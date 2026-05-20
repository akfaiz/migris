package migris

import "database/sql"

// MigrisOption configures a Migrate instance at construction time.
//
//nolint:revive // MigrisOption prefix is intentional to disambiguate from the run-time Option type.
type MigrisOption func(*Migrate)

// WithTableName sets the table name for the migration.
func WithTableName(name string) MigrisOption {
	return func(m *Migrate) {
		m.tableName = name
	}
}

// WithMigrationDir sets the directory for the migration files.
func WithMigrationDir(dir string) MigrisOption {
	return func(m *Migrate) {
		m.migrationDir = dir
	}
}

// WithDB sets the database connection for the migration.
func WithDB(db *sql.DB) MigrisOption {
	return func(m *Migrate) {
		m.db = db
	}
}

// WithDSN opens a database connection from the given DSN string.
// The migrator takes ownership of the connection; call Close when done.
func WithDSN(dsn string) MigrisOption {
	return func(m *Migrate) {
		m.dsn = dsn
	}
}

// WithRegistry sets the migration registry for the migrator.
func WithRegistry(registry *Registry) MigrisOption {
	return func(m *Migrate) {
		if registry != nil {
			m.registry = registry
		}
	}
}

// Option configures a single run operation (Up, Down, Reset, etc.).
type Option func(*runOptions)

type runOptions struct {
	dryRun       bool
	allowMissing bool
}

func applyRunOptions(opts []Option) runOptions {
	ro := runOptions{}
	for _, opt := range opts {
		opt(&ro)
	}
	return ro
}

// WithDryRun simulates the migration without applying changes.
func WithDryRun(enabled bool) Option {
	return func(o *runOptions) {
		o.dryRun = enabled
	}
}

// WithAllowMissing allows out-of-order migrations to be applied.
func WithAllowMissing(enabled bool) Option {
	return func(o *runOptions) {
		o.allowMissing = enabled
	}
}
