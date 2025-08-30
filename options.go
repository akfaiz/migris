package migris

import "database/sql"

type Option func(*Migrate)

// WithTableName sets the table name for the migration.
func WithTableName(name string) Option {
	return func(m *Migrate) {
		m.tableName = name
	}
}

// WithMigrationPath sets the directory for the migration files.
func WithMigrationPath(path string) Option {
	return func(m *Migrate) {
		m.migrationPath = path
	}
}

// WithDB sets the database connection for the migration.
func WithDB(db *sql.DB) Option {
	return func(m *Migrate) {
		m.db = db
	}
}
