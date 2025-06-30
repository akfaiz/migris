package schema

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

var ErrTableIsNotSet = errors.New("table name is not set")
var ErrBlueprintIsNil = errors.New("blueprint function is nil")
var ErrTxIsNil = errors.New("transaction is nil")

type builder interface {
	Create(ctx context.Context, tx *sql.Tx, name string, blueprint func(table *Blueprint)) error
	CreateIfNotExists(ctx context.Context, tx *sql.Tx, name string, blueprint func(table *Blueprint)) error
	Drop(ctx context.Context, tx *sql.Tx, name string) error
	DropIfExists(ctx context.Context, tx *sql.Tx, name string) error
	GetColumns(ctx context.Context, tx *sql.Tx, tableName string) ([]*Column, error)
	GetIndexes(ctx context.Context, tx *sql.Tx, tableName string) ([]*Index, error)
	GetTables(ctx context.Context, tx *sql.Tx) ([]*TableInfo, error)
	HasColumn(ctx context.Context, tx *sql.Tx, tableName string, columnName string) (bool, error)
	HasColumns(ctx context.Context, tx *sql.Tx, tableName string, columnNames []string) (bool, error)
	HasIndex(ctx context.Context, tx *sql.Tx, tableName string, indexes []string) (bool, error)
	HasTable(ctx context.Context, tx *sql.Tx, name string) (bool, error)
	Rename(ctx context.Context, tx *sql.Tx, oldName string, newName string) error
	Table(ctx context.Context, tx *sql.Tx, name string, blueprint func(table *Blueprint)) error
}

func newBuilder(dialect dialectType) (builder, error) {
	switch dialect {
	case postgres:
		return newPostgresBuilder(), nil
	case mysql:
		return newMysqlBuilder(), nil
	default:
		return nil, ErrDialectNotSet
	}
}

type baseBuilder struct {
	dialect dialectType
	grammar grammar
}

func (b *baseBuilder) validateTxAndName(tx *sql.Tx, name string) error {
	if isEmptyString(name) {
		return ErrTableIsNotSet
	}
	if tx == nil {
		return ErrTxIsNil
	}
	return nil
}

func (b *baseBuilder) validateCreateAndAlter(tx *sql.Tx, name string, blueprint func(table *Blueprint)) error {
	if isEmptyString(name) {
		return ErrTableIsNotSet
	}
	if b.dialect == postgres {
		names := strings.Split(name, ".")
		if len(names) > 2 {
			return errors.New("invalid table name: " + name + ", it should be in the format 'schema.table' or just 'table'")
		}
		if len(names) == 2 {
			if isEmptyString(names[0]) {
				return errors.New("schema name is empty in table name: " + name)
			}
			if isEmptyString(names[1]) {
				return errors.New("table name is empty in table name: " + name)
			}
		}
	}
	if blueprint == nil {
		return ErrBlueprintIsNil
	}
	if tx == nil {
		return ErrTxIsNil
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

	sqls, err := bp.toSql(b.grammar)
	if err != nil {
		return err
	}

	return execContext(ctx, tx, sqls...)
}

func (b *baseBuilder) CreateIfNotExists(ctx context.Context, tx *sql.Tx, name string, blueprint func(table *Blueprint)) error {
	if err := b.validateCreateAndAlter(tx, name, blueprint); err != nil {
		return err
	}

	bp := &Blueprint{name: name}
	bp.createIfNotExists()
	blueprint(bp)

	sqls, err := bp.toSql(b.grammar)
	if err != nil {
		return err
	}

	return execContext(ctx, tx, sqls...)
}

func (b *baseBuilder) Drop(ctx context.Context, tx *sql.Tx, name string) error {
	if err := b.validateTxAndName(tx, name); err != nil {
		return err
	}

	bp := &Blueprint{name: name}
	bp.drop()
	sqls, err := bp.toSql(b.grammar)
	if err != nil {
		return err
	}

	return execContext(ctx, tx, sqls...)
}

func (b *baseBuilder) DropIfExists(ctx context.Context, tx *sql.Tx, name string) error {
	if err := b.validateTxAndName(tx, name); err != nil {
		return err
	}

	bp := &Blueprint{name: name}
	bp.dropIfExists()
	sqls, err := bp.toSql(b.grammar)
	if err != nil {
		return err
	}

	return execContext(ctx, tx, sqls...)
}

func (b *baseBuilder) Rename(ctx context.Context, tx *sql.Tx, oldName string, newName string) error {
	if isEmptyString(oldName) || isEmptyString(newName) {
		return ErrTableIsNotSet
	}
	if tx == nil {
		return ErrTxIsNil
	}
	bp := &Blueprint{name: oldName, newName: newName}
	bp.rename()
	sqls, err := bp.toSql(b.grammar)
	if err != nil {
		return err
	}

	return execContext(ctx, tx, sqls...)
}

func (b *baseBuilder) Table(ctx context.Context, tx *sql.Tx, name string, blueprint func(table *Blueprint)) error {
	if err := b.validateCreateAndAlter(tx, name, blueprint); err != nil {
		return err
	}

	bp := &Blueprint{name: name}
	blueprint(bp)

	sqls, err := bp.toSql(b.grammar)
	if err != nil {
		return err
	}

	return execContext(ctx, tx, sqls...)
}
