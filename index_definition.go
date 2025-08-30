package schema

import "github.com/afkdevs/go-schema/internal/util"

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

type indexDefinition struct {
	*command
}

func (id *indexDefinition) Algorithm(algorithm string) IndexDefinition {
	id.algorithm = algorithm
	return id
}

func (id *indexDefinition) Deferrable(value ...bool) IndexDefinition {
	val := util.Optional(true, value...)
	id.deferrable = &val
	return id
}

func (id *indexDefinition) InitiallyImmediate(value ...bool) IndexDefinition {
	val := util.Optional(true, value...)
	id.initiallyImmediate = &val
	return id
}

func (id *indexDefinition) Language(language string) IndexDefinition {
	id.language = language
	return id
}

func (id *indexDefinition) Name(name string) IndexDefinition {
	id.index = name
	return id
}
