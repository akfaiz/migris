# Migris

[![Go Reference](https://pkg.go.dev/badge/github.com/akfaiz/migris.svg)](https://pkg.go.dev/github.com/akfaiz/migris)
[![Go Report Card](https://goreportcard.com/badge/github.com/akfaiz/migris)](https://goreportcard.com/report/github.com/akfaiz/migris)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/akfaiz/migris/blob/main/LICENSE)

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

Create a CLI tool to manage migrations within your Go application:

```go
package main

import (
    "context"
    "database/sql"
    "log"
    "os"

    "github.com/akfaiz/migris"
    _ "github.com/akfaiz/migris/examples/basic/migrations" // Import migrations directory
    _ "github.com/jackc/pgx/v5/stdlib"
    "github.com/joho/godotenv"
    "github.com/urfave/cli/v3"
)

const migrationDir = "migrations"

func loadDatabaseURL() string {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }
    databaseURL := os.Getenv("DATABASE_URL")
    if databaseURL == "" {
        log.Fatal("DATABASE_URL is not set in the environment")
    }
    return databaseURL
}

func main() {
    databaseURL := loadDatabaseURL()
    db, err := sql.Open("pgx", databaseURL)
    if err != nil {
        log.Fatalf("Failed to open database: %v", err)
    }
    defer db.Close()

    migrator, err := migris.New("pgx", migris.WithDB(db), migris.WithMigrationDir(migrationDir))
    if err != nil {
        log.Fatalf("Failed to create migrator: %v", err)
    }

    cmd := &cli.Command{
        Name:  "migrate",
        Usage: "Migration tool",
        Commands: []*cli.Command{
            {
                Name:  "create",
                Usage: "Create a new migration file",
                Flags: []cli.Flag{
                    &cli.StringFlag{
                        Name:     "name",
                        Aliases:  []string{"n"},
                        Usage:    "Name of the migration",
                        Required: true,
                    },
                },
                Action: func(ctx context.Context, c *cli.Command) error {
                    return migrator.Create(c.String("name"))
                },
            },
            {
                Name:  "up",
                Usage: "Run all pending migrations",
                Action: func(ctx context.Context, c *cli.Command) error {
                    return migrator.UpContext(ctx)
                },
            },
            {
                Name:  "reset",
                Usage: "Rollback all migrations",
                Action: func(ctx context.Context, c *cli.Command) error {
                    return migrator.ResetContext(ctx)
                },
            },
            {
                Name:  "down",
                Usage: "Rollback the last migration",
                Action: func(ctx context.Context, c *cli.Command) error {
                    return migrator.DownContext(ctx)
                },
            },
            {
                Name:  "status",
                Usage: "Show the status of migrations",
                Action: func(ctx context.Context, c *cli.Command) error {
                    return migrator.StatusContext(ctx)
                },
            },
        },
    }
    if err := cmd.Run(context.Background(), os.Args); err != nil {
        log.Printf("Error running app: %v\n", err)
        os.Exit(1)
    }
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