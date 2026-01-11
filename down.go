package migris

import (
	"context"
	"errors"
	"fmt"

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
	// Check if dry-run mode is enabled
	if m.dryRun {
		return m.executeDryRunDown(ctx, -1) // -1 means rollback last migration
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
	// Check if dry-run mode is enabled
	if m.dryRun {
		return m.executeDryRunDown(ctx, version)
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

// executeDryRunDown executes migrations in dry-run mode for down operations.
func (m *Migrate) executeDryRunDown(ctx context.Context, version int64) error {
	// Create provider to check migration status
	provider, err := m.newProvider()
	if err != nil {
		return fmt.Errorf("cannot connect to database for dry-run: %w", err)
	}

	// Get current database version
	currentVersion, err := provider.GetDBVersion(ctx)
	if err != nil {
		return fmt.Errorf("cannot get current database version: %w", err)
	}

	if currentVersion == 0 {
		logger.Info("Nothing to rollback.")
		return nil
	}

	logger.DryRunDownStart(version)

	// Determine which migrations to rollback
	migrationsToRollback := m.determineMigrationsToRollback(version, currentVersion)
	if len(migrationsToRollback) == 0 {
		logger.Info("Nothing to rollback.")
		return nil
	}

	// Process migrations in dry-run mode
	totalMigrations, totalStatements, duration, err := m.processDryRunDownMigrations(ctx, migrationsToRollback)
	if err != nil {
		return err
	}

	// Print summary
	operation := "DOWN"
	if version == 0 {
		operation = "RESET"
	}
	logger.DryRunDownSummary(totalMigrations, totalStatements, duration, operation)

	return nil
}

// determineMigrationsToRollback determines which migrations should be rolled back.
func (m *Migrate) determineMigrationsToRollback(version, currentVersion int64) []*Migration {
	var migrationsToRollback []*Migration

	if version == -1 {
		// Rollback last applied migration only
		for i := len(registeredMigrations) - 1; i >= 0; i-- {
			migration := registeredMigrations[i]
			if migration.version <= currentVersion {
				migrationsToRollback = append(migrationsToRollback, migration)
				break // Only the last one
			}
		}
	} else {
		// Rollback migrations down to specified version (only applied ones)
		for i := len(registeredMigrations) - 1; i >= 0; i-- {
			migration := registeredMigrations[i]
			if migration.version > version && migration.version <= currentVersion {
				migrationsToRollback = append(migrationsToRollback, migration)
			}
		}
	}

	return migrationsToRollback
}

// processDryRunDownMigrations processes migrations to rollback in dry-run mode.
func (m *Migrate) processDryRunDownMigrations(
	ctx context.Context,
	migrationsToRollback []*Migration,
) (int, int, float64, error) {
	return m.processDryRunMigrations(ctx, migrationsToRollback, false)
}
