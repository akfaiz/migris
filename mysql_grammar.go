package schema

import (
	"errors"
	"fmt"
	"strings"
)

type mysqlGrammar struct {
	baseGrammar
}

var _ grammar = (*mysqlGrammar)(nil)

func newMysqlGrammar() *mysqlGrammar {
	return &mysqlGrammar{}
}

func (g *mysqlGrammar) compileTableExists(database string, table string) (string, error) {
	return fmt.Sprintf(
		"SELECT 1 FROM information_schema.tables WHERE table_schema = %s AND table_name = %s AND table_type = 'BASE TABLE'",
		g.quoteString(database),
		g.quoteString(table),
	), nil
}

func (g *mysqlGrammar) compileTables(database string) (string, error) {
	return fmt.Sprintf(
		"select table_name as `name`, (data_length + index_length) as `size`, "+
			"table_comment as `comment`, engine as `engine`, table_collation as `collation` "+
			"from information_schema.tables where table_schema = %s and table_type in ('BASE TABLE', 'SYSTEM VERSIONED') "+
			"order by table_name",
		g.quoteString(database),
	), nil
}

func (g *mysqlGrammar) compileColumns(database, table string) (string, error) {
	return fmt.Sprintf(
		"select column_name as `name`, data_type as `type_name`, column_type as `type`, "+
			"collation_name as `collation`, is_nullable as `nullable`, "+
			"column_default as `default`, column_comment as `comment`, extra as `extra` "+
			"from information_schema.columns where table_schema = %s and table_name = %s "+
			"order by ordinal_position asc",
		g.quoteString(database),
		g.quoteString(table),
	), nil
}

func (g *mysqlGrammar) compileIndexes(database, table string) (string, error) {
	return fmt.Sprintf(
		"select index_name as `name`, group_concat(column_name order by seq_in_index) as `columns`, "+
			"index_type as `type`, not non_unique as `unique` "+
			"from information_schema.statistics where table_schema = %s and table_name = %s "+
			"group by index_name, index_type, non_unique",
		g.quoteString(database),
		g.quoteString(table),
	), nil
}

func (g *mysqlGrammar) compileCreate(blueprint *Blueprint) (string, error) {
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

func (g *mysqlGrammar) compileCreateIfNotExists(blueprint *Blueprint) (string, error) {
	return g.compileCreate(blueprint)
}

func (g *mysqlGrammar) compileAdd(blueprint *Blueprint) (string, error) {
	columns, err := g.getColumns(blueprint)
	if err != nil {
		return "", err
	}
	if len(columns) == 0 {
		return "", fmt.Errorf("no columns to add")
	}
	return fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", blueprint.name, strings.Join(columns, ", ")), nil
}

func (g *mysqlGrammar) compileChange(bp *Blueprint) ([]string, error) {
	if len(bp.getChangedColumns()) == 0 {
		return nil, nil
	}

	var sqls []string
	for _, col := range bp.getChangedColumns() {
		if col.name == "" {
			return nil, fmt.Errorf("column name cannot be empty for change operation")
		}
		sql := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s", bp.name, col.name, g.getType(col))
		if col.defaultVal != nil {
			sql += fmt.Sprintf(" SET DEFAULT %s", g.getDefaultValue(col))
		}
		if col.nullable {
			sql += " DROP NOT NULL"
		} else {
			sql += " SET NOT NULL"
		}
		if col.comment != "" {
			sql += fmt.Sprintf(" COMMENT '%s'", col.comment)
		}
		sqls = append(sqls, sql)
	}

	return sqls, nil
}

func (g *mysqlGrammar) compileRename(blueprint *Blueprint) (string, error) {
	if blueprint.newName == "" {
		return "", fmt.Errorf("new table name cannot be empty")
	}
	return fmt.Sprintf("ALTER TABLE %s RENAME TO %s", blueprint.name, blueprint.newName), nil
}

func (g *mysqlGrammar) compileDrop(blueprint *Blueprint) (string, error) {
	if blueprint.name == "" {
		return "", fmt.Errorf("table name cannot be empty")
	}
	return fmt.Sprintf("DROP TABLE %s", blueprint.name), nil
}

func (g *mysqlGrammar) compileDropIfExists(blueprint *Blueprint) (string, error) {
	if blueprint.name == "" {
		return "", fmt.Errorf("table name cannot be empty")
	}
	return fmt.Sprintf("DROP TABLE IF EXISTS %s", blueprint.name), nil
}

func (g *mysqlGrammar) compileDropColumn(blueprint *Blueprint) (string, error) {
	if len(blueprint.dropColumns) == 0 {
		return "", fmt.Errorf("no columns to drop")
	}
	columns := make([]string, len(blueprint.dropColumns))
	for i, col := range blueprint.dropColumns {
		if col == "" {
			return "", fmt.Errorf("column name cannot be empty")
		}
		columns[i] = col
	}
	columns = g.prefixArray("DROP COLUMN ", columns)
	return fmt.Sprintf("ALTER TABLE %s %s", blueprint.name, strings.Join(columns, ", ")), nil
}

func (g *mysqlGrammar) compileDropForeignKey(blueprint *Blueprint, foreignKeyName string) (string, error) {
	if foreignKeyName == "" {
		return "", fmt.Errorf("foreign key name cannot be empty")
	}
	return fmt.Sprintf("ALTER TABLE %s DROP FOREIGN KEY %s", blueprint.name, foreignKeyName), nil
}

func (g *mysqlGrammar) compileDropIndex(indexName string) (string, error) {
	if indexName == "" {
		return "", fmt.Errorf("index name cannot be empty")
	}
	return fmt.Sprintf("DROP INDEX %s", indexName), nil
}

func (g *mysqlGrammar) compileDropUnique(indexName string) (string, error) {
	if indexName == "" {
		return "", fmt.Errorf("unique index name cannot be empty")
	}
	return fmt.Sprintf("DROP INDEX %s", indexName), nil
}

func (g *mysqlGrammar) compileDropPrimaryKey(blueprint *Blueprint, indexName string) (string, error) {
	if indexName == "" {
		return "", fmt.Errorf("primary key index name cannot be empty")
	}
	return fmt.Sprintf("ALTER TABLE %s DROP PRIMARY KEY", blueprint.name), nil
}

func (g *mysqlGrammar) compileRenameIndex(blueprint *Blueprint, oldName, newName string) (string, error) {
	if oldName == "" || newName == "" {
		return "", fmt.Errorf("old and new index names cannot be empty")
	}
	return fmt.Sprintf("ALTER TABLE %s RENAME INDEX %s TO %s", blueprint.name, oldName, newName), nil
}

func (g *mysqlGrammar) compileRenameColumn(blueprint *Blueprint, oldName, newName string) (string, error) {
	if oldName == "" || newName == "" {
		return "", fmt.Errorf("old and new column names cannot be empty")
	}
	return fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s", blueprint.name, oldName, newName), nil
}

func (g *mysqlGrammar) compileIndexSql(blueprint *Blueprint, index *indexDefinition) (string, error) {
	if index == nil || len(index.columns) == 0 {
		return "", fmt.Errorf("index definition cannot be nil or empty")
	}

	indexName := index.name
	if indexName == "" {
		indexName = g.createIndexName(blueprint, index)
	}
	columns := strings.Join(index.columns, ", ")

	switch index.indexType {
	case indexTypePrimary:
		sql := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s PRIMARY KEY (%s)", blueprint.name, indexName, columns)
		return sql, nil
	case indexTypeUnique:
		sql := fmt.Sprintf("CREATE UNIQUE INDEX %s ON %s (%s)", indexName, blueprint.name, columns)
		if index.algorithmn != "" {
			sql += fmt.Sprintf(" USING %s", index.algorithmn)
		}
		return sql, nil
	case indexTypeIndex:
		sql := fmt.Sprintf("CREATE INDEX %s ON %s (%s)", indexName, blueprint.name, columns)
		if index.algorithmn != "" {
			sql += fmt.Sprintf(" USING %s", index.algorithmn)
		}
		return sql, nil
	default:
		return "", errors.New("unknown index type")
	}
}

func (g *mysqlGrammar) compileForeignKeySql(blueprint *Blueprint, foreignKey *foreignKeyDefinition) (string, error) {
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
		g.createForeginKeyName(blueprint, foreignKey),
		foreignKey.column,
		foreignKey.on,
		foreignKey.references,
		onDelete,
		onUpdate,
	), nil
}

func (g *mysqlGrammar) getColumns(blueprint *Blueprint) ([]string, error) {
	var columns []string
	for _, col := range blueprint.getAddeddColumns() {
		if col.name == "" {
			return nil, fmt.Errorf("column name cannot be empty")
		}
		sql := col.name + " " + g.getType(col)
		if col.defaultVal != nil {
			sql += fmt.Sprintf(" DEFAULT %s", g.getDefaultValue(col))
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

func (g *mysqlGrammar) getType(col *columnDefinition) string {
	switch col.columnType {
	case columnTypeBoolean:
		return "TINYINT(1)"
	case columnTypeChar:
		return fmt.Sprintf("CHAR(%d)", col.length)
	case columnTypeString:
		return fmt.Sprintf("VARCHAR(%d)", col.length)
	case columnTypeDecimal:
		return fmt.Sprintf("DECIMAL(%d, %d)", col.precision, col.scale)
	case columnTypeDouble, columnTypeFloat:
		if col.precision > 0 && col.scale > 0 {
			return fmt.Sprintf("DOUBLE(%d, %d)", col.precision, col.scale)
		}
		return "DOUBLE"
	case columnTypeBigInteger:
		return g.modifyUnsignedAndAutoIncrement("BIGINT", col)
	case columnTypeInteger:
		return g.modifyUnsignedAndAutoIncrement("INT", col)
	case columnTypeSmallInteger:
		return g.modifyUnsignedAndAutoIncrement("SMALLINT", col)
	case columnTypeTime:
		return fmt.Sprintf("TIME(%d)", col.precision)
	case columnTypeTimestamp:
		return fmt.Sprintf("TIMESTAMP(%d)", col.precision)
	case columnTypeTimestampTz:
		return fmt.Sprintf("TIMESTAMPTZ(%d)", col.precision)
	case columnTypeGeography:
		return fmt.Sprintf("GEOGRAPHY(%s, %d)", col.subType, col.srid)
	case columnTypeEnum:
		return fmt.Sprintf("ENUM(%s)", g.quoteString(strings.Join(col.allowedEnums, "','")))
	default:
		return map[columnType]string{
			columnTypeBoolean:  "BOOLEAN",
			columnTypeText:     "TEXT",
			columnTypeDate:     "DATE",
			columnTypeYear:     "YEAR",
			columnTypeJSON:     "JSON",
			columnTypeJSONB:    "JSON",
			columnTypeUUID:     "UUID",
			columnTypeBinary:   "BLOB",
			columnTypeGeometry: "GEOMETRY",
		}[col.columnType]
	}
}

func (g *mysqlGrammar) modifyUnsignedAndAutoIncrement(sql string, col *columnDefinition) string {
	if col.unsigned {
		sql += " UNSIGNED"
	}
	if col.autoIncrement {
		sql += " AUTO_INCREMENT"
	}
	return sql
}
