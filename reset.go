package migris

import (
	"context"
	"errors"

	"github.com/afkdevs/migris/internal/logger"
	"github.com/pressly/goose/v3"
)

// Reset rolls back all migrations.
func (m *Migrate) Reset() error {
	ctx := context.Background()
	return m.ResetContext(ctx)
}

// ResetContext rolls back all migrations.
func (m *Migrate) ResetContext(ctx context.Context) error {
	provider, err := m.newProvider()
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
