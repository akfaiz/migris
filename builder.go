package schema

import (
	"context"
	"database/sql"
	"errors"

	"github.com/afkdevs/go-schema/internal/dialect"
)

// Builder is an interface that defines methods for creating, dropping, and managing database tables.
type Builder interface {
	// Create creates a new table with the given name and applies the provided blueprint.
	Create(ctx context.Context, tx *sql.Tx, name string, blueprint func(table *Blueprint)) error
	// Drop removes the table with the given name.
	Drop(ctx context.Context, tx *sql.Tx, name string) error
	// DropIfExists removes the table with the given name if it exists.
	DropIfExists(ctx context.Context, tx *sql.Tx, name string) error
	// GetColumns retrieves the columns of the specified table.
	GetColumns(ctx context.Context, tx *sql.Tx, tableName string) ([]*Column, error)
	// GetIndexes retrieves the indexes of the specified table.
	GetIndexes(ctx context.Context, tx *sql.Tx, tableName string) ([]*Index, error)
	// GetTables retrieves all tables in the database.
	GetTables(ctx context.Context, tx *sql.Tx) ([]*TableInfo, error)
	// HasColumn checks if the specified table has the given column.
	HasColumn(ctx context.Context, tx *sql.Tx, tableName string, columnName string) (bool, error)
	// HasColumns checks if the specified table has all the given columns.
	HasColumns(ctx context.Context, tx *sql.Tx, tableName string, columnNames []string) (bool, error)
	// HasIndex checks if the specified table has the given index.
	HasIndex(ctx context.Context, tx *sql.Tx, tableName string, indexes []string) (bool, error)
	// HasTable checks if a table with the given name exists.
	HasTable(ctx context.Context, tx *sql.Tx, name string) (bool, error)
	// Rename renames a table from oldName to newName.
	Rename(ctx context.Context, tx *sql.Tx, oldName string, newName string) error
	// Table applies the provided blueprint to the specified table.
	Table(ctx context.Context, tx *sql.Tx, name string, blueprint func(table *Blueprint)) error
}

// NewBuilder creates a new Builder instance based on the specified dialect.
// It returns an error if the dialect is not supported.
//
// Supported dialects are "postgres", "pgx", "mysql", and "mariadb".
func NewBuilder(dialectValue string) (Builder, error) {
	dialectVal := dialect.FromString(dialectValue)
	switch dialectVal {
	case dialect.MySQL:
		return newMysqlBuilder(), nil
	case dialect.Postgres:
		return newPostgresBuilder(), nil
	default:
		return nil, errors.New("unsupported dialect: " + dialectValue)
	}
}

type baseBuilder struct {
	grammar grammar
	verbose bool
}

func (b *baseBuilder) newBlueprint(name string) *Blueprint {
	return &Blueprint{name: name, grammar: b.grammar, verbose: b.verbose}
}

func (b *baseBuilder) Create(ctx context.Context, tx *sql.Tx, name string, blueprint func(table *Blueprint)) error {
	if tx == nil || name == "" || blueprint == nil {
		return errors.New("invalid arguments: transaction, name, or blueprint is nil/empty")
	}

	bp := b.newBlueprint(name)
	bp.create()
	blueprint(bp)

	if err := bp.build(ctx, tx); err != nil {
		return err
	}

	return nil
}

func (b *baseBuilder) Drop(ctx context.Context, tx *sql.Tx, name string) error {
	if tx == nil || name == "" {
		return errors.New("invalid arguments: transaction is nil or name is empty")
	}

	bp := b.newBlueprint(name)
	bp.drop()

	if err := bp.build(ctx, tx); err != nil {
		return err
	}

	return nil
}

func (b *baseBuilder) DropIfExists(ctx context.Context, tx *sql.Tx, name string) error {
	if tx == nil || name == "" {
		return errors.New("invalid arguments: transaction is nil or name is empty")
	}

	bp := b.newBlueprint(name)
	bp.dropIfExists()

	if err := bp.build(ctx, tx); err != nil {
		return err
	}

	return nil
}

func (b *baseBuilder) Rename(ctx context.Context, tx *sql.Tx, oldName string, newName string) error {
	if tx == nil || oldName == "" || newName == "" {
		return errors.New("invalid arguments: transaction is nil or old/new table name is empty")
	}

	bp := b.newBlueprint(oldName)
	bp.rename(newName)

	if err := bp.build(ctx, tx); err != nil {
		return err
	}

	return nil
}

func (b *baseBuilder) Table(ctx context.Context, tx *sql.Tx, name string, blueprint func(table *Blueprint)) error {
	if tx == nil || name == "" || blueprint == nil {
		return errors.New("invalid arguments: transaction is nil or name/blueprint is empty")
	}

	bp := b.newBlueprint(name)
	blueprint(bp)

	if err := bp.build(ctx, tx); err != nil {
		return err
	}

	return nil
}
