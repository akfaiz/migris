package schema_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/akfaiz/migris/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDryRunContext(t *testing.T) {
	ctx := context.Background()
	drc := schema.NewDryRunContext(ctx)

	assert.NotNil(t, drc, "NewDryRunContext should not return nil")
	assert.NotNil(t, drc.GetCapturedSQL(), "capturedSQL should be initialized")
	assert.Empty(t, drc.GetCapturedSQL(), "capturedSQL should be empty initially")
}

func TestDryRunContext_Exec(t *testing.T) {
	drc := schema.NewDryRunContext(context.Background())

	query := "INSERT INTO users (name) VALUES ($1)"
	args := []any{"John"}

	result, err := drc.Exec(query, args...)
	require.NoError(t, err, "Exec should not return error")
	assert.NotNil(t, result, "Exec should not return nil result")

	lastID, err := result.LastInsertId()
	require.NoError(t, err, "LastInsertId should not return error")
	assert.Equal(t, int64(1), lastID, "LastInsertId should be 1")

	rowsAffected, err := result.RowsAffected()
	require.NoError(t, err, "RowsAffected should not return error")
	assert.Equal(t, int64(1), rowsAffected, "RowsAffected should be 1")

	captured := drc.GetCapturedSQL()
	assert.Len(t, captured, 1, "Should have 1 captured query")
	assert.Equal(t, query, captured[0], "Captured query should match")
}

func TestDryRunContext_Query(t *testing.T) {
	drc := schema.NewDryRunContext(context.Background())

	query := "SELECT * FROM users WHERE id = $1"
	args := []any{1}

	rows, err := drc.Query(query, args...) //nolint:rowserrcheck // ignore roserrcheck for test
	require.NoError(t, err, "Query should not return error")
	assert.NotNil(t, rows, "Query should not return nil rows")

	captured := drc.GetCapturedSQL()
	assert.Len(t, captured, 1, "Should have 1 captured query")
	assert.Equal(t, query, captured[0], "Captured query should match")
}

func TestDryRunContext_QueryRow(t *testing.T) {
	drc := schema.NewDryRunContext(context.Background())

	query := "SELECT name FROM users WHERE id = $1"
	args := []any{1}

	row := drc.QueryRow(query, args...)
	assert.NotNil(t, row, "QueryRow should not return nil row")

	captured := drc.GetCapturedSQL()
	assert.Len(t, captured, 1, "Should have 1 captured query")
	assert.Equal(t, query, captured[0], "Captured query should match")
}

func TestDryRunContext_GetCapturedSQL(t *testing.T) {
	drc := schema.NewDryRunContext(context.Background())

	queries := []string{
		"INSERT INTO users (name) VALUES ($1)",
		"UPDATE users SET name = $1 WHERE id = $2",
		"DELETE FROM users WHERE id = $1",
	}

	for _, query := range queries {
		drc.Exec(query, "test")
	}

	captured := drc.GetCapturedSQL()
	assert.Len(t, captured, len(queries), "Should capture all queries")
	assert.Equal(t, queries, captured, "Captured queries should match expected queries")
}

func TestDryRunContext_PendingQueries(t *testing.T) {
	drc := schema.NewDryRunContext(context.Background())

	assert.False(t, drc.HasPendingQuery(), "Should not have pending queries initially")

	query := "INSERT INTO users (name) VALUES ($1)"
	args := []any{"John"}
	drc.Exec(query, args...)

	assert.True(t, drc.HasPendingQuery(), "Should have pending query after Exec")

	queries := drc.GetPendingQueries()
	assert.Len(t, queries, 1, "Should have 1 pending query")
	assert.Equal(t, query, queries[0].Query, "Query should match")
	assert.Len(t, queries[0].Args, len(args), "Args length should match")

	assert.False(t, drc.HasPendingQuery(), "HasPendingQuery should return false after GetPendingQueries consumes them")
}

func TestDryRunContext_GetPendingQuery(t *testing.T) {
	drc := schema.NewDryRunContext(context.Background())

	query1 := "INSERT INTO users (name) VALUES ($1)"
	args1 := []any{"John"}
	drc.Exec(query1, args1...)

	query2 := "UPDATE users SET name = $1"
	args2 := []any{"Jane"}
	drc.Exec(query2, args2...)

	gotQuery, gotArgs := drc.GetPendingQuery()
	assert.Equal(t, query1, gotQuery, "First query should match")
	assert.Len(t, gotArgs, len(args1), "First args length should match")

	gotQuery, _ = drc.GetPendingQuery()
	assert.Equal(t, query2, gotQuery, "Second query should match")

	gotQuery, gotArgs = drc.GetPendingQuery()
	assert.Empty(t, gotQuery, "Should return empty query when no more queries")
	assert.Nil(t, gotArgs, "Should return nil args for empty query")
}

func TestMockResult(t *testing.T) {
	mock := &schema.MockResult{
		LastInsertID:      42,
		RowsAffectedValue: 10,
	}

	lastID, err := mock.LastInsertId()
	require.NoError(t, err, "LastInsertId should not return error")
	assert.Equal(t, int64(42), lastID, "LastInsertId should be 42")

	rowsAffected, err := mock.RowsAffected()
	require.NoError(t, err, "RowsAffected should not return error")
	assert.Equal(t, int64(10), rowsAffected, "RowsAffected should be 10")
}

func TestMockRows(t *testing.T) {
	mock := &schema.MockRows{}

	assert.False(t, mock.Closed, "MockRows should not be closed initially")
	assert.False(t, mock.Next(), "MockRows.Next() should always return false")

	err := mock.Scan()
	require.NoError(t, err, "MockRows.Scan() should return nil")

	columns, err := mock.Columns()
	require.NoError(t, err, "MockRows.Columns() should not return error")
	assert.Empty(t, columns, "MockRows.Columns() should return empty slice")

	err = mock.Err()
	require.NoError(t, err, "MockRows.Err() should return nil")

	mock.Close()
	assert.True(t, mock.Closed, "MockRows should be closed after Close()")
}

func TestMockRow(t *testing.T) {
	mock := &schema.MockRow{}

	err := mock.Scan()
	assert.Equal(t, sql.ErrNoRows, err, "MockRow.Scan() should return sql.ErrNoRows")
}

func TestDryRunContext_WhitespaceHandling(t *testing.T) {
	drc := schema.NewDryRunContext(context.Background())

	query := "  INSERT INTO users (name) VALUES ($1)  \n"
	expected := "INSERT INTO users (name) VALUES ($1)"

	drc.Exec(query)

	captured := drc.GetCapturedSQL()
	assert.Len(t, captured, 1, "Should have 1 captured query")
	assert.Equal(t, expected, captured[0], "Query should be trimmed")
}
