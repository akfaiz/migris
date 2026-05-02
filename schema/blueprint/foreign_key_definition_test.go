package blueprint_test

import (
	"testing"

	"github.com/akfaiz/migris/schema/blueprint"
	"github.com/stretchr/testify/assert"
)

func TestForeignKeyDefinition_Modifiers(t *testing.T) {
	bp := blueprint.NewBlueprintForTesting("posts", &mockGrammar{})
	fk := bp.Foreign("user_id")

	fk.References("id").
		On("users").
		OnDelete("CASCADE").
		OnUpdate("RESTRICT").
		Name("fk_posts_user_id").
		Deferrable(true).
		InitiallyImmediate(true)

	// Since fk is an interface wrapping internal command, we check the bp.Commands
	cmd := bp.Commands[0]
	assert.Equal(t, blueprint.CommandForeign, cmd.Name)
	assert.Equal(t, []string{"user_id"}, cmd.Columns)
	assert.Equal(t, []string{"id"}, cmd.References)
	assert.Equal(t, "users", cmd.On)
	assert.Equal(t, "CASCADE", cmd.OnDelete)
	assert.Equal(t, "RESTRICT", cmd.OnUpdate)
	assert.Equal(t, "fk_posts_user_id", cmd.Index)
	assert.True(t, *cmd.Deferrable)
	assert.True(t, *cmd.InitiallyImmediate)
}

func TestForeignKeyDefinition_HelperMethods(t *testing.T) {
	bp := blueprint.NewBlueprintForTesting("posts", &mockGrammar{})

	bp.Foreign("c1").CascadeOnDelete().CascadeOnUpdate()
	bp.Foreign("c2").RestrictOnDelete().RestrictOnUpdate()
	bp.Foreign("c3").NullOnDelete().NullOnUpdate()
	bp.Foreign("c4").NoActionOnDelete().NoActionOnUpdate()

	assert.Equal(t, "CASCADE", bp.Commands[0].OnDelete)
	assert.Equal(t, "CASCADE", bp.Commands[0].OnUpdate)

	assert.Equal(t, "RESTRICT", bp.Commands[1].OnDelete)
	assert.Equal(t, "RESTRICT", bp.Commands[1].OnUpdate)

	assert.Equal(t, "SET NULL", bp.Commands[2].OnDelete)
	assert.Equal(t, "SET NULL", bp.Commands[2].OnUpdate)

	assert.Equal(t, "NO ACTION", bp.Commands[3].OnDelete)
	assert.Equal(t, "NO ACTION", bp.Commands[3].OnUpdate)
}
