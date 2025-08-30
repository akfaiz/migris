package migrate

import (
	"database/sql"
	"fmt"

	"github.com/afkdevs/go-schema"
	"github.com/afkdevs/go-schema/examples/basic/config"
	_ "github.com/afkdevs/go-schema/examples/basic/migrations"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/pressly/goose/v3"
)

const (
	directory = "migrations"
	tableName = "schema_migrations"
)

func Up(dryRun bool) error {
	db, err := initMigrator()
	if err != nil {
		return err
	}
	return goose.Up(db, directory)
}

func Create(name string) error {
	tmpl := schema.GooseMigrationTemplate(name)
	return goose.CreateWithTemplate(nil, directory, tmpl, name, "go")
}

func Reset(dryRun bool) error {
	db, err := initMigrator()
	if err != nil {
		return err
	}
	return goose.Reset(db, directory)
}

func initMigrator() (*sql.DB, error) {
	if err := goose.SetDialect("postgres"); err != nil {
		return nil, fmt.Errorf("failed to set dialect: %w", err)
	}
	goose.SetTableName(tableName)
	if err := schema.SetDialect("postgres"); err != nil {
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
