package blueprint

import (
	"slices"

	"github.com/akfaiz/migris/internal/util"
)

// ColumnDefinition defines the interface for defining a column in a database table.
type ColumnDefinition interface {
	AutoIncrement() ColumnDefinition
	Change() ColumnDefinition
	Charset(charset string) ColumnDefinition
	Collation(collation string) ColumnDefinition
	Comment(comment string) ColumnDefinition
	Default(value any) ColumnDefinition
	Fixed() ColumnDefinition
	Index(params ...any) ColumnDefinition
	Nullable(value ...bool) ColumnDefinition
	OnUpdate(value any) ColumnDefinition
	Primary(value ...bool) ColumnDefinition
	Unique(params ...any) ColumnDefinition
	Unsigned() ColumnDefinition
	UseCurrent() ColumnDefinition
	UseCurrentOnUpdate() ColumnDefinition
	After(column string) ColumnDefinition
	First() ColumnDefinition
	VirtualAs(expression string) ColumnDefinition
	StoredAs(expression string) ColumnDefinition
	Invisible() ColumnDefinition
	GeneratedAs(expression ...string) ColumnDefinition
	Always(value ...bool) ColumnDefinition
	RenameTo(name string) ColumnDefinition
}

// Column represents a database column definition in a blueprint.
type Column struct {
	Commands              []string
	Name                  string
	RenameToVal           string
	ColumnType            string
	CharsetVal            *string
	CollationVal          *string
	CommentVal            *string
	DefaultValue          any
	OnUpdateValue         any
	UseCurrentVal         bool
	UseCurrentOnUpdateVal bool
	NullableVal           *bool
	AutoIncrementVal      *bool
	UnsignedVal           *bool
	PrimaryVal            *bool
	IndexVal              *bool
	IndexName             string
	UniqueVal             *bool
	UniqueName            string
	Length                *int
	Precision             *int
	Total                 *int
	Places                *int
	ChangeVal             bool
	Allowed               []string // for enum type columns
	Subtype               *string  // for geography and geometry types
	Srid                  *int     // for geography and geometry types
	AfterVal              *string
	FirstVal              bool
	FixedVal              *bool
	RawDefinition         *string
	VirtualAsVal          *string
	StoredAsVal           *string
	InvisibleVal          *bool
	GeneratedAsVal        *string
	AlwaysVal             *bool
}

// Expression is a type for expressions that can be used as default values for columns.
type Expression string

func (e Expression) String() string {
	return string(e)
}

var _ ColumnDefinition = &Column{}

func (c *Column) AddCommand(command string) {
	c.Commands = append(c.Commands, command)
}

func (c *Column) HasCommand(command string) bool {
	return slices.Contains(c.Commands, command)
}

func (c *Column) AutoIncrement() ColumnDefinition {
	c.AutoIncrementVal = util.PtrOf(true)
	return c
}

func (c *Column) Charset(charset string) ColumnDefinition {
	c.CharsetVal = &charset
	return c
}

func (c *Column) Change() ColumnDefinition {
	c.ChangeVal = true
	return c
}

func (c *Column) Collation(collation string) ColumnDefinition {
	c.CollationVal = &collation
	return c
}

func (c *Column) Comment(comment string) ColumnDefinition {
	c.AddCommand("comment")
	c.CommentVal = &comment
	return c
}

func (c *Column) Default(value any) ColumnDefinition {
	c.AddCommand("default")
	c.DefaultValue = value
	return c
}

func (c *Column) Index(params ...any) ColumnDefinition {
	index := true
	for _, param := range params {
		switch v := param.(type) {
		case bool:
			index = v
		case string:
			c.IndexName = v
		}
	}
	c.IndexVal = &index
	return c
}

func (c *Column) Nullable(value ...bool) ColumnDefinition {
	c.AddCommand("nullable")
	c.NullableVal = util.OptionalPtr(true, value...)
	return c
}

func (c *Column) OnUpdate(value any) ColumnDefinition {
	c.AddCommand("onUpdate")
	c.OnUpdateValue = value
	return c
}

func (c *Column) Primary(value ...bool) ColumnDefinition {
	val := util.Optional(true, value...)
	c.PrimaryVal = &val
	return c
}

func (c *Column) Unique(params ...any) ColumnDefinition {
	unique := true
	for _, param := range params {
		switch v := param.(type) {
		case bool:
			unique = v
		case string:
			c.UniqueName = v
		}
	}
	c.UniqueVal = &unique
	return c
}

func (c *Column) Unsigned() ColumnDefinition {
	c.UnsignedVal = util.PtrOf(true)
	return c
}

func (c *Column) UseCurrent() ColumnDefinition {
	c.UseCurrentVal = true
	return c
}

func (c *Column) UseCurrentOnUpdate() ColumnDefinition {
	c.UseCurrentOnUpdateVal = true
	return c
}

func (c *Column) After(column string) ColumnDefinition {
	c.AfterVal = &column
	return c
}

func (c *Column) First() ColumnDefinition {
	c.FirstVal = true
	return c
}

func (c *Column) VirtualAs(expression string) ColumnDefinition {
	c.VirtualAsVal = &expression
	return c
}

func (c *Column) StoredAs(expression string) ColumnDefinition {
	c.StoredAsVal = &expression
	return c
}

func (c *Column) Invisible() ColumnDefinition {
	c.InvisibleVal = util.PtrOf(true)
	return c
}

func (c *Column) GeneratedAs(expression ...string) ColumnDefinition {
	if len(expression) > 0 {
		c.GeneratedAsVal = &expression[0]
	} else {
		c.GeneratedAsVal = util.PtrOf("")
	}
	return c
}

func (c *Column) Always(value ...bool) ColumnDefinition {
	c.AlwaysVal = util.OptionalPtr(true, value...)
	return c
}

func (c *Column) RenameTo(name string) ColumnDefinition {
	c.RenameToVal = name
	return c
}

func (c *Column) Fixed() ColumnDefinition {
	c.FixedVal = util.PtrOf(true)
	return c
}
