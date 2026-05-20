# Migris

[![Go Reference](https://pkg.go.dev/badge/github.com/akfaiz/migris.svg)](https://pkg.go.dev/github.com/akfaiz/migris)
[![Go Report Card](https://goreportcard.com/badge/github.com/akfaiz/migris)](https://goreportcard.com/report/github.com/akfaiz/migris)
[![codecov](https://codecov.io/gh/akfaiz/migris/graph/badge.svg?token=7tbSVRaD4b)](https://codecov.io/gh/akfaiz/migris)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/akfaiz/migris/blob/main/LICENSE)

**Migris** is a database migration library for Go, inspired by Laravel's migrations. It combines the power of [pressly/goose](https://github.com/pressly/goose) with a fluent schema builder, making migrations easy to write, run, and maintain.

## Features

- **Migration management** - Run up, down, reset, status, and create operations
- **Dry-run mode** - Preview migrations without executing them to see generated SQL
- **Fluent schema builder** - Laravel-inspired API for defining database schemas
- **Multi-database support** - Works with PostgreSQL, MySQL, MariaDB, and SQLite3
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

For a complete stdlib-only example, see [examples/basic](examples/basic/). For quick setup with a CLI framework, use the helpers below.

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
        AllowMissing:  false, // set true to allow out-of-order migrations
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
        AllowMissing:  false, // set true to allow out-of-order migrations
    }

    cmd := migriscobra.NewCLI(cfg)
    if err := cmd.Execute(); err != nil {
        log.Fatal(err)
    }
}
```

Both CLI helpers support all migration commands: `create`, `up`, `up-to`, `down`, `down-to`, `reset`, `status`. The `--dry-run` flag is available per-command. `AllowMissing` in `Config` applies to all run operations.

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

## Options

Migris uses two distinct option types:

### `MigrisOption` — constructor options

Passed to `New()` to configure the migrator instance:

```go
m, err := migris.New("pgx",
    migris.WithDB(db),
    migris.WithMigrationDir("./migrations"),
    migris.WithTableName("schema_migrations"),
    migris.WithRegistry(registry),
)
```

| Option | Description |
|---|---|
| `WithDB(db)` | Existing `*sql.DB` connection (caller manages lifecycle) |
| `WithDSN(dsn)` | Open a new connection from DSN string (migrator manages lifecycle, call `Close()`) |
| `WithMigrationDir(dir)` | Directory for SQL migration files (default: `"migrations"`) |
| `WithTableName(name)` | Migrations tracking table name (default: `"schema_migrations"`) |
| `WithRegistry(r)` | Use an isolated migration registry instead of the global one |

`WithDB` and `WithDSN` are mutually exclusive; `WithDB` takes precedence if both are set.

When using `WithDSN`, the migrator opens the connection internally and owns it. Call `Close()` when done:

```go
m, err := migris.New("pgx",
    migris.WithDSN("postgres://user:pass@localhost:5432/mydb"),
    migris.WithMigrationDir("./migrations"),
)
if err != nil {
    log.Fatal(err)
}
defer m.Close()

if err := m.Up(); err != nil {
    log.Fatal(err)
}
```

When using `WithDB`, the caller owns the connection lifecycle:

```go
db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
if err != nil {
    log.Fatal(err)
}
defer db.Close()

m, err := migris.New("pgx",
    migris.WithDB(db),
    migris.WithMigrationDir("./migrations"),
)
```

### `Option` — run-time options

Passed per-operation to control execution behaviour:

```go
// Dry-run: print SQL without applying changes
migrator.Up(migris.WithDryRun(true))
migrator.Down(migris.WithDryRun(true))
migrator.Reset(migris.WithDryRun(true))

// Allow out-of-order migrations
migrator.Up(migris.WithAllowMissing(true))

// Combine options
migrator.UpTo(20250101000005, migris.WithDryRun(true), migris.WithAllowMissing(true))
```

| Option | Description |
|---|---|
| `WithDryRun(bool)` | Preview SQL without executing |
| `WithAllowMissing(bool)` | Apply out-of-order migrations instead of failing |

All operation methods accept run-time options:

```go
migrator.Up(opts...)
migrator.UpTo(version, opts...)
migrator.Down(opts...)
migrator.DownTo(version, opts...)
migrator.Reset(opts...)

// Context variants
migrator.UpContext(ctx, opts...)
migrator.UpToContext(ctx, version, opts...)
migrator.DownContext(ctx, opts...)
migrator.DownToContext(ctx, version, opts...)
migrator.ResetContext(ctx, opts...)

// No run-time options
migrator.Status()
migrator.Create(name)
```

### Dry-Run Mode

Preview migrations without executing them:

```go
// In code
if err := migrator.Up(migris.WithDryRun(true)); err != nil {
    log.Fatal(err)
}
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
- **SQLite3**

## Contributing

Contributions are welcome! Please feel free to submit issues, feature requests, or pull requests.

## License

Released under the MIT License. See [LICENSE](./LICENSE) for details.
