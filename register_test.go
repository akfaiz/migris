package migris_test

import (
	"testing"

	"github.com/akfaiz/migris"
	"github.com/akfaiz/migris/schema"
	"github.com/stretchr/testify/require"
)

func TestAddMigrationContext(t *testing.T) {
	migris.ResetRegisteredMigrations()
	noop := func(_ schema.Context) error { return nil }

	require.NotPanics(t, func() {
		migris.AddNamedMigrationContext("20250101000001_init.go", noop, noop)
	})
	require.Panics(t, func() {
		migris.AddNamedMigrationContext("20250101000001_init.go", noop, noop)
	})
	require.Panics(t, func() {
		migris.AddNamedMigrationContext("not_versioned.go", noop, noop)
	})
}

func TestRegistry_AddMigrationContext(t *testing.T) {
	r := migris.NewRegistry()
	noop := func(_ schema.Context) error { return nil }

	require.NoError(t, r.AddNamedMigrationContext("20250101000001_init.go", noop, noop))
	require.Error(t, r.AddNamedMigrationContext("20250101000001_init.go", noop, noop))
	require.Error(t, r.AddNamedMigrationContext("not_versioned.go", noop, noop))
}
