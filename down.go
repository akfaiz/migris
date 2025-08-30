package migris

import (
	"context"
	"errors"

	"github.com/akfaiz/migris/internal/logger"
	"github.com/pressly/goose/v3"
)

// Down rolls back the last migration.
func (m *Migrate) Down() error {
	ctx := context.Background()
	return m.DownContext(ctx)
}

// DownContext rolls back the last migration.
func (m *Migrate) DownContext(ctx context.Context) error {
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

// DownTo rolls back the migrations to the specified version.
func (m *Migrate) DownTo(version int64) error {
	ctx := context.Background()
	return m.DownToContext(ctx, version)
}

// DownToContext rolls back the migrations to the specified version.
func (m *Migrate) DownToContext(ctx context.Context, version int64) error {
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
	results, err := provider.DownTo(ctx, version)
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
