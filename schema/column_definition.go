package schema

import (
	"slices"

	"github.com/akfaiz/migris/internal/util"
)

// ColumnDefinition defines the interface for defining a column in a database table.
type ColumnDefinition interface {
	// AutoIncrement sets the column to auto-increment.
	// This is typically used for primary key columns.
	AutoIncrement() ColumnDefinition
	// Change changes the column definition.
	Change() ColumnDefinition
	// Charset sets the character set for the column.
	Charset(charset string) ColumnDefinition
	// Collation sets the collation for the column.
	Collation(collation string) ColumnDefinition
	// Comment adds a comment to the column definition.
	Comment(comment string) ColumnDefinition
	// Default sets a default value for the column.
	Default(value any) ColumnDefinition
	// Index adds an index to the column.
	Index(params ...any) ColumnDefinition
	// Nullable sets the column to be nullable or not.
	Nullable(value ...bool) ColumnDefinition
	// OnUpdate sets the value to be used when the column is updated.
	OnUpdate(value any) ColumnDefinition
	// Primary sets the column as a primary key.
	Primary(value ...bool) ColumnDefinition
	// Unique sets the column to be unique.
	Unique(params ...any) ColumnDefinition
	// Unsigned sets the column to be unsigned (applicable for numeric types).
	Unsigned() ColumnDefinition
	// UseCurrent sets the column to use the current timestamp as default.
	UseCurrent() ColumnDefinition
	// UseCurrentOnUpdate sets the column to use the current timestamp on update.
	UseCurrentOnUpdate() ColumnDefinition
	// After places the column after the specified column (MySQL).
	After(column string) ColumnDefinition
	// First places the column "first" in the table (MySQL).
	First() ColumnDefinition
	// VirtualAs creates a virtual generated column based on the expression.
	VirtualAs(expression string) ColumnDefinition
	// StoredAs creates a stored generated column based on the expression.
	StoredAs(expression string) ColumnDefinition
	// Invisible sets the column as invisible (MySQL).
	Invisible() ColumnDefinition
	// GeneratedAs specifies that the column is a generated identity column (Postgres).
	GeneratedAs(expression ...string) ColumnDefinition
	// Always specifies that the generated identity column should always use the generated value (Postgres).
	Always(value ...bool) ColumnDefinition
	// RenameTo renames the column to a new name (MySQL).
	RenameTo(name string) ColumnDefinition
}

type columnDefinition struct {
	commands           []string
	name               string
	renameTo           string
	columnType         string
	charset            *string
	collation          *string
	comment            *string
	defaultValue       any
	onUpdateValue      any
	useCurrent         bool
	useCurrentOnUpdate bool
	nullable           *bool
	autoIncrement      *bool
	unsigned           *bool
	primary            *bool
	index              *bool
	indexName          string
	unique             *bool
	uniqueName         string
	length             *int
	precision          *int
	total              *int
	places             *int
	change             bool
	allowed            []string // for enum type columns
	subtype            *string  // for geography and geometry types
	srid               *int     // for geography and geometry types
	after              *string
	first              bool
	virtualAs          *string
	storedAs           *string
	invisible          *bool
	generatedAs        *string
	always             *bool
}

// Expression is a type for expressions that can be used as default values for columns.
//
// Example:
//
// schema.Timestamp("created_at").Default(schema.Expression("CURRENT_TIMESTAMP")).
type Expression string

func (e Expression) String() string {
	return string(e)
}

var _ ColumnDefinition = &columnDefinition{}

func (c *columnDefinition) addCommand(command string) {
	c.commands = append(c.commands, command)
}

func (c *columnDefinition) hasCommand(command string) bool {
	return slices.Contains(c.commands, command)
}

func (c *columnDefinition) SetDefault(value any) {
	c.addCommand("default")
	c.defaultValue = value
}

func (c *columnDefinition) SetOnUpdate(value any) {
	c.addCommand("onUpdate")
	c.onUpdateValue = value
}

func (c *columnDefinition) SetSubtype(value *string) {
	c.subtype = value
}

func (c *columnDefinition) AutoIncrement() ColumnDefinition {
	c.autoIncrement = util.PtrOf(true)
	return c
}

func (c *columnDefinition) Charset(charset string) ColumnDefinition {
	c.charset = &charset
	return c
}

func (c *columnDefinition) Change() ColumnDefinition {
	c.change = true
	return c
}

func (c *columnDefinition) Collation(collation string) ColumnDefinition {
	c.collation = &collation
	return c
}

func (c *columnDefinition) Comment(comment string) ColumnDefinition {
	c.addCommand("comment")
	c.comment = &comment
	return c
}

func (c *columnDefinition) Default(value any) ColumnDefinition {
	c.addCommand("default")
	c.defaultValue = value

	return c
}

func (c *columnDefinition) Index(params ...any) ColumnDefinition {
	index := true
	for _, param := range params {
		switch v := param.(type) {
		case bool:
			index = v
		case string:
			c.indexName = v
		}
	}
	c.index = &index
	return c
}

func (c *columnDefinition) Nullable(value ...bool) ColumnDefinition {
	c.addCommand("nullable")
	c.nullable = util.OptionalPtr(true, value...)
	return c
}

func (c *columnDefinition) OnUpdate(value any) ColumnDefinition {
	c.addCommand("onUpdate")
	c.onUpdateValue = value
	return c
}

func (c *columnDefinition) Primary(value ...bool) ColumnDefinition {
	val := util.Optional(true, value...)
	c.primary = &val
	return c
}

func (c *columnDefinition) Unique(params ...any) ColumnDefinition {
	unique := true
	for _, param := range params {
		switch v := param.(type) {
		case bool:
			unique = v
		case string:
			c.uniqueName = v
		}
	}
	c.unique = &unique
	return c
}

func (c *columnDefinition) Unsigned() ColumnDefinition {
	c.unsigned = util.PtrOf(true)
	return c
}

func (c *columnDefinition) UseCurrent() ColumnDefinition {
	c.useCurrent = true
	return c
}

func (c *columnDefinition) UseCurrentOnUpdate() ColumnDefinition {
	c.useCurrentOnUpdate = true
	return c
}

func (c *columnDefinition) After(column string) ColumnDefinition {
	c.after = &column
	return c
}

func (c *columnDefinition) First() ColumnDefinition {
	c.first = true
	return c
}

func (c *columnDefinition) VirtualAs(expression string) ColumnDefinition {
	c.virtualAs = &expression
	return c
}

func (c *columnDefinition) StoredAs(expression string) ColumnDefinition {
	c.storedAs = &expression
	return c
}

func (c *columnDefinition) Invisible() ColumnDefinition {
	c.invisible = util.PtrOf(true)
	return c
}

func (c *columnDefinition) GeneratedAs(expression ...string) ColumnDefinition {
	if len(expression) > 0 {
		c.generatedAs = &expression[0]
	} else {
		c.generatedAs = util.PtrOf("")
	}
	return c
}

func (c *columnDefinition) Always(value ...bool) ColumnDefinition {
	c.always = util.OptionalPtr(true, value...)
	return c
}

func (c *columnDefinition) RenameTo(name string) ColumnDefinition {
	c.renameTo = name
	return c
}
