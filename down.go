package migris

import (
	"context"
	"errors"
	"fmt"

	"github.com/pressly/goose/v3"
)

// Down rolls back the last migration.
func (m *Migrate) Down(opts ...Option) error {
	ctx := context.Background()
	return m.DownContext(ctx, opts...)
}

// DownContext rolls back the last migration.
func (m *Migrate) DownContext(ctx context.Context, opts ...Option) error {
	ro := applyRunOptions(opts)
	if ro.dryRun {
		return m.executeDryRunDown(ctx, -1, ro)
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
	result, err := provider.Down(ctx)
	if err != nil {
		var partialErr *goose.PartialError
		if errors.As(err, &partialErr) {
			m.logger.PrintResult(partialErr.Failed)
		}
		return err
	}
	if result != nil {
		m.logger.PrintResult(result)
	}
	return nil
}

// DownTo rolls back the migrations to the specified version.
func (m *Migrate) DownTo(version int64, opts ...Option) error {
	ctx := context.Background()
	return m.DownToContext(ctx, version, opts...)
}

// DownToContext rolls back the migrations to the specified version.
func (m *Migrate) DownToContext(ctx context.Context, version int64, opts ...Option) error {
	ro := applyRunOptions(opts)
	if ro.dryRun {
		return m.executeDryRunDown(ctx, version, ro)
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
	results, err := provider.DownTo(ctx, version)
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

// executeDryRunDown executes migrations in dry-run mode for down operations.
func (m *Migrate) executeDryRunDown(ctx context.Context, version int64, ro runOptions) error {
	provider, err := m.newProvider(ro)
	if err != nil {
		return fmt.Errorf("cannot connect to database for dry-run: %w", err)
	}

	currentVersion, err := provider.GetDBVersion(ctx)
	if err != nil {
		return fmt.Errorf("cannot get current database version: %w", err)
	}

	if currentVersion == 0 {
		m.logger.Info("Nothing to rollback.")
		return nil
	}

	m.logger.DryRunDownStart(version)
	migrationsToRollback := m.determineMigrationsToRollback(version, currentVersion)
	if len(migrationsToRollback) == 0 {
		m.logger.Info("Nothing to rollback.")
		return nil
	}

	totalMigrations, totalStatements, _, err := m.processDryRunDownMigrations(ctx, migrationsToRollback)
	if err != nil {
		return err
	}

	operation := "DOWN"
	if version == 0 {
		operation = "RESET"
	}
	m.logger.DryRunDownSummary(totalMigrations, totalStatements, operation)

	return nil
}

// determineMigrationsToRollback determines which migrations should be rolled back.
func (m *Migrate) determineMigrationsToRollback(version, currentVersion int64) []*Migration {
	var migrationsToRollback []*Migration

	if version == -1 {
		registeredMigrations := m.registry.migrationsSnapshot()
		for i := len(registeredMigrations) - 1; i >= 0; i-- {
			migration := registeredMigrations[i]
			if migration.version <= currentVersion {
				migrationsToRollback = append(migrationsToRollback, migration)
				break
			}
		}
	} else {
		registeredMigrations := m.registry.migrationsSnapshot()
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
