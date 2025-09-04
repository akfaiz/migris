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
		table.Increments("id").Primary()
		table.String("username")
		table.String("email").Unique()
		table.Timestamp("created_at").UseCurrent()
	})
}

func downCreateUsersTable(c *schema.Context) error {
	return schema.DropIfExists(c, "users")
}
