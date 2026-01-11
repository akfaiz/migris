package migris

import (
	"context"
	"os"
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
	assert.True(t, migrator.dryRunConfig.PrintSQL)
	assert.True(t, migrator.dryRunConfig.PrintMigrations)
}

func TestDryRunContext(t *testing.T) {
	config := schema.DryRunConfig{
		PrintSQL:        true,
		PrintMigrations: true,
		ColorOutput:     false,
		OutputWriter:    os.Stdout,
	}

	ctx := context.Background()
	dryRunCtx := schema.NewDryRunContext(ctx, config)

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
	var regularCtx schema.Context = schema.NewContext(ctx, nil)
	assert.NotNil(t, regularCtx)

	// Test that DryRunContext also implements Context interface
	config := schema.DryRunConfig{OutputWriter: os.Stdout}
	var dryRunCtx schema.Context = schema.NewDryRunContext(ctx, config)
	assert.NotNil(t, dryRunCtx)
}
