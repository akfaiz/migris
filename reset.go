package migris

import (
	"context"
	"database/sql"
)

func Reset(db *sql.DB, dir string) error {
	ctx := context.Background()
	return ResetContext(ctx, db, dir)
}

func ResetContext(ctx context.Context, db *sql.DB, dir string) error {
	provider, err := newProvider(db, dir)
	if err != nil {
		return err
	}
	_, err = provider.DownTo(ctx, 0)
	if err != nil {
		return err
	}
	return nil
}
