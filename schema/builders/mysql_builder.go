package builders

import (
	"database/sql"
	"errors"
	"slices"
	"strings"

	"github.com/akfaiz/migris/schema/blueprint"
	"github.com/akfaiz/migris/schema/core"
	"github.com/akfaiz/migris/schema/grammars"
)

type mysqlBuilder struct {
	baseBuilder
}

var _ Builder = (*mysqlBuilder)(nil)

func NewMysqlBuilder() Builder {
	g, _ := grammars.NewGrammar("mysql")
	b := &mysqlBuilder{}
	b.baseBuilder = baseBuilder{Grammar: g, Outer: b}
	return b
}

func (b *mysqlBuilder) Drop(c core.Context, name string) error {
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

func (b *mysqlBuilder) DropIfExists(c core.Context, name string) error {
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

func (b *mysqlBuilder) Rename(c core.Context, from, to string) error {
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

func (b *mysqlBuilder) GetColumns(c core.Context, tableName string) ([]*core.Column, error) {
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
		var nullableStr, keyVal, privilegesVal string
		if err = rows.Scan(
			&col.Name, &col.TypeFull, &col.Collation,
			&nullableStr, &keyVal, &col.DefaultVal,
			&col.Extra, &privilegesVal, &col.Comment,
		); err != nil {
			return nil, err
		}

		// Extract TypeName from TypeFull (e.g., "varchar(255)" -> "varchar")
		col.TypeName = col.TypeFull
		if idx := strings.Index(col.TypeFull, "("); idx != -1 {
			col.TypeName = col.TypeFull[:idx]
		}

		if nullableStr == "YES" {
			col.Nullable = true
		}
		columns = append(columns, &col)
	}
	return columns, rows.Err()
}

func (b *mysqlBuilder) GetIndexes(c core.Context, tableName string) ([]*core.Index, error) {
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

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var nonUniqueVal int
		var keyNameVal, columnNameVal string

		dest := make([]any, len(cols))
		for i, col := range cols {
			switch strings.ToLower(col) {
			case "non_unique":
				dest[i] = &nonUniqueVal
			case "key_name":
				dest[i] = &keyNameVal
			case "column_name":
				dest[i] = &columnNameVal
			default:
				dest[i] = new(any)
			}
		}

		err = rows.Scan(dest...)
		if err != nil {
			return nil, err
		}

		if idx, ok := indexMap[keyNameVal]; ok {
			idx.Columns = append(idx.Columns, columnNameVal)
		} else {
			idx = &core.Index{
				Name:    keyNameVal,
				Columns: []string{columnNameVal},
				Unique:  nonUniqueVal == 0,
				Primary: keyNameVal == "PRIMARY",
			}
			indexMap[keyNameVal] = idx
			indexNames = append(indexNames, keyNameVal)
		}
	}

	var result []*core.Index
	for _, name := range indexNames {
		result = append(result, indexMap[name])
	}
	return result, rows.Err()
}

func (b *mysqlBuilder) GetTables(c core.Context) ([]*core.TableInfo, error) {
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
		if err = rows.Scan(&table.Name, &table.Comment); err != nil {
			return nil, err
		}
		tables = append(tables, &table)
	}
	return tables, rows.Err()
}

func (b *mysqlBuilder) HasTable(c core.Context, name string) (bool, error) {
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

func (b *mysqlBuilder) HasColumn(c core.Context, table, column string) (bool, error) {
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

func (b *mysqlBuilder) HasColumns(c core.Context, table string, columns []string) (bool, error) {
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

func (b *mysqlBuilder) HasIndex(c core.Context, table string, columns []string) (bool, error) {
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
