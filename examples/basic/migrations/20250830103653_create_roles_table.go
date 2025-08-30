package migrations

import (
	"context"
	"database/sql"

	"github.com/afkdevs/migris"
	"github.com/afkdevs/migris/schema"
)

func init() {
	migris.AddMigrationContext(upCreateRolesTable, downCreateRolesTable)
}

func upCreateRolesTable(ctx context.Context, tx *sql.Tx) error {
	return schema.Create(ctx, tx, "roles", func(table *schema.Blueprint) {
		table.Increments("id").Primary()
		table.String("name").Unique().Nullable(false)
	})
}

func downCreateRolesTable(ctx context.Context, tx *sql.Tx) error {
	return schema.DropIfExists(ctx, tx, "roles")
}
