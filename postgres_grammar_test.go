package schema

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
			want: []string{"CREATE TABLE users (id BIGSERIAL NOT NULL PRIMARY KEY, name VARCHAR(255) NOT NULL, email VARCHAR(255) NOT NULL UNIQUE, password VARCHAR(255) NULL, created_at TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP NOT NULL, updated_at TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP NOT NULL)"},
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
		{
			name: "Create table with column name is empty",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "empty_column_table"}
				bp.String("", 255) // Intentionally empty column name
				return bp
			}(),
			wantErr: true,
		},
		{
			name: "Create table with index name is empty",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.ID()
				bp.String("name", 255)
				bp.Index("")
				return bp
			}(),
			wantErr: true,
		},
		{
			name: "Create table with invalid foreign key",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "post"}
				bp.ID()
				bp.Integer("user_id")
				bp.Foreign("user_id")
				return bp
			}(),
			wantErr: true,
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
		{
			name: "Drop column from table",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.DropColumn("address")
				return bp
			}(),
			want:    []string{"ALTER TABLE users DROP COLUMN address"},
			wantErr: false,
		},
		{
			name: "Drop foreign key",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "posts"}
				bp.DropForeign("fk_posts_users")
				return bp
			}(),
			want: []string{"ALTER TABLE posts DROP CONSTRAINT IF EXISTS fk_posts_users"},
		},
		{
			name: "Drop primary key",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.DropPrimary("pk_users")
				return bp
			}(),
			want:    []string{"ALTER TABLE users DROP CONSTRAINT IF EXISTS pk_users"},
			wantErr: false,
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

func TestPgGrammar_GetColumns(t *testing.T) {
	grammar := newPgGrammar()

	tests := []struct {
		name      string
		blueprint *Blueprint
		want      []string
		wantErr   bool
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
			name: "All column types",
			blueprint: func() *Blueprint {
				bp := &Blueprint{}
				bp.ID()
				bp.Char("code", 10).Comment("Unique code for user")
				bp.String("name", 255)
				bp.Text("bio").Nullable()
				bp.BigInteger("company_id")
				bp.Double("price").Default(30.5)
				bp.Increments("stock").Default(100)
				bp.Integer("age").Default(30)
				bp.SmallIncrements("level").Default(1)
				bp.SmallInteger("rank").Default(5)
				bp.Boolean("is_active").Default(true)
				bp.Float("rating").Default(4.5)
				bp.Decimal("balance", 10, 2).Default(100.00)
				bp.Date("birth_date").Nullable()
				bp.Time("last_login").Nullable()
				bp.Timestamp("created_at").Default("CURRENT_TIMESTAMP")
				bp.TimestampTz("created_at_tz").Default("CURRENT_TIMESTAMP")
				bp.Year("year").Default(2023)
				bp.JSON("settings").Nullable()
				bp.JSONB("preferences").Nullable()
				bp.UUID("session_id").Nullable()
				return bp
			}(),
			want: []string{
				"id BIGSERIAL NOT NULL PRIMARY KEY",
				"code CHAR(10) NOT NULL COMMENT 'Unique code for user'",
				"name VARCHAR(255) NOT NULL",
				"bio TEXT NULL",
				"company_id BIGINT NOT NULL",
				"price DOUBLE PRECISION DEFAULT 30.5 NOT NULL",
				"stock SERIAL DEFAULT 100 NOT NULL",
				"age INTEGER DEFAULT 30 NOT NULL",
				"level SMALLSERIAL DEFAULT 1 NOT NULL",
				"rank SMALLINT DEFAULT 5 NOT NULL",
				"is_active BOOLEAN DEFAULT true NOT NULL",
				"rating REAL DEFAULT 4.5 NOT NULL",
				"balance DECIMAL(10, 2) DEFAULT 100 NOT NULL",
				"birth_date DATE NULL",
				"last_login TIME(0) NULL",
				"created_at TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP NOT NULL",
				"created_at_tz TIMESTAMPTZ(0) DEFAULT CURRENT_TIMESTAMP NOT NULL",
				"year INTEGER DEFAULT 2023 NOT NULL",
				"settings JSON NULL",
				"preferences JSONB NULL",
				"session_id UUID NULL",
			},
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
		{
			name: "Error on empty column",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.String("", 255)
				return bp
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := grammar.getColumns(tt.blueprint)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.Equal(t, tt.want, got)
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

func TestPgGrammar_GetIndexSqls(t *testing.T) {
	grammar := newPgGrammar()

	tests := []struct {
		name      string
		blueprint *Blueprint
		want      []string
		wantErr   bool
	}{
		{
			name: "Column with unique index and name",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.String("email").Unique("unique_email_idx")
				return bp
			}(),
			want: []string{"CREATE UNIQUE INDEX unique_email_idx ON users(email)"},
		},
		{
			name: "Column with regular index",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.Timestamp("created_at").Index()
				return bp
			}(),
			want: []string{"CREATE INDEX idx_users_created_at ON users(created_at)"},
		},
		{
			name: "Column with named index",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.Timestamp("last_login").Index("last_login_index")
				return bp
			}(),
			want: []string{"CREATE INDEX last_login_index ON users(last_login)"},
		},
		{
			name: "Primary key index",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.Primary("id")
				return bp
			}(),
			want: []string{"ALTER TABLE users ADD CONSTRAINT pk_users PRIMARY KEY (id)"},
		},
		{
			name: "Index with empty column",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.Index("")
				return bp
			}(),
			wantErr: true,
		},
		{
			name: "Index with algorithm",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.Index("email").Algorithm("HASH")
				return bp
			}(),
			want: []string{"CREATE INDEX idx_users_email ON users (email) USING HASH"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := grammar.getIndexSqls(tt.blueprint)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_GetDropColumnSqls(t *testing.T) {
	grammar := newPgGrammar()

	tests := []struct {
		name      string
		blueprint *Blueprint
		want      []string
		wantErr   bool
	}{
		{
			name: "Drop single column",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.DropColumn("email")
				return bp
			}(),
			want:    []string{"ALTER TABLE users DROP COLUMN email"},
			wantErr: false,
		},
		{
			name: "Drop multiple columns",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.DropColumn("email", "phone", "address")
				return bp
			}(),
			want: []string{
				"ALTER TABLE users DROP COLUMN email",
				"ALTER TABLE users DROP COLUMN phone",
				"ALTER TABLE users DROP COLUMN address",
			},
			wantErr: false,
		},
		{
			name: "Empty column name",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.DropColumn("")
				return bp
			}(),
			wantErr: true,
		},
		{
			name: "Mixed valid and empty column names",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.DropColumn("email", "")
				return bp
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := grammar.getDropColumnSqls(tt.blueprint)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_GetRenameColumnSqls(t *testing.T) {
	grammar := newPgGrammar()

	tests := []struct {
		name      string
		blueprint *Blueprint
		want      []string
		wantErr   bool
	}{
		{
			name: "Rename single column",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.RenameColumn("email", "email_address")
				return bp
			}(),
			want:    []string{"ALTER TABLE users RENAME COLUMN email TO email_address"},
			wantErr: false,
		},
		{
			name: "Rename multiple columns",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.RenameColumn("email", "email_address")
				bp.RenameColumn("name", "full_name")
				return bp
			}(),
			want: []string{
				"ALTER TABLE users RENAME COLUMN email TO email_address",
				"ALTER TABLE users RENAME COLUMN name TO full_name",
			},
			wantErr: false,
		},
		{
			name: "Empty old column name",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.RenameColumn("", "email_address")
				return bp
			}(),
			wantErr: true,
		},
		{
			name: "Empty new column name",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.RenameColumn("email", "")
				return bp
			}(),
			wantErr: true,
		},
		{
			name: "Both column names empty",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.RenameColumn("", "")
				return bp
			}(),
			wantErr: true,
		},
		{
			name: "Mixed valid and invalid column names",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.RenameColumn("email", "email_address")
				bp.RenameColumn("name", "")
				return bp
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := grammar.getRenameColumnSqls(tt.blueprint)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			// Since map iteration order is not guaranteed, we need to check that all
			// expected SQL statements are present regardless of order
			require.Len(t, got, len(tt.want), "Expected %d SQL statements, got %d", len(tt.want), len(got))

			// Sort both slices for comparison
			sort.Strings(got)
			sort.Strings(tt.want)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_GetDropIndexSqls(t *testing.T) {
	grammar := newPgGrammar()

	tests := []struct {
		name      string
		blueprint *Blueprint
		want      []string
		wantErr   bool
	}{
		{
			name: "Drop single index",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.DropIndex("idx_users_email")
				return bp
			}(),
			want:    []string{"DROP INDEX IF EXISTS idx_users_email"},
			wantErr: false,
		},
		{
			name: "Drop multiple indexes",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.DropIndex("idx_users_email")
				bp.DropIndex("idx_users_created_at")
				return bp
			}(),
			want: []string{
				"DROP INDEX IF EXISTS idx_users_email",
				"DROP INDEX IF EXISTS idx_users_created_at",
			},
			wantErr: false,
		},
		{
			name: "Drop unique indexes",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.DropUnique("uk_users_email")
				return bp
			}(),
			want: []string{
				"DROP INDEX IF EXISTS uk_users_email",
			},
			wantErr: false,
		},
		{
			name: "Drop both regular and unique indexes",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.DropIndex("idx_users_created_at")
				bp.DropUnique("uk_users_email")
				return bp
			}(),
			want: []string{
				"DROP INDEX IF EXISTS idx_users_created_at",
				"DROP INDEX IF EXISTS uk_users_email",
			},
			wantErr: false,
		},
		{
			name: "Empty index name",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.DropIndex("")
				return bp
			}(),
			wantErr: true,
		},
		{
			name: "Empty unique index name",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.DropUnique("")
				return bp
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := grammar.getDropIndexSqls(tt.blueprint)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			require.Len(t, got, len(tt.want), "Expected %d SQL statements, got %d", len(tt.want), len(got))

			// Sort both slices for comparison as the order of indexes might not be guaranteed
			sort.Strings(got)
			sort.Strings(tt.want)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_GetRenameIndexSqls(t *testing.T) {
	grammar := newPgGrammar()

	tests := []struct {
		name      string
		blueprint *Blueprint
		want      []string
		wantErr   bool
	}{
		{
			name: "Rename single index",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.RenameIndex("idx_users_email", "idx_users_email_address")
				return bp
			}(),
			want:    []string{"ALTER INDEX idx_users_email RENAME TO idx_users_email_address"},
			wantErr: false,
		},
		{
			name: "Rename multiple indexes",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.RenameIndex("idx_users_email", "idx_users_email_address")
				bp.RenameIndex("idx_users_name", "idx_users_full_name")
				return bp
			}(),
			want: []string{
				"ALTER INDEX idx_users_email RENAME TO idx_users_email_address",
				"ALTER INDEX idx_users_name RENAME TO idx_users_full_name",
			},
			wantErr: false,
		},
		{
			name: "Empty old index name",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.RenameIndex("", "idx_users_email_address")
				return bp
			}(),
			wantErr: true,
		},
		{
			name: "Empty new index name",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.RenameIndex("idx_users_email", "")
				return bp
			}(),
			wantErr: true,
		},
		{
			name: "Both index names empty",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.RenameIndex("", "")
				return bp
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := grammar.getRenameIndexSqls(tt.blueprint)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			// Since map iteration order is not guaranteed, we need to check that all
			// expected SQL statements are present regardless of order
			require.Len(t, got, len(tt.want), "Expected %d SQL statements, got %d", len(tt.want), len(got))

			// Sort both slices for comparison
			sort.Strings(got)
			sort.Strings(tt.want)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_GetDropForeignKeySqls(t *testing.T) {
	grammar := newPgGrammar()

	tests := []struct {
		name      string
		blueprint *Blueprint
		want      []string
		wantErr   bool
	}{
		{
			name: "Drop single foreign key",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "posts"}
				bp.DropForeign("fk_posts_users")
				return bp
			}(),
			want:    []string{"ALTER TABLE posts DROP CONSTRAINT IF EXISTS fk_posts_users"},
			wantErr: false,
		},
		{
			name: "Drop multiple foreign keys",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "comments"}
				bp.DropForeign("fk_comments_posts")
				bp.DropForeign("fk_comments_users")
				return bp
			}(),
			want: []string{
				"ALTER TABLE comments DROP CONSTRAINT IF EXISTS fk_comments_posts",
				"ALTER TABLE comments DROP CONSTRAINT IF EXISTS fk_comments_users",
			},
			wantErr: false,
		},
		{
			name: "Empty foreign key name",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "posts"}
				bp.DropForeign("")
				bp.dropForeignKeys = []string{""}
				return bp
			}(),
			wantErr: true,
		},
		{
			name: "Mixed valid and invalid foreign key names",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "orders"}
				bp.DropForeign("fk_orders_customers")
				bp.DropForeign("")
				return bp
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := grammar.getDropForeignKeySqls(tt.blueprint)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			require.Len(t, got, len(tt.want), "Expected %d SQL statements, got %d", len(tt.want), len(got))

			// Sort both slices for comparison as order might not be guaranteed
			sort.Strings(got)
			sort.Strings(tt.want)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_GetDropPrimaryKeySqls(t *testing.T) {
	grammar := newPgGrammar()

	tests := []struct {
		name      string
		blueprint *Blueprint
		want      []string
		wantErr   bool
	}{
		{
			name: "Drop single primary key",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.DropPrimary("pk_users")
				return bp
			}(),
			want:    []string{"ALTER TABLE users DROP CONSTRAINT IF EXISTS pk_users"},
			wantErr: false,
		},
		{
			name: "Empty primary key name",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.DropPrimary("")
				return bp
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := grammar.getDropPrimaryKeySqls(tt.blueprint)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			require.Len(t, got, len(tt.want), "Expected %d SQL statements, got %d", len(tt.want), len(got))

			// Sort both slices for comparison as order might not be guaranteed
			sort.Strings(got)
			sort.Strings(tt.want)
			assert.Equal(t, tt.want, got)
		})
	}
}
