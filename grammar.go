package schema

import (
	"errors"
	"fmt"
	"strings"
)

type grammar interface {
	compileTableExists(schema string, table string) (string, error)
	compileCreate(bp *Blueprint) (string, error)
	compileCreateIfNotExists(bp *Blueprint) (string, error)
	compileAdd(bp *Blueprint) (string, error)
	compileDrop(bp *Blueprint) (string, error)
	compileDropIfExists(bp *Blueprint) (string, error)
	compileRename(bp *Blueprint) (string, error)
	compileDropColumn(blueprint *Blueprint) (string, error)
	compileRenameColumn(blueprint *Blueprint, oldName, newName string) (string, error)
	compileIndexSql(blueprint *Blueprint, index *indexDefinition) (string, error)
	compileDropIndex(indexName string) (string, error)
	compileDropUnique(indexName string) (string, error)
	compileRenameIndex(blueprint *Blueprint, oldName, newName string) (string, error)
	compileDropPrimaryKey(blueprint *Blueprint, indexName string) (string, error)
	compileForeignKeySql(blueprint *Blueprint, foreignKey *foreignKeyDefinition) (string, error)
	compileDropForeignKey(blueprint *Blueprint, foreignKeyName string) (string, error)
}

type grammarImpl struct{}

var _ grammar = (*grammarImpl)(nil)

func (g *grammarImpl) compileTableExists(schema string, table string) (string, error) {
	return "", errors.New("not implemented")
}

func (g *grammarImpl) compileCreate(_ *Blueprint) (string, error) {
	return "", errors.New("not implemented")
}

func (g *grammarImpl) compileCreateIfNotExists(_ *Blueprint) (string, error) {
	return "", errors.New("not implemented")
}

func (g *grammarImpl) compileAdd(_ *Blueprint) (string, error) {
	return "", errors.New("not implemented")
}

func (g *grammarImpl) compileDrop(_ *Blueprint) (string, error) {
	return "", errors.New("not implemented")
}

func (g *grammarImpl) compileDropIfExists(_ *Blueprint) (string, error) {
	return "", errors.New("not implemented")
}

func (g *grammarImpl) compileRename(_ *Blueprint) (string, error) {
	return "", errors.New("not implemented")
}

func (g *grammarImpl) compileDropColumn(_ *Blueprint) (string, error) {
	return "", errors.New("not implemented")
}

func (g *grammarImpl) compileRenameColumn(_ *Blueprint, oldName, newName string) (string, error) {
	return "", errors.New("not implemented")
}

func (g *grammarImpl) compileIndexSql(_ *Blueprint, index *indexDefinition) (string, error) {
	return "", errors.New("not implemented")
}

func (g *grammarImpl) compileDropIndex(indexName string) (string, error) {
	return "", errors.New("not implemented")
}

func (g *grammarImpl) compileDropUnique(indexName string) (string, error) {
	return "", errors.New("not implemented")
}

func (g *grammarImpl) compileRenameIndex(blueprint *Blueprint, oldName, newName string) (string, error) {
	return "", errors.New("not implemented")
}

func (g *grammarImpl) compileDropPrimaryKey(blueprint *Blueprint, indexName string) (string, error) {
	return "", errors.New("not implemented")
}

func (g *grammarImpl) compileForeignKeySql(blueprint *Blueprint, foreignKey *foreignKeyDefinition) (string, error) {
	return "", errors.New("not implemented")
}

func (g *grammarImpl) compileDropForeignKey(blueprint *Blueprint, foreignKeyName string) (string, error) {
	return "", errors.New("not implemented")
}

func (g *grammarImpl) quoteString(s string) string {
	return "'" + s + "'"
}

func (p *grammarImpl) prefixArray(prefix string, items []string) []string {
	if len(items) == 0 {
		return nil
	}
	prefixed := make([]string, len(items))
	for i, item := range items {
		prefixed[i] = fmt.Sprintf("%s%s", prefix, item)
	}
	return prefixed
}

func (p *grammarImpl) createIndexName(blueprint *Blueprint, index *indexDefinition) string {
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

func (p *grammarImpl) createForeginKeyName(blueprint *Blueprint, foreignKey *foreignKeyDefinition) string {
	return fmt.Sprintf("fk_%s_%s", blueprint.name, foreignKey.on)
}
