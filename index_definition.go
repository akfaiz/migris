package schema

// IndexDefinition defines the interface for defining an index in a database table.
type IndexDefinition interface {
	// Algorithm sets the algorithm for the index.
	Algorithm(algorithm string) IndexDefinition
	// Deferrable sets the index as deferrable.
	Deferrable(value ...bool) IndexDefinition
	// InitiallyImmediate sets the index to be initially immediate.
	InitiallyImmediate(value ...bool) IndexDefinition
	// Language sets the language for the index.
	// Used for full-text indexes in PostgreSQL.
	Language(language string) IndexDefinition
	// Name sets the name of the index.
	Name(name string) IndexDefinition
}
