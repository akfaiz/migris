package schema

// IndexDefinition defines the interface for defining an index in a database table.
type IndexDefinition interface {
	// Algorithm sets the algorithm for the index.
	Algorithm(algorithm string) IndexDefinition
	// Name sets the name of the index.
	Name(name string) IndexDefinition
}
