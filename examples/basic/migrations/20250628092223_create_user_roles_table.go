package migrations

import (
	"context"
	"database/sql"

	"github.com/afkdevs/go-schema"
	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upCreateUserRolesTable, downCreateUserRolesTable)
}

func upCreateUserRolesTable(ctx context.Context, tx *sql.Tx) error {
	return schema.Create(ctx, tx, "user_roles", func(table *schema.Blueprint) {
		table.BigInteger("user_id")
		table.Integer("role_id")
		table.Primary("user_id", "role_id")
		table.Foreign("user_id").References("id").On("users")
		table.Foreign("role_id").References("id").On("roles")
	})
}

func downCreateUserRolesTable(ctx context.Context, tx *sql.Tx) error {
	return schema.DropIfExists(ctx, tx, "user_roles")
}
