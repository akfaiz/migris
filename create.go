package migris

import (
	"text/template"

	"github.com/akfaiz/migris/internal/parser"
	"github.com/pressly/goose/v3"
)

// Create creates a new migration file with the given name in the specified directory.
func (m *Migrate) Create(name string) error {
	tmpl := getMigrationTemplate(name)
	return goose.CreateWithTemplate(nil, m.migrationDir, tmpl, name, "go")
}

func getMigrationTemplate(name string) *template.Template {
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
	"github.com/akfaiz/migris"
	"github.com/akfaiz/migris/schema"
)

func init() {
	migris.AddMigrationContext(up{{.CamelName}}, down{{.CamelName}})
}

func up{{.CamelName}}(c *schema.Context) error {
	// This code is executed when the migration is applied.
	return nil
}

func down{{.CamelName}}(c *schema.Context) error {
	// This code is executed when the migration is rolled back.
	return nil
}
`))

func migrationCreateTemplate(table string) *template.Template {
	tmpl := `package migrations

import (
	"github.com/akfaiz/migris"
	"github.com/akfaiz/migris/schema"
)

func init() {
	migris.AddMigrationContext(up{{.CamelName}}, down{{.CamelName}})
}

func up{{.CamelName}}(c *schema.Context) error {
	return schema.Create(c, "` + table + `", func(table *schema.Blueprint) {
		// Define your table schema here
	})
}

func down{{.CamelName}}(c *schema.Context) error {
	return schema.DropIfExists(c, "` + table + `")
}
`
	return template.Must(template.New("migration-create").Parse(tmpl))
}

func migrationUpdateTemplate(table string) *template.Template {
	tmpl := `package migrations

import (
	"github.com/akfaiz/migris"
	"github.com/akfaiz/migris/schema"
)

func init() {
	migris.AddMigrationContext(up{{.CamelName}}, down{{.CamelName}})
}

func up{{.CamelName}}(c *schema.Context) error {
	return schema.Table(c, "` + table + `", func(table *schema.Blueprint) {
		// Define your table schema changes here
	})
}

func down{{.CamelName}}(c *schema.Context) error {
	return schema.Table(c, "` + table + `", func(table *schema.Blueprint) {
		// Define your table schema changes here
	})
}
`
	return template.Must(template.New("migration-update").Parse(tmpl))
}
