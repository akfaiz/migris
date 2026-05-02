package migris

import (
	"context"
	"errors"

	"github.com/pressly/goose/v3"
)

// Reset rolls back all migrations.
func (m *Migrate) Reset() error {
	ctx := context.Background()
	return m.ResetContext(ctx)
}

// ResetContext rolls back all migrations.
func (m *Migrate) ResetContext(ctx context.Context) error {
	// Check if dry-run mode is enabled
	if m.dryRun {
		return m.DownToContext(ctx, 0) // Use DownToContext with version 0 for reset
	}

	provider, err := m.newProvider()
	if err != nil {
		return err
	}
	currentVersion, err := provider.GetDBVersion(ctx)
	if err != nil {
		return err
	}
	if currentVersion == 0 {
		m.logger.Info("Nothing to rollback.")
		return nil
	}
	m.logger.Info("Rolling back migrations.\n")
	results, err := provider.DownTo(ctx, 0)
	if err != nil {
		var partialErr *goose.PartialError
		if errors.As(err, &partialErr) {
			m.logger.PrintResults(partialErr.Applied)
			m.logger.PrintResult(partialErr.Failed)
		}

		return err
	}
	m.logger.PrintResults(results)
	return nil
}
