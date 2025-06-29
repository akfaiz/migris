package schema

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

type pgGrammar struct {
	grammarImpl
}

var _ grammar = (*pgGrammar)(nil)

func newPgGrammar() *pgGrammar {
	return &pgGrammar{}
}

func (g *pgGrammar) compileTableExists(schema string, table string) (string, error) {
	if schema == "" {
		schema = "public" // Default schema for PostgreSQL
	}
	return fmt.Sprintf(
		"SELECT 1 FROM information_schema.tables WHERE table_schema = %s AND table_name = %s AND table_type = 'BASE TABLE'",
		g.quoteString(schema),
		g.quoteString(table),
	), nil
}

func (g *pgGrammar) compileCreate(blueprint *Blueprint) (string, error) {
	columns, err := g.getColumns(blueprint)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("CREATE TABLE %s (%s)", blueprint.name, strings.Join(columns, ", ")), nil
}

func (g *pgGrammar) compileCreateIfNotExists(blueprint *Blueprint) (string, error) {
	columns, err := g.getColumns(blueprint)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", blueprint.name, strings.Join(columns, ", ")), nil
}

func (g *pgGrammar) compileAdd(blueprint *Blueprint) (string, error) {
	if len(blueprint.getAddeddColumns()) == 0 {
		return "", nil
	}

	columns, err := g.getColumns(blueprint)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("ALTER TABLE %s %s",
		blueprint.name,
		strings.Join(g.prefixArray("ADD COLUMN ", columns), ", "),
	), nil
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

func (p *pgGrammar) compileRenameColumn(blueprint *Blueprint, oldName, newName string) (string, error) {
	if oldName == "" || newName == "" {
		return "", fmt.Errorf("table name, old column name, and new column name cannot be empty for rename operation")
	}
	return fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s", blueprint.name, oldName, newName), nil
}

func (p *pgGrammar) compileIndexSql(blueprint *Blueprint, index *indexDefinition) (string, error) {
	if slices.Contains(index.columns, "") {
		return "", fmt.Errorf("index column cannot be empty")
	}
	indexName := index.name
	if indexName == "" {
		indexName = p.createIndexName(blueprint, index)
	}
	columns := strings.Join(index.columns, ", ")

	switch index.indexType {
	case indexTypePrimary:
		sql := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s PRIMARY KEY (%s)", blueprint.name, indexName, columns)
		return sql, nil
	case indexTypeUnique:
		sql := fmt.Sprintf("CREATE UNIQUE INDEX %s ON %s(%s)", indexName, blueprint.name, columns)
		return sql, nil
	case indexTypeIndex:
		sql := fmt.Sprintf("CREATE INDEX %s ON %s(%s)", indexName, blueprint.name, columns)
		if index.algorithmn != "" {
			sql += fmt.Sprintf(" USING %s", index.algorithmn)
		}
		return sql, nil
	default:
		return "", errors.New("unknown index type")
	}
}

func (g *pgGrammar) compileDropIndex(indexName string) (string, error) {
	if indexName == "" {
		return "", fmt.Errorf("index name cannot be empty for drop operation")
	}
	return fmt.Sprintf("DROP INDEX %s", indexName), nil
}

func (g *pgGrammar) compileDropUnique(indexName string) (string, error) {
	if indexName == "" {
		return "", fmt.Errorf("index name cannot be empty for drop operation")
	}
	return fmt.Sprintf("DROP INDEX %s", indexName), nil
}

func (p *pgGrammar) compileRenameIndex(blueprint *Blueprint, oldName, newName string) (string, error) {
	if oldName == "" || newName == "" {
		return "", fmt.Errorf("index names for rename operation cannot be empty: oldName=%s, newName=%s", oldName, newName)
	}
	return fmt.Sprintf("ALTER INDEX %s RENAME TO %s", oldName, newName), nil
}

func (p *pgGrammar) compileDropPrimaryKey(blueprint *Blueprint, indexName string) (string, error) {
	if indexName == "" {
		indexName = p.createIndexName(blueprint, &indexDefinition{indexType: indexTypePrimary})
	}
	return fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", blueprint.name, indexName), nil
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
		p.createForeginKeyName(blueprint, foreignKey),
		foreignKey.column,
		foreignKey.on,
		foreignKey.references,
		onDelete,
		onUpdate,
	), nil
}

func (p *pgGrammar) compileDropForeignKey(blueprint *Blueprint, foreignKeyName string) (string, error) {
	if foreignKeyName == "" {
		return "", fmt.Errorf("foreign key name cannot be empty for drop operation")
	}
	return fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", blueprint.name, foreignKeyName), nil
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

func (p *pgGrammar) getType(col *columnDefinition) string {
	switch col.columnType {
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
