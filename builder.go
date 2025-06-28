package schema

import (
	"context"
	"database/sql"
	"errors"
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
	if blueprint == nil {
		return ErrBlueprintIsNil
	}
	if tx == nil {
		return ErrTxIsNil
	}
	return nil
}

func (b *builder) Create(ctx context.Context, tx *sql.Tx, name string, blueprint func(table *Blueprint)) error {
	if err := b.validateCreateAndAlter(tx, name, blueprint); err != nil {
		return err
	}

	bp := &Blueprint{name: name}
	blueprint(bp)

	sqls, err := b.grammar.compileCreate(bp)
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
	blueprint(bp)

	sqls, err := b.grammar.compileCreateIfNotExists(bp)
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

	sqls, err := b.grammar.compileAlter(bp)
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
	sql, err := b.grammar.compileDrop(bp)
	if err != nil {
		return err
	}

	return execContext(ctx, tx, sql)
}

func (b *builder) DropIfExists(ctx context.Context, tx *sql.Tx, name string) error {
	if err := b.validateTxAndName(tx, name); err != nil {
		return err
	}

	bp := &Blueprint{name: name}
	sql, err := b.grammar.compileDropIfExists(bp)
	if err != nil {
		return err
	}

	return execContext(ctx, tx, sql)
}

func (b *builder) Rename(ctx context.Context, tx *sql.Tx, oldName string, newName string) error {
	if isEmptyString(oldName) || isEmptyString(newName) {
		return ErrTableIsNotSet
	}
	if tx == nil {
		return ErrTxIsNil
	}
	bp := &Blueprint{name: oldName, newName: newName}
	sql, err := b.grammar.compileRename(bp)
	if err != nil {
		return err
	}

	return execContext(ctx, tx, sql)
}
