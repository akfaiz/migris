package schema

import "github.com/afkdevs/go-schema/internal/util"

// ForeignKeyDefinition defines the interface for defining a foreign key constraint in a database table.
type ForeignKeyDefinition interface {
	// CascadeOnDelete sets the foreign key to cascade on delete.
	CascadeOnDelete() ForeignKeyDefinition
	// CascadeOnUpdate sets the foreign key to cascade on update.
	CascadeOnUpdate() ForeignKeyDefinition
	// Deferrable sets the foreign key as deferrable.
	Deferrable(value ...bool) ForeignKeyDefinition
	// InitiallyImmediate sets the foreign key to be initially immediate.
	InitiallyImmediate(value ...bool) ForeignKeyDefinition
	// Name sets the name of the foreign key constraint.
	// This is optional and can be used to give a specific name to the foreign key.
	Name(name string) ForeignKeyDefinition
	// NoActionOnDelete set the foreign key to do nothing on delete.
	NoActionOnDelete() ForeignKeyDefinition
	// NoActionOnUpdate set the foreign key to do nothing on the update.
	NoActionOnUpdate() ForeignKeyDefinition
	// NullOnDelete set the foreign key to set the column to NULL on delete.
	NullOnDelete() ForeignKeyDefinition
	// NullOnUpdate set the foreign key to set the column to NULL on update.
	NullOnUpdate() ForeignKeyDefinition
	// On sets the table that these foreign key references.
	On(table string) ForeignKeyDefinition
	// OnDelete set the action to take when the referenced row is deleted.
	OnDelete(action string) ForeignKeyDefinition
	// OnUpdate set the action to take when the referenced row is updated.
	OnUpdate(action string) ForeignKeyDefinition
	// References set the column that this foreign key references in the other table.
	References(column string) ForeignKeyDefinition
	// RestrictOnDelete set the foreign key to restrict deletion of the referenced row.
	RestrictOnDelete() ForeignKeyDefinition
	// RestrictOnUpdate set the foreign key to restrict updating of the referenced row.
	RestrictOnUpdate() ForeignKeyDefinition
}

type foreignKeyDefinition struct {
	*command
}

func (fd *foreignKeyDefinition) CascadeOnDelete() ForeignKeyDefinition {
	return fd.OnDelete("CASCADE")
}

func (fd *foreignKeyDefinition) CascadeOnUpdate() ForeignKeyDefinition {
	return fd.OnUpdate("CASCADE")
}

func (fd *foreignKeyDefinition) Deferrable(value ...bool) ForeignKeyDefinition {
	val := util.Optional(true, value...)
	fd.deferrable = &val
	return fd
}

func (fd *foreignKeyDefinition) InitiallyImmediate(value ...bool) ForeignKeyDefinition {
	val := util.Optional(true, value...)
	fd.initiallyImmediate = &val
	return fd
}

func (fd *foreignKeyDefinition) Name(name string) ForeignKeyDefinition {
	fd.index = name
	return fd
}

func (fd *foreignKeyDefinition) NoActionOnDelete() ForeignKeyDefinition {
	return fd.OnDelete("NO ACTION")
}

func (fd *foreignKeyDefinition) NoActionOnUpdate() ForeignKeyDefinition {
	return fd.OnUpdate("NO ACTION")
}

func (fd *foreignKeyDefinition) NullOnDelete() ForeignKeyDefinition {
	return fd.OnDelete("SET NULL")
}

func (fd *foreignKeyDefinition) NullOnUpdate() ForeignKeyDefinition {
	return fd.OnUpdate("SET NULL")
}

func (fd *foreignKeyDefinition) On(table string) ForeignKeyDefinition {
	fd.on = table
	return fd
}

func (fd *foreignKeyDefinition) OnDelete(action string) ForeignKeyDefinition {
	fd.onDelete = action
	return fd
}

func (fd *foreignKeyDefinition) OnUpdate(action string) ForeignKeyDefinition {
	fd.onUpdate = action
	return fd
}

func (fd *foreignKeyDefinition) References(columns string) ForeignKeyDefinition {
	fd.references = []string{columns}
	return fd
}

func (fd *foreignKeyDefinition) RestrictOnDelete() ForeignKeyDefinition {
	return fd.OnDelete("RESTRICT")
}

func (fd *foreignKeyDefinition) RestrictOnUpdate() ForeignKeyDefinition {
	return fd.OnUpdate("RESTRICT")
}
