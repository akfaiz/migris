package schema

import (
	"fmt"
	"slices"
	"strings"
)

type postgresGrammar struct {
	baseGrammar
}

func newPostgresGrammar() *postgresGrammar {
	return &postgresGrammar{}
}

func (g *postgresGrammar) CompileTableExists(schema string, table string) (string, error) {
	return fmt.Sprintf(
		"SELECT 1 FROM information_schema.tables WHERE table_schema = %s AND table_name = %s AND table_type = 'BASE TABLE'",
		g.QuoteString(schema),
		g.QuoteString(table),
	), nil
}

func (g *postgresGrammar) CompileTables() (string, error) {
	return "select c.relname as name, n.nspname as schema, pg_total_relation_size(c.oid) as size, " +
		"obj_description(c.oid, 'pg_class') as comment from pg_class c, pg_namespace n " +
		"where c.relkind in ('r', 'p') and n.oid = c.relnamespace and n.nspname not in ('pg_catalog', 'information_schema') " +
		"order by c.relname", nil
}

func (g *postgresGrammar) CompileColumns(schema, table string) (string, error) {
	return fmt.Sprintf(
		"select a.attname as name, t.typname as type_name, format_type(a.atttypid, a.atttypmod) as type, "+
			"(select tc.collcollate from pg_catalog.pg_collation tc where tc.oid = a.attcollation) as collation, "+
			"not a.attnotnull as nullable, "+
			"(select pg_get_expr(adbin, adrelid) from pg_attrdef where c.oid = pg_attrdef.adrelid and pg_attrdef.adnum = a.attnum) as default, "+
			"col_description(c.oid, a.attnum) as comment "+
			"from pg_attribute a, pg_class c, pg_type t, pg_namespace n "+
			"where c.relname = %s and n.nspname = %s and a.attnum > 0 and a.attrelid = c.oid and a.atttypid = t.oid and n.oid = c.relnamespace "+
			"order by a.attnum",
		g.QuoteString(table),
		g.QuoteString(schema),
	), nil
}

func (g *postgresGrammar) CompileIndexes(schema, table string) (string, error) {
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
		g.QuoteString(table),
		g.QuoteString(schema),
	), nil
}

func (g *postgresGrammar) CompileCreate(blueprint *Blueprint) (string, error) {
	columns, err := g.getColumns(blueprint)
	if err != nil {
		return "", err
	}
	columns = append(columns, g.getConstraints(blueprint)...)
	return fmt.Sprintf("CREATE TABLE %s (%s)", blueprint.name, strings.Join(columns, ", ")), nil
}

func (g *postgresGrammar) CompileAdd(blueprint *Blueprint) (string, error) {
	if len(blueprint.getAddedColumns()) == 0 {
		return "", nil
	}

	columns, err := g.getColumns(blueprint)
	if err != nil {
		return "", err
	}
	columns = g.PrefixArray("ADD COLUMN ", columns)
	constraints := g.getConstraints(blueprint)
	if len(constraints) > 0 {
		constraints = g.PrefixArray("ADD ", constraints)
		columns = append(columns, constraints...)
	}

	return fmt.Sprintf("ALTER TABLE %s %s",
		blueprint.name,
		strings.Join(columns, ", "),
	), nil
}

func (g *postgresGrammar) CompileChange(bp *Blueprint, command *command) (string, error) {
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
		strings.Join(g.PrefixArray(fmt.Sprintf("ALTER COLUMN %s ", column.name), changes), ", "),
	), nil
}

func (g *postgresGrammar) CompileDrop(blueprint *Blueprint) (string, error) {
	return fmt.Sprintf("DROP TABLE %s", blueprint.name), nil
}

func (g *postgresGrammar) CompileDropIfExists(blueprint *Blueprint) (string, error) {
	return fmt.Sprintf("DROP TABLE IF EXISTS %s", blueprint.name), nil
}

func (g *postgresGrammar) CompileRename(blueprint *Blueprint, command *command) (string, error) {
	return fmt.Sprintf("ALTER TABLE %s RENAME TO %s", blueprint.name, command.to), nil
}

func (g *postgresGrammar) CompileDropColumn(blueprint *Blueprint, command *command) (string, error) {
	if len(command.columns) == 0 {
		return "", nil
	}
	columns := g.PrefixArray("DROP COLUMN ", command.columns)

	return fmt.Sprintf("ALTER TABLE %s %s", blueprint.name, strings.Join(columns, ", ")), nil
}

func (g *postgresGrammar) CompileRenameColumn(blueprint *Blueprint, command *command) (string, error) {
	if command.from == "" || command.to == "" {
		return "", fmt.Errorf("table name, old column name, and new column name cannot be empty for rename operation")
	}
	return fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s", blueprint.name, command.from, command.to), nil
}

func (g *postgresGrammar) CompileFullText(blueprint *Blueprint, command *command) (string, error) {
	if slices.Contains(command.columns, "") {
		return "", fmt.Errorf("fulltext index column cannot be empty")
	}
	indexName := command.index
	if indexName == "" {
		indexName = g.CreateIndexName(blueprint, "fulltext", command.columns...)
	}
	language := command.language
	if language == "" {
		language = "english" // Default language for full-text search
	}
	var columns []string
	for _, col := range command.columns {
		columns = append(columns, fmt.Sprintf("to_tsvector(%s, %s)", g.QuoteString(language), col))
	}

	return fmt.Sprintf("CREATE INDEX %s ON %s USING GIN (%s)", indexName, blueprint.name, strings.Join(columns, " || ")), nil
}

func (g *postgresGrammar) CompileIndex(blueprint *Blueprint, command *command) (string, error) {
	if slices.Contains(command.columns, "") {
		return "", fmt.Errorf("index column cannot be empty")
	}
	indexName := command.index
	if indexName == "" {
		indexName = g.CreateIndexName(blueprint, "index", command.columns...)
	}

	sql := fmt.Sprintf("CREATE INDEX %s ON %s", indexName, blueprint.name)
	if command.algorithm != "" {
		sql += fmt.Sprintf(" USING %s", command.algorithm)
	}
	return fmt.Sprintf("%s (%s)", sql, g.Columnize(command.columns)), nil
}

func (g *postgresGrammar) CompileUnique(blueprint *Blueprint, command *command) (string, error) {
	if slices.Contains(command.columns, "") {
		return "", fmt.Errorf("unique index column cannot be empty")
	}
	indexName := command.index
	if indexName == "" {
		indexName = g.CreateIndexName(blueprint, "unique", command.columns...)
	}
	sql := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s UNIQUE (%s)",
		blueprint.name,
		indexName,
		g.Columnize(command.columns),
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

func (g *postgresGrammar) CompilePrimary(blueprint *Blueprint, command *command) (string, error) {
	if slices.Contains(command.columns, "") {
		return "", fmt.Errorf("primary key index column cannot be empty")
	}
	indexName := command.index
	if indexName == "" {
		indexName = g.CreateIndexName(blueprint, "primary", command.columns...)
	}
	return fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s PRIMARY KEY (%s)", blueprint.name, indexName, g.Columnize(command.columns)), nil
}

func (g *postgresGrammar) CompileDropIndex(_ *Blueprint, command *command) (string, error) {
	if command.index == "" {
		return "", fmt.Errorf("index name cannot be empty for drop operation")
	}
	return fmt.Sprintf("DROP INDEX %s", command.index), nil
}

func (g *postgresGrammar) CompileDropFulltext(blueprint *Blueprint, command *command) (string, error) {
	return g.CompileDropIndex(blueprint, command)
}

func (g *postgresGrammar) CompileDropUnique(blueprint *Blueprint, command *command) (string, error) {
	if command.index == "" {
		return "", fmt.Errorf("index name cannot be empty for drop operation")
	}
	return fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", blueprint.name, command.index), nil
}

func (g *postgresGrammar) CompileDropPrimary(blueprint *Blueprint, command *command) (string, error) {
	index := command.index
	if index == "" {
		index = g.CreateIndexName(blueprint, "primary", command.columns...)
	}
	return fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", blueprint.name, index), nil
}

func (g *postgresGrammar) CompileRenameIndex(_ *Blueprint, command *command) (string, error) {
	if command.from == "" || command.to == "" {
		return "", fmt.Errorf("index names for rename operation cannot be empty: oldName=%s, newName=%s", command.from, command.to)
	}
	return fmt.Sprintf("ALTER INDEX %s RENAME TO %s", command.from, command.to), nil
}

func (g *postgresGrammar) CompileForeign(blueprint *Blueprint, command *command) (string, error) {
	sql, err := g.baseGrammar.CompileForeign(blueprint, command)
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

func (g *postgresGrammar) CompileDropForeign(blueprint *Blueprint, command *command) (string, error) {
	if command.index == "" {
		return "", fmt.Errorf("foreign key name cannot be empty for drop operation")
	}
	return fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", blueprint.name, command.index), nil
}

func (g *postgresGrammar) GetFluentCommands() []func(blueprint *Blueprint, command *command) string {
	return []func(blueprint *Blueprint, command *command) string{
		g.CompileComment,
	}
}

func (g *postgresGrammar) CompileComment(blueprint *Blueprint, command *command) string {
	if command.column.comment != nil {
		sql := fmt.Sprintf("COMMENT ON COLUMN %s.%s IS ", blueprint.name, command.column.name)
		if command.column.comment == nil {
			return sql + "NULL"
		} else {
			return sql + fmt.Sprintf("'%s'", *command.column.comment)
		}
	}
	return ""
}

func (g *postgresGrammar) getColumns(blueprint *Blueprint) ([]string, error) {
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

func (g *postgresGrammar) getConstraints(blueprint *Blueprint) []string {
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

func (g *postgresGrammar) getType(col *columnDefinition) string {
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
	if fn, ok := typeMapFunc[col.columnType]; ok {
		return fn(col)
	}
	return col.columnType
}

func (g *postgresGrammar) typeChar(col *columnDefinition) string {
	if col.length != nil && *col.length > 0 {
		return fmt.Sprintf("CHAR(%d)", *col.length)
	}
	return "CHAR"
}

func (g *postgresGrammar) typeString(col *columnDefinition) string {
	if col.length != nil && *col.length > 0 {
		return fmt.Sprintf("VARCHAR(%d)", *col.length)
	}
	return "VARCHAR"
}

func (g *postgresGrammar) typeTinyText(_ *columnDefinition) string {
	return "VARCHAR(255)"
}

func (g *postgresGrammar) typeText(_ *columnDefinition) string {
	return "TEXT"
}

func (g *postgresGrammar) typeMediumText(_ *columnDefinition) string {
	return "TEXT"
}

func (g *postgresGrammar) typeLongText(_ *columnDefinition) string {
	return "TEXT"
}

func (g *postgresGrammar) typeBigInteger(col *columnDefinition) string {
	if col.autoIncrement != nil && *col.autoIncrement {
		return "BIGSERIAL"
	}
	return "BIGINT"
}

func (g *postgresGrammar) typeInteger(col *columnDefinition) string {
	if col.autoIncrement != nil && *col.autoIncrement {
		return "SERIAL"
	}
	return "INTEGER"
}

func (g *postgresGrammar) typeMediumInteger(col *columnDefinition) string {
	return g.typeInteger(col)
}

func (g *postgresGrammar) typeSmallInteger(col *columnDefinition) string {
	if col.autoIncrement != nil && *col.autoIncrement {
		return "SMALLSERIAL"
	}
	return "SMALLINT"
}

func (g *postgresGrammar) typeTinyInteger(col *columnDefinition) string {
	return g.typeSmallInteger(col)
}

func (g *postgresGrammar) typeFloat(_ *columnDefinition) string {
	return "REAL"
}

func (g *postgresGrammar) typeDouble(_ *columnDefinition) string {
	return "DOUBLE PRECISION"
}

func (g *postgresGrammar) typeDecimal(col *columnDefinition) string {
	return fmt.Sprintf("DECIMAL(%d, %d)", *col.total, *col.places)
}

func (g *postgresGrammar) typeBoolean(_ *columnDefinition) string {
	return "BOOLEAN"
}

func (g *postgresGrammar) typeEnum(col *columnDefinition) string {
	enumValues := make([]string, len(col.allowed))
	for i, v := range col.allowed {
		enumValues[i] = g.QuoteString(v)
	}
	return "VARCHAR(255) CHECK (" + col.name + " IN (" + strings.Join(enumValues, ", ") + "))"
}

func (g *postgresGrammar) typeJson(_ *columnDefinition) string {
	return "JSON"
}

func (g *postgresGrammar) typeJsonb(_ *columnDefinition) string {
	return "JSONB"
}

func (g *postgresGrammar) typeDate(_ *columnDefinition) string {
	return "DATE"
}

func (g *postgresGrammar) typeDateTime(col *columnDefinition) string {
	return g.typeTimestamp(col)
}

func (g *postgresGrammar) typeDateTimeTz(col *columnDefinition) string {
	return g.typeTimestampTz(col)
}

func (g *postgresGrammar) typeTime(col *columnDefinition) string {
	if col.precision != nil {
		return fmt.Sprintf("TIME(%d)", *col.precision)
	}
	return "TIME"
}

func (g *postgresGrammar) typeTimeTz(col *columnDefinition) string {
	if col.precision != nil {
		return fmt.Sprintf("TIMETZ(%d)", *col.precision)
	}
	return "TIMETZ"
}

func (g *postgresGrammar) typeTimestamp(col *columnDefinition) string {
	if col.useCurrent {
		col.SetDefault(Expression("CURRENT_TIMESTAMP"))
	}
	if col.precision != nil {
		return fmt.Sprintf("TIMESTAMP(%d)", *col.precision)
	}
	return "TIMESTAMP"
}

func (g *postgresGrammar) typeTimestampTz(col *columnDefinition) string {
	if col.useCurrent {
		col.SetDefault(Expression("CURRENT_TIMESTAMP"))
	}
	if col.precision != nil {
		return fmt.Sprintf("TIMESTAMPTZ(%d)", *col.precision)
	}
	return "TIMESTAMPTZ"
}

func (g *postgresGrammar) typeYear(_ *columnDefinition) string {
	return "INTEGER"
}

func (g *postgresGrammar) typeBinary(_ *columnDefinition) string {
	return "BYTEA"
}

func (g *postgresGrammar) typeUuid(_ *columnDefinition) string {
	return "UUID"
}

func (g *postgresGrammar) typeGeography(col *columnDefinition) string {
	if col.subtype != nil && col.srid != nil {
		return fmt.Sprintf("GEOGRAPHY(%s, %d)", *col.subtype, *col.srid)
	} else if col.subtype != nil {
		return fmt.Sprintf("GEOGRAPHY(%s)", *col.subtype)
	}
	return "GEOGRAPHY"
}

func (g *postgresGrammar) typeGeometry(col *columnDefinition) string {
	if col.subtype != nil && col.srid != nil {
		return fmt.Sprintf("GEOMETRY(%s, %d)", *col.subtype, *col.srid)
	} else if col.subtype != nil {
		return fmt.Sprintf("GEOMETRY(%s)", *col.subtype)
	}
	return "GEOMETRY"
}

func (g *postgresGrammar) typePoint(col *columnDefinition) string {
	if col.srid != nil {
		return fmt.Sprintf("POINT(%d)", *col.srid)
	}
	return "POINT"
}

func (g *postgresGrammar) modifiers() []func(*columnDefinition) string {
	return []func(*columnDefinition) string{
		g.modifyDefault,
		g.modifyNullable,
	}
}

func (g *postgresGrammar) modifyNullable(col *columnDefinition) string {
	if col.change {
		if col.nullable == nil {
			return ""
		}
		if *col.nullable {
			return " DROP NOT NULL"
		}
		return " SET NOT NULL"
	}
	if col.nullable != nil && *col.nullable {
		return " NULL"
	}
	return " NOT NULL"
}

func (g *postgresGrammar) modifyDefault(col *columnDefinition) string {
	if col.hasCommand("default") {
		if col.change {
			return fmt.Sprintf(" SET DEFAULT %s", g.GetDefaultValue(col.defaultValue))
		}
		return fmt.Sprintf(" DEFAULT %s", g.GetDefaultValue(col.defaultValue))
	}
	return ""
}
