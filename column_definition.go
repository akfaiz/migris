package schema

import "slices"

// ColumnDefinition defines the interface for defining a column in a database table.
type ColumnDefinition interface {
	// AutoIncrement sets the column to auto-increment.
	// This is typically used for primary key columns.
	AutoIncrement() ColumnDefinition
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
	// Unsigned sets the column to be unsigned (applicable for numeric types).
	Unsigned() ColumnDefinition
	// UseCurrent sets the column to use the current timestamp as default.
	UseCurrent() ColumnDefinition
	// UseCurrentOnUpdate sets the column to use the current timestamp on update.
	UseCurrentOnUpdate() ColumnDefinition
}

var _ ColumnDefinition = &columnDefinition{}

type columnDefinition struct {
	name             string
	columnType       columnType
	customColumnType string // for custom column types
	commands         []string
	comment          string
	defaultValue     any
	onUpdateValue    string
	nullable         bool
	autoIncrement    bool
	unsigned         bool
	primary          bool
	index            bool
	indexName        string
	unique           bool
	uniqueName       string
	length           int
	precision        int
	total            int
	places           int
	changed          bool
	allowedEnums     []string // for enum type columns
	subType          string   // for geography and geometry types
	srid             int      // for geography and geometry types
}

func (c *columnDefinition) addCommand(command string) {
	if command == "" {
		return
	}
	if !slices.Contains(c.commands, command) {
		c.commands = append(c.commands, command)
	}
}

func (c *columnDefinition) hasCommand(command string) bool {
	return slices.Contains(c.commands, command)
}

func (c *columnDefinition) AutoIncrement() ColumnDefinition {
	c.addCommand("autoIncrement")
	c.autoIncrement = true
	return c
}

func (c *columnDefinition) Comment(comment string) ColumnDefinition {
	c.addCommand("comment")
	c.comment = comment
	return c
}

func (c *columnDefinition) Default(value any) ColumnDefinition {
	c.addCommand("default")
	c.defaultValue = value

	return c
}

func (c *columnDefinition) Index(indexName ...string) ColumnDefinition {
	c.index = true
	c.indexName = optional("", indexName...)
	c.addCommand("index")
	return c
}

func (c *columnDefinition) Nullable(value ...bool) ColumnDefinition {
	c.addCommand("nullable")
	c.nullable = optional(true, value...)
	return c
}

func (c *columnDefinition) Primary() ColumnDefinition {
	c.addCommand("primary")
	c.primary = true
	return c
}

func (c *columnDefinition) Unique(indexName ...string) ColumnDefinition {
	c.addCommand("unique")
	c.unique = true
	c.uniqueName = optional("", indexName...)
	return c
}

func (c *columnDefinition) Change() ColumnDefinition {
	c.addCommand("change")
	c.changed = true
	return c
}

func (c *columnDefinition) Unsigned() ColumnDefinition {
	c.unsigned = true
	return c
}

func (c *columnDefinition) UseCurrent() ColumnDefinition {
	c.Default("CURRENT_TIMESTAMP")
	return c
}

func (c *columnDefinition) UseCurrentOnUpdate() ColumnDefinition {
	c.onUpdateValue = "CURRENT_TIMESTAMP"
	return c
}
