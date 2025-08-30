package schema

import (
	"fmt"
	"slices"
	"strings"

	"github.com/afkdevs/go-schema/internal/util"
)

type mysqlGrammar struct {
	baseGrammar

	serials []string
}

func newMysqlGrammar() *mysqlGrammar {
	return &mysqlGrammar{
		serials: []string{
			"bigInteger", "integer", "mediumInteger", "smallInteger",
			"tinyInteger",
		},
	}
}

func (g *mysqlGrammar) CompileCurrentDatabase() string {
	return "SELECT DATABASE()"
}

func (g *mysqlGrammar) CompileTableExists(database string, table string) (string, error) {
	return fmt.Sprintf(
		"SELECT 1 FROM information_schema.tables WHERE table_schema = %s AND table_name = %s AND table_type = 'BASE TABLE'",
		g.QuoteString(database),
		g.QuoteString(table),
	), nil
}

func (g *mysqlGrammar) CompileTables(database string) (string, error) {
	return fmt.Sprintf(
		"select table_name as `name`, (data_length + index_length) as `size`, "+
			"table_comment as `comment`, engine as `engine`, table_collation as `collation` "+
			"from information_schema.tables where table_schema = %s and table_type in ('BASE TABLE', 'SYSTEM VERSIONED') "+
			"order by table_name",
		g.QuoteString(database),
	), nil
}

func (g *mysqlGrammar) CompileColumns(database, table string) (string, error) {
	return fmt.Sprintf(
		"select column_name as `name`, data_type as `type_name`, column_type as `type`, "+
			"collation_name as `collation`, is_nullable as `nullable`, "+
			"column_default as `default`, column_comment as `comment`, extra as `extra` "+
			"from information_schema.columns where table_schema = %s and table_name = %s "+
			"order by ordinal_position asc",
		g.QuoteString(database),
		g.QuoteString(table),
	), nil
}

func (g *mysqlGrammar) CompileIndexes(database, table string) (string, error) {
	return fmt.Sprintf(
		"select index_name as `name`, group_concat(column_name order by seq_in_index) as `columns`, "+
			"index_type as `type`, not non_unique as `unique` "+
			"from information_schema.statistics where table_schema = %s and table_name = %s "+
			"group by index_name, index_type, non_unique",
		g.QuoteString(database),
		g.QuoteString(table),
	), nil
}

func (g *mysqlGrammar) CompileCreate(blueprint *Blueprint) (string, error) {
	sql, err := g.compileCreateTable(blueprint)
	if err != nil {
		return "", err
	}
	sql = g.compileCreateEncoding(sql, blueprint)

	return g.compileCreateEngine(sql, blueprint), nil
}

func (g *mysqlGrammar) compileCreateTable(blueprint *Blueprint) (string, error) {
	columns, err := g.getColumns(blueprint)
	if err != nil {
		return "", err
	}

	constraints := g.getConstraints(blueprint)
	columns = append(columns, constraints...)

	return fmt.Sprintf("CREATE TABLE %s (%s)", blueprint.name, strings.Join(columns, ", ")), nil
}

func (g *mysqlGrammar) compileCreateEncoding(sql string, blueprint *Blueprint) string {
	if blueprint.charset != "" {
		sql += fmt.Sprintf(" DEFAULT CHARACTER SET %s", blueprint.charset)
	}
	if blueprint.collation != "" {
		sql += fmt.Sprintf(" COLLATE %s", blueprint.collation)
	}

	return sql
}

func (g *mysqlGrammar) compileCreateEngine(sql string, blueprint *Blueprint) string {
	if blueprint.engine != "" {
		sql += fmt.Sprintf(" ENGINE = %s", blueprint.engine)
	}
	return sql
}

func (g *mysqlGrammar) CompileAdd(blueprint *Blueprint) (string, error) {
	if len(blueprint.getAddedColumns()) == 0 {
		return "", nil
	}

	columns, err := g.getColumns(blueprint)
	if err != nil {
		return "", err
	}
	columns = g.PrefixArray("ADD COLUMN ", columns)
	constraints := g.getConstraints(blueprint)
	constraints = g.PrefixArray("ADD  ", constraints)
	columns = append(columns, constraints...)

	return fmt.Sprintf("ALTER TABLE %s %s",
		blueprint.name,
		strings.Join(columns, ", "),
	), nil
}

func (g *mysqlGrammar) CompileChange(bp *Blueprint, command *command) (string, error) {
	column := command.column
	if column.name == "" {
		return "", fmt.Errorf("column name cannot be empty for change operation")
	}

	sql := fmt.Sprintf("ALTER TABLE %s MODIFY COLUMN %s %s", bp.name, column.name, g.getType(column))
	for _, modifier := range g.modifiers() {
		sql += modifier(column)
	}

	return sql, nil
}

func (g *mysqlGrammar) CompileRename(blueprint *Blueprint, command *command) (string, error) {
	return fmt.Sprintf("ALTER TABLE %s RENAME TO %s", blueprint.name, command.to), nil
}

func (g *mysqlGrammar) CompileDrop(blueprint *Blueprint) (string, error) {
	if blueprint.name == "" {
		return "", fmt.Errorf("table name cannot be empty")
	}
	return fmt.Sprintf("DROP TABLE %s", blueprint.name), nil
}

func (g *mysqlGrammar) CompileDropIfExists(blueprint *Blueprint) (string, error) {
	if blueprint.name == "" {
		return "", fmt.Errorf("table name cannot be empty")
	}
	return fmt.Sprintf("DROP TABLE IF EXISTS %s", blueprint.name), nil
}

func (g *mysqlGrammar) CompileDropColumn(blueprint *Blueprint, command *command) (string, error) {
	if len(command.columns) == 0 {
		return "", fmt.Errorf("no columns to drop")
	}
	columns := make([]string, len(command.columns))
	for i, col := range command.columns {
		if col == "" {
			return "", fmt.Errorf("column name cannot be empty")
		}
		columns[i] = col
	}
	columns = g.PrefixArray("DROP COLUMN ", columns)
	return fmt.Sprintf("ALTER TABLE %s %s", blueprint.name, strings.Join(columns, ", ")), nil
}

func (g *mysqlGrammar) CompileRenameColumn(blueprint *Blueprint, command *command) (string, error) {
	if command.from == "" || command.to == "" {
		return "", fmt.Errorf("old and new column names cannot be empty")
	}
	return fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s", blueprint.name, command.from, command.to), nil
}

func (g *mysqlGrammar) CompileIndex(blueprint *Blueprint, command *command) (string, error) {
	if slices.Contains(command.columns, "") {
		return "", fmt.Errorf("index column cannot be empty")
	}

	indexName := command.index
	if indexName == "" {
		indexName = g.CreateIndexName(blueprint, "index", command.columns...)
	}

	sql := fmt.Sprintf("CREATE INDEX %s ON %s (%s)", indexName, blueprint.name, g.Columnize(command.columns))
	if command.algorithm != "" {
		sql += fmt.Sprintf(" USING %s", command.algorithm)
	}

	return sql, nil
}

func (g *mysqlGrammar) CompileUnique(blueprint *Blueprint, command *command) (string, error) {
	if slices.Contains(command.columns, "") {
		return "", fmt.Errorf("unique column cannot be empty")
	}

	indexName := command.index
	if indexName == "" {
		indexName = g.CreateIndexName(blueprint, "unique", command.columns...)
	}
	sql := fmt.Sprintf("CREATE UNIQUE INDEX %s ON %s (%s)", indexName, blueprint.name, g.Columnize(command.columns))
	if command.algorithm != "" {
		sql += fmt.Sprintf(" USING %s", command.algorithm)
	}

	return sql, nil
}

func (g *mysqlGrammar) CompileFullText(blueprint *Blueprint, command *command) (string, error) {
	if slices.Contains(command.columns, "") {
		return "", fmt.Errorf("fulltext index column cannot be empty")
	}

	indexName := command.index
	if indexName == "" {
		indexName = g.CreateIndexName(blueprint, "fulltext", command.columns...)
	}

	return fmt.Sprintf("CREATE FULLTEXT INDEX %s ON %s (%s)", indexName, blueprint.name, g.Columnize(command.columns)), nil
}

func (g *mysqlGrammar) CompilePrimary(blueprint *Blueprint, command *command) (string, error) {
	if slices.Contains(command.columns, "") {
		return "", fmt.Errorf("primary key column cannot be empty")
	}

	indexName := command.index
	if indexName == "" {
		indexName = g.CreateIndexName(blueprint, "primary", command.columns...)
	}

	return fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s PRIMARY KEY (%s)", blueprint.name, indexName, g.Columnize(command.columns)), nil
}

func (g *mysqlGrammar) CompileDropIndex(blueprint *Blueprint, command *command) (string, error) {
	if command.index == "" {
		return "", fmt.Errorf("index name cannot be empty")
	}
	return fmt.Sprintf("ALTER TABLE %s DROP INDEX %s", blueprint.name, command.index), nil
}

func (g *mysqlGrammar) CompileDropUnique(blueprint *Blueprint, command *command) (string, error) {
	if command.index == "" {
		return "", fmt.Errorf("unique index name cannot be empty")
	}
	return fmt.Sprintf("ALTER TABLE %s DROP INDEX %s", blueprint.name, command.index), nil
}

func (g *mysqlGrammar) CompileDropFulltext(blueprint *Blueprint, command *command) (string, error) {
	return g.CompileDropIndex(blueprint, command)
}

func (g *mysqlGrammar) CompileDropPrimary(blueprint *Blueprint, _ *command) (string, error) {
	return fmt.Sprintf("ALTER TABLE %s DROP PRIMARY KEY", blueprint.name), nil
}

func (g *mysqlGrammar) CompileRenameIndex(blueprint *Blueprint, command *command) (string, error) {
	if command.from == "" || command.to == "" {
		return "", fmt.Errorf("old and new index names cannot be empty")
	}
	return fmt.Sprintf("ALTER TABLE %s RENAME INDEX %s TO %s", blueprint.name, command.from, command.to), nil
}

func (g *mysqlGrammar) CompileDropForeign(blueprint *Blueprint, command *command) (string, error) {
	if command.index == "" {
		return "", fmt.Errorf("foreign key name cannot be empty")
	}
	return fmt.Sprintf("ALTER TABLE %s DROP FOREIGN KEY %s", blueprint.name, command.index), nil
}

func (g *mysqlGrammar) GetFluentCommands() []func(*Blueprint, *command) string {
	return []func(*Blueprint, *command) string{}
}

func (g *mysqlGrammar) getColumns(blueprint *Blueprint) ([]string, error) {
	var columns []string
	for _, col := range blueprint.getAddedColumns() {
		if col.name == "" {
			return nil, fmt.Errorf("column name cannot be empty")
		}
		sql := col.name + " " + g.getType(col)
		sql += g.modifyUnsigned(col)
		sql += g.modifyIncrement(col)
		sql += g.modifyDefault(col)
		sql += g.modifyOnUpdate(col)
		sql += g.modifyCharset(col)
		sql += g.modifyCollate(col)
		sql += g.modifyNullable(col)
		sql += g.modifyComment(col)

		columns = append(columns, sql)
	}

	return columns, nil
}

func (g *mysqlGrammar) getConstraints(blueprint *Blueprint) []string {
	var constrains []string
	for _, col := range blueprint.getAddedColumns() {
		if col.primary != nil && *col.primary {
			pkConstraintName := g.CreateIndexName(blueprint, "primary")
			sql := "CONSTRAINT " + pkConstraintName + " PRIMARY KEY (" + col.name + ")"
			constrains = append(constrains, sql)
			continue
		}
	}

	return constrains
}

func (g *mysqlGrammar) getType(col *columnDefinition) string {
	typeFuncMap := map[string]func(*columnDefinition) string{
		columnTypeChar:          g.typeChar,
		columnTypeString:        g.typeString,
		columnTypeTinyText:      g.typeTinyText,
		columnTypeText:          g.typeText,
		columnTypeMediumText:    g.typeMediumText,
		columnTypeLongText:      g.typeLongText,
		columnTypeInteger:       g.typeInteger,
		columnTypeBigInteger:    g.typeBigInteger,
		columnTypeMediumInteger: g.typeMediumInteger,
		columnTypeSmallInteger:  g.typeSmallInteger,
		columnTypeTinyInteger:   g.typeTinyInteger,
		columnTypeFloat:         g.typeFloat,
		columnTypeDouble:        g.typeDouble,
		columnTypeDecimal:       g.typeDecimal,
		columnTypeBoolean:       g.typeBoolean,
		columnTypeEnum:          g.typeEnum,
		columnTypeJson:          g.typeJson,
		columnTypeJsonb:         g.typeJsonb,
		columnTypeDate:          g.typeDate,
		columnTypeDateTime:      g.typeDateTime,
		columnTypeDateTimeTz:    g.typeDateTimeTz,
		columnTypeTime:          g.typeTime,
		columnTypeTimeTz:        g.typeTimeTz,
		columnTypeTimestamp:     g.typeTimestamp,
		columnTypeTimestampTz:   g.typeTimestampTz,
		columnTypeYear:          g.typeYear,
		columnTypeBinary:        g.typeBinary,
		columnTypeUuid:          g.typeUuid,
		columnTypeGeography:     g.typeGeography,
		columnTypeGeometry:      g.typeGeometry,
		columnTypePoint:         g.typePoint,
	}
	if fn, ok := typeFuncMap[col.columnType]; ok {
		return fn(col)
	}
	return col.columnType
}

func (g *mysqlGrammar) typeChar(col *columnDefinition) string {
	return fmt.Sprintf("CHAR(%d)", *col.length)
}

func (g *mysqlGrammar) typeString(col *columnDefinition) string {
	return fmt.Sprintf("VARCHAR(%d)", *col.length)
}

func (g *mysqlGrammar) typeTinyText(col *columnDefinition) string {
	return "TINYTEXT"
}

func (g *mysqlGrammar) typeText(col *columnDefinition) string {
	return "TEXT"
}

func (g *mysqlGrammar) typeMediumText(col *columnDefinition) string {
	return "MEDIUMTEXT"
}

func (g *mysqlGrammar) typeLongText(col *columnDefinition) string {
	return "LONGTEXT"
}

func (g *mysqlGrammar) typeBigInteger(col *columnDefinition) string {
	return "BIGINT"
}

func (g *mysqlGrammar) typeInteger(col *columnDefinition) string {
	return "INT"
}

func (g *mysqlGrammar) typeMediumInteger(col *columnDefinition) string {
	return "MEDIUMINT"
}

func (g *mysqlGrammar) typeSmallInteger(col *columnDefinition) string {
	return "SMALLINT"
}

func (g *mysqlGrammar) typeTinyInteger(col *columnDefinition) string {
	return "TINYINT"
}

func (g *mysqlGrammar) typeFloat(col *columnDefinition) string {
	if col.precision != nil && *col.precision > 0 {
		return fmt.Sprintf("FLOAT(%d)", *col.precision)
	}
	return "FLOAT"
}

func (g *mysqlGrammar) typeDouble(col *columnDefinition) string {
	return "DOUBLE"
}

func (g *mysqlGrammar) typeDecimal(col *columnDefinition) string {
	return fmt.Sprintf("DECIMAL(%d, %d)", *col.total, *col.places)
}

func (g *mysqlGrammar) typeBoolean(col *columnDefinition) string {
	return "TINYINT(1)"
}

func (g *mysqlGrammar) typeEnum(col *columnDefinition) string {
	allowedValues := make([]string, len(col.allowed))
	for i, e := range col.allowed {
		allowedValues[i] = g.QuoteString(e)
	}
	return fmt.Sprintf("ENUM(%s)", strings.Join(allowedValues, ", "))
}

func (g *mysqlGrammar) typeJson(col *columnDefinition) string {
	return "JSON"
}

func (g *mysqlGrammar) typeJsonb(col *columnDefinition) string {
	return "JSON"
}

func (g *mysqlGrammar) typeDate(col *columnDefinition) string {
	return "DATE"
}

func (g *mysqlGrammar) typeDateTime(col *columnDefinition) string {
	current := "CURRENT_TIMESTAMP"
	if col.precision != nil && *col.precision > 0 {
		current = fmt.Sprintf("CURRENT_TIMESTAMP(%d)", *col.precision)
	}
	if col.useCurrent {
		col.SetDefault(Expression(current))
	}
	if col.useCurrentOnUpdate {
		col.SetOnUpdate(Expression(current))
	}
	if col.precision != nil && *col.precision > 0 {
		return fmt.Sprintf("DATETIME(%d)", *col.precision)
	}
	return "DATETIME"
}

func (g *mysqlGrammar) typeDateTimeTz(col *columnDefinition) string {
	return g.typeDateTime(col)
}

func (g *mysqlGrammar) typeTime(col *columnDefinition) string {
	if col.precision != nil && *col.precision > 0 {
		return fmt.Sprintf("TIME(%d)", *col.precision)
	}
	return "TIME"
}

func (g *mysqlGrammar) typeTimeTz(col *columnDefinition) string {
	return g.typeTime(col)
}

func (g *mysqlGrammar) typeTimestamp(col *columnDefinition) string {
	current := "CURRENT_TIMESTAMP"
	if col.precision != nil && *col.precision > 0 {
		current = fmt.Sprintf("CURRENT_TIMESTAMP(%d)", *col.precision)
	}
	if col.useCurrent {
		col.SetDefault(Expression(current))
	}
	if col.useCurrentOnUpdate {
		col.SetOnUpdate(Expression(current))
	}
	if col.precision != nil && *col.precision > 0 {
		return fmt.Sprintf("TIMESTAMP(%d)", *col.precision)
	}
	return "TIMESTAMP"
}

func (g *mysqlGrammar) typeTimestampTz(col *columnDefinition) string {
	return g.typeTimestamp(col)
}

func (g *mysqlGrammar) typeYear(col *columnDefinition) string {
	return "YEAR"
}

func (g *mysqlGrammar) typeBinary(col *columnDefinition) string {
	if col.length != nil && *col.length > 0 {
		return fmt.Sprintf("BINARY(%d)", *col.length)
	}
	return "BLOB"
}

func (g *mysqlGrammar) typeUuid(col *columnDefinition) string {
	return "CHAR(36)" // Default UUID length
}

func (g *mysqlGrammar) typeGeometry(col *columnDefinition) string {
	subtype := util.Ternary(col.subtype != nil, util.PtrOf(strings.ToUpper(*col.subtype)), nil)
	if subtype != nil {
		if !slices.Contains([]string{"POINT", "LINESTRING", "POLYGON", "GEOMETRYCOLLECTION", "MULTIPOINT", "MULTILINESTRING"}, *subtype) {
			subtype = nil
		}
	}

	if subtype == nil {
		subtype = util.PtrOf("GEOMETRY")
	}
	if col.srid != nil && *col.srid > 0 {
		return fmt.Sprintf("%s SRID %d", *subtype, *col.srid)
	}
	return *subtype
}

func (g *mysqlGrammar) typeGeography(col *columnDefinition) string {
	return g.typeGeometry(col)
}

func (g *mysqlGrammar) typePoint(col *columnDefinition) string {
	col.SetSubtype(util.PtrOf("POINT"))
	return g.typeGeometry(col)
}

func (g *mysqlGrammar) modifiers() []func(*columnDefinition) string {
	return []func(*columnDefinition) string{
		g.modifyUnsigned,
		g.modifyCharset,
		g.modifyCollate,
		g.modifyNullable,
		g.modifyDefault,
		g.modifyOnUpdate,
		g.modifyIncrement,
		g.modifyComment,
	}
}

func (g *mysqlGrammar) modifyCharset(col *columnDefinition) string {
	if col.charset != nil && *col.charset != "" {
		return fmt.Sprintf(" CHARACTER SET %s", *col.charset)
	}
	return ""
}

func (g *mysqlGrammar) modifyCollate(col *columnDefinition) string {
	if col.collation != nil && *col.collation != "" {
		return fmt.Sprintf(" COLLATE %s", *col.collation)
	}
	return ""
}

func (g *mysqlGrammar) modifyComment(col *columnDefinition) string {
	if col.comment != nil {
		return fmt.Sprintf(" COMMENT '%s'", *col.comment)
	}
	return ""
}

func (g *mysqlGrammar) modifyDefault(col *columnDefinition) string {
	if col.hasCommand("default") {
		return fmt.Sprintf(" DEFAULT %s", g.GetDefaultValue(col.defaultValue))
	}
	return ""
}

func (g *mysqlGrammar) modifyIncrement(col *columnDefinition) string {
	if slices.Contains(g.serials, col.columnType) &&
		col.autoIncrement != nil && *col.autoIncrement &&
		col.primary != nil && *col.primary {
		return " AUTO_INCREMENT"
	}
	return ""
}

func (g *mysqlGrammar) modifyNullable(col *columnDefinition) string {
	if col.nullable != nil && *col.nullable {
		return " NULL"
	}
	return " NOT NULL"
}

func (g *mysqlGrammar) modifyOnUpdate(col *columnDefinition) string {
	if col.hasCommand("onUpdate") {
		return fmt.Sprintf(" ON UPDATE %s", g.GetValue(col.onUpdateValue))
	}
	return ""
}

func (g *mysqlGrammar) modifyUnsigned(col *columnDefinition) string {
	if col.unsigned != nil && *col.unsigned {
		return " UNSIGNED"
	}
	return ""
}
