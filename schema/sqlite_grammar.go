package schema

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

type sqliteGrammar struct {
	baseGrammar
}

func newSqliteGrammar() *sqliteGrammar {
	return &sqliteGrammar{}
}

// QuoteString overrides the base implementation to use double quotes for SQLite identifiers.
func (g *sqliteGrammar) QuoteString(s string) string {
	return "\"" + s + "\""
}

// GetDefaultValue overrides the base implementation for SQLite-specific default value formatting.
func (g *sqliteGrammar) GetDefaultValue(value any) string {
	if value == nil {
		return "NULL"
	}
	switch v := value.(type) {
	case Expression:
		return v.String()
	case bool:
		// SQLite boolean values as integers without quotes
		if v {
			return "1"
		}
		return "0"
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		// Numeric values without quotes
		return fmt.Sprintf("%v", v)
	case float32, float64:
		// Numeric values without quotes
		return fmt.Sprintf("%v", v)
	default:
		// String values with single quotes (for SQL literals, not identifiers)
		return fmt.Sprintf("'%v'", v)
	}
}

// CreateIndexName overrides the base implementation for SQLite-specific naming conventions.
func (g *sqliteGrammar) CreateIndexName(blueprint *Blueprint, idxType string, columns ...string) string {
	tableName := blueprint.name
	if strings.Contains(tableName, ".") {
		parts := strings.Split(tableName, ".")
		tableName = parts[len(parts)-1] // Use the last part as the table name
	}

	switch idxType {
	case "primary":
		return fmt.Sprintf("pk_%s", tableName)
	case "unique":
		return fmt.Sprintf("uq_%s_%s", tableName, strings.Join(columns, "_"))
	case "index":
		return fmt.Sprintf("idx_%s_%s", tableName, strings.Join(columns, "_"))
	case "fulltext":
		return fmt.Sprintf("ft_%s_%s", tableName, strings.Join(columns, "_"))
	default:
		return ""
	}
}

func (g *sqliteGrammar) CompileTableExists(_ string, table string) (string, error) {
	return fmt.Sprintf(
		"SELECT 1 FROM sqlite_master WHERE type = 'table' AND name = %s",
		g.QuoteString(table),
	), nil
}

func (g *sqliteGrammar) CompileTables(_ string) (string, error) {
	return "SELECT name, 0 as size, '' as comment FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%' ORDER BY name", nil
}

func (g *sqliteGrammar) CompileColumns(_, table string) (string, error) {
	return fmt.Sprintf("PRAGMA table_info(%s)", g.QuoteString(table)), nil
}

func (g *sqliteGrammar) CompileIndexes(_, table string) (string, error) {
	return fmt.Sprintf(
		"SELECT name, 0 as unique_flag, '' as columns FROM sqlite_master WHERE type = 'index' AND tbl_name = %s",
		g.QuoteString(table),
	), nil
}

func (g *sqliteGrammar) CompileCreate(blueprint *Blueprint) (string, error) {
	sql := g.compileCreateTable(blueprint)
	return sql, nil
}

func (g *sqliteGrammar) compileCreateTable(blueprint *Blueprint) string {
	columns := g.getColumns(blueprint)
	constraints := g.getTableConstraints(blueprint)

	var items []string
	items = append(items, columns...)
	items = append(items, constraints...)

	return fmt.Sprintf("CREATE TABLE %s (%s)",
		g.QuoteString(blueprint.name),
		strings.Join(items, ", "),
	)
}

// getTableConstraints extracts table-level constraints that should be included in CREATE TABLE.
func (g *sqliteGrammar) getTableConstraints(blueprint *Blueprint) []string {
	var constraints []string

	// Process commands to find constraints that should be included in CREATE TABLE
	for _, cmd := range blueprint.commands {
		switch cmd.name {
		case commandForeign:
			// Foreign keys must be defined at table creation time in SQLite
			if sql := g.compileForeignConstraint(blueprint, cmd); sql != "" {
				constraints = append(constraints, sql)
			}
		case commandPrimary:
			// Composite primary keys should be defined at table level
			if sql := g.compilePrimaryConstraint(blueprint, cmd); sql != "" {
				constraints = append(constraints, sql)
			}
		}
	}

	return constraints
}

// compileForeignConstraint compiles foreign key constraint for inclusion in CREATE TABLE.
func (g *sqliteGrammar) compileForeignConstraint(_ *Blueprint, command *command) string {
	if len(command.columns) == 0 || command.on == "" {
		return ""
	}

	foreign := fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s",
		g.QuoteColumnize(command.columns),
		g.QuoteString(command.on),
	)

	if len(command.references) > 0 {
		foreign += fmt.Sprintf(" (%s)", g.QuoteColumnize(command.references))
	}

	if command.onUpdate != "" {
		foreign += " ON UPDATE " + strings.ToUpper(command.onUpdate)
	}

	if command.onDelete != "" {
		foreign += " ON DELETE " + strings.ToUpper(command.onDelete)
	}

	return foreign
}

// compilePrimaryConstraint compiles composite primary key constraint for inclusion in CREATE TABLE.
func (g *sqliteGrammar) compilePrimaryConstraint(_ *Blueprint, command *command) string {
	if len(command.columns) <= 1 {
		// Single column primary keys should be handled at column level
		return ""
	}

	return fmt.Sprintf("PRIMARY KEY (%s)", g.QuoteColumnize(command.columns))
}

func (g *sqliteGrammar) CompileAdd(blueprint *Blueprint) (string, error) {
	columns := g.getColumns(blueprint)

	return fmt.Sprintf("ALTER TABLE %s %s",
		g.QuoteString(blueprint.name),
		strings.Join(g.PrefixArray("ADD COLUMN ", columns), ", "),
	), nil
}

func (g *sqliteGrammar) CompileChange(_ *Blueprint, _ *command) (string, error) {
	return "", errors.New("SQLite does not support modifying columns")
}

func (g *sqliteGrammar) CompileRename(bp *Blueprint, command *command) (string, error) {
	return fmt.Sprintf("ALTER TABLE %s RENAME TO %s",
		g.QuoteString(bp.name),
		g.QuoteString(command.to),
	), nil
}

func (g *sqliteGrammar) CompileDrop(blueprint *Blueprint) (string, error) {
	return fmt.Sprintf("DROP TABLE %s", g.QuoteString(blueprint.name)), nil
}

func (g *sqliteGrammar) CompileDropIfExists(blueprint *Blueprint) (string, error) {
	return fmt.Sprintf("DROP TABLE IF EXISTS %s", g.QuoteString(blueprint.name)), nil
}

func (g *sqliteGrammar) CompileDropColumn(_ *Blueprint, _ *command) (string, error) {
	return "", errors.New("SQLite does not support dropping columns")
}

func (g *sqliteGrammar) CompileRenameColumn(_ *Blueprint, _ *command) (string, error) {
	return "", errors.New("SQLite does not support renaming columns")
}

func (g *sqliteGrammar) CompileIndex(blueprint *Blueprint, command *command) (string, error) {
	if slices.Contains(command.columns, "") {
		return "", errors.New("index column cannot be empty")
	}

	indexName := command.index
	if indexName == "" {
		indexName = g.CreateIndexName(blueprint, "index", command.columns...)
	}

	return fmt.Sprintf("CREATE INDEX %s ON %s (%s)",
		g.QuoteString(indexName),
		g.QuoteString(blueprint.name),
		g.QuoteColumnize(command.columns),
	), nil
}

func (g *sqliteGrammar) CompileUnique(blueprint *Blueprint, command *command) (string, error) {
	if slices.Contains(command.columns, "") {
		return "", errors.New("unique index column cannot be empty")
	}

	indexName := command.index
	if indexName == "" {
		indexName = g.CreateIndexName(blueprint, "unique", command.columns...)
	}

	return fmt.Sprintf("CREATE UNIQUE INDEX %s ON %s (%s)",
		g.QuoteString(indexName),
		g.QuoteString(blueprint.name),
		g.QuoteColumnize(command.columns),
	), nil
}

// QuoteColumnize quotes individual column names for SQLite.
func (g *sqliteGrammar) QuoteColumnize(columns []string) string {
	if len(columns) == 0 {
		return ""
	}
	quoted := make([]string, len(columns))
	for i, col := range columns {
		quoted[i] = g.QuoteString(col)
	}
	return strings.Join(quoted, ", ")
}

func (g *sqliteGrammar) CompileFullText(_ *Blueprint, _ *command) (string, error) {
	return "", errors.New("SQLite full-text search requires special setup")
}

func (g *sqliteGrammar) CompilePrimary(_ *Blueprint, _ *command) (string, error) {
	// SQLite composite primary keys are handled at table creation time via getTableConstraints
	// Single column primary keys are handled at column level via modifiers
	return "", nil
}

func (g *sqliteGrammar) CompileDropIndex(_ *Blueprint, command *command) (string, error) {
	indexName := command.index
	if indexName == "" {
		return "", errors.New("index name cannot be empty")
	}

	return fmt.Sprintf("DROP INDEX %s", g.QuoteString(indexName)), nil
}

func (g *sqliteGrammar) CompileDropUnique(blueprint *Blueprint, command *command) (string, error) {
	return g.CompileDropIndex(blueprint, command)
}

func (g *sqliteGrammar) CompileDropFulltext(blueprint *Blueprint, command *command) (string, error) {
	return g.CompileDropIndex(blueprint, command)
}

func (g *sqliteGrammar) CompileDropPrimary(_ *Blueprint, _ *command) (string, error) {
	return "", errors.New("SQLite does not support dropping primary keys")
}

func (g *sqliteGrammar) CompileRenameIndex(_ *Blueprint, _ *command) (string, error) {
	return "", errors.New("SQLite does not support renaming indexes")
}

func (g *sqliteGrammar) CompileForeign(_ *Blueprint, _ *command) (string, error) {
	// SQLite foreign keys must be defined at table creation time, not as separate statements
	// They are handled in compileCreateTable via getTableConstraints
	return "", nil
}

func (g *sqliteGrammar) CompileDropForeign(_ *Blueprint, _ *command) (string, error) {
	return "", errors.New("SQLite does not support dropping foreign keys")
}

func (g *sqliteGrammar) GetFluentCommands() []func(blueprint *Blueprint, command *command) string {
	return []func(blueprint *Blueprint, command *command) string{
		// Add fluent command handlers here if needed
	}
}

func (g *sqliteGrammar) getColumns(blueprint *Blueprint) []string {
	var columns []string

	for _, column := range blueprint.getAddedColumns() {
		sql := strings.TrimSpace(strings.Join(g.modifiers(column, blueprint), " "))
		columns = append(columns, sql)
	}

	return columns
}

func (g *sqliteGrammar) getType(column *columnDefinition) string {
	typeFuncMap := map[string]func(*columnDefinition) string{
		"char":          g.typeChar,
		"string":        g.typeString,
		"tinyText":      g.typeTinyText,
		"text":          g.typeText,
		"mediumText":    g.typeMediumText,
		"longText":      g.typeLongText,
		"bigInteger":    g.typeBigInteger,
		"integer":       g.typeInteger,
		"mediumInteger": g.typeMediumInteger,
		"smallInteger":  g.typeSmallInteger,
		"tinyInteger":   g.typeTinyInteger,
		"float":         g.typeFloat,
		"double":        g.typeDouble,
		"decimal":       g.typeDecimal,
		"boolean":       g.typeBoolean,
		"enum":          g.typeEnum,
		"json":          g.typeJSON,
		"jsonb":         g.typeJSONB,
		"date":          g.typeDate,
		"dateTime":      g.typeDateTime,
		"dateTimeTz":    g.typeDateTimeTz,
		"time":          g.typeTime,
		"timeTz":        g.typeTimeTz,
		"timestamp":     g.typeTimestamp,
		"timestampTz":   g.typeTimestampTz,
		"year":          g.typeYear,
		"binary":        g.typeBinary,
		"uuid":          g.typeUUID,
		"geometry":      g.typeGeometry,
		"geography":     g.typeGeography,
		"point":         g.typePoint,
	}
	if fn, ok := typeFuncMap[column.columnType]; ok {
		return fn(column)
	}
	return "TEXT"
}

// SQLite type mappings.
func (g *sqliteGrammar) typeChar(_ *columnDefinition) string {
	return "TEXT"
}

func (g *sqliteGrammar) typeString(_ *columnDefinition) string {
	return "TEXT"
}

func (g *sqliteGrammar) typeTinyText(_ *columnDefinition) string {
	return "TEXT"
}

func (g *sqliteGrammar) typeText(_ *columnDefinition) string {
	return "TEXT"
}

func (g *sqliteGrammar) typeMediumText(_ *columnDefinition) string {
	return "TEXT"
}

func (g *sqliteGrammar) typeLongText(_ *columnDefinition) string {
	return "TEXT"
}

func (g *sqliteGrammar) typeBigInteger(_ *columnDefinition) string {
	return "INTEGER"
}

func (g *sqliteGrammar) typeInteger(_ *columnDefinition) string {
	return "INTEGER"
}

func (g *sqliteGrammar) typeMediumInteger(_ *columnDefinition) string {
	return "INTEGER"
}

func (g *sqliteGrammar) typeSmallInteger(_ *columnDefinition) string {
	return "INTEGER"
}

func (g *sqliteGrammar) typeTinyInteger(_ *columnDefinition) string {
	return "INTEGER"
}

func (g *sqliteGrammar) typeFloat(_ *columnDefinition) string {
	return "REAL"
}

func (g *sqliteGrammar) typeDouble(_ *columnDefinition) string {
	return "REAL"
}

func (g *sqliteGrammar) typeDecimal(_ *columnDefinition) string {
	return "NUMERIC"
}

func (g *sqliteGrammar) typeBoolean(_ *columnDefinition) string {
	return "INTEGER"
}

func (g *sqliteGrammar) typeEnum(_ *columnDefinition) string {
	return "TEXT"
}

func (g *sqliteGrammar) typeJSON(_ *columnDefinition) string {
	return "TEXT"
}

func (g *sqliteGrammar) typeJSONB(_ *columnDefinition) string {
	return "TEXT"
}

func (g *sqliteGrammar) typeDate(_ *columnDefinition) string {
	return "DATE"
}

func (g *sqliteGrammar) typeDateTime(col *columnDefinition) string {
	if col.useCurrent {
		col.SetDefault(Expression("CURRENT_TIMESTAMP"))
	}
	return "DATETIME"
}

func (g *sqliteGrammar) typeDateTimeTz(col *columnDefinition) string {
	if col.useCurrent {
		col.SetDefault(Expression("CURRENT_TIMESTAMP"))
	}
	return "DATETIME"
}

func (g *sqliteGrammar) typeTime(_ *columnDefinition) string {
	return "TIME"
}

func (g *sqliteGrammar) typeTimeTz(_ *columnDefinition) string {
	return "TIME"
}

func (g *sqliteGrammar) typeTimestamp(col *columnDefinition) string {
	if col.useCurrent {
		col.SetDefault(Expression("CURRENT_TIMESTAMP"))
	}
	return "DATETIME"
}

func (g *sqliteGrammar) typeTimestampTz(col *columnDefinition) string {
	if col.useCurrent {
		col.SetDefault(Expression("CURRENT_TIMESTAMP"))
	}
	return "DATETIME"
}

func (g *sqliteGrammar) typeYear(_ *columnDefinition) string {
	return "INTEGER"
}

func (g *sqliteGrammar) typeBinary(_ *columnDefinition) string {
	return "BLOB"
}

func (g *sqliteGrammar) typeUUID(_ *columnDefinition) string {
	return "TEXT"
}

func (g *sqliteGrammar) typeGeometry(_ *columnDefinition) string {
	return "TEXT"
}

func (g *sqliteGrammar) typeGeography(_ *columnDefinition) string {
	return "TEXT"
}

func (g *sqliteGrammar) typePoint(_ *columnDefinition) string {
	return "TEXT"
}

func (g *sqliteGrammar) modifiers(column *columnDefinition, blueprint *Blueprint) []string {
	var modifiers []string

	// Add the column name and type first
	modifiers = append(modifiers, g.QuoteString(column.name), g.getType(column))

	// SQLite modifier order: [CONSTRAINT name] [PRIMARY KEY | UNIQUE] [NOT NULL] [DEFAULT value]
	for _, method := range []string{"primary", "unique", "nullable", "default"} {
		if modifier := g.getModifier(method, column, blueprint); modifier != "" {
			modifiers = append(modifiers, modifier)
		}
	}

	return modifiers
}

func (g *sqliteGrammar) getModifier(name string, column *columnDefinition, blueprint *Blueprint) string {
	switch name {
	case "primary":
		return g.modifyPrimary(column, blueprint)
	case "unique":
		return g.modifyUnique(column, blueprint)
	case "nullable":
		return g.modifyNullable(column, blueprint)
	case "default":
		return g.modifyDefault(column, blueprint)
	default:
		return ""
	}
}

func (g *sqliteGrammar) modifyPrimary(column *columnDefinition, _ *Blueprint) string {
	if column.primary != nil && *column.primary {
		primary := "PRIMARY KEY"
		if column.autoIncrement != nil && *column.autoIncrement {
			primary += " AUTOINCREMENT"
		}
		return primary
	}
	return ""
}

func (g *sqliteGrammar) modifyUnique(column *columnDefinition, _ *Blueprint) string {
	if column.unique != nil && *column.unique {
		return "UNIQUE"
	}
	return ""
}

func (g *sqliteGrammar) modifyNullable(column *columnDefinition, _ *Blueprint) string {
	if column.nullable == nil || !*column.nullable {
		return "NOT NULL"
	}
	return ""
}

func (g *sqliteGrammar) modifyDefault(column *columnDefinition, _ *Blueprint) string {
	if column.defaultValue != nil {
		return fmt.Sprintf("DEFAULT %s", g.GetDefaultValue(column.defaultValue))
	}
	return ""
}
