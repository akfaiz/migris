package migris

import "database/sql"

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
