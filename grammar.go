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
	compileChange(bp *Blueprint, command *command) (string, error)
	compileDrop(bp *Blueprint) (string, error)
	compileDropIfExists(bp *Blueprint) (string, error)
	compileRename(bp *Blueprint, command *command) (string, error)
	compileDropColumn(blueprint *Blueprint, command *command) (string, error)
	compileRenameColumn(blueprint *Blueprint, command *command) (string, error)
	compileIndex(blueprint *Blueprint, command *command) (string, error)
	compileUnique(blueprint *Blueprint, command *command) (string, error)
	compilePrimary(blueprint *Blueprint, command *command) (string, error)
	compileFullText(blueprint *Blueprint, command *command) (string, error)
	compileDropIndex(blueprint *Blueprint, command *command) (string, error)
	compileDropUnique(blueprint *Blueprint, command *command) (string, error)
	compileDropFulltext(blueprint *Blueprint, command *command) (string, error)
	compileDropPrimary(blueprint *Blueprint, command *command) (string, error)
	compileRenameIndex(blueprint *Blueprint, command *command) (string, error)
	compileForeign(blueprint *Blueprint, command *command) (string, error)
	compileDropForeign(blueprint *Blueprint, command *command) (string, error)
	getFluentCommands() []func(blueprint *Blueprint, command *command) string
	createIndexName(blueprint *Blueprint, idxType string, columns ...string) string
}

type baseGrammar struct{}

func (g *baseGrammar) compileForeign(blueprint *Blueprint, command *command) (string, error) {
	if len(command.columns) == 0 || slices.Contains(command.columns, "") || command.on == "" ||
		len(command.references) == 0 || slices.Contains(command.references, "") {
		return "", fmt.Errorf("foreign key definition is incomplete: column, on, and references must be set")
	}
	onDelete := ""
	if command.onDelete != "" {
		onDelete = fmt.Sprintf(" ON DELETE %s", command.onDelete)
	}
	onUpdate := ""
	if command.onUpdate != "" {
		onUpdate = fmt.Sprintf(" ON UPDATE %s", command.onUpdate)
	}
	index := command.index
	if index == "" {
		index = g.createForeignKeyName(blueprint, command)
	}

	return fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s(%s)%s%s",
		blueprint.name,
		index,
		command.columns[0],
		command.on,
		command.references[0],
		onDelete,
		onUpdate,
	), nil
}

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

func (g *baseGrammar) getValue(value any) string {
	switch v := value.(type) {
	case Expression:
		return v.String()
	default:
		return fmt.Sprintf("'%v'", v)
	}
}

func (g *baseGrammar) getDefaultValue(value any) string {
	if value == nil {
		return "NULL"
	}
	switch v := value.(type) {
	case Expression:
		return v.String()
	case bool:
		return ternary(v, "'1'", "'0'")
	default:
		return fmt.Sprintf("'%v'", v)
	}
}

func (g *baseGrammar) createIndexName(blueprint *Blueprint, idxType string, columns ...string) string {
	tableName := blueprint.name
	if strings.Contains(tableName, ".") {
		parts := strings.Split(tableName, ".")
		tableName = parts[len(parts)-1] // Use the last part as the table name
	}

	switch idxType {
	case commandPrimary:
		return fmt.Sprintf("pk_%s", tableName)
	case commandUnique:
		return fmt.Sprintf("uk_%s_%s", tableName, strings.Join(columns, "_"))
	case commandIndex:
		return fmt.Sprintf("idx_%s_%s", tableName, strings.Join(columns, "_"))
	case commandFullText:
		return fmt.Sprintf("ft_%s_%s", tableName, strings.Join(columns, "_"))
	default:
		return ""
	}
}

func (g *baseGrammar) createForeignKeyName(blueprint *Blueprint, command *command) string {
	tableName := blueprint.name
	if strings.Contains(tableName, ".") {
		parts := strings.Split(tableName, ".")
		tableName = parts[len(parts)-1] // Use the last part as the table name
	}
	on := command.on
	if strings.Contains(on, ".") {
		parts := strings.Split(on, ".")
		on = parts[len(parts)-1] // Use the last part as the column name
	}
	return fmt.Sprintf("fk_%s_%s", tableName, on)
}
