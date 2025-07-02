package schema

import (
	"fmt"
	"slices"
	"strings"
)

type grammar interface {
	compileCreate(bp *Blueprint) (string, error)
	compileCreateIfNotExists(bp *Blueprint) (string, error)
	compileAdd(bp *Blueprint) (string, error)
	compileChange(bp *Blueprint) ([]string, error)
	compileDrop(bp *Blueprint) (string, error)
	compileDropIfExists(bp *Blueprint) (string, error)
	compileRename(bp *Blueprint) (string, error)
	compileDropColumn(blueprint *Blueprint) (string, error)
	compileRenameColumn(blueprint *Blueprint, oldName, newName string) (string, error)
	compileIndex(blueprint *Blueprint, index *indexDefinition) (string, error)
	compileUnique(blueprint *Blueprint, index *indexDefinition) (string, error)
	compilePrimary(blueprint *Blueprint, index *indexDefinition) (string, error)
	compileFullText(blueprint *Blueprint, index *indexDefinition) (string, error)
	compileDropIndex(indexName string) (string, error)
	compileDropUnique(indexName string) (string, error)
	compileDropFulltext(indexName string) (string, error)
	compileDropPrimary(blueprint *Blueprint, indexName string) (string, error)
	compileRenameIndex(blueprint *Blueprint, oldName, newName string) (string, error)
	compileForeign(blueprint *Blueprint, foreignKey *foreignKeyDefinition) (string, error)
	compileDropForeign(blueprint *Blueprint, foreignKeyName string) (string, error)
}

type baseGrammar struct{}

func (g *baseGrammar) quoteString(s string) string {
	return "'" + s + "'"
}

func (g *baseGrammar) prefixArray(prefix string, items []string) []string {
	prefixed := make([]string, len(items))
	for i, item := range items {
		prefixed[i] = fmt.Sprintf("%s%s", prefix, item)
	}
	return prefixed
}

func (g *baseGrammar) columnize(columns []string) string {
	if len(columns) == 0 {
		return ""
	}
	return strings.Join(columns, ", ")
}

func (g *baseGrammar) getDefaultValue(col *columnDefinition) string {
	if col.defaultValue == nil {
		return "NULL"
	}
	useQuote := slices.Contains([]columnType{columnTypeString, columnTypeChar, columnTypeEnum}, col.columnType)

	switch v := col.defaultValue.(type) {
	case string:
		if useQuote {
			return g.quoteString(v)
		}
		return v
	case int, int64, float64:
		return fmt.Sprintf("%v", v)
	case bool:
		if v {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("'%v'", v) // Fallback for other types
	}
}

func (g *baseGrammar) createIndexName(blueprint *Blueprint, index *indexDefinition) string {
	switch index.indexType {
	case indexTypePrimary:
		return fmt.Sprintf("pk_%s", blueprint.name)
	case indexTypeUnique:
		return fmt.Sprintf("uk_%s_%s", blueprint.name, strings.Join(index.columns, "_"))
	case indexTypeIndex:
		return fmt.Sprintf("idx_%s_%s", blueprint.name, strings.Join(index.columns, "_"))
	case indexTypeFulltext:
		return fmt.Sprintf("idx_%s_%s", blueprint.name, strings.Join(index.columns, "_"))
	default:
		return ""
	}
}

func (g *baseGrammar) createForeignKeyName(blueprint *Blueprint, foreignKey *foreignKeyDefinition) string {
	return fmt.Sprintf("fk_%s_%s", blueprint.name, foreignKey.on)
}
