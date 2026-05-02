package builders

import (
	"database/sql"
	"errors"
	"fmt"
	"slices"

	"github.com/akfaiz/migris/schema/blueprint"
	"github.com/akfaiz/migris/schema/core"
	"github.com/akfaiz/migris/schema/grammars"
)

type sqliteBuilder struct {
	baseBuilder
}

var _ Builder = (*sqliteBuilder)(nil)

func NewSqliteBuilder() Builder {
	g, _ := grammars.NewGrammar("sqlite3")
	b := &sqliteBuilder{}
	b.baseBuilder = baseBuilder{Grammar: g, Outer: b}
	return b
}

func (b *sqliteBuilder) Drop(c core.Context, name string) error {
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

func (b *sqliteBuilder) DropIfExists(c core.Context, name string) error {
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

func (b *sqliteBuilder) Rename(c core.Context, from, to string) error {
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

func (b *sqliteBuilder) GetColumns(c core.Context, tableName string) ([]*core.Column, error) {
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
		var notNull int
		var discard any
		if err = rows.Scan(&discard, &col.Name, &col.TypeName, &notNull, &col.DefaultVal, &discard); err != nil {
			return nil, err
		}
		col.Nullable = (notNull == 0)
		columns = append(columns, &col)
	}
	return columns, rows.Err()
}

func (b *sqliteBuilder) GetIndexes(c core.Context, tableName string) ([]*core.Index, error) {
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

	var indexes []*core.Index
	for rows.Next() {
		var idx core.Index
		var unique int
		var origin string
		var discard any
		if err = rows.Scan(&discard, &idx.Name, &unique, &origin, &discard); err != nil {
			return nil, err
		}
		idx.Unique = (unique == 1)
		idx.Primary = (origin == "pk")

		// Fetch columns for each index
		idx.Columns = b.fetchIndexColumns(c, idx.Name)
		indexes = append(indexes, &idx)
	}

	// SQLite's PRAGMA index_list does not always include the primary key if it's an INTEGER PRIMARY KEY.
	if !b.hasPrimaryIndex(indexes) {
		if pkIdx := b.fetchPrimaryKey(c, tableName); pkIdx != nil {
			indexes = append(indexes, pkIdx)
		}
	}

	return indexes, rows.Err()
}

func (b *sqliteBuilder) fetchIndexColumns(c core.Context, indexName string) []string {
	var columns []string
	columnQuery := fmt.Sprintf("PRAGMA index_info(%q)", indexName)
	colRows, colErr := c.Query(columnQuery)
	if colErr != nil {
		return columns
	}
	defer colRows.Close()

	for colRows.Next() {
		var colName string
		var discard any
		if scanErr := colRows.Scan(&discard, &discard, &colName); scanErr == nil {
			columns = append(columns, colName)
		}
	}
	return columns
}

func (b *sqliteBuilder) hasPrimaryIndex(indexes []*core.Index) bool {
	for _, idx := range indexes {
		if idx.Primary {
			return true
		}
	}
	return false
}

func (b *sqliteBuilder) fetchPrimaryKey(c core.Context, tableName string) *core.Index {
	tableInfoQuery := fmt.Sprintf("PRAGMA table_info(%q)", tableName)
	tableInfoRows, infoErr := c.Query(tableInfoQuery)
	if infoErr != nil {
		return nil
	}
	defer tableInfoRows.Close()

	var pkColumns []string
	for tableInfoRows.Next() {
		var colName string
		var pk int
		var discard any
		if scanErr := tableInfoRows.Scan(
			&discard,
			&colName,
			&discard,
			&discard,
			&discard,
			&pk,
		); scanErr == nil && pk > 0 {
			pkColumns = append(pkColumns, colName)
		}
	}

	if len(pkColumns) > 0 {
		return &core.Index{
			Name:    "primary",
			Columns: pkColumns,
			Unique:  true,
			Primary: true,
		}
	}
	return nil
}

func (b *sqliteBuilder) GetTables(c core.Context) ([]*core.TableInfo, error) {
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
		if err = rows.Scan(&table.Name); err != nil {
			return nil, err
		}
		tables = append(tables, &table)
	}
	return tables, rows.Err()
}

func (b *sqliteBuilder) HasTable(c core.Context, name string) (bool, error) {
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

func (b *sqliteBuilder) HasColumn(c core.Context, table, column string) (bool, error) {
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

func (b *sqliteBuilder) HasColumns(c core.Context, table string, columns []string) (bool, error) {
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

func (b *sqliteBuilder) HasIndex(c core.Context, table string, columns []string) (bool, error) {
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
