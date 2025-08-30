package migris

import "database/sql"

// Migrate handles database migrations
type Migrate struct {
	dir string
	db  *sql.DB
}

// New creates a new Migrate instance
func New(db *sql.DB, dir string) *Migrate {
	return &Migrate{
		dir: dir,
		db:  db,
	}
}
