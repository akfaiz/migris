package schema //nolint:testpackage // Need to access unexported members for testing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSqliteGrammar_CompileCreate(t *testing.T) {
	g := newSqliteGrammar()

	tests := []struct {
		name      string
		table     string
		blueprint func(table *Blueprint)
		want      string
		wantErr   bool
	}{
		{
			name:  "basic table creation",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.ID()
				table.String("name", 255)
			},
			want:    "CREATE TABLE \"users\" (\"id\" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, \"name\" TEXT NOT NULL)",
			wantErr: false,
		},
		{
			name:  "table with nullable column",
			table: "posts",
			blueprint: func(table *Blueprint) {
				table.ID()
				table.String("title", 255)
				table.Text("content").Nullable()
			},
			want:    "CREATE TABLE \"posts\" (\"id\" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, \"title\" TEXT NOT NULL, \"content\" TEXT)",
			wantErr: false,
		},
		{
			name:  "table with default values",
			table: "settings",
			blueprint: func(table *Blueprint) {
				table.ID()
				table.String("key", 255)
				table.Boolean("enabled").Default(true)
				table.Integer("count").Default(0)
			},
			want:    "CREATE TABLE \"settings\" (\"id\" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, \"key\" TEXT NOT NULL, \"enabled\" INTEGER NOT NULL DEFAULT 1, \"count\" INTEGER NOT NULL DEFAULT 0)",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: tt.table, grammar: g}
			tt.blueprint(bp)

			got, err := g.CompileCreate(bp)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestSqliteGrammar_CompileAdd(t *testing.T) {
	g := newSqliteGrammar()

	tests := []struct {
		name      string
		table     string
		blueprint func(table *Blueprint)
		want      string
		wantErr   bool
	}{
		{
			name:  "add single column",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.String("email", 255)
			},
			want:    "ALTER TABLE \"users\" ADD COLUMN \"email\" TEXT NOT NULL",
			wantErr: false,
		},
		{
			name:  "add multiple columns",
			table: "posts",
			blueprint: func(table *Blueprint) {
				table.String("slug", 255)
				table.Timestamp("published_at").Nullable()
			},
			want:    "ALTER TABLE \"posts\" ADD COLUMN \"slug\" TEXT NOT NULL, ADD COLUMN \"published_at\" DATETIME",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: tt.table, grammar: g}
			tt.blueprint(bp)

			got, err := g.CompileAdd(bp)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestSqliteGrammar_CompileDrop(t *testing.T) {
	g := newSqliteGrammar()

	tests := []struct {
		name    string
		table   string
		want    string
		wantErr bool
	}{
		{
			name:    "drop table",
			table:   "users",
			want:    "DROP TABLE \"users\"",
			wantErr: false,
		},
		{
			name:    "drop table with special chars",
			table:   "user_profiles",
			want:    "DROP TABLE \"user_profiles\"",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: tt.table, grammar: g}

			got, err := g.CompileDrop(bp)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestSqliteGrammar_CompileDropIfExists(t *testing.T) {
	g := newSqliteGrammar()

	tests := []struct {
		name    string
		table   string
		want    string
		wantErr bool
	}{
		{
			name:    "drop table if exists",
			table:   "users",
			want:    "DROP TABLE IF EXISTS \"users\"",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: tt.table, grammar: g}

			got, err := g.CompileDropIfExists(bp)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestSqliteGrammar_CompileRename(t *testing.T) {
	g := newSqliteGrammar()

	tests := []struct {
		name    string
		table   string
		newName string
		want    string
		wantErr bool
	}{
		{
			name:    "rename table",
			table:   "users",
			newName: "customers",
			want:    "ALTER TABLE \"users\" RENAME TO \"customers\"",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: tt.table, grammar: g}
			bp.rename(tt.newName)

			got, err := g.CompileRename(bp, bp.commands[0])
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestSqliteGrammar_CompileIndex(t *testing.T) {
	g := newSqliteGrammar()

	tests := []struct {
		name      string
		table     string
		blueprint func(table *Blueprint)
		want      string
		wantErr   bool
	}{
		{
			name:  "create index",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Index("email")
			},
			want:    "CREATE INDEX \"idx_users_email\" ON \"users\" (\"email\")",
			wantErr: false,
		},
		{
			name:  "create composite index",
			table: "posts",
			blueprint: func(table *Blueprint) {
				table.Index("user_id", "created_at")
			},
			want:    "CREATE INDEX \"idx_posts_user_id_created_at\" ON \"posts\" (\"user_id\", \"created_at\")",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: tt.table, grammar: g}
			tt.blueprint(bp)

			got, err := g.CompileIndex(bp, bp.commands[0])
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestSqliteGrammar_CompileUnique(t *testing.T) {
	g := newSqliteGrammar()

	tests := []struct {
		name      string
		table     string
		blueprint func(table *Blueprint)
		want      string
		wantErr   bool
	}{
		{
			name:  "create unique index",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Unique("email")
			},
			want:    "CREATE UNIQUE INDEX \"uq_users_email\" ON \"users\" (\"email\")",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: tt.table, grammar: g}
			tt.blueprint(bp)

			got, err := g.CompileUnique(bp, bp.commands[0])
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestSqliteGrammar_TypeMethods(t *testing.T) {
	g := newSqliteGrammar()

	tests := []struct {
		name string
		fn   func(*columnDefinition) string
		want string
	}{
		{"typeChar", g.typeChar, "TEXT"},
		{"typeString", g.typeString, "TEXT"},
		{"typeText", g.typeText, "TEXT"},
		{"typeInteger", g.typeInteger, "INTEGER"},
		{"typeBigInteger", g.typeBigInteger, "INTEGER"},
		{"typeFloat", g.typeFloat, "REAL"},
		{"typeDouble", g.typeDouble, "REAL"},
		{"typeDecimal", g.typeDecimal, "NUMERIC"},
		{"typeBoolean", g.typeBoolean, "INTEGER"},
		{"typeDate", g.typeDate, "DATE"},
		{"typeDateTime", g.typeDateTime, "DATETIME"},
		{"typeTimestamp", g.typeTimestamp, "DATETIME"},
		{"typeBinary", g.typeBinary, "BLOB"},
		{"typeUUID", g.typeUUID, "TEXT"},
		{"typeJSON", g.typeJSON, "TEXT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := &columnDefinition{} // Empty column definition for testing
			got := tt.fn(col)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSqliteGrammar_UnsupportedOperations(t *testing.T) {
	g := newSqliteGrammar()
	bp := &Blueprint{name: "test_table", grammar: g}
	cmd := &command{} // Empty command for testing

	tests := []struct {
		name string
		fn   func() (string, error)
	}{
		{"CompileChange", func() (string, error) { return g.CompileChange(bp, cmd) }},
		{"CompileDropColumn", func() (string, error) { return g.CompileDropColumn(bp, cmd) }},
		{"CompileRenameColumn", func() (string, error) { return g.CompileRenameColumn(bp, cmd) }},
		{"CompileFullText", func() (string, error) { return g.CompileFullText(bp, cmd) }},
		{"CompileDropPrimary", func() (string, error) { return g.CompileDropPrimary(bp, cmd) }},
		{"CompileRenameIndex", func() (string, error) { return g.CompileRenameIndex(bp, cmd) }},
		{"CompileDropForeign", func() (string, error) { return g.CompileDropForeign(bp, cmd) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, err := tt.fn()
			require.Error(t, err)
			assert.Empty(t, sql)
		})
	}
}

func TestSqliteGrammar_SupportedButDelegated(t *testing.T) {
	g := newSqliteGrammar()
	bp := &Blueprint{name: "test_table", grammar: g}
	cmd := &command{} // Empty command for testing

	// Test operations that are supported but handled at table creation time (not as separate commands)
	tests := []struct {
		name string
		fn   func() (string, error)
	}{
		{"CompilePrimary", func() (string, error) { return g.CompilePrimary(bp, cmd) }},
		{"CompileForeign", func() (string, error) { return g.CompileForeign(bp, cmd) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, err := tt.fn()
			require.NoError(t, err, "Should not return error as these are handled at table creation time")
			assert.Empty(t, sql, "Should return empty SQL as these are handled elsewhere")
		})
	}
}
