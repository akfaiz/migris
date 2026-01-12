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

func TestUp(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err, "failed to open in-memory sqlite3 database")
	defer db.Close()

	m, err := migris.New("sqlite3", migris.WithDB(db))
	require.NoError(t, err, "failed to create migris instance")

	migris.AddNamedMigrationContext("20250101000002_create_posts_table.go", func(ctx schema.Context) error {
		return schema.Create(ctx, "posts", func(t *schema.Blueprint) {
			t.Increments("id")
			t.String("title", 255)
		})
	}, func(ctx schema.Context) error {
		return schema.DropIfExists(ctx, "posts")
	})

	// Apply migrations
	err = m.Up()
	require.NoError(t, err, "failed to apply migrations")

	// Verify table exists
	var name string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='posts'").Scan(&name)
	require.NoError(t, err, "failed to find posts table")
	require.Equal(t, "posts", name, "table name should be posts")
}

func TestUp_DryRun(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err, "failed to open in-memory sqlite3 database")
	defer db.Close()

	m, err := migris.New("sqlite3", migris.WithDB(db), migris.WithDryRun(true))
	require.NoError(t, err, "failed to create migris instance")

	migris.AddNamedMigrationContext("20250101000003_insert_data.go", func(ctx schema.Context) error {
		// Insert SQL generated via schema builder should be captured in dry-run
		return schema.Create(ctx, "dryrun_table", func(t *schema.Blueprint) {
			t.Increments("id")
			t.String("data", 255)
		})
	}, func(ctx schema.Context) error {
		return schema.DropIfExists(ctx, "dryrun_table")
	})

	// Capture logger output
	var buf bytes.Buffer
	lg := logger.Get()
	lg.SetOutput(&buf)

	// Run Up in dry-run mode
	err = m.Up()
	require.NoError(t, err, "failed to run Up in dry-run mode")

	out := buf.String()
	require.Contains(t, out, "DRY RUN", "output should contain DRY RUN badge")
	require.Contains(t, out, "dryrun_table", "output should contain the migration/table name or SQL")
}
