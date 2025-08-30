package schema

import (
	"context"
	"database/sql"
)

type Context struct {
	ctx      context.Context
	tx       *sql.Tx
	filename string
}

type ContextOptions func(*Context)

func WithFilename(filename string) ContextOptions {
	return func(c *Context) {
		c.filename = filename
	}
}

func NewContext(ctx context.Context, tx *sql.Tx, opts ...ContextOptions) *Context {
	c := &Context{
		ctx: ctx,
		tx:  tx,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Context) Exec(query string, args ...any) (sql.Result, error) {
	return c.tx.ExecContext(c.ctx, query, args...)
}

func (c *Context) Query(query string, args ...any) (*sql.Rows, error) {
	return c.tx.QueryContext(c.ctx, query, args...)
}

func (c *Context) QueryRow(query string, args ...any) *sql.Row {
	return c.tx.QueryRowContext(c.ctx, query, args...)
}
