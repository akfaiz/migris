package grammars

import (
	"errors"
	"fmt"
	"strings"

	"github.com/akfaiz/migris/schema/blueprint"
)

type sqliteGrammar struct {
	blueprint.BaseGrammar
}

func newSqliteGrammar() *sqliteGrammar {
	return &sqliteGrammar{}
}

func (g *sqliteGrammar) CompileTableExists(_, table string) (string, error) {
	return fmt.Sprintf("SELECT 1 FROM sqlite_master WHERE type='table' AND name=%s", g.QuoteString(table)), nil
}

func (g *sqliteGrammar) CompileTables(_ string) (string, error) {
	return "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'", nil
}

func (g *sqliteGrammar) CompileColumns(_, table string) (string, error) {
	return fmt.Sprintf("PRAGMA table_info(%q)", table), nil
}

func (g *sqliteGrammar) CompileIndexes(_, table string) (string, error) {
	return fmt.Sprintf("PRAGMA index_list(%q)", table), nil
}

func (g *sqliteGrammar) CompileCreate(bp *blueprint.Blueprint) (string, error) {
	columns, err := g.getColumns(bp)
	if err != nil {
		return "", err
	}
	sql := fmt.Sprintf("CREATE TABLE %q (%s)", bp.Name, strings.Join(columns, ", "))
	return sql, nil
}

func (g *sqliteGrammar) CompileAdd(bp *blueprint.Blueprint) (string, error) {
	columns, err := g.getColumns(bp)
	if err != nil {
		return "", err
	}
	if len(columns) == 0 {
		return "", nil
	}
	for i, col := range columns {
		columns[i] = "ADD COLUMN " + col
	}
	return fmt.Sprintf("ALTER TABLE %q %s", bp.Name, strings.Join(columns, ", ")), nil
}

func (g *sqliteGrammar) CompileChange(_ *blueprint.Blueprint, _ *blueprint.Command) (string, error) {
	return "", errors.New("sqlite does not support changing columns directly")
}

func (g *sqliteGrammar) CompileRename(blueprint *blueprint.Blueprint, command *blueprint.Command) (string, error) {
	return fmt.Sprintf("ALTER TABLE %q RENAME TO %q", blueprint.Name, command.To), nil
}

func (g *sqliteGrammar) CompileDrop(blueprint *blueprint.Blueprint) (string, error) {
	return fmt.Sprintf("DROP TABLE %q", blueprint.Name), nil
}

func (g *sqliteGrammar) CompileDropIfExists(blueprint *blueprint.Blueprint) (string, error) {
	return fmt.Sprintf("DROP TABLE IF EXISTS %q", blueprint.Name), nil
}

func (g *sqliteGrammar) CompileDropColumn(_ *blueprint.Blueprint, _ *blueprint.Command) (string, error) {
	return "", errors.New("sqlite drop column has limited support")
}

func (g *sqliteGrammar) CompileIndex(blueprint *blueprint.Blueprint, command *blueprint.Command) (string, error) {
	index := g.WrapIndexName(command.Index, `"`)
	if index == "" {
		index = g.WrapIndexName(g.CreateIndexName(blueprint, "index", command.Columns...), `"`)
	}
	return fmt.Sprintf("CREATE INDEX %s ON %q (%s)", index, blueprint.Name, g.WrapColumnize(command.Columns, `"`)), nil
}

func (g *sqliteGrammar) CompileUnique(blueprint *blueprint.Blueprint, command *blueprint.Command) (string, error) {
	index := g.WrapIndexName(command.Index, `"`)
	if index == "" {
		index = g.WrapIndexName(g.CreateIndexName(blueprint, "unique", command.Columns...), `"`)
	}
	return fmt.Sprintf(
		"CREATE UNIQUE INDEX %s ON %q (%s)",
		index,
		blueprint.Name,
		g.WrapColumnize(command.Columns, `"`),
	), nil
}

func (g *sqliteGrammar) CompilePrimary(_ *blueprint.Blueprint, _ *blueprint.Command) (string, error) {
	return "", nil
}

func (g *sqliteGrammar) CompileForeign(_ *blueprint.Blueprint, _ *blueprint.Command) (string, error) {
	return "", nil
}

func (g *sqliteGrammar) CompileDropIndex(_ *blueprint.Blueprint, command *blueprint.Command) (string, error) {
	if command.Index == "" {
		return "", errors.New("index name cannot be empty")
	}
	return fmt.Sprintf("DROP INDEX %s", g.WrapIndexName(command.Index, `"`)), nil
}

func (g *sqliteGrammar) GetType(col *blueprint.Column) string {
	typeFuncMap := map[string]func(*blueprint.Column) string{
		blueprint.ColumnTypeChar:          g.typeChar,
		blueprint.ColumnTypeString:        g.typeString,
		blueprint.ColumnTypeTinyText:      g.typeTinyText,
		blueprint.ColumnTypeText:          g.typeText,
		blueprint.ColumnTypeMediumText:    g.typeMediumText,
		blueprint.ColumnTypeLongText:      g.typeLongText,
		blueprint.ColumnTypeBigInteger:    g.typeBigInteger,
		blueprint.ColumnTypeInteger:       g.typeInteger,
		blueprint.ColumnTypeMediumInteger: g.typeMediumInteger,
		blueprint.ColumnTypeSmallInteger:  g.typeSmallInteger,
		blueprint.ColumnTypeTinyInteger:   g.typeTinyInteger,
		blueprint.ColumnTypeFloat:         g.typeFloat,
		blueprint.ColumnTypeDouble:        g.typeDouble,
		blueprint.ColumnTypeDecimal:       g.typeDecimal,
		blueprint.ColumnTypeBoolean:       g.typeBoolean,
		blueprint.ColumnTypeEnum:          g.typeEnum,
		blueprint.ColumnTypeJSON:          g.typeJSON,
		blueprint.ColumnTypeJSONB:         g.typeJSONB,
		blueprint.ColumnTypeDate:          g.typeDate,
		blueprint.ColumnTypeDateTime:      g.typeDateTime,
		blueprint.ColumnTypeDateTimeTz:    g.typeDateTimeTz,
		blueprint.ColumnTypeTime:          g.typeTime,
		blueprint.ColumnTypeTimeTz:        g.typeTimeTz,
		blueprint.ColumnTypeTimestamp:     g.typeTimestamp,
		blueprint.ColumnTypeTimestampTz:   g.typeTimestampTz,
		blueprint.ColumnTypeYear:          g.typeYear,
		blueprint.ColumnTypeBinary:        g.typeBinary,
		blueprint.ColumnTypeUUID:          g.typeUUID,
		blueprint.ColumnTypeULID:          g.typeULID,
		blueprint.ColumnTypeIPAddress:     g.typeIPAddress,
		blueprint.ColumnTypeMacAddress:    g.typeMacAddress,
		blueprint.ColumnTypeGeometry:      g.typeGeometry,
		blueprint.ColumnTypeGeography:     g.typeGeography,
		blueprint.ColumnTypePoint:         g.typePoint,
	}
	if fn, ok := typeFuncMap[col.ColumnType]; ok {
		return fn(col)
	}
	return col.ColumnType
}

func (g *sqliteGrammar) typeChar(_ *blueprint.Column) string {
	return "TEXT"
}

func (g *sqliteGrammar) typeString(_ *blueprint.Column) string {
	return "TEXT"
}

func (g *sqliteGrammar) typeTinyText(_ *blueprint.Column) string {
	return "TEXT"
}

func (g *sqliteGrammar) typeText(_ *blueprint.Column) string {
	return "TEXT"
}

func (g *sqliteGrammar) typeMediumText(_ *blueprint.Column) string {
	return "TEXT"
}

func (g *sqliteGrammar) typeLongText(_ *blueprint.Column) string {
	return "TEXT"
}

func (g *sqliteGrammar) typeBigInteger(_ *blueprint.Column) string {
	return "INTEGER"
}

func (g *sqliteGrammar) typeInteger(_ *blueprint.Column) string {
	return "INTEGER"
}

func (g *sqliteGrammar) typeMediumInteger(_ *blueprint.Column) string {
	return "INTEGER"
}

func (g *sqliteGrammar) typeSmallInteger(_ *blueprint.Column) string {
	return "INTEGER"
}

func (g *sqliteGrammar) typeTinyInteger(_ *blueprint.Column) string {
	return "INTEGER"
}

func (g *sqliteGrammar) typeFloat(_ *blueprint.Column) string {
	return "REAL"
}

func (g *sqliteGrammar) typeDouble(_ *blueprint.Column) string {
	return "REAL"
}

func (g *sqliteGrammar) typeDecimal(_ *blueprint.Column) string {
	return "NUMERIC"
}

func (g *sqliteGrammar) typeBoolean(_ *blueprint.Column) string {
	return "INTEGER"
}

func (g *sqliteGrammar) typeEnum(_ *blueprint.Column) string {
	return "TEXT"
}

func (g *sqliteGrammar) typeJSON(_ *blueprint.Column) string {
	return "TEXT"
}

func (g *sqliteGrammar) typeJSONB(_ *blueprint.Column) string {
	return "TEXT"
}

func (g *sqliteGrammar) typeDate(_ *blueprint.Column) string {
	return "DATE"
}

func (g *sqliteGrammar) typeDateTime(_ *blueprint.Column) string {
	return "DATETIME"
}

func (g *sqliteGrammar) typeDateTimeTz(_ *blueprint.Column) string {
	return "DATETIME"
}

func (g *sqliteGrammar) typeTime(_ *blueprint.Column) string {
	return "TEXT"
}

func (g *sqliteGrammar) typeTimeTz(_ *blueprint.Column) string {
	return "TEXT"
}

func (g *sqliteGrammar) typeTimestamp(_ *blueprint.Column) string {
	return "DATETIME"
}

func (g *sqliteGrammar) typeTimestampTz(_ *blueprint.Column) string {
	return "DATETIME"
}

func (g *sqliteGrammar) typeYear(_ *blueprint.Column) string {
	return "INTEGER"
}

func (g *sqliteGrammar) typeBinary(_ *blueprint.Column) string {
	return "BLOB"
}

func (g *sqliteGrammar) typeUUID(_ *blueprint.Column) string {
	return "TEXT"
}

func (g *sqliteGrammar) typeULID(_ *blueprint.Column) string {
	return "TEXT"
}

func (g *sqliteGrammar) typeIPAddress(_ *blueprint.Column) string {
	return "TEXT"
}

func (g *sqliteGrammar) typeMacAddress(_ *blueprint.Column) string {
	return "TEXT"
}

func (g *sqliteGrammar) typeGeometry(_ *blueprint.Column) string {
	return "TEXT"
}

func (g *sqliteGrammar) typeGeography(_ *blueprint.Column) string {
	return "TEXT"
}

func (g *sqliteGrammar) typePoint(_ *blueprint.Column) string {
	return "TEXT"
}

func (g *sqliteGrammar) GetDefaultValue(value any) string {
	if value == nil {
		return "NULL"
	}
	switch v := value.(type) {
	case blueprint.Expression:
		return v.String()
	case bool:
		if v {
			return "1"
		}
		return "0"
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return fmt.Sprintf("%v", v)
	default:
		return fmt.Sprintf("'%v'", v)
	}
}

func (g *sqliteGrammar) getColumns(bp *blueprint.Blueprint) ([]string, error) {
	var columns []string
	for _, col := range bp.GetAddedColumns() {
		if col.Name == "" {
			return nil, errors.New("column name cannot be empty")
		}
		sql := fmt.Sprintf("%q %s", col.Name, g.GetType(col))
		sql += g.modifiers(col)
		columns = append(columns, sql)
	}
	return columns, nil
}

func (g *sqliteGrammar) modifiers(col *blueprint.Column) string {
	var sql string
	if col.PrimaryVal != nil && *col.PrimaryVal {
		sql += " PRIMARY KEY"
		if col.AutoIncrementVal != nil && *col.AutoIncrementVal {
			sql += " AUTOINCREMENT"
		}
	}
	if col.NullableVal != nil {
		if !*col.NullableVal {
			sql += " NOT NULL"
		}
	} else {
		sql += " NOT NULL"
	}
	if col.DefaultValue != nil {
		sql += fmt.Sprintf(" DEFAULT %s", g.GetDefaultValue(col.DefaultValue))
	} else if col.UseCurrentVal {
		sql += " DEFAULT CURRENT_TIMESTAMP"
	}
	return sql
}
