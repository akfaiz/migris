# Migris

**Migris** is a database migration library for Go, inspired by Laravel's migrations.  
It combines the power of [pressly/goose](https://github.com/pressly/goose) with a fluent schema builder, making migrations easy to write, run, and maintain.

## ✨ Features

- 📦 Migration management (`up`, `down`, `reset`, `status`, `create`)
- 🏗️ Fluent schema builder (similar to Laravel migrations)
- 🗄️ Supports PostgreSQL, MySQL, and MariaDB
- 🔄 Transaction-based migrations
- 🛠️ Integration with Go projects (no external CLI required)

## 🚀 Installation

```bash
go get github.com/afkdevs/migris
```

## 📚 Usage

### 1. Create a Migration

Migrations are defined in Go files using the schema builder:

```go
package migrations

import (
    "github.com/afkdevs/migris"
    "github.com/afkdevs/migris/schema"
)

func init() {
    migris.AddMigrationContext(upCreateUsersTable, downCreateUsersTable)
}

func upCreateUsersTable(c *schema.Context) error {
    return schema.Create(c, "users", func(table *schema.Blueprint) {
        table.ID()
        table.String("name")
        table.String("email")
        table.Timestamp("email_verified_at").Nullable()
        table.String("password")
        table.Timestamps()
    })
}

func downCreateUsersTable(c *schema.Context) error {
    return schema.DropIfExists(c, "users")
}
```

This creates a `users` table with common fields.

### 2. Run Migrations

You can manage migrations directly from Go code:

```go
package migrate

import (
	"database/sql"
	"fmt"

	"github.com/afkdevs/migris"
	"github.com/afkdevs/migris/examples/basic/config"
	_ "github.com/afkdevs/migris/examples/basic/migrations"
	_ "github.com/lib/pq" // PostgreSQL driver
)

func Up() error {
	m, err := newMigrate()
	if err != nil {
		return err
	}
	return m.Up()
}

func Create(name string) error {
	m, err := newMigrate()
	if err != nil {
		return err
	}
	return m.Create(name)
}

func Reset() error {
	m, err := newMigrate()
	if err != nil {
		return err
	}
	return m.Reset()
}

func Down() error {
	m, err := newMigrate()
	if err != nil {
		return err
	}
	return m.Down()
}

func Status() error {
	m, err := newMigrate()
	if err != nil {
		return err
	}
	return m.Status()
}

func newMigrate() (*migris.Migrate, error) {
	if err := migris.SetDialect("postgres"); err != nil {
		return nil, fmt.Errorf("failed to set schema dialect: %w", err)
	}
	dsn := "postgres://user:pass@localhost:5432/mydb?sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	return migris.New(db, "migrations"), nil
}
```

## 🔧 Commands

Here are the available migration commands:

| Function          | Description                                |
|-------------------|--------------------------------------------|
| `migris.Up`       | Apply all pending migrations               |
| `migris.Down`     | Rollback the last migration                |
| `migris.Reset`    | Rollback all migrations                    |
| `migris.Status`   | Show migration status                      |
| `migris.Create`   | Create a new migration file with timestamp |

## 🛠️ Example Schema

```go
schema.Create(c, "posts", func(table *schema.Blueprint) {
    table.ID()
    table.String("title")
    table.Text("body")
    table.ForeignID("user_id").Constrained("users")
    table.Timestamps()
})
```

## 📖 Roadmap

- [ ] Add dry-run mode
- [ ] Add SQLite support
- [ ] CLI wrapper for quick usage

## 📄 License

MIT License.  
See [LICENSE](./LICENSE) for details.