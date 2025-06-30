package schema

// ColumnDefinition defines the interface for defining a column in a database table.
type ColumnDefinition interface {
	// Change changes the column definition.
	Change() ColumnDefinition
	// Comment adds a comment to the column definition.
	Comment(comment string) ColumnDefinition
	// Default sets a default value for the column.
	Default(value any) ColumnDefinition
	// Index adds an index to the column.
	Index(indexName ...string) ColumnDefinition
	// Nullable sets the column to be nullable or not.
	Nullable(value ...bool) ColumnDefinition
	// Primary sets the column as a primary key.
	Primary() ColumnDefinition
	// Unique sets the column to be unique.
	Unique(indexName ...string) ColumnDefinition
}
