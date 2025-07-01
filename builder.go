package schema

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

// Builder is an interface that defines methods for creating, dropping, and managing database tables.
type Builder interface {
	// Create creates a new table with the given name and applies the provided blueprint.
	Create(ctx context.Context, tx *sql.Tx, name string, blueprint func(table *Blueprint)) error
	// CreateIfNotExists creates a new table with the given name and applies the provided blueprint if it does not already exist.
	CreateIfNotExists(ctx context.Context, tx *sql.Tx, name string, blueprint func(table *Blueprint)) error
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
func NewBuilder(dialect string) (Builder, error) {
	switch dialect {
	case "postgres", "pgx":
		return newPostgresBuilder(), nil
	case "mysql", "mariadb":
		return newMysqlBuilder(), nil
	default:
		return nil, errors.New("unknown dialect: " + dialect)
	}
}

type baseBuilder struct {
	dialect string
	grammar grammar
}

func (b *baseBuilder) validateTxAndName(tx *sql.Tx, name string) error {
	if name == "" {
		return errors.New("table name is empty")
	}
	if tx == nil {
		return errors.New("transaction is nil")
	}
	return nil
}

func (b *baseBuilder) validateCreateAndAlter(tx *sql.Tx, name string, blueprint func(table *Blueprint)) error {
	if name == "" {
		return errors.New("table name is empty")
	}
	if b.dialect == "postgres" || b.dialect == "pgx" {
		names := strings.Split(name, ".")
		if len(names) > 2 {
			return errors.New("invalid table name: " + name + ", it should be in the format 'schema.table' or just 'table'")
		}
		if len(names) == 2 {
			if names[0] == "" || names[1] == "" {
				return errors.New("invalid table name: " + name + ", schema and table name cannot be empty")
			}
		}
	}
	if blueprint == nil {
		return errors.New("blueprint function is nil")
	}
	if tx == nil {
		return errors.New("transaction is nil")
	}
	return nil
}

func (b *baseBuilder) Create(ctx context.Context, tx *sql.Tx, name string, blueprint func(table *Blueprint)) error {
	if err := b.validateCreateAndAlter(tx, name, blueprint); err != nil {
		return err
	}

	bp := &Blueprint{name: name}
	bp.create()
	blueprint(bp)

	statements, err := bp.toSql(b.grammar)
	if err != nil {
		return err
	}

	return execContext(ctx, tx, statements...)
}

func (b *baseBuilder) CreateIfNotExists(ctx context.Context, tx *sql.Tx, name string, blueprint func(table *Blueprint)) error {
	if err := b.validateCreateAndAlter(tx, name, blueprint); err != nil {
		return err
	}

	bp := &Blueprint{name: name}
	bp.createIfNotExists()
	blueprint(bp)

	statements, err := bp.toSql(b.grammar)
	if err != nil {
		return err
	}

	return execContext(ctx, tx, statements...)
}

func (b *baseBuilder) Drop(ctx context.Context, tx *sql.Tx, name string) error {
	if err := b.validateTxAndName(tx, name); err != nil {
		return err
	}

	bp := &Blueprint{name: name}
	bp.drop()
	statements, err := bp.toSql(b.grammar)
	if err != nil {
		return err
	}

	return execContext(ctx, tx, statements...)
}

func (b *baseBuilder) DropIfExists(ctx context.Context, tx *sql.Tx, name string) error {
	if err := b.validateTxAndName(tx, name); err != nil {
		return err
	}

	bp := &Blueprint{name: name}
	bp.dropIfExists()
	statements, err := bp.toSql(b.grammar)
	if err != nil {
		return err
	}

	return execContext(ctx, tx, statements...)
}

func (b *baseBuilder) Rename(ctx context.Context, tx *sql.Tx, oldName string, newName string) error {
	if oldName == "" || newName == "" {
		return errors.New("old or new table name is empty")
	}
	if tx == nil {
		return errors.New("transaction is nil")
	}
	bp := &Blueprint{name: oldName, newName: newName}
	bp.rename()
	statements, err := bp.toSql(b.grammar)
	if err != nil {
		return err
	}

	return execContext(ctx, tx, statements...)
}

func (b *baseBuilder) Table(ctx context.Context, tx *sql.Tx, name string, blueprint func(table *Blueprint)) error {
	if err := b.validateCreateAndAlter(tx, name, blueprint); err != nil {
		return err
	}

	bp := &Blueprint{name: name}
	blueprint(bp)

	statements, err := bp.toSql(b.grammar)
	if err != nil {
		return err
	}

	return execContext(ctx, tx, statements...)
}
