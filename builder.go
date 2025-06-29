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

type builder struct {
	grammar grammar
}

func newBuilder() (*builder, error) {
	grammar, err := newGrammar()
	if err != nil {
		return nil, err
	}
	return &builder{grammar: grammar}, nil
}

func (b *builder) validateTxAndName(tx *sql.Tx, name string) error {
	if isEmptyString(name) {
		return ErrTableIsNotSet
	}
	if tx == nil {
		return ErrTxIsNil
	}
	return nil
}

func (b *builder) validateCreateAndAlter(tx *sql.Tx, name string, blueprint func(table *Blueprint)) error {
	if isEmptyString(name) {
		return ErrTableIsNotSet
	}
	if dialect == postgres {
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

func (b *builder) parseSchemaAndTable(name string) (string, string) {
	if dialect == postgres {
		names := strings.Split(name, ".")
		if len(names) == 2 {
			return names[0], names[1]
		}
		return "", names[0]
	}
	return "", name
}

func (b *builder) Create(ctx context.Context, tx *sql.Tx, name string, blueprint func(table *Blueprint)) error {
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

func (b *builder) CreateIfNotExists(ctx context.Context, tx *sql.Tx, name string, blueprint func(table *Blueprint)) error {
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

func (b *builder) Drop(ctx context.Context, tx *sql.Tx, name string) error {
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

func (b *builder) DropIfExists(ctx context.Context, tx *sql.Tx, name string) error {
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

func (b *builder) HasTable(ctx context.Context, tx *sql.Tx, name string) (bool, error) {
	if err := b.validateTxAndName(tx, name); err != nil {
		return false, err
	}

	schema, name := b.parseSchemaAndTable(name)
	query, err := b.grammar.compileTableExists(schema, name)
	if err != nil {
		return false, err
	}

	var exists bool
	if err := queryRowContext(ctx, tx, query).Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil // Table does not exist
		}
		return false, err // Other error
	}
	return exists, nil
}

func (b *builder) Rename(ctx context.Context, tx *sql.Tx, oldName string, newName string) error {
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

func (b *builder) Table(ctx context.Context, tx *sql.Tx, name string, blueprint func(table *Blueprint)) error {
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
