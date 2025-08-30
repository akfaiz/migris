package migris

import (
	"context"
	"errors"

	"github.com/afkdevs/migris/internal/logger"
	"github.com/pressly/goose/v3"
)

// Up applies the migrations in the specified directory.
func (m *Migrate) Up() error {
	ctx := context.Background()
	return m.UpContext(ctx)
}

// UpContext applies the migrations in the specified directory.
func (m *Migrate) UpContext(ctx context.Context) error {
	provider, err := newProvider(m.db, m.dir)
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
