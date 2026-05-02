package grammars_test

import (
	"testing"

	"github.com/akfaiz/migris/schema/blueprint"
	"github.com/akfaiz/migris/schema/grammars"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSqliteGrammar_CompileCreate(t *testing.T) {
	g, err := grammars.NewGrammar("sqlite3")
	require.NoError(t, err)

	tests := []struct {
		name      string
		table     string
		blueprint func(table *blueprint.Blueprint)
		want      string
		wantErr   bool
	}{
		{
			name:  "basic table creation",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.ID()
				table.String("name", 255)
			},
			want:    "CREATE TABLE \"users\" (\"id\" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, \"name\" TEXT NOT NULL)",
			wantErr: false,
		},
		{
			name:  "table with nullable column",
			table: "posts",
			blueprint: func(table *blueprint.Blueprint) {
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
			blueprint: func(table *blueprint.Blueprint) {
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
			tableName := tt.table
			if tableName == "" && !tt.wantErr {
				tableName = tt.name
			}
			bp := &blueprint.Blueprint{Name: tableName, Grammar: g}
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
	g, err := grammars.NewGrammar("sqlite3")
	require.NoError(t, err)

	tests := []struct {
		name      string
		table     string
		blueprint func(table *blueprint.Blueprint)
		want      string
		wantErr   bool
	}{
		{
			name:  "add single column",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.String("email", 255)
			},
			want:    "ALTER TABLE \"users\" ADD COLUMN \"email\" TEXT NOT NULL",
			wantErr: false,
		},
		{
			name:  "add multiple columns",
			table: "posts",
			blueprint: func(table *blueprint.Blueprint) {
				table.String("slug", 255)
				table.Timestamp("published_at").Nullable()
			},
			want:    "ALTER TABLE \"posts\" ADD COLUMN \"slug\" TEXT NOT NULL, ADD COLUMN \"published_at\" DATETIME",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tableName := tt.table
			if tableName == "" && !tt.wantErr {
				tableName = tt.name
			}
			bp := &blueprint.Blueprint{Name: tableName, Grammar: g}
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
	g, err := grammars.NewGrammar("sqlite3")
	require.NoError(t, err)

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
			tableName := tt.table
			if tableName == "" && !tt.wantErr {
				tableName = tt.name
			}
			bp := &blueprint.Blueprint{Name: tableName, Grammar: g}

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
	g, err := grammars.NewGrammar("sqlite3")
	require.NoError(t, err)

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
			tableName := tt.table
			if tableName == "" && !tt.wantErr {
				tableName = tt.name
			}
			bp := &blueprint.Blueprint{Name: tableName, Grammar: g}

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
	g, err := grammars.NewGrammar("sqlite3")
	require.NoError(t, err)

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
			tableName := tt.table
			if tableName == "" && !tt.wantErr {
				tableName = tt.name
			}
			bp := &blueprint.Blueprint{Name: tableName, Grammar: g}
			bp.Rename(tt.newName)

			got, err := g.CompileRename(bp, bp.Commands[0])
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
	g, err := grammars.NewGrammar("sqlite3")
	require.NoError(t, err)

	tests := []struct {
		name      string
		table     string
		blueprint func(table *blueprint.Blueprint)
		want      string
		wantErr   bool
	}{
		{
			name:  "create index",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Index("email")
			},
			want:    "CREATE INDEX \"users_email_index\" ON \"users\" (\"email\")",
			wantErr: false,
		},
		{
			name:  "create composite index",
			table: "posts",
			blueprint: func(table *blueprint.Blueprint) {
				table.Index("user_id", "created_at")
			},
			want:    "CREATE INDEX \"posts_user_id_created_at_index\" ON \"posts\" (\"user_id\", \"created_at\")",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tableName := tt.table
			if tableName == "" && !tt.wantErr {
				tableName = tt.name
			}
			bp := &blueprint.Blueprint{Name: tableName, Grammar: g}
			tt.blueprint(bp)

			got, err := g.CompileIndex(bp, bp.Commands[0])
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
	g, err := grammars.NewGrammar("sqlite3")
	require.NoError(t, err)

	tests := []struct {
		name      string
		table     string
		blueprint func(table *blueprint.Blueprint)
		want      string
		wantErr   bool
	}{
		{
			name:  "create unique index",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Unique("email")
			},
			want:    "CREATE UNIQUE INDEX \"users_email_unique\" ON \"users\" (\"email\")",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tableName := tt.table
			if tableName == "" && !tt.wantErr {
				tableName = tt.name
			}
			bp := &blueprint.Blueprint{Name: tableName, Grammar: g}
			tt.blueprint(bp)

			got, err := g.CompileUnique(bp, bp.Commands[0])
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestSqliteGrammar_UnsupportedOperations(t *testing.T) {
	g, err := grammars.NewGrammar("sqlite3")
	require.NoError(t, err)
	bp := &blueprint.Blueprint{Name: "test_table", Grammar: g}
	cmd := &blueprint.Command{} // Empty command for testing

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
	g, err := grammars.NewGrammar("sqlite3")
	require.NoError(t, err)
	bp := &blueprint.Blueprint{Name: "test_table", Grammar: g}
	cmd := &blueprint.Command{} // Empty command for testing

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

func TestSqliteGrammar_GetType(t *testing.T) {
	g, err := grammars.NewGrammar("sqlite3")
	require.NoError(t, err)

	tests := []struct {
		name      string
		blueprint func(table *blueprint.Blueprint)
		want      string
	}{
		{
			name: "custom column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Column("name", "CUSTOM_TYPE")
			},
			want: "CUSTOM_TYPE",
		},
		{
			name: "char column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Char("code", 10)
			},
			want: "TEXT",
		},
		{
			name: "string column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.String("name", 255)
			},
			want: "TEXT",
		},
		{
			name: "tiny text column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.TinyText("content")
			},
			want: "TEXT",
		},
		{
			name: "text column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Text("content")
			},
			want: "TEXT",
		},
		{
			name: "medium text column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.MediumText("content")
			},
			want: "TEXT",
		},
		{
			name: "long text column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.LongText("content")
			},
			want: "TEXT",
		},
		{
			name: "big integer column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.BigInteger("id")
			},
			want: "INTEGER",
		},
		{
			name: "integer column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Integer("id")
			},
			want: "INTEGER",
		},
		{
			name: "medium integer column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.MediumInteger("id")
			},
			want: "INTEGER",
		},
		{
			name: "small integer column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.SmallInteger("id")
			},
			want: "INTEGER",
		},
		{
			name: "tiny integer column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.TinyInteger("id")
			},
			want: "INTEGER",
		},
		{
			name: "float column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Float("val")
			},
			want: "REAL",
		},
		{
			name: "double column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Double("val")
			},
			want: "REAL",
		},
		{
			name: "decimal column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Decimal("val", 8, 2)
			},
			want: "NUMERIC",
		},
		{
			name: "boolean column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Boolean("val")
			},
			want: "INTEGER",
		},
		{
			name: "enum column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Enum("val", []string{"a", "b"})
			},
			want: "TEXT",
		},
		{
			name: "json column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.JSON("val")
			},
			want: "TEXT",
		},
		{
			name: "jsonb column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.JSONB("val")
			},
			want: "TEXT",
		},
		{
			name: "date column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Date("val")
			},
			want: "DATE",
		},
		{
			name: "datetime column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.DateTime("val", 0)
			},
			want: "DATETIME",
		},
		{
			name: "datetime tz column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.DateTimeTz("val", 0)
			},
			want: "DATETIME",
		},
		{
			name: "time column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Time("val")
			},
			want: "TEXT",
		},
		{
			name: "time tz column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.TimeTz("val", 0)
			},
			want: "TEXT",
		},
		{
			name: "timestamp column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Timestamp("val", 0)
			},
			want: "DATETIME",
		},
		{
			name: "timestamp tz column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.TimestampTz("val", 0)
			},
			want: "DATETIME",
		},
		{
			name: "year column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Year("val")
			},
			want: "INTEGER",
		},
		{
			name: "binary column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Binary("val")
			},
			want: "BLOB",
		},
		{
			name: "uuid column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.UUID("val")
			},
			want: "TEXT",
		},
		{
			name: "ulid column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.ULID("val")
			},
			want: "TEXT",
		},
		{
			name: "ip address column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.IPAddress("val")
			},
			want: "TEXT",
		},
		{
			name: "mac address column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.MacAddress("val")
			},
			want: "TEXT",
		},
		{
			name: "geometry column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Geometry("val", "POLYGON", 4326)
			},
			want: "TEXT",
		},
		{
			name: "geography column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Geography("val", "POLYGON", 4326)
			},
			want: "TEXT",
		},
		{
			name: "point column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Point("val")
			},
			want: "TEXT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &blueprint.Blueprint{Name: "test_table"}
			tt.blueprint(bp)
			got := g.GetType(bp.Columns[0])
			assert.Equal(t, tt.want, got, "Expected type to match for test case: %s", tt.name)
		})
	}
}
