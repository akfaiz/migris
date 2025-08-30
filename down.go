package migris

import (
	"context"
	"database/sql"
)

func Down(db *sql.DB, dir string) error {
	ctx := context.Background()
	return DownContext(ctx, db, dir)
}

func DownContext(ctx context.Context, db *sql.DB, dir string) error {
	provider, err := newProvider(db, dir)
	if err != nil {
		return err
	}
	_, err = provider.Down(ctx)
	if err != nil {
		return err
	}
	return nil
}
