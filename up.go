package migris

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/akfaiz/migris/internal/logger"
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
		logger.Info("Nothing to migrate.")
		return nil
	}

	logger.Infof("Running migrations.\n")
	results, err := provider.UpTo(ctx, version)
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

// executeDryRunUp executes migrations in dry-run mode
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
		logger.Info("Nothing to migrate.")
		return nil
	}

	logger.DryRunStart(version)

	startTime := time.Now()
	totalStatements := 0
	totalMigrations := 0

	// Get current database version
	currentVersion, err := provider.GetDBVersion(ctx)
	if err != nil {
		return fmt.Errorf("cannot get current database version: %w", err)
	}

	// Get all registered migrations that need to be applied (only pending ones)
	for _, migration := range registeredMigrations {
		// Skip migrations that are already applied
		if migration.version <= currentVersion {
			continue
		}

		if version != goose.MaxVersion && migration.version > version {
			break
		}

		migrationStartTime := time.Now()
		totalMigrations++

		logger.DryRunMigrationStart(filepath.Base(migration.source), migration.version)

		// Create dry-run context for this migration
		dryRunCtx := schema.NewDryRunContext(ctx)

		// Execute the migration in dry-run mode
		if migration.upFnContext != nil {
			err := migration.upFnContext(dryRunCtx)
			if err != nil {
				return fmt.Errorf("dry-run migration %s failed: %w", migration.source, err)
			}

			capturedSQL := dryRunCtx.GetCapturedSQL()
			totalStatements += len(capturedSQL)

			// Print captured SQL
			if dryRunCtx.HasPendingQuery() {
				queries := dryRunCtx.GetPendingQueries()
				for _, q := range queries {
					logger.DryRunSQL(q.Query, q.Args...)
				}
			}
		}

		migrationDuration := time.Since(migrationStartTime).Seconds() * 1000

		logger.DryRunMigrationComplete(filepath.Base(migration.source), migrationDuration)
	}

	duration := time.Since(startTime).Seconds() * 1000

	logger.DryRunSummary(totalMigrations, totalStatements, duration)

	return nil
}
