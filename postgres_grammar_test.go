package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPgGrammar_CompileCreate(t *testing.T) {
	grammar := newPgGrammar()

	tests := []struct {
		name      string
		blueprint *Blueprint
		want      string
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
			want: "CREATE TABLE users (id BIGSERIAL NOT NULL PRIMARY KEY, name VARCHAR(255) NOT NULL, email VARCHAR(255) NOT NULL UNIQUE, password VARCHAR NULL, created_at TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP NOT NULL, updated_at TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP NOT NULL)",
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
			want: "CREATE TABLE posts (id BIGSERIAL NOT NULL PRIMARY KEY, user_id INTEGER NOT NULL, title VARCHAR(255) NOT NULL, content TEXT NULL)",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := grammar.compileCreate(tt.blueprint)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got, "SQL statement mismatch for %s", tt.name)
		})
	}
}

func TestPgGrammar_CompileCreateIfNotExists(t *testing.T) {
	grammar := newPgGrammar()

	tests := []struct {
		name      string
		blueprint *Blueprint
		want      string
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
			want:    "CREATE TABLE IF NOT EXISTS users (id BIGSERIAL NOT NULL PRIMARY KEY, name VARCHAR(255) NOT NULL, email VARCHAR(255) NOT NULL UNIQUE, password VARCHAR NULL, created_at TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP NOT NULL, updated_at TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP NOT NULL)",
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
			assert.Equal(t, tt.want, got, "SQL statement mismatch for %s", tt.name)
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
			col:  &columnDefinition{columnType: columnTypeChar},
			want: "CHAR",
		},
		{
			name: "String type with length",
			col:  &columnDefinition{columnType: columnTypeString, length: 255},
			want: "VARCHAR(255)",
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

func TestPgGrammar_CompileDropColumn(t *testing.T) {
	grammar := newPgGrammar()

	tests := []struct {
		name      string
		blueprint *Blueprint
		want      string
		wantErr   bool
	}{
		{
			name: "Drop single column",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.dropColumns = []string{"email"}
				return bp
			}(),
			want:    "ALTER TABLE users DROP COLUMN email",
			wantErr: false,
		},
		{
			name: "Drop multiple columns",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.dropColumns = []string{"email", "phone", "address"}
				return bp
			}(),
			want:    "ALTER TABLE users DROP COLUMN email, DROP COLUMN phone, DROP COLUMN address",
			wantErr: false,
		},
		{
			name: "No columns to drop",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				bp.dropColumns = []string{}
				return bp
			}(),
			want:    "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := grammar.compileDropColumn(tt.blueprint)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompileRenameColumn(t *testing.T) {
	grammar := newPgGrammar()

	tests := []struct {
		name      string
		blueprint *Blueprint
		oldName   string
		newName   string
		want      string
		wantErr   bool
	}{
		{
			name: "Rename column",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				return bp
			}(),
			oldName: "email",
			newName: "user_email",
			want:    "ALTER TABLE users RENAME COLUMN email TO user_email",
			wantErr: false,
		},
		{
			name: "Empty old name",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				return bp
			}(),
			oldName: "",
			newName: "user_email",
			wantErr: true,
		},
		{
			name: "Empty new name",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				return bp
			}(),
			oldName: "email",
			newName: "",
			wantErr: true,
		},
		{
			name: "Both names empty",
			blueprint: func() *Blueprint {
				bp := &Blueprint{name: "users"}
				return bp
			}(),
			oldName: "",
			newName: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := grammar.compileRenameColumn(tt.blueprint, tt.oldName, tt.newName)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompileIndexSql(t *testing.T) {
	grammar := newPgGrammar()

	tests := []struct {
		name      string
		blueprint *Blueprint
		index     *indexDefinition
		want      string
		wantErr   bool
	}{
		{
			name: "Primary key index",
			blueprint: func() *Blueprint {
				return &Blueprint{name: "users"}
			}(),
			index: &indexDefinition{
				indexType: indexTypePrimary,
				columns:   []string{"id"},
				name:      "users_pkey",
			},
			want:    "ALTER TABLE users ADD CONSTRAINT users_pkey PRIMARY KEY (id)",
			wantErr: false,
		},
		{
			name: "Unique index",
			blueprint: func() *Blueprint {
				return &Blueprint{name: "users"}
			}(),
			index: &indexDefinition{
				indexType: indexTypeUnique,
				columns:   []string{"email"},
				name:      "users_email_unique",
			},
			want:    "CREATE UNIQUE INDEX users_email_unique ON users(email)",
			wantErr: false,
		},
		{
			name: "Regular index",
			blueprint: func() *Blueprint {
				return &Blueprint{name: "users"}
			}(),
			index: &indexDefinition{
				indexType: indexTypeIndex,
				columns:   []string{"name"},
				name:      "users_name_index",
			},
			want:    "CREATE INDEX users_name_index ON users(name)",
			wantErr: false,
		},
		{
			name: "Index with algorithm",
			blueprint: func() *Blueprint {
				return &Blueprint{name: "users"}
			}(),
			index: &indexDefinition{
				indexType:  indexTypeIndex,
				columns:    []string{"name", "email"},
				name:       "users_name_email_index",
				algorithmn: "hash",
			},
			want:    "CREATE INDEX users_name_email_index ON users(name, email) USING hash",
			wantErr: false,
		},
		{
			name: "Composite index",
			blueprint: func() *Blueprint {
				return &Blueprint{name: "users"}
			}(),
			index: &indexDefinition{
				indexType: indexTypeIndex,
				columns:   []string{"first_name", "last_name"},
				name:      "users_name_index",
			},
			want:    "CREATE INDEX users_name_index ON users(first_name, last_name)",
			wantErr: false,
		},
		{
			name: "Index with empty column",
			blueprint: func() *Blueprint {
				return &Blueprint{name: "users"}
			}(),
			index: &indexDefinition{
				indexType: indexTypeIndex,
				columns:   []string{"name", ""},
				name:      "users_name_index",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := grammar.compileIndexSql(tt.blueprint, tt.index)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompileDropIndex(t *testing.T) {
	grammar := newPgGrammar()

	tests := []struct {
		name      string
		indexName string
		want      string
		wantErr   bool
	}{
		{
			name:      "Drop index with valid name",
			indexName: "users_email_index",
			want:      "DROP INDEX users_email_index",
			wantErr:   false,
		},
		{
			name:      "Drop index with complex name",
			indexName: "idx_users_email_name",
			want:      "DROP INDEX idx_users_email_name",
			wantErr:   false,
		},
		{
			name:      "Empty index name",
			indexName: "",
			want:      "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := grammar.compileDropIndex(tt.indexName)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, got)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompileRenameIndex(t *testing.T) {
	grammar := newPgGrammar()

	tests := []struct {
		name      string
		blueprint *Blueprint
		oldName   string
		newName   string
		want      string
		wantErr   bool
	}{
		{
			name: "Rename index with valid names",
			blueprint: func() *Blueprint {
				return &Blueprint{name: "users"}
			}(),
			oldName: "users_email_index",
			newName: "users_email_unique",
			want:    "ALTER INDEX users_email_index RENAME TO users_email_unique",
			wantErr: false,
		},
		{
			name: "Rename index with complex names",
			blueprint: func() *Blueprint {
				return &Blueprint{name: "users"}
			}(),
			oldName: "idx_users_email_name",
			newName: "idx_users_email_name_unique",
			want:    "ALTER INDEX idx_users_email_name RENAME TO idx_users_email_name_unique",
			wantErr: false,
		},
		{
			name: "Empty old name",
			blueprint: func() *Blueprint {
				return &Blueprint{name: "users"}
			}(),
			oldName: "",
			newName: "users_email_unique",
			wantErr: true,
		},
		{
			name: "Empty new name",
			blueprint: func() *Blueprint {
				return &Blueprint{name: "users"}
			}(),
			oldName: "users_email_index",
			newName: "",
			wantErr: true,
		},
		{
			name: "Both names empty",
			blueprint: func() *Blueprint {
				return &Blueprint{name: "users"}
			}(),
			oldName: "",
			newName: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := grammar.compileRenameIndex(tt.blueprint, tt.oldName, tt.newName)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompileDropPrimaryKey(t *testing.T) {
	grammar := newPgGrammar()

	tests := []struct {
		name      string
		blueprint *Blueprint
		indexName string
		want      string
		wantErr   bool
	}{
		{
			name: "Drop primary key with specified name",
			blueprint: func() *Blueprint {
				return &Blueprint{name: "users"}
			}(),
			indexName: "users_pkey",
			want:      "ALTER TABLE users DROP CONSTRAINT users_pkey",
			wantErr:   false,
		},
		{
			name: "Drop primary key with empty index name (should use default naming)",
			blueprint: func() *Blueprint {
				return &Blueprint{name: "posts"}
			}(),
			indexName: "",
			want:      "ALTER TABLE posts DROP CONSTRAINT pk_posts",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := grammar.compileDropPrimaryKey(tt.blueprint, tt.indexName)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompileForeignKeySql(t *testing.T) {
	grammar := newPgGrammar()

	tests := []struct {
		name       string
		blueprint  *Blueprint
		foreignKey *foreignKeyDefinition
		want       string
		wantErr    bool
	}{
		{
			name: "Basic foreign key",
			blueprint: func() *Blueprint {
				return &Blueprint{name: "posts"}
			}(),
			foreignKey: &foreignKeyDefinition{
				column:     "user_id",
				on:         "users",
				references: "id",
			},
			want:    "ALTER TABLE posts ADD CONSTRAINT fk_posts_users FOREIGN KEY (user_id) REFERENCES users(id)",
			wantErr: false,
		},
		{
			name: "Foreign key with onDelete CASCADE",
			blueprint: func() *Blueprint {
				return &Blueprint{name: "comments"}
			}(),
			foreignKey: &foreignKeyDefinition{
				column:     "post_id",
				on:         "posts",
				references: "id",
				onDelete:   "CASCADE",
			},
			want:    "ALTER TABLE comments ADD CONSTRAINT fk_comments_posts FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE",
			wantErr: false,
		},
		{
			name: "Foreign key with onUpdate CASCADE",
			blueprint: func() *Blueprint {
				return &Blueprint{name: "order_items"}
			}(),
			foreignKey: &foreignKeyDefinition{
				column:     "order_id",
				on:         "orders",
				references: "id",
				onUpdate:   "CASCADE",
			},
			want:    "ALTER TABLE order_items ADD CONSTRAINT fk_order_items_orders FOREIGN KEY (order_id) REFERENCES orders(id) ON UPDATE CASCADE",
			wantErr: false,
		},
		{
			name: "Foreign key with empty column",
			blueprint: func() *Blueprint {
				return &Blueprint{name: "users"}
			}(),
			foreignKey: &foreignKeyDefinition{
				column:     "",
				on:         "roles",
				references: "id",
			},
			wantErr: true,
		},
		{
			name: "Foreign key with empty on",
			blueprint: func() *Blueprint {
				return &Blueprint{name: "users"}
			}(),
			foreignKey: &foreignKeyDefinition{
				column:     "role_id",
				on:         "",
				references: "id",
			},
			wantErr: true,
		},
		{
			name: "Foreign key with empty references",
			blueprint: func() *Blueprint {
				return &Blueprint{name: "users"}
			}(),
			foreignKey: &foreignKeyDefinition{
				column:     "role_id",
				on:         "roles",
				references: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := grammar.compileForeignKeySql(tt.blueprint, tt.foreignKey)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompileDropForeignKey(t *testing.T) {
	grammar := newPgGrammar()

	tests := []struct {
		name           string
		blueprint      *Blueprint
		foreignKeyName string
		want           string
		wantErr        bool
	}{
		{
			name: "Drop foreign key with valid name",
			blueprint: func() *Blueprint {
				return &Blueprint{name: "posts"}
			}(),
			foreignKeyName: "fk_posts_users",
			want:           "ALTER TABLE posts DROP CONSTRAINT fk_posts_users",
			wantErr:        false,
		},
		{
			name: "Drop foreign key with complex name",
			blueprint: func() *Blueprint {
				return &Blueprint{name: "order_items"}
			}(),
			foreignKeyName: "fk_order_items_products_cascade",
			want:           "ALTER TABLE order_items DROP CONSTRAINT fk_order_items_products_cascade",
			wantErr:        false,
		},
		{
			name: "Empty foreign key name",
			blueprint: func() *Blueprint {
				return &Blueprint{name: "users"}
			}(),
			foreignKeyName: "",
			want:           "",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := grammar.compileDropForeignKey(tt.blueprint, tt.foreignKeyName)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
