# Migris

**Migris** is a database migration library for Go, inspired by Laravel's migrations. It combines the power of [pressly/goose](https://github.com/pressly/goose) with a fluent schema builder, making migrations easy to write, run, and maintain.

## Features

- **Migration management** - Run up, down, reset, status, and create operations
- **Fluent schema builder** - Laravel-inspired API for defining database schemas
- **Multi-database support** - Works with PostgreSQL, MySQL, and MariaDB
- **Transaction safety** - All migrations run within database transactions
- **Native Go integration** - No external CLI tools required

## Installation

```bash
go get -u github.com/akfaiz/migris
```

## Quick Start

### Creating Migrations

Define migrations using the fluent schema builder API:

```go
package migrations

import (
    "github.com/akfaiz/migris"
    "github.com/akfaiz/migris/schema"
)

func init() {
    migris.AddMigrationContext(upCreateUsersTable, downCreateUsersTable)
}

func upCreateUsersTable(c *schema.Context) error {
    return schema.Create(c, "users", func(table *schema.Blueprint) {
        table.ID()
        table.String("name")
        table.String("email").Unique()
        table.Timestamp("email_verified_at").Nullable()
        table.String("password")
        table.Timestamps()
    })
}

func downCreateUsersTable(c *schema.Context) error {
    return schema.DropIfExists(c, "users")
}
```

### Running Migrations

Execute migrations programmatically within your Go application:

```go
package main

import (
    "database/sql"
    "log"

    "github.com/akfaiz/migris"
    _ "migrations" // Import migrations directory
    _ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
    db, err := sql.Open("pgx", "postgres://user:password@localhost:5432/mydb?sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    m, err := migris.New("pgx", migris.WithDB(db), migris.WithMigrationDir("migrations"))
    if err != nil {
        log.Fatal(err)
    }
    
    if err := m.Up(); err != nil {
        log.Fatal(err)
    }
    
    log.Println("Migrations completed successfully")
}
```

## Schema Builder API

The schema builder provides a fluent interface for defining database schemas:

```go
// Creating tables
schema.Create(c, "posts", func(table *schema.Blueprint) {
    table.ID()
    table.String("title")
    table.Text("content")
    table.UnsignedBigInteger("user_id")
    table.Boolean("published").Default(false)
    table.Timestamps()
    
    // Foreign key constraints
    table.Foreign("user_id").References("id").On("users")
    
    // Indexes
    table.Index([]string{"title", "published"})
})

// Modifying existing tables
schema.Table(c, "posts", func(table *schema.Blueprint) {
    table.String("slug")
    table.DropColumn("old_column")
})
```

## Migration Operations

Migris supports all standard migration operations:

```go
migrator.Up()           // Run all pending migrations
migrator.Down()         // Rollback the last migration
migrator.Reset()        // Rollback all migrations
migrator.Status()       // Show migration status
migrator.Create(name)   // Create a new migration file
```

## Database Support

Currently supported databases:
- **PostgreSQL** (via pgx driver)
- **MySQL** 
- **MariaDB**

## Roadmap

- [ ] SQLite support
- [ ] Advanced schema introspection

## Contributing

Contributions are welcome! Please feel free to submit issues, feature requests, or pull requests.

## License

Released under the MIT License. See [LICENSE](./LICENSE) for details.