package migrations

import (
	"github.com/akfaiz/migris"
	"github.com/akfaiz/migris/schema"
)

func init() {
	migris.AddMigrationContext(upCreateCommentsTable, downCreateCommentsTable)
}

func upCreateCommentsTable(c *schema.Context) error {
	return schema.Create(c, "comments", func(table *schema.Blueprint) {
		table.Increments("id").Primary()
		table.Integer("post_id").Unsigned()
		table.Integer("user_id").Unsigned()
		table.Text("content")
		table.Timestamp("created_at").UseCurrent()
		table.Foreign("post_id").References("id").On("posts")
		table.Foreign("user_id").References("id").On("users")
	})
}

func downCreateCommentsTable(c *schema.Context) error {
	return schema.DropIfExists(c, "comments")
}
