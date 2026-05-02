package blueprint_test

import (
	"testing"

	"github.com/akfaiz/migris/schema/blueprint"
	"github.com/stretchr/testify/assert"
)

func TestBaseGrammar_Helpers(t *testing.T) {
	g := &blueprint.BaseGrammar{}
	bp := blueprint.NewBlueprintForTesting("users", nil)

	t.Run("CreateIndexName", func(t *testing.T) {
		name := g.CreateIndexName(bp, "unique", "email", "phone")
		assert.Equal(t, "users_email_phone_unique", name)
	})

	t.Run("CreateForeignKeyName", func(t *testing.T) {
		name := g.CreateForeignKeyName(bp, &blueprint.Command{Columns: []string{"user_id"}})
		assert.Equal(t, "users_user_id_foreign", name)
	})

	t.Run("QuoteString", func(t *testing.T) {
		assert.Equal(t, "'test'", g.QuoteString("test"))
	})

	t.Run("PrefixArray", func(t *testing.T) {
		assert.Equal(t, []string{"p1", "p2"}, g.PrefixArray("p", []string{"1", "2"}))
	})

	t.Run("Columnize", func(t *testing.T) {
		assert.Equal(t, "c1, c2", g.Columnize([]string{"c1", "c2"}))
		assert.Empty(t, g.Columnize([]string{}))
	})

	t.Run("Wrap", func(t *testing.T) {
		assert.Equal(t, "\"users\"", g.Wrap("users", "\""))
		assert.Equal(t, "\"schema\".\"table\"", g.Wrap("schema.table", "\""))
		assert.Equal(t, "*", g.Wrap("*", "\""))
		assert.Empty(t, g.Wrap("", "\""))
	})

	t.Run("WrapColumnize", func(t *testing.T) {
		assert.Equal(t, "\"c1\", \"c2\"", g.WrapColumnize([]string{"c1", "c2"}, "\""))
	})

	t.Run("WrapIndexName", func(t *testing.T) {
		assert.Equal(t, "\"idx\"", g.WrapIndexName("idx", "\""))
		assert.Equal(t, "idx(col)", g.WrapIndexName("idx(col)", "\""))
	})

	t.Run("GetValue", func(t *testing.T) {
		assert.Equal(t, "'123'", g.GetValue(123))
		assert.Equal(t, "CURRENT_TIMESTAMP", g.GetValue(blueprint.Expression("CURRENT_TIMESTAMP")))
	})

	t.Run("GetDefaultValue", func(t *testing.T) {
		assert.Equal(t, "NULL", g.GetDefaultValue(nil))
		assert.Equal(t, "'1'", g.GetDefaultValue(true))
		assert.Equal(t, "'0'", g.GetDefaultValue(false))
		assert.Equal(t, "'val'", g.GetDefaultValue("val"))
	})
}
