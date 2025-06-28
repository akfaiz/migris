package schema

import (
	"fmt"
	"strings"
)

type pgGrammar struct{}

func newPgGrammar() *pgGrammar {
	return &pgGrammar{}
}

func (g *pgGrammar) compileCreate(blueprint *Blueprint) ([]string, error) {
	var sqls []string
	sql := fmt.Sprintf("CREATE TABLE %s (%s)", blueprint.name, strings.Join(g.getColumns(blueprint), ", "))
	sqls = append(sqls, sql)
	sqls = append(sqls, g.getIndexSqls(blueprint)...)
	sqls = append(sqls, g.getForeignKeySqls(blueprint)...)

	return sqls, nil
}

func (g *pgGrammar) compileCreateIfNotExists(blueprint *Blueprint) ([]string, error) {
	var sqls []string
	sql := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", blueprint.name, strings.Join(g.getColumns(blueprint), ", "))
	sqls = append(sqls, sql)
	sqls = append(sqls, g.getIndexSqls(blueprint)...)
	sqls = append(sqls, g.getForeignKeySqls(blueprint)...)

	return sqls, nil
}

func (g *pgGrammar) compileAlter(blueprint *Blueprint) ([]string, error) {
	var sqls []string
	for _, col := range g.getColumns(blueprint) {
		sql := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", blueprint.name, col)
		sqls = append(sqls, sql)
	}
	sqls = append(sqls, g.getIndexSqls(blueprint)...)
	sqls = append(sqls, g.getForeignKeySqls(blueprint)...)
	sqls = append(sqls, g.getDropColumnSqls(blueprint)...)
	sqls = append(sqls, g.getRenameColumnSqls(blueprint)...)
	sqls = append(sqls, g.getDropIndexSqls(blueprint)...)
	sqls = append(sqls, g.getDropForeignKeySqls(blueprint)...)
	sqls = append(sqls, g.getDropPrimaryKeySqls(blueprint)...)

	return sqls, nil
}

func (g *pgGrammar) compileDrop(blueprint *Blueprint) (string, error) {
	return fmt.Sprintf("DROP TABLE %s", blueprint.name), nil
}

func (g *pgGrammar) compileDropIfExists(blueprint *Blueprint) (string, error) {
	return fmt.Sprintf("DROP TABLE IF EXISTS %s", blueprint.name), nil
}

func (g *pgGrammar) compileRename(blueprint *Blueprint) (string, error) {
	if blueprint.newName == "" {
		return "", fmt.Errorf("new name must be set for renaming table")
	}
	return fmt.Sprintf("ALTER TABLE %s RENAME TO %s", blueprint.name, blueprint.newName), nil
}

func (p *pgGrammar) getColumns(blueprint *Blueprint) []string {
	var columns []string
	for _, col := range blueprint.getAddeddColumns() {
		sql := col.name + " " + p.getType(col)
		if col.defaultVal != nil {
			sql += p.getDefaultValue(col)
		}
		if col.nullable {
			sql += " NULL"
		} else {
			sql += " NOT NULL"
		}
		if col.comment != "" {
			sql += fmt.Sprintf(" COMMENT '%s'", col.comment)
		}
		if col.primary {
			sql += " PRIMARY KEY"
		}
		if col.unique && col.uniqueIndexName == "" {
			sql += " UNIQUE"
		}
		columns = append(columns, sql)
	}

	return columns
}

func (p *pgGrammar) getDefaultValue(col *columnDefinition) string {
	if col.defaultVal == nil {
		return ""
	}
	switch col.columnType {
	case columnTypeBoolean:
		return fmt.Sprintf(" DEFAULT %t", col.defaultVal)
	case columnTypeString, columnTypeText, columnTypeChar:
		return fmt.Sprintf(" DEFAULT '%s'", col.defaultVal)
	case columnTypeDate, columnTypeTime, columnTypeTimestamp, columnTypeTimestampTz:
		return fmt.Sprintf(" DEFAULT %s", col.defaultVal)
	case columnTypeDecimal, columnTypeDouble, columnTypeFloat:
		return fmt.Sprintf(" DEFAULT %f", col.defaultVal)
	case columnTypeInteger, columnTypeBigInteger, columnTypeSmallInteger, columnTypeIncrements, columnTypeBigIncrements, columnTypeSmallIncrements:
		return fmt.Sprintf(" DEFAULT %d", col.defaultVal)
	case columnTypeJSON, columnTypeJSONB:
		if str, ok := col.defaultVal.(string); ok {
			return fmt.Sprintf(" DEFAULT '%s'", str)
		}
		return fmt.Sprintf(" DEFAULT '%v'", col.defaultVal) // Fallback for non-string
	default:
		return fmt.Sprintf(" DEFAULT '%v'", col.defaultVal)
	}
}

func (p *pgGrammar) getIndexSqls(blueprint *Blueprint) []string {
	var sqls []string

	for _, col := range blueprint.getAddeddColumns() {
		if col.unique && col.uniqueIndexName != "" {
			sqls = append(sqls, "CREATE UNIQUE INDEX "+col.uniqueIndexName+" ON "+blueprint.name+"("+col.name+")")
		}
		if col.index {
			indexName := col.indexName
			if indexName == "" {
				indexName = p.compileIndexName(blueprint, &indexDefinition{
					tableName: blueprint.name,
					indexType: indexTypeIndex,
					columns:   []string{col.name},
				})
			}
			sqls = append(sqls, "CREATE INDEX "+indexName+" ON "+blueprint.name+"("+col.name+")")
		}
	}

	for _, index := range blueprint.indexes {
		sql := p.compileIndexSql(blueprint, index)
		if sql != "" {
			sqls = append(sqls, sql)
		}
	}

	return sqls
}

func (p *pgGrammar) getForeignKeySqls(blueprint *Blueprint) []string {
	var sqls []string

	for _, foreignKey := range blueprint.foreignKeys {
		sql := p.compileForeignKeySql(blueprint, foreignKey)
		if sql != "" {
			sqls = append(sqls, sql)
		}
	}

	return sqls
}

func (p *pgGrammar) getDropColumnSqls(blueprint *Blueprint) []string {
	var sqls []string
	for _, col := range blueprint.dropColumns {
		sqls = append(sqls, fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", blueprint.name, col))
	}
	return sqls
}

func (p *pgGrammar) getRenameColumnSqls(blueprint *Blueprint) []string {
	var sqls []string
	for oldCol, newCol := range blueprint.renameColumns {
		sqls = append(sqls, fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s", blueprint.name, oldCol, newCol))
	}
	return sqls
}

func (p *pgGrammar) getDropIndexSqls(blueprint *Blueprint) []string {
	var sqls []string
	for _, index := range blueprint.dropIndexes {
		sqls = append(sqls, fmt.Sprintf("DROP INDEX IF EXISTS %s", index))
	}
	for _, col := range blueprint.dropUniqueKeys {
		sqls = append(sqls, fmt.Sprintf("DROP INDEX IF EXISTS %s", col))
	}
	return sqls
}

func (p *pgGrammar) getDropForeignKeySqls(blueprint *Blueprint) []string {
	var sqls []string
	for _, foreignKey := range blueprint.dropForeignKeys {
		sqls = append(sqls, fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT IF EXISTS %s", blueprint.name, foreignKey))
	}
	return sqls
}

func (p *pgGrammar) getDropPrimaryKeySqls(blueprint *Blueprint) []string {
	var sqls []string
	for _, index := range blueprint.dropPrimaryKeys {
		sqls = append(sqls, fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT IF EXISTS %s", blueprint.name, p.compileIndexName(blueprint, &indexDefinition{
			tableName: blueprint.name,
			indexType: indexTypePrimary,
			columns:   []string{index},
		})))
	}
	return sqls
}

func (p *pgGrammar) getType(col *columnDefinition) string {
	switch col.columnType {
	case columnTypeChar:
		return fmt.Sprintf("CHAR(%d)", col.length)
	case columnTypeString:
		return fmt.Sprintf("VARCHAR(%d)", col.length)
	case columnTypeDecimal:
		return fmt.Sprintf("DECIMAL(%d, %d)", col.precision, col.scale)
	case columnTypeTime:
		return fmt.Sprintf("TIME(%d)", col.precision)
	case columnTypeTimestamp:
		return fmt.Sprintf("TIMESTAMP(%d)", col.precision)
	case columnTypeTimestampTz:
		return fmt.Sprintf("TIMESTAMPTZ(%d)", col.precision)
	default:
		return map[columnType]string{
			columnTypeBoolean:         "BOOLEAN",
			columnTypeText:            "TEXT",
			columnTypeBigInteger:      "BIGINT",
			columnTypeInteger:         "INTEGER",
			columnTypeSmallInteger:    "SMALLINT",
			columnTypeBigIncrements:   "BIGSERIAL",
			columnTypeIncrements:      "SERIAL",
			columnTypeSmallIncrements: "SMALLSERIAL",
			columnTypeDouble:          "DOUBLE PRECISION",
			columnTypeFloat:           "REAL",
			columnTypeDate:            "DATE",
			columnTypeYear:            "INTEGER", // PostgreSQL does not have a YEAR type, using INTEGER instead
			columnTypeJSON:            "JSON",
			columnTypeJSONB:           "JSONB",
			columnTypeUUID:            "UUID",
		}[col.columnType]
	}
}

func (p *pgGrammar) compileIndexSql(blueprint *Blueprint, index *indexDefinition) string {
	columns := strings.Join(index.columns, ", ")
	switch index.indexType {
	case indexTypePrimary:
		return fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s PRIMARY KEY (%s)", blueprint.name, p.compileIndexName(blueprint, index), columns)
	case indexTypeUnique:
		return fmt.Sprintf("CREATE UNIQUE INDEX %s ON %s (%s)", p.compileIndexName(blueprint, index), blueprint.name, columns)
	case indexTypeIndex:
		sql := fmt.Sprintf("CREATE INDEX %s ON %s (%s)", p.compileIndexName(blueprint, index), blueprint.name, columns)
		if index.algorithmn != "" {
			sql += fmt.Sprintf(" USING %s", index.algorithmn)
		}
		return sql
	default:
		return ""
	}
}

func (p *pgGrammar) compileIndexName(blueprint *Blueprint, index *indexDefinition) string {
	switch index.indexType {
	case indexTypePrimary:
		return fmt.Sprintf("pk_%s", blueprint.name)
	case indexTypeUnique:
		return fmt.Sprintf("uk_%s_%s", blueprint.name, strings.Join(index.columns, "_"))
	case indexTypeIndex:
		return fmt.Sprintf("idx_%s_%s", blueprint.name, strings.Join(index.columns, "_"))
	default:
		return ""
	}
}

func (p *pgGrammar) compileForeignKeySql(blueprint *Blueprint, foreignKey *foreignKeyDefinition) string {
	onDelete := ""
	if foreignKey.onDelete != "" {
		onDelete = fmt.Sprintf(" ON DELETE %s", foreignKey.onDelete)
	}
	onUpdate := ""
	if foreignKey.onUpdate != "" {
		onUpdate = fmt.Sprintf(" ON UPDATE %s", foreignKey.onUpdate)
	}

	return fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s(%s)%s%s",
		blueprint.name,
		p.compileForeginKeyName(blueprint, foreignKey),
		foreignKey.column,
		foreignKey.on,
		foreignKey.references,
		onDelete,
		onUpdate,
	)
}

func (p *pgGrammar) compileForeginKeyName(blueprint *Blueprint, foreignKey *foreignKeyDefinition) string {
	return fmt.Sprintf("fk_%s_%s", blueprint.name, foreignKey.on)
}
