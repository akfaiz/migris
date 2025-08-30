package schema

import (
	"text/template"

	"github.com/afkdevs/go-schema/internal/parser"
)

func GooseMigrationTemplate(name string) *template.Template {
	tableName, create := parser.ParseMigrationName(name)
	if create {
		return migrationCreateTemplate(tableName)
	}
	if tableName != "" {
		return migrationUpdateTemplate(tableName)
	}
	return migrationTemplate
}

var migrationTemplate = template.Must(template.New("migrator.go-migration").Parse(`package migrations

import (
	"context"
	"database/sql"

	"github.com/afkdevs/go-schema"
	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(up{{.CamelName}}, down{{.CamelName}})
}

func up{{.CamelName}}(ctx context.Context, tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	return nil
}

func down{{.CamelName}}(ctx context.Context, tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	return nil
}
`))

func migrationCreateTemplate(table string) *template.Template {
	tmpl := `package migrations

import (
	"context"
	"database/sql"

	"github.com/afkdevs/go-schema"
	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(up{{.CamelName}}, down{{.CamelName}})
}

func up{{.CamelName}}(ctx context.Context, tx *sql.Tx) error {
	return schema.Create(ctx, tx, "` + table + `", func(table *schema.Blueprint) {
		// Define your table schema here
	})
}

func down{{.CamelName}}(ctx context.Context, tx *sql.Tx) error {
	return schema.DropIfExists(ctx, tx, "` + table + `")
}
`
	return template.Must(template.New("migration-create").Parse(tmpl))
}

func migrationUpdateTemplate(table string) *template.Template {
	tmpl := `package migrations

import (
	"context"
	"database/sql"

	"github.com/afkdevs/go-schema"
	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(up{{.CamelName}}, down{{.CamelName}})
}

func up{{.CamelName}}(ctx context.Context, tx *sql.Tx) error {
	return schema.Table(ctx, tx, "` + table + `", func(table *schema.Blueprint) {
		// Define your table schema changes here
	})
}

func down{{.CamelName}}(ctx context.Context, tx *sql.Tx) error {
	return schema.Table(ctx, tx, "` + table + `", func(table *schema.Blueprint) {
		// Define your table schema changes here
	})
}
`
	return template.Must(template.New("migration-update").Parse(tmpl))
}
