package migris

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/akfaiz/migris/internal/dialect"
	"github.com/akfaiz/migris/internal/logger"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"
)

// Migrate handles database migrations.
type Migrate struct {
	dialect      dialect.Dialect
	driverName   string
	db           *sql.DB
	dsn          string
	ownDB        bool
	migrationDir string
	tableName    string
	logger       *logger.Logger
	registry     *Registry
}

// New creates a new Migrate instance.
func New(dialectValue string, opts ...MigrisOption) (*Migrate, error) {
	dialectVal := dialect.FromString(dialectValue)
	if dialectVal == dialect.Unknown {
		return nil, errors.New("unknown database dialect")
	}

	m := &Migrate{
		dialect:      dialectVal,
		driverName:   dialect.DriverName(dialectVal, dialectValue),
		migrationDir: "migrations",
		tableName:    "schema_migrations",
		logger:       logger.Get(),
		registry:     defaultRegistry,
	}
	for _, opt := range opts {
		opt(m)
	}
	if m.db == nil && m.dsn != "" {
		db, err := sql.Open(m.driverName, m.dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to open database: %w", err)
		}
		m.db = db
		m.ownDB = true
	}
	if m.db == nil {
		return nil, errors.New("database connection is not set, use WithDB or WithDSN")
	}
	return m, nil
}

// Close closes the database connection if it was opened by the migrator via WithDSN.
// If the connection was provided externally via WithDB, this is a no-op.
func (m *Migrate) Close() error {
	if m.ownDB && m.db != nil {
		return m.db.Close()
	}
	return nil
}

// queryCurrentVersion queries the max applied migration version directly,
// returning 0 if the migration table does not yet exist.
func (m *Migrate) queryCurrentVersion(ctx context.Context) int64 {
	//nolint:gosec // tableName is user-configured, not external input
	q := fmt.Sprintf("SELECT COALESCE(MAX(version_id), 0) FROM %s WHERE is_applied = TRUE", m.tableName)
	var version int64
	_ = m.db.QueryRowContext(ctx, q).Scan(&version)
	return version
}

// queryAppliedVersions returns a set of all applied migration versions,
// returning an empty set if the migration table does not yet exist.
func (m *Migrate) queryAppliedVersions(ctx context.Context) (map[int64]bool, error) {
	//nolint:gosec // tableName is user-configured, not external input
	q := fmt.Sprintf("SELECT version_id FROM %s WHERE is_applied = TRUE", m.tableName)
	rows, err := m.db.QueryContext(ctx, q)
	if err != nil {
		return make(map[int64]bool), nil
	}
	defer rows.Close()

	applied := make(map[int64]bool)
	for rows.Next() {
		var v int64
		if scanErr := rows.Scan(&v); scanErr != nil {
			return nil, scanErr
		}
		applied[v] = true
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, rowsErr
	}
	return applied, nil
}

func (m *Migrate) newProvider(ro runOptions) (*goose.Provider, error) {
	gooseDialect := m.dialect.GooseDialect()
	store, err := database.NewStore(gooseDialect, m.tableName)
	if err != nil {
		return nil, err
	}
	provider, err := goose.NewProvider(database.DialectCustom, m.db, os.DirFS(m.migrationDir),
		goose.WithStore(store),
		goose.WithDisableGlobalRegistry(true),
		goose.WithGoMigrations(m.registry.gooseMigrations(m.dialect)...),
		goose.WithAllowOutofOrder(ro.allowMissing),
	)
	if err != nil {
		return nil, err
	}
	return provider, nil
}
