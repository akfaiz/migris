package migris

import (
	"context"
	"errors"

	"github.com/afkdevs/migris/internal/logger"
	"github.com/pressly/goose/v3"
)

// Down rolls back the last migration.
func (m *Migrate) Down() error {
	ctx := context.Background()
	return m.DownContext(ctx)
}

// DownContext rolls back the last migration.
func (m *Migrate) DownContext(ctx context.Context) error {
	provider, err := newProvider(m.db, m.dir)
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
