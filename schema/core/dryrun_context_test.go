package core_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/akfaiz/migris/schema/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDryRunContext(t *testing.T) {
	ctx := context.Background()
	drc := core.NewDryRunContext(ctx)

	assert.NotNil(t, drc, "NewDryRunContext should not return nil")
	assert.NotNil(t, drc.GetCapturedSQL(), "capturedSQL should be initialized")
	assert.Empty(t, drc.GetCapturedSQL(), "capturedSQL should be empty initially")
}

func TestDryRunContext_Exec(t *testing.T) {
	drc := core.NewDryRunContext(context.Background())

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

func TestMockResult(t *testing.T) {
	mock := &core.MockResult{
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
	mock := &core.MockRows{}

	assert.False(t, mock.Closed, "MockRows should not be closed initially")
	assert.False(t, mock.Next(), "MockRows.Next() should always return false")

	err := mock.Scan()
	require.NoError(t, err, "MockRows.Scan() should return nil")

	columns, err := mock.Columns()
	require.NoError(t, err, "MockRows.Columns() should not return error")
	assert.Empty(t, columns, "MockRows.Columns() should return empty slice")

	err = mock.Err()
	require.NoError(t, err, "MockRows.Err() should return nil")

	err = mock.Close()
	require.NoError(t, err, "MockRows.Close() should return nil")
	assert.True(t, mock.Closed, "MockRows should be closed after Close()")
}

func TestMockRow(t *testing.T) {
	mock := &core.MockRow{}

	err := mock.Scan()
	assert.Equal(t, sql.ErrNoRows, err, "MockRow.Scan() should return sql.ErrNoRows")
}
