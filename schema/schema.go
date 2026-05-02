package schema

import (
	"context"
	"database/sql"
	"errors"

	"github.com/akfaiz/migris/internal/config"
	"github.com/akfaiz/migris/internal/dialect"
	"github.com/akfaiz/migris/schema/blueprint"
	"github.com/akfaiz/migris/schema/builders"
	"github.com/akfaiz/migris/schema/core"
	"github.com/akfaiz/migris/schema/grammars"
)

// Type aliases for the public API.

type (
	Blueprint            = blueprint.Blueprint
	ColumnDefinition     = blueprint.ColumnDefinition
	IndexDefinition      = blueprint.IndexDefinition
	ForeignKeyDefinition = blueprint.ForeignKeyDefinition
	Context              = core.Context
	Rows                 = core.Rows
	Row                  = core.Row
	Builder              = builders.Builder
	Expression           = blueprint.Expression
	Column               = core.Column
	Index                = core.Index
	TableInfo            = core.TableInfo
)

func newBuilder(c Context) (Builder, error) {
	dialectVal := dialect.Unknown
	if c != nil {
		dialectVal = dialect.FromString(c.Dialect())
	}
	if dialectVal == dialect.Unknown {
		dialectVal = config.GetDialect()
	}
	if dialectVal == dialect.Unknown {
		return nil, errors.New("schema dialect is not set")
	}

	return builders.NewBuilder(dialectVal.String())
}

// Create creates a new table with the given name and blueprint.
//
// Example:
//
//	schema.Create(ctx, "users", func(table *schema.Blueprint) {
//	    table.String("name")
//	    table.Integer("age")
//	})
func Create(c Context, name string, bp func(table *Blueprint)) error {
	if c == nil || name == "" || bp == nil {
		return errors.New("invalid arguments")
	}
	builder, err := newBuilder(c)
	if err != nil {
		return err
	}
	return builder.Create(c, name, bp)
}

// Drop drops the table with the given name.
//
// Example:
//
//	schema.Drop(ctx, "users")
func Drop(c Context, name string) error {
	if c == nil || name == "" {
		return errors.New("invalid arguments")
	}
	builder, err := newBuilder(c)
	if err != nil {
		return err
	}
	return builder.Drop(c, name)
}

// DropIfExists drops the table with the given name if it exists.
//
// Example:
//
//	schema.DropIfExists(ctx, "users")
func DropIfExists(c Context, name string) error {
	if c == nil || name == "" {
		return errors.New("invalid arguments")
	}
	builder, err := newBuilder(c)
	if err != nil {
		return err
	}
	return builder.DropIfExists(c, name)
}

// Table modifies the table with the given name using the provided blueprint.
//
// Example:
//
//	schema.Table(ctx, "users", func(table *schema.Blueprint) {
//	    table.String("email").Unique()
//	})
func Table(c Context, name string, bp func(table *Blueprint)) error {
	if c == nil || name == "" || bp == nil {
		return errors.New("invalid arguments")
	}
	builder, err := newBuilder(c)
	if err != nil {
		return err
	}
	return builder.Table(c, name, bp)
}

// Rename renames a table from the given name to the new name.
//
// Example:
//
//	schema.Rename(ctx, "users", "app_users")
func Rename(c Context, from, to string) error {
	if c == nil || from == "" || to == "" {
		return errors.New("invalid arguments")
	}
	builder, err := newBuilder(c)
	if err != nil {
		return err
	}
	return builder.Rename(c, from, to)
}

// HasTable checks if a table with the given name exists.
//
// Example:
//
//	exists, err := schema.HasTable(ctx, "users")
func HasTable(c Context, name string) (bool, error) {
	if c == nil || name == "" {
		return false, errors.New("invalid arguments")
	}
	builder, err := newBuilder(c)
	if err != nil {
		return false, err
	}
	return builder.HasTable(c, name)
}

// HasColumn checks if a column with the given name exists in the specified table.
//
// Example:
//
//	exists, err := schema.HasColumn(ctx, "users", "email")
func HasColumn(c Context, table, column string) (bool, error) {
	if c == nil || table == "" || column == "" {
		return false, errors.New("invalid arguments")
	}
	builder, err := newBuilder(c)
	if err != nil {
		return false, err
	}
	return builder.HasColumn(c, table, column)
}

// HasColumns checks if all specified columns exist in the given table.
//
// Example:
//
//	exists, err := schema.HasColumns(ctx, "users", []string{"email", "name"})
func HasColumns(c Context, table string, columns []string) (bool, error) {
	if c == nil || table == "" || len(columns) == 0 {
		return false, errors.New("invalid arguments")
	}
	builder, err := newBuilder(c)
	if err != nil {
		return false, err
	}
	return builder.HasColumns(c, table, columns)
}

// HasIndex checks if an index exists on the specified table for the given columns.
//
// Example:
//
//	exists, err := schema.HasIndex(ctx, "users", []string{"email"})
func HasIndex(c Context, table string, columns []string) (bool, error) {
	if c == nil || table == "" || len(columns) == 0 {
		return false, errors.New("invalid arguments")
	}
	builder, err := newBuilder(c)
	if err != nil {
		return false, err
	}
	return builder.HasIndex(c, table, columns)
}

// GetColumns retrieves the columns of the specified table.
//
// Example:
//
//	columns, err := schema.GetColumns(ctx, "users")
func GetColumns(c Context, table string) ([]*Column, error) {
	if c == nil || table == "" {
		return nil, errors.New("invalid arguments")
	}
	builder, err := newBuilder(c)
	if err != nil {
		return nil, err
	}
	return builder.GetColumns(c, table)
}

// GetIndexes retrieves the indexes of the specified table.
//
// Example:
//
//	indexes, err := schema.GetIndexes(ctx, "users")
func GetIndexes(c Context, table string) ([]*Index, error) {
	if c == nil || table == "" {
		return nil, errors.New("invalid arguments")
	}
	builder, err := newBuilder(c)
	if err != nil {
		return nil, err
	}
	return builder.GetIndexes(c, table)
}

// GetTables retrieves the list of tables in the database.
//
// Example:
//
//	tables, err := schema.GetTables(ctx)
func GetTables(c Context) ([]*TableInfo, error) {
	if c == nil {
		return nil, errors.New("invalid arguments")
	}
	builder, err := newBuilder(c)
	if err != nil {
		return nil, err
	}
	return builder.GetTables(c)
}

// NewBuilder creates a new Builder instance for the specified dialect.
func NewBuilder(dialectValue string) (Builder, error) {
	return builders.NewBuilder(dialectValue)
}

// NewGrammar creates a new Grammar instance for the specified dialect.
func NewGrammar(dialectValue string) (blueprint.Grammar, error) {
	return grammars.NewGrammar(dialectValue)
}

// NewContext creates a new Context with the given base context, transaction, and options.
func NewContext(ctx context.Context, tx *sql.Tx, opts ...core.ContextOptions) Context {
	return core.NewContext(ctx, tx, opts...)
}

// NewDryRunContext creates a new DryRunContext with the given base context and options.
func NewDryRunContext(ctx context.Context, opts ...core.DryRunContextOptions) *core.DryRunContext {
	return core.NewDryRunContext(ctx, opts...)
}

// WithFilename returns a ContextOption that sets the filename for the context.
func WithFilename(filename string) core.ContextOptions {
	return core.WithFilename(filename)
}

// WithDialect returns a ContextOption that sets the dialect for the context.
func WithDialect(dialect string) core.ContextOptions {
	return core.WithDialect(dialect)
}

// WithDryRunDialect returns a DryRunContextOption that sets the dialect for the dry run context.
func WithDryRunDialect(dialect string) core.DryRunContextOptions {
	return core.WithDryRunDialect(dialect)
}

// NewBlueprintForTesting creates a new Blueprint instance for testing purposes with the given name and grammar.
func NewBlueprintForTesting(name string, g blueprint.Grammar) *Blueprint {
	return blueprint.NewBlueprintForTesting(name, g)
}
