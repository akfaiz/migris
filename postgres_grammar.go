package schema

import (
	"fmt"
	"slices"
	"strings"
)

type pgGrammar struct {
	baseGrammar
}

var _ grammar = (*pgGrammar)(nil)

func newPgGrammar() *pgGrammar {
	return &pgGrammar{}
}

func (g *pgGrammar) compileTableExists(schema string, table string) (string, error) {
	return fmt.Sprintf(
		"SELECT 1 FROM information_schema.tables WHERE table_schema = %s AND table_name = %s AND table_type = 'BASE TABLE'",
		g.quoteString(schema),
		g.quoteString(table),
	), nil
}

func (g *pgGrammar) compileTables() (string, error) {
	return "select c.relname as name, n.nspname as schema, pg_total_relation_size(c.oid) as size, " +
		"obj_description(c.oid, 'pg_class') as comment from pg_class c, pg_namespace n " +
		"where c.relkind in ('r', 'p') and n.oid = c.relnamespace and n.nspname not in ('pg_catalog', 'information_schema') " +
		"order by c.relname", nil
}

func (g *pgGrammar) compileColumns(schema, table string) (string, error) {
	return fmt.Sprintf(
		"select a.attname as name, t.typname as type_name, format_type(a.atttypid, a.atttypmod) as type, "+
			"(select tc.collcollate from pg_catalog.pg_collation tc where tc.oid = a.attcollation) as collation, "+
			"not a.attnotnull as nullable, "+
			"(select pg_get_expr(adbin, adrelid) from pg_attrdef where c.oid = pg_attrdef.adrelid and pg_attrdef.adnum = a.attnum) as default, "+
			"col_description(c.oid, a.attnum) as comment "+
			"from pg_attribute a, pg_class c, pg_type t, pg_namespace n "+
			"where c.relname = %s and n.nspname = %s and a.attnum > 0 and a.attrelid = c.oid and a.atttypid = t.oid and n.oid = c.relnamespace "+
			"order by a.attnum",
		g.quoteString(table),
		g.quoteString(schema),
	), nil
}

func (g *pgGrammar) compileIndexes(schema, table string) (string, error) {
	return fmt.Sprintf(
		"select ic.relname as name, string_agg(a.attname, ',' order by indseq.ord) as columns, "+
			"am.amname as \"type\", i.indisunique as \"unique\", i.indisprimary as \"primary\" "+
			"from pg_index i "+
			"join pg_class tc on tc.oid = i.indrelid "+
			"join pg_namespace tn on tn.oid = tc.relnamespace "+
			"join pg_class ic on ic.oid = i.indexrelid "+
			"join pg_am am on am.oid = ic.relam "+
			"join lateral unnest(i.indkey) with ordinality as indseq(num, ord) on true "+
			"left join pg_attribute a on a.attrelid = i.indrelid and a.attnum = indseq.num "+
			"where tc.relname = %s and tn.nspname = %s "+
			"group by ic.relname, am.amname, i.indisunique, i.indisprimary",
		g.quoteString(table),
		g.quoteString(schema),
	), nil
}

func (g *pgGrammar) compileCreate(blueprint *Blueprint) (string, error) {
	columns, err := g.getColumns(blueprint)
	if err != nil {
		return "", err
	}
	columns = append(columns, g.getConstraints(blueprint)...)
	return fmt.Sprintf("CREATE TABLE %s (%s)", blueprint.name, strings.Join(columns, ", ")), nil
}

func (g *pgGrammar) compileCreateIfNotExists(blueprint *Blueprint) (string, error) {
	columns, err := g.getColumns(blueprint)
	if err != nil {
		return "", err
	}
	columns = append(columns, g.getConstraints(blueprint)...)
	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", blueprint.name, strings.Join(columns, ", ")), nil
}

func (g *pgGrammar) compileAdd(blueprint *Blueprint) (string, error) {
	if len(blueprint.getAddedColumns()) == 0 {
		return "", nil
	}

	columns, err := g.getColumns(blueprint)
	if err != nil {
		return "", err
	}
	columns = g.prefixArray("ADD COLUMN ", columns)
	constraints := g.getConstraints(blueprint)
	if len(constraints) > 0 {
		constraints = g.prefixArray("ADD ", constraints)
		columns = append(columns, constraints...)
	}

	return fmt.Sprintf("ALTER TABLE %s %s",
		blueprint.name,
		strings.Join(columns, ", "),
	), nil
}

func (g *pgGrammar) compileChange(bp *Blueprint, command *command) (string, error) {
	column := command.column
	if column.name == "" {
		return "", fmt.Errorf("column name cannot be empty for change operation")
	}

	var changes []string
	changes = append(changes, fmt.Sprintf("TYPE %s", g.getType(command.column)))
	for _, modifier := range g.modifiers() {
		change := modifier(command.column)
		if change != "" {
			changes = append(changes, strings.TrimSpace(change))
		}
	}

	return fmt.Sprintf("ALTER TABLE %s %s",
		bp.name,
		strings.Join(g.prefixArray(fmt.Sprintf("ALTER COLUMN %s ", column.name), changes), ", "),
	), nil
}

func (g *pgGrammar) compileDrop(blueprint *Blueprint) (string, error) {
	return fmt.Sprintf("DROP TABLE %s", blueprint.name), nil
}

func (g *pgGrammar) compileDropIfExists(blueprint *Blueprint) (string, error) {
	return fmt.Sprintf("DROP TABLE IF EXISTS %s", blueprint.name), nil
}

func (g *pgGrammar) compileRename(blueprint *Blueprint, command *command) (string, error) {
	return fmt.Sprintf("ALTER TABLE %s RENAME TO %s", blueprint.name, command.to), nil
}

func (g *pgGrammar) compileDropColumn(blueprint *Blueprint, command *command) (string, error) {
	if len(blueprint.columns) == 0 {
		return "", nil
	}
	columns := g.prefixArray("DROP COLUMN ", command.columns)

	return fmt.Sprintf("ALTER TABLE %s %s", blueprint.name, strings.Join(columns, ", ")), nil
}

func (g *pgGrammar) compileRenameColumn(blueprint *Blueprint, command *command) (string, error) {
	if command.from == "" || command.to == "" {
		return "", fmt.Errorf("table name, old column name, and new column name cannot be empty for rename operation")
	}
	return fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s", blueprint.name, command.from, command.to), nil
}

func (g *pgGrammar) compileFullText(blueprint *Blueprint, command *command) (string, error) {
	if slices.Contains(command.columns, "") {
		return "", fmt.Errorf("fulltext index column cannot be empty")
	}
	indexName := command.index
	if indexName == "" {
		indexName = g.createIndexName(blueprint, commandFullText, command.columns...)
	}
	language := command.language
	if language == "" {
		language = "english" // Default language for full-text search
	}
	var columns []string
	for _, col := range command.columns {
		columns = append(columns, fmt.Sprintf("to_tsvector(%s, %s)", g.quoteString(language), col))
	}

	return fmt.Sprintf("CREATE INDEX %s ON %s USING GIN (%s)", indexName, blueprint.name, strings.Join(columns, " || ")), nil
}

func (g *pgGrammar) compileIndex(blueprint *Blueprint, command *command) (string, error) {
	if slices.Contains(command.columns, "") {
		return "", fmt.Errorf("index column cannot be empty")
	}
	indexName := command.index
	if indexName == "" {
		indexName = g.createIndexName(blueprint, commandIndex, command.columns...)
	}

	sql := fmt.Sprintf("CREATE INDEX %s ON %s", indexName, blueprint.name)
	if command.algorithm != "" {
		sql += fmt.Sprintf(" USING %s", command.algorithm)
	}
	return fmt.Sprintf("%s (%s)", sql, g.columnize(command.columns)), nil
}

func (g *pgGrammar) compileUnique(blueprint *Blueprint, command *command) (string, error) {
	if slices.Contains(command.columns, "") {
		return "", fmt.Errorf("unique index column cannot be empty")
	}
	indexName := command.index
	if indexName == "" {
		indexName = g.createIndexName(blueprint, commandUnique, command.columns...)
	}
	sql := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s UNIQUE (%s)",
		blueprint.name,
		indexName,
		g.columnize(command.columns),
	)

	if command.deferrable != nil {
		if *command.deferrable {
			sql += " DEFERRABLE"
		} else {
			sql += " NOT DEFERRABLE"
		}
	}
	if command.deferrable != nil && *command.deferrable && command.initiallyImmediate != nil {
		if *command.initiallyImmediate {
			sql += " INITIALLY IMMEDIATE"
		} else {
			sql += " INITIALLY DEFERRED"
		}
	}

	return sql, nil
}

func (g *pgGrammar) compilePrimary(blueprint *Blueprint, command *command) (string, error) {
	if slices.Contains(command.columns, "") {
		return "", fmt.Errorf("primary key index column cannot be empty")
	}
	indexName := command.index
	if indexName == "" {
		indexName = g.createIndexName(blueprint, commandPrimary, command.columns...)
	}
	return fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s PRIMARY KEY (%s)", blueprint.name, indexName, g.columnize(command.columns)), nil
}

func (g *pgGrammar) compileDropIndex(_ *Blueprint, command *command) (string, error) {
	if command.index == "" {
		return "", fmt.Errorf("index name cannot be empty for drop operation")
	}
	return fmt.Sprintf("DROP INDEX %s", command.index), nil
}

func (g *pgGrammar) compileDropFulltext(blueprint *Blueprint, command *command) (string, error) {
	return g.compileDropIndex(blueprint, command)
}

func (g *pgGrammar) compileDropUnique(blueprint *Blueprint, command *command) (string, error) {
	if command.index == "" {
		return "", fmt.Errorf("index name cannot be empty for drop operation")
	}
	return fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", blueprint.name, command.index), nil
}

func (g *pgGrammar) compileDropPrimary(blueprint *Blueprint, command *command) (string, error) {
	if command.index == "" {
		command.index = g.createIndexName(blueprint, commandPrimary)
	}
	return fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", blueprint.name, command.index), nil
}

func (g *pgGrammar) compileRenameIndex(_ *Blueprint, command *command) (string, error) {
	if command.from == "" || command.to == "" {
		return "", fmt.Errorf("index names for rename operation cannot be empty: oldName=%s, newName=%s", command.from, command.to)
	}
	return fmt.Sprintf("ALTER INDEX %s RENAME TO %s", command.from, command.to), nil
}

func (g *pgGrammar) compileForeign(blueprint *Blueprint, command *command) (string, error) {
	sql, err := g.baseGrammar.compileForeign(blueprint, command)
	if err != nil {
		return "", err
	}

	if command.deferrable != nil {
		if *command.deferrable {
			sql += " DEFERRABLE"
		} else {
			sql += " NOT DEFERRABLE"
		}
	}
	if command.deferrable != nil && *command.deferrable && command.initiallyImmediate != nil {
		if *command.initiallyImmediate {
			sql += " INITIALLY IMMEDIATE"
		} else {
			sql += " INITIALLY DEFERRED"
		}
	}

	return sql, nil
}

func (g *pgGrammar) compileDropForeign(blueprint *Blueprint, command *command) (string, error) {
	if command.index == "" {
		return "", fmt.Errorf("foreign key name cannot be empty for drop operation")
	}
	return fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", blueprint.name, command.index), nil
}

func (g *pgGrammar) getFluentCommands() []func(blueprint *Blueprint, command *command) string {
	return []func(blueprint *Blueprint, command *command) string{
		g.compileComment,
	}
}

func (g *pgGrammar) compileComment(blueprint *Blueprint, command *command) string {
	if command.column.comment != nil || command.column.change {
		sql := fmt.Sprintf("COMMENT ON COLUMN %s.%s IS ", blueprint.name, command.column.name)
		if command.column.comment == nil {
			return sql + "NULL"
		} else {
			return sql + fmt.Sprintf("'%s'", *command.column.comment)
		}
	}
	return ""
}

func (g *pgGrammar) getColumns(blueprint *Blueprint) ([]string, error) {
	var columns []string
	for _, col := range blueprint.getAddedColumns() {
		if col.name == "" {
			return nil, fmt.Errorf("column name cannot be empty")
		}
		sql := col.name + " " + g.getType(col)
		for _, modifier := range g.modifiers() {
			sql += modifier(col)
		}
		columns = append(columns, sql)
	}

	return columns, nil
}

func (g *pgGrammar) getConstraints(blueprint *Blueprint) []string {
	var constrains []string
	for _, col := range blueprint.getAddedColumns() {
		if col.primary != nil && *col.primary {
			pkConstraintName := g.createIndexName(blueprint, commandPrimary)
			sql := "CONSTRAINT " + pkConstraintName + " PRIMARY KEY (" + col.name + ")"
			constrains = append(constrains, sql)
			continue
		}
	}

	return constrains
}

func (g *pgGrammar) getType(col *columnDefinition) string {
	typeMapFunc := map[string]func(*columnDefinition) string{
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
		columnTypeJSON:          g.typeJson,
		columnTypeJSONB:         g.typeJsonb,
		columnTypeDate:          g.typeDate,
		columnTypeDateTime:      g.typeDateTime,
		columnTypeDateTimeTz:    g.typeDateTimeTz,
		columnTypeTime:          g.typeTime,
		columnTypeTimeTz:        g.typeTimeTz,
		columnTypeTimestamp:     g.typeTimestamp,
		columnTypeTimestampTz:   g.typeTimestampTz,
		columnTypeYear:          g.typeYear,
		columnTypeBinary:        g.typeBinary,
		columnTypeUUID:          g.typeUUID,
		columnTypeGeography:     g.typeGeography,
		columnTypeGeometry:      g.typeGeometry,
		columnTypePoint:         g.typePoint,
	}
	if fn, ok := typeMapFunc[col.columnType]; ok {
		return fn(col)
	}
	return col.columnType
}

func (g *pgGrammar) typeChar(col *columnDefinition) string {
	if col.length != nil && *col.length > 0 {
		return fmt.Sprintf("CHAR(%d)", *col.length)
	}
	return "CHAR"
}

func (g *pgGrammar) typeString(col *columnDefinition) string {
	if col.length != nil && *col.length > 0 {
		return fmt.Sprintf("VARCHAR(%d)", *col.length)
	}
	return "VARCHAR"
}

func (g *pgGrammar) typeTinyText(_ *columnDefinition) string {
	return "VARCHAR(255)"
}

func (g *pgGrammar) typeText(_ *columnDefinition) string {
	return "TEXT"
}

func (g *pgGrammar) typeMediumText(_ *columnDefinition) string {
	return "TEXT"
}

func (g *pgGrammar) typeLongText(_ *columnDefinition) string {
	return "TEXT"
}

func (g *pgGrammar) typeBigInteger(col *columnDefinition) string {
	if col.autoIncrement != nil && *col.autoIncrement {
		return "BIGSERIAL"
	}
	return "BIGINT"
}

func (g *pgGrammar) typeInteger(col *columnDefinition) string {
	if col.autoIncrement != nil && *col.autoIncrement {
		return "SERIAL"
	}
	return "INTEGER"
}

func (g *pgGrammar) typeMediumInteger(col *columnDefinition) string {
	return g.typeInteger(col)
}

func (g *pgGrammar) typeSmallInteger(col *columnDefinition) string {
	if col.autoIncrement != nil && *col.autoIncrement {
		return "SMALLSERIAL"
	}
	return "SMALLINT"
}

func (g *pgGrammar) typeTinyInteger(col *columnDefinition) string {
	return g.typeSmallInteger(col)
}

func (g *pgGrammar) typeFloat(_ *columnDefinition) string {
	return "REAL"
}

func (g *pgGrammar) typeDouble(_ *columnDefinition) string {
	return "DOUBLE PRECISION"
}

func (g *pgGrammar) typeDecimal(col *columnDefinition) string {
	return fmt.Sprintf("DECIMAL(%d, %d)", *col.total, *col.places)
}

func (g *pgGrammar) typeBoolean(_ *columnDefinition) string {
	return "BOOLEAN"
}

func (g *pgGrammar) typeEnum(col *columnDefinition) string {
	enumValues := make([]string, len(col.allowed))
	for i, v := range col.allowed {
		enumValues[i] = g.quoteString(v)
	}
	return "VARCHAR(255) CHECK (" + col.name + " IN (" + strings.Join(enumValues, ", ") + "))"
}

func (g *pgGrammar) typeJson(_ *columnDefinition) string {
	return "JSON"
}

func (g *pgGrammar) typeJsonb(_ *columnDefinition) string {
	return "JSONB"
}

func (g *pgGrammar) typeDate(_ *columnDefinition) string {
	return "DATE"
}

func (g *pgGrammar) typeDateTime(col *columnDefinition) string {
	return g.typeTimestamp(col)
}

func (g *pgGrammar) typeDateTimeTz(col *columnDefinition) string {
	return g.typeTimestampTz(col)
}

func (g *pgGrammar) typeTime(col *columnDefinition) string {
	if col.precision != nil && *col.precision > 0 {
		return fmt.Sprintf("TIME(%d)", *col.precision)
	}
	return "TIME"
}

func (g *pgGrammar) typeTimeTz(col *columnDefinition) string {
	if col.precision != nil && *col.precision > 0 {
		return fmt.Sprintf("TIMETZ(%d)", *col.precision)
	}
	return "TIMETZ"
}

func (g *pgGrammar) typeTimestamp(col *columnDefinition) string {
	if col.useCurrent != nil && *col.useCurrent {
		col.Default(Expression("CURRENT_TIMESTAMP"))
	}
	if col.precision != nil {
		return fmt.Sprintf("TIMESTAMP(%d)", *col.precision)
	}
	return "TIMESTAMP"
}

func (g *pgGrammar) typeTimestampTz(col *columnDefinition) string {
	if col.useCurrent != nil && *col.useCurrent {
		col.Default(Expression("CURRENT_TIMESTAMP"))
	}
	if col.precision != nil {
		return fmt.Sprintf("TIMESTAMPTZ(%d)", *col.precision)
	}
	return "TIMESTAMPTZ"
}

func (g *pgGrammar) typeYear(_ *columnDefinition) string {
	return "INTEGER"
}

func (g *pgGrammar) typeBinary(_ *columnDefinition) string {
	return "BYTEA"
}

func (g *pgGrammar) typeUUID(_ *columnDefinition) string {
	return "UUID"
}

func (g *pgGrammar) typeGeography(col *columnDefinition) string {
	if col.subtype != nil && col.srid != nil {
		return fmt.Sprintf("GEOGRAPHY(%s, %d)", *col.subtype, *col.srid)
	} else if col.subtype != nil {
		return fmt.Sprintf("GEOGRAPHY(%s)", *col.subtype)
	}
	return "GEOGRAPHY"
}

func (g *pgGrammar) typeGeometry(col *columnDefinition) string {
	if col.subtype != nil && col.srid != nil {
		return fmt.Sprintf("GEOMETRY(%s, %d)", *col.subtype, *col.srid)
	} else if col.subtype != nil {
		return fmt.Sprintf("GEOMETRY(%s)", *col.subtype)
	}
	return "GEOMETRY"
}

func (g *pgGrammar) typePoint(col *columnDefinition) string {
	if col.srid != nil {
		return fmt.Sprintf("POINT(%d)", *col.srid)
	}
	return "POINT"
}

func (g *pgGrammar) modifiers() []func(*columnDefinition) string {
	return []func(*columnDefinition) string{
		g.modifyNullable,
		g.modifyDefault,
	}
}

func (g *pgGrammar) modifyNullable(col *columnDefinition) string {
	if col.change {
		if col.nullable != nil && *col.nullable {
			return " DROP NOT NULL"
		}
		return " SET NOT NULL"
	}
	if col.nullable != nil && *col.nullable {
		return " NULL"
	}
	return " NOT NULL"
}

func (g *pgGrammar) modifyDefault(col *columnDefinition) string {
	if col.defaultValue != nil {
		return fmt.Sprintf(" DEFAULT %s", g.getDefaultValue(col.defaultValue))
	}
	return ""
}
