package grammars

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/akfaiz/migris/internal/util"
	"github.com/akfaiz/migris/schema/blueprint"
)

type mysqlGrammar struct {
	blueprint.BaseGrammar

	serials []string
	self    blueprint.Grammar
}

func newMysqlGrammar() *mysqlGrammar {
	g := &mysqlGrammar{
		serials: []string{
			blueprint.ColumnTypeBigInteger,
			blueprint.ColumnTypeInteger,
			blueprint.ColumnTypeMediumInteger,
			blueprint.ColumnTypeSmallInteger,
			blueprint.ColumnTypeTinyInteger,
		},
	}
	g.self = g
	return g
}

func (g *mysqlGrammar) CompileTableExists(_, table string) (string, error) {
	schema, table := g.parseTable(table)
	schemaClause := "table_schema = DATABASE()"
	if schema != "" {
		schemaClause = "table_schema = " + g.QuoteString(schema)
	}
	return fmt.Sprintf(
		"SELECT 1 FROM information_schema.tables WHERE %s AND table_name = %s",
		schemaClause,
		g.QuoteString(table),
	), nil
}

func (g *mysqlGrammar) CompileTables(_ string) (string, error) {
	return "SELECT table_name, table_comment FROM information_schema.tables WHERE table_schema = DATABASE() AND table_type = 'BASE TABLE'", nil
}

func (g *mysqlGrammar) CompileColumns(_, table string) (string, error) {
	return fmt.Sprintf("SHOW FULL COLUMNS FROM %s", g.WrapTable(table, "`")), nil
}

func (g *mysqlGrammar) CompileIndexes(_, table string) (string, error) {
	return fmt.Sprintf("SHOW INDEX FROM %s", g.WrapTable(table, "`")), nil
}

func (g *mysqlGrammar) parseTable(table string) (string, string) {
	parts := strings.Split(table, ".")
	if len(parts) > 1 {
		return parts[0], parts[1]
	}
	return "", table
}

func (g *mysqlGrammar) CompileCreate(bp *blueprint.Blueprint) (string, error) {
	columns, err := g.getColumns(bp)
	if err != nil {
		return "", err
	}
	columns = append(columns, g.getConstraints(bp)...)

	create := "CREATE TABLE"
	if bp.TemporaryVal {
		create = "CREATE TEMPORARY TABLE"
	}
	sql := fmt.Sprintf("%s %s (%s)", create, g.WrapTable(bp.Name, "`"), strings.Join(columns, ", "))
	sql += g.compileTableModifiers(bp)

	return sql, nil
}

func (g *mysqlGrammar) CompileAdd(bp *blueprint.Blueprint) (string, error) {
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
	return fmt.Sprintf("ALTER TABLE %s %s", g.WrapTable(bp.Name, "`"), strings.Join(columns, ", ")), nil
}

func (g *mysqlGrammar) CompileChange(bp *blueprint.Blueprint, command *blueprint.Command) (string, error) {
	column := command.Column
	if column.Name == "" {
		return "", errors.New("column name cannot be empty for change operation")
	}

	operation := "MODIFY"
	name := g.Wrap(column.Name, "`")
	if column.RenameToVal != "" {
		operation = "CHANGE"
		name = fmt.Sprintf("%s %s", name, g.Wrap(column.RenameToVal, "`"))
	}

	sql := fmt.Sprintf(
		"ALTER TABLE %s %s COLUMN %s %s",
		g.WrapTable(bp.Name, "`"),
		operation,
		name,
		g.self.GetType(column),
	)
	var sqlBuilder strings.Builder
	for _, modifier := range g.modifiers() {
		sqlBuilder.WriteString(modifier(column))
	}
	sql += sqlBuilder.String()

	return sql, nil
}

func (g *mysqlGrammar) CompileDrop(blueprint *blueprint.Blueprint) (string, error) {
	if blueprint.Name == "" {
		return "", errors.New("table name cannot be empty")
	}
	return fmt.Sprintf("DROP TABLE %s", g.WrapTable(blueprint.Name, "`")), nil
}

func (g *mysqlGrammar) CompileDropIfExists(blueprint *blueprint.Blueprint) (string, error) {
	if blueprint.Name == "" {
		return "", errors.New("table name cannot be empty")
	}
	return fmt.Sprintf("DROP TABLE IF EXISTS %s", g.WrapTable(blueprint.Name, "`")), nil
}

func (g *mysqlGrammar) CompileDropColumn(blueprint *blueprint.Blueprint, command *blueprint.Command) (string, error) {
	var dropped []string
	for _, col := range command.Columns {
		if col == "" {
			return "", errors.New("column name cannot be empty")
		}
		dropped = append(dropped, "DROP COLUMN "+g.Wrap(col, "`"))
	}
	return fmt.Sprintf("ALTER TABLE %s %s", g.WrapTable(blueprint.Name, "`"), strings.Join(dropped, ", ")), nil
}

func (g *mysqlGrammar) CompileRenameColumn(blueprint *blueprint.Blueprint, command *blueprint.Command) (string, error) {
	if command.From == "" || command.To == "" {
		return "", errors.New("old and new column names cannot be empty")
	}
	return fmt.Sprintf(
		"ALTER TABLE %s RENAME COLUMN %s TO %s",
		g.WrapTable(blueprint.Name, "`"),
		g.Wrap(command.From, "`"),
		g.Wrap(command.To, "`"),
	), nil
}

func (g *mysqlGrammar) CompileIndex(blueprint *blueprint.Blueprint, command *blueprint.Command) (string, error) {
	if len(command.Columns) == 0 || slices.Contains(command.Columns, "") {
		return "", errors.New("index columns cannot be empty")
	}
	return g.compileKey(blueprint, command, "INDEX"), nil
}

func (g *mysqlGrammar) CompileUnique(blueprint *blueprint.Blueprint, command *blueprint.Command) (string, error) {
	if len(command.Columns) == 0 || slices.Contains(command.Columns, "") {
		return "", errors.New("unique index columns cannot be empty")
	}
	return g.compileKey(blueprint, command, "UNIQUE INDEX"), nil
}

func (g *mysqlGrammar) CompileFullText(blueprint *blueprint.Blueprint, command *blueprint.Command) (string, error) {
	if len(command.Columns) == 0 || slices.Contains(command.Columns, "") {
		return "", errors.New("fulltext index columns cannot be empty")
	}
	return g.compileKey(blueprint, command, "FULLTEXT INDEX"), nil
}

func (g *mysqlGrammar) CompilePrimary(blueprint *blueprint.Blueprint, command *blueprint.Command) (string, error) {
	if len(command.Columns) == 0 || slices.Contains(command.Columns, "") {
		return "", errors.New("primary key columns cannot be empty")
	}
	index := command.Index
	if index == "" {
		index = g.CreateIndexName(blueprint, "primary", command.Columns...)
	}

	return fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s PRIMARY KEY (%s)",
		g.WrapTable(blueprint.Name, "`"),
		g.WrapIndexName(index, "`"),
		g.WrapColumnize(command.Columns, "`"),
	), nil
}

func (g *mysqlGrammar) compileKey(blueprint *blueprint.Blueprint, command *blueprint.Command, keyType string) string {
	idxType := strings.ToLower(keyType)
	switch {
	case strings.Contains(idxType, "unique"):
		idxType = "unique"
	case strings.Contains(idxType, "fulltext"):
		idxType = "fulltext"
	case strings.Contains(idxType, "index"):
		idxType = "index"
	}

	index := g.WrapIndexName(command.Index, "`")
	if index == "" {
		index = g.WrapIndexName(g.CreateIndexName(blueprint, idxType, command.Columns...), "`")
	}

	columns := g.WrapColumnize(command.Columns, "`")
	algorithm := ""
	if command.Algorithm != "" {
		algorithm = " USING " + strings.ToUpper(command.Algorithm)
	}

	return fmt.Sprintf(
		"CREATE %s %s ON %s (%s)%s",
		keyType,
		index,
		g.WrapTable(blueprint.Name, "`"),
		columns,
		algorithm,
	)
}

func (g *mysqlGrammar) CompileDropIndex(blueprint *blueprint.Blueprint, command *blueprint.Command) (string, error) {
	if command.Index == "" {
		return "", errors.New("index name cannot be empty")
	}
	return fmt.Sprintf(
		"ALTER TABLE %s DROP INDEX %s",
		g.WrapTable(blueprint.Name, "`"),
		g.WrapIndexName(command.Index, "`"),
	), nil
}

func (g *mysqlGrammar) CompileDropUnique(blueprint *blueprint.Blueprint, command *blueprint.Command) (string, error) {
	if command.Index == "" {
		return "", errors.New("unique index name cannot be empty")
	}
	return fmt.Sprintf(
		"ALTER TABLE %s DROP INDEX %s",
		g.WrapTable(blueprint.Name, "`"),
		g.WrapIndexName(command.Index, "`"),
	), nil
}

func (g *mysqlGrammar) CompileDropFulltext(blueprint *blueprint.Blueprint, command *blueprint.Command) (string, error) {
	if command.Index == "" {
		return "", errors.New("fulltext index name cannot be empty")
	}
	return fmt.Sprintf(
		"ALTER TABLE %s DROP INDEX %s",
		g.WrapTable(blueprint.Name, "`"),
		g.WrapIndexName(command.Index, "`"),
	), nil
}

func (g *mysqlGrammar) CompileDropPrimary(blueprint *blueprint.Blueprint, _ *blueprint.Command) (string, error) {
	return fmt.Sprintf("ALTER TABLE %s DROP PRIMARY KEY", g.WrapTable(blueprint.Name, "`")), nil
}

func (g *mysqlGrammar) CompileRenameIndex(blueprint *blueprint.Blueprint, command *blueprint.Command) (string, error) {
	if command.From == "" || command.To == "" {
		return "", errors.New("old and new index names cannot be empty")
	}
	return fmt.Sprintf(
		"ALTER TABLE %s RENAME INDEX %s TO %s",
		g.WrapTable(blueprint.Name, "`"),
		g.WrapIndexName(command.From, "`"),
		g.WrapIndexName(command.To, "`"),
	), nil
}

func (g *mysqlGrammar) CompileDropForeign(blueprint *blueprint.Blueprint, command *blueprint.Command) (string, error) {
	if command.Index == "" {
		return "", errors.New("foreign key name cannot be empty")
	}
	return fmt.Sprintf(
		"ALTER TABLE %s DROP FOREIGN KEY %s",
		g.WrapTable(blueprint.Name, "`"),
		g.WrapIndexName(command.Index, "`"),
	), nil
}

func (g *mysqlGrammar) CompileForeign(blueprint *blueprint.Blueprint, command *blueprint.Command) (string, error) {
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

	return fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s)%s%s",
		g.WrapTable(blueprint.Name, "`"),
		g.WrapIndexName(index, "`"),
		g.WrapColumnize(command.Columns, "`"),
		g.WrapTable(command.On, "`"),
		g.WrapColumnize(command.References, "`"),
		onDelete,
		onUpdate,
	), nil
}

func (g *mysqlGrammar) getColumns(blueprint *blueprint.Blueprint) ([]string, error) {
	var columns []string
	for _, col := range blueprint.GetAddedColumns() {
		if col.Name == "" {
			return nil, errors.New("column name cannot be empty")
		}

		sql := g.Wrap(col.Name, "`") + " " + g.self.GetType(col)
		var sqlBuilder strings.Builder
		for _, modifier := range g.modifiers() {
			sqlBuilder.WriteString(modifier(col))
		}
		sql += sqlBuilder.String()
		columns = append(columns, sql)
	}
	return columns, nil
}

func (g *mysqlGrammar) getConstraints(blueprint *blueprint.Blueprint) []string {
	var constrains []string
	for _, col := range blueprint.GetAddedColumns() {
		if col.PrimaryVal != nil && *col.PrimaryVal {
			pkConstraintName := g.CreateIndexName(blueprint, "primary", col.Name)
			sql := "CONSTRAINT " + g.WrapIndexName(
				pkConstraintName,
				"`",
			) + " PRIMARY KEY (" + g.Wrap(
				col.Name,
				"`",
			) + ")"
			constrains = append(constrains, sql)
			continue
		}
	}
	return constrains
}

func (g *mysqlGrammar) GetType(col *blueprint.Column) string {
	typeFuncMap := g.getTypeFuncMap()
	if fn, ok := typeFuncMap[col.ColumnType]; ok {
		return fn(col)
	}
	return col.ColumnType
}

func (g *mysqlGrammar) getTypeFuncMap() map[string]func(*blueprint.Column) string {
	return map[string]func(*blueprint.Column) string{
		blueprint.ColumnTypeChar:          g.typeChar,
		blueprint.ColumnTypeString:        g.typeString,
		blueprint.ColumnTypeTinyText:      g.typeTinyText,
		blueprint.ColumnTypeText:          g.typeText,
		blueprint.ColumnTypeMediumText:    g.typeMediumText,
		blueprint.ColumnTypeLongText:      g.typeLongText,
		blueprint.ColumnTypeInteger:       g.typeInteger,
		blueprint.ColumnTypeBigInteger:    g.typeBigInteger,
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
		blueprint.ColumnTypeSet:           g.typeSet,
		blueprint.ColumnTypeVector:        g.typeVector,
		blueprint.ColumnTypeGeography:     g.typeGeography,
		blueprint.ColumnTypeGeometry:      g.typeGeometry,
		blueprint.ColumnTypePoint:         g.typePoint,
		blueprint.ColumnTypeRaw:           g.typeRaw,
	}
}

func (g *mysqlGrammar) typeChar(col *blueprint.Column) string {
	return fmt.Sprintf("CHAR(%d)", *col.Length)
}

func (g *mysqlGrammar) typeString(col *blueprint.Column) string {
	return fmt.Sprintf("VARCHAR(%d)", *col.Length)
}

func (g *mysqlGrammar) typeTinyText(_ *blueprint.Column) string {
	return "TINYTEXT"
}

func (g *mysqlGrammar) typeText(_ *blueprint.Column) string {
	return "TEXT"
}

func (g *mysqlGrammar) typeMediumText(_ *blueprint.Column) string {
	return "MEDIUMTEXT"
}

func (g *mysqlGrammar) typeLongText(_ *blueprint.Column) string {
	return "LONGTEXT"
}

func (g *mysqlGrammar) typeBigInteger(_ *blueprint.Column) string {
	return "BIGINT"
}

func (g *mysqlGrammar) typeInteger(_ *blueprint.Column) string {
	return "INT"
}

func (g *mysqlGrammar) typeMediumInteger(_ *blueprint.Column) string {
	return "MEDIUMINT"
}

func (g *mysqlGrammar) typeSmallInteger(_ *blueprint.Column) string {
	return "SMALLINT"
}

func (g *mysqlGrammar) typeTinyInteger(_ *blueprint.Column) string {
	return "TINYINT"
}

func (g *mysqlGrammar) typeFloat(col *blueprint.Column) string {
	if col.Precision != nil && *col.Precision > 0 {
		return fmt.Sprintf("FLOAT(%d)", *col.Precision)
	}
	return "FLOAT"
}

func (g *mysqlGrammar) typeDouble(_ *blueprint.Column) string {
	return "DOUBLE"
}

func (g *mysqlGrammar) typeDecimal(col *blueprint.Column) string {
	return fmt.Sprintf("DECIMAL(%d, %d)", *col.Total, *col.Places)
}

func (g *mysqlGrammar) typeBoolean(_ *blueprint.Column) string {
	return "TINYINT(1)"
}

func (g *mysqlGrammar) typeEnum(col *blueprint.Column) string {
	return fmt.Sprintf("ENUM(%s)", g.QuoteString(strings.Join(col.Allowed, "', '")))
}

func (g *mysqlGrammar) typeSet(col *blueprint.Column) string {
	return fmt.Sprintf("SET(%s)", g.QuoteString(strings.Join(col.Allowed, "', '")))
}

func (g *mysqlGrammar) typeJSON(_ *blueprint.Column) string {
	return "JSON"
}

func (g *mysqlGrammar) typeJSONB(_ *blueprint.Column) string {
	return "JSON"
}

func (g *mysqlGrammar) typeDate(_ *blueprint.Column) string {
	return "DATE"
}

func (g *mysqlGrammar) typeDateTime(col *blueprint.Column) string {
	if col.Precision != nil && *col.Precision > 0 {
		return fmt.Sprintf("DATETIME(%d)", *col.Precision)
	}
	return "DATETIME"
}

func (g *mysqlGrammar) typeDateTimeTz(col *blueprint.Column) string {
	return g.typeDateTime(col)
}

func (g *mysqlGrammar) typeTime(col *blueprint.Column) string {
	if col.Precision != nil && *col.Precision > 0 {
		return fmt.Sprintf("TIME(%d)", *col.Precision)
	}
	return "TIME"
}

func (g *mysqlGrammar) typeTimeTz(col *blueprint.Column) string {
	return g.typeTime(col)
}

func (g *mysqlGrammar) typeTimestamp(col *blueprint.Column) string {
	if col.Precision != nil && *col.Precision > 0 {
		return fmt.Sprintf("TIMESTAMP(%d)", *col.Precision)
	}
	return "TIMESTAMP"
}

func (g *mysqlGrammar) typeTimestampTz(col *blueprint.Column) string {
	return g.typeTimestamp(col)
}

func (g *mysqlGrammar) typeYear(_ *blueprint.Column) string {
	return "YEAR"
}

func (g *mysqlGrammar) typeBinary(col *blueprint.Column) string {
	if col.Length != nil && *col.Length > 0 {
		if col.FixedVal != nil && *col.FixedVal {
			return fmt.Sprintf("BINARY(%d)", *col.Length)
		}
		return fmt.Sprintf("VARBINARY(%d)", *col.Length)
	}
	return "BLOB"
}

func (g *mysqlGrammar) typeUUID(_ *blueprint.Column) string {
	return "CHAR(36)"
}

func (g *mysqlGrammar) typeULID(_ *blueprint.Column) string {
	return "CHAR(26)"
}

func (g *mysqlGrammar) typeIPAddress(_ *blueprint.Column) string {
	return "VARCHAR(45)"
}

func (g *mysqlGrammar) typeMacAddress(_ *blueprint.Column) string {
	return "VARCHAR(17)"
}

func (g *mysqlGrammar) typeVector(col *blueprint.Column) string {
	if col.Places != nil {
		return fmt.Sprintf("VECTOR(%d)", *col.Places)
	}
	return "VECTOR"
}

func (g *mysqlGrammar) typeGeography(col *blueprint.Column) string {
	return g.typeGeometry(col)
}

func (g *mysqlGrammar) typeGeometry(col *blueprint.Column) string {
	subtype := util.Ternary(col.Subtype != nil, util.PtrOf(strings.ToUpper(*col.Subtype)), nil)
	if subtype != nil {
		if !slices.Contains(
			[]string{
				"POINT",
				"LINESTRING",
				"POLYGON",
				"GEOMETRYCOLLECTION",
				"MULTIPOINT",
				"MULTILINESTRING",
				"MULTIPOLYGON",
			},
			*subtype,
		) {
			subtype = nil
		}
	}

	if subtype == nil {
		subtype = util.PtrOf("GEOMETRY")
	}

	if col.Srid != nil && *col.Srid > 0 {
		return fmt.Sprintf("%s SRID %d", *subtype, *col.Srid)
	}

	return *subtype
}

func (g *mysqlGrammar) typePoint(col *blueprint.Column) string {
	if col.Srid != nil && *col.Srid > 0 {
		return fmt.Sprintf("POINT SRID %d", *col.Srid)
	}
	return "POINT"
}

func (g *mysqlGrammar) typeRaw(col *blueprint.Column) string {
	if col.RawDefinition != nil {
		return *col.RawDefinition
	}
	return ""
}

func (g *mysqlGrammar) modifiers() []func(*blueprint.Column) string {
	return []func(*blueprint.Column) string{
		g.modifyUnsigned,
		g.modifyCharset,
		g.modifyCollate,
		g.modifyVirtualAs,
		g.modifyStoredAs,
		g.modifyNullable,
		g.modifyDefault,
		g.modifyIncrement,
		g.modifyOnUpdate,
		g.modifyInvisible,
		g.modifyComment,
		g.modifyAfter,
		g.modifyFirst,
	}
}

func (g *mysqlGrammar) compileTableModifiers(bp *blueprint.Blueprint) string {
	var sql string
	if bp.CharsetVal != "" {
		sql += " DEFAULT CHARACTER SET " + bp.CharsetVal
	}
	if bp.CollationVal != "" {
		sql += " COLLATE " + bp.CollationVal
	}
	if bp.EngineVal != "" {
		sql += " ENGINE = " + bp.EngineVal
	}
	if bp.CommentVal != "" {
		sql += " COMMENT = " + g.QuoteString(bp.CommentVal)
	}
	return sql
}

func (g *mysqlGrammar) modifyNullable(col *blueprint.Column) string {
	if col.VirtualAsVal != nil || col.StoredAsVal != nil {
		if col.NullableVal != nil && !*col.NullableVal {
			return " NOT NULL"
		}
		return ""
	}
	if col.NullableVal != nil && *col.NullableVal {
		return " NULL"
	}
	return " NOT NULL"
}

func (g *mysqlGrammar) modifyDefault(col *blueprint.Column) string {
	if col.DefaultValue != nil {
		return fmt.Sprintf(" DEFAULT %s", g.GetDefaultValue(col.DefaultValue))
	}
	if col.HasCommand("default") {
		return " DEFAULT NULL"
	}
	if col.UseCurrentVal {
		return " DEFAULT CURRENT_TIMESTAMP"
	}
	return ""
}

func (g *mysqlGrammar) modifyIncrement(col *blueprint.Column) string {
	if slices.Contains(g.serials, col.ColumnType) && col.AutoIncrementVal != nil && *col.AutoIncrementVal {
		if col.PrimaryVal != nil || col.ChangeVal {
			return " AUTO_INCREMENT"
		}
		return " AUTO_INCREMENT PRIMARY KEY"
	}
	return ""
}

func (g *mysqlGrammar) modifyComment(col *blueprint.Column) string {
	if col.CommentVal != nil {
		return fmt.Sprintf(" COMMENT %s", g.QuoteString(*col.CommentVal))
	}
	return ""
}

func (g *mysqlGrammar) modifyAfter(col *blueprint.Column) string {
	if col.AfterVal != nil {
		return fmt.Sprintf(" AFTER %s", g.Wrap(*col.AfterVal, "`"))
	}
	return ""
}

func (g *mysqlGrammar) modifyFirst(col *blueprint.Column) string {
	if col.FirstVal {
		return " FIRST"
	}
	return ""
}

func (g *mysqlGrammar) modifyOnUpdate(col *blueprint.Column) string {
	if col.UseCurrentOnUpdateVal {
		return " ON UPDATE CURRENT_TIMESTAMP"
	}
	if col.OnUpdateValue != nil {
		return fmt.Sprintf(" ON UPDATE %s", g.GetValue(col.OnUpdateValue))
	}
	return ""
}

func (g *mysqlGrammar) modifyUnsigned(col *blueprint.Column) string {
	if col.UnsignedVal != nil && *col.UnsignedVal {
		return " UNSIGNED"
	}
	return ""
}

func (g *mysqlGrammar) modifyCharset(col *blueprint.Column) string {
	if col.CharsetVal != nil {
		return fmt.Sprintf(" CHARACTER SET %s", *col.CharsetVal)
	}
	return ""
}

func (g *mysqlGrammar) modifyCollate(col *blueprint.Column) string {
	if col.CollationVal != nil {
		return fmt.Sprintf(" COLLATE %s", *col.CollationVal)
	}
	return ""
}

func (g *mysqlGrammar) modifyVirtualAs(col *blueprint.Column) string {
	if col.VirtualAsVal != nil {
		return fmt.Sprintf(" GENERATED ALWAYS AS (%s) VIRTUAL", *col.VirtualAsVal)
	}
	return ""
}

func (g *mysqlGrammar) modifyStoredAs(col *blueprint.Column) string {
	if col.StoredAsVal != nil {
		return fmt.Sprintf(" GENERATED ALWAYS AS (%s) STORED", *col.StoredAsVal)
	}
	return ""
}

func (g *mysqlGrammar) modifyInvisible(col *blueprint.Column) string {
	if col.InvisibleVal != nil && *col.InvisibleVal {
		return " INVISIBLE"
	}
	return ""
}

func (g *mysqlGrammar) CompileSpatialIndex(bp *blueprint.Blueprint, command *blueprint.Command) (string, error) {
	if len(command.Columns) == 0 || slices.Contains(command.Columns, "") {
		return "", errors.New("spatial index columns cannot be empty")
	}
	index := g.WrapIndexName(command.Index, "`")
	if index == "" {
		index = g.WrapIndexName(g.CreateIndexName(bp, "spatialindex", command.Columns...), "`")
	}
	return fmt.Sprintf(
		"CREATE SPATIAL INDEX %s ON %s (%s)",
		index,
		g.WrapTable(bp.Name, "`"),
		g.WrapColumnize(command.Columns, "`"),
	), nil
}

func (g *mysqlGrammar) CompileDropSpatialIndex(bp *blueprint.Blueprint, command *blueprint.Command) (string, error) {
	return g.CompileDropIndex(bp, command)
}

func (g *mysqlGrammar) CompileTableComment(bp *blueprint.Blueprint, command *blueprint.Command) (string, error) {
	if bp.IsCreating() {
		return "", nil
	}
	return fmt.Sprintf(
		"ALTER TABLE %s COMMENT = %s",
		g.WrapTable(bp.Name, "`"),
		g.QuoteString(strings.ReplaceAll(command.Comment, "'", "''")),
	), nil
}

func (g *mysqlGrammar) CompileAutoIncrementStartingValues(
	bp *blueprint.Blueprint, command *blueprint.Command,
) (string, error) {
	if command.Value <= 0 {
		return "", nil
	}
	return fmt.Sprintf(
		"ALTER TABLE %s AUTO_INCREMENT = %d",
		g.WrapTable(bp.Name, "`"),
		command.Value,
	), nil
}

func (g *mysqlGrammar) CompileRename(bp *blueprint.Blueprint, command *blueprint.Command) (string, error) {
	return fmt.Sprintf(
		"RENAME TABLE %s TO %s",
		g.WrapTable(bp.Name, "`"),
		g.WrapTable(command.To, "`"),
	), nil
}

func (g *mysqlGrammar) GetFluentCommands() []func(*blueprint.Blueprint, *blueprint.Command) string {
	return []func(*blueprint.Blueprint, *blueprint.Command) string{}
}

func (g *mysqlGrammar) GetTableFluentCommands() []func(*blueprint.Blueprint) string {
	return []func(*blueprint.Blueprint) string{}
}
