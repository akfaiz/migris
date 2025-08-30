package migris

import (
	"context"
	"database/sql"
	"errors"

	"github.com/afkdevs/migris/internal/logger"
	"github.com/pressly/goose/v3"
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
	hasPending, err := provider.HasPending(ctx)
	if err != nil {
		return err
	}
	if !hasPending {
		logger.Info("Nothing to migrate.")
		return nil
	}
	logger.Infof("Running migrations.\n")

	results, err := provider.Up(ctx)
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
