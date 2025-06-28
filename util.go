package schema

import (
	"context"
	"database/sql"
	"log"
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
		if debug {
			log.Printf("Executing SQL: %s\n", sql)
		}
		if _, err := tx.ExecContext(ctx, sql); err != nil {
			if debug {
				log.Printf("Error executing SQL: %s\nError: %v\n", sql, err)
			}
			return err
		}
	}
	return nil
}
