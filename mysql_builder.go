package schema

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/afkdevs/go-schema/internal/config"
)

type mysqlBuilder struct {
	baseBuilder
	grammar *mysqlGrammar
}

var _ Builder = (*mysqlBuilder)(nil)

func newMysqlBuilder() Builder {
	grammar := newMysqlGrammar()
	cfg := config.Get()

	return &mysqlBuilder{
		baseBuilder: baseBuilder{grammar: grammar, verbose: cfg.Verbose},
		grammar:     grammar,
	}
}

func (b *mysqlBuilder) getCurrentDatabase(ctx context.Context, tx *sql.Tx) (string, error) {
	query := b.grammar.CompileCurrentDatabase()
	row := tx.QueryRowContext(ctx, query)
	var dbName string
	if err := row.Scan(&dbName); err != nil {
		return "", err
	}
	return dbName, nil
}

func (b *mysqlBuilder) GetColumns(ctx context.Context, tx *sql.Tx, tableName string) ([]*Column, error) {
	if tx == nil || tableName == "" {
		return nil, errors.New("invalid arguments: transaction is nil or table name is empty")
	}

	database, err := b.getCurrentDatabase(ctx, tx)
	if err != nil {
		return nil, err
	}

	query, err := b.grammar.CompileColumns(database, tableName)
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	var columns []*Column
	for rows.Next() {
		var col Column
		var nullableStr string
		if err := rows.Scan(
			&col.Name, &col.TypeName, &col.TypeFull,
			&col.Collation, &nullableStr,
			&col.DefaultVal, &col.Comment,
			&col.Extra,
		); err != nil {
			return nil, err
		}
		if nullableStr == "YES" {
			col.Nullable = true
		}
		columns = append(columns, &col)
	}
	return columns, nil
}

func (b *mysqlBuilder) GetIndexes(ctx context.Context, tx *sql.Tx, tableName string) ([]*Index, error) {
	if tx == nil || tableName == "" {
		return nil, errors.New("invalid arguments: transaction is nil or table name is empty")
	}

	database, err := b.getCurrentDatabase(ctx, tx)
	if err != nil {
		return nil, err
	}

	query, err := b.grammar.CompileIndexes(database, tableName)
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	var indexes []*Index
	for rows.Next() {
		var idx Index
		var columnsStr string
		if err := rows.Scan(&idx.Name, &columnsStr, &idx.Type, &idx.Unique); err != nil {
			return nil, err
		}
		idx.Columns = strings.Split(columnsStr, ",")
		indexes = append(indexes, &idx)
	}
	return indexes, nil
}

func (b *mysqlBuilder) GetTables(ctx context.Context, tx *sql.Tx) ([]*TableInfo, error) {
	if tx == nil {
		return nil, errors.New("invalid arguments: transaction is nil")
	}

	database, err := b.getCurrentDatabase(ctx, tx)
	if err != nil {
		return nil, err
	}

	query, err := b.grammar.CompileTables(database)
	if err != nil {
		return nil, err
	}
	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	var tables []*TableInfo
	for rows.Next() {
		var table TableInfo
		if err := rows.Scan(&table.Name, &table.Size, &table.Comment, &table.Engine, &table.Collation); err != nil {
			return nil, err
		}
		tables = append(tables, &table)
	}
	return tables, nil
}

func (b *mysqlBuilder) HasColumn(ctx context.Context, tx *sql.Tx, tableName string, columnName string) (bool, error) {
	if tx == nil || columnName == "" {
		return false, errors.New("invalid arguments: transaction is nil or column name is empty")
	}
	return b.HasColumns(ctx, tx, tableName, []string{columnName})
}

func (b *mysqlBuilder) HasColumns(ctx context.Context, tx *sql.Tx, tableName string, columnNames []string) (bool, error) {
	if tx == nil || tableName == "" {
		return false, errors.New("invalid arguments: transaction is nil or table name is empty")
	}
	if len(columnNames) == 0 {
		return false, errors.New("no column names provided")
	}

	columns, err := b.GetColumns(ctx, tx, tableName)
	if err != nil {
		return false, err
	}
	columnsMap := make(map[string]bool)
	for _, col := range columns {
		columnsMap[col.Name] = true
	}
	for _, colName := range columnNames {
		if _, exists := columnsMap[colName]; !exists {
			return false, nil // If any column does not exist, return false
		}
	}
	return true, nil // All specified columns exist
}

func (b *mysqlBuilder) HasIndex(ctx context.Context, tx *sql.Tx, tableName string, indexes []string) (bool, error) {
	if tx == nil || tableName == "" {
		return false, errors.New("invalid arguments: transaction is nil or table name is empty")
	}

	existingIndexes, err := b.GetIndexes(ctx, tx, tableName)
	if err != nil {
		return false, err
	}
	if len(existingIndexes) == 0 {
		return false, nil // No indexes found, so the specified indexes cannot exist
	}
	if len(indexes) == 0 {
		return true, nil // No specific indexes to check, so we assume they exist
	}
	if len(indexes) == 1 {
		for _, idx := range existingIndexes {
			if idx.Name == indexes[0] {
				return true, nil // If any specified index exists, return true
			}
		}
	}
	indexColumns := make(map[string]bool)
	for _, idx := range indexes {
		indexColumns[idx] = true // Create a map of specified indexes for quick lookup
	}

	for _, index := range existingIndexes {
		found := true
		for _, indexCol := range index.Columns {
			if _, exists := indexColumns[indexCol]; !exists {
				found = false // If any column in the index does not match the specified indexes, set found to false
				break
			}
		}
		// If all columns in the index match the specified indexes, we found a match
		if found {
			return true, nil
		}
	}

	return false, nil // If no specified index exists, return false
}

func (b *mysqlBuilder) HasTable(ctx context.Context, tx *sql.Tx, name string) (bool, error) {
	if tx == nil || name == "" {
		return false, errors.New("invalid arguments: transaction is nil or table name is empty")
	}

	database, err := b.getCurrentDatabase(ctx, tx)
	if err != nil {
		return false, err
	}

	query, err := b.grammar.CompileTableExists(database, name)
	if err != nil {
		return false, err
	}

	row := tx.QueryRowContext(ctx, query)
	var exists bool
	if err := row.Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil // Table does not exist
		}
		return false, err // Other error occurred
	}
	return exists, nil // Return true if the table exists
}
