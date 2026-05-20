package migris

import (
	"context"
	"errors"

	"github.com/pressly/goose/v3"
)

// Reset rolls back all migrations.
func (m *Migrate) Reset(opts ...Option) error {
	ctx := context.Background()
	return m.ResetContext(ctx, opts...)
}

// ResetContext rolls back all migrations.
func (m *Migrate) ResetContext(ctx context.Context, opts ...Option) error {
	ro := applyRunOptions(opts)
	if ro.dryRun {
		return m.DownToContext(ctx, 0, opts...)
	}

	provider, err := m.newProvider(ro)
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
