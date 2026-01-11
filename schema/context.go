package schema

import (
	"context"
	"database/sql"
)

// Context interface defines the contract for database operations
// This allows us to switch between normal execution and dry-run mode.
type Context interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

// RegularContext implements Context for normal database operations.
type RegularContext struct {
	ctx      context.Context
	tx       *sql.Tx
	filename string
}

type ContextOptions func(*RegularContext)

func WithFilename(filename string) ContextOptions {
	return func(c *RegularContext) {
		c.filename = filename
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

func (c *RegularContext) Query(query string, args ...any) (*sql.Rows, error) {
	return c.tx.QueryContext(c.ctx, query, args...)
}

func (c *RegularContext) QueryRow(query string, args ...any) *sql.Row {
	return c.tx.QueryRowContext(c.ctx, query, args...)
}
