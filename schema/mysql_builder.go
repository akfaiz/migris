package schema

import (
	"database/sql"
	"errors"
	"strings"
)

type mysqlBuilder struct {
	baseBuilder
}

var _ Builder = (*mysqlBuilder)(nil)

func newMysqlBuilder() Builder {
	grammar := newMysqlGrammar()

	return &mysqlBuilder{
		baseBuilder: baseBuilder{grammar: grammar},
	}
}

func (b *mysqlBuilder) GetColumns(c *Context, tableName string) ([]*Column, error) {
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
		var nullableStr string
		if err = rows.Scan(
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
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return columns, nil
}

func (b *mysqlBuilder) GetIndexes(c *Context, tableName string) ([]*Index, error) {
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
		var columnsStr string
		if err = rows.Scan(&idx.Name, &columnsStr, &idx.Type, &idx.Unique); err != nil {
			return nil, err
		}
		idx.Columns = strings.Split(columnsStr, ",")
		indexes = append(indexes, &idx)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return indexes, nil
}

func (b *mysqlBuilder) GetTables(c *Context) ([]*TableInfo, error) {
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
		if err = rows.Scan(&table.Name, &table.Size, &table.Comment, &table.Engine, &table.Collation); err != nil {
			return nil, err
		}
		tables = append(tables, &table)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tables, nil
}

func (b *mysqlBuilder) HasColumn(c *Context, tableName string, columnName string) (bool, error) {
	if c == nil || columnName == "" {
		return false, errors.New("invalid arguments: context is nil or column name is empty")
	}
	return b.HasColumns(c, tableName, []string{columnName})
}

func (b *mysqlBuilder) HasColumns(c *Context, tableName string, columnNames []string) (bool, error) {
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

// nolint: dupl,godoclint // Similar code exists in other builder files
func (b *mysqlBuilder) HasIndex(c *Context, tableName string, indexes []string) (bool, error) {
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

func (b *mysqlBuilder) HasTable(c *Context, name string) (bool, error) {
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
