package migris

import (
	"context"
	"database/sql"
)

// Up applies the migrations in the specified directory.
func Up(db *sql.DB, dir string) error {
	ctx := context.Background()
	return UpContext(ctx, db, dir)
}

// UpContext applies the migrations in the specified directory.
func UpContext(ctx context.Context, db *sql.DB, dir string) error {
	provider, err := newProvider(db, dir)
	if err != nil {
		return err
	}
	_, err = provider.Up(ctx)
	if err != nil {
		return err
	}
	return nil
}
