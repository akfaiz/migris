package builders

import (
	"errors"

	"github.com/akfaiz/migris/internal/dialect"
	"github.com/akfaiz/migris/schema/blueprint"
	"github.com/akfaiz/migris/schema/core"
)

// Builder is an interface that defines methods for creating, dropping, and managing database tables.
type Builder interface {
	Create(c core.Context, name string, blueprint func(table *blueprint.Blueprint)) error
	Table(c core.Context, name string, blueprint func(table *blueprint.Blueprint)) error
	Drop(c core.Context, name string) error
	DropIfExists(c core.Context, name string) error
	Rename(c core.Context, from, to string) error
	HasTable(c core.Context, name string) (bool, error)
	HasColumn(c core.Context, table, column string) (bool, error)
	HasColumns(c core.Context, table string, columns []string) (bool, error)
	HasIndex(c core.Context, table string, columns []string) (bool, error)
	GetColumns(c core.Context, table string) ([]*core.Column, error)
	GetIndexes(c core.Context, table string) ([]*core.Index, error)
	GetTables(c core.Context) ([]*core.TableInfo, error)
}

func NewBuilder(dialectValue string) (Builder, error) {
	dialectVal := dialect.FromString(dialectValue)
	switch dialectVal {
	case dialect.MySQL:
		return NewMysqlBuilder(), nil
	case dialect.MariaDB:
		return NewMariadbBuilder(), nil
	case dialect.Postgres:
		return NewPostgresBuilder(), nil
	case dialect.SQLite3:
		return NewSqliteBuilder(), nil
	case dialect.Unknown:
		return nil, errors.New("unsupported dialect: " + dialectValue)
	default:
		return nil, errors.New("unsupported dialect: " + dialectValue)
	}
}

type baseBuilder struct {
	Grammar blueprint.Grammar
	Outer   Builder
}

func (b *baseBuilder) newBlueprint(name string) *blueprint.Blueprint {
	bp := &blueprint.Blueprint{
		Name:    name,
		Grammar: b.Grammar,
		Builder: b.Outer,
	}
	return bp
}

func (b *baseBuilder) Create(c core.Context, name string, bp func(table *blueprint.Blueprint)) error {
	if c == nil || name == "" || bp == nil {
		return errors.New("invalid arguments")
	}

	blueprint := b.newBlueprint(name)
	blueprint.SetDialect(dialect.FromString(c.Dialect()))
	blueprint.Create()
	bp(blueprint)

	if err := blueprint.Build(c); err != nil {
		return err
	}

	return nil
}

func (b *baseBuilder) Table(c core.Context, name string, bp func(table *blueprint.Blueprint)) error {
	if c == nil || name == "" || bp == nil {
		return errors.New("invalid arguments")
	}

	blueprint := b.newBlueprint(name)
	blueprint.SetDialect(dialect.FromString(c.Dialect()))
	bp(blueprint)

	if err := blueprint.Build(c); err != nil {
		return err
	}

	return nil
}
