package migrations

import (
	"context"
	"database/sql"

	"github.com/afkdevs/go-schema"
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
	return schema.DropIfExists(ctx, tx, "users")
}
