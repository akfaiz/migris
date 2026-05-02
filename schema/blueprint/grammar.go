package blueprint

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/akfaiz/migris/internal/util"
)

type Grammar interface {
	CompileTableExists(schema string, table string) (string, error)
	CompileTables(schema string) (string, error)
	CompileColumns(schema, table string) (string, error)
	CompileIndexes(schema, table string) (string, error)
	CompileCreate(bp *Blueprint) (string, error)
	CompileAdd(bp *Blueprint) (string, error)
	CompileChange(bp *Blueprint, command *Command) (string, error)
	CompileDrop(bp *Blueprint) (string, error)
	CompileDropIfExists(bp *Blueprint) (string, error)
	CompileRename(bp *Blueprint, command *Command) (string, error)
	CompileDropColumn(blueprint *Blueprint, command *Command) (string, error)
	CompileRenameColumn(blueprint *Blueprint, command *Command) (string, error)
	CompileIndex(blueprint *Blueprint, command *Command) (string, error)
	CompileUnique(blueprint *Blueprint, command *Command) (string, error)
	CompilePrimary(blueprint *Blueprint, command *Command) (string, error)
	CompileFullText(blueprint *Blueprint, command *Command) (string, error)
	CompileDropIndex(blueprint *Blueprint, command *Command) (string, error)
	CompileDropUnique(blueprint *Blueprint, command *Command) (string, error)
	CompileDropFulltext(blueprint *Blueprint, command *Command) (string, error)
	CompileDropPrimary(blueprint *Blueprint, command *Command) (string, error)
	CompileRenameIndex(blueprint *Blueprint, command *Command) (string, error)
	CompileForeign(blueprint *Blueprint, command *Command) (string, error)
	CompileDropForeign(blueprint *Blueprint, command *Command) (string, error)
	GetFluentCommands() []func(blueprint *Blueprint, command *Command) string
	GetTableFluentCommands() []func(blueprint *Blueprint) string
	CreateIndexName(blueprint *Blueprint, idxType string, columns ...string) string
	GetType(column *Column) string
}

type BaseGrammar struct{}

func (g *BaseGrammar) CompileChange(_ *Blueprint, _ *Command) (string, error) {
	return "", errors.New("change operation not supported by this grammar")
}

func (g *BaseGrammar) CompileDropColumn(_ *Blueprint, _ *Command) (string, error) {
	return "", errors.New("drop column operation not supported by this grammar")
}

func (g *BaseGrammar) CompileRenameColumn(_ *Blueprint, _ *Command) (string, error) {
	return "", errors.New("rename column operation not supported by this grammar")
}

func (g *BaseGrammar) CompileIndex(_ *Blueprint, _ *Command) (string, error) {
	return "", errors.New("index operation not supported by this grammar")
}

func (g *BaseGrammar) CompileUnique(_ *Blueprint, _ *Command) (string, error) {
	return "", errors.New("unique index operation not supported by this grammar")
}

func (g *BaseGrammar) CompilePrimary(_ *Blueprint, _ *Command) (string, error) {
	return "", errors.New("primary key operation not supported by this grammar")
}

func (g *BaseGrammar) CompileFullText(_ *Blueprint, _ *Command) (string, error) {
	return "", errors.New("fulltext index operation not supported by this grammar")
}

func (g *BaseGrammar) CompileDropIndex(_ *Blueprint, _ *Command) (string, error) {
	return "", errors.New("drop index operation not supported by this grammar")
}

func (g *BaseGrammar) CompileDropUnique(_ *Blueprint, _ *Command) (string, error) {
	return "", errors.New("drop unique operation not supported by this grammar")
}

func (g *BaseGrammar) CompileDropFulltext(_ *Blueprint, _ *Command) (string, error) {
	return "", errors.New("drop fulltext operation not supported by this grammar")
}

func (g *BaseGrammar) CompileDropPrimary(_ *Blueprint, _ *Command) (string, error) {
	return "", errors.New("drop primary operation not supported by this grammar")
}

func (g *BaseGrammar) CompileRenameIndex(_ *Blueprint, _ *Command) (string, error) {
	return "", errors.New("rename index operation not supported by this grammar")
}

func (g *BaseGrammar) CompileDropForeign(_ *Blueprint, _ *Command) (string, error) {
	return "", errors.New("drop foreign operation not supported by this grammar")
}

func (g *BaseGrammar) CompileForeign(blueprint *Blueprint, command *Command) (string, error) {
	if len(command.Columns) == 0 || slices.Contains(command.Columns, "") || command.On == "" ||
		len(command.References) == 0 || slices.Contains(command.References, "") {
		return "", errors.New("foreign key definition is incomplete: column, on, and references must be set")
	}
	onDelete := ""
	if command.OnDelete != "" {
		onDelete = fmt.Sprintf(" ON DELETE %s", command.OnDelete)
	}
	onUpdate := ""
	if command.OnUpdate != "" {
		onUpdate = fmt.Sprintf(" ON UPDATE %s", command.OnUpdate)
	}
	index := command.Index
	if index == "" {
		index = g.CreateForeignKeyName(blueprint, command)
	}

	return fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s)%s%s",
		blueprint.Name,
		index,
		g.Columnize(command.Columns),
		command.On,
		g.Columnize(command.References),
		onDelete,
		onUpdate,
	), nil
}

func (g *BaseGrammar) GetFluentCommands() []func(blueprint *Blueprint, command *Command) string {
	return []func(blueprint *Blueprint, command *Command) string{}
}

func (g *BaseGrammar) GetTableFluentCommands() []func(blueprint *Blueprint) string {
	return []func(blueprint *Blueprint) string{}
}

func (g *BaseGrammar) CreateIndexName(blueprint *Blueprint, idxType string, columns ...string) string {
	parts := []string{blueprint.Name}
	parts = append(parts, columns...)
	parts = append(parts, idxType)

	index := strings.ToLower(strings.Join(parts, "_"))
	return strings.NewReplacer("-", "_", ".", "_").Replace(index)
}

func (g *BaseGrammar) CreateForeignKeyName(blueprint *Blueprint, command *Command) string {
	return g.CreateIndexName(blueprint, "foreign", command.Columns...)
}

func (g *BaseGrammar) QuoteString(s string) string {
	return "'" + s + "'"
}

func (g *BaseGrammar) PrefixArray(prefix string, items []string) []string {
	prefixed := make([]string, len(items))
	for i, item := range items {
		prefixed[i] = fmt.Sprintf("%s%s", prefix, item)
	}
	return prefixed
}

func (g *BaseGrammar) Columnize(columns []string) string {
	if len(columns) == 0 {
		return ""
	}
	return strings.Join(columns, ", ")
}

func (g *BaseGrammar) Wrap(ident string, quote string) string {
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

func (g *BaseGrammar) WrapTable(table string, quote string) string {
	return g.Wrap(table, quote)
}

func (g *BaseGrammar) WrapColumnize(columns []string, quote string) string {
	if len(columns) == 0 {
		return ""
	}
	wrapped := make([]string, 0, len(columns))
	for _, column := range columns {
		wrapped = append(wrapped, g.Wrap(column, quote))
	}
	return strings.Join(wrapped, ", ")
}

func (g *BaseGrammar) WrapIndexName(index string, quote string) string {
	if index == "" {
		return index
	}
	if regexp.MustCompile(`[()\s]`).MatchString(index) {
		return index
	}
	return g.Wrap(index, quote)
}

func (g *BaseGrammar) GetValue(value any) string {
	switch v := value.(type) {
	case Expression:
		return v.String()
	default:
		return fmt.Sprintf("'%v'", v)
	}
}

func (g *BaseGrammar) GetDefaultValue(value any) string {
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
