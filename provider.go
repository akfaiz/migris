package migris

import (
	"database/sql"
	"errors"
	"os"

	"github.com/afkdevs/migris/internal/config"
	"github.com/afkdevs/migris/internal/dialect"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"
)

func newProvider(db *sql.DB, dir string) (*goose.Provider, error) {
	cfg := config.Get()
	if cfg.Dialect == dialect.Unknown {
		return nil, errors.New("unknown database dialect")
	}
	dialect := cfg.Dialect.GooseDialect()
	store, err := database.NewStore(dialect, cfg.TableName)
	if err != nil {
		return nil, err
	}
	provider, err := goose.NewProvider(database.DialectCustom, db, os.DirFS(dir),
		goose.WithStore(store),
		goose.WithDisableGlobalRegistry(true),
		goose.WithGoMigrations(getMigrations()...),
		goose.WithVerbose(cfg.Verbose),
	)
	if err != nil {
		return nil, err
	}
	return provider, nil
}
