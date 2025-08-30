package migris

import (
	"context"
	"database/sql"
	"errors"

	"github.com/afkdevs/migris/internal/logger"
	"github.com/pressly/goose/v3"
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
	currentVersion, err := provider.GetDBVersion(ctx)
	if err != nil {
		return err
	}
	if currentVersion == 0 {
		logger.Info("Nothing to rollback.")
		return nil
	}
	logger.Info("Rolling back migrations.\n")
	result, err := provider.Down(ctx)
	if err != nil {
		var partialErr *goose.PartialError
		if errors.As(err, &partialErr) {
			logger.PrintResult(partialErr.Failed)
		}
		return err
	}
	if result != nil {
		logger.PrintResult(result)
	}
	return nil
}
