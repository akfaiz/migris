package schema

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
	// NoActionOnDelete sets the foreign key to do nothing on delete.
	NoActionOnDelete() ForeignKeyDefinition
	// NoActionOnUpdate sets the foreign key to do nothing on update.
	NoActionOnUpdate() ForeignKeyDefinition
	// NullOnDelete sets the foreign key to set the column to NULL on delete.
	NullOnDelete() ForeignKeyDefinition
	// NullOnUpdate sets the foreign key to set the column to NULL on update.
	NullOnUpdate() ForeignKeyDefinition
	// On sets the table that this foreign key references.
	On(table string) ForeignKeyDefinition
	// OnDelete sets the action to take when the referenced row is deleted.
	OnDelete(action string) ForeignKeyDefinition
	// OnUpdate sets the action to take when the referenced row is updated.
	OnUpdate(action string) ForeignKeyDefinition
	// References sets the column that this foreign key references in the other table.
	References(column string) ForeignKeyDefinition
	// RestrictOnDelete sets the foreign key to restrict deletion of the referenced row.
	RestrictOnDelete() ForeignKeyDefinition
	// RestrictOnUpdate sets the foreign key to restrict updating of the referenced row.
	RestrictOnUpdate() ForeignKeyDefinition
}

var _ ForeignKeyDefinition = &foreignKeyDefinition{}

type foreignKeyDefinition struct {
	tableName          string
	column             string
	constaintName      string // name of the foreign key constraint
	references         string
	on                 string
	onDelete           string
	onUpdate           string
	deferrable         *bool
	initiallyImmediate *bool
}

func (fk *foreignKeyDefinition) CascadeOnDelete() ForeignKeyDefinition {
	fk.onDelete = "CASCADE"
	return fk
}

func (fk *foreignKeyDefinition) CascadeOnUpdate() ForeignKeyDefinition {
	fk.onUpdate = "CASCADE"
	return fk
}

func (fk *foreignKeyDefinition) Deferrable(value ...bool) ForeignKeyDefinition {
	val := optional(true, value...)
	fk.deferrable = &val
	return fk
}

func (fk *foreignKeyDefinition) InitiallyImmediate(value ...bool) ForeignKeyDefinition {
	val := optional(true, value...)
	fk.initiallyImmediate = &val
	return fk
}

func (fk *foreignKeyDefinition) Name(name string) ForeignKeyDefinition {
	fk.constaintName = name
	return fk
}

func (fk *foreignKeyDefinition) NoActionOnDelete() ForeignKeyDefinition {
	fk.onDelete = "NO ACTION"
	return fk
}

func (fk *foreignKeyDefinition) NoActionOnUpdate() ForeignKeyDefinition {
	fk.onUpdate = "NO ACTION"
	return fk
}

func (fk *foreignKeyDefinition) NullOnDelete() ForeignKeyDefinition {
	fk.onDelete = "SET NULL"
	return fk
}

func (fk *foreignKeyDefinition) NullOnUpdate() ForeignKeyDefinition {
	fk.onUpdate = "SET NULL"
	return fk
}

func (fk *foreignKeyDefinition) On(table string) ForeignKeyDefinition {
	fk.on = table
	return fk
}

func (fk *foreignKeyDefinition) OnDelete(action string) ForeignKeyDefinition {
	fk.onDelete = action
	return fk
}

func (fk *foreignKeyDefinition) OnUpdate(action string) ForeignKeyDefinition {
	fk.onUpdate = action
	return fk
}

func (fk *foreignKeyDefinition) References(column string) ForeignKeyDefinition {
	fk.references = column
	return fk
}

func (fk *foreignKeyDefinition) RestrictOnDelete() ForeignKeyDefinition {
	fk.onDelete = "RESTRICT"
	return fk
}

func (fk *foreignKeyDefinition) RestrictOnUpdate() ForeignKeyDefinition {
	fk.onUpdate = "RESTRICT"
	return fk
}
