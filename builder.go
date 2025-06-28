package schema

import (
	"context"
	"database/sql"
)

type Builder struct {
	grammar grammar
}

func newBuilder() (*Builder, error) {
	grammar, err := newGrammar()
	if err != nil {
		return nil, err
	}
	return &Builder{grammar: grammar}, nil
}

func (b *Builder) Create(ctx context.Context, tx *sql.Tx, name string, blueprint func(table *Blueprint)) error {
	bp := &Blueprint{name: name}
	blueprint(bp)

	sqls, err := b.grammar.compileCreate(bp)
	if err != nil {
		return err
	}

	return execContext(ctx, tx, sqls...)
}

func (b *Builder) CreateIfNotExists(ctx context.Context, tx *sql.Tx, name string, blueprint func(table *Blueprint)) error {
	bp := &Blueprint{name: name}
	blueprint(bp)

	sqls, err := b.grammar.compileCreate(bp)
	if err != nil {
		return err
	}

	return execContext(ctx, tx, sqls...)
}

func (b *Builder) Table(ctx context.Context, tx *sql.Tx, name string, blueprint func(table *Blueprint)) error {
	bp := &Blueprint{name: name}
	blueprint(bp)

	sqls, err := b.grammar.compileAlter(bp)
	if err != nil {
		return err
	}

	return execContext(ctx, tx, sqls...)
}

func (b *Builder) Drop(ctx context.Context, tx *sql.Tx, name string) error {
	bp := &Blueprint{name: name}
	sql, err := b.grammar.compileDrop(bp)
	if err != nil {
		return err
	}

	return execContext(ctx, tx, sql)
}

func (b *Builder) DropIfExists(ctx context.Context, tx *sql.Tx, name string) error {
	bp := &Blueprint{name: name}
	sql, err := b.grammar.compileDrop(bp)
	if err != nil {
		return err
	}

	return execContext(ctx, tx, sql)
}

func (b *Builder) Rename(ctx context.Context, tx *sql.Tx, oldName string, newName string) error {
	bp := &Blueprint{name: oldName, newName: newName}
	sql, err := b.grammar.compileDrop(bp)
	if err != nil {
		return err
	}

	return execContext(ctx, tx, sql)
}
