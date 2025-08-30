package migrations

import (
	"github.com/akfaiz/migris"
	"github.com/akfaiz/migris/schema"
)

func init() {
	migris.AddMigrationContext(upCreateUserRolesTable, downCreateUserRolesTable)
}

func upCreateUserRolesTable(c *schema.Context) error {
	return schema.Create(c, "user_roles", func(table *schema.Blueprint) {
		table.BigInteger("user_id")
		table.Integer("role_id")
		table.Primary("user_id", "role_id")
		table.Foreign("user_id").References("id").On("users")
		table.Foreign("role_id").References("id").On("roles")
	})
}

func downCreateUserRolesTable(c *schema.Context) error {
	return schema.DropIfExists(c, "user_roles")
}
