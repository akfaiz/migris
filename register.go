package migris

import (
	"context"
	"database/sql"
	"fmt"
	"path"
	"runtime"
	"sort"
	"sync"

	"github.com/akfaiz/migris/internal/dialect"
	"github.com/akfaiz/migris/schema"
	"github.com/pressly/goose/v3"
)

type Registry struct {
	mu         sync.RWMutex
	versions   map[int64]string
	migrations []*Migration
}

var defaultRegistry = NewRegistry()

type Migration struct {
	version                    int64
	source                     string
	upFnContext, downFnContext MigrationContext
}

// MigrationContext is a Go migration func that is run within a transaction and receives a
// context.
type MigrationContext func(ctx schema.Context) error

// NewRegistry creates an isolated migration registry.
func NewRegistry() *Registry {
	return &Registry{
		versions:   make(map[int64]string),
		migrations: make([]*Migration, 0),
	}
}

func (m MigrationContext) runTxFunc(source string, dialectVal dialect.Dialect) func(ctx context.Context, tx *sql.Tx) error {
	return func(ctx context.Context, tx *sql.Tx) error {
		filename := path.Base(source)
		c := schema.NewContext(ctx, tx,
			schema.WithFilename(filename),
			schema.WithDialect(dialectVal.String()),
		)

		return m(c)
	}
}

// AddMigrationContext adds Go migrations to the default registry.
func AddMigrationContext(up, down MigrationContext) {
	_, filename, _, _ := runtime.Caller(1)
	AddNamedMigrationContext(filename, up, down)
}

// AddNamedMigrationContext adds named Go migrations to the default registry.
func AddNamedMigrationContext(source string, up, down MigrationContext) {
	if err := defaultRegistry.AddNamedMigrationContext(source, up, down); err != nil {
		panic(err)
	}
}

// AddMigrationContext adds Go migrations to the registry.
func (r *Registry) AddMigrationContext(up, down MigrationContext) {
	_, filename, _, _ := runtime.Caller(1)
	if err := r.AddNamedMigrationContext(filename, up, down); err != nil {
		panic(err)
	}
}

// AddNamedMigrationContext adds named Go migrations to the registry.
func (r *Registry) AddNamedMigrationContext(source string, up, down MigrationContext) error {
	v, _ := goose.NumericComponent(source)

	r.mu.Lock()
	defer r.mu.Unlock()

	if existing, ok := r.versions[v]; ok {
		return fmt.Errorf("failed to add migration %q: version %d conflicts with %q",
			source,
			v,
			existing,
		)
	}
	m := &Migration{
		version:       v,
		source:        source,
		upFnContext:   up,
		downFnContext: down,
	}
	r.versions[v] = source
	r.migrations = append(r.migrations, m)
	return nil
}

func (r *Registry) migrationsSnapshot() []*Migration {
	r.mu.RLock()
	defer r.mu.RUnlock()

	migrations := make([]*Migration, len(r.migrations))
	copy(migrations, r.migrations)
	sort.SliceStable(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})
	return migrations
}

// ResetRegisteredMigrations clears the default global registry of migrations.
// This is primarily intended for tests to avoid cross-test interference.
func ResetRegisteredMigrations() {
	defaultRegistry = NewRegistry()
}

func (r *Registry) gooseMigrations(dialectVal dialect.Dialect) []*goose.Migration {
	registeredMigrations := r.migrationsSnapshot()
	migrations := make([]*goose.Migration, 0, len(registeredMigrations))
	for _, m := range registeredMigrations {
		upFunc := &goose.GoFunc{
			RunTx: m.upFnContext.runTxFunc(m.source, dialectVal),
			Mode:  goose.TransactionEnabled,
		}
		downFunc := &goose.GoFunc{
			RunTx: m.downFnContext.runTxFunc(m.source, dialectVal),
			Mode:  goose.TransactionEnabled,
		}
		gm := goose.NewGoMigration(m.version, upFunc, downFunc)
		migrations = append(migrations, gm)
	}
	return migrations
}
