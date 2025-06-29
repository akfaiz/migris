package schema

import (
	"context"
	"database/sql"
)

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
func Create(ctx context.Context, tx *sql.Tx, name string, blueprint func(table *Blueprint)) error {
	builder, err := newBuilder()
	if err != nil {
		return err
	}

	return builder.Create(ctx, tx, name, blueprint)
}

// CreateIfNotExists creates a new table with the given name and blueprint if it does not already exist.
// The blueprint function is used to define the structure of the table.
// It returns an error if the table creation fails.
//
// Example:
//
//	err := schema.CreateIfNotExists(ctx, tx, "users", func(table *schema.Blueprint) {
//	    table.ID()
//	    table.String("name").Nullable(false)
//	    table.String("email").Unique().Nullable(false)
//	    table.String("password").Nullable()
//	    table.Timestamp("created_at").Default("CURRENT_TIMESTAMP").Nullable(false)
//	    table.Timestamp("updated_at").Default("CURRENT_TIMESTAMP").Nullable(false)
//	})
func CreateIfNotExists(ctx context.Context, tx *sql.Tx, name string, blueprint func(table *Blueprint)) error {
	builder, err := newBuilder()
	if err != nil {
		return err
	}

	return builder.CreateIfNotExists(ctx, tx, name, blueprint)
}

// Drop removes the table with the given name.
// It returns an error if the table removal fails.
//
// Example:
//
//	err := schema.Drop(ctx, tx, "users")
func Drop(ctx context.Context, tx *sql.Tx, name string) error {
	builder, err := newBuilder()
	if err != nil {
		return err
	}

	return builder.Drop(ctx, tx, name)
}

// DropIfExists removes the table with the given name if it exists.
// It returns an error if the table removal fails.
//
// Example:
//
//	err := schema.DropIfExists(ctx, tx, "users")
func DropIfExists(ctx context.Context, tx *sql.Tx, name string) error {
	builder, err := newBuilder()
	if err != nil {
		return err
	}

	return builder.DropIfExists(ctx, tx, name)
}

// HasTable checks if a table with the given name exists in the database.
// It returns true if the table exists, false otherwise.
// It returns an error if the check fails.
//
// Example:
//
//	exists, err := schema.HasTable(ctx, tx, "users")
func HasTable(ctx context.Context, tx *sql.Tx, name string) (bool, error) {
	builder, err := newBuilder()
	if err != nil {
		return false, err
	}

	return builder.HasTable(ctx, tx, name)
}

// Rename changes the name of the table from name to newName.
// It returns an error if the renaming fails.
//
// Example:
//
//	err := schema.Rename(ctx, tx, "users", "people")
func Rename(ctx context.Context, tx *sql.Tx, name string, newName string) error {
	builder, err := newBuilder()
	if err != nil {
		return err
	}

	return builder.Rename(ctx, tx, name, newName)
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
func Table(ctx context.Context, tx *sql.Tx, name string, blueprint func(table *Blueprint)) error {
	builder, err := newBuilder()
	if err != nil {
		return err
	}

	return builder.Table(ctx, tx, name, blueprint)
}
