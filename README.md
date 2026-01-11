# Migris

[![Go Reference](https://pkg.go.dev/badge/github.com/akfaiz/migris.svg)](https://pkg.go.dev/github.com/akfaiz/migris)
[![Go Report Card](https://goreportcard.com/badge/github.com/akfaiz/migris)](https://goreportcard.com/report/github.com/akfaiz/migris)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/akfaiz/migris/blob/main/LICENSE)

**Migris** is a database migration library for Go, inspired by Laravel's migrations. It combines the power of [pressly/goose](https://github.com/pressly/goose) with a fluent schema builder, making migrations easy to write, run, and maintain.

## Features

- **Migration management** - Run up, down, reset, status, and create operations
- **Dry-run mode** - Preview migrations without executing them to see generated SQL
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

func upCreateUsersTable(c schema.Context) error {
    return schema.Create(c, "users", func(table *schema.Blueprint) {
        table.ID()
        table.String("name")
        table.String("email").Unique()
        table.Timestamp("email_verified_at").Nullable()
        table.String("password")
        table.Timestamps()
    })
}

func downCreateUsersTable(c schema.Context) error {
    return schema.DropIfExists(c, "users")
}
```

### Running Migrations

For a complete CLI setup example, see [examples/basic](examples/basic/). For quick setup, use the CLI helpers below.

### CLI Helpers

For simpler CLI integration, use the pre-built CLI helpers:

#### Using urfave/cli

```go
package main

import (
    "context"
    "database/sql"
    "log"
    "os"

    "github.com/akfaiz/migris/extra/migriscli"
    _ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
    db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    cfg := migriscli.Config{
        DB:            db,
        Dialect:       "pgx",
        MigrationsDir: "./migrations",
    }

    cmd := migriscli.NewCLI(cfg)
    if err := cmd.Run(context.Background(), os.Args); err != nil {
        log.Fatal(err)
    }
}
```

#### Using Cobra

```go
package main

import (
    "database/sql"
    "log"
    "os"

    "github.com/akfaiz/migris/extra/migriscobra"
    _ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
    db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    cfg := migriscobra.Config{
        DB:            db,
        Dialect:       "pgx",
        MigrationsDir: "./migrations",
    }

    cmd := migriscobra.NewCLI(cfg)
    if err := cmd.Execute(); err != nil {
        log.Fatal(err)
    }
}
```

Both CLI helpers support all migration commands: `create`, `up`, `up-to`, `down`, `down-to`, `reset`, `status` with `--dry-run` support.

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

### Dry-Run Mode

Preview migrations without executing them:

```bash
# Preview pending migrations
go run main.go --dry-run up

# Preview rollback operations
go run main.go --dry-run down
go run main.go --dry-run reset
```

Dry-run mode shows:

- Which migrations would be executed
- The exact SQL statements that would be generated
- Execution timing and summary statistics
- Clear indication that no database changes are made

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
