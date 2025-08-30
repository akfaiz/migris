package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPgGrammar_CompileCreate(t *testing.T) {
	grammar := newPostgresGrammar()

	tests := []struct {
		name      string
		table     string
		blueprint func(table *Blueprint)
		want      string
		wantErr   bool
	}{
		{
			name:  "Create simple table",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.ID()
				table.String("name")
				table.String("email")
				table.String("password").Nullable()
				table.Timestamp("created_at").UseCurrent()
				table.Timestamp("updated_at").UseCurrent()
			},
			want: "CREATE TABLE users (id BIGSERIAL NOT NULL, name VARCHAR(255) NOT NULL, email VARCHAR(255) NOT NULL, password VARCHAR(255) NULL, created_at TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP NOT NULL, updated_at TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP NOT NULL, CONSTRAINT pk_users PRIMARY KEY (id))",
		},
		{
			name:  "Create table with foreign key",
			table: "posts",
			blueprint: func(table *Blueprint) {
				table.ID()
				table.Integer("user_id")
				table.String("title")
				table.Text("content").Nullable()
				table.Foreign("user_id").References("id").On("users").OnDelete("CASCADE").OnUpdate("CASCADE")
			},
			want: "CREATE TABLE posts (id BIGSERIAL NOT NULL, user_id INTEGER NOT NULL, title VARCHAR(255) NOT NULL, content TEXT NULL, CONSTRAINT pk_posts PRIMARY KEY (id))",
		},
		{
			name:  "Create table with column name is empty",
			table: "empty_column_table",
			blueprint: func(table *Blueprint) {
				table.String("") // Intentionally empty column name
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: tt.table}
			tt.blueprint(bp)
			got, err := grammar.CompileCreate(bp)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got, "SQL statement mismatch for %s", tt.name)
		})
	}
}

func TestPgGrammar_CompileAdd(t *testing.T) {
	grammar := newPostgresGrammar()

	tests := []struct {
		name      string
		table     string
		blueprint func(table *Blueprint)
		want      string
		wantErr   bool
	}{
		{
			name:  "Add single column",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.String("phone", 20)
			},
			want:    "ALTER TABLE users ADD COLUMN phone VARCHAR(20) NOT NULL",
			wantErr: false,
		},
		{
			name:  "Add multiple columns",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.String("phone", 20)
				table.String("address", 255).Nullable()
				table.Integer("age")
			},
			want:    "ALTER TABLE users ADD COLUMN phone VARCHAR(20) NOT NULL, ADD COLUMN address VARCHAR(255) NULL, ADD COLUMN age INTEGER NOT NULL",
			wantErr: false,
		},
		{
			name:  "Add column with default value",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Boolean("active").Default(true)
			},
			want:    "ALTER TABLE users ADD COLUMN active BOOLEAN DEFAULT '1' NOT NULL",
			wantErr: false,
		},
		{
			name:  "Add column with comment",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.String("notes", 500).Comment("User notes")
			},
			want:    "ALTER TABLE users ADD COLUMN notes VARCHAR(500) NOT NULL",
			wantErr: false,
		},
		{
			name:  "Add primary key column",
			table: "categories",
			blueprint: func(table *Blueprint) {
				table.Integer("id").Primary()
			},
			want:    "ALTER TABLE categories ADD COLUMN id INTEGER NOT NULL, ADD CONSTRAINT pk_categories PRIMARY KEY (id)",
			wantErr: false,
		},
		{
			name:  "Add auto increment column",
			table: "logs",
			blueprint: func(table *Blueprint) {
				table.BigInteger("id").AutoIncrement()
			},
			want:    "ALTER TABLE logs ADD COLUMN id BIGSERIAL NOT NULL",
			wantErr: false,
		},
		{
			name:  "Add complex column with all attributes",
			table: "products",
			blueprint: func(table *Blueprint) {
				table.Decimal("price", 10, 2).Default(0)
			},
			want:    "ALTER TABLE products ADD COLUMN price DECIMAL(10, 2) DEFAULT '0' NOT NULL",
			wantErr: false,
		},
		{
			name:  "Add timestamp columns",
			table: "orders",
			blueprint: func(table *Blueprint) {
				table.Timestamp("created_at").UseCurrent()
				table.Timestamp("updated_at").UseCurrent().Nullable()
			},
			want:    "ALTER TABLE orders ADD COLUMN created_at TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP NOT NULL, ADD COLUMN updated_at TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP NULL",
			wantErr: false,
		},
		{
			name:  "Add different data types",
			table: "mixed_table",
			blueprint: func(table *Blueprint) {
				table.Text("description")
				table.JSON("metadata").Nullable()
				table.UUID("reference_id")
				table.Date("event_date")
			},
			want:    "ALTER TABLE mixed_table ADD COLUMN description TEXT NOT NULL, ADD COLUMN metadata JSON NULL, ADD COLUMN reference_id UUID NOT NULL, ADD COLUMN event_date DATE NOT NULL",
			wantErr: false,
		},
		{
			name:      "No columns to add",
			table:     "users",
			blueprint: func(table *Blueprint) {},
			want:      "",
			wantErr:   false,
		},
		{
			name:  "Error on empty column name",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.String("", 255) // Intentionally empty column name
			},
			wantErr: true,
		},
		{
			name:  "Add enum column",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Enum("status", []string{"active", "inactive", "pending"})
			},
			want:    "ALTER TABLE users ADD COLUMN status VARCHAR(255) CHECK (status IN ('active', 'inactive', 'pending')) NOT NULL",
			wantErr: false,
		},
		{
			name:  "Add geography column",
			table: "locations",
			blueprint: func(table *Blueprint) {
				table.Geography("coordinates", "POINT", 4326)
			},
			want:    "ALTER TABLE locations ADD COLUMN coordinates GEOGRAPHY(POINT, 4326) NOT NULL",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: tt.table}
			tt.blueprint(bp)
			got, err := grammar.CompileAdd(bp)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompileChange(t *testing.T) {
	grammar := newPostgresGrammar()

	tests := []struct {
		name      string
		table     string
		blueprint func(blueprint *Blueprint)
		want      []string
		wantErr   bool
	}{
		{
			name:  "Change single column type",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.String("email", 500).Nullable().Change()
			},
			want: []string{"ALTER TABLE users ALTER COLUMN email TYPE VARCHAR(500), ALTER COLUMN email DROP NOT NULL"},
		},
		{
			name:  "Change column with default value",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.String("email", 500).Default("user@mail.com").Change()
			},
			want: []string{"ALTER TABLE users ALTER COLUMN email TYPE VARCHAR(500), ALTER COLUMN email SET DEFAULT 'user@mail.com'"},
		},
		{
			name:  "Change multiple columns",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.String("email", 500).Nullable().Change()
				table.String("name", 255).Default("Anonymous").Change()
			},
			want: []string{
				"ALTER TABLE users ALTER COLUMN email TYPE VARCHAR(500), ALTER COLUMN email DROP NOT NULL",
				"ALTER TABLE users ALTER COLUMN name TYPE VARCHAR(255), ALTER COLUMN name SET DEFAULT 'Anonymous'",
			},
		},
		{
			name:  "Drop default value from column",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.String("email", 500).Default(nil).Change()
			},
			want: []string{"ALTER TABLE users ALTER COLUMN email TYPE VARCHAR(500), ALTER COLUMN email SET DEFAULT NULL"},
		},
		{
			name:  "Add comment to column",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.String("email", 500).Comment("User email address").Change()
			},
			want: []string{
				"ALTER TABLE users ALTER COLUMN email TYPE VARCHAR(500)",
				"COMMENT ON COLUMN users.email IS 'User email address'",
			},
		},
		{
			name:  "Remove comment from column",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.String("email", 500).Comment("").Change()
			},
			want: []string{"ALTER TABLE users ALTER COLUMN email TYPE VARCHAR(500)", "COMMENT ON COLUMN users.email IS ''"},
		},
		{
			name:  "Set column to not nullable",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.String("email", 500).Nullable(false).Change()
			},
			want: []string{"ALTER TABLE users ALTER COLUMN email TYPE VARCHAR(500), ALTER COLUMN email SET NOT NULL"},
		},
		{
			name:  "Column name with empty string",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.String("", 255).Change() // Intentionally empty column name
			},
			wantErr: true,
		},
		{
			name:      "No changes",
			table:     "users",
			blueprint: func(table *Blueprint) {},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: tt.table, grammar: grammar}
			tt.blueprint(bp)
			got, err := bp.toSql()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompileDrop(t *testing.T) {
	grammar := newPostgresGrammar()

	tests := []struct {
		name    string
		table   string
		want    string
		wantErr bool
	}{
		{
			name:    "Drop table",
			table:   "users",
			want:    "DROP TABLE users",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: tt.table}
			got, err := grammar.CompileDrop(bp)
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
	grammar := newPostgresGrammar()

	tests := []struct {
		name    string
		table   string
		want    string
		wantErr bool
	}{
		{
			name:    "Drop table if exists",
			table:   "users",
			want:    "DROP TABLE IF EXISTS users",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: tt.table}
			got, err := grammar.CompileDropIfExists(bp)
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
	grammar := newPostgresGrammar()

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
			want:    "ALTER TABLE users RENAME TO people",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: tt.oldName}
			bp.rename(tt.newName)
			got, err := grammar.CompileRename(bp, bp.commands[0])
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
	grammar := newPostgresGrammar()

	tests := []struct {
		name      string
		blueprint func(table *Blueprint)
		want      []string
		wantErr   bool
	}{
		{
			name: "Simple column",
			blueprint: func(table *Blueprint) {
				table.String("name", 255)
			},
			want: []string{"name VARCHAR(255) NOT NULL"},
		},
		{
			name: "Nullable column",
			blueprint: func(table *Blueprint) {
				table.String("email", 255).Nullable()
			},
			want: []string{"email VARCHAR(255) NULL"},
		},
		{
			name: "Nullable column with default null",
			blueprint: func(table *Blueprint) {
				table.Text("description").Nullable().Default(nil)
			},
			want: []string{"description TEXT DEFAULT NULL NULL"},
		},
		{
			name: "Column with default value",
			blueprint: func(table *Blueprint) {
				table.Boolean("active").Default(true)
			},
			want: []string{"active BOOLEAN DEFAULT '1' NOT NULL"},
		},
		{
			name: "Primary key column",
			blueprint: func(table *Blueprint) {
				table.Integer("id").Primary()
			},
			want: []string{"id INTEGER NOT NULL"},
		},
		{
			name: "Error on empty column",
			blueprint: func(table *Blueprint) {
				table.String("", 255) // Intentionally empty column name
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: "test_table"}
			tt.blueprint(bp)
			got, err := grammar.getColumns(bp)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompileDropColumn(t *testing.T) {
	grammar := newPostgresGrammar()

	tests := []struct {
		name      string
		table     string
		blueprint func(table *Blueprint)
		wants     []string
		wantErr   bool
	}{
		{
			name:  "Drop single column",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.DropColumn("email")
			},
			wants:   []string{"ALTER TABLE users DROP COLUMN email"},
			wantErr: false,
		},
		{
			name:  "Drop multiple columns",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.DropColumn("email", "phone")
				table.DropColumn("address")
			},
			wants:   []string{"ALTER TABLE users DROP COLUMN email, DROP COLUMN phone", "ALTER TABLE users DROP COLUMN address"},
			wantErr: false,
		},
		{
			name:      "No columns to drop",
			table:     "users",
			blueprint: func(table *Blueprint) {},
			wants:     nil,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: tt.table, grammar: grammar}
			tt.blueprint(bp)
			got, err := bp.toSql()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wants, got)
		})
	}
}

func TestPgGrammar_CompileRenameColumn(t *testing.T) {
	grammar := newPostgresGrammar()

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
			want:    "ALTER TABLE users RENAME COLUMN email TO user_email",
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
			bp := &Blueprint{name: tt.table}
			command := &command{from: tt.oldName, to: tt.newName}
			got, err := grammar.CompileRenameColumn(bp, command)
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
	grammar := newPostgresGrammar()

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
			bp := &Blueprint{}
			command := &command{index: tt.indexName}
			got, err := grammar.CompileDropIndex(bp, command)
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

func TestPgGrammar_CompileDropPrimary(t *testing.T) {
	grammar := newPostgresGrammar()

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
			command := &command{index: tt.indexName}
			got, err := grammar.CompileDropPrimary(tt.blueprint, command)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompileRenameIndex(t *testing.T) {
	grammar := newPostgresGrammar()

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
			want:    "ALTER INDEX users_email_index RENAME TO users_email_unique",
			wantErr: false,
		},
		{
			name:    "Rename index with complex names",
			table:   "users",
			oldName: "idx_users_email_name",
			newName: "idx_users_email_name_unique",
			want:    "ALTER INDEX idx_users_email_name RENAME TO idx_users_email_name_unique",
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
			bp := &Blueprint{name: tt.table}
			command := &command{from: tt.oldName, to: tt.newName}
			got, err := grammar.CompileRenameIndex(bp, command)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompileForeign(t *testing.T) {
	grammar := newPostgresGrammar()

	tests := []struct {
		name      string
		table     string
		blueprint func(table *Blueprint)
		want      string
		wantErr   bool
	}{
		{
			name:  "Basic foreign key",
			table: "posts",
			blueprint: func(table *Blueprint) {
				table.Foreign("user_id").References("id").On("users")
			},
			want:    "ALTER TABLE posts ADD CONSTRAINT fk_posts_users FOREIGN KEY (user_id) REFERENCES users(id)",
			wantErr: false,
		},
		{
			name:  "Foreign key with custom constraint name",
			table: "orders",
			blueprint: func(table *Blueprint) {
				table.Foreign("customer_id").References("id").On("customers").Name("fk_orders_customers")
			},
			want:    "ALTER TABLE orders ADD CONSTRAINT fk_orders_customers FOREIGN KEY (customer_id) REFERENCES customers(id)",
			wantErr: false,
		},
		{
			name:  "Foreign key with ON DELETE CASCADE",
			table: "comments",
			blueprint: func(table *Blueprint) {
				table.Foreign("post_id").References("id").On("posts").CascadeOnDelete()
			},
			want:    "ALTER TABLE comments ADD CONSTRAINT fk_comments_posts FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE",
			wantErr: false,
		},
		{
			name:  "Foreign key with ON UPDATE SET NULL",
			table: "orders",
			blueprint: func(table *Blueprint) {
				table.Foreign("customer_id").References("id").On("customers").NullOnUpdate()
			},
			want:    "ALTER TABLE orders ADD CONSTRAINT fk_orders_customers FOREIGN KEY (customer_id) REFERENCES customers(id) ON UPDATE SET NULL",
			wantErr: false,
		},
		{
			name:  "Foreign key with both ON DELETE and ON UPDATE",
			table: "order_items",
			blueprint: func(table *Blueprint) {
				table.Foreign("order_id").References("id").On("orders").CascadeOnDelete().RestrictOnUpdate()
			},
			want:    "ALTER TABLE order_items ADD CONSTRAINT fk_order_items_orders FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE ON UPDATE RESTRICT",
			wantErr: false,
		},
		{
			name:  "Foreign key with deferrable true",
			table: "posts",
			blueprint: func(table *Blueprint) {
				table.Foreign("user_id").References("id").On("users").Deferrable(true)
			},
			want:    "ALTER TABLE posts ADD CONSTRAINT fk_posts_users FOREIGN KEY (user_id) REFERENCES users(id) DEFERRABLE",
			wantErr: false,
		},
		{
			name:  "Foreign key with deferrable false",
			table: "posts",
			blueprint: func(table *Blueprint) {
				table.Foreign("user_id").References("id").On("users").Deferrable(false)
			},
			want:    "ALTER TABLE posts ADD CONSTRAINT fk_posts_users FOREIGN KEY (user_id) REFERENCES users(id) NOT DEFERRABLE",
			wantErr: false,
		},
		{
			name:  "Foreign key with deferrable true and initially immediate true",
			table: "posts",
			blueprint: func(table *Blueprint) {
				table.Foreign("user_id").References("id").On("users").Deferrable().InitiallyImmediate(true)
			},
			want:    "ALTER TABLE posts ADD CONSTRAINT fk_posts_users FOREIGN KEY (user_id) REFERENCES users(id) DEFERRABLE INITIALLY IMMEDIATE",
			wantErr: false,
		},
		{
			name:  "Foreign key with deferrable true and initially immediate false",
			table: "posts",
			blueprint: func(table *Blueprint) {
				table.Foreign("user_id").References("id").On("users").Deferrable().InitiallyImmediate(false)
			},
			want:    "ALTER TABLE posts ADD CONSTRAINT fk_posts_users FOREIGN KEY (user_id) REFERENCES users(id) DEFERRABLE INITIALLY DEFERRED",
			wantErr: false,
		},
		{
			name:  "Foreign key with deferrable false and initially immediate true (should be ignored)",
			table: "posts",
			blueprint: func(table *Blueprint) {
				table.Foreign("user_id").References("id").On("users").Deferrable(false).InitiallyImmediate(true)
			},
			want:    "ALTER TABLE posts ADD CONSTRAINT fk_posts_users FOREIGN KEY (user_id) REFERENCES users(id) NOT DEFERRABLE",
			wantErr: false,
		},
		{
			name:  "Complex foreign key with all options",
			table: "user_roles",
			blueprint: func(table *Blueprint) {
				table.Foreign("role_id").References("id").On("roles").
					CascadeOnDelete().RestrictOnUpdate().
					Deferrable().InitiallyImmediate(true)
			},
			want:    "ALTER TABLE user_roles ADD CONSTRAINT fk_user_roles_roles FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE ON UPDATE RESTRICT DEFERRABLE INITIALLY IMMEDIATE",
			wantErr: false,
		},
		{
			name:  "Empty column name",
			table: "posts",
			blueprint: func(table *Blueprint) {
				table.Foreign("").References("id").On("users")
			},
			wantErr: true,
		},
		{
			name:  "Empty on table name",
			table: "posts",
			blueprint: func(table *Blueprint) {
				table.Foreign("user_id").References("id").On("")
			},
			wantErr: true,
		},
		{
			name:  "Empty references column name",
			table: "posts",
			blueprint: func(table *Blueprint) {
				table.Foreign("user_id").References("").On("users")
			},
			wantErr: true,
		},
		{
			name:  "All required fields empty",
			table: "posts",
			blueprint: func(table *Blueprint) {
				table.Foreign("").References("").On("")
			},
			wantErr: true,
		},
		{
			name:  "Foreign key with RESTRICT actions",
			table: "invoices",
			blueprint: func(table *Blueprint) {
				table.Foreign("customer_id").References("id").On("customers").RestrictOnDelete().RestrictOnUpdate()
			},
			want:    "ALTER TABLE invoices ADD CONSTRAINT fk_invoices_customers FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE RESTRICT ON UPDATE RESTRICT",
			wantErr: false,
		},
		{
			name:  "Foreign key with no actions",
			table: "payments",
			blueprint: func(table *Blueprint) {
				table.Foreign("invoice_id").References("id").On("invoices").NoActionOnDelete().NoActionOnUpdate()
			},
			want: "ALTER TABLE payments ADD CONSTRAINT fk_payments_invoices FOREIGN KEY (invoice_id) REFERENCES invoices(id) ON DELETE NO ACTION ON UPDATE NO ACTION",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: tt.table}
			tt.blueprint(bp)
			got, err := grammar.CompileForeign(bp, bp.commands[0])
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompileDropForeign(t *testing.T) {
	grammar := newPostgresGrammar()

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
			want:           "ALTER TABLE posts DROP CONSTRAINT fk_posts_users",
			wantErr:        false,
		},
		{
			name:           "Drop foreign key with complex name",
			table:          "order_items",
			foreignKeyName: "fk_order_items_products_cascade",
			want:           "ALTER TABLE order_items DROP CONSTRAINT fk_order_items_products_cascade",
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
			bp := &Blueprint{name: tt.table}
			command := &command{index: tt.foreignKeyName}
			got, err := grammar.CompileDropForeign(bp, command)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompileIndex(t *testing.T) {
	grammar := newPostgresGrammar()

	tests := []struct {
		name      string
		table     string
		blueprint func(table *Blueprint)
		want      string
		wantErr   bool
	}{
		{
			name:  "Basic index with single column",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Index("email").Name("users_email_index")
			},
			want:    "CREATE INDEX users_email_index ON users (email)",
			wantErr: false,
		},
		{
			name:  "Basic index with multiple columns",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Index("name", "email").Name("users_name_email_index")
			},
			want:    "CREATE INDEX users_name_email_index ON users (name, email)",
			wantErr: false,
		},
		{
			name:  "Index with algorithm",
			table: "products",
			blueprint: func(table *Blueprint) {
				table.Index("sku").Name("products_sku_index").Algorithm("btree")
			},
			want:    "CREATE INDEX products_sku_index ON products USING btree (sku)",
			wantErr: false,
		},
		{
			name:  "Index without name (should use generated name)",
			table: "orders",
			blueprint: func(table *Blueprint) {
				table.Index("sku")
			},
			want:    "CREATE INDEX idx_orders_sku ON orders (sku)",
			wantErr: false,
		},
		{
			name:  "Index with empty column in list",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Index("name", "", "email").Name("users_invalid_index")
			},
			wantErr: true,
		},
		{
			name:  "Index with only empty column",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Index("").Name("users_empty_index")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: tt.table}
			tt.blueprint(bp)
			got, err := grammar.CompileIndex(bp, bp.commands[0])
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompileUnique(t *testing.T) {
	grammar := newPostgresGrammar()

	tests := []struct {
		name      string
		table     string
		blueprint func(table *Blueprint)
		want      string
		wantErr   bool
	}{
		{
			name:  "Basic unique index with single column",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Unique("email").Name("users_email_unique")
			},
			want:    "ALTER TABLE users ADD CONSTRAINT users_email_unique UNIQUE (email)",
			wantErr: false,
		},
		{
			name:  "Basic unique index with multiple columns",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Unique("name", "email").Name("users_name_email_unique")
			},
			want:    "ALTER TABLE users ADD CONSTRAINT users_name_email_unique UNIQUE (name, email)",
			wantErr: false,
		},
		{
			name:  "Unique index without name (should use generated name)",
			table: "orders",
			blueprint: func(table *Blueprint) {
				table.Unique("order_number")
			},
			want:    "ALTER TABLE orders ADD CONSTRAINT uk_orders_order_number UNIQUE (order_number)",
			wantErr: false,
		},
		{
			name:  "Unique index with deferrable true",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Unique("email").Name("users_email_unique").Deferrable(true)
			},
			want:    "ALTER TABLE users ADD CONSTRAINT users_email_unique UNIQUE (email) DEFERRABLE",
			wantErr: false,
		},
		{
			name:  "Unique index with deferrable false",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Unique("email").Name("users_email_unique").Deferrable(false)
			},
			want:    "ALTER TABLE users ADD CONSTRAINT users_email_unique UNIQUE (email) NOT DEFERRABLE",
			wantErr: false,
		},
		{
			name:  "Unique index with deferrable and initially immediate true",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Unique("email").Name("users_email_unique").Deferrable().InitiallyImmediate(true)
			},
			want:    "ALTER TABLE users ADD CONSTRAINT users_email_unique UNIQUE (email) DEFERRABLE INITIALLY IMMEDIATE",
			wantErr: false,
		},
		{
			name:  "Unique index with deferrable false and initially immediate false",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Unique("email").Name("users_email_unique").Deferrable().InitiallyImmediate(false)
			},
			want:    "ALTER TABLE users ADD CONSTRAINT users_email_unique UNIQUE (email) DEFERRABLE INITIALLY DEFERRED",
			wantErr: false,
		},
		{
			name:  "Unique index with deferrable false and initially immediate true (should be ignored)",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Unique("email").Name("users_email_unique").Deferrable(false).InitiallyImmediate(true)
			},
			want:    "ALTER TABLE users ADD CONSTRAINT users_email_unique UNIQUE (email) NOT DEFERRABLE",
			wantErr: false,
		},
		{
			name:  "Unique index with empty column in list",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Unique("name", "", "email").Name("users_invalid_unique")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: tt.table}
			tt.blueprint(bp)
			got, err := grammar.CompileUnique(bp, bp.commands[0])
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompileFullText(t *testing.T) {
	grammar := newPostgresGrammar()

	tests := []struct {
		name      string
		table     string
		blueprint func(table *Blueprint)
		want      string
		wantErr   bool
	}{
		{
			name:  "Basic fulltext index with single column",
			table: "articles",
			blueprint: func(table *Blueprint) {
				table.FullText("title").Name("articles_title_fulltext").Language("english")
			},
			want:    "CREATE INDEX articles_title_fulltext ON articles USING GIN (to_tsvector('english', title))",
			wantErr: false,
		},
		{
			name:  "Fulltext index with multiple columns",
			table: "documents",
			blueprint: func(table *Blueprint) {
				table.FullText("title", "content").Name("documents_title_content_fulltext").Language("english")
			},
			want:    "CREATE INDEX documents_title_content_fulltext ON documents USING GIN (to_tsvector('english', title) || to_tsvector('english', content))",
			wantErr: false,
		},
		{
			name:  "Fulltext index with different language",
			table: "posts",
			blueprint: func(table *Blueprint) {
				table.FullText("content").Name("posts_content_spanish_fulltext").Language("spanish")
			},
			want:    "CREATE INDEX posts_content_spanish_fulltext ON posts USING GIN (to_tsvector('spanish', content))",
			wantErr: false,
		},
		{
			name:  "Fulltext index without language (should use default english)",
			table: "blogs",
			blueprint: func(table *Blueprint) {
				table.FullText("body").Name("blogs_body_fulltext")
			},
			want:    "CREATE INDEX blogs_body_fulltext ON blogs USING GIN (to_tsvector('english', body))",
			wantErr: false,
		},
		{
			name:  "Fulltext index without name (should use generated name)",
			table: "news",
			blueprint: func(table *Blueprint) {
				table.FullText("headline")
			},
			want:    "CREATE INDEX ft_news_headline ON news USING GIN (to_tsvector('english', headline))",
			wantErr: false,
		},
		{
			name:  "Fulltext index with three columns",
			table: "products",
			blueprint: func(table *Blueprint) {
				table.FullText("name", "description", "tags").Name("products_search_fulltext").Language("english")
			},
			want:    "CREATE INDEX products_search_fulltext ON products USING GIN (to_tsvector('english', name) || to_tsvector('english', description) || to_tsvector('english', tags))",
			wantErr: false,
		},
		{
			name:  "Fulltext index with empty column in list",
			table: "articles",
			blueprint: func(table *Blueprint) {
				table.FullText("title", "", "content").Name("articles_invalid_fulltext")
			},
			wantErr: true,
		},
		{
			name:  "Fulltext index with only empty column",
			table: "articles",
			blueprint: func(table *Blueprint) {
				table.FullText("").Name("articles_empty_fulltext")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: tt.table}
			tt.blueprint(bp)
			got, err := grammar.CompileFullText(bp, bp.commands[0])
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_CompileDropUnique(t *testing.T) {
	grammar := newPostgresGrammar()

	tests := []struct {
		name      string
		indexName string
		want      string
		wantErr   bool
	}{
		{
			name:      "Drop unique index with valid name",
			indexName: "users_email_unique",
			want:      "ALTER TABLE  DROP CONSTRAINT users_email_unique",
			wantErr:   false,
		},
		{
			name:      "Drop unique index with complex name",
			indexName: "uk_users_email_name",
			want:      "ALTER TABLE  DROP CONSTRAINT uk_users_email_name",
			wantErr:   false,
		},
		{
			name:      "Drop unique index with numeric suffix",
			indexName: "users_email_unique_2",
			want:      "ALTER TABLE  DROP CONSTRAINT users_email_unique_2",
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
			bp := &Blueprint{}
			command := &command{index: tt.indexName}
			got, err := grammar.CompileDropUnique(bp, command)
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

func TestPgGrammar_CompileDropFulltext(t *testing.T) {
	grammar := newPostgresGrammar()

	tests := []struct {
		name      string
		indexName string
		want      string
		wantErr   bool
	}{
		{
			name:      "Drop fulltext index with valid name",
			indexName: "articles_title_fulltext",
			want:      "DROP INDEX articles_title_fulltext",
			wantErr:   false,
		},
		{
			name:      "Drop fulltext index with complex name",
			indexName: "documents_title_content_fulltext",
			want:      "DROP INDEX documents_title_content_fulltext",
			wantErr:   false,
		},
		{
			name:      "Drop fulltext index with underscore prefix",
			indexName: "idx_posts_content_fulltext",
			want:      "DROP INDEX idx_posts_content_fulltext",
			wantErr:   false,
		},
		{
			name:      "Drop fulltext index with numeric suffix",
			indexName: "search_index_fulltext_1",
			want:      "DROP INDEX search_index_fulltext_1",
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
			bp := &Blueprint{}
			command := &command{index: tt.indexName}
			got, err := grammar.CompileDropFulltext(bp, command)
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

func TestPgGrammar_CompilePrimary(t *testing.T) {
	grammar := newPostgresGrammar()

	tests := []struct {
		name      string
		table     string
		blueprint func(table *Blueprint)
		want      string
		wantErr   bool
	}{
		{
			name:  "Basic primary key with single column",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Primary("id").Name("users_id_primary")
			},
			want:    "ALTER TABLE users ADD CONSTRAINT users_id_primary PRIMARY KEY (id)",
			wantErr: false,
		},
		{
			name:  "Primary key with multiple columns",
			table: "user_roles",
			blueprint: func(table *Blueprint) {
				table.Primary("user_id", "role_id").Name("user_roles_primary")
			},
			want:    "ALTER TABLE user_roles ADD CONSTRAINT user_roles_primary PRIMARY KEY (user_id, role_id)",
			wantErr: false,
		},
		{
			name:  "Primary key without name (should use generated name)",
			table: "orders",
			blueprint: func(table *Blueprint) {
				table.Primary("order_id")
			},
			want:    "ALTER TABLE orders ADD CONSTRAINT pk_orders PRIMARY KEY (order_id)",
			wantErr: false,
		},
		{
			name:  "Primary key with three columns",
			table: "order_items",
			blueprint: func(table *Blueprint) {
				table.Primary("order_id", "product_id", "variant_id").Name("order_items_composite_pk")
			},
			want:    "ALTER TABLE order_items ADD CONSTRAINT order_items_composite_pk PRIMARY KEY (order_id, product_id, variant_id)",
			wantErr: false,
		},
		{
			name:  "Primary key with empty column in list",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Primary("id", "", "tenant_id").Name("users_invalid_primary")
			},
			wantErr: true,
		},
		{
			name:  "Primary key with only empty column",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Primary("").Name("users_empty_primary")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: tt.table}
			tt.blueprint(bp)
			got, err := grammar.CompilePrimary(bp, bp.commands[0])
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPgGrammar_GetType(t *testing.T) {
	grammar := newPostgresGrammar()

	tests := []struct {
		name      string
		blueprint func(table *Blueprint)
		want      string
	}{
		{
			name: "custom column type",
			blueprint: func(table *Blueprint) {
				table.Column("name", "CUSTOM_TYPE")
			},
			want: "CUSTOM_TYPE",
		},
		{
			name: "boolean column type",
			blueprint: func(table *Blueprint) {
				table.Boolean("active")
			},
			want: "BOOLEAN",
		},
		{
			name: "char column type",
			blueprint: func(table *Blueprint) {
				table.Char("code", 10)
			},
			want: "CHAR(10)",
		},
		{
			name: "char column type without length",
			blueprint: func(table *Blueprint) {
				table.Char("code")
			},
			want: "CHAR(255)",
		},
		{
			name: "string column type",
			blueprint: func(table *Blueprint) {
				table.String("name", 255)
			},
			want: "VARCHAR(255)",
		},
		{
			name: "decimal column type",
			blueprint: func(table *Blueprint) {
				table.Decimal("price", 10, 2)
			},
			want: "DECIMAL(10, 2)",
		},
		{
			name: "double column type",
			blueprint: func(table *Blueprint) {
				table.Double("value")
			},
			want: "DOUBLE PRECISION",
		},
		{
			name: "float column type with precision",
			blueprint: func(table *Blueprint) {
				table.Float("value", 6, 2)
			},
			want: "REAL",
		},
		{
			name: "float column type without precision",
			blueprint: func(table *Blueprint) {
				table.Float("value", 0, 0)
			},
			want: "REAL",
		},
		{
			name: "big integer column type",
			blueprint: func(table *Blueprint) {
				table.BigInteger("id")
			},
			want: "BIGINT",
		},
		{
			name: "big integer auto increment",
			blueprint: func(table *Blueprint) {
				table.BigInteger("id").AutoIncrement()
			},
			want: "BIGSERIAL",
		},
		{
			name: "integer column type",
			blueprint: func(table *Blueprint) {
				table.Integer("count")
			},
			want: "INTEGER",
		},
		{
			name: "integer auto increment",
			blueprint: func(table *Blueprint) {
				table.Integer("id").AutoIncrement()
			},
			want: "SERIAL",
		},
		{
			name: "small integer column type",
			blueprint: func(table *Blueprint) {
				table.SmallInteger("status")
			},
			want: "SMALLINT",
		},
		{
			name: "small integer auto increment",
			blueprint: func(table *Blueprint) {
				table.SmallInteger("id").Unsigned().AutoIncrement()
			},
			want: "SMALLSERIAL",
		},
		{
			name: "medium integer column type",
			blueprint: func(table *Blueprint) {
				table.MediumInteger("value")
			},
			want: "INTEGER",
		},
		{
			name: "medium auto increment",
			blueprint: func(table *Blueprint) {
				table.MediumIncrements("id")
			},
			want: "SERIAL",
		},
		{
			name: "tiny integer column type",
			blueprint: func(table *Blueprint) {
				table.TinyInteger("flag")
			},
			want: "SMALLINT",
		},
		{
			name: "tiny integer auto increment",
			blueprint: func(table *Blueprint) {
				table.TinyInteger("id").AutoIncrement()
			},
			want: "SMALLSERIAL",
		},
		{
			name: "time column type",
			blueprint: func(table *Blueprint) {
				table.Time("created_at")
			},
			want: "TIME(0)",
		},
		{
			name: "datetime column type with precision",
			blueprint: func(table *Blueprint) {
				table.DateTime("created_at", 6)
			},
			want: "TIMESTAMP(6)",
		},
		{
			name: "datetime column type without precision",
			blueprint: func(table *Blueprint) {
				table.DateTime("created_at", 0)
			},
			want: "TIMESTAMP(0)",
		},
		{
			name: "datetime tz column type with precision",
			blueprint: func(table *Blueprint) {
				table.DateTimeTz("created_at", 3)
			},
			want: "TIMESTAMPTZ(3)",
		},
		{
			name: "datetime tz column type without precision",
			blueprint: func(table *Blueprint) {
				table.DateTimeTz("created_at", 0)
			},
			want: "TIMESTAMPTZ(0)",
		},
		{
			name: "timestamp column type with precision",
			blueprint: func(table *Blueprint) {
				table.Timestamp("created_at", 6)
			},
			want: "TIMESTAMP(6)",
		},
		{
			name: "timestamp column type without precision",
			blueprint: func(table *Blueprint) {
				table.Timestamp("created_at", 0)
			},
			want: "TIMESTAMP(0)",
		},
		{
			name: "timestamp tz column type with precision",
			blueprint: func(table *Blueprint) {
				table.TimestampTz("created_at", 3)
			},
			want: "TIMESTAMPTZ(3)",
		},
		{
			name: "timestamp tz column type without precision",
			blueprint: func(table *Blueprint) {
				table.TimestampTz("created_at", 0)
			},
			want: "TIMESTAMPTZ(0)",
		},
		{
			name: "geography column type",
			blueprint: func(table *Blueprint) {
				table.Geography("location", "POINT", 4326)
			},
			want: "GEOGRAPHY(POINT, 4326)",
		},
		{
			name: "long text column type",
			blueprint: func(table *Blueprint) {
				table.LongText("content")
			},
			want: "TEXT",
		},
		{
			name: "text column type",
			blueprint: func(table *Blueprint) {
				table.Text("description")
			},
			want: "TEXT",
		},
		{
			name: "tiny text column type",
			blueprint: func(table *Blueprint) {
				table.TinyText("notes")
			},
			want: "VARCHAR(255)",
		},
		{
			name: "date column type",
			blueprint: func(table *Blueprint) {
				table.Date("birth_date")
			},
			want: "DATE",
		},
		{
			name: "year column type",
			blueprint: func(table *Blueprint) {
				table.Year("graduation_year")
			},
			want: "INTEGER",
		},
		{
			name: "json column type",
			blueprint: func(table *Blueprint) {
				table.JSON("metadata")
			},
			want: "JSON",
		},
		{
			name: "jsonb column type",
			blueprint: func(table *Blueprint) {
				table.JSONB("data")
			},
			want: "JSONB",
		},
		{
			name: "uuid column type",
			blueprint: func(table *Blueprint) {
				table.UUID("uuid")
			},
			want: "UUID",
		},
		{
			name: "binary column type",
			blueprint: func(table *Blueprint) {
				table.Binary("data")
			},
			want: "BYTEA",
		},
		{
			name: "point column type",
			blueprint: func(table *Blueprint) {
				table.Point("location")
			},
			want: "POINT(4326)",
		},
		{
			name: "Geography type",
			blueprint: func(table *Blueprint) {
				table.Geography("location", "POINT", 4326)
			},
			want: "GEOGRAPHY(POINT, 4326)",
		},
		{
			name: "Geometry type with SRID",
			blueprint: func(table *Blueprint) {
				table.Geometry("shape", "POLYGON", 4326)
			},
			want: "GEOMETRY(POLYGON, 4326)",
		},
		{
			name: "Geometry type without SRID",
			blueprint: func(table *Blueprint) {
				table.Geometry("shape", "POLYGON")
			},
			want: "GEOMETRY(POLYGON)",
		},
		{
			name: "Enum type",
			blueprint: func(table *Blueprint) {
				table.Enum("status", []string{"active", "inactive"})
			},
			want: "VARCHAR(255) CHECK (status IN ('active', 'inactive'))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: "test_table"}
			tt.blueprint(bp)
			got := grammar.getType(bp.columns[0])
			assert.Equal(t, tt.want, got)
		})
	}
}
