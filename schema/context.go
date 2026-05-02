package schema

import (
	"context"
	"database/sql"
)

// Context interface defines the contract for database operations
// This allows us to switch between normal execution and dry-run mode.
type Context interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (Rows, error)
	QueryRow(query string, args ...any) Row
}

// Rows interface defines the methods required for database rows.
type Rows interface {
	Next() bool
	Scan(dest ...any) error
	Close() error
	Err() error
	Columns() ([]string, error)
}

// Row interface defines the method required for a single database row.
type Row interface {
	Scan(dest ...any) error
}

// RegularContext implements Context for normal database operations.
type RegularContext struct {
	ctx      context.Context
	tx       *sql.Tx
	filename string
	dialect  string
}

type ContextOptions func(*RegularContext)

func WithFilename(filename string) ContextOptions {
	return func(c *RegularContext) {
		c.filename = filename
	}
}

func WithDialect(dialect string) ContextOptions {
	return func(c *RegularContext) {
		c.dialect = dialect
	}
}

func NewContext(ctx context.Context, tx *sql.Tx, opts ...ContextOptions) Context {
	c := &RegularContext{
		ctx: ctx,
		tx:  tx,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *RegularContext) Exec(query string, args ...any) (sql.Result, error) {
	return c.tx.ExecContext(c.ctx, query, args...)
}

func (c *RegularContext) Query(query string, args ...any) (Rows, error) {
	rows, err := c.tx.QueryContext(c.ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = rows.Close()
		}
	}()
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return rows, nil
}

func (c *RegularContext) QueryRow(query string, args ...any) Row {
	return c.tx.QueryRowContext(c.ctx, query, args...)
}

func (c *RegularContext) Dialect() string {
	return c.dialect
}
