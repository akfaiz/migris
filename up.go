package migris

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/akfaiz/migris/schema"
	"github.com/pressly/goose/v3"
)

// Up applies the migrations in the specified directory.
func (m *Migrate) Up() error {
	ctx := context.Background()
	return m.UpContext(ctx)
}

// UpContext applies the migrations in the specified directory.
func (m *Migrate) UpContext(ctx context.Context) error {
	return m.UpToContext(ctx, goose.MaxVersion)
}

// UpTo applies the migrations up to the specified version.
func (m *Migrate) UpTo(version int64) error {
	ctx := context.Background()
	return m.UpToContext(ctx, version)
}

// UpToContext applies the migrations up to the specified version.
func (m *Migrate) UpToContext(ctx context.Context, version int64) error {
	// Set global dry-run state for migration execution
	setGlobalDryRunState(m.dryRun)
	defer setGlobalDryRunState(false) // Reset after execution

	if m.dryRun {
		return m.executeDryRunUp(ctx, version)
	}

	provider, err := m.newProvider()
	if err != nil {
		return err
	}
	hasPending, err := provider.HasPending(ctx)
	if err != nil {
		return err
	}
	if !hasPending {
		m.logger.Info("Nothing to migrate.")
		return nil
	}

	m.logger.Infof("Running migrations.\n")
	results, err := provider.UpTo(ctx, version)
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

// executeDryRunUp executes migrations in dry-run mode.
func (m *Migrate) executeDryRunUp(ctx context.Context, version int64) error {
	// Create provider to check migration status
	provider, err := m.newProvider()
	if err != nil {
		return fmt.Errorf("cannot connect to database for dry-run: %w", err)
	}

	// Check if there are pending migrations
	hasPending, err := provider.HasPending(ctx)
	if err != nil {
		return fmt.Errorf("cannot check pending migrations: %w", err)
	}
	if !hasPending {
		m.logger.Info("Nothing to migrate.")
		return nil
	}

	m.logger.DryRunStart(version)
	// Get current database version
	currentVersion, err := provider.GetDBVersion(ctx)
	if err != nil {
		return fmt.Errorf("cannot get current database version: %w", err)
	}

	// Get migrations to apply
	migrationsToApply := m.determineMigrationsToApply(version, currentVersion)

	// Process migrations in dry-run mode
	totalMigrations, totalStatements, _, err := m.processDryRunUpMigrations(ctx, migrationsToApply)
	if err != nil {
		return err
	}

	// Print summary
	m.logger.DryRunSummary(totalMigrations, totalStatements)

	return nil
}

// determineMigrationsToApply determines which migrations should be applied.
func (m *Migrate) determineMigrationsToApply(version, currentVersion int64) []*Migration {
	var migrationsToApply []*Migration

	// Get all registered migrations that need to be applied (only pending ones)
	for _, migration := range registeredMigrations {
		// Skip migrations that are already applied
		if migration.version <= currentVersion {
			continue
		}

		if version != goose.MaxVersion && migration.version > version {
			break
		}

		migrationsToApply = append(migrationsToApply, migration)
	}

	return migrationsToApply
}

// processDryRunMigrations processes migrations in dry-run mode (common logic for up and down).
func (m *Migrate) processDryRunMigrations(
	ctx context.Context,
	migrations []*Migration,
	isUp bool,
) (int, int, float64, error) {
	startTime := time.Now()
	totalStatements := 0
	totalMigrations := 0

	for _, migration := range migrations {
		migrationStartTime := time.Now()
		totalMigrations++

		m.logger.DryRunMigrationStart(filepath.Base(migration.source), migration.version)

		// Create dry-run context for this migration
		dryRunCtx := schema.NewDryRunContext(ctx)

		// Execute the migration in dry-run mode
		var migrationFunc MigrationContext
		var direction string
		if isUp {
			migrationFunc = migration.upFnContext
			direction = "up"
		} else {
			migrationFunc = migration.downFnContext
			direction = "down"
		}

		if migrationFunc != nil {
			err := migrationFunc(dryRunCtx)
			if err != nil {
				return 0, 0, 0, fmt.Errorf("dry-run %s migration %s failed: %w", direction, migration.source, err)
			}

			capturedSQL := dryRunCtx.GetCapturedSQL()
			totalStatements += len(capturedSQL)

			// Print captured SQL
			if dryRunCtx.HasPendingQuery() {
				queries := dryRunCtx.GetPendingQueries()
				for _, q := range queries {
					m.logger.DryRunSQL(q.Query, q.Args...)
				}
			}
		}

		migrationDuration := time.Since(migrationStartTime).Seconds() * 1000
		m.logger.DryRunMigrationComplete(filepath.Base(migration.source), migrationDuration)
	}

	duration := time.Since(startTime).Seconds() * 1000
	return totalMigrations, totalStatements, duration, nil
}

// processDryRunUpMigrations processes migrations to apply in dry-run mode.
func (m *Migrate) processDryRunUpMigrations(
	ctx context.Context,
	migrationsToApply []*Migration,
) (int, int, float64, error) {
	return m.processDryRunMigrations(ctx, migrationsToApply, true)
}
