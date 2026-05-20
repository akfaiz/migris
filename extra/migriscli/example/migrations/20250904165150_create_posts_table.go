package migrations

import (
	"github.com/akfaiz/migris"
	"github.com/akfaiz/migris/schema"
)

func init() {
	migris.AddMigrationContext(upCreatePostsTable, downCreatePostsTable)
}

func upCreatePostsTable(c schema.Context) error {
	return schema.Create(c, "posts", func(table *schema.Blueprint) {
		table.Increments("id").Primary()
		table.String("title")
		table.Text("content")
		table.Integer("author_id").Unsigned()
		table.Timestamp("created_at").UseCurrent()
		table.Foreign("author_id").References("id").On("users")
	})
}

func downCreatePostsTable(c schema.Context) error {
	return schema.DropIfExists(c, "posts")
}
