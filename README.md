# Schema

Schema is a fluent, expressive Go library for generating and executing DDL SQL (e.g., `CREATE TABLE`) within Go code — ideal for use with migration tools like [Goose](https://github.com/pressly/goose).

Inspired by Laravel's schema builder, Schema simplifies defining and evolving database schemas using idiomatic Go.


## ✨ Features

- ✅ Programmatic DDL builder for Go
- 🧱 Supports `CREATE TABLE`, columns, primary keys, unique constraints, default values, and nullable fields
- 🔄 Works seamlessly with Goose and other transaction-based migration tools
- 🧩 Clean, fluent API for easy schema design

## Supported Databases

- PostgreSQL
- MySQL (TODO)
- MariaDB (TODO)
- SQLite (TODO)

## Installation

```bash
go get github.com/ahmadfaizk/schema
```

## Usage
```go
package migrations

import (
	"context"
	"database/sql"

	"github.com/ahmadfaizk/schema"
	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upCreateUsersTable, downCreateUsersTable)
}

func upCreateUsersTable(ctx context.Context, tx *sql.Tx) error {
	return schema.Create(ctx, tx, "users", func(table *schema.Blueprint) {
		table.ID()
		table.String("name")
		table.String("email")
		table.Timestamp("email_verified_at").Nullable()
		table.String("password")
		table.Timestamp("created_at").Default("CURRENT_TIMESTAMP")
		table.Timestamp("updated_at").Default("CURRENT_TIMESTAMP")
	})
}

func downCreateUsersTable(ctx context.Context, tx *sql.Tx) error {
	return schema.Drop(ctx, tx, "users")
}

```