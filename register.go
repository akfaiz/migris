package migris

import (
	"context"
	"database/sql"
	"fmt"
	"path"
	"runtime"

	"github.com/afkdevs/migris/schema"
	"github.com/pressly/goose/v3"
)

var (
	registeredVersions   = make(map[int64]string)
	registeredMigrations = make([]*goose.Migration, 0)
)

// GoMigrationContext is a Go migration func that is run within a transaction and receives a
// context.
type GoMigrationContext func(ctx *schema.Context) error

func (m GoMigrationContext) RunTxFunc(filename string) func(ctx context.Context, tx *sql.Tx) error {
	return func(ctx context.Context, tx *sql.Tx) error {
		filename = path.Base(filename)
		c := schema.NewContext(ctx, tx, schema.WithFilename(filename))
		return m(c)
	}
}

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
		&goose.GoFunc{RunTx: up.RunTxFunc(filename), Mode: goose.TransactionEnabled},
		&goose.GoFunc{RunTx: down.RunTxFunc(filename), Mode: goose.TransactionEnabled},
	); err != nil {
		panic(err)
	}
}

func register(filename string, useTx bool, up, down *goose.GoFunc) error {
	v, _ := goose.NumericComponent(filename)
	if existing, ok := registeredVersions[v]; ok {
		return fmt.Errorf("failed to add migration %q: version %d conflicts with %q",
			filename,
			v,
			existing,
		)
	}
	// Add to global as a registered migration.
	m := goose.NewGoMigration(v, up, down)
	m.Source = filename
	// We explicitly set transaction to maintain existing behavior. Both up and down may be nil, but
	// we know based on the register function what the user is requesting.
	m.UseTx = useTx
	registeredVersions[v] = filename
	registeredMigrations = append(registeredMigrations, m)
	return nil
}
