package builders

import (
	"database/sql"
	"errors"
	"slices"

	"github.com/akfaiz/migris/schema/blueprint"
	"github.com/akfaiz/migris/schema/core"
	"github.com/akfaiz/migris/schema/grammars"
)

type postgresBuilder struct {
	baseBuilder
}

var _ Builder = (*postgresBuilder)(nil)

func NewPostgresBuilder() Builder {
	g, _ := grammars.NewGrammar("postgres")
	b := &postgresBuilder{}
	b.baseBuilder = baseBuilder{Grammar: g, Outer: b}
	return b
}

func (b *postgresBuilder) Drop(c core.Context, name string) error {
	if c == nil || name == "" {
		return errors.New("invalid arguments")
	}
	query, err := b.Grammar.CompileDrop(&blueprint.Blueprint{Name: name})
	if err != nil {
		return err
	}
	_, err = c.Exec(query)
	return err
}

func (b *postgresBuilder) DropIfExists(c core.Context, name string) error {
	if c == nil || name == "" {
		return errors.New("invalid arguments")
	}
	query, err := b.Grammar.CompileDropIfExists(&blueprint.Blueprint{Name: name})
	if err != nil {
		return err
	}
	_, err = c.Exec(query)
	return err
}

func (b *postgresBuilder) Rename(c core.Context, from, to string) error {
	if c == nil || from == "" || to == "" {
		return errors.New("invalid arguments")
	}
	query, err := b.Grammar.CompileRename(&blueprint.Blueprint{Name: from}, &blueprint.Command{To: to})
	if err != nil {
		return err
	}
	_, err = c.Exec(query)
	return err
}

func (b *postgresBuilder) GetColumns(c core.Context, tableName string) ([]*core.Column, error) {
	if c == nil || tableName == "" {
		return nil, errors.New("invalid arguments")
	}

	exists, err := b.HasTable(c, tableName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}

	query, err := b.Grammar.CompileColumns("", tableName)
	if err != nil {
		return nil, err
	}

	rows, err := c.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []*core.Column
	for rows.Next() {
		var col core.Column
		var isNullable string
		if err = rows.Scan(&col.Name, &col.TypeName, &isNullable, &col.DefaultVal, &col.Comment); err != nil {
			return nil, err
		}
		col.Nullable = (isNullable == "YES")
		columns = append(columns, &col)
	}
	return columns, rows.Err()
}

func (b *postgresBuilder) GetIndexes(c core.Context, tableName string) ([]*core.Index, error) {
	if c == nil || tableName == "" {
		return nil, errors.New("invalid arguments")
	}

	exists, err := b.HasTable(c, tableName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}

	query, err := b.Grammar.CompileIndexes("", tableName)
	if err != nil {
		return nil, err
	}
	rows, err := c.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	indexMap := make(map[string]*core.Index)
	var indexNames []string

	for rows.Next() {
		var indexName string
		var isUnique, isPrimary bool
		var columnName string

		if err = rows.Scan(&indexName, &isUnique, &isPrimary, &columnName); err != nil {
			return nil, err
		}

		if idx, ok := indexMap[indexName]; ok {
			idx.Columns = append(idx.Columns, columnName)
		} else {
			idx = &core.Index{
				Name:    indexName,
				Columns: []string{columnName},
				Unique:  isUnique,
				Primary: isPrimary,
			}
			indexMap[indexName] = idx
			indexNames = append(indexNames, indexName)
		}
	}

	var result []*core.Index
	for _, name := range indexNames {
		result = append(result, indexMap[name])
	}
	return result, rows.Err()
}

func (b *postgresBuilder) GetTables(c core.Context) ([]*core.TableInfo, error) {
	if c == nil {
		return nil, errors.New("invalid arguments")
	}
	query, err := b.Grammar.CompileTables("")
	if err != nil {
		return nil, err
	}
	rows, err := c.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []*core.TableInfo
	for rows.Next() {
		var table core.TableInfo
		if err = rows.Scan(&table.Name, &table.Schema, &table.Comment); err != nil {
			return nil, err
		}
		tables = append(tables, &table)
	}
	return tables, rows.Err()
}

func (b *postgresBuilder) HasTable(c core.Context, name string) (bool, error) {
	if c == nil || name == "" {
		return false, errors.New("invalid arguments")
	}
	query, err := b.Grammar.CompileTableExists("", name)
	if err != nil {
		return false, err
	}
	row := c.QueryRow(query)
	var exists int
	err = row.Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (b *postgresBuilder) HasColumn(c core.Context, table, column string) (bool, error) {
	if c == nil || table == "" || column == "" {
		return false, errors.New("invalid arguments")
	}
	cols, err := b.GetColumns(c, table)
	if err != nil {
		return false, err
	}
	for _, col := range cols {
		if col.Name == column {
			return true, nil
		}
	}
	return false, nil
}

func (b *postgresBuilder) HasColumns(c core.Context, table string, columns []string) (bool, error) {
	if c == nil || table == "" || len(columns) == 0 {
		return false, errors.New("invalid arguments")
	}
	cols, err := b.GetColumns(c, table)
	if err != nil {
		return false, err
	}
	colMap := make(map[string]bool)
	for _, col := range cols {
		colMap[col.Name] = true
	}
	for _, col := range columns {
		if col == "" || !colMap[col] {
			return false, nil
		}
	}
	return true, nil
}

func (b *postgresBuilder) HasIndex(c core.Context, table string, columns []string) (bool, error) {
	if c == nil || table == "" || len(columns) == 0 {
		return false, errors.New("invalid arguments")
	}
	indexes, err := b.GetIndexes(c, table)
	if err != nil {
		return false, err
	}
	for _, idx := range indexes {
		if slices.Equal(idx.Columns, columns) {
			return true, nil
		}
		if len(columns) == 1 && idx.Name == columns[0] {
			return true, nil
		}
	}
	return false, nil
}
