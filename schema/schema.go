package schema

import (
	"database/sql"
	"errors"

	"github.com/akfaiz/migris/internal/config"
	"github.com/akfaiz/migris/internal/dialect"
)

// Column represents a database column with its properties.
type Column struct {
	Name       string         // Name is the name of the column.
	TypeName   string         // TypeName is the name of the column type (e.g., "VARCHAR", "INT").
	TypeFull   string         // TypeFull is the full type name including any modifiers (e.g., "VARCHAR(255)", "INT(11)").
	Collation  sql.NullString // Collation is the collation of the column, if applicable.
	Nullable   bool           // Nullable indicates whether the column can contain NULL values.
	DefaultVal sql.NullString // DefaultVal is the default value for the column, if any.
	Comment    sql.NullString // Comment is an optional comment for the column.
	Extra      sql.NullString // Extra contains additional information about the column (e.g., "auto_increment").
}

// Index represents a database index with its properties.
type Index struct {
	Name    string   // Name is the name of the index.
	Columns []string // Columns is a slice of column names that are part of the index.
	Type    string   // e.g., "btree", "hash"
	Unique  bool     // Indicates if the index is unique
	Primary bool     // Indicates if the index is a primary key
}

// TableInfo represents information about a database table.
// It includes the table name, schema, size, and an optional comment.
type TableInfo struct {
	Name      string         // Name is the name of the table.
	Schema    string         // Schema is the schema where the table resides.
	Size      int64          // Size is the size of the table in bytes.
	Comment   sql.NullString // Comment is an optional comment for the table.
	Engine    sql.NullString // Engine is the storage engine used for the table (e.g., "InnoDB", "MyISAM").
	Collation sql.NullString // Collation is the collation used for the table (e.g., "utf8mb4_general_ci").
}

func newBuilder() (Builder, error) {
	dialectVal := config.GetDialect()
	if dialectVal == dialect.Unknown {
		return nil, errors.New("schema dialect is not set, please call schema.SetDialect() before using schema functions")
	}

	builder, err := NewBuilder(dialectVal.String())
	if err != nil {
		return nil, err
	}

	return builder, nil
}

// Create creates a new table with the given name and blueprint.
// The blueprint function is used to define the structure of the table.
// It returns an error if the table creation fails.
//
// Example:
//
//	err := schema.Create(ctx, tx, "users", func(table *schema.Blueprint) {
//	    table.ID()
//	    table.String("name").Nullable(false)
//	    table.String("email").Unique().Nullable(false)
//	    table.String("password").Nullable()
//	    table.Timestamp("created_at").Default("CURRENT_TIMESTAMP").Nullable(false)
//	    table.Timestamp("updated_at").Default("CURRENT_TIMESTAMP").Nullable(false)
//	})
func Create(c *Context, name string, blueprint func(table *Blueprint)) error {
	builder, err := newBuilder()
	if err != nil {
		return err
	}

	return builder.Create(c, name, blueprint)
}

// Drop removes the table with the given name.
// It returns an error if the table removal fails.
//
// Example:
//
//	err := schema.Drop(ctx, tx, "users")
func Drop(c *Context, name string) error {
	builder, err := newBuilder()
	if err != nil {
		return err
	}

	return builder.Drop(c, name)
}

// DropIfExists removes the table with the given name if it exists.
// It returns an error if the table removal fails.
//
// Example:
//
//	err := schema.DropIfExists(ctx, tx, "users")
func DropIfExists(c *Context, name string) error {
	builder, err := newBuilder()
	if err != nil {
		return err
	}

	return builder.DropIfExists(c, name)
}

// GetColumns retrieves the columns of the specified table.
// It returns a slice of Column structs representing the columns in the table.
//
// Example:
//
//	columns, err := schema.GetColumns(ctx, tx, "users")
func GetColumns(c *Context, tableName string) ([]*Column, error) {
	builder, err := newBuilder()
	if err != nil {
		return nil, err
	}

	return builder.GetColumns(c, tableName)
}

// GetIndexes retrieves the indexes of the specified table.
// It returns a slice of Index structs representing the indexes in the table.
//
// Example:
//
//	indexes, err := schema.GetIndexes(ctx, tx, "users")
func GetIndexes(c *Context, tableName string) ([]*Index, error) {
	builder, err := newBuilder()
	if err != nil {
		return nil, err
	}

	return builder.GetIndexes(c, tableName)
}

// GetTables retrieves all tables in the database.
// It returns a slice of TableInfo structs containing information about each table.
//
// Example:
//
//	tables, err := schema.GetTables(ctx, tx)
func GetTables(c *Context) ([]*TableInfo, error) {
	builder, err := newBuilder()
	if err != nil {
		return nil, err
	}

	return builder.GetTables(c)
}

// HasColumn checks if a column with the given name exists in the specified table.
// It returns true if the column exists, false otherwise.
//
// Example:
//
//	exists, err := schema.HasColumn(ctx, tx, "users", "email")
func HasColumn(c *Context, tableName string, columnName string) (bool, error) {
	builder, err := newBuilder()
	if err != nil {
		return false, err
	}

	return builder.HasColumn(c, tableName, columnName)
}

// HasColumns checks if the specified columns exist in the given table.
// It returns true if all specified columns exist, false otherwise.
//
// Example:
//
//	exists, err := schema.HasColumns(ctx, tx, "users", []string{"email", "name"})
//
// If any of the specified columns do not exist, it returns false.
func HasColumns(c *Context, tableName string, columnNames []string) (bool, error) {
	builder, err := newBuilder()
	if err != nil {
		return false, err
	}

	return builder.HasColumns(c, tableName, columnNames)
}

// HasIndex checks if an index with the given name exists in the specified table.
// It returns true if the index exists, false otherwise.
//
// Example:
//
//	exists, err := schema.HasIndex(ctx, tx, "users", []string{"uk_users_email"}) // Checks if the index with name "uk_users_email" exists in the "users" table.
//
//	exists, err := schema.HasIndex(ctx, tx, "users", []string{"email", "name"}) // Checks if a composite index exists on the "email" and "name" columns in the "users" table.
func HasIndex(c *Context, tableName string, indexes []string) (bool, error) {
	builder, err := newBuilder()
	if err != nil {
		return false, err
	}

	return builder.HasIndex(c, tableName, indexes)
}

// HasTable checks if a table with the given name exists in the database.
// It returns true if the table exists, false otherwise.
// It returns an error if the check fails.
//
// Example:
//
//	exists, err := schema.HasTable(ctx, tx, "users")
func HasTable(c *Context, name string) (bool, error) {
	builder, err := newBuilder()
	if err != nil {
		return false, err
	}

	return builder.HasTable(c, name)
}

// Rename changes the name of the table from name to newName.
// It returns an error if the renaming fails.
//
// Example:
//
//	err := schema.Rename(ctx, tx, "users", "people")
func Rename(c *Context, name string, newName string) error {
	builder, err := newBuilder()
	if err != nil {
		return err
	}

	return builder.Rename(c, name, newName)
}

// Table modifies an existing table with the given name and blueprint.
// The blueprint function is used to define the modifications to the table.
// It returns an error if the table modification fails.
//
// Example:
//
//	err := schema.Table(ctx, tx, "users", func(table *schema.Blueprint) {
//	    table.Column("name").String().Nullable(false)
//	    table.DropColumn("password")
//	    table.RenameColumn("email", "contact_email")
//	})
func Table(c *Context, name string, blueprint func(table *Blueprint)) error {
	builder, err := newBuilder()
	if err != nil {
		return err
	}

	return builder.Table(c, name, blueprint)
}
