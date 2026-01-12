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

func TestReset(t *testing.T) {
	migris.ResetRegisteredMigrations()
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	m, err := migris.New("sqlite3", migris.WithDB(db))
	require.NoError(t, err)

	migris.AddNamedMigrationContext("20250101000006_create_c.go", func(ctx schema.Context) error {
		return schema.Create(ctx, "c_table", func(t *schema.Blueprint) {
			t.Increments("id")
		})
	}, func(ctx schema.Context) error {
		return schema.DropIfExists(ctx, "c_table")
	})

	// Apply migration
	err = m.Up()
	require.NoError(t, err)

	// Reset (rollback all)
	err = m.Reset()
	require.NoError(t, err)

	// c_table should be gone
	var name string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='c_table'").Scan(&name)
	require.Error(t, err)
}

func TestReset_DryRun(t *testing.T) {
	migris.ResetRegisteredMigrations()
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Apply a migration first
	mApply, err := migris.New("sqlite3", migris.WithDB(db))
	require.NoError(t, err)

	migris.AddNamedMigrationContext("20250101000008_create_e.go", func(ctx schema.Context) error {
		return schema.Create(ctx, "e_table", func(t *schema.Blueprint) {
			t.Increments("id")
		})
	}, func(ctx schema.Context) error {
		return schema.DropIfExists(ctx, "e_table")
	})

	err = mApply.Up()
	require.NoError(t, err)

	// Run Reset in dry-run mode and capture output
	mDry, err := migris.New("sqlite3", migris.WithDB(db), migris.WithDryRun(true))
	require.NoError(t, err)

	var buf bytes.Buffer
	lg := logger.Get()
	lg.SetOutput(&buf)

	err = mDry.Reset()
	require.NoError(t, err)

	out := buf.String()
	require.Contains(t, out, "DRY RUN", "reset dry-run output should contain DRY RUN badge")
	require.Contains(t, out, "e_table", "reset dry-run output should reference the table to rollback")
}
