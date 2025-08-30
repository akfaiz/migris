package migris

import (
	"context"
	"database/sql"
	"errors"

	"github.com/afkdevs/migris/internal/logger"
	"github.com/pressly/goose/v3"
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
	currentVersion, err := provider.GetDBVersion(ctx)
	if err != nil {
		return err
	}
	if currentVersion == 0 {
		logger.Info("Nothing to rollback.")
		return nil
	}
	logger.Info("Rolling back migrations.\n")
	results, err := provider.DownTo(ctx, 0)
	if err != nil {
		var partialErr *goose.PartialError
		if errors.As(err, &partialErr) {
			logger.PrintResults(partialErr.Applied)
			logger.PrintResult(partialErr.Failed)
		}

		return err
	}
	logger.PrintResults(results)
	return nil
}
