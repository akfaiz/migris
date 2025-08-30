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
go get -u github.com/akfaiz/migris
```

## 📚 Usage

### 1. Create a Migration

Migrations are defined in Go files using the schema builder:

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

	"github.com/akfaiz/migris"
	_ "migrations" // Import migrations
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
	dsn := "postgres://user:pass@localhost:5432/mydb?sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	return migris.New("postgres", migris.WithDB(db), migris.WithMigrationDir("migrations")), nil
}
```

## 📖 Roadmap

- [ ] Add SQLite support
- [ ] CLI wrapper for quick usage

## 📄 License

MIT License.  
See [LICENSE](./LICENSE) for details.