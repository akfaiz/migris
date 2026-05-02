package blueprint_test

import (
	"testing"

	"github.com/akfaiz/migris/schema/blueprint"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBaseGrammar_UnsupportedOperations(t *testing.T) {
	g := &blueprint.BaseGrammar{}
	bp := &blueprint.Blueprint{Name: "users"}
	cmd := &blueprint.Command{}

	tests := []struct {
		name string
		fn   func() (string, error)
	}{
		{"CompileChange", func() (string, error) { return g.CompileChange(bp, cmd) }},
		{"CompileDropColumn", func() (string, error) { return g.CompileDropColumn(bp, cmd) }},
		{"CompileRenameColumn", func() (string, error) { return g.CompileRenameColumn(bp, cmd) }},
		{"CompileIndex", func() (string, error) { return g.CompileIndex(bp, cmd) }},
		{"CompileUnique", func() (string, error) { return g.CompileUnique(bp, cmd) }},
		{"CompilePrimary", func() (string, error) { return g.CompilePrimary(bp, cmd) }},
		{"CompileFullText", func() (string, error) { return g.CompileFullText(bp, cmd) }},
		{"CompileDropIndex", func() (string, error) { return g.CompileDropIndex(bp, cmd) }},
		{"CompileDropUnique", func() (string, error) { return g.CompileDropUnique(bp, cmd) }},
		{"CompileDropFulltext", func() (string, error) { return g.CompileDropFulltext(bp, cmd) }},
		{"CompileDropPrimary", func() (string, error) { return g.CompileDropPrimary(bp, cmd) }},
		{"CompileRenameIndex", func() (string, error) { return g.CompileRenameIndex(bp, cmd) }},
		{"CompileDropForeign", func() (string, error) { return g.CompileDropForeign(bp, cmd) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.fn()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "not supported")
		})
	}
}

func TestBaseGrammar_CompileForeign(t *testing.T) {
	g := &blueprint.BaseGrammar{}
	bp := &blueprint.Blueprint{Name: "posts"}

	t.Run("incomplete command", func(t *testing.T) {
		cmd := &blueprint.Command{}
		_, err := g.CompileForeign(bp, cmd)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incomplete")
	})
}

func TestBaseGrammar_GetFluentCommands(t *testing.T) {
	g := &blueprint.BaseGrammar{}
	cmds := g.GetFluentCommands()
	assert.Empty(t, cmds)

	tableCmds := g.GetTableFluentCommands()
	assert.Empty(t, tableCmds)
}

func TestBaseGrammar_CreateIndexName(t *testing.T) {
	g := &blueprint.BaseGrammar{}
	bp := &blueprint.Blueprint{Name: "users"}

	idxName := g.CreateIndexName(bp, "index", "email")
	assert.Equal(t, "users_email_index", idxName)
}

func TestBaseGrammar_Helpers(t *testing.T) {
	g := &blueprint.BaseGrammar{}
	bp := &blueprint.Blueprint{Name: "users"}

	t.Run("CreateForeignKeyName", func(t *testing.T) {
		cmd := &blueprint.Command{Columns: []string{"user_id"}}
		name := g.CreateForeignKeyName(bp, cmd)
		assert.Equal(t, "users_user_id_foreign", name)
	})

	t.Run("QuoteString", func(t *testing.T) {
		assert.Equal(t, "'test'", g.QuoteString("test"))
	})

	t.Run("PrefixArray", func(t *testing.T) {
		assert.Equal(t, []string{"prefix_a", "prefix_b"}, g.PrefixArray("prefix_", []string{"a", "b"}))
	})

	t.Run("Columnize", func(t *testing.T) {
		assert.Equal(t, "a, b", g.Columnize([]string{"a", "b"}))
	})

	t.Run("Wrap", func(t *testing.T) {
		assert.Equal(t, "\"test\"", g.Wrap("test", "\""))
	})

	t.Run("WrapTable", func(t *testing.T) {
		assert.Equal(t, "\"test\"", g.WrapTable("test", "\""))
	})

	t.Run("WrapColumnize", func(t *testing.T) {
		assert.Equal(t, "\"a\", \"b\"", g.WrapColumnize([]string{"a", "b"}, "\""))
	})

	t.Run("WrapIndexName", func(t *testing.T) {
		assert.Equal(t, "\"idx\"", g.WrapIndexName("idx", "\""))
		assert.Empty(t, g.WrapIndexName("", "\""))
	})

	t.Run("GetValue", func(t *testing.T) {
		assert.Equal(t, "'val'", g.GetValue("val"))
		assert.Equal(t, "'1'", g.GetValue(1))
		assert.Equal(t, "'true'", g.GetValue(true))
		assert.Equal(t, "'false'", g.GetValue(false))
	})

	t.Run("GetDefaultValue", func(t *testing.T) {
		assert.Equal(t, "'val'", g.GetDefaultValue("val"))
		assert.Equal(t, "NULL", g.GetDefaultValue(nil))
	})
}

func TestBaseGrammar_CompileForeignValid(t *testing.T) {
	g := &blueprint.BaseGrammar{}
	bp := &blueprint.Blueprint{Name: "posts"}

	cmd := &blueprint.Command{
		Columns:    []string{"user_id"},
		On:         "users",
		References: []string{"id"},
		OnDelete:   "CASCADE",
		OnUpdate:   "RESTRICT",
	}

	sql, err := g.CompileForeign(bp, cmd)
	require.NoError(t, err)
	assert.Equal(
		t,
		"ALTER TABLE posts ADD CONSTRAINT posts_user_id_foreign FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE ON UPDATE RESTRICT",
		sql,
	)
}
