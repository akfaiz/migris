package migris

import (
	"context"
	"database/sql"

	"github.com/afkdevs/migris/internal/logger"
)

func Status(db *sql.DB, dir string) error {
	ctx := context.Background()
	return StatusContext(ctx, db, dir)
}

func StatusContext(ctx context.Context, db *sql.DB, dir string) error {
	provider, err := newProvider(db, dir)
	if err != nil {
		return err
	}
	migrations, err := provider.Status(ctx)
	if err != nil {
		return err
	}
	logger.PrintStatuses(migrations)
	return nil
}
