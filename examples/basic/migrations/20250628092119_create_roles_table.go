package migrations

import (
	"context"
	"database/sql"

	"github.com/afkdevs/go-schema"
	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upCreateRolesTable, downCreateRolesTable)
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
