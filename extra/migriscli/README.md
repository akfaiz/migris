# Migris CLI Helper

A simple CLI interface for the migris database migration tool using [urfave/cli/v3](https://github.com/urfave/cli).

## Installation

```bash
go get github.com/akfaiz/migris/extra/migriscli
```

## Usage

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
    db, err := sql.Open("pgx", "postgres://user:pass@localhost/mydb")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    cfg := migriscli.Config{
        DB:            db,
        Dialect:       "pgx",
        MigrationsDir: "./migrations",
    }

    cmd := migriscli.NewCommand(cfg)
    if err := cmd.Run(context.Background(), os.Args); err != nil {
        log.Fatal(err)
    }
}
```

## Commands

- `create --name <name>` - Create a new migration file
- `up` - Apply all pending migrations
- `up-to --version <version>` - Apply migrations up to specific version
- `down` - Rollback the last migration
- `down-to --version <version>` - Rollback to specific version
- `reset` - Rollback all migrations
- `status` - Show migration status

All migration commands support `--dry-run` to preview changes without executing them.

## Configuration

```go
type Config struct {
    DB            *sql.DB  // Database connection
    Dialect       string   // "pgx", "mysql", or "maria"
    MigrationsDir string   // Migration files directory
}
```
