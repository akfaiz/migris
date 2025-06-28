package schema

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

type pgGrammar struct{}

func newPgGrammar() *pgGrammar {
	return &pgGrammar{}
}

func (g *pgGrammar) compileCreate(blueprint *Blueprint) ([]string, error) {
	var result arr[string]

	columns, err := g.getColumns(blueprint)
	if err != nil {
		return nil, err
	}
	sql := fmt.Sprintf("CREATE TABLE %s (%s)", blueprint.name, strings.Join(columns, ", "))
	result.append(sql)

	if err := result.appendIfNotError(g.getIndexSqls(blueprint)); err != nil {
		return nil, err
	}
	if err := result.appendIfNotError(g.getForeignKeySqls(blueprint)); err != nil {
		return nil, err
	}

	return result.toSlice(), nil
}

func (g *pgGrammar) compileCreateIfNotExists(blueprint *Blueprint) ([]string, error) {
	var result arr[string]

	columns, err := g.getColumns(blueprint)
	if err != nil {
		return nil, err
	}
	sql := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", blueprint.name, strings.Join(columns, ", "))
	result.append(sql)

	if err := result.appendIfNotError(g.getIndexSqls(blueprint)); err != nil {
		return nil, err
	}
	if err := result.appendIfNotError(g.getForeignKeySqls(blueprint)); err != nil {
		return nil, err
	}

	return result.toSlice(), nil
}

func (g *pgGrammar) compileAlter(blueprint *Blueprint) ([]string, error) {
	var result arr[string]

	columns, err := g.getColumns(blueprint)
	if err != nil {
		return nil, err
	}
	for _, col := range columns {
		sql := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", blueprint.name, col)
		result.append(sql)
	}

	if err := result.appendIfNotError(g.getIndexSqls(blueprint)); err != nil {
		return nil, err
	}
	if err := result.appendIfNotError(g.getForeignKeySqls(blueprint)); err != nil {
		return nil, err
	}
	if err := result.appendIfNotError(g.getDropColumnSqls(blueprint)); err != nil {
		return nil, err
	}
	if err := result.appendIfNotError(g.getRenameColumnSqls(blueprint)); err != nil {
		return nil, err
	}
	if err := result.appendIfNotError(g.getDropIndexSqls(blueprint)); err != nil {
		return nil, err
	}
	if err := result.appendIfNotError(g.getRenameIndexSqls(blueprint)); err != nil {
		return nil, err
	}
	if err := result.appendIfNotError(g.getDropForeignKeySqls(blueprint)); err != nil {
		return nil, err
	}
	if err := result.appendIfNotError(g.getDropPrimaryKeySqls(blueprint)); err != nil {
		return nil, err
	}

	return result.toSlice(), nil
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

func (p *pgGrammar) getColumns(blueprint *Blueprint) ([]string, error) {
	var columns []string
	for _, col := range blueprint.getAddeddColumns() {
		if col.name == "" {
			return nil, fmt.Errorf("column name cannot be empty")
		}
		sql := col.name + " " + p.getType(col)
		if col.defaultVal != nil {
			sql += fmt.Sprintf(" DEFAULT %s", toString(col.defaultVal))
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

	return columns, nil
}

func (p *pgGrammar) getIndexSqls(blueprint *Blueprint) ([]string, error) {
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
		sql, err := p.compileIndexSql(blueprint, index)
		if err != nil {
			return nil, fmt.Errorf("error compiling index SQL: %w", err)
		}

		if sql != "" {
			sqls = append(sqls, sql)
		}
	}

	return sqls, nil
}

func (p *pgGrammar) getForeignKeySqls(blueprint *Blueprint) ([]string, error) {
	var sqls []string

	for _, foreignKey := range blueprint.foreignKeys {
		sql, err := p.compileForeignKeySql(blueprint, foreignKey)
		if err != nil {
			return nil, fmt.Errorf("error compiling foreign key SQL: %w", err)
		}
		if sql != "" {
			sqls = append(sqls, sql)
		}
	}

	return sqls, nil
}

func (p *pgGrammar) getDropColumnSqls(blueprint *Blueprint) ([]string, error) {
	var sqls []string
	for _, col := range blueprint.dropColumns {
		if col == "" {
			return nil, fmt.Errorf("column name cannot be empty for drop operation")
		}
		sqls = append(sqls, fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", blueprint.name, col))
	}
	return sqls, nil
}

func (p *pgGrammar) getRenameColumnSqls(blueprint *Blueprint) ([]string, error) {
	var sqls []string
	for oldCol, newCol := range blueprint.renameColumns {
		if oldCol == "" || newCol == "" {
			return nil, fmt.Errorf("column names for rename operation cannot be empty: oldCol=%s, newCol=%s", oldCol, newCol)
		}
		sqls = append(sqls, fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s", blueprint.name, oldCol, newCol))
	}
	return sqls, nil
}

func (p *pgGrammar) getDropIndexSqls(blueprint *Blueprint) ([]string, error) {
	var sqls []string
	for _, index := range blueprint.dropIndexes {
		if index == "" {
			return nil, fmt.Errorf("index name cannot be empty for drop operation")
		}
		sqls = append(sqls, fmt.Sprintf("DROP INDEX IF EXISTS %s", index))
	}
	for _, index := range blueprint.dropUniqueKeys {
		if index == "" {
			return nil, fmt.Errorf("unique index name cannot be empty for drop operation")
		}
		sqls = append(sqls, fmt.Sprintf("DROP INDEX IF EXISTS %s", index))
	}
	return sqls, nil
}

func (p *pgGrammar) getRenameIndexSqls(blueprint *Blueprint) ([]string, error) {
	var sqls []string
	for oldName, newName := range blueprint.renameIndexes {
		if oldName == "" || newName == "" {
			return nil, fmt.Errorf("index names for rename operation cannot be empty: oldName=%s, newName=%s", oldName, newName)
		}
		sqls = append(sqls, fmt.Sprintf("ALTER INDEX %s RENAME TO %s", oldName, newName))
	}
	return sqls, nil
}

func (p *pgGrammar) getDropForeignKeySqls(blueprint *Blueprint) ([]string, error) {
	var sqls []string
	for _, foreignKey := range blueprint.dropForeignKeys {
		if foreignKey == "" {
			return nil, fmt.Errorf("foreign key name cannot be empty for drop operation")
		}
		sqls = append(sqls, fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT IF EXISTS %s", blueprint.name, foreignKey))
	}
	return sqls, nil
}

func (p *pgGrammar) getDropPrimaryKeySqls(blueprint *Blueprint) ([]string, error) {
	var sqls []string
	for _, index := range blueprint.dropPrimaryKeys {
		if index == "" {
			return nil, fmt.Errorf("primary key index name cannot be empty for drop operation")
		}
		sql := fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT IF EXISTS %s", blueprint.name, index)
		sqls = append(sqls, sql)
	}
	return sqls, nil
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

func (p *pgGrammar) compileIndexSql(blueprint *Blueprint, index *indexDefinition) (string, error) {
	if slices.Contains(index.columns, "") {
		return "", fmt.Errorf("index column cannot be empty")
	}
	columns := strings.Join(index.columns, ", ")
	switch index.indexType {
	case indexTypePrimary:
		sql := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s PRIMARY KEY (%s)", blueprint.name, p.compileIndexName(blueprint, index), columns)
		return sql, nil
	case indexTypeUnique:
		sql := fmt.Sprintf("CREATE UNIQUE INDEX %s ON %s (%s)", p.compileIndexName(blueprint, index), blueprint.name, columns)
		return sql, nil
	case indexTypeIndex:
		sql := fmt.Sprintf("CREATE INDEX %s ON %s (%s)", p.compileIndexName(blueprint, index), blueprint.name, columns)
		if index.algorithmn != "" {
			sql += fmt.Sprintf(" USING %s", index.algorithmn)
		}
		return sql, nil
	default:
		return "", errors.New("unknown index type")
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

func (p *pgGrammar) compileForeignKeySql(blueprint *Blueprint, foreignKey *foreignKeyDefinition) (string, error) {
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

	return fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s(%s)%s%s",
		blueprint.name,
		p.compileForeginKeyName(blueprint, foreignKey),
		foreignKey.column,
		foreignKey.on,
		foreignKey.references,
		onDelete,
		onUpdate,
	), nil
}

func (p *pgGrammar) compileForeginKeyName(blueprint *Blueprint, foreignKey *foreignKeyDefinition) string {
	return fmt.Sprintf("fk_%s_%s", blueprint.name, foreignKey.on)
}
