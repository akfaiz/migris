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
