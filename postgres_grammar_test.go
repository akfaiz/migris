package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPgGrammar_GetColumns(t *testing.T) {
	grammar := newPgGrammar()

	tests := []struct {
		name      string
		blueprint *Blueprint
		want      []string
	}{
		{
			name: "Simple column",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.String("name", 255)
				return bp
			}(),
			want: []string{"name VARCHAR(255) NOT NULL"},
		},
		{
			name: "Nullable column",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.String("email", 255).Nullable()
				return bp
			}(),
			want: []string{"email VARCHAR(255) NULL"},
		},
		{
			name: "Column with default value",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.Boolean("active").Default(true)
				return bp
			}(),
			want: []string{"active BOOLEAN DEFAULT true NOT NULL"},
		},
		{
			name: "Primary key column",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.Integer("id").Primary()
				return bp
			}(),
			want: []string{"id INTEGER NOT NULL PRIMARY KEY"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := grammar.getColumns(tt.blueprint)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompileCreate(t *testing.T) {
	grammar := newPgGrammar()

	tests := []struct {
		name      string
		blueprint *Blueprint
		want      []string
		wantErr   bool
	}{
		{
			name: "Create simple table",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.ID()
				bp.String("name", 255)
				bp.String("email", 255).Unique()
				bp.String("password").Nullable()
				bp.Timestamp("created_at").Default("CURRENT_TIMESTAMP")
				bp.Timestamp("updated_at").Default("CURRENT_TIMESTAMP")
				return bp
			}(),
			want:    []string{"CREATE TABLE users (id BIGSERIAL NOT NULL PRIMARY KEY, name VARCHAR(255) NOT NULL, email VARCHAR(255) NOT NULL UNIQUE, password VARCHAR(255) NULL, created_at TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP NOT NULL, updated_at TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP NOT NULL)"},
			wantErr: false,
		},
		{
			name: "Create table with foreign key",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "posts"}
				bp.ID()
				bp.Integer("user_id")
				bp.String("title", 255)
				bp.Text("content").Nullable()
				bp.Foreign("user_id").References("id").On("users").OnDelete("CASCADE").OnUpdate("CASCADE")
				return bp
			}(),
			want: []string{
				"CREATE TABLE posts (id BIGSERIAL NOT NULL PRIMARY KEY, user_id INTEGER NOT NULL, title VARCHAR(255) NOT NULL, content TEXT NULL)",
				"ALTER TABLE posts ADD CONSTRAINT fk_posts_users FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := grammar.compileCreate(tt.blueprint)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			require.Len(t, got, len(tt.want), "Expected %d SQL statements, got %d", len(tt.want), len(got))
			for i, sql := range got {
				assert.Equal(t, tt.want[i], sql, "SQL statement mismatch at index %d", i)
			}
		})
	}
}

func TestPgGrammar_CompileCreateIfNotExists(t *testing.T) {
	grammar := newPgGrammar()

	tests := []struct {
		name      string
		blueprint *Blueprint
		want      []string
		wantErr   bool
	}{
		{
			name: "Create simple table if not exists",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.ID()
				bp.String("name", 255)
				bp.String("email", 255).Unique()
				bp.String("password").Nullable()
				bp.Timestamp("created_at").Default("CURRENT_TIMESTAMP")
				bp.Timestamp("updated_at").Default("CURRENT_TIMESTAMP")
				return bp
			}(),
			want:    []string{"CREATE TABLE IF NOT EXISTS users (id BIGSERIAL NOT NULL PRIMARY KEY, name VARCHAR(255) NOT NULL, email VARCHAR(255) NOT NULL UNIQUE, password VARCHAR(255) NULL, created_at TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP NOT NULL, updated_at TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP NOT NULL)"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := grammar.compileCreateIfNotExists(tt.blueprint)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			require.Len(t, got, len(tt.want), "Expected %d SQL statements, got %d", len(tt.want), len(got))
			for i, sql := range got {
				assert.Equal(t, tt.want[i], sql, "SQL statement mismatch at index %d", i)
			}
		})
	}
}

func TestPgGrammar_CompileAlter(t *testing.T) {
	grammar := newPgGrammar()

	tests := []struct {
		name      string
		blueprint *Blueprint
		want      []string
		wantErr   bool
	}{
		{
			name: "Add column to table",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.String("address", 255)
				return bp
			}(),
			want:    []string{"ALTER TABLE users ADD COLUMN address VARCHAR(255) NOT NULL"},
			wantErr: false,
		},
		{
			name: "Add index to column",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.Unique("email")
				bp.Primary("id")
				bp.Index("created_at").Algorithm("BTREE")
				return bp
			}(),
			want: []string{
				"CREATE UNIQUE INDEX uk_users_email ON users (email)",
				"ALTER TABLE users ADD CONSTRAINT pk_users PRIMARY KEY (id)",
				"CREATE INDEX idx_users_created_at ON users (created_at) USING BTREE",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := grammar.compileAlter(tt.blueprint)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			require.Len(t, got, len(tt.want), "Expected %d SQL statements, got %d", len(tt.want), len(got))
			for i, sql := range got {
				assert.Equal(t, tt.want[i], sql, "SQL statement mismatch at index %d", i)
			}
		})
	}
}

func TestPgGrammar_GetType(t *testing.T) {
	grammar := newPgGrammar()

	tests := []struct {
		name string
		col  *columnDefinition
		want string
	}{
		{
			name: "Integer type",
			col:  &columnDefinition{columnType: columnTypeInteger},
			want: "INTEGER",
		},
		{
			name: "Char type",
			col:  &columnDefinition{columnType: columnTypeChar, length: 2},
			want: "CHAR(2)",
		},
		{
			name: "String type with length",
			col:  &columnDefinition{columnType: columnTypeString, length: 100},
			want: "VARCHAR(100)",
		},
		{
			name: "Decimal type with precision and scale",
			col:  &columnDefinition{columnType: columnTypeDecimal, precision: 10, scale: 2},
			want: "DECIMAL(10, 2)",
		},
		{
			name: "Time",
			col:  &columnDefinition{columnType: columnTypeTime, precision: 0},
			want: "TIME(0)",
		},
		{
			name: "Timestamp with precision",
			col:  &columnDefinition{columnType: columnTypeTimestamp, precision: 3},
			want: "TIMESTAMP(3)",
		},
		{
			name: "Timestamp with time zone",
			col:  &columnDefinition{columnType: columnTypeTimestampTz, precision: 3},
			want: "TIMESTAMPTZ(3)",
		},
		{
			name: "JSON type",
			col:  &columnDefinition{columnType: columnTypeJSON},
			want: "JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := grammar.getType(tt.col)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompileDrop(t *testing.T) {
	grammar := newPgGrammar()

	tests := []struct {
		name      string
		blueprint *Blueprint
		want      string
		wantErr   bool
	}{
		{
			name: "Drop table",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				return bp
			}(),
			want:    "DROP TABLE users",
			wantErr: false,
		},
		{
			name: "Drop table with schema",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "public.users"}
				return bp
			}(),
			want:    "DROP TABLE public.users",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := grammar.compileDrop(tt.blueprint)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompileDropIfExists(t *testing.T) {
	grammar := newPgGrammar()

	tests := []struct {
		name      string
		blueprint *Blueprint
		want      string
		wantErr   bool
	}{
		{
			name: "Drop table if exists",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				return bp
			}(),
			want:    "DROP TABLE IF EXISTS users",
			wantErr: false,
		},
		{
			name: "Drop table with schema if exists",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "public.users"}
				return bp
			}(),
			want:    "DROP TABLE IF EXISTS public.users",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := grammar.compileDropIfExists(tt.blueprint)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompileRename(t *testing.T) {
	grammar := newPgGrammar()

	tests := []struct {
		name      string
		blueprint *Blueprint
		want      string
		wantErr   bool
	}{
		{
			name: "Rename table",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users", newName: "people"}
				return bp
			}(),
			want:    "ALTER TABLE users RENAME TO people",
			wantErr: false,
		},
		{
			name: "Rename table with schema",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "public.users", newName: "public.people"}
				return bp
			}(),
			want:    "ALTER TABLE public.users RENAME TO public.people",
			wantErr: false,
		},
		{
			name: "Rename table without new name",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				return bp
			}(),
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := grammar.compileRename(tt.blueprint)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
