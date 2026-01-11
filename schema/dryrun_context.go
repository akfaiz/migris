package schema

import (
	"context"
	"database/sql"
	"strings"
)

// DryRunContext implements Context for dry-run mode (captures SQL without executing)
type DryRunContext struct {
	ctx            context.Context
	capturedSQL    []string
	config         DryRunConfig
	pendingQueries []QueryWithArgs
}

// QueryWithArgs stores a query and its arguments
type QueryWithArgs struct {
	Query string
	Args  []any
}

// DryRunConfig holds configuration for dry-run mode
type DryRunConfig struct {
	PrintSQL        bool // Print SQL statements as they're captured
	PrintMigrations bool // Print migration info
}

// MockResult implements sql.Result for dry-run mode
type MockResult struct {
	lastInsertID int64
	rowsAffected int64
}

func (m *MockResult) LastInsertId() (int64, error) {
	return m.lastInsertID, nil
}

func (m *MockResult) RowsAffected() (int64, error) {
	return m.rowsAffected, nil
}

// MockRows implements basic sql.Rows functionality for dry-run mode
type MockRows struct {
	closed bool
}

func (m *MockRows) Close() error {
	m.closed = true
	return nil
}

func (m *MockRows) Next() bool {
	return false // No rows in dry-run mode
}

func (m *MockRows) Scan(dest ...interface{}) error {
	return nil // No data to scan in dry-run mode
}

func (m *MockRows) Columns() ([]string, error) {
	return []string{}, nil
}

func (m *MockRows) Err() error {
	return nil
}

// MockRow implements basic sql.Row functionality for dry-run mode
type MockRow struct{}

func (m *MockRow) Scan(dest ...interface{}) error {
	return sql.ErrNoRows // Always return no rows in dry-run mode
}

// NewDryRunContext creates a new DryRunContext
func NewDryRunContext(ctx context.Context, config DryRunConfig) *DryRunContext {
	return &DryRunContext{
		ctx:         ctx,
		capturedSQL: make([]string, 0),
		config:      config,
	}
}

func (drc *DryRunContext) Exec(query string, args ...any) (sql.Result, error) {
	// Clean up the query for display
	cleanQuery := strings.TrimSpace(query)
	drc.capturedSQL = append(drc.capturedSQL, cleanQuery)

	if drc.config.PrintSQL {
		drc.printSQL(cleanQuery, args...)
	}

	// Return a mock result that simulates successful execution
	return &MockResult{
		lastInsertID: 1,
		rowsAffected: 1,
	}, nil
}

func (drc *DryRunContext) Query(query string, args ...any) (*sql.Rows, error) {
	// Capture the query but don't execute
	cleanQuery := strings.TrimSpace(query)
	drc.capturedSQL = append(drc.capturedSQL, cleanQuery)

	if drc.config.PrintSQL {
		drc.printSQL(cleanQuery, args...)
	}

	// Return empty mock rows
	return &sql.Rows{}, nil
}

func (drc *DryRunContext) QueryRow(query string, args ...any) *sql.Row {
	// Capture the query but don't execute
	cleanQuery := strings.TrimSpace(query)
	drc.capturedSQL = append(drc.capturedSQL, cleanQuery)

	if drc.config.PrintSQL {
		drc.printSQL(cleanQuery, args...)
	}

	// Return a row that will return sql.ErrNoRows when scanned
	return &sql.Row{}
}

// GetCapturedSQL returns all captured SQL statements
func (drc *DryRunContext) GetCapturedSQL() []string {
	return drc.capturedSQL
}

// HasPendingQuery returns true if there's a pending SQL query to be printed
func (drc *DryRunContext) HasPendingQuery() bool {
	return len(drc.pendingQueries) > 0
}

// GetPendingQueries returns all pending queries and clears them
func (drc *DryRunContext) GetPendingQueries() []QueryWithArgs {
	queries := drc.pendingQueries
	drc.pendingQueries = nil
	return queries
}

// GetPendingQuery returns the first pending query and arguments (for backward compatibility)
func (drc *DryRunContext) GetPendingQuery() (string, []any) {
	if len(drc.pendingQueries) == 0 {
		return "", nil
	}
	first := drc.pendingQueries[0]
	drc.pendingQueries = drc.pendingQueries[1:]
	return first.Query, first.Args
}

// printSQL stores the query for later printing by logger
func (drc *DryRunContext) printSQL(query string, args ...any) {
	// Store query and args for later printing
	drc.pendingQueries = append(drc.pendingQueries, QueryWithArgs{
		Query: query,
		Args:  args,
	})
}
