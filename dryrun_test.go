package migris

import (
	"context"
	"testing"

	"github.com/akfaiz/migris/schema"
	"github.com/stretchr/testify/assert"
)

func TestDryRunMode(t *testing.T) {
	// Test creating a migrator with dry-run mode enabled
	migrator, err := New("postgres",
		WithDryRun(true),
	)
	assert.NoError(t, err)
	assert.True(t, migrator.dryRun)
}

func TestDryRunContext(t *testing.T) {
	ctx := context.Background()
	dryRunCtx := schema.NewDryRunContext(ctx)

	// Test that SQL is captured instead of executed
	result, err := dryRunCtx.Exec("CREATE TABLE test (id SERIAL PRIMARY KEY)")
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify SQL was captured
	capturedSQL := dryRunCtx.GetCapturedSQL()
	assert.Len(t, capturedSQL, 1)
	assert.Contains(t, capturedSQL[0], "CREATE TABLE test")
}

func TestRegularContextInterface(t *testing.T) {
	// Test that RegularContext implements Context interface
	ctx := context.Background()

	// This would normally require a real DB, but we're just testing the interface
	regularCtx := schema.NewContext(ctx, nil)
	assert.NotNil(t, regularCtx)

	// Test that DryRunContext also implements Context interface
	dryRunCtx := schema.NewDryRunContext(ctx)
	assert.NotNil(t, dryRunCtx)
}
