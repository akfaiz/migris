# Schema
[![Go](https://github.com/afkdevs/go-schema/actions/workflows/ci.yml/badge.svg)](https://github.com/afkdevs/go-schema/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/afkdevs/go-schema)](https://goreportcard.com/report/github.com/afkdevs/go-schema)
[![codecov](https://codecov.io/gh/afkdevs/go-schema/graph/badge.svg?token=7tbSVRaD4b)](https://codecov.io/gh/afkdevs/go-schema)
[![GoDoc](https://pkg.go.dev/badge/github.com/afkdevs/go-schema)](https://pkg.go.dev/github.com/afkdevs/go-schema)
[![Go Version](https://img.shields.io/github/go-mod/go-version/afkdevs/go-schema)](https://golang.org/doc/devel/release.html)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

`Schema` is a simple Go library for building and running SQL schema (DDL) code in a clean, readable, and migration-friendly way. Inspired by Laravel's Schema Builder, it helps you easily create or change database tables—and works well with tools like [`goose`](https://github.com/pressly/goose).

## Features

- 📊 Programmatic table and column definitions
- 🗃️ Support for common data types and constraints
- ⚙️ Auto-generates `CREATE TABLE`, `ALTER TABLE`, index and foreign key SQL
- 🔀 Designed to work with database transactions
- 🧪 Built-in types and functions make migration code clear and testable
- 🔍 Provides helper functions to get list tables, columns, and indexes

## Supported Databases

Currently, `schema` is tested and optimized for:

* PostgreSQL
* MySQL / MariaDB
* SQLite (TODO)

## Installation

```bash
go get github.com/afkdevs/schema
```

## Integration Example (with goose)
```go
package migrations

import (
	"context"
	"database/sql"

	"github.com/afkdevs/schema"
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
		table.Timestamps()
	})
}

func downCreateUsersTable(ctx context.Context, tx *sql.Tx) error {
	return schema.Drop(ctx, tx, "users")
}
```
For more examples, check out the [examples](examples/basic) directory.

## Documentation
For detailed documentation, please refer to the [GoDoc](https://pkg.go.dev/github.com/afkdevs/schema) page.

## Contributing
Contributions are welcome! Please read the [contributing guidelines](CONTRIBUTING.md) and submit a pull request.

## License
This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.