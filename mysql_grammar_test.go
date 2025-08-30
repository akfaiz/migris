package schema

import (
	"testing"

	"github.com/afkdevs/go-schema/internal/dialect"
	"github.com/stretchr/testify/assert"
)

func TestMysqlGrammar_CompileCreate(t *testing.T) {
	g := newMysqlGrammar()

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
			want:    "CREATE TABLE users (id BIGINT UNSIGNED AUTO_INCREMENT NOT NULL, name VARCHAR(255) NOT NULL, CONSTRAINT pk_users PRIMARY KEY (id))",
			wantErr: false,
		},
		{
			name:  "table with charset",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Charset("utf8mb4")
				table.ID()
			},
			want:    "CREATE TABLE users (id BIGINT UNSIGNED AUTO_INCREMENT NOT NULL, CONSTRAINT pk_users PRIMARY KEY (id)) DEFAULT CHARACTER SET utf8mb4",
			wantErr: false,
		},
		{
			name:  "table with collation",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Collation("utf8mb4_unicode_ci")
				table.ID()
			},
			want:    "CREATE TABLE users (id BIGINT UNSIGNED AUTO_INCREMENT NOT NULL, CONSTRAINT pk_users PRIMARY KEY (id)) COLLATE utf8mb4_unicode_ci",
			wantErr: false,
		},
		{
			name:  "table with engine",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Engine("InnoDB")
				table.ID()
			},
			want:    "CREATE TABLE users (id BIGINT UNSIGNED AUTO_INCREMENT NOT NULL, CONSTRAINT pk_users PRIMARY KEY (id)) ENGINE = InnoDB",
			wantErr: false,
		},
		{
			name:  "table with charset, collation and engine",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Charset("utf8mb4")
				table.Collation("utf8mb4_unicode_ci")
				table.Engine("InnoDB")
				table.ID()
			},
			want:    "CREATE TABLE users (id BIGINT UNSIGNED AUTO_INCREMENT NOT NULL, CONSTRAINT pk_users PRIMARY KEY (id)) DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci ENGINE = InnoDB",
			wantErr: false,
		},
		{
			name:  "table with empty column name should return error",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Integer("")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: tt.table}
			tt.blueprint(bp)
			got, err := g.CompileCreate(bp)
			if tt.wantErr {
				assert.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			assert.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_CompileAdd(t *testing.T) {
	g := newMysqlGrammar()

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
			want:    "ALTER TABLE users ADD COLUMN email VARCHAR(255) NOT NULL",
			wantErr: false,
		},
		{
			name:  "add multiple columns",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.String("email", 255)
				table.Integer("age")
			},
			want:    "ALTER TABLE users ADD COLUMN email VARCHAR(255) NOT NULL, ADD COLUMN age INT NOT NULL",
			wantErr: false,
		},
		{
			name:  "add column with nullable",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.String("email", 255).Nullable()
			},
			want:    "ALTER TABLE users ADD COLUMN email VARCHAR(255) NULL",
			wantErr: false,
		},
		{
			name:  "add column with default value",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.String("status", 50).Default("active")
			},
			want:    "ALTER TABLE users ADD COLUMN status VARCHAR(50) DEFAULT 'active' NOT NULL",
			wantErr: false,
		},
		{
			name:  "add column with comment",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.String("name", 255).Comment("User full name")
			},
			want:    "ALTER TABLE users ADD COLUMN name VARCHAR(255) NOT NULL COMMENT 'User full name'",
			wantErr: false,
		},
		{
			name:  "no columns to add returns empty string",
			table: "users",
			blueprint: func(table *Blueprint) {
				// No columns added
			},
			want:    "",
			wantErr: false,
		},
		{
			name:  "add column with empty name should return error",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.String("", 255)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: tt.table}
			tt.blueprint(bp)
			got, err := g.CompileAdd(bp)
			if tt.wantErr {
				assert.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			assert.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_CompileChange(t *testing.T) {
	g := newMysqlGrammar()

	tests := []struct {
		name      string
		table     string
		blueprint func(table *Blueprint)
		want      []string
		wantErr   bool
	}{
		{
			name:      "no changed columns returns nil",
			table:     "users",
			blueprint: func(table *Blueprint) {},
			want:      nil,
			wantErr:   false,
		},
		{
			name:  "change single column type",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Integer("age").Change()
			},
			want:    []string{"ALTER TABLE users MODIFY COLUMN age INT NOT NULL"},
			wantErr: false,
		},
		{
			name:  "change column with nullable command",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.String("email", 255).Nullable().Change()
			},
			want:    []string{"ALTER TABLE users MODIFY COLUMN email VARCHAR(255) NULL"},
			wantErr: false,
		},
		{
			name:  "change column with not nullable command",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.String("name", 100).Nullable(false).Change()
			},
			want:    []string{"ALTER TABLE users MODIFY COLUMN name VARCHAR(100) NOT NULL"},
			wantErr: false,
		},
		{
			name:  "change column with default value",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.String("status", 50).Default("active").Change()
			},
			want:    []string{"ALTER TABLE users MODIFY COLUMN status VARCHAR(50) NOT NULL DEFAULT 'active'"},
			wantErr: false,
		},
		{
			name:  "change column with null default",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Text("description").Nullable().Default(nil).Change()
			},
			want:    []string{"ALTER TABLE users MODIFY COLUMN description TEXT NULL DEFAULT NULL"},
			wantErr: false,
		},
		{
			name:  "change column with comment",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Integer("age").Comment("User age in years").Change()
			},
			want:    []string{"ALTER TABLE users MODIFY COLUMN age INT NOT NULL COMMENT 'User age in years'"},
			wantErr: false,
		},
		{
			name:  "change column with empty comment",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Text("notes").Comment("").Change()
			},
			want:    []string{"ALTER TABLE users MODIFY COLUMN notes TEXT NOT NULL COMMENT ''"},
			wantErr: false,
		},
		{
			name:  "change column with all commands",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.String("email", 255).
					Nullable(false).
					Default("example@test.com").
					Comment("User email address").
					Change()
			},
			want:    []string{"ALTER TABLE users MODIFY COLUMN email VARCHAR(255) NOT NULL DEFAULT 'example@test.com' COMMENT 'User email address'"},
			wantErr: false,
		},
		{
			name:  "change multiple columns",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.String("name", 200).Change()
				table.SmallInteger("age").Nullable().Change()
			},
			want: []string{
				"ALTER TABLE users MODIFY COLUMN name VARCHAR(200) NOT NULL",
				"ALTER TABLE users MODIFY COLUMN age SMALLINT NULL",
			},
			wantErr: false,
		},
		{
			name:  "change column with empty name should return error",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Integer("").Change()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: tt.table, grammar: g, dialect: dialect.MySQL}
			tt.blueprint(bp)
			statements, err := bp.toSql()
			if tt.wantErr {
				assert.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			assert.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, statements, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_CompileRename(t *testing.T) {
	g := newMysqlGrammar()

	tests := []struct {
		name    string
		table   string
		newName string
		want    string
		wantErr bool
	}{
		{
			name:    "rename table with valid names",
			table:   "users",
			newName: "customers",
			want:    "ALTER TABLE users RENAME TO customers",
			wantErr: false,
		},
		{
			name:    "rename table with underscore names",
			table:   "old_table_name",
			newName: "new_table_name",
			want:    "ALTER TABLE old_table_name RENAME TO new_table_name",
			wantErr: false,
		},
		{
			name:    "rename table with numeric names",
			table:   "table1",
			newName: "table2",
			want:    "ALTER TABLE table1 RENAME TO table2",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{
				name: tt.table,
			}
			bp.rename(tt.newName)
			got, err := g.CompileRename(bp, bp.commands[0])
			if tt.wantErr {
				assert.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			assert.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_CompileDrop(t *testing.T) {
	g := newMysqlGrammar()

	tests := []struct {
		name    string
		table   string
		want    string
		wantErr bool
	}{
		{
			name:    "drop table with valid name",
			table:   "users",
			want:    "DROP TABLE users",
			wantErr: false,
		},
		{
			name:    "drop table with underscore name",
			table:   "user_profiles",
			want:    "DROP TABLE user_profiles",
			wantErr: false,
		},
		{
			name:    "drop table with numeric name",
			table:   "table123",
			want:    "DROP TABLE table123",
			wantErr: false,
		},
		{
			name:    "drop table with mixed case name",
			table:   "UserTable",
			want:    "DROP TABLE UserTable",
			wantErr: false,
		},
		{
			name:    "empty table name should return error",
			table:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: tt.table}
			got, err := g.CompileDrop(bp)
			if tt.wantErr {
				assert.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			assert.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_CompileDropIfExists(t *testing.T) {
	g := newMysqlGrammar()

	tests := []struct {
		name    string
		table   string
		want    string
		wantErr bool
	}{
		{
			name:    "drop table if exists with valid name",
			table:   "users",
			want:    "DROP TABLE IF EXISTS users",
			wantErr: false,
		},
		{
			name:    "drop table if exists with underscore name",
			table:   "user_profiles",
			want:    "DROP TABLE IF EXISTS user_profiles",
			wantErr: false,
		},
		{
			name:    "drop table if exists with numeric name",
			table:   "table123",
			want:    "DROP TABLE IF EXISTS table123",
			wantErr: false,
		},
		{
			name:    "drop table if exists with mixed case name",
			table:   "UserTable",
			want:    "DROP TABLE IF EXISTS UserTable",
			wantErr: false,
		},
		{
			name:    "empty table name should return error",
			table:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: tt.table}
			got, err := g.CompileDropIfExists(bp)
			if tt.wantErr {
				assert.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			assert.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_CompileDropColumn(t *testing.T) {
	g := newMysqlGrammar()

	tests := []struct {
		name      string
		table     string
		blueprint func(table *Blueprint)
		want      string
		wantErr   bool
	}{
		{
			name:  "drop single column",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.DropColumn("email")
			},
			want:    "ALTER TABLE users DROP COLUMN email",
			wantErr: false,
		},
		{
			name:  "drop multiple columns",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.DropColumn("email", "phone", "address")
			},
			want:    "ALTER TABLE users DROP COLUMN email, DROP COLUMN phone, DROP COLUMN address",
			wantErr: false,
		},
		{
			name:  "empty column name should return error",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.DropColumn("email", "", "phone")
			},
			wantErr: true,
		},
		{
			name:  "single empty column name should return error",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.DropColumn("")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{
				name: tt.table,
			}
			tt.blueprint(bp)
			got, err := g.CompileDropColumn(bp, bp.commands[0])
			if tt.wantErr {
				assert.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			assert.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_CompileRenameColumn(t *testing.T) {
	g := newMysqlGrammar()

	tests := []struct {
		name    string
		table   string
		oldName string
		newName string
		want    string
		wantErr bool
	}{
		{
			name:    "rename column with valid names",
			table:   "users",
			oldName: "email",
			newName: "email_address",
			want:    "ALTER TABLE users RENAME COLUMN email TO email_address",
			wantErr: false,
		},
		{
			name:    "empty old name should return error",
			table:   "users",
			oldName: "",
			newName: "new_name",
			wantErr: true,
		},
		{
			name:    "empty new name should return error",
			table:   "users",
			oldName: "old_name",
			newName: "",
			wantErr: true,
		},
		{
			name:    "both empty names should return error",
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
			got, err := g.CompileRenameColumn(bp, command)
			if tt.wantErr {
				assert.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			assert.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_CompileForeign(t *testing.T) {
	g := newMysqlGrammar()

	tests := []struct {
		name      string
		table     string
		blueprint func(table *Blueprint)
		want      string
		wantErr   bool
	}{
		{
			name:  "basic foreign key",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Foreign("company_id").References("id").On("companies")
			},
			want:    "ALTER TABLE users ADD CONSTRAINT fk_users_companies FOREIGN KEY (company_id) REFERENCES companies(id)",
			wantErr: false,
		},
		{
			name:  "foreign key with on delete cascade",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Foreign("company_id").References("id").On("companies").CascadeOnDelete()
			},
			want:    "ALTER TABLE users ADD CONSTRAINT fk_users_companies FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE",
			wantErr: false,
		},
		{
			name:  "foreign key with on update cascade",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Foreign("company_id").References("id").On("companies").CascadeOnUpdate()
			},
			want:    "ALTER TABLE users ADD CONSTRAINT fk_users_companies FOREIGN KEY (company_id) REFERENCES companies(id) ON UPDATE CASCADE",
			wantErr: false,
		},
		{
			name:  "foreign key with both on delete and on update",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Foreign("company_id").References("id").On("companies").
					CascadeOnDelete().NullOnUpdate()
			},
			want:    "ALTER TABLE users ADD CONSTRAINT fk_users_companies FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE ON UPDATE SET NULL",
			wantErr: false,
		},
		{
			name:  "foreign key with on delete restrict",
			table: "orders",
			blueprint: func(table *Blueprint) {
				table.Foreign("user_id").References("id").On("users").RestrictOnDelete()
			},
			want:    "ALTER TABLE orders ADD CONSTRAINT fk_orders_users FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT",
			wantErr: false,
		},
		{
			name:  "foreign key with on delete set null",
			table: "posts",
			blueprint: func(table *Blueprint) {
				table.Foreign("author_id").References("id").On("users").NullOnDelete()
			},
			want:    "ALTER TABLE posts ADD CONSTRAINT fk_posts_users FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE SET NULL",
			wantErr: false,
		},
		{
			name:  "foreign key with no actions on delete or update",
			table: "comments",
			blueprint: func(table *Blueprint) {
				table.Foreign("post_id").References("id").On("posts").NoActionOnDelete().NoActionOnUpdate()
			},
			want: "ALTER TABLE comments ADD CONSTRAINT fk_comments_posts FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE NO ACTION ON UPDATE NO ACTION",
		},
		{
			name:  "empty column should return error",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Foreign("").References("id").On("companies")
			},
			wantErr: true,
		},
		{
			name:  "empty on table should return error",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Foreign("company_id").References("id").On("")
			},
			wantErr: true,
		},
		{
			name:  "empty references should return error",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Foreign("company_id").References("").On("companies")
			},
			wantErr: true,
		},
		{
			name:  "all empty values should return error",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Foreign("").References("").On("")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: tt.table}
			tt.blueprint(bp)
			got, err := g.CompileForeign(bp, bp.commands[0])
			if tt.wantErr {
				assert.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			assert.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_CompileDropForeign(t *testing.T) {
	g := newMysqlGrammar()

	tests := []struct {
		name    string
		table   string
		fkName  string
		want    string
		wantErr bool
	}{
		{
			name:    "drop single foreign key",
			table:   "users",
			fkName:  "fk_users_company_id",
			want:    "ALTER TABLE users DROP FOREIGN KEY fk_users_company_id",
			wantErr: false,
		},
		{
			name:    "empty foreign key name should return error",
			table:   "users",
			fkName:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: tt.table}
			command := &command{index: tt.fkName}
			got, err := g.CompileDropForeign(bp, command)
			if tt.wantErr {
				assert.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			assert.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_CompileIndex(t *testing.T) {
	g := newMysqlGrammar()

	tests := []struct {
		name      string
		table     string
		blueprint func(table *Blueprint)
		want      string
		wantErr   bool
	}{
		{
			name:  "basic index on single column",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Index("email")
			},
			want:    "CREATE INDEX idx_users_email ON users (email)",
			wantErr: false,
		},
		{
			name:  "index on multiple columns",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Index("first_name", "last_name")
			},
			want:    "CREATE INDEX idx_users_first_name_last_name ON users (first_name, last_name)",
			wantErr: false,
		},
		{
			name:  "index with custom name",
			table: "products",
			blueprint: func(table *Blueprint) {
				table.Index("category_id").Name("idx_product_category")
			},
			want:    "CREATE INDEX idx_product_category ON products (category_id)",
			wantErr: false,
		},
		{
			name:  "index with algorithm",
			table: "logs",
			blueprint: func(table *Blueprint) {
				table.Index("created_at").Algorithm("BTREE")
			},
			want:    "CREATE INDEX idx_logs_created_at ON logs (created_at) USING BTREE",
			wantErr: false,
		},
		{
			name:  "index with custom name and algorithm",
			table: "orders",
			blueprint: func(table *Blueprint) {
				table.Index("status", "created_at").Name("idx_order_status_date").Algorithm("HASH")
			},
			want:    "CREATE INDEX idx_order_status_date ON orders (status, created_at) USING HASH",
			wantErr: false,
		},
		{
			name:  "empty columns should return error",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Index("")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: tt.table}
			tt.blueprint(bp)
			got, err := g.CompileIndex(bp, bp.commands[0])
			if tt.wantErr {
				assert.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			assert.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_CompileUnique(t *testing.T) {
	g := newMysqlGrammar()

	tests := []struct {
		name      string
		table     string
		blueprint func(table *Blueprint)
		want      string
		wantErr   bool
	}{
		{
			name:  "basic unique index on single column",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Unique("email")
			},
			want:    "CREATE UNIQUE INDEX uk_users_email ON users (email)",
			wantErr: false,
		},
		{
			name:  "unique index on multiple columns",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Unique("first_name", "last_name")
			},
			want:    "CREATE UNIQUE INDEX uk_users_first_name_last_name ON users (first_name, last_name)",
			wantErr: false,
		},
		{
			name:  "unique index with custom name",
			table: "products",
			blueprint: func(table *Blueprint) {
				table.Unique("sku").Name("unique_product_sku")
			},
			want:    "CREATE UNIQUE INDEX unique_product_sku ON products (sku)",
			wantErr: false,
		},
		{
			name:  "unique index with algorithm",
			table: "logs",
			blueprint: func(table *Blueprint) {
				table.Unique("transaction_id").Algorithm("BTREE")
			},
			want:    "CREATE UNIQUE INDEX uk_logs_transaction_id ON logs (transaction_id) USING BTREE",
			wantErr: false,
		},
		{
			name:  "unique index with custom name and algorithm",
			table: "orders",
			blueprint: func(table *Blueprint) {
				table.Unique("order_number", "customer_id").Name("unique_order_customer").Algorithm("HASH")
			},
			want:    "CREATE UNIQUE INDEX unique_order_customer ON orders (order_number, customer_id) USING HASH",
			wantErr: false,
		},
		{
			name:  "empty column should return error",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Unique("")
			},
			wantErr: true,
		},
		{
			name:  "one empty column among multiple should return error",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Unique("email", "", "username")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: tt.table}
			tt.blueprint(bp)
			got, err := g.CompileUnique(bp, bp.commands[0])
			if tt.wantErr {
				assert.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			assert.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_CompilePrimary(t *testing.T) {
	g := newMysqlGrammar()

	tests := []struct {
		name      string
		table     string
		blueprint func(table *Blueprint)
		want      string
		wantErr   bool
	}{
		{
			name:  "basic primary key on single column",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Primary("id")
			},
			want:    "ALTER TABLE users ADD CONSTRAINT pk_users PRIMARY KEY (id)",
			wantErr: false,
		},
		{
			name:  "composite primary key on multiple columns",
			table: "order_items",
			blueprint: func(table *Blueprint) {
				table.Primary("order_id", "product_id")
			},
			want:    "ALTER TABLE order_items ADD CONSTRAINT pk_order_items PRIMARY KEY (order_id, product_id)",
			wantErr: false,
		},
		{
			name:  "primary key with custom name",
			table: "products",
			blueprint: func(table *Blueprint) {
				table.Primary("sku").Name("primary_product_sku")
			},
			want:    "ALTER TABLE products ADD CONSTRAINT primary_product_sku PRIMARY KEY (sku)",
			wantErr: false,
		},
		{
			name:  "primary key on three columns",
			table: "user_permissions",
			blueprint: func(table *Blueprint) {
				table.Primary("user_id", "resource_id", "permission_id")
			},
			want:    "ALTER TABLE user_permissions ADD CONSTRAINT pk_user_permissions PRIMARY KEY (user_id, resource_id, permission_id)",
			wantErr: false,
		},
		{
			name:  "primary key with custom name on multiple columns",
			table: "audit_logs",
			blueprint: func(table *Blueprint) {
				table.Primary("timestamp", "user_id", "action").Name("pk_audit_composite")
			},
			want:    "ALTER TABLE audit_logs ADD CONSTRAINT pk_audit_composite PRIMARY KEY (timestamp, user_id, action)",
			wantErr: false,
		},
		{
			name:  "empty column should return error",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Primary("")
			},
			wantErr: true,
		},
		{
			name:  "one empty column among multiple should return error",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Primary("id", "", "email")
			},
			wantErr: true,
		},
		{
			name:  "empty column in the middle should return error",
			table: "orders",
			blueprint: func(table *Blueprint) {
				table.Primary("user_id", "", "order_number")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: tt.table}
			tt.blueprint(bp)
			got, err := g.CompilePrimary(bp, bp.commands[0])
			if tt.wantErr {
				assert.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			assert.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_CompileFullText(t *testing.T) {
	g := newMysqlGrammar()

	tests := []struct {
		name      string
		table     string
		blueprint func(table *Blueprint)
		want      string
		wantErr   bool
	}{
		{
			name:  "basic fulltext index on single column",
			table: "articles",
			blueprint: func(table *Blueprint) {
				table.FullText("content")
			},
			want:    "CREATE FULLTEXT INDEX ft_articles_content ON articles (content)",
			wantErr: false,
		},
		{
			name:  "fulltext index on multiple columns",
			table: "posts",
			blueprint: func(table *Blueprint) {
				table.FullText("title", "content")
			},
			want:    "CREATE FULLTEXT INDEX ft_posts_title_content ON posts (title, content)",
			wantErr: false,
		},
		{
			name:  "fulltext index with custom name",
			table: "documents",
			blueprint: func(table *Blueprint) {
				table.FullText("body").Name("fulltext_document_body")
			},
			want:    "CREATE FULLTEXT INDEX fulltext_document_body ON documents (body)",
			wantErr: false,
		},
		{
			name:  "fulltext index on three columns",
			table: "news",
			blueprint: func(table *Blueprint) {
				table.FullText("title", "summary", "content")
			},
			want:    "CREATE FULLTEXT INDEX ft_news_title_summary_content ON news (title, summary, content)",
			wantErr: false,
		},
		{
			name:  "fulltext index with custom name on multiple columns",
			table: "blog_posts",
			blueprint: func(table *Blueprint) {
				table.FullText("title", "excerpt", "body").Name("ft_blog_search")
			},
			want:    "CREATE FULLTEXT INDEX ft_blog_search ON blog_posts (title, excerpt, body)",
			wantErr: false,
		},
		{
			name:  "empty column should return error",
			table: "articles",
			blueprint: func(table *Blueprint) {
				table.FullText("")
			},
			wantErr: true,
		},
		{
			name:  "one empty column among multiple should return error",
			table: "posts",
			blueprint: func(table *Blueprint) {
				table.FullText("title", "", "content")
			},
			wantErr: true,
		},
		{
			name:  "empty column in the middle should return error",
			table: "documents",
			blueprint: func(table *Blueprint) {
				table.FullText("title", "", "body")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: tt.table}
			tt.blueprint(bp)
			got, err := g.CompileFullText(bp, bp.commands[0])
			if tt.wantErr {
				assert.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			assert.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_CompileDropIndex(t *testing.T) {
	g := newMysqlGrammar()

	tests := []struct {
		name      string
		table     string
		indexName string
		want      string
		wantErr   bool
	}{
		{
			name:      "drop index with valid name",
			table:     "users",
			indexName: "idx_users_email",
			want:      "ALTER TABLE users DROP INDEX idx_users_email",
			wantErr:   false,
		},
		{
			name:      "empty index name should return error",
			table:     "users",
			indexName: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: tt.table}
			command := &command{index: tt.indexName}
			got, err := g.CompileDropIndex(bp, command)
			if tt.wantErr {
				assert.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			assert.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_CompileDropUnique(t *testing.T) {
	g := newMysqlGrammar()

	tests := []struct {
		name      string
		table     string
		indexName string
		want      string
		wantErr   bool
	}{
		{
			name:      "drop unique index with valid name",
			table:     "users",
			indexName: "uk_users_email",
			want:      "ALTER TABLE users DROP INDEX uk_users_email",
			wantErr:   false,
		},
		{
			name:      "empty unique index name should return error",
			indexName: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: tt.table}
			command := &command{index: tt.indexName}
			got, err := g.CompileDropUnique(bp, command)
			if tt.wantErr {
				assert.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			assert.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_CompileDropFulltext(t *testing.T) {
	g := newMysqlGrammar()

	tests := []struct {
		name      string
		table     string
		indexName string
		want      string
		wantErr   bool
	}{
		{
			name:      "drop fulltext index with valid name",
			table:     "articles",
			indexName: "ft_articles_content",
			want:      "ALTER TABLE articles DROP INDEX ft_articles_content",
			wantErr:   false,
		},
		{
			name:      "empty fulltext index name should return error",
			indexName: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: tt.table}
			command := &command{index: tt.indexName}
			got, err := g.CompileDropFulltext(bp, command)
			if tt.wantErr {
				assert.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			assert.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_CompileDropPrimary(t *testing.T) {
	g := newMysqlGrammar()

	tests := []struct {
		name      string
		table     string
		indexName string
		want      string
		wantErr   bool
	}{
		{
			name:      "drop primary key with valid name",
			table:     "users",
			indexName: "pk_users",
			want:      "ALTER TABLE users DROP PRIMARY KEY",
			wantErr:   false,
		},
		{
			name:      "drop primary key with underscore name",
			table:     "user_profiles",
			indexName: "pk_user_profiles",
			want:      "ALTER TABLE user_profiles DROP PRIMARY KEY",
			wantErr:   false,
		},
		{
			name:      "drop primary key with numeric name",
			table:     "table123",
			indexName: "pk_123",
			want:      "ALTER TABLE table123 DROP PRIMARY KEY",
			wantErr:   false,
		},
		{
			name:      "drop primary key with mixed case name",
			table:     "UserTable",
			indexName: "PkUserTable",
			want:      "ALTER TABLE UserTable DROP PRIMARY KEY",
			wantErr:   false,
		},
		{
			name:      "drop primary key with special characters in name",
			table:     "orders",
			indexName: "pk_order$id",
			want:      "ALTER TABLE orders DROP PRIMARY KEY",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: tt.table}
			command := &command{index: tt.indexName}
			got, err := g.CompileDropPrimary(bp, command)
			if tt.wantErr {
				assert.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			assert.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_CompileRenameIndex(t *testing.T) {
	g := newMysqlGrammar()

	tests := []struct {
		name    string
		table   string
		oldName string
		newName string
		want    string
		wantErr bool
	}{
		{
			name:    "rename index with valid names",
			table:   "users",
			oldName: "idx_users_email",
			newName: "idx_users_email_address",
			want:    "ALTER TABLE users RENAME INDEX idx_users_email TO idx_users_email_address",
			wantErr: false,
		},
		{
			name:    "rename index with underscore names",
			table:   "user_profiles",
			oldName: "idx_user_profiles_name",
			newName: "idx_user_profiles_full_name",
			want:    "ALTER TABLE user_profiles RENAME INDEX idx_user_profiles_name TO idx_user_profiles_full_name",
			wantErr: false,
		},
		{
			name:    "rename index with numeric names",
			table:   "orders",
			oldName: "idx_123",
			newName: "idx_456",
			want:    "ALTER TABLE orders RENAME INDEX idx_123 TO idx_456",
			wantErr: false,
		},
		{
			name:    "rename index with mixed case names",
			table:   "Products",
			oldName: "IdxProductSku",
			newName: "IdxProductCode",
			want:    "ALTER TABLE Products RENAME INDEX IdxProductSku TO IdxProductCode",
			wantErr: false,
		},
		{
			name:    "rename index with special characters",
			table:   "logs",
			oldName: "idx_log$date",
			newName: "idx_log$timestamp",
			want:    "ALTER TABLE logs RENAME INDEX idx_log$date TO idx_log$timestamp",
			wantErr: false,
		},
		{
			name:    "empty old name should return error",
			table:   "users",
			oldName: "",
			newName: "new_index_name",
			wantErr: true,
		},
		{
			name:    "empty new name should return error",
			table:   "users",
			oldName: "old_index_name",
			newName: "",
			wantErr: true,
		},
		{
			name:    "both empty names should return error",
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
			got, err := g.CompileRenameIndex(bp, command)
			if tt.wantErr {
				assert.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			assert.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_GetType(t *testing.T) {
	g := newMysqlGrammar()

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
			want: "TINYINT(1)",
		},
		{
			name: "char column type",
			blueprint: func(table *Blueprint) {
				table.Char("code", 10)
			},
			want: "CHAR(10)",
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
			want: "DOUBLE",
		},
		{
			name: "float column type with precision",
			blueprint: func(table *Blueprint) {
				table.Float("value", 6)
			},
			want: "FLOAT(6)",
		},
		{
			name: "float column type without precision",
			blueprint: func(table *Blueprint) {
				table.Float("value")
			},
			want: "FLOAT(53)",
		},
		{
			name: "big integer column type",
			blueprint: func(table *Blueprint) {
				table.BigInteger("id")
			},
			want: "BIGINT",
		},
		{
			name: "integer column type",
			blueprint: func(table *Blueprint) {
				table.Integer("count")
			},
			want: "INT",
		},
		{
			name: "small integer column type",
			blueprint: func(table *Blueprint) {
				table.SmallInteger("status")
			},
			want: "SMALLINT",
		},
		{
			name: "medium integer column type",
			blueprint: func(table *Blueprint) {
				table.MediumInteger("value")
			},
			want: "MEDIUMINT",
		},
		{
			name: "small integer column type",
			blueprint: func(table *Blueprint) {
				table.SmallInteger("level")
			},
			want: "SMALLINT",
		},
		{
			name: "tiny integer column type",
			blueprint: func(table *Blueprint) {
				table.TinyInteger("flag")
			},
			want: "TINYINT",
		},
		{
			name: "time column type",
			blueprint: func(table *Blueprint) {
				table.Time("created_at")
			},
			want: "TIME",
		},
		{
			name: "datetime column type with precision",
			blueprint: func(table *Blueprint) {
				table.DateTime("created_at", 6)
			},
			want: "DATETIME(6)",
		},
		{
			name: "datetime column type without precision",
			blueprint: func(table *Blueprint) {
				table.DateTime("created_at", 0)
			},
			want: "DATETIME",
		},
		{
			name: "datetime tz column type with precision",
			blueprint: func(table *Blueprint) {
				table.DateTimeTz("created_at", 3)
			},
			want: "DATETIME(3)",
		},
		{
			name: "datetime tz column type without precision",
			blueprint: func(table *Blueprint) {
				table.DateTimeTz("created_at", 0)
			},
			want: "DATETIME",
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
			want: "TIMESTAMP",
		},
		{
			name: "timestamp tz column type with precision",
			blueprint: func(table *Blueprint) {
				table.TimestampTz("created_at", 3)
			},
			want: "TIMESTAMP(3)",
		},
		{
			name: "timestamp tz column type without precision",
			blueprint: func(table *Blueprint) {
				table.TimestampTz("created_at", 0)
			},
			want: "TIMESTAMP",
		},
		{
			name: "enum column type",
			blueprint: func(table *Blueprint) {
				table.Enum("status", []string{"active", "inactive", "pending"})
			},
			want: "ENUM('active', 'inactive', 'pending')",
		},
		{
			name: "long text column type",
			blueprint: func(table *Blueprint) {
				table.LongText("content")
			},
			want: "LONGTEXT",
		},
		{
			name: "text column type",
			blueprint: func(table *Blueprint) {
				table.Text("description")
			},
			want: "TEXT",
		},
		{
			name: "medium text column type",
			blueprint: func(table *Blueprint) {
				table.MediumText("summary")
			},
			want: "MEDIUMTEXT",
		},
		{
			name: "tiny text column type",
			blueprint: func(table *Blueprint) {
				table.TinyText("notes")
			},
			want: "TINYTEXT",
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
			want: "YEAR",
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
			want: "JSON",
		},
		{
			name: "uuid column type",
			blueprint: func(table *Blueprint) {
				table.UUID("uuid")
			},
			want: "CHAR(36)",
		},
		{
			name: "binary column type",
			blueprint: func(table *Blueprint) {
				table.Binary("data")
			},
			want: "BLOB",
		},
		{
			name: "geography column type",
			blueprint: func(table *Blueprint) {
				table.Geography("location", "LINESTRING", 4326)
			},
			want: "LINESTRING SRID 4326",
		},
		{
			name: "geometry column type",
			blueprint: func(table *Blueprint) {
				table.Geometry("shape", "", 4326)
			},
			want: "GEOMETRY SRID 4326",
		},
		{
			name: "point column type",
			blueprint: func(table *Blueprint) {
				table.Point("location")
			},
			want: "POINT SRID 4326",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: "test_table"}
			tt.blueprint(bp)
			got := g.getType(bp.columns[0])
			assert.Equal(t, tt.want, got, "Expected type to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_GetColumns(t *testing.T) {
	g := newMysqlGrammar()

	tests := []struct {
		name      string
		blueprint func(table *Blueprint)
		want      []string
		wantErr   bool
	}{
		{
			name: "single basic column",
			blueprint: func(table *Blueprint) {
				table.String("name", 255)
			},
			want:    []string{"name VARCHAR(255) NOT NULL"},
			wantErr: false,
		},
		{
			name: "multiple basic columns",
			blueprint: func(table *Blueprint) {
				table.String("name", 255)
				table.Integer("age")
			},
			want:    []string{"name VARCHAR(255) NOT NULL", "age INT NOT NULL"},
			wantErr: false,
		},
		{
			name: "column with default value",
			blueprint: func(table *Blueprint) {
				table.String("status", 50).Default("active")
			},
			want:    []string{"status VARCHAR(50) DEFAULT 'active' NOT NULL"},
			wantErr: false,
		},
		{
			name: "nullable column",
			blueprint: func(table *Blueprint) {
				table.String("email", 255).Nullable()
			},
			want:    []string{"email VARCHAR(255) NULL"},
			wantErr: false,
		},
		{
			name: "not nullable column",
			blueprint: func(table *Blueprint) {
				table.String("username", 100).Nullable(false)
			},
			want:    []string{"username VARCHAR(100) NOT NULL"},
			wantErr: false,
		},
		{
			name: "column with comment",
			blueprint: func(table *Blueprint) {
				table.String("name", 255).Comment("User full name")
			},
			want:    []string{"name VARCHAR(255) NOT NULL COMMENT 'User full name'"},
			wantErr: false,
		},
		{
			name: "primary key column",
			blueprint: func(table *Blueprint) {
				table.BigInteger("id").Primary()
			},
			want:    []string{"id BIGINT NOT NULL"},
			wantErr: false,
		},
		{
			name: "column with all attributes",
			blueprint: func(table *Blueprint) {
				table.String("email", 255).
					Default("user@example.com").
					Nullable().
					Comment("User email address")
			},
			want:    []string{"email VARCHAR(255) DEFAULT 'user@example.com' NULL COMMENT 'User email address'"},
			wantErr: false,
		},
		{
			name: "auto increment primary key",
			blueprint: func(table *Blueprint) {
				table.BigInteger("id").Unsigned().AutoIncrement().Primary()
			},
			want:    []string{"id BIGINT UNSIGNED AUTO_INCREMENT NOT NULL"},
			wantErr: false,
		},
		{
			name: "multiple columns with different attributes",
			blueprint: func(table *Blueprint) {
				table.BigInteger("id").Unsigned().AutoIncrement().Primary()
				table.String("name", 255).Comment("User name")
				table.String("email", 255).Nullable()
				table.Timestamp("created_at", 0).UseCurrent()
			},
			want: []string{
				"id BIGINT UNSIGNED AUTO_INCREMENT NOT NULL",
				"name VARCHAR(255) NOT NULL COMMENT 'User name'",
				"email VARCHAR(255) NULL",
				"created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL",
			},
			wantErr: false,
		},
		{
			name: "column with null default value",
			blueprint: func(table *Blueprint) {
				table.Text("description").Nullable().Default(nil)
			},
			want:    []string{"description TEXT DEFAULT NULL NULL"},
			wantErr: false,
		},
		{
			name: "empty column name should return error",
			blueprint: func(table *Blueprint) {
				table.String("", 255)
			},
			wantErr: true,
		},
		{
			name: "multiple columns with one empty name should return error",
			blueprint: func(table *Blueprint) {
				table.String("name", 255)
				table.Integer("")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: "test_table"}
			tt.blueprint(bp)
			got, err := g.getColumns(bp)
			if tt.wantErr {
				assert.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			assert.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected columns to match for test case: %s", tt.name)
		})
	}
}
