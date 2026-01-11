package schema

import (
	"errors"

	"github.com/akfaiz/migris/internal/dialect"
)

// Builder is an interface that defines methods for creating, dropping, and managing database tables.
type Builder interface {
	// Create creates a new table with the given name and applies the provided blueprint.
	Create(c Context, name string, blueprint func(table *Blueprint)) error
	// Drop removes the table with the given name.
	Drop(c Context, name string) error
	// DropIfExists removes the table with the given name if it exists.
	DropIfExists(c Context, name string) error
	// GetColumns retrieves the columns of the specified table.
	GetColumns(c Context, tableName string) ([]*Column, error)
	// GetIndexes retrieves the indexes of the specified table.
	GetIndexes(c Context, tableName string) ([]*Index, error)
	// GetTables retrieves all tables in the database.
	GetTables(c Context) ([]*TableInfo, error)
	// HasColumn checks if the specified table has the given column.
	HasColumn(c Context, tableName string, columnName string) (bool, error)
	// HasColumns checks if the specified table has all the given columns.
	HasColumns(c Context, tableName string, columnNames []string) (bool, error)
	// HasIndex checks if the specified table has the given index.
	HasIndex(c Context, tableName string, indexes []string) (bool, error)
	// HasTable checks if a table with the given name exists.
	HasTable(c Context, name string) (bool, error)
	// Rename renames a table from oldName to newName.
	Rename(c Context, oldName string, newName string) error
	// Table applies the provided blueprint to the specified table.
	Table(c Context, name string, blueprint func(table *Blueprint)) error
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
	case dialect.Unknown:
		return nil, errors.New("unsupported dialect: " + dialectValue)
	default:
		return nil, errors.New("unsupported dialect: " + dialectValue)
	}
}

type baseBuilder struct {
	grammar grammar
}

func (b *baseBuilder) newBlueprint(name string) *Blueprint {
	return &Blueprint{name: name, grammar: b.grammar}
}

func (b *baseBuilder) Create(c Context, name string, blueprint func(table *Blueprint)) error {
	if c == nil || name == "" || blueprint == nil {
		return errors.New("invalid arguments: context, name, or blueprint is nil/empty")
	}

	bp := b.newBlueprint(name)
	bp.create()
	blueprint(bp)

	if err := bp.build(c); err != nil {
		return err
	}

	return nil
}

func (b *baseBuilder) Drop(c Context, name string) error {
	if c == nil || name == "" {
		return errors.New("invalid arguments: context is nil or name is empty")
	}

	bp := b.newBlueprint(name)
	bp.drop()

	if err := bp.build(c); err != nil {
		return err
	}

	return nil
}

func (b *baseBuilder) DropIfExists(c Context, name string) error {
	if c == nil || name == "" {
		return errors.New("invalid arguments: context is nil or name is empty")
	}

	bp := b.newBlueprint(name)
	bp.dropIfExists()

	if err := bp.build(c); err != nil {
		return err
	}

	return nil
}

func (b *baseBuilder) Rename(c Context, oldName string, newName string) error {
	if c == nil || oldName == "" || newName == "" {
		return errors.New("invalid arguments: context is nil or old/new table name is empty")
	}

	bp := b.newBlueprint(oldName)
	bp.rename(newName)

	if err := bp.build(c); err != nil {
		return err
	}

	return nil
}

func (b *baseBuilder) Table(c Context, name string, blueprint func(table *Blueprint)) error {
	if c == nil || name == "" || blueprint == nil {
		return errors.New("invalid arguments: context is nil or name/blueprint is empty")
	}

	bp := b.newBlueprint(name)
	blueprint(bp)

	if err := bp.build(c); err != nil {
		return err
	}

	return nil
}
