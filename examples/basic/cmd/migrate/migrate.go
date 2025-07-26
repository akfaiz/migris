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

func Up() error {
	m, err := newMigrator()
	if err != nil {
		return err
	}
	return goose.Up(m.db, m.dir)
}

func Create(name string) error {
	m, err := newMigrator()
	if err != nil {
		return err
	}
	return goose.Create(m.db, m.dir, name, m.migrationType)
}

func Down() error {
	m, err := newMigrator()
	if err != nil {
		return err
	}
	return goose.Reset(m.db, m.dir)
}

type migrator struct {
	dir           string
	dialect       string
	tableName     string
	migrationType string
	db            *sql.DB
}

func newMigrator() (*migrator, error) {
	db, err := newDatabase(config.GetDatabase())
	if err != nil {
		return nil, err
	}
	m := &migrator{
		dir:           "migrations",
		dialect:       "postgres",
		tableName:     "schema_migrations",
		migrationType: "go",
		db:            db,
	}
	if err := m.init(); err != nil {
		return nil, err
	}

	return m, nil
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

func (m *migrator) init() error {
	goose.SetTableName(m.tableName)
	if err := goose.SetDialect(m.dialect); err != nil {
		return err
	}
	if err := schema.SetDialect(m.dialect); err != nil {
		return err
	}
	return nil
}
