package schema

import (
	"fmt"
	"slices"
	"strings"

	"github.com/afkdevs/go-schema/internal/util"
)

type grammar interface {
	CompileCreate(bp *Blueprint) (string, error)
	CompileAdd(bp *Blueprint) (string, error)
	CompileChange(bp *Blueprint, command *command) (string, error)
	CompileDrop(bp *Blueprint) (string, error)
	CompileDropIfExists(bp *Blueprint) (string, error)
	CompileRename(bp *Blueprint, command *command) (string, error)
	CompileDropColumn(blueprint *Blueprint, command *command) (string, error)
	CompileRenameColumn(blueprint *Blueprint, command *command) (string, error)
	CompileIndex(blueprint *Blueprint, command *command) (string, error)
	CompileUnique(blueprint *Blueprint, command *command) (string, error)
	CompilePrimary(blueprint *Blueprint, command *command) (string, error)
	CompileFullText(blueprint *Blueprint, command *command) (string, error)
	CompileDropIndex(blueprint *Blueprint, command *command) (string, error)
	CompileDropUnique(blueprint *Blueprint, command *command) (string, error)
	CompileDropFulltext(blueprint *Blueprint, command *command) (string, error)
	CompileDropPrimary(blueprint *Blueprint, command *command) (string, error)
	CompileRenameIndex(blueprint *Blueprint, command *command) (string, error)
	CompileForeign(blueprint *Blueprint, command *command) (string, error)
	CompileDropForeign(blueprint *Blueprint, command *command) (string, error)
	GetFluentCommands() []func(blueprint *Blueprint, command *command) string
	CreateIndexName(blueprint *Blueprint, idxType string, columns ...string) string
}

type baseGrammar struct{}

func (g *baseGrammar) CompileForeign(blueprint *Blueprint, command *command) (string, error) {
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
		index = g.CreateForeignKeyName(blueprint, command)
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

func (g *baseGrammar) CreateIndexName(blueprint *Blueprint, idxType string, columns ...string) string {
	tableName := blueprint.name
	if strings.Contains(tableName, ".") {
		parts := strings.Split(tableName, ".")
		tableName = parts[len(parts)-1] // Use the last part as the table name
	}

	switch idxType {
	case "primary":
		return fmt.Sprintf("pk_%s", tableName)
	case "unique":
		return fmt.Sprintf("uk_%s_%s", tableName, strings.Join(columns, "_"))
	case "index":
		return fmt.Sprintf("idx_%s_%s", tableName, strings.Join(columns, "_"))
	case "fulltext":
		return fmt.Sprintf("ft_%s_%s", tableName, strings.Join(columns, "_"))
	default:
		return ""
	}
}

func (g *baseGrammar) CreateForeignKeyName(blueprint *Blueprint, command *command) string {
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

func (g *baseGrammar) QuoteString(s string) string {
	return "'" + s + "'"
}

func (g *baseGrammar) PrefixArray(prefix string, items []string) []string {
	prefixed := make([]string, len(items))
	for i, item := range items {
		prefixed[i] = fmt.Sprintf("%s%s", prefix, item)
	}
	return prefixed
}

func (g *baseGrammar) Columnize(columns []string) string {
	if len(columns) == 0 {
		return ""
	}
	return strings.Join(columns, ", ")
}

func (g *baseGrammar) GetValue(value any) string {
	switch v := value.(type) {
	case Expression:
		return v.String()
	default:
		return fmt.Sprintf("'%v'", v)
	}
}

func (g *baseGrammar) GetDefaultValue(value any) string {
	if value == nil {
		return "NULL"
	}
	switch v := value.(type) {
	case Expression:
		return v.String()
	case bool:
		return util.Ternary(v, "'1'", "'0'")
	default:
		return fmt.Sprintf("'%v'", v)
	}
}
