package migrations

import (
	"github.com/afkdevs/migris"
	"github.com/afkdevs/migris/schema"
)

func init() {
	migris.AddMigrationContext(upCreateRolesTable, downCreateRolesTable)
}

func upCreateRolesTable(c *schema.Context) error {
	return schema.Create(c, "roles", func(table *schema.Blueprint) {
		table.Increments("id").Primary()
		table.String("name").Unique().Nullable(false)
	})
}

func downCreateRolesTable(c *schema.Context) error {
	return schema.DropIfExists(c, "roles")
}
