package migrations

import (
	"context"
	"database/sql"

	"github.com/afkdevs/migris"
	"github.com/afkdevs/migris/schema"
)

func init() {
	migris.AddMigrationContext(upCreateUsersTable, downCreateUsersTable)
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
	return schema.DropIfExists(ctx, tx, "users")
}
