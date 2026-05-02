package blueprint_test

import (
	"testing"

	"github.com/akfaiz/migris/schema/blueprint"
	"github.com/stretchr/testify/assert"
)

func TestIndexDefinition_Modifiers(t *testing.T) {
	bp := blueprint.NewBlueprintForTesting("users", &mockGrammar{})
	idx := bp.Index("email")

	idx.Name("idx_email_custom").
		Algorithm("hash").
		Deferrable(true).
		InitiallyImmediate(false)

	cmd := bp.Commands[0]
	assert.Equal(t, blueprint.CommandIndex, cmd.Name)
	assert.Equal(t, "idx_email_custom", cmd.Index)
	assert.Equal(t, "hash", cmd.Algorithm)
	assert.True(t, *cmd.Deferrable)
	assert.False(t, *cmd.InitiallyImmediate)
}

func TestIndexDefinition_Language(t *testing.T) {
	bp := blueprint.NewBlueprintForTesting("users", &mockGrammar{})
	idx := bp.FullText("bio")

	idx.Language("spanish")

	cmd := bp.Commands[0]
	assert.Equal(t, blueprint.CommandFullText, cmd.Name)
	assert.Equal(t, "spanish", cmd.Language)
}
