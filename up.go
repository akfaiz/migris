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
func (m *Migrate) Up(opts ...Option) error {
	ctx := context.Background()
	return m.UpContext(ctx, opts...)
}

// UpContext applies the migrations in the specified directory.
func (m *Migrate) UpContext(ctx context.Context, opts ...Option) error {
	return m.UpToContext(ctx, goose.MaxVersion, opts...)
}

// UpTo applies the migrations up to the specified version.
func (m *Migrate) UpTo(version int64, opts ...Option) error {
	ctx := context.Background()
	return m.UpToContext(ctx, version, opts...)
}

// UpToContext applies the migrations up to the specified version.
func (m *Migrate) UpToContext(ctx context.Context, version int64, opts ...Option) error {
	ro := applyRunOptions(opts)
	if ro.dryRun {
		return m.executeDryRunUp(ctx, version, ro)
	}

	provider, err := m.newProvider(ro)
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

// executeDryRunUp executes migrations in dry-run mode without touching the DB schema.
func (m *Migrate) executeDryRunUp(ctx context.Context, version int64, ro runOptions) error {
	appliedVersions, err := m.queryAppliedVersions(ctx)
	if err != nil {
		return fmt.Errorf("cannot get applied migrations: %w", err)
	}

	migrationsToApply := m.determineMigrationsToApply(version, appliedVersions, ro.allowMissing)
	if len(migrationsToApply) == 0 {
		m.logger.Info("Nothing to migrate.")
		return nil
	}

	m.logger.DryRunStart(version)

	totalMigrations, totalStatements, _, err := m.processDryRunUpMigrations(ctx, migrationsToApply)
	if err != nil {
		return err
	}

	m.logger.DryRunSummary(totalMigrations, totalStatements)
	return nil
}

// determineMigrationsToApply determines which migrations should be applied.
// When allowMissing is true, out-of-order pending migrations are included.
func (m *Migrate) determineMigrationsToApply(
	version int64,
	appliedVersions map[int64]bool,
	allowMissing bool,
) []*Migration {
	currentVersion := maxAppliedVersion(appliedVersions)
	var migrationsToApply []*Migration

	for _, migration := range m.registry.migrationsSnapshot() {
		if version != goose.MaxVersion && migration.version > version {
			break
		}
		if allowMissing {
			if !appliedVersions[migration.version] {
				migrationsToApply = append(migrationsToApply, migration)
			}
		} else {
			if migration.version > currentVersion {
				migrationsToApply = append(migrationsToApply, migration)
			}
		}
	}

	return migrationsToApply
}

func maxAppliedVersion(applied map[int64]bool) int64 {
	var result int64
	for v := range applied {
		result = max(result, v)
	}
	return result
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

		dryRunCtx := schema.NewDryRunContext(ctx, schema.WithDryRunDialect(m.dialect.String()))

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

			queries := dryRunCtx.GetPendingQueries()
			totalStatements += len(queries)
			for _, q := range queries {
				m.logger.DryRunSQL(q.Query, q.Args...)
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
