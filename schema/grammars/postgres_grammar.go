package grammars

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/akfaiz/migris/schema/blueprint"
)

type postgresGrammar struct {
	blueprint.BaseGrammar
}

func newPostgresGrammar() *postgresGrammar {
	return &postgresGrammar{}
}

func (g *postgresGrammar) CompileTableExists(_, table string) (string, error) {
	schema, table := g.parseTable(table)
	return fmt.Sprintf(
		"SELECT 1 FROM information_schema.tables WHERE table_schema = %s AND table_name = %s",
		g.QuoteString(schema),
		g.QuoteString(table),
	), nil
}

func (g *postgresGrammar) CompileTables(_ string) (string, error) {
	return `SELECT 
			t.table_name, 
			t.table_schema, 
			pg_catalog.obj_description(c.oid, 'pg_class') as table_comment 
		FROM 
			information_schema.tables t
		JOIN 
			pg_catalog.pg_class c ON c.relname = t.table_name
		JOIN 
			pg_catalog.pg_namespace n ON n.oid = c.relnamespace AND n.nspname = t.table_schema
		WHERE 
			t.table_schema = 'public' 
			AND t.table_type = 'BASE TABLE'`, nil
}

func (g *postgresGrammar) CompileColumns(_, table string) (string, error) {
	schema, table := g.parseTable(table)
	return fmt.Sprintf(`SELECT 
			column_name, 
			data_type, 
			is_nullable, 
			column_default,
			pg_catalog.col_description(c.oid, cols.ordinal_position::int) AS column_comment
		FROM 
			information_schema.columns cols
		JOIN 
			pg_catalog.pg_class c ON c.relname = cols.table_name
		JOIN 
			pg_catalog.pg_namespace n ON n.oid = c.relnamespace AND n.nspname = cols.table_schema
		WHERE 
			cols.table_schema = %s 
			AND cols.table_name = %s`, g.QuoteString(schema), g.QuoteString(table)), nil
}

func (g *postgresGrammar) CompileIndexes(_, table string) (string, error) {
	schema, table := g.parseTable(table)
	return fmt.Sprintf(`SELECT
			i.relname AS index_name,
			ix.indisunique AS is_unique,
			ix.indisprimary AS is_primary,
			a.attname AS column_name
		FROM
			pg_class t
		JOIN
			pg_index ix ON t.oid = ix.indrelid
		JOIN
			pg_class i ON i.oid = ix.indexrelid
		JOIN
			pg_namespace n ON n.oid = t.relnamespace
		CROSS JOIN
			unnest(ix.indkey) WITH ORDINALITY AS k(attnum, ord)
		JOIN
			pg_attribute a ON a.attrelid = t.oid AND a.attnum = k.attnum
		WHERE
			t.relname = %s
			AND n.nspname = %s
			AND t.relkind = 'r'
		ORDER BY
			i.relname, k.ord`, g.QuoteString(table), g.QuoteString(schema)), nil
}

func (g *postgresGrammar) parseTable(table string) (string, string) {
	parts := strings.Split(table, ".")
	if len(parts) > 1 {
		return parts[0], parts[1]
	}
	return "public", table
}

func (g *postgresGrammar) CompileCreate(bp *blueprint.Blueprint) (string, error) {
	columns, err := g.getColumns(bp)
	if err != nil {
		return "", err
	}
	columns = append(columns, g.getConstraints(bp)...)

	sql := fmt.Sprintf("CREATE TABLE %s (%s)", g.WrapTable(bp.Name, `"`), strings.Join(columns, ", "))
	return sql, nil
}

func (g *postgresGrammar) CompileAdd(bp *blueprint.Blueprint) (string, error) {
	columns, err := g.getColumns(bp)
	if err != nil {
		return "", err
	}

	for i, col := range columns {
		columns[i] = "ADD COLUMN " + col
	}

	for _, constraint := range g.getConstraints(bp) {
		columns = append(columns, "ADD "+constraint)
	}

	if len(columns) == 0 {
		return "", nil
	}

	return fmt.Sprintf("ALTER TABLE %s %s", g.WrapTable(bp.Name, `"`), strings.Join(columns, ", ")), nil
}

func (g *postgresGrammar) CompileChange(bp *blueprint.Blueprint, command *blueprint.Command) (string, error) {
	col := command.Column
	if col == nil || col.Name == "" {
		return "", errors.New("column name cannot be empty")
	}
	var sqls []string

	// Type change
	sqls = append(sqls, fmt.Sprintf("ALTER COLUMN %s TYPE %s", g.Wrap(col.Name, `"`), g.GetType(col)))

	// Nullability
	if col.NullableVal != nil {
		if *col.NullableVal {
			sqls = append(sqls, fmt.Sprintf("ALTER COLUMN %s DROP NOT NULL", g.Wrap(col.Name, `"`)))
		} else {
			sqls = append(sqls, fmt.Sprintf("ALTER COLUMN %s SET NOT NULL", g.Wrap(col.Name, `"`)))
		}
	}

	// Default
	if col.HasCommand("default") {
		sqls = append(
			sqls,
			fmt.Sprintf("ALTER COLUMN %s SET DEFAULT %s", g.Wrap(col.Name, `"`), g.GetDefaultValue(col.DefaultValue)),
		)
	} else if col.UseCurrentVal {
		sqls = append(sqls, fmt.Sprintf("ALTER COLUMN %s SET DEFAULT CURRENT_TIMESTAMP", g.Wrap(col.Name, `"`)))
	}

	return fmt.Sprintf("ALTER TABLE %s %s", g.WrapTable(bp.Name, `"`), strings.Join(sqls, ", ")), nil
}

func (g *postgresGrammar) CompileRename(blueprint *blueprint.Blueprint, command *blueprint.Command) (string, error) {
	return fmt.Sprintf(
		"ALTER TABLE %s RENAME TO %s",
		g.WrapTable(blueprint.Name, `"`),
		g.WrapTable(command.To, `"`),
	), nil
}

func (g *postgresGrammar) CompileDrop(blueprint *blueprint.Blueprint) (string, error) {
	return fmt.Sprintf("DROP TABLE %s", g.WrapTable(blueprint.Name, `"`)), nil
}

func (g *postgresGrammar) CompileDropIfExists(blueprint *blueprint.Blueprint) (string, error) {
	return fmt.Sprintf("DROP TABLE IF EXISTS %s", g.WrapTable(blueprint.Name, `"`)), nil
}

func (g *postgresGrammar) CompileDropColumn(
	blueprint *blueprint.Blueprint,
	command *blueprint.Command,
) (string, error) {
	var dropped []string
	for _, col := range command.Columns {
		dropped = append(dropped, "DROP COLUMN "+g.Wrap(col, `"`))
	}
	return fmt.Sprintf("ALTER TABLE %s %s", g.WrapTable(blueprint.Name, `"`), strings.Join(dropped, ", ")), nil
}

func (g *postgresGrammar) CompileRenameColumn(
	blueprint *blueprint.Blueprint,
	command *blueprint.Command,
) (string, error) {
	if command.From == "" || command.To == "" {
		return "", errors.New("old and new column names cannot be empty")
	}
	return fmt.Sprintf(
		"ALTER TABLE %s RENAME COLUMN %s TO %s",
		g.WrapTable(blueprint.Name, `"`),
		g.Wrap(command.From, `"`),
		g.Wrap(command.To, `"`),
	), nil
}

func (g *postgresGrammar) CompileIndex(blueprint *blueprint.Blueprint, command *blueprint.Command) (string, error) {
	if len(command.Columns) == 0 || slices.Contains(command.Columns, "") {
		return "", errors.New("index columns cannot be empty")
	}
	return g.compileKey(blueprint, command, "INDEX"), nil
}

func (g *postgresGrammar) CompileUnique(blueprint *blueprint.Blueprint, command *blueprint.Command) (string, error) {
	if len(command.Columns) == 0 || slices.Contains(command.Columns, "") {
		return "", errors.New("unique index columns cannot be empty")
	}
	index := command.Index
	if index == "" {
		index = g.CreateIndexName(blueprint, "unique", command.Columns...)
	}

	deferrable := ""
	if command.Deferrable != nil {
		if *command.Deferrable {
			deferrable = " DEFERRABLE"
		} else {
			deferrable = " NOT DEFERRABLE"
		}
	}

	if (command.Deferrable == nil || *command.Deferrable) && command.InitiallyImmediate != nil {
		if *command.InitiallyImmediate {
			deferrable += " INITIALLY IMMEDIATE"
		} else {
			deferrable += " INITIALLY DEFERRED"
		}
	}

	return fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s UNIQUE (%s)%s",
		g.WrapTable(blueprint.Name, `"`),
		g.WrapIndexName(index, `"`),
		g.WrapColumnize(command.Columns, `"`),
		deferrable,
	), nil
}

func (g *postgresGrammar) CompilePrimary(blueprint *blueprint.Blueprint, command *blueprint.Command) (string, error) {
	if len(command.Columns) == 0 || slices.Contains(command.Columns, "") {
		return "", errors.New("primary key columns cannot be empty")
	}
	index := command.Index
	if index == "" {
		index = g.CreateIndexName(blueprint, "primary", command.Columns...)
	}
	return fmt.Sprintf(
		"ALTER TABLE %s ADD CONSTRAINT %s PRIMARY KEY (%s)",
		g.WrapTable(blueprint.Name, `"`),
		g.WrapIndexName(index, `"`),
		g.WrapColumnize(command.Columns, `"`),
	), nil
}

func (g *postgresGrammar) CompileFullText(blueprint *blueprint.Blueprint, command *blueprint.Command) (string, error) {
	if len(command.Columns) == 0 || slices.Contains(command.Columns, "") {
		return "", errors.New("fulltext index columns cannot be empty")
	}
	language := command.Language
	if language == "" {
		language = "english"
	}

	columns := make([]string, len(command.Columns))
	for i, column := range command.Columns {
		columns[i] = fmt.Sprintf("to_tsvector('%s', %s)", language, column)
	}

	index := command.Index
	if index == "" {
		index = g.CreateIndexName(blueprint, "fulltext", command.Columns...)
	}

	return fmt.Sprintf("CREATE INDEX %s ON %s USING GIN (%s)",
		g.WrapIndexName(index, `"`),
		g.WrapTable(blueprint.Name, `"`),
		strings.Join(columns, " || "),
	), nil
}

func (g *postgresGrammar) CompileDropIndex(_ *blueprint.Blueprint, command *blueprint.Command) (string, error) {
	if command.Index == "" {
		return "", errors.New("index name cannot be empty")
	}
	return fmt.Sprintf("DROP INDEX %s", g.WrapIndexName(command.Index, `"`)), nil
}

func (g *postgresGrammar) CompileDropUnique(
	blueprint *blueprint.Blueprint,
	command *blueprint.Command,
) (string, error) {
	if command.Index == "" {
		return "", errors.New("index name cannot be empty")
	}
	return fmt.Sprintf(
		"ALTER TABLE %s DROP CONSTRAINT %s",
		g.WrapTable(blueprint.Name, `"`),
		g.WrapIndexName(command.Index, `"`),
	), nil
}

func (g *postgresGrammar) CompileDropFulltext(
	blueprint *blueprint.Blueprint,
	command *blueprint.Command,
) (string, error) {
	if command.Index == "" {
		return "", errors.New("index name cannot be empty")
	}
	return g.CompileDropIndex(blueprint, command)
}

func (g *postgresGrammar) CompileDropPrimary(
	blueprint *blueprint.Blueprint,
	command *blueprint.Command,
) (string, error) {
	index := command.Index
	if index == "" {
		index = blueprint.Name + "_primary"
	}
	return fmt.Sprintf(
		"ALTER TABLE %s DROP CONSTRAINT %s",
		g.WrapTable(blueprint.Name, `"`),
		g.WrapIndexName(index, `"`),
	), nil
}

func (g *postgresGrammar) CompileRenameIndex(
	_ *blueprint.Blueprint,
	command *blueprint.Command,
) (string, error) {
	if command.From == "" || command.To == "" {
		return "", errors.New("old and new index names cannot be empty")
	}
	return fmt.Sprintf(
		"ALTER INDEX %s RENAME TO %s",
		g.WrapIndexName(command.From, `"`),
		g.WrapIndexName(command.To, `"`),
	), nil
}

func (g *postgresGrammar) CompileForeign(blueprint *blueprint.Blueprint, command *blueprint.Command) (string, error) {
	if len(command.Columns) == 0 || slices.Contains(command.Columns, "") || command.On == "" ||
		len(command.References) == 0 || slices.Contains(command.References, "") {
		return "", errors.New("foreign key definition is incomplete: column, on, and references must be set")
	}
	onDelete := ""
	if command.OnDelete != "" {
		onDelete = fmt.Sprintf(" ON DELETE %s", command.OnDelete)
	}
	onUpdate := ""
	if command.OnUpdate != "" {
		onUpdate = fmt.Sprintf(" ON UPDATE %s", command.OnUpdate)
	}
	index := command.Index
	if index == "" {
		index = g.CreateForeignKeyName(blueprint, command)
	}

	deferrable := ""
	if command.Deferrable != nil {
		if *command.Deferrable {
			deferrable = " DEFERRABLE"
		} else {
			deferrable = " NOT DEFERRABLE"
		}
	}

	if (command.Deferrable == nil || *command.Deferrable) && command.InitiallyImmediate != nil {
		if *command.InitiallyImmediate {
			deferrable += " INITIALLY IMMEDIATE"
		} else {
			deferrable += " INITIALLY DEFERRED"
		}
	}

	return fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s)%s%s%s",
		g.WrapTable(blueprint.Name, `"`),
		g.WrapIndexName(index, `"`),
		g.WrapColumnize(command.Columns, `"`),
		g.WrapTable(command.On, `"`),
		g.WrapColumnize(command.References, `"`),
		onDelete,
		onUpdate,
		deferrable,
	), nil
}

func (g *postgresGrammar) CompileDropForeign(
	blueprint *blueprint.Blueprint,
	command *blueprint.Command,
) (string, error) {
	if command.Index == "" {
		return "", errors.New("index name cannot be empty")
	}
	return fmt.Sprintf(
		"ALTER TABLE %s DROP CONSTRAINT %s",
		g.WrapTable(blueprint.Name, `"`),
		g.WrapIndexName(command.Index, `"`),
	), nil
}

func (g *postgresGrammar) GetFluentCommands() []func(blueprint *blueprint.Blueprint, command *blueprint.Command) string {
	return []func(blueprint *blueprint.Blueprint, command *blueprint.Command) string{
		g.compileFluentComment,
	}
}

func (g *postgresGrammar) compileFluentComment(bp *blueprint.Blueprint, command *blueprint.Command) string {
	col := command.Column
	if col != nil && col.CommentVal != nil {
		return fmt.Sprintf(
			"COMMENT ON COLUMN %s.%s IS '%s'",
			g.WrapTable(bp.Name, `"`),
			g.Wrap(col.Name, `"`),
			*col.CommentVal,
		)
	}
	return ""
}

func (g *postgresGrammar) compileKey(
	blueprint *blueprint.Blueprint,
	command *blueprint.Command,
	keyType string,
) string {
	index := g.WrapIndexName(command.Index, `"`)
	if index == "" {
		index = g.WrapIndexName(g.CreateIndexName(blueprint, strings.ToLower(keyType), command.Columns...), `"`)
	}
	algorithm := ""
	if command.Algorithm != "" {
		algorithm = " USING " + command.Algorithm
	}
	return fmt.Sprintf(
		"CREATE %s %s ON %s%s (%s)",
		keyType,
		index,
		g.WrapTable(blueprint.Name, `"`),
		algorithm,
		g.WrapColumnize(command.Columns, `"`),
	)
}

func (g *postgresGrammar) GetType(col *blueprint.Column) string {
	if col.AutoIncrementVal != nil && *col.AutoIncrementVal {
		switch col.ColumnType {
		case blueprint.ColumnTypeInteger, blueprint.ColumnTypeMediumInteger:
			return "SERIAL"
		case blueprint.ColumnTypeSmallInteger, blueprint.ColumnTypeTinyInteger:
			return "SMALLSERIAL"
		case blueprint.ColumnTypeBigInteger:
			return "BIGSERIAL"
		}
	}
	typeMapFunc := map[string]func(*blueprint.Column) string{
		blueprint.ColumnTypeChar:          g.typeChar,
		blueprint.ColumnTypeString:        g.typeString,
		blueprint.ColumnTypeTinyText:      g.typeTinyText,
		blueprint.ColumnTypeText:          g.typeText,
		blueprint.ColumnTypeMediumText:    g.typeText,
		blueprint.ColumnTypeLongText:      g.typeText,
		blueprint.ColumnTypeInteger:       g.typeInteger,
		blueprint.ColumnTypeBigInteger:    g.typeBigInteger,
		blueprint.ColumnTypeMediumInteger: g.typeInteger,
		blueprint.ColumnTypeSmallInteger:  g.typeSmallInteger,
		blueprint.ColumnTypeTinyInteger:   g.typeSmallInteger,
		blueprint.ColumnTypeFloat:         g.typeFloat,
		blueprint.ColumnTypeDouble:        g.typeDouble,
		blueprint.ColumnTypeDecimal:       g.typeDecimal,
		blueprint.ColumnTypeBoolean:       g.typeBoolean,
		blueprint.ColumnTypeEnum:          g.typeEnum,
		blueprint.ColumnTypeJSON:          g.typeJSON,
		blueprint.ColumnTypeJSONB:         g.typeJSONB,
		blueprint.ColumnTypeDate:          g.typeDate,
		blueprint.ColumnTypeDateTime:      g.typeTimestamp,
		blueprint.ColumnTypeDateTimeTz:    g.typeTimestampTz,
		blueprint.ColumnTypeTime:          g.typeTime,
		blueprint.ColumnTypeTimeTz:        g.typeTimeTz,
		blueprint.ColumnTypeTimestamp:     g.typeTimestamp,
		blueprint.ColumnTypeTimestampTz:   g.typeTimestampTz,
		blueprint.ColumnTypeYear:          g.typeInteger,
		blueprint.ColumnTypeBinary:        g.typeBinary,
		blueprint.ColumnTypeUUID:          g.typeUUID,
		blueprint.ColumnTypeULID:          g.typeUUID,
		blueprint.ColumnTypeIPAddress:     g.typeIPAddress,
		blueprint.ColumnTypeMacAddress:    g.typeMacAddress,
		blueprint.ColumnTypeTSVector:      g.typeTSVector,
		blueprint.ColumnTypeGeography:     g.typeGeography,
		blueprint.ColumnTypeGeometry:      g.typeGeometry,
		blueprint.ColumnTypePoint:         g.typePoint,
	}
	if fn, ok := typeMapFunc[col.ColumnType]; ok {
		return fn(col)
	}
	return col.ColumnType
}

func (g *postgresGrammar) typeChar(col *blueprint.Column) string {
	return fmt.Sprintf("CHAR(%d)", *col.Length)
}

func (g *postgresGrammar) typeString(col *blueprint.Column) string {
	return fmt.Sprintf("VARCHAR(%d)", *col.Length)
}

func (g *postgresGrammar) typeText(_ *blueprint.Column) string {
	return "TEXT"
}

func (g *postgresGrammar) typeTinyText(_ *blueprint.Column) string {
	return "VARCHAR(255)"
}

func (g *postgresGrammar) typeInteger(_ *blueprint.Column) string {
	return "INTEGER"
}

func (g *postgresGrammar) typeBigInteger(_ *blueprint.Column) string {
	return "BIGINT"
}

func (g *postgresGrammar) typeSmallInteger(_ *blueprint.Column) string {
	return "SMALLINT"
}

func (g *postgresGrammar) typeFloat(_ *blueprint.Column) string {
	return "REAL"
}

func (g *postgresGrammar) typeDouble(_ *blueprint.Column) string {
	return "DOUBLE PRECISION"
}

func (g *postgresGrammar) typeDecimal(col *blueprint.Column) string {
	return fmt.Sprintf("DECIMAL(%d, %d)", *col.Total, *col.Places)
}

func (g *postgresGrammar) typeBoolean(_ *blueprint.Column) string {
	return "BOOLEAN"
}

func (g *postgresGrammar) typeEnum(col *blueprint.Column) string {
	if len(col.Allowed) > 0 {
		return fmt.Sprintf("VARCHAR(255) CHECK (%s IN ('%s'))", col.Name, strings.Join(col.Allowed, "', '"))
	}
	return "VARCHAR(255)"
}

func (g *postgresGrammar) typeJSON(_ *blueprint.Column) string {
	return "JSON"
}

func (g *postgresGrammar) typeJSONB(_ *blueprint.Column) string {
	return "JSONB"
}

func (g *postgresGrammar) typeDate(_ *blueprint.Column) string {
	return "DATE"
}

func (g *postgresGrammar) typeTimestamp(col *blueprint.Column) string {
	return g.typeDateTime(col)
}

func (g *postgresGrammar) typeTimestampTz(col *blueprint.Column) string {
	return g.typeDateTimeTz(col)
}

func (g *postgresGrammar) typeDateTime(col *blueprint.Column) string {
	if col.Precision != nil && *col.Precision > 0 {
		return fmt.Sprintf("TIMESTAMP(%d)", *col.Precision)
	}
	return "TIMESTAMP(0)"
}

func (g *postgresGrammar) typeDateTimeTz(col *blueprint.Column) string {
	if col.Precision != nil && *col.Precision > 0 {
		return fmt.Sprintf("TIMESTAMPTZ(%d)", *col.Precision)
	}
	return "TIMESTAMPTZ(0)"
}

func (g *postgresGrammar) typeTime(col *blueprint.Column) string {
	if col.Precision != nil && *col.Precision > 0 {
		return fmt.Sprintf("TIME(%d)", *col.Precision)
	}
	return "TIME(0)"
}

func (g *postgresGrammar) typeTimeTz(col *blueprint.Column) string {
	if col.Precision != nil && *col.Precision > 0 {
		return fmt.Sprintf("TIMETZ(%d)", *col.Precision)
	}
	return "TIMETZ(0)"
}

func (g *postgresGrammar) typeBinary(_ *blueprint.Column) string {
	return "BYTEA"
}

func (g *postgresGrammar) typeUUID(_ *blueprint.Column) string {
	return "UUID"
}

func (g *postgresGrammar) typeIPAddress(_ *blueprint.Column) string {
	return "inet"
}

func (g *postgresGrammar) typeMacAddress(_ *blueprint.Column) string {
	return "MACADDR"
}

func (g *postgresGrammar) typeTSVector(_ *blueprint.Column) string {
	return "tsvector"
}

func (g *postgresGrammar) typeGeography(col *blueprint.Column) string {
	if col.Subtype != nil && *col.Subtype != "" {
		return fmt.Sprintf("GEOGRAPHY(%s, %d)", strings.ToUpper(*col.Subtype), *col.Srid)
	}
	return "geography"
}

func (g *postgresGrammar) typeGeometry(col *blueprint.Column) string {
	if col.Subtype != nil && *col.Subtype != "" {
		if col.Srid != nil {
			return fmt.Sprintf("GEOMETRY(%s, %d)", strings.ToUpper(*col.Subtype), *col.Srid)
		}
		return fmt.Sprintf("GEOMETRY(%s)", strings.ToUpper(*col.Subtype))
	}
	return "geometry"
}

func (g *postgresGrammar) typePoint(col *blueprint.Column) string {
	if col.Srid != nil {
		return fmt.Sprintf("POINT(%d)", *col.Srid)
	}
	return "point"
}

func (g *postgresGrammar) getColumns(bp *blueprint.Blueprint) ([]string, error) {
	var columns []string
	for _, col := range bp.GetAddedColumns() {
		if col.Name == "" {
			return nil, errors.New("column name cannot be empty")
		}
		sql := g.Wrap(col.Name, `"`) + " " + g.GetType(col)
		sql += g.modifiers(col)
		columns = append(columns, sql)
	}
	return columns, nil
}

func (g *postgresGrammar) getConstraints(bp *blueprint.Blueprint) []string {
	var constraints []string
	for _, col := range bp.GetAddedColumns() {
		if col.PrimaryVal != nil && *col.PrimaryVal {
			constraints = append(
				constraints,
				fmt.Sprintf(
					"CONSTRAINT %s PRIMARY KEY (%s)",
					g.WrapIndexName(g.CreateIndexName(bp, "primary", col.Name), `"`),
					g.Wrap(col.Name, `"`),
				),
			)
		}
	}
	return constraints
}

func (g *postgresGrammar) modifiers(col *blueprint.Column) string {
	var sql string
	if col.NullableVal != nil {
		if *col.NullableVal {
			sql += " NULL"
		} else {
			sql += " NOT NULL"
		}
	} else {
		sql += " NOT NULL"
	}
	if col.HasCommand("default") {
		sql += fmt.Sprintf(" DEFAULT %s", g.GetDefaultValue(col.DefaultValue))
	} else if col.UseCurrentVal {
		sql += " DEFAULT CURRENT_TIMESTAMP"
	}
	return sql
}
