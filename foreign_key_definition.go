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
