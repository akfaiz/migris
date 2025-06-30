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

func execContext(ctx context.Context, tx *sql.Tx, queries ...string) error {
	for _, query := range queries {
		if debug {
			log.Printf("Executing SQL: %s\n", query)
		}
		if _, err := tx.ExecContext(ctx, query); err != nil {
			if debug {
				log.Printf("Error executing SQL: %s\nError: %v\n", query, err)
			}
			return err
		}
	}
	return nil
}

func queryRowContext(ctx context.Context, tx *sql.Tx, query string, args ...any) *sql.Row {
	if debug {
		log.Printf("Executing Query: %s with args: %v\n", query, args)
	}
	row := tx.QueryRowContext(ctx, query, args...)
	if row.Err() != nil {
		if debug {
			log.Printf("Error executing Query: %s\nError: %v\n", query, row.Err())
		}
	}
	return row
}

func queryContext(ctx context.Context, tx *sql.Tx, query string, args ...any) (*sql.Rows, error) {
	if debug {
		log.Printf("Executing Query: %s with args: %v\n", query, args)
	}
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		if debug {
			log.Printf("Error executing Query: %s\nError: %v\n", query, err)
		}
		return nil, err
	}
	return rows, nil
}
