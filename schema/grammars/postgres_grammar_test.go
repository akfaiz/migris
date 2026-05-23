package grammars_test

import (
	"testing"

	"github.com/akfaiz/migris/schema/blueprint"
	"github.com/akfaiz/migris/schema/grammars"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPgGrammar_CompileCreate(t *testing.T) {
	g, err := grammars.NewGrammar("postgres")
	require.NoError(t, err)

	tests := []struct {
		name      string
		table     string
		blueprint func(table *blueprint.Blueprint)
		want      string
		wantErr   bool
	}{
		{
			name: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.ID()
				table.String("name")
				table.String("email")
				table.String("password").Nullable()
				table.Timestamp("created_at").UseCurrent()
				table.Timestamp("updated_at").UseCurrent()
			},
			want: "CREATE TABLE \"users\" (\"id\" BIGSERIAL NOT NULL, \"name\" VARCHAR(255) NOT NULL, \"email\" VARCHAR(255) NOT NULL, \"password\" VARCHAR(255) NULL, \"created_at\" TIMESTAMP(0) NOT NULL DEFAULT CURRENT_TIMESTAMP, \"updated_at\" TIMESTAMP(0) NOT NULL DEFAULT CURRENT_TIMESTAMP, CONSTRAINT \"users_id_primary\" PRIMARY KEY (\"id\"))",
		},
		{
			name: "posts",
			blueprint: func(table *blueprint.Blueprint) {
				table.ID()
				table.Integer("user_id")
				table.String("title")
				table.Text("content").Nullable()
				table.Foreign("user_id").References("id").On("users").OnDelete("CASCADE").OnUpdate("CASCADE")
			},
			want: "CREATE TABLE \"posts\" (\"id\" BIGSERIAL NOT NULL, \"user_id\" INTEGER NOT NULL, \"title\" VARCHAR(255) NOT NULL, \"content\" TEXT NULL, CONSTRAINT \"posts_id_primary\" PRIMARY KEY (\"id\"))",
		},
		{
			name: "empty_column_table",
			blueprint: func(table *blueprint.Blueprint) {
				table.String("") // Intentionally empty column name
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tableName := tt.table
			if tableName == "" && !tt.wantErr {
				tableName = tt.name
			}
			bp := &blueprint.Blueprint{Name: tableName}
			tt.blueprint(bp)
			got, err := g.CompileCreate(bp)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got, "SQL statement mismatch for %s", tt.name)
		})
	}
}

func TestPgGrammar_CompileTableExists(t *testing.T) {
	g, err := grammars.NewGrammar("postgres")
	require.NoError(t, err)

	sql, err := g.CompileTableExists("", "public.users")
	require.NoError(t, err)
	assert.Equal(
		t,
		"SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'users'",
		sql,
	)
}

func TestPgGrammar_CompileTables(t *testing.T) {
	g, err := grammars.NewGrammar("postgres")
	require.NoError(t, err)

	sql, err := g.CompileTables("")
	require.NoError(t, err)
	assert.Contains(t, sql, "SELECT \n\t\t\tt.table_name")
}

func TestPgGrammar_CompileColumns(t *testing.T) {
	g, err := grammars.NewGrammar("postgres")
	require.NoError(t, err)

	sql, err := g.CompileColumns("", "public.users")
	require.NoError(t, err)
	assert.Contains(t, sql, "cols.table_schema = 'public' \n\t\t\tAND cols.table_name = 'users'")
}

func TestPgGrammar_CompileIndexes(t *testing.T) {
	g, err := grammars.NewGrammar("postgres")
	require.NoError(t, err)

	sql, err := g.CompileIndexes("", "public.users")
	require.NoError(t, err)
	assert.Contains(t, sql, "t.relname = 'users'\n\t\t\tAND n.nspname = 'public'")
}

func TestPgGrammar_CompileAdd(t *testing.T) {
	g, err := grammars.NewGrammar("postgres")
	require.NoError(t, err)

	tests := []struct {
		name      string
		table     string
		blueprint func(table *blueprint.Blueprint)
		want      string
		wantErr   bool
	}{
		{
			name: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.String("phone", 20)
			},
			want:    "ALTER TABLE \"users\" ADD COLUMN \"phone\" VARCHAR(20) NOT NULL",
			wantErr: false,
		},
		{
			name: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.String("phone", 20)
				table.String("address", 255).Nullable()
				table.Integer("age")
			},
			want:    "ALTER TABLE \"users\" ADD COLUMN \"phone\" VARCHAR(20) NOT NULL, ADD COLUMN \"address\" VARCHAR(255) NULL, ADD COLUMN \"age\" INTEGER NOT NULL",
			wantErr: false,
		},
		{
			name: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Boolean("active").Default(true)
			},
			want:    "ALTER TABLE \"users\" ADD COLUMN \"active\" BOOLEAN NOT NULL DEFAULT '1'",
			wantErr: false,
		},
		{
			name: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.String("notes", 500).Comment("User notes")
			},
			want:    "ALTER TABLE \"users\" ADD COLUMN \"notes\" VARCHAR(500) NOT NULL",
			wantErr: false,
		},
		{
			name: "categories",
			blueprint: func(table *blueprint.Blueprint) {
				table.Integer("id").Primary()
			},
			want:    "ALTER TABLE \"categories\" ADD COLUMN \"id\" INTEGER NOT NULL, ADD CONSTRAINT \"categories_id_primary\" PRIMARY KEY (\"id\")",
			wantErr: false,
		},
		{
			name: "logs",
			blueprint: func(table *blueprint.Blueprint) {
				table.BigInteger("id").AutoIncrement()
			},
			want:    "ALTER TABLE \"logs\" ADD COLUMN \"id\" BIGSERIAL NOT NULL",
			wantErr: false,
		},
		{
			name: "products",
			blueprint: func(table *blueprint.Blueprint) {
				table.Decimal("price", 10, 2).Default(0)
			},
			want:    "ALTER TABLE \"products\" ADD COLUMN \"price\" DECIMAL(10, 2) NOT NULL DEFAULT '0'",
			wantErr: false,
		},
		{
			name: "orders",
			blueprint: func(table *blueprint.Blueprint) {
				table.Timestamp("created_at").UseCurrent()
				table.Timestamp("updated_at").UseCurrent().Nullable()
			},
			want:    "ALTER TABLE \"orders\" ADD COLUMN \"created_at\" TIMESTAMP(0) NOT NULL DEFAULT CURRENT_TIMESTAMP, ADD COLUMN \"updated_at\" TIMESTAMP(0) NULL DEFAULT CURRENT_TIMESTAMP",
			wantErr: false,
		},
		{
			name: "mixed_table",
			blueprint: func(table *blueprint.Blueprint) {
				table.Text("description")
				table.JSON("metadata").Nullable()
				table.UUID("reference_id")
				table.Date("event_date")
			},
			want:    "ALTER TABLE \"mixed_table\" ADD COLUMN \"description\" TEXT NOT NULL, ADD COLUMN \"metadata\" JSON NULL, ADD COLUMN \"reference_id\" UUID NOT NULL, ADD COLUMN \"event_date\" DATE NOT NULL",
			wantErr: false,
		},
		{
			name:      "No columns to add",
			table:     "users",
			blueprint: func(_ *blueprint.Blueprint) {},
			want:      "",
			wantErr:   false,
		},
		{
			name: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.String("", 255) // Intentionally empty column name
			},
			wantErr: true,
		},
		{
			name: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Enum("status", []string{"active", "inactive", "pending"})
			},
			want:    "ALTER TABLE \"users\" ADD COLUMN \"status\" VARCHAR(255) CHECK (status IN ('active', 'inactive', 'pending')) NOT NULL",
			wantErr: false,
		},
		{
			name: "locations",
			blueprint: func(table *blueprint.Blueprint) {
				table.Geography("coordinates", "POINT", 4326)
			},
			want:    "ALTER TABLE \"locations\" ADD COLUMN \"coordinates\" GEOGRAPHY(POINT, 4326) NOT NULL",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tableName := tt.table
			if tableName == "" && !tt.wantErr {
				tableName = tt.name
			}
			bp := &blueprint.Blueprint{Name: tableName}
			tt.blueprint(bp)
			got, err := g.CompileAdd(bp)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompileChange(t *testing.T) {
	g, err := grammars.NewGrammar("postgres")
	require.NoError(t, err)

	tests := []struct {
		name      string
		table     string
		blueprint func(blueprint *blueprint.Blueprint)
		want      []string
		wantErr   bool
	}{
		{
			name: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.String("email", 500).Nullable().Change()
			},
			want: []string{
				"ALTER TABLE \"users\" ALTER COLUMN \"email\" TYPE VARCHAR(500), ALTER COLUMN \"email\" DROP NOT NULL",
			},
		},
		{
			name: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.String("email", 500).Default("user@mail.com").Change()
			},
			want: []string{
				"ALTER TABLE \"users\" ALTER COLUMN \"email\" TYPE VARCHAR(500), ALTER COLUMN \"email\" SET DEFAULT 'user@mail.com'",
			},
		},
		{
			name: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.String("email", 500).Nullable().Change()
				table.String("name", 255).Default("Anonymous").Change()
			},
			want: []string{
				"ALTER TABLE \"users\" ALTER COLUMN \"email\" TYPE VARCHAR(500), ALTER COLUMN \"email\" DROP NOT NULL",
				"ALTER TABLE \"users\" ALTER COLUMN \"name\" TYPE VARCHAR(255), ALTER COLUMN \"name\" SET DEFAULT 'Anonymous'",
			},
		},
		{
			name: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.String("email", 500).Default(nil).Change()
			},
			want: []string{
				"ALTER TABLE \"users\" ALTER COLUMN \"email\" TYPE VARCHAR(500), ALTER COLUMN \"email\" SET DEFAULT NULL",
			},
		},
		{
			name: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.String("email", 500).Comment("User email address").Change()
			},
			want: []string{
				"ALTER TABLE \"users\" ALTER COLUMN \"email\" TYPE VARCHAR(500)",
				"COMMENT ON COLUMN \"users\".\"email\" IS 'User email address'",
			},
		},
		{
			name: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.String("email", 500).Comment("").Change()
			},
			want: []string{
				"ALTER TABLE \"users\" ALTER COLUMN \"email\" TYPE VARCHAR(500)",
				"COMMENT ON COLUMN \"users\".\"email\" IS ''",
			},
		},
		{
			name: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.String("email", 500).Nullable(false).Change()
			},
			want: []string{
				"ALTER TABLE \"users\" ALTER COLUMN \"email\" TYPE VARCHAR(500), ALTER COLUMN \"email\" SET NOT NULL",
			},
		},
		{
			name: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.String("", 255).Change() // Intentionally empty column name
			},
			wantErr: true,
		},
		{
			name:      "No changes",
			table:     "users",
			blueprint: func(_ *blueprint.Blueprint) {},
			wantErr:   false,
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
			got, err := bp.ToSQL()
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompileDrop(t *testing.T) {
	g, err := grammars.NewGrammar("postgres")
	require.NoError(t, err)

	tests := []struct {
		name    string
		table   string
		want    string
		wantErr bool
	}{
		{
			name:    "Drop table",
			table:   "users",
			want:    "DROP TABLE \"users\"",
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
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompileDropIfExists(t *testing.T) {
	g, err := grammars.NewGrammar("postgres")
	require.NoError(t, err)

	tests := []struct {
		name    string
		table   string
		want    string
		wantErr bool
	}{
		{
			name:    "Drop table if exists",
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
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompileRename(t *testing.T) {
	g, err := grammars.NewGrammar("postgres")
	require.NoError(t, err)

	tests := []struct {
		name    string
		oldName string
		newName string
		want    string
		wantErr bool
	}{
		{
			name:    "Rename table",
			oldName: "users",
			newName: "people",
			want:    "ALTER TABLE \"users\" RENAME TO \"people\"",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &blueprint.Blueprint{Name: tt.oldName, Grammar: g}
			bp.Rename(tt.newName)
			got, err := g.CompileRename(bp, bp.Commands[0])
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompileDropColumn(t *testing.T) {
	g, err := grammars.NewGrammar("postgres")
	require.NoError(t, err)

	tests := []struct {
		name      string
		table     string
		blueprint func(table *blueprint.Blueprint)
		wants     []string
		wantErr   bool
	}{
		{
			name: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.DropColumn("email")
			},
			wants:   []string{"ALTER TABLE \"users\" DROP COLUMN \"email\""},
			wantErr: false,
		},
		{
			name: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.DropColumn("email", "phone")
				table.DropColumn("address")
			},
			wants: []string{
				"ALTER TABLE \"users\" DROP COLUMN \"email\", DROP COLUMN \"phone\"",
				"ALTER TABLE \"users\" DROP COLUMN \"address\"",
			},
			wantErr: false,
		},
		{
			name:      "No columns to drop",
			table:     "users",
			blueprint: func(_ *blueprint.Blueprint) {},
			wants:     nil,
			wantErr:   false,
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
			got, err := bp.ToSQL()
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wants, got)
		})
	}
}

func TestPgGrammar_CompileRenameColumn(t *testing.T) {
	g, err := grammars.NewGrammar("postgres")
	require.NoError(t, err)

	tests := []struct {
		name    string
		table   string
		oldName string
		newName string
		want    string
		wantErr bool
	}{
		{
			name:    "Rename column",
			table:   "users",
			oldName: "email",
			newName: "user_email",
			want:    "ALTER TABLE \"users\" RENAME COLUMN \"email\" TO \"user_email\"",
			wantErr: false,
		},
		{
			name:    "Empty old name",
			table:   "users",
			oldName: "",
			newName: "user_email",
			wantErr: true,
		},
		{
			name:    "Empty new name",
			table:   "users",
			oldName: "email",
			newName: "",
			wantErr: true,
		},
		{
			name:    "Both names empty",
			table:   "users",
			oldName: "",
			newName: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tableName := tt.table
			if tableName == "" && !tt.wantErr {
				tableName = tt.name
			}
			bp := &blueprint.Blueprint{Name: tableName, Grammar: g}
			command := &blueprint.Command{From: tt.oldName, To: tt.newName}
			got, err := g.CompileRenameColumn(bp, command)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompileDropIndex(t *testing.T) {
	g, err := grammars.NewGrammar("postgres")
	require.NoError(t, err)

	tests := []struct {
		name      string
		indexName string
		want      string
		wantErr   bool
	}{
		{
			name:      "Drop index with valid name",
			indexName: "users_email_index",
			want:      "DROP INDEX \"users_email_index\"",
			wantErr:   false,
		},
		{
			name:      "Drop index with complex name",
			indexName: "idx_users_email_name",
			want:      "DROP INDEX \"idx_users_email_name\"",
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
			bp := &blueprint.Blueprint{Grammar: g}
			command := &blueprint.Command{Index: tt.indexName}
			got, err := g.CompileDropIndex(bp, command)
			if tt.wantErr {
				require.Error(t, err)
				assert.Empty(t, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompileDropPrimary(t *testing.T) {
	g, err := grammars.NewGrammar("postgres")
	require.NoError(t, err)

	tests := []struct {
		name      string
		blueprint *blueprint.Blueprint
		indexName string
		want      string
		wantErr   bool
	}{
		{
			name: "Drop primary key with specified name",
			blueprint: func() *blueprint.Blueprint {
				return &blueprint.Blueprint{Name: "users", Grammar: g}
			}(),
			indexName: "users_pkey",
			want:      "ALTER TABLE \"users\" DROP CONSTRAINT \"users_pkey\"",
			wantErr:   false,
		},
		{
			name: "Drop primary key with empty index name (should use default naming)",
			blueprint: func() *blueprint.Blueprint {
				return &blueprint.Blueprint{Name: "posts", Grammar: g}
			}(),
			indexName: "",
			want:      "ALTER TABLE \"posts\" DROP CONSTRAINT \"posts_primary\"",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command := &blueprint.Command{Index: tt.indexName}
			got, err := g.CompileDropPrimary(tt.blueprint, command)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompileRenameIndex(t *testing.T) {
	g, err := grammars.NewGrammar("postgres")
	require.NoError(t, err)

	tests := []struct {
		name    string
		table   string
		oldName string
		newName string
		want    string
		wantErr bool
	}{
		{
			name:    "Rename index with valid names",
			table:   "users",
			oldName: "users_email_index",
			newName: "users_email_unique",
			want:    "ALTER INDEX \"users_email_index\" RENAME TO \"users_email_unique\"",
			wantErr: false,
		},
		{
			name:    "Rename index with complex names",
			table:   "users",
			oldName: "idx_users_email_name",
			newName: "idx_users_email_name_unique",
			want:    "ALTER INDEX \"idx_users_email_name\" RENAME TO \"idx_users_email_name_unique\"",
			wantErr: false,
		},
		{
			name:    "Empty old name",
			table:   "users",
			oldName: "",
			newName: "users_email_unique",
			wantErr: true,
		},
		{
			name:    "Empty new name",
			table:   "users",
			oldName: "users_email_index",
			newName: "",
			wantErr: true,
		},
		{
			name:    "Both names empty",
			table:   "users",
			oldName: "",
			newName: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tableName := tt.table
			if tableName == "" && !tt.wantErr {
				tableName = tt.name
			}
			bp := &blueprint.Blueprint{Name: tableName, Grammar: g}
			command := &blueprint.Command{From: tt.oldName, To: tt.newName}
			got, err := g.CompileRenameIndex(bp, command)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompileForeign(t *testing.T) {
	g, err := grammars.NewGrammar("postgres")
	require.NoError(t, err)

	tests := []struct {
		name      string
		table     string
		blueprint func(table *blueprint.Blueprint)
		want      string
		wantErr   bool
	}{
		{
			name: "posts",
			blueprint: func(table *blueprint.Blueprint) {
				table.Foreign("user_id").References("id").On("users")
			},
			want:    "ALTER TABLE \"posts\" ADD CONSTRAINT \"posts_user_id_foreign\" FOREIGN KEY (\"user_id\") REFERENCES \"users\" (\"id\")",
			wantErr: false,
		},
		{
			name: "orders",
			blueprint: func(table *blueprint.Blueprint) {
				table.Foreign("customer_id").References("id").On("customers").Name("fk_orders_customers")
			},
			want:    "ALTER TABLE \"orders\" ADD CONSTRAINT \"fk_orders_customers\" FOREIGN KEY (\"customer_id\") REFERENCES \"customers\" (\"id\")",
			wantErr: false,
		},
		{
			name: "comments",
			blueprint: func(table *blueprint.Blueprint) {
				table.Foreign("post_id").References("id").On("posts").CascadeOnDelete()
			},
			want:    "ALTER TABLE \"comments\" ADD CONSTRAINT \"comments_post_id_foreign\" FOREIGN KEY (\"post_id\") REFERENCES \"posts\" (\"id\") ON DELETE CASCADE",
			wantErr: false,
		},
		{
			name: "orders",
			blueprint: func(table *blueprint.Blueprint) {
				table.Foreign("customer_id").References("id").On("customers").NullOnUpdate()
			},
			want:    "ALTER TABLE \"orders\" ADD CONSTRAINT \"orders_customer_id_foreign\" FOREIGN KEY (\"customer_id\") REFERENCES \"customers\" (\"id\") ON UPDATE SET NULL",
			wantErr: false,
		},
		{
			name: "order_items",
			blueprint: func(table *blueprint.Blueprint) {
				table.Foreign("order_id").References("id").On("orders").CascadeOnDelete().RestrictOnUpdate()
			},
			want:    "ALTER TABLE \"order_items\" ADD CONSTRAINT \"order_items_order_id_foreign\" FOREIGN KEY (\"order_id\") REFERENCES \"orders\" (\"id\") ON DELETE CASCADE ON UPDATE RESTRICT",
			wantErr: false,
		},
		{
			name: "posts",
			blueprint: func(table *blueprint.Blueprint) {
				table.Foreign("user_id").References("id").On("users").Deferrable(true)
			},
			want:    "ALTER TABLE \"posts\" ADD CONSTRAINT \"posts_user_id_foreign\" FOREIGN KEY (\"user_id\") REFERENCES \"users\" (\"id\") DEFERRABLE",
			wantErr: false,
		},
		{
			name: "posts",
			blueprint: func(table *blueprint.Blueprint) {
				table.Foreign("user_id").References("id").On("users").Deferrable(false)
			},
			want:    "ALTER TABLE \"posts\" ADD CONSTRAINT \"posts_user_id_foreign\" FOREIGN KEY (\"user_id\") REFERENCES \"users\" (\"id\") NOT DEFERRABLE",
			wantErr: false,
		},
		{
			name: "posts",
			blueprint: func(table *blueprint.Blueprint) {
				table.Foreign("user_id").References("id").On("users").Deferrable().InitiallyImmediate(true)
			},
			want:    "ALTER TABLE \"posts\" ADD CONSTRAINT \"posts_user_id_foreign\" FOREIGN KEY (\"user_id\") REFERENCES \"users\" (\"id\") DEFERRABLE INITIALLY IMMEDIATE",
			wantErr: false,
		},
		{
			name: "posts",
			blueprint: func(table *blueprint.Blueprint) {
				table.Foreign("user_id").References("id").On("users").Deferrable().InitiallyImmediate(false)
			},
			want:    "ALTER TABLE \"posts\" ADD CONSTRAINT \"posts_user_id_foreign\" FOREIGN KEY (\"user_id\") REFERENCES \"users\" (\"id\") DEFERRABLE INITIALLY DEFERRED",
			wantErr: false,
		},
		{
			name: "posts",
			blueprint: func(table *blueprint.Blueprint) {
				table.Foreign("user_id").References("id").On("users").Deferrable(false).InitiallyImmediate(true)
			},
			want:    "ALTER TABLE \"posts\" ADD CONSTRAINT \"posts_user_id_foreign\" FOREIGN KEY (\"user_id\") REFERENCES \"users\" (\"id\") NOT DEFERRABLE",
			wantErr: false,
		},
		{
			name: "user_roles",
			blueprint: func(table *blueprint.Blueprint) {
				table.Foreign("role_id").References("id").On("roles").
					CascadeOnDelete().RestrictOnUpdate().
					Deferrable().InitiallyImmediate(true)
			},
			want:    "ALTER TABLE \"user_roles\" ADD CONSTRAINT \"user_roles_role_id_foreign\" FOREIGN KEY (\"role_id\") REFERENCES \"roles\" (\"id\") ON DELETE CASCADE ON UPDATE RESTRICT DEFERRABLE INITIALLY IMMEDIATE",
			wantErr: false,
		},
		{
			name: "posts",
			blueprint: func(table *blueprint.Blueprint) {
				table.Foreign("").References("id").On("users")
			},
			wantErr: true,
		},
		{
			name: "posts",
			blueprint: func(table *blueprint.Blueprint) {
				table.Foreign("user_id").References("id").On("")
			},
			wantErr: true,
		},
		{
			name: "posts",
			blueprint: func(table *blueprint.Blueprint) {
				table.Foreign("user_id").References("").On("users")
			},
			wantErr: true,
		},
		{
			name: "posts",
			blueprint: func(table *blueprint.Blueprint) {
				table.Foreign("").References("").On("")
			},
			wantErr: true,
		},
		{
			name: "invoices",
			blueprint: func(table *blueprint.Blueprint) {
				table.Foreign("customer_id").References("id").On("customers").RestrictOnDelete().RestrictOnUpdate()
			},
			want:    "ALTER TABLE \"invoices\" ADD CONSTRAINT \"invoices_customer_id_foreign\" FOREIGN KEY (\"customer_id\") REFERENCES \"customers\" (\"id\") ON DELETE RESTRICT ON UPDATE RESTRICT",
			wantErr: false,
		},
		{
			name: "payments",
			blueprint: func(table *blueprint.Blueprint) {
				table.Foreign("invoice_id").References("id").On("invoices").NoActionOnDelete().NoActionOnUpdate()
			},
			want: "ALTER TABLE \"payments\" ADD CONSTRAINT \"payments_invoice_id_foreign\" FOREIGN KEY (\"invoice_id\") REFERENCES \"invoices\" (\"id\") ON DELETE NO ACTION ON UPDATE NO ACTION",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tableName := tt.table
			if tableName == "" && !tt.wantErr {
				tableName = tt.name
			}
			bp := &blueprint.Blueprint{Name: tableName}
			tt.blueprint(bp)
			got, err := g.CompileForeign(bp, bp.Commands[0])
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompileDropForeign(t *testing.T) {
	g, err := grammars.NewGrammar("postgres")
	require.NoError(t, err)

	tests := []struct {
		name           string
		table          string
		foreignKeyName string
		want           string
		wantErr        bool
	}{
		{
			name:           "Drop foreign key with valid name",
			table:          "posts",
			foreignKeyName: "fk_posts_users",
			want:           "ALTER TABLE \"posts\" DROP CONSTRAINT \"fk_posts_users\"",
			wantErr:        false,
		},
		{
			name:           "Drop foreign key with complex name",
			table:          "order_items",
			foreignKeyName: "fk_order_items_products_cascade",
			want:           "ALTER TABLE \"order_items\" DROP CONSTRAINT \"fk_order_items_products_cascade\"",
			wantErr:        false,
		},
		{
			name:           "Empty foreign key name",
			table:          "users",
			foreignKeyName: "",
			want:           "",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tableName := tt.table
			if tableName == "" && !tt.wantErr {
				tableName = tt.name
			}
			bp := &blueprint.Blueprint{Name: tableName}
			command := &blueprint.Command{Index: tt.foreignKeyName}
			got, err := g.CompileDropForeign(bp, command)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompileIndex(t *testing.T) {
	g, err := grammars.NewGrammar("postgres")
	require.NoError(t, err)

	tests := []struct {
		name      string
		table     string
		blueprint func(table *blueprint.Blueprint)
		want      string
		wantErr   bool
	}{
		{
			name: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Index("email").Name("users_email_index")
			},
			want:    "CREATE INDEX \"users_email_index\" ON \"users\" (\"email\")",
			wantErr: false,
		},
		{
			name: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Index("name", "email").Name("users_name_email_index")
			},
			want:    "CREATE INDEX \"users_name_email_index\" ON \"users\" (\"name\", \"email\")",
			wantErr: false,
		},
		{
			name: "products",
			blueprint: func(table *blueprint.Blueprint) {
				table.Index("sku").Name("products_sku_index").Algorithm("btree")
			},
			want:    "CREATE INDEX \"products_sku_index\" ON \"products\" USING btree (\"sku\")",
			wantErr: false,
		},
		{
			name: "orders",
			blueprint: func(table *blueprint.Blueprint) {
				table.Index("sku")
			},
			want:    "CREATE INDEX \"orders_sku_index\" ON \"orders\" (\"sku\")",
			wantErr: false,
		},
		{
			name: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Index("name", "", "email").Name("users_invalid_index")
			},
			wantErr: true,
		},
		{
			name: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Index("").Name("users_empty_index")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tableName := tt.table
			if tableName == "" && !tt.wantErr {
				tableName = tt.name
			}
			bp := &blueprint.Blueprint{Name: tableName}
			tt.blueprint(bp)
			got, err := g.CompileIndex(bp, bp.Commands[0])
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompileUnique(t *testing.T) {
	g, err := grammars.NewGrammar("postgres")
	require.NoError(t, err)

	tests := []struct {
		name      string
		table     string
		blueprint func(table *blueprint.Blueprint)
		want      string
		wantErr   bool
	}{
		{
			name: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Unique("email").Name("users_email_unique")
			},
			want:    "ALTER TABLE \"users\" ADD CONSTRAINT \"users_email_unique\" UNIQUE (\"email\")",
			wantErr: false,
		},
		{
			name: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Unique("name", "email").Name("users_name_email_unique")
			},
			want:    "ALTER TABLE \"users\" ADD CONSTRAINT \"users_name_email_unique\" UNIQUE (\"name\", \"email\")",
			wantErr: false,
		},
		{
			name: "orders",
			blueprint: func(table *blueprint.Blueprint) {
				table.Unique("order_number")
			},
			want:    "ALTER TABLE \"orders\" ADD CONSTRAINT \"orders_order_number_unique\" UNIQUE (\"order_number\")",
			wantErr: false,
		},
		{
			name: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Unique("email").Name("users_email_unique").Deferrable(true)
			},
			want:    "ALTER TABLE \"users\" ADD CONSTRAINT \"users_email_unique\" UNIQUE (\"email\") DEFERRABLE",
			wantErr: false,
		},
		{
			name: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Unique("email").Name("users_email_unique").Deferrable(false)
			},
			want:    "ALTER TABLE \"users\" ADD CONSTRAINT \"users_email_unique\" UNIQUE (\"email\") NOT DEFERRABLE",
			wantErr: false,
		},
		{
			name: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Unique("email").Name("users_email_unique").Deferrable().InitiallyImmediate(true)
			},
			want:    "ALTER TABLE \"users\" ADD CONSTRAINT \"users_email_unique\" UNIQUE (\"email\") DEFERRABLE INITIALLY IMMEDIATE",
			wantErr: false,
		},
		{
			name: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Unique("email").Name("users_email_unique").Deferrable().InitiallyImmediate(false)
			},
			want:    "ALTER TABLE \"users\" ADD CONSTRAINT \"users_email_unique\" UNIQUE (\"email\") DEFERRABLE INITIALLY DEFERRED",
			wantErr: false,
		},
		{
			name: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Unique("email").Name("users_email_unique").Deferrable(false).InitiallyImmediate(true)
			},
			want:    "ALTER TABLE \"users\" ADD CONSTRAINT \"users_email_unique\" UNIQUE (\"email\") NOT DEFERRABLE",
			wantErr: false,
		},
		{
			name: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Unique("name", "", "email").Name("users_invalid_unique")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tableName := tt.table
			if tableName == "" && !tt.wantErr {
				tableName = tt.name
			}
			bp := &blueprint.Blueprint{Name: tableName}
			tt.blueprint(bp)
			got, err := g.CompileUnique(bp, bp.Commands[0])
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompileFullText(t *testing.T) {
	g, err := grammars.NewGrammar("postgres")
	require.NoError(t, err)

	tests := []struct {
		name      string
		table     string
		blueprint func(table *blueprint.Blueprint)
		want      string
		wantErr   bool
	}{
		{
			name: "articles",
			blueprint: func(table *blueprint.Blueprint) {
				table.FullText("title").Name("articles_title_fulltext").Language("english")
			},
			want:    "CREATE INDEX \"articles_title_fulltext\" ON \"articles\" USING GIN (to_tsvector('english', title))",
			wantErr: false,
		},
		{
			name: "documents",
			blueprint: func(table *blueprint.Blueprint) {
				table.FullText("title", "content").Name("documents_title_content_fulltext").Language("english")
			},
			want:    "CREATE INDEX \"documents_title_content_fulltext\" ON \"documents\" USING GIN (to_tsvector('english', title) || to_tsvector('english', content))",
			wantErr: false,
		},
		{
			name: "posts",
			blueprint: func(table *blueprint.Blueprint) {
				table.FullText("content").Name("posts_content_spanish_fulltext").Language("spanish")
			},
			want:    "CREATE INDEX \"posts_content_spanish_fulltext\" ON \"posts\" USING GIN (to_tsvector('spanish', content))",
			wantErr: false,
		},
		{
			name: "blogs",
			blueprint: func(table *blueprint.Blueprint) {
				table.FullText("body").Name("blogs_body_fulltext")
			},
			want:    "CREATE INDEX \"blogs_body_fulltext\" ON \"blogs\" USING GIN (to_tsvector('english', body))",
			wantErr: false,
		},
		{
			name: "news",
			blueprint: func(table *blueprint.Blueprint) {
				table.FullText("headline")
			},
			want:    "CREATE INDEX \"news_headline_fulltext\" ON \"news\" USING GIN (to_tsvector('english', headline))",
			wantErr: false,
		},
		{
			name: "products",
			blueprint: func(table *blueprint.Blueprint) {
				table.FullText("name", "description", "tags").Name("products_search_fulltext").Language("english")
			},
			want:    "CREATE INDEX \"products_search_fulltext\" ON \"products\" USING GIN (to_tsvector('english', name) || to_tsvector('english', description) || to_tsvector('english', tags))",
			wantErr: false,
		},
		{
			name: "articles",
			blueprint: func(table *blueprint.Blueprint) {
				table.FullText("title", "", "content").Name("articles_invalid_fulltext")
			},
			wantErr: true,
		},
		{
			name: "articles",
			blueprint: func(table *blueprint.Blueprint) {
				table.FullText("").Name("articles_empty_fulltext")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tableName := tt.table
			if tableName == "" && !tt.wantErr {
				tableName = tt.name
			}
			bp := &blueprint.Blueprint{Name: tableName}
			tt.blueprint(bp)
			got, err := g.CompileFullText(bp, bp.Commands[0])
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompileDropUnique(t *testing.T) {
	g, err := grammars.NewGrammar("postgres")
	require.NoError(t, err)

	tests := []struct {
		name      string
		indexName string
		want      string
		wantErr   bool
	}{
		{
			name:      "Drop unique index with valid name",
			indexName: "users_email_unique",
			want:      "ALTER TABLE  DROP CONSTRAINT \"users_email_unique\"",
			wantErr:   false,
		},
		{
			name:      "Drop unique index with complex name",
			indexName: "uk_users_email_name",
			want:      "ALTER TABLE  DROP CONSTRAINT \"uk_users_email_name\"",
			wantErr:   false,
		},
		{
			name:      "Drop unique index with numeric suffix",
			indexName: "users_email_unique_2",
			want:      "ALTER TABLE  DROP CONSTRAINT \"users_email_unique_2\"",
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
			bp := &blueprint.Blueprint{}
			command := &blueprint.Command{Index: tt.indexName}
			got, err := g.CompileDropUnique(bp, command)
			if tt.wantErr {
				require.Error(t, err)
				assert.Empty(t, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompileDropFulltext(t *testing.T) {
	g, err := grammars.NewGrammar("postgres")
	require.NoError(t, err)

	tests := []struct {
		name      string
		indexName string
		want      string
		wantErr   bool
	}{
		{
			name:      "Drop fulltext index with valid name",
			indexName: "articles_title_fulltext",
			want:      "DROP INDEX \"articles_title_fulltext\"",
			wantErr:   false,
		},
		{
			name:      "Drop fulltext index with complex name",
			indexName: "documents_title_content_fulltext",
			want:      "DROP INDEX \"documents_title_content_fulltext\"",
			wantErr:   false,
		},
		{
			name:      "Drop fulltext index with underscore prefix",
			indexName: "idx_posts_content_fulltext",
			want:      "DROP INDEX \"idx_posts_content_fulltext\"",
			wantErr:   false,
		},
		{
			name:      "Drop fulltext index with numeric suffix",
			indexName: "search_index_fulltext_1",
			want:      "DROP INDEX \"search_index_fulltext_1\"",
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
			bp := &blueprint.Blueprint{}
			command := &blueprint.Command{Index: tt.indexName}
			got, err := g.CompileDropFulltext(bp, command)
			if tt.wantErr {
				require.Error(t, err)
				assert.Empty(t, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompilePrimary(t *testing.T) {
	g, err := grammars.NewGrammar("postgres")
	require.NoError(t, err)

	tests := []struct {
		name      string
		table     string
		blueprint func(table *blueprint.Blueprint)
		want      string
		wantErr   bool
	}{
		{
			name: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Primary("id").Name("users_id_primary")
			},
			want:    "ALTER TABLE \"users\" ADD CONSTRAINT \"users_id_primary\" PRIMARY KEY (\"id\")",
			wantErr: false,
		},
		{
			name: "user_roles",
			blueprint: func(table *blueprint.Blueprint) {
				table.Primary("user_id", "role_id").Name("user_roles_primary")
			},
			want:    "ALTER TABLE \"user_roles\" ADD CONSTRAINT \"user_roles_primary\" PRIMARY KEY (\"user_id\", \"role_id\")",
			wantErr: false,
		},
		{
			name: "orders",
			blueprint: func(table *blueprint.Blueprint) {
				table.Primary("order_id")
			},
			want:    "ALTER TABLE \"orders\" ADD CONSTRAINT \"orders_order_id_primary\" PRIMARY KEY (\"order_id\")",
			wantErr: false,
		},
		{
			name: "order_items",
			blueprint: func(table *blueprint.Blueprint) {
				table.Primary("order_id", "product_id", "variant_id").Name("order_items_composite_pk")
			},
			want:    "ALTER TABLE \"order_items\" ADD CONSTRAINT \"order_items_composite_pk\" PRIMARY KEY (\"order_id\", \"product_id\", \"variant_id\")",
			wantErr: false,
		},
		{
			name: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Primary("id", "", "tenant_id").Name("users_invalid_primary")
			},
			wantErr: true,
		},
		{
			name: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Primary("").Name("users_empty_primary")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tableName := tt.table
			if tableName == "" && !tt.wantErr {
				tableName = tt.name
			}
			bp := &blueprint.Blueprint{Name: tableName}
			tt.blueprint(bp)
			got, err := g.CompilePrimary(bp, bp.Commands[0])
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_GetType(t *testing.T) {
	g, err := grammars.NewGrammar("postgres")
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
			name: "boolean column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Boolean("active")
			},
			want: "BOOLEAN",
		},
		{
			name: "char column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Char("code", 10)
			},
			want: "CHAR(10)",
		},
		{
			name: "char column type without length",
			blueprint: func(table *blueprint.Blueprint) {
				table.Char("code")
			},
			want: "CHAR(255)",
		},
		{
			name: "string column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.String("name", 255)
			},
			want: "VARCHAR(255)",
		},
		{
			name: "decimal column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Decimal("price", 10, 2)
			},
			want: "DECIMAL(10, 2)",
		},
		{
			name: "double column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Double("value")
			},
			want: "DOUBLE PRECISION",
		},
		{
			name: "float column type with precision",
			blueprint: func(table *blueprint.Blueprint) {
				table.Float("value", 6, 2)
			},
			want: "REAL",
		},
		{
			name: "float column type without precision",
			blueprint: func(table *blueprint.Blueprint) {
				table.Float("value", 0, 0)
			},
			want: "REAL",
		},
		{
			name: "big integer column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.BigInteger("id")
			},
			want: "BIGINT",
		},
		{
			name: "big integer auto increment",
			blueprint: func(table *blueprint.Blueprint) {
				table.BigInteger("id").AutoIncrement()
			},
			want: "BIGSERIAL",
		},
		{
			name: "integer column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Integer("count")
			},
			want: "INTEGER",
		},
		{
			name: "integer auto increment",
			blueprint: func(table *blueprint.Blueprint) {
				table.Integer("id").AutoIncrement()
			},
			want: "SERIAL",
		},
		{
			name: "small integer column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.SmallInteger("status")
			},
			want: "SMALLINT",
		},
		{
			name: "small integer auto increment",
			blueprint: func(table *blueprint.Blueprint) {
				table.SmallInteger("id").Unsigned().AutoIncrement()
			},
			want: "SMALLSERIAL",
		},
		{
			name: "medium integer column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.MediumInteger("value")
			},
			want: "INTEGER",
		},
		{
			name: "medium auto increment",
			blueprint: func(table *blueprint.Blueprint) {
				table.MediumIncrements("id")
			},
			want: "SERIAL",
		},
		{
			name: "tiny integer column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.TinyInteger("flag")
			},
			want: "SMALLINT",
		},
		{
			name: "tiny integer auto increment",
			blueprint: func(table *blueprint.Blueprint) {
				table.TinyInteger("id").AutoIncrement()
			},
			want: "SMALLSERIAL",
		},
		{
			name: "time column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Time("created_at")
			},
			want: "TIME(0)",
		},
		{
			name: "datetime column type with precision",
			blueprint: func(table *blueprint.Blueprint) {
				table.DateTime("created_at", 6)
			},
			want: "TIMESTAMP(6)",
		},
		{
			name: "datetime column type without precision",
			blueprint: func(table *blueprint.Blueprint) {
				table.DateTime("created_at", 0)
			},
			want: "TIMESTAMP(0)",
		},
		{
			name: "datetime tz column type with precision",
			blueprint: func(table *blueprint.Blueprint) {
				table.DateTimeTz("created_at", 3)
			},
			want: "TIMESTAMPTZ(3)",
		},
		{
			name: "datetime tz column type without precision",
			blueprint: func(table *blueprint.Blueprint) {
				table.DateTimeTz("created_at", 0)
			},
			want: "TIMESTAMPTZ(0)",
		},
		{
			name: "timestamp column type with precision",
			blueprint: func(table *blueprint.Blueprint) {
				table.Timestamp("created_at", 6)
			},
			want: "TIMESTAMP(6)",
		},
		{
			name: "timestamp column type without precision",
			blueprint: func(table *blueprint.Blueprint) {
				table.Timestamp("created_at", 0)
			},
			want: "TIMESTAMP(0)",
		},
		{
			name: "timestamp tz column type with precision",
			blueprint: func(table *blueprint.Blueprint) {
				table.TimestampTz("created_at", 3)
			},
			want: "TIMESTAMPTZ(3)",
		},
		{
			name: "timestamp tz column type without precision",
			blueprint: func(table *blueprint.Blueprint) {
				table.TimestampTz("created_at", 0)
			},
			want: "TIMESTAMPTZ(0)",
		},
		{
			name: "geography column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Geography("location", "POINT", 4326)
			},
			want: "GEOGRAPHY(POINT, 4326)",
		},
		{
			name: "long text column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.LongText("content")
			},
			want: "TEXT",
		},
		{
			name: "text column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Text("description")
			},
			want: "TEXT",
		},
		{
			name: "tiny text column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.TinyText("notes")
			},
			want: "VARCHAR(255)",
		},
		{
			name: "date column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Date("birth_date")
			},
			want: "DATE",
		},
		{
			name: "year column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Year("graduation_year")
			},
			want: "INTEGER",
		},
		{
			name: "json column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.JSON("metadata")
			},
			want: "JSON",
		},
		{
			name: "jsonb column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.JSONB("data")
			},
			want: "JSONB",
		},
		{
			name: "uuid column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.UUID("uuid")
			},
			want: "UUID",
		},
		{
			name: "binary column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Binary("data")
			},
			want: "BYTEA",
		},
		{
			name: "point column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Point("location")
			},
			want: "POINT(4326)",
		},
		{
			name: "Geography type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Geography("location", "POINT", 4326)
			},
			want: "GEOGRAPHY(POINT, 4326)",
		},
		{
			name: "Geometry type with SRID",
			blueprint: func(table *blueprint.Blueprint) {
				table.Geometry("shape", "POLYGON", 4326)
			},
			want: "GEOMETRY(POLYGON, 4326)",
		},
		{
			name: "Geometry type without SRID",
			blueprint: func(table *blueprint.Blueprint) {
				table.Geometry("shape", "POLYGON")
			},
			want: "GEOMETRY(POLYGON)",
		},
		{
			name: "Enum type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Enum("status", []string{"active", "inactive"})
			},
			want: "VARCHAR(255) CHECK (status IN ('active', 'inactive'))",
		},
		{
			name: "Enum type without allowed values",
			blueprint: func(table *blueprint.Blueprint) {
				table.Enum("status", []string{})
			},
			want: "VARCHAR(255)",
		},
		{
			name: "TimeTz with precision",
			blueprint: func(table *blueprint.Blueprint) {
				table.TimeTz("created_at", 3)
			},
			want: "TIMETZ(3)",
		},
		{
			name: "TimeTz without precision",
			blueprint: func(table *blueprint.Blueprint) {
				table.TimeTz("created_at")
			},
			want: "TIMETZ(0)",
		},
		{
			name: "IP Address type",
			blueprint: func(table *blueprint.Blueprint) {
				table.IPAddress("ip")
			},
			want: "inet",
		},
		{
			name: "Mac Address type",
			blueprint: func(table *blueprint.Blueprint) {
				table.MacAddress("mac")
			},
			want: "MACADDR",
		},
		{
			name: "TSVector type",
			blueprint: func(table *blueprint.Blueprint) {
				table.TSVector("ts")
			},
			want: "TSVECTOR",
		},
		{
			name: "Geography type without subtype",
			blueprint: func(table *blueprint.Blueprint) {
				table.Geography("location", "")
			},
			want: "GEOGRAPHY",
		},
		{
			name: "Point type without srid",
			blueprint: func(table *blueprint.Blueprint) {
				table.Point("location")
			},
			want: "POINT(4326)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &blueprint.Blueprint{Name: "test_table"}
			tt.blueprint(bp)
			got := g.GetType(bp.Columns[0])
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompileCreate_Temporary(t *testing.T) {
	g, err := grammars.NewGrammar("postgres")
	require.NoError(t, err)

	bp := &blueprint.Blueprint{Name: "temp_logs", TemporaryVal: true}
	bp.String("message")
	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)
	assert.Contains(t, sql, "CREATE TEMPORARY TABLE")
	assert.Contains(t, sql, `"temp_logs"`)
}

func TestPgGrammar_CompileVectorIndex(t *testing.T) {
	g, err := grammars.NewGrammar("postgres")
	require.NoError(t, err)

	tests := []struct {
		name    string
		setup   func(*blueprint.Blueprint)
		want    string
		wantErr bool
	}{
		{
			name: "basic vector index uses cosine ops",
			setup: func(bp *blueprint.Blueprint) {
				bp.VectorIndex("embedding")
			},
			want: `CREATE INDEX "items_embedding_vectorindex" ON "items" USING hnsw ("embedding" vector_cosine_ops)`,
		},
		{
			name: "vector index with custom name",
			setup: func(bp *blueprint.Blueprint) {
				bp.VectorIndex("embedding").Name("idx_emb")
			},
			want: `CREATE INDEX "idx_emb" ON "items" USING hnsw ("embedding" vector_cosine_ops)`,
		},
		{
			name:    "empty column returns error",
			setup:   func(bp *blueprint.Blueprint) { bp.VectorIndex("") },
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &blueprint.Blueprint{Name: "items"}
			tt.setup(bp)
			got, err := g.CompileVectorIndex(bp, bp.Commands[0])
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompileTableComment(t *testing.T) {
	g, err := grammars.NewGrammar("postgres")
	require.NoError(t, err)

	t.Run("alter table comment", func(t *testing.T) {
		bp := &blueprint.Blueprint{Name: "users"}
		bp.Comment("User accounts")
		got, err := g.CompileTableComment(bp, bp.Commands[0])
		require.NoError(t, err)
		assert.Equal(t, `COMMENT ON TABLE "users" IS 'User accounts'`, got)
	})

	t.Run("skip comment during create", func(t *testing.T) {
		bp := &blueprint.Blueprint{Name: "users"}
		bp.Create()
		bp.Comment("User accounts")
		var cmd *blueprint.Command
		for _, c := range bp.Commands {
			if c.Name == blueprint.CommandTableComment {
				cmd = c
			}
		}
		require.NotNil(t, cmd)
		got, err := g.CompileTableComment(bp, cmd)
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestPgGrammar_GetType_RawColumn(t *testing.T) {
	g, err := grammars.NewGrammar("postgres")
	require.NoError(t, err)

	bp := &blueprint.Blueprint{Name: "t"}
	bp.RawColumn("payload", "BYTEA NOT NULL")
	got := g.GetType(bp.Columns[0])
	assert.Equal(t, "BYTEA NOT NULL", got)
}
