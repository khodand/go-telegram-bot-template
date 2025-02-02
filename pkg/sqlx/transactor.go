package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type DB interface {
	sqlx.ExtContext
}

type DBGetter interface {
	Primary(ctx context.Context) DB
	Replica() DB
}

type DBTransactor interface {
	Exec(baseContext context.Context, th func(ctx context.Context) error) error
}

type TransactorConfig struct {
	Isolation sql.IsolationLevel `yaml:"isolation_level"`
	ReadOnly  bool               `yaml:"read_only"`
}

type Transactor struct {
	logger *zap.Logger
	db     *sqlx.DB
	opts   *sql.TxOptions
}

type ConnContainer struct {
	primary *sqlx.DB
	replica *sqlx.DB
}

type txKey struct{}

// NewConnContainer creates container with primary and replica connections.
func NewConnContainer(primary, replica *sqlx.DB) *ConnContainer {
	return &ConnContainer{
		primary: primary,
		replica: replica,
	}
}

// NewTransactor creates new transaction manager with tx options support.
func NewTransactor(db *sqlx.DB, logger *zap.Logger, cfg *TransactorConfig) *Transactor {
	return &Transactor{
		logger: logger,
		db:     db,
		opts: &sql.TxOptions{
			Isolation: cfg.Isolation,
			ReadOnly:  cfg.ReadOnly,
		},
	}
}

func (t *Transactor) Exec(baseContext context.Context, th func(ctx context.Context) error) (err error) {
	tx, ctx, parentTx, err := t.GetTx(baseContext)
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			t.logger.Error("Panic recovered", zap.Any("error", r), zap.Stack("stack"))
			err = fmt.Errorf("panic recovered: %v", r)
		}

		if parentTx && err != nil {
			if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
				t.logger.Error("Rolling back tx error", zap.Error(rbErr))
				err = fmt.Errorf("rollback error: %w, rollback cause: %w", rbErr, err)
			}
		}
	}()

	err = th(ctx)
	if err != nil {
		return err
	}

	if parentTx {
		if err = tx.Commit(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			return fmt.Errorf("committing tx: %w", err)
		}
	}

	return err
}

func (t *Transactor) GetTx(baseContext context.Context) (
	tx *sqlx.Tx,
	enrichedCtx context.Context,
	isParent bool,
	err error,
) {
	if tx = getTx(baseContext); tx != nil {
		return tx, baseContext, false, nil
	}

	tx, err = t.db.BeginTxx(baseContext, t.opts)
	if err != nil {
		return nil, baseContext, true, err
	}

	return tx, context.WithValue(baseContext, txKey{}, tx), true, nil
}

func (g *ConnContainer) Primary(ctx context.Context) DB {
	if tx := getTx(ctx); tx != nil {
		return tx
	}
	return g.primary
}

func (g *ConnContainer) Replica() DB {
	return g.replica
}

func getTx(ctx context.Context) *sqlx.Tx {
	o := ctx.Value(txKey{})
	if o == nil {
		return nil
	}
	if tx, ok := o.(*sqlx.Tx); ok {
		return tx
	}
	return nil
}
