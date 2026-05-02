package core_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/akfaiz/migris/schema/core"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegularContext(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer tx.Rollback()

	c := core.NewContext(ctx, tx, core.WithDialect("sqlite3"), core.WithFilename("test.go"))

	t.Run("Exec", func(t *testing.T) {
		res, err := c.Exec("CREATE TABLE test (id INTEGER)")
		require.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("Query", func(t *testing.T) {
		_, err := c.Exec("INSERT INTO test (id) VALUES (1)")
		require.NoError(t, err)

		rows, err := c.Query("SELECT id FROM test")
		require.NoError(t, err)
		defer rows.Close()

		assert.True(t, rows.Next())
		var id int
		err = rows.Scan(&id)
		require.NoError(t, err)
		assert.Equal(t, 1, id)
	})

	t.Run("QueryRow", func(t *testing.T) {
		row := c.QueryRow("SELECT id FROM test WHERE id = 1")
		var id int
		err := row.Scan(&id)
		require.NoError(t, err)
		assert.Equal(t, 1, id)
	})

	t.Run("Dialect", func(t *testing.T) {
		assert.Equal(t, "sqlite3", c.Dialect())
	})
}

func TestDryRunContext_More(t *testing.T) {
	ctx := context.Background()
	drc := core.NewDryRunContext(ctx, core.WithDryRunDialect("postgres"))

	t.Run("Dialect", func(t *testing.T) {
		assert.Equal(t, "postgres", drc.Dialect())
	})

	t.Run("Exec", func(t *testing.T) {
		res, err := drc.Exec("CREATE TABLE users")
		require.NoError(t, err)
		lastID, _ := res.LastInsertId()
		rowsAff, _ := res.RowsAffected()
		assert.Equal(t, int64(1), lastID)
		assert.Equal(t, int64(1), rowsAff)
		assert.Contains(t, drc.GetCapturedSQL(), "CREATE TABLE users")
	})

	t.Run("Query", func(t *testing.T) {
		rows, err := drc.Query("SELECT * FROM users")
		require.NoError(t, err)
		assert.False(t, rows.Next())
		require.NoError(t, rows.Err())
		cols, _ := rows.Columns()
		assert.Empty(t, cols)
		assert.NoError(t, rows.Close())
	})

	t.Run("QueryRow", func(t *testing.T) {
		row := drc.QueryRow("SELECT 1")
		err := row.Scan()
		assert.ErrorIs(t, err, sql.ErrNoRows)
	})

	t.Run("PendingQueries", func(t *testing.T) {
		assert.True(t, drc.HasPendingQuery())
		queries := drc.GetPendingQueries()
		assert.Len(t, queries, 3)
		assert.False(t, drc.HasPendingQuery())
	})

	t.Run("GetPendingQuery", func(t *testing.T) {
		drc.Exec("DROP TABLE users", 1, 2)
		query, args := drc.GetPendingQuery()
		assert.Equal(t, "DROP TABLE users", query)
		assert.Equal(t, []any{1, 2}, args)

		q2, a2 := drc.GetPendingQuery()
		assert.Empty(t, q2)
		assert.Nil(t, a2)
	})
}
