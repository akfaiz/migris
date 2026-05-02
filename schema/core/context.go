package core

import (
	"context"
	"database/sql"
)

// Context interface defines the contract for database operations.
type Context interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (Rows, error)
	QueryRow(query string, args ...any) Row
	Dialect() string
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

// Column represents a database column with its properties.
type Column struct {
	Name       string         // Name is the name of the column.
	TypeName   string         // TypeName is the name of the column type (e.g., "VARCHAR", "INT").
	TypeFull   string         // TypeFull is the full type name including any modifiers (e.g., "VARCHAR(255)", "INT(11)").
	Collation  sql.NullString // Collation is the collation of the column, if applicable.
	Nullable   bool           // Nullable indicates whether the column can contain NULL values.
	DefaultVal sql.NullString // DefaultVal is the default value for the column, if any.
	Comment    sql.NullString // Comment is an optional comment for the column.
	Extra      sql.NullString // Extra contains additional information about the column (e.g., "auto_increment").
}

// Index represents a database index with its properties.
type Index struct {
	Name    string   // Name is the name of the index.
	Columns []string // Columns is a slice of column names that are part of the index.
	Type    string   // e.g., "btree", "hash"
	Unique  bool     // Indicates if the index is unique
	Primary bool     // Indicates if the index is a primary key
}

// TableInfo represents information about a database table.
type TableInfo struct {
	Name      string         // Name is the name of the table.
	Schema    string         // Schema is the schema where the table resides.
	Size      int64          // Size is the size of the table in bytes.
	Comment   sql.NullString // Comment is an optional comment for the table.
	Engine    sql.NullString // Engine is the storage engine used for the table (e.g., "InnoDB", "MyISAM").
	Collation sql.NullString // Collation is the collation used for the table (e.g., "utf8mb4_general_ci").
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
