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
	constraints := g.getConstraints(blueprint)
	columns = append(columns, constraints...)
	return fmt.Sprintf("CREATE TABLE %s (%s)", blueprint.name, strings.Join(columns, ", ")), nil
}

func (g *pgGrammar) compileCreateIfNotExists(blueprint *Blueprint) (string, error) {
	columns, err := g.getColumns(blueprint)
	if err != nil {
		return "", err
	}
	constraints := g.getConstraints(blueprint)
	columns = append(columns, constraints...)
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
	constraints = g.prefixArray("ADD ", constraints)

	columns = append(columns, constraints...)

	return fmt.Sprintf("ALTER TABLE %s %s",
		blueprint.name,
		strings.Join(columns, ", "),
	), nil
}

func (g *pgGrammar) compileChange(bp *Blueprint) ([]string, error) {
	if len(bp.getChangedColumns()) == 0 {
		return nil, nil
	}

	var queries []string
	for _, col := range bp.getChangedColumns() {
		if col.name == "" {
			return nil, fmt.Errorf("column name cannot be empty for change operation")
		}
		var statements []string

		statements = append(statements, fmt.Sprintf("ALTER COLUMN %s TYPE %s", col.name, g.getType(col)))
		if col.hasCommand("default") {
			if col.defaultValue != nil {
				statements = append(statements, fmt.Sprintf("ALTER COLUMN %s SET DEFAULT %s", col.name, g.getDefaultValue(col)))
			} else {
				statements = append(statements, fmt.Sprintf("ALTER COLUMN %s DROP DEFAULT", col.name))
			}
		}
		if col.hasCommand("nullable") {
			if col.nullable {
				statements = append(statements, fmt.Sprintf("ALTER COLUMN %s DROP NOT NULL", col.name))
			} else {
				statements = append(statements, fmt.Sprintf("ALTER COLUMN %s SET NOT NULL", col.name))
			}
		}
		sql := "ALTER TABLE " + bp.name + " " + strings.Join(statements, ", ")
		queries = append(queries, sql)

		if col.hasCommand("comment") {
			if col.comment != "" {
				queries = append(queries, fmt.Sprintf("COMMENT ON COLUMN %s.%s IS '%s'", bp.name, col.name, col.comment))
			} else {
				queries = append(queries, fmt.Sprintf("COMMENT ON COLUMN %s.%s IS NULL", bp.name, col.name))
			}
		}
	}

	return queries, nil
}

func (g *pgGrammar) compileDrop(blueprint *Blueprint) (string, error) {
	return fmt.Sprintf("DROP TABLE %s", blueprint.name), nil
}

func (g *pgGrammar) compileDropIfExists(blueprint *Blueprint) (string, error) {
	return fmt.Sprintf("DROP TABLE IF EXISTS %s", blueprint.name), nil
}

func (g *pgGrammar) compileRename(blueprint *Blueprint) (string, error) {
	return fmt.Sprintf("ALTER TABLE %s RENAME TO %s", blueprint.name, blueprint.newName), nil
}

func (g *pgGrammar) compileDropColumn(blueprint *Blueprint) (string, error) {
	if len(blueprint.dropColumns) == 0 {
		return "", nil
	}
	columns := g.prefixArray("DROP COLUMN ", blueprint.dropColumns)

	return fmt.Sprintf("ALTER TABLE %s %s", blueprint.name, strings.Join(columns, ", ")), nil
}

func (g *pgGrammar) compileRenameColumn(blueprint *Blueprint, oldName, newName string) (string, error) {
	if oldName == "" || newName == "" {
		return "", fmt.Errorf("table name, old column name, and new column name cannot be empty for rename operation")
	}
	return fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s", blueprint.name, oldName, newName), nil
}

func (g *pgGrammar) compileIndex(blueprint *Blueprint, index *indexDefinition) (string, error) {
	if slices.Contains(index.columns, "") {
		return "", fmt.Errorf("index column cannot be empty")
	}
	indexName := index.name
	if indexName == "" {
		indexName = g.createIndexName(blueprint, index)
	}

	sql := fmt.Sprintf("CREATE INDEX %s ON %s", indexName, blueprint.name)
	if index.algorithm != "" {
		sql += fmt.Sprintf(" USING %s", index.algorithm)
	}
	return fmt.Sprintf("%s (%s)", sql, g.columnize(index.columns)), nil
}

func (g *pgGrammar) compileUnique(blueprint *Blueprint, index *indexDefinition) (string, error) {
	if slices.Contains(index.columns, "") {
		return "", fmt.Errorf("unique index column cannot be empty")
	}
	indexName := index.name
	if indexName == "" {
		indexName = g.createIndexName(blueprint, index)
	}
	sql := fmt.Sprintf("CREATE UNIQUE INDEX %s ON %s", indexName, blueprint.name)
	if index.algorithm != "" {
		sql += fmt.Sprintf(" USING %s", index.algorithm)
	}
	sql += fmt.Sprintf(" (%s)", g.columnize(index.columns))
	if index.deferrable != nil {
		if *index.deferrable {
			sql += " DEFERRABLE"
		} else {
			sql += " NOT DEFERRABLE"
		}
	}
	if index.deferrable != nil && !*index.deferrable && index.initiallyImmediate != nil {
		if *index.initiallyImmediate {
			sql += " INITIALLY IMMEDIATE"
		} else {
			sql += " INITIALLY DEFERRED"
		}
	}

	return sql, nil
}

func (g *pgGrammar) compilePrimary(blueprint *Blueprint, index *indexDefinition) (string, error) {
	if slices.Contains(index.columns, "") {
		return "", fmt.Errorf("primary key index column cannot be empty")
	}
	indexName := index.name
	if indexName == "" {
		indexName = g.createIndexName(blueprint, index)
	}
	return fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s PRIMARY KEY (%s)", blueprint.name, indexName, g.columnize(index.columns)), nil
}

func (g *pgGrammar) compileFullText(blueprint *Blueprint, index *indexDefinition) (string, error) {
	if slices.Contains(index.columns, "") {
		return "", fmt.Errorf("fulltext index column cannot be empty")
	}
	indexName := index.name
	if indexName == "" {
		indexName = g.createIndexName(blueprint, index)
	}
	language := index.language
	if language == "" {
		language = "english" // Default language for full-text search
	}
	var columns []string
	for _, col := range index.columns {
		columns = append(columns, fmt.Sprintf("to_tsvector(%s, %s)", g.quoteString(language), col))
	}

	return fmt.Sprintf("CREATE INDEX %s ON %s USING GIN (%s)", indexName, blueprint.name, strings.Join(columns, " || ")), nil
}

func (g *pgGrammar) compileDropIndex(_ *Blueprint, indexName string) (string, error) {
	if indexName == "" {
		return "", fmt.Errorf("index name cannot be empty for drop operation")
	}
	return fmt.Sprintf("DROP INDEX %s", indexName), nil
}

func (g *pgGrammar) compileDropUnique(blueprint *Blueprint, indexName string) (string, error) {
	if indexName == "" {
		return "", fmt.Errorf("index name cannot be empty for drop operation")
	}
	return fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", blueprint.name, indexName), nil
}

func (g *pgGrammar) compileDropFulltext(_ *Blueprint, indexName string) (string, error) {
	if indexName == "" {
		return "", fmt.Errorf("index name cannot be empty for drop operation")
	}
	return fmt.Sprintf("DROP INDEX %s", indexName), nil
}

func (g *pgGrammar) compileRenameIndex(_ *Blueprint, oldName, newName string) (string, error) {
	if oldName == "" || newName == "" {
		return "", fmt.Errorf("index names for rename operation cannot be empty: oldName=%s, newName=%s", oldName, newName)
	}
	return fmt.Sprintf("ALTER INDEX %s RENAME TO %s", oldName, newName), nil
}

func (g *pgGrammar) compileDropPrimary(blueprint *Blueprint, indexName string) (string, error) {
	if indexName == "" {
		indexName = g.createIndexName(blueprint, &indexDefinition{indexType: indexTypePrimary})
	}
	return fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", blueprint.name, indexName), nil
}

func (g *pgGrammar) compileForeign(blueprint *Blueprint, foreignKey *foreignKeyDefinition) (string, error) {
	if foreignKey.column == "" || foreignKey.on == "" || foreignKey.references == "" {
		return "", fmt.Errorf("foreign key definition is incomplete: column, on, and references must be set")
	}
	onDelete := ""
	if foreignKey.onDelete != "" {
		onDelete = fmt.Sprintf(" ON DELETE %s", foreignKey.onDelete)
	}
	onUpdate := ""
	if foreignKey.onUpdate != "" {
		onUpdate = fmt.Sprintf(" ON UPDATE %s", foreignKey.onUpdate)
	}
	containtName := foreignKey.constaintName
	if containtName == "" {
		containtName = g.createForeignKeyName(blueprint, foreignKey)
	}

	sql := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s(%s)%s%s",
		blueprint.name,
		containtName,
		foreignKey.column,
		foreignKey.on,
		foreignKey.references,
		onDelete,
		onUpdate,
	)
	if foreignKey.deferrable != nil {
		if *foreignKey.deferrable {
			sql += " DEFERRABLE"
		} else {
			sql += " NOT DEFERRABLE"
		}
	}
	if foreignKey.deferrable != nil && !*foreignKey.deferrable && foreignKey.initiallyImmediate != nil {
		if *foreignKey.initiallyImmediate {
			sql += " INITIALLY IMMEDIATE"
		} else {
			sql += " INITIALLY DEFERRED"
		}
	}

	return sql, nil
}

func (g *pgGrammar) compileDropForeign(blueprint *Blueprint, foreignKeyName string) (string, error) {
	if foreignKeyName == "" {
		return "", fmt.Errorf("foreign key name cannot be empty for drop operation")
	}
	return fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", blueprint.name, foreignKeyName), nil
}

func (g *pgGrammar) getColumns(blueprint *Blueprint) ([]string, error) {
	var columns []string
	for _, col := range blueprint.getAddedColumns() {
		if col.name == "" {
			return nil, fmt.Errorf("column name cannot be empty")
		}
		sql := col.name + " " + g.getType(col)
		if col.hasCommand("default") {
			if col.defaultValue != nil {
				sql += fmt.Sprintf(" DEFAULT %s", g.getDefaultValue(col))
			} else {
				sql += " DEFAULT NULL"
			}
		}
		if col.nullable {
			sql += " NULL"
		} else {
			sql += " NOT NULL"
		}
		if col.comment != "" {
			sql += fmt.Sprintf(" COMMENT '%s'", col.comment)
		}
		columns = append(columns, sql)
	}

	return columns, nil
}

func (g *pgGrammar) getConstraints(blueprint *Blueprint) []string {
	var constrains []string
	for _, col := range blueprint.getAddedColumns() {
		if col.primary {
			pkConstraintName := g.createIndexName(blueprint, &indexDefinition{indexType: indexTypePrimary})
			sql := "CONSTRAINT " + pkConstraintName + " PRIMARY KEY (" + col.name + ")"
			constrains = append(constrains, sql)
			continue
		}
		if col.unique {
			uniqueName := col.uniqueName
			if uniqueName == "" {
				uniqueName = g.createIndexName(blueprint, &indexDefinition{
					indexType: indexTypeUnique,
					columns:   []string{col.name},
				})
			}
			sql := fmt.Sprintf("CONSTRAINT %s UNIQUE (%s)", uniqueName, col.name)
			constrains = append(constrains, sql)
		}
	}

	return constrains
}

func (g *pgGrammar) getType(col *columnDefinition) string {
	switch col.columnType {
	case columnTypeCustom:
		return col.customColumnType
	case columnTypeChar:
		if col.length > 0 {
			return fmt.Sprintf("CHAR(%d)", col.length)
		}
		return "CHAR"
	case columnTypeString:
		if col.length > 0 {
			return fmt.Sprintf("VARCHAR(%d)", col.length)
		}
		return "VARCHAR"
	case columnTypeDecimal:
		return fmt.Sprintf("DECIMAL(%d, %d)", col.total, col.places)
	case columnTypeBigInteger:
		if col.autoIncrement {
			return "BIGSERIAL"
		}
		return "BIGINT"
	case columnTypeInteger, columnTypeMediumInteger:
		if col.autoIncrement {
			return "SERIAL"
		}
		return "INTEGER"
	case columnTypeSmallInteger, columnTypeTinyInteger:
		if col.autoIncrement {
			return "SMALLSERIAL"
		}
		return "SMALLINT"
	case columnTypeTime:
		return fmt.Sprintf("TIME(%d)", col.precision)
	case columnTypeTimestamp, columnTypeDateTime:
		return fmt.Sprintf("TIMESTAMP(%d)", col.precision)
	case columnTypeTimestampTz, columnTypeDateTimeTz:
		return fmt.Sprintf("TIMESTAMPTZ(%d)", col.precision)
	case columnTypeGeography:
		return fmt.Sprintf("GEOGRAPHY(%s, %d)", col.subType, col.srid)
	case columnTypeGeometry:
		if col.srid > 0 {
			return fmt.Sprintf("GEOMETRY(%s, %d)", col.subType, col.srid)
		}
		return fmt.Sprintf("GEOMETRY(%s)", col.subType)
	case columnTypePoint:
		return fmt.Sprintf("POINT(%d)", col.srid)
	case columnTypeEnum:
		if len(col.allowedEnums) == 0 {
			return "VARCHAR(255)" // Fallback to VARCHAR if no enums are defined
		}
		enumValues := make([]string, len(col.allowedEnums))
		for i, v := range col.allowedEnums {
			enumValues[i] = fmt.Sprintf("'%s'", v)
		}
		return "VARCHAR(255) CHECK (" + col.name + " IN (" + strings.Join(enumValues, ", ") + "))"
	default:
		colType, ok := map[columnType]string{
			columnTypeBoolean:    "BOOLEAN",
			columnTypeLongText:   "TEXT",
			columnTypeText:       "TEXT",
			columnTypeMediumText: "TEXT",
			columnTypeTinyText:   "TEXT",
			columnTypeDouble:     "DOUBLE PRECISION",
			columnTypeFloat:      "REAL",
			columnTypeDate:       "DATE",
			columnTypeYear:       "INTEGER", // PostgreSQL does not have a YEAR type, using INTEGER instead
			columnTypeJSON:       "JSON",
			columnTypeJSONB:      "JSONB",
			columnTypeUUID:       "UUID",
			columnTypeBinary:     "BYTEA",
		}[col.columnType]
		if !ok {
			return "ERROR: Unknown column type " + string(col.columnType)
		}
		return colType
	}
}
