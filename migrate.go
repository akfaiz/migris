package migris

import (
	"database/sql"
	"errors"
	"os"

	"github.com/akfaiz/migris/internal/config"
	"github.com/akfaiz/migris/internal/dialect"
	"github.com/akfaiz/migris/internal/logger"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"
)

// Migrate handles database migrations.
type Migrate struct {
	dialect      dialect.Dialect
	db           *sql.DB
	migrationDir string
	tableName    string
	dryRun       bool
	logger       *logger.Logger
}

// New creates a new Migrate instance.
func New(dialectValue string, opts ...Option) (*Migrate, error) {
	dialectVal := dialect.FromString(dialectValue)
	if dialectVal == dialect.Unknown {
		return nil, errors.New("unknown database dialect")
	}
	config.SetDialect(dialectVal)

	m := &Migrate{
		dialect:      dialectVal,
		migrationDir: "migrations",
		tableName:    "schema_migrations",
		logger:       logger.Get(),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m, nil
}

func (m *Migrate) newProvider() (*goose.Provider, error) {
	val := config.GetDialect()
	if val == dialect.Unknown {
		return nil, errors.New("unknown database dialect")
	}
	if m.db == nil {
		return nil, errors.New("database connection is not set, please call WithDB option")
	}
	gooseDialect := val.GooseDialect()
	store, err := database.NewStore(gooseDialect, m.tableName)
	if err != nil {
		return nil, err
	}
	provider, err := goose.NewProvider(database.DialectCustom, m.db, os.DirFS(m.migrationDir),
		goose.WithStore(store),
		goose.WithDisableGlobalRegistry(true),
		goose.WithGoMigrations(gooseMigrations()...),
	)
	if err != nil {
		return nil, err
	}
	return provider, nil
}
