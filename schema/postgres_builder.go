package schema

import (
	"database/sql"
	"errors"
	"strings"
)

type postgresBuilder struct {
	baseBuilder
	grammar *postgresGrammar
}

func newPostgresBuilder() Builder {
	grammar := newPostgresGrammar()

	return &postgresBuilder{
		baseBuilder: baseBuilder{grammar: grammar},
		grammar:     grammar,
	}
}

func (b *postgresBuilder) parseSchemaAndTable(name string) (string, string) {
	names := strings.Split(name, ".")
	if len(names) == 2 {
		return names[0], names[1]
	}
	return "", names[0]
}

func (b *postgresBuilder) GetColumns(c *Context, tableName string) ([]*Column, error) {
	if c == nil || tableName == "" {
		return nil, errors.New("invalid arguments: context is nil or table name is empty")
	}

	schema, name := b.parseSchemaAndTable(tableName)
	if schema == "" {
		schema = "public" // Default schema for PostgreSQL
	}
	query, err := b.grammar.CompileColumns(schema, name)
	if err != nil {
		return nil, err
	}

	rows, err := c.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	var columns []*Column
	for rows.Next() {
		var col Column
		if err := rows.Scan(
			&col.Name, &col.TypeName, &col.TypeFull, &col.Collation,
			&col.Nullable, &col.DefaultVal, &col.Comment,
		); err != nil {
			return nil, err
		}
		columns = append(columns, &col)
	}

	return columns, nil
}

func (b *postgresBuilder) GetIndexes(c *Context, tableName string) ([]*Index, error) {
	if c == nil || tableName == "" {
		return nil, errors.New("invalid arguments: context is nil or table name is empty")
	}
	schema, name := b.parseSchemaAndTable(tableName)
	if schema == "" {
		schema = "public" // Default schema for PostgreSQL
	}
	query, err := b.grammar.CompileIndexes(schema, name)
	if err != nil {
		return nil, err
	}
	rows, err := c.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	var indexes []*Index
	for rows.Next() {
		var index Index
		var columnsStr string
		if err := rows.Scan(&index.Name, &columnsStr, &index.Type, &index.Unique, &index.Primary); err != nil {
			return nil, err
		}
		index.Columns = strings.Split(columnsStr, ",")
		indexes = append(indexes, &index)
	}

	return indexes, nil
}

func (b *postgresBuilder) GetTables(c *Context) ([]*TableInfo, error) {
	if c == nil {
		return nil, errors.New("invalid arguments: context is nil")
	}

	query, err := b.grammar.CompileTables()
	if err != nil {
		return nil, err
	}

	rows, err := c.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	var tables []*TableInfo
	for rows.Next() {
		var table TableInfo
		if err := rows.Scan(&table.Name, &table.Schema, &table.Size, &table.Comment); err != nil {
			return nil, err
		}
		tables = append(tables, &table)
	}

	return tables, nil
}

func (b *postgresBuilder) HasColumn(c *Context, tableName string, columnName string) (bool, error) {
	return b.HasColumns(c, tableName, []string{columnName})
}

func (b *postgresBuilder) HasColumns(c *Context, tableName string, columnNames []string) (bool, error) {
	if c == nil || tableName == "" {
		return false, errors.New("invalid arguments: context is nil or table name is empty")
	}
	if len(columnNames) == 0 {
		return false, errors.New("no column names provided")
	}
	existingColumns, err := b.GetColumns(c, tableName)
	if err != nil {
		return false, err
	}
	if len(existingColumns) == 0 {
		return false, nil // No columns found, so the specified columns cannot exist
	}
	if len(columnNames) == 0 {
		return true, nil // No specific columns to check, so we assume they exist
	}
	existingColumnMap := make(map[string]bool)
	for _, col := range existingColumns {
		existingColumnMap[col.Name] = true
	}
	for _, colName := range columnNames {
		if colName == "" {
			return false, errors.New("column name is empty")
		}
		if _, exists := existingColumnMap[colName]; !exists {
			return false, nil // If any specified column does not exist, return false
		}
	}
	return true, nil // All specified columns exist
}

func (b *postgresBuilder) HasIndex(c *Context, tableName string, indexes []string) (bool, error) {
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

func (b *postgresBuilder) HasTable(c *Context, name string) (bool, error) {
	if c == nil || name == "" {
		return false, errors.New("invalid arguments: context is nil or table name is empty")
	}

	schema, name := b.parseSchemaAndTable(name)
	if schema == "" {
		schema = "public" // Default schema for PostgreSQL
	}
	query, err := b.grammar.CompileTableExists(schema, name)
	if err != nil {
		return false, err
	}

	var exists bool
	if err := c.QueryRow(query).Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil // Table does not exist
		}
		return false, err // Other error
	}
	return exists, nil
}
