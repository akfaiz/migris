package migrate

import (
	"database/sql"
	"fmt"

	"github.com/afkdevs/migris"
	"github.com/afkdevs/migris/examples/basic/config"
	_ "github.com/afkdevs/migris/examples/basic/migrations"
	_ "github.com/lib/pq" // PostgreSQL driver
)

const (
	directory = "migrations"
)

func Up() error {
	db, err := initMigrator()
	if err != nil {
		return err
	}
	return migris.Up(db, directory)
}

func Create(name string) error {
	return migris.Create(directory, name)
}

func Reset() error {
	db, err := initMigrator()
	if err != nil {
		return err
	}
	return migris.Reset(db, directory)
}

func initMigrator() (*sql.DB, error) {
	if err := migris.SetDialect("postgres"); err != nil {
		return nil, fmt.Errorf("failed to set schema dialect: %w", err)
	}
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	db, err := newDatabase(cfg.Database)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func newDatabase(cfg config.Database) (*sql.DB, error) {
	dsn := cfg.DSN()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
