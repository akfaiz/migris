package blueprint

import "strings"

// ForeignIDColumnDefinition extends ColumnDefinition with FK constraint helpers.
type ForeignIDColumnDefinition interface {
	ColumnDefinition
	Constrained(table ...string) ForeignKeyDefinition
}

type foreignIDColumnDefinition struct {
	column    *Column
	blueprint *Blueprint
}

var _ ForeignIDColumnDefinition = &foreignIDColumnDefinition{}

func (f *foreignIDColumnDefinition) Constrained(table ...string) ForeignKeyDefinition {
	var tableName string
	if len(table) > 0 && table[0] != "" {
		tableName = table[0]
	} else {
		tableName = strings.TrimSuffix(f.column.Name, "_id") + "s"
	}
	return f.blueprint.Foreign(f.column.Name).References("id").On(tableName)
}

func (f *foreignIDColumnDefinition) GetColumn() *Column { return f.column }

func (f *foreignIDColumnDefinition) AutoIncrement() ColumnDefinition {
	return f.column.AutoIncrement()
}

func (f *foreignIDColumnDefinition) Change() ColumnDefinition { return f.column.Change() }

func (f *foreignIDColumnDefinition) Charset(charset string) ColumnDefinition {
	return f.column.Charset(charset)
}

func (f *foreignIDColumnDefinition) Collation(collation string) ColumnDefinition {
	return f.column.Collation(collation)
}

func (f *foreignIDColumnDefinition) Comment(comment string) ColumnDefinition {
	return f.column.Comment(comment)
}

func (f *foreignIDColumnDefinition) Default(value any) ColumnDefinition {
	return f.column.Default(value)
}

func (f *foreignIDColumnDefinition) Fixed() ColumnDefinition { return f.column.Fixed() }

func (f *foreignIDColumnDefinition) Index(params ...any) ColumnDefinition {
	return f.column.Index(params...)
}

func (f *foreignIDColumnDefinition) Nullable(value ...bool) ColumnDefinition {
	return f.column.Nullable(value...)
}

func (f *foreignIDColumnDefinition) OnUpdate(value any) ColumnDefinition {
	return f.column.OnUpdate(value)
}

func (f *foreignIDColumnDefinition) Primary(value ...bool) ColumnDefinition {
	return f.column.Primary(value...)
}

func (f *foreignIDColumnDefinition) Unique(params ...any) ColumnDefinition {
	return f.column.Unique(params...)
}

func (f *foreignIDColumnDefinition) Unsigned() ColumnDefinition { return f.column.Unsigned() }

func (f *foreignIDColumnDefinition) UseCurrent() ColumnDefinition { return f.column.UseCurrent() }

func (f *foreignIDColumnDefinition) UseCurrentOnUpdate() ColumnDefinition {
	return f.column.UseCurrentOnUpdate()
}

func (f *foreignIDColumnDefinition) After(column string) ColumnDefinition {
	return f.column.After(column)
}

func (f *foreignIDColumnDefinition) First() ColumnDefinition { return f.column.First() }

func (f *foreignIDColumnDefinition) VirtualAs(expression string) ColumnDefinition {
	return f.column.VirtualAs(expression)
}

func (f *foreignIDColumnDefinition) StoredAs(expression string) ColumnDefinition {
	return f.column.StoredAs(expression)
}

func (f *foreignIDColumnDefinition) Invisible() ColumnDefinition { return f.column.Invisible() }

func (f *foreignIDColumnDefinition) GeneratedAs(expression ...string) ColumnDefinition {
	return f.column.GeneratedAs(expression...)
}

func (f *foreignIDColumnDefinition) Always(value ...bool) ColumnDefinition {
	return f.column.Always(value...)
}

func (f *foreignIDColumnDefinition) RenameTo(name string) ColumnDefinition {
	return f.column.RenameTo(name)
}
