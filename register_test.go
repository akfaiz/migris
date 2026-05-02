package migris_test

import (
	"testing"

	"github.com/akfaiz/migris"
	"github.com/akfaiz/migris/schema"
	"github.com/stretchr/testify/require"
)

func TestAddMigrationContext(t *testing.T) {
	migris.ResetRegisteredMigrations()

	require.NotPanics(t, func() {
		migris.AddMigrationContext(
			func(ctx schema.Context) error { return nil },
			func(ctx schema.Context) error { return nil },
		)
	})

	require.Panics(t, func() {
		migris.AddMigrationContext(
			func(ctx schema.Context) error { return nil },
			func(ctx schema.Context) error { return nil },
		)
	})
}

func TestRegistry_AddMigrationContext(t *testing.T) {
	r := migris.NewRegistry()

	require.NotPanics(t, func() {
		r.AddMigrationContext(
			func(ctx schema.Context) error { return nil },
			func(ctx schema.Context) error { return nil },
		)
	})

	require.Panics(t, func() {
		r.AddMigrationContext(
			func(ctx schema.Context) error { return nil },
			func(ctx schema.Context) error { return nil },
		)
	})
}
