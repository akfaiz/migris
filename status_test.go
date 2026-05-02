package migris_test

import (
	"database/sql"
	"testing"

	"github.com/akfaiz/migris"
	"github.com/akfaiz/migris/schema"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func TestStatus(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err, "failed to open in-memory sqlite3 database")
	defer db.Close()

	// Initialize migris
	m, err := migris.New("sqlite3", migris.WithDB(db))
	require.NoError(t, err, "failed to create migris instance")

	// Register a sample migration
	migris.AddNamedMigrationContext("1_create_users_table.go",
		func(ctx schema.Context) error {
			return schema.Create(ctx, "users", func(t *schema.Blueprint) {
				t.Increments("id")
				t.String("name", 255)
			})
		}, func(ctx schema.Context) error {
			return schema.DropIfExists(ctx, "users")
		},
	)

	// Apply the migration
	err = m.Up()
	require.NoError(t, err, "failed to apply migration")

	// Check status
	err = m.Status()
	require.NoError(t, err, "failed to get migration status")
}
