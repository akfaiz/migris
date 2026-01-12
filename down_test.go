package migris_test

import (
	"bytes"
	"database/sql"
	"testing"

	"github.com/akfaiz/migris"
	"github.com/akfaiz/migris/internal/logger"
	"github.com/akfaiz/migris/schema"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func TestDown(t *testing.T) {
	migris.ResetRegisteredMigrations()
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	m, err := migris.New("sqlite3", migris.WithDB(db))
	require.NoError(t, err)

	// Register two migrations, apply them, then roll back the last one
	migris.AddNamedMigrationContext("20250101000004_create_a.go", func(ctx schema.Context) error {
		return schema.Create(ctx, "a_table", func(t *schema.Blueprint) {
			t.Increments("id")
		})
	}, func(ctx schema.Context) error {
		return schema.DropIfExists(ctx, "a_table")
	})

	migris.AddNamedMigrationContext("20250101000005_create_b.go", func(ctx schema.Context) error {
		return schema.Create(ctx, "b_table", func(t *schema.Blueprint) {
			t.Increments("id")
		})
	}, func(ctx schema.Context) error {
		return schema.DropIfExists(ctx, "b_table")
	})

	// Apply both migrations
	err = m.Up()
	require.NoError(t, err)

	// Ensure both tables exist
	var name string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='a_table'").Scan(&name)
	require.NoError(t, err)
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='b_table'").Scan(&name)
	require.NoError(t, err)

	// Rollback last migration (b_table)
	err = m.Down()
	require.NoError(t, err)

	// b_table should be gone, a_table should still exist
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='b_table'").Scan(&name)
	require.Error(t, err)
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='a_table'").Scan(&name)
	require.NoError(t, err)
}

func TestDown_DryRun(t *testing.T) {
	migris.ResetRegisteredMigrations()
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// First apply a migration using non-dry-run migrator
	mApply, err := migris.New("sqlite3", migris.WithDB(db))
	require.NoError(t, err)

	migris.AddNamedMigrationContext("20250101000007_create_d.go", func(ctx schema.Context) error {
		return schema.Create(ctx, "d_table", func(t *schema.Blueprint) {
			t.Increments("id")
		})
	}, func(ctx schema.Context) error {
		return schema.DropIfExists(ctx, "d_table")
	})

	err = mApply.Up()
	require.NoError(t, err)

	// Now run Down in dry-run mode and capture logger output
	mDry, err := migris.New("sqlite3", migris.WithDB(db), migris.WithDryRun(true))
	require.NoError(t, err)

	var buf bytes.Buffer
	lg := logger.Get()
	lg.SetOutput(&buf)

	err = mDry.Down()
	require.NoError(t, err)

	out := buf.String()
	require.Contains(t, out, "DRY RUN", "dry-run output should contain DRY RUN badge")
	require.Contains(t, out, "d_table", "dry-run output should reference the table to rollback")
}

func TestDownTo(t *testing.T) {
	migris.ResetRegisteredMigrations()
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	m, err := migris.New("sqlite3", migris.WithDB(db))
	require.NoError(t, err)

	// Register three migrations
	migris.AddNamedMigrationContext("20250101000030_create_t1.go", func(ctx schema.Context) error {
		return schema.Create(ctx, "t1", func(t *schema.Blueprint) { t.Increments("id") })
	}, func(ctx schema.Context) error { return schema.DropIfExists(ctx, "t1") })

	migris.AddNamedMigrationContext("20250101000031_create_t2.go", func(ctx schema.Context) error {
		return schema.Create(ctx, "t2", func(t *schema.Blueprint) { t.Increments("id") })
	}, func(ctx schema.Context) error { return schema.DropIfExists(ctx, "t2") })

	migris.AddNamedMigrationContext("20250101000032_create_t3.go", func(ctx schema.Context) error {
		return schema.Create(ctx, "t3", func(t *schema.Blueprint) { t.Increments("id") })
	}, func(ctx schema.Context) error { return schema.DropIfExists(ctx, "t3") })

	// Apply all migrations
	err = m.Up()
	require.NoError(t, err)

	// Now rollback to the first version (keep t1 only)
	err = m.DownTo(20250101000030)
	require.NoError(t, err)

	// t1 should exist, t2 and t3 should be removed
	var name string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='t1'").Scan(&name)
	require.NoError(t, err)
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='t2'").Scan(&name)
	require.Error(t, err)
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='t3'").Scan(&name)
	require.Error(t, err)
}

func TestDownTo_DryRun(t *testing.T) {
	migris.ResetRegisteredMigrations()
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Apply migrations normally first
	mApply, err := migris.New("sqlite3", migris.WithDB(db))
	require.NoError(t, err)

	migris.AddNamedMigrationContext("20250101000033_create_dt1.go", func(ctx schema.Context) error {
		return schema.Create(ctx, "dt1", func(t *schema.Blueprint) { t.Increments("id") })
	}, func(ctx schema.Context) error { return schema.DropIfExists(ctx, "dt1") })

	migris.AddNamedMigrationContext("20250101000034_create_dt2.go", func(ctx schema.Context) error {
		return schema.Create(ctx, "dt2", func(t *schema.Blueprint) { t.Increments("id") })
	}, func(ctx schema.Context) error { return schema.DropIfExists(ctx, "dt2") })

	err = mApply.Up()
	require.NoError(t, err)

	// Run DownTo in dry-run mode to rollback to dt1 only
	mDry, err := migris.New("sqlite3", migris.WithDB(db), migris.WithDryRun(true))
	require.NoError(t, err)

	var buf bytes.Buffer
	lg := logger.Get()
	lg.SetOutput(&buf)

	err = mDry.DownTo(20250101000033)
	require.NoError(t, err)

	out := buf.String()
	require.Contains(t, out, "DRY RUN", "dry-run output should contain DRY RUN badge")
	require.Contains(t, out, "dt2", "dry-run output should reference dt2 as rolled back")
	require.NotContains(t, out, "dt1", "dry-run output should not rollback dt1 since target is dt1")
}
