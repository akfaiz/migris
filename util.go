package schema

import (
	"context"
	"database/sql"
	"log"
)

func optional[T any](defaultValue T, values ...T) T {
	return optionalAtIndex(0, defaultValue, values...)
}

func optionalAtIndex[T any](index int, defaultValue T, values ...T) T {
	if index < len(values) {
		return values[index]
	}
	return defaultValue
}

func execContext(ctx context.Context, tx *sql.Tx, queries ...string) error {
	for _, query := range queries {
		if debug {
			log.Printf("Executing SQL: %s\n", query)
		}
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

func queryRowContext(ctx context.Context, tx *sql.Tx, query string, args ...any) *sql.Row {
	if debug {
		log.Printf("Executing Query: %s with args: %v\n", query, args)
	}
	return tx.QueryRowContext(ctx, query, args...)
}

func queryContext(ctx context.Context, tx *sql.Tx, query string, args ...any) (*sql.Rows, error) {
	if debug {
		log.Printf("Executing Query: %s with args: %v\n", query, args)
	}
	return tx.QueryContext(ctx, query, args...)
}
