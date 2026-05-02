package schema

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/akfaiz/migris/internal/util"
)

type grammar interface {
	CompileTableExists(schema string, table string) (string, error)
	CompileTables(schema string) (string, error)
	CompileColumns(schema, table string) (string, error)
	CompileIndexes(schema, table string) (string, error)
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
	GetTableFluentCommands() []func(blueprint *Blueprint) string
	CreateIndexName(blueprint *Blueprint, idxType string, columns ...string) string
	getType(column *columnDefinition) string
}

type baseGrammar struct{}

func (g *baseGrammar) CompileForeign(blueprint *Blueprint, command *command) (string, error) {
	if len(command.columns) == 0 || slices.Contains(command.columns, "") || command.on == "" ||
		len(command.references) == 0 || slices.Contains(command.references, "") {
		return "", errors.New("foreign key definition is incomplete: column, on, and references must be set")
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

	return fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s)%s%s",
		blueprint.name,
		index,
		g.Columnize(command.columns),
		command.on,
		g.Columnize(command.references),
		onDelete,
		onUpdate,
	), nil
}

func (g *baseGrammar) GetFluentCommands() []func(blueprint *Blueprint, command *command) string {
	return []func(blueprint *Blueprint, command *command) string{}
}

func (g *baseGrammar) GetTableFluentCommands() []func(blueprint *Blueprint) string {
	return []func(blueprint *Blueprint) string{}
}

func (g *baseGrammar) CreateIndexName(blueprint *Blueprint, idxType string, columns ...string) string {
	parts := []string{blueprint.name}
	parts = append(parts, columns...)
	parts = append(parts, idxType)

	index := strings.ToLower(strings.Join(parts, "_"))
	return strings.NewReplacer("-", "_", ".", "_").Replace(index)
}

func (g *baseGrammar) CreateForeignKeyName(blueprint *Blueprint, command *command) string {
	return g.CreateIndexName(blueprint, "foreign", command.columns...)
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

func (g *baseGrammar) Wrap(ident string, quote string) string {
	if ident == "" {
		return ident
	}
	parts := strings.Split(ident, ".")
	wrapped := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "*" {
			wrapped = append(wrapped, part)
			continue
		}
		wrapped = append(wrapped, quote+part+quote)
	}
	return strings.Join(wrapped, ".")
}

func (g *baseGrammar) WrapTable(table string, quote string) string {
	return g.Wrap(table, quote)
}

func (g *baseGrammar) WrapColumnize(columns []string, quote string) string {
	if len(columns) == 0 {
		return ""
	}
	wrapped := make([]string, 0, len(columns))
	for _, column := range columns {
		wrapped = append(wrapped, g.Wrap(column, quote))
	}
	return strings.Join(wrapped, ", ")
}

func (g *baseGrammar) WrapIndexName(index string, quote string) string {
	if index == "" {
		return index
	}
	// Keep raw expressions such as function calls unwrapped.
	if regexp.MustCompile(`[()\s]`).MatchString(index) {
		return index
	}
	return g.Wrap(index, quote)
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
