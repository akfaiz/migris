package migris

import (
	"context"
	"database/sql"
	"fmt"
	"runtime"
	"slices"

	"github.com/pressly/goose/v3"
)

var (
	registeredGoMigrations = make(map[int64]*goose.Migration)
)

// GoMigrationContext is a Go migration func that is run within a transaction and receives a
// context.
type GoMigrationContext func(ctx context.Context, tx *sql.Tx) error

// AddMigrationContext adds Go migrations.
func AddMigrationContext(up, down GoMigrationContext) {
	_, filename, _, _ := runtime.Caller(1)
	AddNamedMigrationContext(filename, up, down)
}

// AddNamedMigrationContext adds named Go migrations.
func AddNamedMigrationContext(filename string, up, down GoMigrationContext) {
	if err := register(
		filename,
		true,
		&goose.GoFunc{RunTx: up, Mode: goose.TransactionEnabled},
		&goose.GoFunc{RunTx: down, Mode: goose.TransactionEnabled},
	); err != nil {
		panic(err)
	}
}

func register(filename string, useTx bool, up, down *goose.GoFunc) error {
	v, _ := goose.NumericComponent(filename)
	if existing, ok := registeredGoMigrations[v]; ok {
		return fmt.Errorf("failed to add migration %q: version %d conflicts with %q",
			filename,
			v,
			existing.Source,
		)
	}
	// Add to global as a registered migration.
	m := goose.NewGoMigration(v, up, down)
	m.Source = filename
	// We explicitly set transaction to maintain existing behavior. Both up and down may be nil, but
	// we know based on the register function what the user is requesting.
	m.UseTx = useTx
	registeredGoMigrations[v] = m
	return nil
}

func getMigrations() []*goose.Migration {
	type migrationWithVersion struct {
		version   int64
		migration *goose.Migration
	}
	var migrations []*migrationWithVersion
	for _, m := range registeredGoMigrations {
		migrations = append(migrations, &migrationWithVersion{
			version:   m.Version,
			migration: m,
		})
	}
	slices.SortFunc(migrations, func(a, b *migrationWithVersion) int {
		return int(a.version - b.version)
	})
	var results []*goose.Migration
	for _, m := range migrations {
		results = append(results, m.migration)
	}
	return results
}
