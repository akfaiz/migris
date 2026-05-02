package migris_test

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/akfaiz/migris"
	"github.com/akfaiz/migris/schema"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func TestNew_ValidOptions(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	m, err := migris.New("sqlite3",
		migris.WithDB(db),
		migris.WithMigrationDir("migrations_test_dir"),
		migris.WithTableName("test_schema_migrations"),
		migris.WithDryRun(true),
	)
	require.NoError(t, err)
	require.NotNil(t, m)
}

func TestNew_DBNotSet(t *testing.T) {
	m, err := migris.New("sqlite3")
	require.Error(t, err)
	require.Nil(t, m)
}

func TestNew_UnknownDialect(t *testing.T) {
	m, err := migris.New("not-a-dialect")
	require.Error(t, err)
	require.Nil(t, m)
}

func TestMigrator_WithRegistryUsesIsolatedMigrations(t *testing.T) {
	migris.ResetRegisteredMigrations()
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	migris.AddNamedMigrationContext("20250101000100_create_global_table.go", func(ctx schema.Context) error {
		return schema.Create(ctx, "global_table", func(t *schema.Blueprint) {
			t.Increments("id")
		})
	}, func(ctx schema.Context) error {
		return schema.DropIfExists(ctx, "global_table")
	})

	registry := migris.NewRegistry()
	require.NoError(
		t,
		registry.AddNamedMigrationContext("20250101000101_create_registry_table.go", func(ctx schema.Context) error {
			return schema.Create(ctx, "registry_table", func(t *schema.Blueprint) {
				t.Increments("id")
			})
		}, func(ctx schema.Context) error {
			return schema.DropIfExists(ctx, "registry_table")
		}),
	)

	m, err := migris.New("sqlite3", migris.WithDB(db), migris.WithRegistry(registry))
	require.NoError(t, err)
	require.NoError(t, m.Up())

	var name string
	require.NoError(
		t,
		db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='registry_table'").Scan(&name),
	)
	require.Equal(t, "registry_table", name)
	require.Error(
		t,
		db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='global_table'").Scan(&name),
	)
}

func TestMigrator_UsesInstanceDialectDuringMigrationExecution(t *testing.T) {
	migris.ResetRegisteredMigrations()
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	registry := migris.NewRegistry()
	require.NoError(
		t,
		registry.AddNamedMigrationContext("20250101000102_create_context_table.go", func(ctx schema.Context) error {
			dialectContext, ok := ctx.(interface{ Dialect() string })
			if !ok {
				return errors.New("expected migration context to expose dialect")
			}
			if dialectContext.Dialect() != "sqlite3" {
				return fmt.Errorf("expected sqlite3 dialect, got %q", dialectContext.Dialect())
			}
			return schema.Create(ctx, "context_table", func(t *schema.Blueprint) {
				t.Increments("id")
			})
		}, func(ctx schema.Context) error {
			return schema.DropIfExists(ctx, "context_table")
		}),
	)

	mSQLite, err := migris.New("sqlite3", migris.WithDB(db), migris.WithRegistry(registry))
	require.NoError(t, err)

	// Constructing a second migrator must not affect the first migrator's schema context.
	_, err = migris.New("postgres", migris.WithDB(db))
	require.NoError(t, err)

	require.NoError(t, mSQLite.Up())
}
