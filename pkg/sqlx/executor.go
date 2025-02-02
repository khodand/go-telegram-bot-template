package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"go-template/pkg/contexts"
	"go-template/pkg/errx"
)

func QueryRowxContext[T any](ctx context.Context, conn DBGetter, query string, args ...interface{}) (*T, error) {
	var result *T
	err := wrapExecution(ctx, query, func() error {
		// we always use primary connection because we don't have replica
		row := conn.Primary(ctx).QueryRowxContext(ctx, query, args...)

		var err error
		result, err = parseRow[T](row)
		return err
	})
	return result, err
}

func QueryxContext[T any](ctx context.Context, conn DBGetter, query string, args ...interface{}) ([]*T, error) {
	var result []*T
	err := wrapExecution(ctx, query, func() error {
		// we always use primary connection because we don't have replica
		rows, err := conn.Primary(ctx).QueryxContext(ctx, query, args...)
		if err != nil {
			return err
		}

		result, err = parseRows[T](rows)
		return err
	})

	return result, err
}

func NamedQueryContext[T any](ctx context.Context, conn DBGetter, query string, args interface{}) ([]*T, error) {
	var result []*T
	err := wrapExecution(ctx, query, func() error {
		// we always use primary connection because we don't have replica
		rows, err := sqlx.NamedQueryContext(ctx, conn.Primary(ctx), query, args)
		if err != nil {
			return err
		}

		result, err = parseRows[T](rows)
		return err
	})
	return result, err
}

func ExecContext(ctx context.Context, conn DBGetter, query string, args ...interface{}) error {
	return wrapExecution(ctx, query, func() error {
		_, err := conn.Primary(ctx).ExecContext(ctx, query, args...)
		return err
	})
}

func wrapExecution(ctx context.Context, query string, exec func() error) error {
	logger := contexts.GetLogger(ctx)
	logger.Debug("SQL query start", zap.String("query", query))
	startTime := time.Now()

	err := exec()

	duration := time.Since(startTime)
	if err != nil {
		logger.Debug("SQL query failed", zap.Error(err), zap.Duration("duration", duration))
	} else {
		logger.Debug("SQL query finished", zap.Duration("duration", duration))
	}

	return handleError(err)
}

func parseRows[T any](rows *sqlx.Rows) ([]*T, error) {
	defer rows.Close()
	var result []*T

	for rows.Next() {
		var obj T
		if e := rows.StructScan(&obj); e != nil {
			return nil, e
		}

		result = append(result, &obj)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func parseRow[T any](row *sqlx.Row) (*T, error) {
	model := new(T)
	if err := row.StructScan(model); err != nil {
		return nil, err
	}

	return model, nil
}

func handleError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return errx.ErrNotFound
	case isUniqueViolatesError(err):
		return fmt.Errorf("%w: %w", errx.ErrDuplicateKey, err)
	}

	return err
}

// isUniqueViolatesError: 23505 (A violation of the constraint imposed by a unique index or a unique constraint).
func isUniqueViolatesError(err error) bool {
	const code = "23505"
	return isMatchPGError(err, code)
}

// isExclusionViolatesError: 23P01 (A violation of the constraint imposed by an exclusion constraint).
// func isExclusionViolatesError(err error) bool {
//	const code = "23P01"
//	return isMatchPGError(err, code)
//}

func isMatchPGError(err error, code string) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}

	return pgErr.Code == code
}
