package migris_test

import (
	"database/sql"
	"testing"

	"github.com/akfaiz/migris"
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
