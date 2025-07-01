package schema

import (
	"testing"

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
			want:    "CREATE TABLE users (id BIGINT UNSIGNED AUTO_INCREMENT NOT NULL PRIMARY KEY, name VARCHAR(255) NOT NULL)",
			wantErr: false,
		},
		{
			name:  "table with charset",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Charset("utf8mb4")
				table.ID()
			},
			want:    "CREATE TABLE users (id BIGINT UNSIGNED AUTO_INCREMENT NOT NULL PRIMARY KEY) DEFAULT CHARACTER SET utf8mb4",
			wantErr: false,
		},
		{
			name:  "table with collation",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Collation("utf8mb4_unicode_ci")
				table.ID()
			},
			want:    "CREATE TABLE users (id BIGINT UNSIGNED AUTO_INCREMENT NOT NULL PRIMARY KEY) COLLATE utf8mb4_unicode_ci",
			wantErr: false,
		},
		{
			name:  "table with engine",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Engine("InnoDB")
				table.ID()
			},
			want:    "CREATE TABLE users (id BIGINT UNSIGNED AUTO_INCREMENT NOT NULL PRIMARY KEY) ENGINE = InnoDB",
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
			want:    "CREATE TABLE users (id BIGINT UNSIGNED AUTO_INCREMENT NOT NULL PRIMARY KEY) DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci ENGINE = InnoDB",
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
			got, err := g.compileCreate(bp)
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
			name:  "add unique column",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.String("email", 255).Unique()
			},
			want:    "ALTER TABLE users ADD COLUMN email VARCHAR(255) NOT NULL UNIQUE",
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
			got, err := g.compileAdd(bp)
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
			name:  "no changed columns returns nil",
			table: "users",
			blueprint: func(table *Blueprint) {
				// No changed columns
			},
			want:    nil,
			wantErr: false,
		},
		{
			name:  "change single column type",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Integer("age").Change()
			},
			want:    []string{"ALTER TABLE users MODIFY COLUMN age INT"},
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
			want:    []string{"ALTER TABLE users MODIFY COLUMN status VARCHAR(50) DEFAULT 'active'"},
			wantErr: false,
		},
		{
			name:  "change column with null default",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Text("description").Default(nil).Change()
			},
			want:    []string{"ALTER TABLE users MODIFY COLUMN description TEXT DEFAULT NULL"},
			wantErr: false,
		},
		{
			name:  "change column with comment",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Integer("age").Comment("User age in years").Change()
			},
			want:    []string{"ALTER TABLE users MODIFY COLUMN age INT COMMENT 'User age in years'"},
			wantErr: false,
		},
		{
			name:  "change column with empty comment",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.Text("notes").Comment("").Change()
			},
			want:    []string{"ALTER TABLE users MODIFY COLUMN notes TEXT COMMENT ''"},
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
				"ALTER TABLE users MODIFY COLUMN name VARCHAR(200)",
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
			bp := &Blueprint{name: tt.table}
			tt.blueprint(bp)
			got, err := g.compileChange(bp)
			if tt.wantErr {
				assert.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			assert.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected SQL to match for test case: %s", tt.name)
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
		{
			name:    "empty new name should return error",
			table:   "users",
			newName: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{
				name:    tt.table,
				newName: tt.newName,
			}
			got, err := g.compileRename(bp)
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
			got, err := g.compileDrop(bp)
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
			got, err := g.compileDropIfExists(bp)
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
			name:      "no columns to drop should return error",
			table:     "users",
			blueprint: func(table *Blueprint) {},
			wantErr:   true,
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
			got, err := g.compileDropColumn(bp)
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
			got, err := g.compileRenameColumn(bp, tt.oldName, tt.newName)
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
		name       string
		table      string
		foreignKey *foreignKeyDefinition
		want       string
		wantErr    bool
	}{
		{
			name:  "basic foreign key",
			table: "users",
			foreignKey: &foreignKeyDefinition{
				column:     "company_id",
				on:         "companies",
				references: "id",
			},
			want:    "ALTER TABLE users ADD CONSTRAINT fk_users_companies FOREIGN KEY (company_id) REFERENCES companies(id)",
			wantErr: false,
		},
		{
			name:  "foreign key with on delete cascade",
			table: "users",
			foreignKey: &foreignKeyDefinition{
				column:     "company_id",
				on:         "companies",
				references: "id",
				onDelete:   "CASCADE",
			},
			want:    "ALTER TABLE users ADD CONSTRAINT fk_users_companies FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE",
			wantErr: false,
		},
		{
			name:  "foreign key with on update cascade",
			table: "users",
			foreignKey: &foreignKeyDefinition{
				column:     "company_id",
				on:         "companies",
				references: "id",
				onUpdate:   "CASCADE",
			},
			want:    "ALTER TABLE users ADD CONSTRAINT fk_users_companies FOREIGN KEY (company_id) REFERENCES companies(id) ON UPDATE CASCADE",
			wantErr: false,
		},
		{
			name:  "foreign key with both on delete and on update",
			table: "users",
			foreignKey: &foreignKeyDefinition{
				column:     "company_id",
				on:         "companies",
				references: "id",
				onDelete:   "CASCADE",
				onUpdate:   "SET NULL",
			},
			want:    "ALTER TABLE users ADD CONSTRAINT fk_users_companies FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE ON UPDATE SET NULL",
			wantErr: false,
		},
		{
			name:  "foreign key with on delete restrict",
			table: "orders",
			foreignKey: &foreignKeyDefinition{
				column:     "user_id",
				on:         "users",
				references: "id",
				onDelete:   "RESTRICT",
			},
			want:    "ALTER TABLE orders ADD CONSTRAINT fk_orders_users FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT",
			wantErr: false,
		},
		{
			name:  "foreign key with on delete set null",
			table: "posts",
			foreignKey: &foreignKeyDefinition{
				column:     "author_id",
				on:         "users",
				references: "id",
				onDelete:   "SET NULL",
			},
			want:    "ALTER TABLE posts ADD CONSTRAINT fk_posts_users FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE SET NULL",
			wantErr: false,
		},
		{
			name:  "empty column should return error",
			table: "users",
			foreignKey: &foreignKeyDefinition{
				column:     "",
				on:         "companies",
				references: "id",
			},
			wantErr: true,
		},
		{
			name:  "empty on table should return error",
			table: "users",
			foreignKey: &foreignKeyDefinition{
				column:     "company_id",
				on:         "",
				references: "id",
			},
			wantErr: true,
		},
		{
			name:  "empty references should return error",
			table: "users",
			foreignKey: &foreignKeyDefinition{
				column:     "company_id",
				on:         "companies",
				references: "",
			},
			wantErr: true,
		},
		{
			name:  "all empty values should return error",
			table: "users",
			foreignKey: &foreignKeyDefinition{
				column:     "",
				on:         "",
				references: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: tt.table}
			got, err := g.compileForeign(bp, tt.foreignKey)
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
			got, err := g.compileDropForeign(bp, tt.fkName)
			if tt.wantErr {
				assert.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			assert.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}
