package migrations

import (
	"context"
	"database/sql"

	"github.com/ahmadfaizk/schema"
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
		table.Timestamp("created_at").Default("CURRENT_TIMESTAMP")
		table.Timestamp("updated_at").Default("CURRENT_TIMESTAMP")
	})
}

func downCreateUsersTable(ctx context.Context, tx *sql.Tx) error {
	return schema.Drop(ctx, tx, "users")
}
