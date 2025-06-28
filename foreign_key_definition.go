package schema

// ForeignKeyDefinition defines the interface for defining a foreign key constraint in a database table.
type ForeignKeyDefinition interface {
	// References sets the column that this foreign key references in the other table.
	References(column string) ForeignKeyDefinition
	// On sets the table that this foreign key references.
	On(table string) ForeignKeyDefinition
	// OnDelete sets the action to take when the referenced row is deleted.
	OnDelete(action string) ForeignKeyDefinition
	// OnUpdate sets the action to take when the referenced row is updated.
	OnUpdate(action string) ForeignKeyDefinition
}
