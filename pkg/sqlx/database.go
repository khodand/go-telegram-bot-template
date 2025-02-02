package sqlx

import (
	"database/sql"
	"fmt"
	"runtime"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	// MaxIdleConns is the maximum number of connections in the idle connection
	// pool.
	// Default: 10 * GOMAXPROCS.
	MaxIdleConns int `yaml:"maxIdleConns" env:"MAX_IDLE_CONNS"`

	// MaxOpenConns is the maximum number of open connections to the database.
	// Default: 10 * GOMAXPROCS.
	MaxOpenConns int `yaml:"maxOpenConns" env:"MAX_OPEN_CONNS"`

	// ConnMaxIdleTime is the maximum amount of time a connection may be idle.
	// Default: 1 minute.
	ConnMaxIdleTime time.Duration `yaml:"connMaxIdleTime" env:"CONN_MAX_IDLE_TIME"`

	// ConnMaxLifeTime is the maximum amount of time a connection may be reused.
	// Default: 5 minutes.
	ConnMaxLifeTime time.Duration `yaml:"connMaxLifeTime" env:"CONN_MAX_LIFE_TIME"`

	Database string `yaml:"database" env:"DATABASE"`
	Username string `yaml:"username" env:"USERNAME"`
	Password string `yaml:"password" env:"PASSWORD"`
	Host     string `yaml:"host" env:"HOST_PRIMARY"`
	Port     string `yaml:"port" env:"PORT"`
	SSLMode  string `yaml:"sslmode" env:"SSLMODE"`
}

func NewDatabase(config Config) (*sqlx.DB, error) {
	conf, err := pgx.ParseConfig(config.createDSN())
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	db := stdlib.OpenDB(*conf)

	setConnectionOptions(db, config)

	return sqlx.NewDb(db, "pgx").Unsafe(), nil
}

func (c Config) createDSN() string {
	return fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=%s",
		c.Username, c.Password, c.Host, c.Port, c.Database, c.SSLMode)
}

//nolint:gomnd // default config
func setConnectionOptions(db *sql.DB, config Config) {
	gomaxprocs := runtime.GOMAXPROCS(0)

	maxOpenConns := 10 * gomaxprocs
	if config.MaxOpenConns != 0 {
		maxOpenConns = config.MaxOpenConns
	}
	maxIdleConns := 10 * gomaxprocs
	if config.MaxIdleConns != 0 {
		maxIdleConns = config.MaxIdleConns
	}
	connMaxLifetime := 5 * time.Minute
	if config.ConnMaxLifeTime != 0 {
		connMaxLifetime = config.ConnMaxLifeTime
	}
	connMaxIdleTime := 1 * time.Minute
	if config.ConnMaxIdleTime != 0 {
		connMaxIdleTime = config.ConnMaxIdleTime
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)
	db.SetConnMaxIdleTime(connMaxIdleTime)
}
