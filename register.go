package migris

import (
	"context"
	"database/sql"
	"fmt"
	"path"
	"runtime"

	"github.com/akfaiz/migris/schema"
	"github.com/pressly/goose/v3"
)

var (
	registeredVersions   = make(map[int64]string)
	registeredMigrations = make([]*Migration, 0)
)

type Migration struct {
	version                    int64
	source                     string
	upFnContext, downFnContext MigrationContext
}

// MigrationContext is a Go migration func that is run within a transaction and receives a
// context.
type MigrationContext func(ctx schema.Context) error

func (m MigrationContext) runTxFunc(source string) func(ctx context.Context, tx *sql.Tx) error {
	return func(ctx context.Context, tx *sql.Tx) error {
		filename := path.Base(source)

		// Check if we're in dry-run mode
		isDryRun := getGlobalDryRunState()

		var c schema.Context
		if isDryRun {
			// Create dry-run context
			c = schema.NewDryRunContext(ctx)
		} else {
			// Create regular context
			c = schema.NewContext(ctx, tx, schema.WithFilename(filename))
		}

		return m(c)
	}
}

// AddMigrationContext adds Go migrations.
func AddMigrationContext(up, down MigrationContext) {
	_, filename, _, _ := runtime.Caller(1)
	AddNamedMigrationContext(filename, up, down)
}

// AddNamedMigrationContext adds named Go migrations.
func AddNamedMigrationContext(source string, up, down MigrationContext) {
	if err := register(
		source,
		up,
		down,
	); err != nil {
		panic(err)
	}
}

func register(source string, up, down MigrationContext) error {
	v, _ := goose.NumericComponent(source)
	if existing, ok := registeredVersions[v]; ok {
		return fmt.Errorf("failed to add migration %q: version %d conflicts with %q",
			source,
			v,
			existing,
		)
	}
	// Add to global as a registered migration.
	m := &Migration{
		version:       v,
		source:        source,
		upFnContext:   up,
		downFnContext: down,
	}
	registeredVersions[v] = source
	registeredMigrations = append(registeredMigrations, m)
	return nil
}

func gooseMigrations() []*goose.Migration {
	migrations := make([]*goose.Migration, 0, len(registeredMigrations))
	for _, m := range registeredMigrations {
		upFunc := &goose.GoFunc{
			RunTx: m.upFnContext.runTxFunc(m.source),
			Mode:  goose.TransactionEnabled,
		}
		downFunc := &goose.GoFunc{
			RunTx: m.downFnContext.runTxFunc(m.source),
			Mode:  goose.TransactionEnabled,
		}
		gm := goose.NewGoMigration(m.version, upFunc, downFunc)
		migrations = append(migrations, gm)
	}
	return migrations
}
