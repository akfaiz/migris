package schema

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type sqliteBuilder struct {
	baseBuilder
}

var _ Builder = (*sqliteBuilder)(nil)

func newSqliteBuilder() Builder {
	grammar := newSqliteGrammar()

	return &sqliteBuilder{
		baseBuilder: baseBuilder{grammar: grammar},
	}
}

func (b *sqliteBuilder) GetColumns(c Context, tableName string) ([]*Column, error) {
	if c == nil || tableName == "" {
		return nil, errors.New("invalid arguments: context is nil or table name is empty")
	}

	query, err := b.grammar.CompileColumns("", tableName)
	if err != nil {
		return nil, err
	}

	rows, err := c.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []*Column
	for rows.Next() {
		var col Column
		var cid int
		var notNullInt int
		var pk int
		if err = rows.Scan(
			&cid, &col.Name, &col.TypeFull,
			&notNullInt, &col.DefaultVal, &pk,
		); err != nil {
			return nil, err
		}

		// SQLite PRAGMA table_info returns: cid, name, type, notnull, dflt_value, pk
		col.TypeName = col.TypeFull
		col.Nullable = notNullInt == 0 // In SQLite, notnull=1 means NOT NULL, notnull=0 means nullable
		// pk > 0 indicates this column is part of the primary key, but we don't store it in Column struct

		columns = append(columns, &col)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return columns, nil
}

//nolint:gocognit
func (b *sqliteBuilder) GetIndexes(c Context, tableName string) ([]*Index, error) {
	if c == nil || tableName == "" {
		return nil, errors.New("invalid arguments: context is nil or table name is empty")
	}

	query, err := b.grammar.CompileIndexes("", tableName)
	if err != nil {
		return nil, err
	}

	rows, err := c.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indexes []*Index
	for rows.Next() {
		var idx Index
		var uniqueFlag int
		var columnsStr string // Change to string for scanning
		if err = rows.Scan(&idx.Name, &uniqueFlag, &columnsStr); err != nil {
			return nil, err
		}
		idx.Unique = uniqueFlag == 1
		// Parse the columns string if it's not empty
		if columnsStr != "" {
			// Split comma-separated column names (basic parsing)
			idx.Columns = strings.Split(columnsStr, ",")
			for i, col := range idx.Columns {
				idx.Columns[i] = strings.TrimSpace(col)
			}
		}
		indexes = append(indexes, &idx)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// For each index, get detailed column information
	for _, idx := range indexes {
		func() {
			columnQuery := fmt.Sprintf("PRAGMA index_info(%q)", idx.Name)
			var columnRows *sql.Rows
			columnRows, err = c.Query(columnQuery)
			if err != nil {
				return // Skip if we can't get column info
			}

			// Ensure rows are closed when this anonymous function returns
			defer func() {
				if closeErr := columnRows.Close(); closeErr != nil {
					_ = closeErr // Acknowledge the error
				}
			}()

			var columns []string
			for columnRows.Next() {
				var seqno, cid int
				var name string
				if err = columnRows.Scan(&seqno, &cid, &name); err != nil {
					return
				}
				columns = append(columns, name)
			}

			// Check for iteration errors
			if err = columnRows.Err(); err != nil {
				return
			}

			if len(columns) > 0 {
				idx.Columns = columns
			}
		}()
	}

	return indexes, nil
}

func (b *sqliteBuilder) GetTables(c Context) ([]*TableInfo, error) {
	if c == nil {
		return nil, errors.New("invalid arguments: context is nil")
	}

	query, err := b.grammar.CompileTables("")
	if err != nil {
		return nil, err
	}
	rows, err := c.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []*TableInfo
	for rows.Next() {
		var table TableInfo
		if err = rows.Scan(&table.Name, &table.Size, &table.Comment); err != nil {
			return nil, err
		}
		tables = append(tables, &table)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tables, nil
}

func (b *sqliteBuilder) HasColumn(c Context, tableName string, columnName string) (bool, error) {
	if c == nil || columnName == "" {
		return false, errors.New("invalid arguments: context is nil or column name is empty")
	}
	return b.HasColumns(c, tableName, []string{columnName})
}

func (b *sqliteBuilder) HasColumns(c Context, tableName string, columnNames []string) (bool, error) {
	if c == nil || tableName == "" {
		return false, errors.New("invalid arguments: context is nil or table name is empty")
	}
	if len(columnNames) == 0 {
		return false, errors.New("no column names provided")
	}

	columns, err := b.GetColumns(c, tableName)
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

//nolint:dupl // Similar code exists in other builder files
func (b *sqliteBuilder) HasIndex(c Context, tableName string, indexes []string) (bool, error) {
	if c == nil || tableName == "" {
		return false, errors.New("invalid arguments: context is nil or table name is empty")
	}

	existingIndexes, err := b.GetIndexes(c, tableName)
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

func (b *sqliteBuilder) HasTable(c Context, name string) (bool, error) {
	if c == nil || name == "" {
		return false, errors.New("invalid arguments: context is nil or table name is empty")
	}

	query, err := b.grammar.CompileTableExists("", name)
	if err != nil {
		return false, err
	}

	row := c.QueryRow(query)
	var exists bool
	if err = row.Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil // Table does not exist
		}
		return false, err // Other error occurred
	}
	return exists, nil // Return true if the table exists
}
