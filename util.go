package schema

import (
	"context"
	"database/sql"
	"log/slog"
)

func optionalInt(defaultValue int, values ...int) int {
	if len(values) > 0 {
		return values[0]
	}
	return defaultValue
}

func optionalString(defaultValue string, values ...string) string {
	if len(values) > 0 {
		return values[0]
	}
	return defaultValue
}

func optionalBool(defaultValue bool, values ...bool) bool {
	if len(values) > 0 {
		return values[0]
	}
	return defaultValue
}

func execContext(ctx context.Context, tx *sql.Tx, sqls ...string) error {
	for _, sql := range sqls {
		slog.DebugContext(ctx, "Executing SQL", "sql", sql)
		if _, err := tx.ExecContext(ctx, sql); err != nil {
			slog.ErrorContext(ctx, "Failed to execute SQL", "error", err)
			return err
		}
	}
	return nil
}
