package blueprint_test

import (
	"testing"

	"github.com/akfaiz/migris/schema/blueprint"
	"github.com/stretchr/testify/assert"
)

func TestColumnDefinition_Modifiers(t *testing.T) {
	col := &blueprint.Column{Name: "test"}

	col.AutoIncrement().
		Charset("utf8").
		Collation("utf8_bin").
		Comment("some comment").
		Default("val").
		Nullable(true).
		Unsigned().
		UseCurrent().
		UseCurrentOnUpdate().
		After("other").
		First().
		VirtualAs("expr1").
		StoredAs("expr2").
		Invisible().
		Always(true).
		RenameTo("new_name")

	assert.True(t, *col.AutoIncrementVal)
	assert.Equal(t, "utf8", *col.CharsetVal)
	assert.Equal(t, "utf8_bin", *col.CollationVal)
	assert.Equal(t, "some comment", *col.CommentVal)
	assert.Equal(t, "val", col.DefaultValue)
	assert.True(t, *col.NullableVal)
	assert.True(t, *col.UnsignedVal)
	assert.True(t, col.UseCurrentVal)
	assert.True(t, col.UseCurrentOnUpdateVal)
	assert.Equal(t, "other", *col.AfterVal)
	assert.True(t, col.FirstVal)
	assert.Equal(t, "expr1", *col.VirtualAsVal)
	assert.Equal(t, "expr2", *col.StoredAsVal)
	assert.True(t, *col.InvisibleVal)
	assert.True(t, *col.AlwaysVal)
	assert.Equal(t, "new_name", col.RenameToVal)
}

func TestColumnDefinition_IndexUnique(t *testing.T) {
	t.Run("index with boolean", func(t *testing.T) {
		col := &blueprint.Column{Name: "test"}
		col.Index(true)
		assert.True(t, *col.IndexVal)
	})

	t.Run("index with name", func(t *testing.T) {
		col := &blueprint.Column{Name: "test"}
		col.Index("idx_test")
		assert.True(t, *col.IndexVal)
		assert.Equal(t, "idx_test", col.IndexName)
	})

	t.Run("unique with boolean", func(t *testing.T) {
		col := &blueprint.Column{Name: "test"}
		col.Unique(true)
		assert.True(t, *col.UniqueVal)
	})

	t.Run("unique with name", func(t *testing.T) {
		col := &blueprint.Column{Name: "test"}
		col.Unique("uk_test")
		assert.True(t, *col.UniqueVal)
		assert.Equal(t, "uk_test", col.UniqueName)
	})
}
