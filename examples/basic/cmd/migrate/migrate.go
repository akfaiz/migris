package migrate

import (
	"database/sql"
	"fmt"

	"github.com/akfaiz/migris"
	"github.com/akfaiz/migris/examples/basic/config"
	_ "github.com/akfaiz/migris/examples/basic/migrations"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	migrationDir = "migrations"
)

func Up() error {
	m, err := newMigrate()
	if err != nil {
		return err
	}
	return m.Up()
}

func Create(name string) error {
	return migris.Create(migrationDir, name)
}

func Reset() error {
	m, err := newMigrate()
	if err != nil {
		return err
	}
	return m.Reset()
}

func Down() error {
	m, err := newMigrate()
	if err != nil {
		return err
	}
	return m.Down()
}

func Status() error {
	m, err := newMigrate()
	if err != nil {
		return err
	}
	return m.Status()
}

func newMigrate() (*migris.Migrate, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	db, err := openDatabase(cfg.Database)
	if err != nil {
		return nil, err
	}
	migrate, err := migris.New("pgx", migris.WithDB(db), migris.WithMigrationDir(migrationDir))
	if err != nil {
		return nil, fmt.Errorf("failed to create migris instance: %w", err)
	}
	return migrate, nil
}

func openDatabase(cfg config.Database) (*sql.DB, error) {
	dsn := cfg.DSN()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	return db, nil
}
