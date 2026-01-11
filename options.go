package migris

import (
	"database/sql"

	"github.com/akfaiz/migris/schema"
)

type Option func(*Migrate)

// WithTableName sets the table name for the migration.
func WithTableName(name string) Option {
	return func(m *Migrate) {
		m.tableName = name
	}
}

// WithMigrationDir sets the directory for the migration files.
func WithMigrationDir(dir string) Option {
	return func(m *Migrate) {
		m.migrationDir = dir
	}
}

// WithDB sets the database connection for the migration.
func WithDB(db *sql.DB) Option {
	return func(m *Migrate) {
		m.db = db
	}
}

// WithDryRun enables or disables dry-run mode.
func WithDryRun(enabled bool) Option {
	return func(m *Migrate) {
		m.dryRun = enabled
		m.dryRunConfig.PrintSQL = true
		m.dryRunConfig.PrintMigrations = true
	}
}

// WithDryRunConfig sets the complete dry-run configuration.
func WithDryRunConfig(config schema.DryRunConfig) Option {
	return func(m *Migrate) {
		m.dryRun = true
		m.dryRunConfig = config
	}
}
