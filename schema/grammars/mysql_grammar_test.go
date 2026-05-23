package grammars_test

import (
	"testing"

	"github.com/akfaiz/migris/internal/dialect"
	"github.com/akfaiz/migris/schema/blueprint"
	"github.com/akfaiz/migris/schema/grammars"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMysqlGrammar_CompileCreate(t *testing.T) {
	g, err := grammars.NewGrammar("mysql")
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
			want:    "CREATE TABLE `users` (`id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, `name` VARCHAR(255) NOT NULL, CONSTRAINT `users_id_primary` PRIMARY KEY (`id`))",
			wantErr: false,
		},
		{
			name:  "table with charset",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Charset("utf8mb4")
				table.ID()
			},
			want:    "CREATE TABLE `users` (`id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, CONSTRAINT `users_id_primary` PRIMARY KEY (`id`)) DEFAULT CHARACTER SET utf8mb4",
			wantErr: false,
		},
		{
			name:  "table with collation",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Collation("utf8mb4_unicode_ci")
				table.ID()
			},
			want:    "CREATE TABLE `users` (`id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, CONSTRAINT `users_id_primary` PRIMARY KEY (`id`)) COLLATE utf8mb4_unicode_ci",
			wantErr: false,
		},
		{
			name:  "table with engine",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Engine("InnoDB")
				table.ID()
			},
			want:    "CREATE TABLE `users` (`id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, CONSTRAINT `users_id_primary` PRIMARY KEY (`id`)) ENGINE = InnoDB",
			wantErr: false,
		},
		{
			name:  "table with charset, collation and engine",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Charset("utf8mb4")
				table.Collation("utf8mb4_unicode_ci")
				table.Engine("InnoDB")
				table.ID()
			},
			want:    "CREATE TABLE `users` (`id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, CONSTRAINT `users_id_primary` PRIMARY KEY (`id`)) DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci ENGINE = InnoDB",
			wantErr: false,
		},
		{
			name:  "table with empty column name should return error",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Integer("")
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
				require.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			require.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_CompileTableExists(t *testing.T) {
	g, err := grammars.NewGrammar("mysql")
	require.NoError(t, err)

	sql, err := g.CompileTableExists("", "db.users")
	require.NoError(t, err)
	assert.Equal(t, "SELECT 1 FROM information_schema.tables WHERE table_schema = 'db' AND table_name = 'users'", sql)

	sql2, err := g.CompileTableExists("", "users")
	require.NoError(t, err)
	assert.Equal(
		t,
		"SELECT 1 FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'users'",
		sql2,
	)
}

func TestMysqlGrammar_CompileTables(t *testing.T) {
	g, err := grammars.NewGrammar("mysql")
	require.NoError(t, err)

	sql, err := g.CompileTables("")
	require.NoError(t, err)
	assert.Equal(
		t,
		"SELECT table_name, table_comment FROM information_schema.tables WHERE table_schema = DATABASE() AND table_type = 'BASE TABLE'",
		sql,
	)
}

func TestMysqlGrammar_CompileColumns(t *testing.T) {
	g, err := grammars.NewGrammar("mysql")
	require.NoError(t, err)

	sql, err := g.CompileColumns("", "users")
	require.NoError(t, err)
	assert.Equal(t, "SHOW FULL COLUMNS FROM `users`", sql)
}

func TestMysqlGrammar_CompileIndexes(t *testing.T) {
	g, err := grammars.NewGrammar("mysql")
	require.NoError(t, err)

	sql, err := g.CompileIndexes("", "users")
	require.NoError(t, err)
	assert.Equal(t, "SHOW INDEX FROM `users`", sql)
}

func TestMysqlGrammar_CompileAdd(t *testing.T) {
	g, err := grammars.NewGrammar("mysql")
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
			want:    "ALTER TABLE `users` ADD COLUMN `email` VARCHAR(255) NOT NULL",
			wantErr: false,
		},
		{
			name:  "add multiple columns",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.String("email", 255)
				table.Integer("age")
			},
			want:    "ALTER TABLE `users` ADD COLUMN `email` VARCHAR(255) NOT NULL, ADD COLUMN `age` INT NOT NULL",
			wantErr: false,
		},
		{
			name:  "add column with nullable",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.String("email", 255).Nullable()
			},
			want:    "ALTER TABLE `users` ADD COLUMN `email` VARCHAR(255) NULL",
			wantErr: false,
		},
		{
			name:  "add column with default value",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.String("status", 50).Default("active")
			},
			want:    "ALTER TABLE `users` ADD COLUMN `status` VARCHAR(50) NOT NULL DEFAULT 'active'",
			wantErr: false,
		},
		{
			name:  "add column with comment",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.String("name", 255).Comment("User full name")
			},
			want:    "ALTER TABLE `users` ADD COLUMN `name` VARCHAR(255) NOT NULL COMMENT 'User full name'",
			wantErr: false,
		},
		{
			name:  "add column with after modifier",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.String("name", 255).After("id")
			},
			want:    "ALTER TABLE `users` ADD COLUMN `name` VARCHAR(255) NOT NULL AFTER `id`",
			wantErr: false,
		},
		{
			name:  "add column with first modifier",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.String("name", 255).First()
			},
			want:    "ALTER TABLE `users` ADD COLUMN `name` VARCHAR(255) NOT NULL FIRST",
			wantErr: false,
		},
		{
			name:  "add column with virtual as",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.String("full_name").VirtualAs("concat(first_name, ' ', last_name)")
			},
			want:    "ALTER TABLE `users` ADD COLUMN `full_name` VARCHAR(255) GENERATED ALWAYS AS (concat(first_name, ' ', last_name)) VIRTUAL",
			wantErr: false,
		},
		{
			name:  "add column with stored as",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.String("full_name").StoredAs("concat(first_name, ' ', last_name)")
			},
			want:    "ALTER TABLE `users` ADD COLUMN `full_name` VARCHAR(255) GENERATED ALWAYS AS (concat(first_name, ' ', last_name)) STORED",
			wantErr: false,
		},
		{
			name:  "no columns to add returns empty string",
			table: "users",
			blueprint: func(_ *blueprint.Blueprint) {
				// No columns added
			},
			want:    "",
			wantErr: false,
		},
		{
			name:  "add column with empty name should return error",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.String("", 255)
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
			got, err := g.CompileAdd(bp)
			if tt.wantErr {
				require.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			require.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_CompileChange(t *testing.T) {
	g, err := grammars.NewGrammar("mysql")
	require.NoError(t, err)

	tests := []struct {
		name      string
		table     string
		blueprint func(table *blueprint.Blueprint)
		want      []string
		wantErr   bool
	}{
		{
			name:      "no changed columns returns nil",
			table:     "users",
			blueprint: func(_ *blueprint.Blueprint) {},
			want:      nil,
			wantErr:   false,
		},
		{
			name:  "change single column type",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Integer("age").Change()
			},
			want:    []string{"ALTER TABLE `users` MODIFY COLUMN `age` INT NOT NULL"},
			wantErr: false,
		},
		{
			name:  "change column with nullable blueprint.Command",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.String("email", 255).Nullable().Change()
			},
			want:    []string{"ALTER TABLE `users` MODIFY COLUMN `email` VARCHAR(255) NULL"},
			wantErr: false,
		},
		{
			name:  "change column with not nullable blueprint.Command",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.String("name", 100).Nullable(false).Change()
			},
			want:    []string{"ALTER TABLE `users` MODIFY COLUMN `name` VARCHAR(100) NOT NULL"},
			wantErr: false,
		},
		{
			name:  "change column with default value",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.String("status", 50).Default("active").Change()
			},
			want:    []string{"ALTER TABLE `users` MODIFY COLUMN `status` VARCHAR(50) NOT NULL DEFAULT 'active'"},
			wantErr: false,
		},
		{
			name:  "change column with null default",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Text("description").Nullable().Default(nil).Change()
			},
			want:    []string{"ALTER TABLE `users` MODIFY COLUMN `description` TEXT NULL DEFAULT NULL"},
			wantErr: false,
		},
		{
			name:  "change column with comment",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Integer("age").Comment("User age in years").Change()
			},
			want:    []string{"ALTER TABLE `users` MODIFY COLUMN `age` INT NOT NULL COMMENT 'User age in years'"},
			wantErr: false,
		},
		{
			name:  "change column with empty comment",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Text("notes").Comment("").Change()
			},
			want:    []string{"ALTER TABLE `users` MODIFY COLUMN `notes` TEXT NOT NULL COMMENT ''"},
			wantErr: false,
		},
		{
			name:  "change column with all blueprint.Commands",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.String("email", 255).
					Nullable(false).
					Default("example@test.com").
					Comment("User email address").
					Change()
			},
			want: []string{
				"ALTER TABLE `users` MODIFY COLUMN `email` VARCHAR(255) NOT NULL DEFAULT 'example@test.com' COMMENT 'User email address'",
			},
			wantErr: false,
		},
		{
			name:  "change multiple columns",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.String("name", 200).Change()
				table.SmallInteger("age").Nullable().Change()
			},
			want: []string{
				"ALTER TABLE `users` MODIFY COLUMN `name` VARCHAR(200) NOT NULL",
				"ALTER TABLE `users` MODIFY COLUMN `age` SMALLINT NULL",
			},
			wantErr: false,
		},
		{
			name:  "change column with empty name should return error",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Integer("").Change()
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
			bp := &blueprint.Blueprint{Name: tableName, Grammar: g, Dialect: dialect.MySQL}
			tt.blueprint(bp)
			statements, err := bp.ToSQL()
			if tt.wantErr {
				require.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			require.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, statements, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_CompileRename(t *testing.T) {
	g, err := grammars.NewGrammar("mysql")
	require.NoError(t, err)

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
			want:    "RENAME TABLE `users` TO `customers`",
			wantErr: false,
		},
		{
			name:    "rename table with underscore names",
			table:   "old_table_name",
			newName: "new_table_name",
			want:    "RENAME TABLE `old_table_name` TO `new_table_name`",
			wantErr: false,
		},
		{
			name:    "rename table with numeric names",
			table:   "table1",
			newName: "table2",
			want:    "RENAME TABLE `table1` TO `table2`",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tableName := tt.table
			if tableName == "" && !tt.wantErr {
				tableName = tt.name
			}
			bp := &blueprint.Blueprint{
				Name: tableName,
			}
			bp.Rename(tt.newName)
			got, err := g.CompileRename(bp, bp.Commands[0])
			if tt.wantErr {
				require.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			require.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_CompileDrop(t *testing.T) {
	g, err := grammars.NewGrammar("mysql")
	require.NoError(t, err)

	tests := []struct {
		name    string
		table   string
		want    string
		wantErr bool
	}{
		{
			name:    "drop table with valid name",
			table:   "users",
			want:    "DROP TABLE `users`",
			wantErr: false,
		},
		{
			name:    "drop table with underscore name",
			table:   "user_profiles",
			want:    "DROP TABLE `user_profiles`",
			wantErr: false,
		},
		{
			name:    "drop table with numeric name",
			table:   "table123",
			want:    "DROP TABLE `table123`",
			wantErr: false,
		},
		{
			name:    "drop table with mixed case name",
			table:   "UserTable",
			want:    "DROP TABLE `UserTable`",
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
			tableName := tt.table
			if tableName == "" && !tt.wantErr {
				tableName = tt.name
			}
			bp := &blueprint.Blueprint{Name: tableName}
			got, err := g.CompileDrop(bp)
			if tt.wantErr {
				require.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			require.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_CompileDropIfExists(t *testing.T) {
	g, err := grammars.NewGrammar("mysql")
	require.NoError(t, err)

	tests := []struct {
		name    string
		table   string
		want    string
		wantErr bool
	}{
		{
			name:    "drop table if exists with valid name",
			table:   "users",
			want:    "DROP TABLE IF EXISTS `users`",
			wantErr: false,
		},
		{
			name:    "drop table if exists with underscore name",
			table:   "user_profiles",
			want:    "DROP TABLE IF EXISTS `user_profiles`",
			wantErr: false,
		},
		{
			name:    "drop table if exists with numeric name",
			table:   "table123",
			want:    "DROP TABLE IF EXISTS `table123`",
			wantErr: false,
		},
		{
			name:    "drop table if exists with mixed case name",
			table:   "UserTable",
			want:    "DROP TABLE IF EXISTS `UserTable`",
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
			tableName := tt.table
			if tableName == "" && !tt.wantErr {
				tableName = tt.name
			}
			bp := &blueprint.Blueprint{Name: tableName}
			got, err := g.CompileDropIfExists(bp)
			if tt.wantErr {
				require.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			require.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_CompileDropColumn(t *testing.T) {
	g, err := grammars.NewGrammar("mysql")
	require.NoError(t, err)

	tests := []struct {
		name      string
		table     string
		blueprint func(table *blueprint.Blueprint)
		want      string
		wantErr   bool
	}{
		{
			name:  "drop single column",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.DropColumn("email")
			},
			want:    "ALTER TABLE `users` DROP COLUMN `email`",
			wantErr: false,
		},
		{
			name:  "drop multiple columns",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.DropColumn("email", "phone", "address")
			},
			want:    "ALTER TABLE `users` DROP COLUMN `email`, DROP COLUMN `phone`, DROP COLUMN `address`",
			wantErr: false,
		},
		{
			name:  "empty column name should return error",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.DropColumn("email", "", "phone")
			},
			wantErr: true,
		},
		{
			name:  "single empty column name should return error",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.DropColumn("")
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
			bp := &blueprint.Blueprint{
				Name: tableName,
			}
			tt.blueprint(bp)
			got, err := g.CompileDropColumn(bp, bp.Commands[0])
			if tt.wantErr {
				require.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			require.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_CompileRenameColumn(t *testing.T) {
	g, err := grammars.NewGrammar("mysql")
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
			name:    "rename column with valid names",
			table:   "users",
			oldName: "email",
			newName: "email_address",
			want:    "ALTER TABLE `users` RENAME COLUMN `email` TO `email_address`",
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
			tableName := tt.table
			if tableName == "" && !tt.wantErr {
				tableName = tt.name
			}
			bp := &blueprint.Blueprint{Name: tableName}
			command := &blueprint.Command{From: tt.oldName, To: tt.newName}
			got, err := g.CompileRenameColumn(bp, command)
			if tt.wantErr {
				require.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			require.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_CompileForeign(t *testing.T) {
	g, err := grammars.NewGrammar("mysql")
	require.NoError(t, err)

	tests := []struct {
		name      string
		table     string
		blueprint func(table *blueprint.Blueprint)
		want      string
		wantErr   bool
	}{
		{
			name:  "basic foreign key",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Foreign("company_id").References("id").On("companies")
			},
			want:    "ALTER TABLE `users` ADD CONSTRAINT `users_company_id_foreign` FOREIGN KEY (`company_id`) REFERENCES `companies` (`id`)",
			wantErr: false,
		},
		{
			name:  "foreign key with on delete cascade",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Foreign("company_id").References("id").On("companies").CascadeOnDelete()
			},
			want:    "ALTER TABLE `users` ADD CONSTRAINT `users_company_id_foreign` FOREIGN KEY (`company_id`) REFERENCES `companies` (`id`) ON DELETE CASCADE",
			wantErr: false,
		},
		{
			name:  "foreign key with on update cascade",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Foreign("company_id").References("id").On("companies").CascadeOnUpdate()
			},
			want:    "ALTER TABLE `users` ADD CONSTRAINT `users_company_id_foreign` FOREIGN KEY (`company_id`) REFERENCES `companies` (`id`) ON UPDATE CASCADE",
			wantErr: false,
		},
		{
			name:  "foreign key with both on delete and on update",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Foreign("company_id").References("id").On("companies").
					CascadeOnDelete().NullOnUpdate()
			},
			want:    "ALTER TABLE `users` ADD CONSTRAINT `users_company_id_foreign` FOREIGN KEY (`company_id`) REFERENCES `companies` (`id`) ON DELETE CASCADE ON UPDATE SET NULL",
			wantErr: false,
		},
		{
			name:  "foreign key with on delete restrict",
			table: "orders",
			blueprint: func(table *blueprint.Blueprint) {
				table.Foreign("user_id").References("id").On("users").RestrictOnDelete()
			},
			want:    "ALTER TABLE `orders` ADD CONSTRAINT `orders_user_id_foreign` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE RESTRICT",
			wantErr: false,
		},
		{
			name:  "foreign key with on delete set null",
			table: "posts",
			blueprint: func(table *blueprint.Blueprint) {
				table.Foreign("author_id").References("id").On("users").NullOnDelete()
			},
			want:    "ALTER TABLE `posts` ADD CONSTRAINT `posts_author_id_foreign` FOREIGN KEY (`author_id`) REFERENCES `users` (`id`) ON DELETE SET NULL",
			wantErr: false,
		},
		{
			name:  "foreign key with no actions on delete or update",
			table: "comments",
			blueprint: func(table *blueprint.Blueprint) {
				table.Foreign("post_id").References("id").On("posts").NoActionOnDelete().NoActionOnUpdate()
			},
			want: "ALTER TABLE `comments` ADD CONSTRAINT `comments_post_id_foreign` FOREIGN KEY (`post_id`) REFERENCES `posts` (`id`) ON DELETE NO ACTION ON UPDATE NO ACTION",
		},
		{
			name:  "empty column should return error",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Foreign("").References("id").On("companies")
			},
			wantErr: true,
		},
		{
			name:  "empty on table should return error",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Foreign("company_id").References("id").On("")
			},
			wantErr: true,
		},
		{
			name:  "empty references should return error",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Foreign("company_id").References("").On("companies")
			},
			wantErr: true,
		},
		{
			name:  "all empty values should return error",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Foreign("").References("").On("")
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
			got, err := g.CompileForeign(bp, bp.Commands[0])
			if tt.wantErr {
				require.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			require.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_CompileDropForeign(t *testing.T) {
	g, err := grammars.NewGrammar("mysql")
	require.NoError(t, err)

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
			want:    "ALTER TABLE `users` DROP FOREIGN KEY `fk_users_company_id`",
			wantErr: false,
		},
		{
			name:    "empty foreign key name should return error",
			table:   "users",
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
			command := &blueprint.Command{Index: tt.fkName}
			got, err := g.CompileDropForeign(bp, command)
			if tt.wantErr {
				require.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			require.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_CompileIndex(t *testing.T) {
	g, err := grammars.NewGrammar("mysql")
	require.NoError(t, err)

	tests := []struct {
		name      string
		table     string
		blueprint func(table *blueprint.Blueprint)
		want      string
		wantErr   bool
	}{
		{
			name:  "basic index on single column",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Index("email")
			},
			want:    "CREATE INDEX `users_email_index` ON `users` (`email`)",
			wantErr: false,
		},
		{
			name:  "index on multiple columns",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Index("first_name", "last_name")
			},
			want:    "CREATE INDEX `users_first_name_last_name_index` ON `users` (`first_name`, `last_name`)",
			wantErr: false,
		},
		{
			name:  "index with custom name",
			table: "products",
			blueprint: func(table *blueprint.Blueprint) {
				table.Index("category_id").Name("idx_product_category")
			},
			want:    "CREATE INDEX `idx_product_category` ON `products` (`category_id`)",
			wantErr: false,
		},
		{
			name:  "index with algorithm",
			table: "logs",
			blueprint: func(table *blueprint.Blueprint) {
				table.Index("created_at").Algorithm("BTREE")
			},
			want:    "CREATE INDEX `logs_created_at_index` ON `logs` (`created_at`) USING BTREE",
			wantErr: false,
		},
		{
			name:  "index with custom name and algorithm",
			table: "orders",
			blueprint: func(table *blueprint.Blueprint) {
				table.Index("status", "created_at").Name("idx_order_status_date").Algorithm("HASH")
			},
			want:    "CREATE INDEX `idx_order_status_date` ON `orders` (`status`, `created_at`) USING HASH",
			wantErr: false,
		},
		{
			name:  "empty columns should return error",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Index("")
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
				require.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			require.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_CompileUnique(t *testing.T) {
	g, err := grammars.NewGrammar("mysql")
	require.NoError(t, err)

	tests := []struct {
		name      string
		table     string
		blueprint func(table *blueprint.Blueprint)
		want      string
		wantErr   bool
	}{
		{
			name:  "basic unique index on single column",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Unique("email")
			},
			want:    "CREATE UNIQUE INDEX `users_email_unique` ON `users` (`email`)",
			wantErr: false,
		},
		{
			name:  "unique index on multiple columns",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Unique("first_name", "last_name")
			},
			want:    "CREATE UNIQUE INDEX `users_first_name_last_name_unique` ON `users` (`first_name`, `last_name`)",
			wantErr: false,
		},
		{
			name:  "unique index with custom name",
			table: "products",
			blueprint: func(table *blueprint.Blueprint) {
				table.Unique("sku").Name("unique_product_sku")
			},
			want:    "CREATE UNIQUE INDEX `unique_product_sku` ON `products` (`sku`)",
			wantErr: false,
		},
		{
			name:  "unique index with algorithm",
			table: "logs",
			blueprint: func(table *blueprint.Blueprint) {
				table.Unique("transaction_id").Algorithm("BTREE")
			},
			want:    "CREATE UNIQUE INDEX `logs_transaction_id_unique` ON `logs` (`transaction_id`) USING BTREE",
			wantErr: false,
		},
		{
			name:  "unique index with custom name and algorithm",
			table: "orders",
			blueprint: func(table *blueprint.Blueprint) {
				table.Unique("order_number", "customer_id").Name("unique_order_customer").Algorithm("HASH")
			},
			want:    "CREATE UNIQUE INDEX `unique_order_customer` ON `orders` (`order_number`, `customer_id`) USING HASH",
			wantErr: false,
		},
		{
			name:  "empty column should return error",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Unique("")
			},
			wantErr: true,
		},
		{
			name:  "one empty column among multiple should return error",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Unique("email", "", "username")
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
				require.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			require.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_CompilePrimary(t *testing.T) {
	g, err := grammars.NewGrammar("mysql")
	require.NoError(t, err)

	tests := []struct {
		name      string
		table     string
		blueprint func(table *blueprint.Blueprint)
		want      string
		wantErr   bool
	}{
		{
			name:  "basic primary key on single column",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Primary("id")
			},
			want:    "ALTER TABLE `users` ADD CONSTRAINT `users_id_primary` PRIMARY KEY (`id`)",
			wantErr: false,
		},
		{
			name:  "composite primary key on multiple columns",
			table: "order_items",
			blueprint: func(table *blueprint.Blueprint) {
				table.Primary("order_id", "product_id")
			},
			want:    "ALTER TABLE `order_items` ADD CONSTRAINT `order_items_order_id_product_id_primary` PRIMARY KEY (`order_id`, `product_id`)",
			wantErr: false,
		},
		{
			name:  "primary key with custom name",
			table: "products",
			blueprint: func(table *blueprint.Blueprint) {
				table.Primary("sku").Name("primary_product_sku")
			},
			want:    "ALTER TABLE `products` ADD CONSTRAINT `primary_product_sku` PRIMARY KEY (`sku`)",
			wantErr: false,
		},
		{
			name:  "primary key on three columns",
			table: "user_permissions",
			blueprint: func(table *blueprint.Blueprint) {
				table.Primary("user_id", "resource_id", "permission_id")
			},
			want:    "ALTER TABLE `user_permissions` ADD CONSTRAINT `user_permissions_user_id_resource_id_permission_id_primary` PRIMARY KEY (`user_id`, `resource_id`, `permission_id`)",
			wantErr: false,
		},
		{
			name:  "primary key with custom name on multiple columns",
			table: "audit_logs",
			blueprint: func(table *blueprint.Blueprint) {
				table.Primary("timestamp", "user_id", "action").Name("pk_audit_composite")
			},
			want:    "ALTER TABLE `audit_logs` ADD CONSTRAINT `pk_audit_composite` PRIMARY KEY (`timestamp`, `user_id`, `action`)",
			wantErr: false,
		},
		{
			name:  "empty column should return error",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Primary("")
			},
			wantErr: true,
		},
		{
			name:  "one empty column among multiple should return error",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Primary("id", "", "email")
			},
			wantErr: true,
		},
		{
			name:  "empty column in the middle should return error",
			table: "orders",
			blueprint: func(table *blueprint.Blueprint) {
				table.Primary("user_id", "", "order_number")
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
				require.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			require.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_CompileFullText(t *testing.T) {
	g, err := grammars.NewGrammar("mysql")
	require.NoError(t, err)

	tests := []struct {
		name      string
		table     string
		blueprint func(table *blueprint.Blueprint)
		want      string
		wantErr   bool
	}{
		{
			name:  "basic fulltext index on single column",
			table: "articles",
			blueprint: func(table *blueprint.Blueprint) {
				table.FullText("content")
			},
			want:    "CREATE FULLTEXT INDEX `articles_content_fulltext` ON `articles` (`content`)",
			wantErr: false,
		},
		{
			name:  "fulltext index on multiple columns",
			table: "posts",
			blueprint: func(table *blueprint.Blueprint) {
				table.FullText("title", "content")
			},
			want:    "CREATE FULLTEXT INDEX `posts_title_content_fulltext` ON `posts` (`title`, `content`)",
			wantErr: false,
		},
		{
			name:  "fulltext index with custom name",
			table: "documents",
			blueprint: func(table *blueprint.Blueprint) {
				table.FullText("body").Name("fulltext_document_body")
			},
			want:    "CREATE FULLTEXT INDEX `fulltext_document_body` ON `documents` (`body`)",
			wantErr: false,
		},
		{
			name:  "fulltext index on three columns",
			table: "news",
			blueprint: func(table *blueprint.Blueprint) {
				table.FullText("title", "summary", "content")
			},
			want:    "CREATE FULLTEXT INDEX `news_title_summary_content_fulltext` ON `news` (`title`, `summary`, `content`)",
			wantErr: false,
		},
		{
			name:  "fulltext index with custom name on multiple columns",
			table: "blog_posts",
			blueprint: func(table *blueprint.Blueprint) {
				table.FullText("title", "excerpt", "body").Name("ft_blog_search")
			},
			want:    "CREATE FULLTEXT INDEX `ft_blog_search` ON `blog_posts` (`title`, `excerpt`, `body`)",
			wantErr: false,
		},
		{
			name:  "empty column should return error",
			table: "articles",
			blueprint: func(table *blueprint.Blueprint) {
				table.FullText("")
			},
			wantErr: true,
		},
		{
			name:  "one empty column among multiple should return error",
			table: "posts",
			blueprint: func(table *blueprint.Blueprint) {
				table.FullText("title", "", "content")
			},
			wantErr: true,
		},
		{
			name:  "empty column in the middle should return error",
			table: "documents",
			blueprint: func(table *blueprint.Blueprint) {
				table.FullText("title", "", "body")
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
				require.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			require.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_CompileDropIndex(t *testing.T) {
	g, err := grammars.NewGrammar("mysql")
	require.NoError(t, err)

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
			want:      "ALTER TABLE `users` DROP INDEX `idx_users_email`",
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
			tableName := tt.table
			if tableName == "" && !tt.wantErr {
				tableName = tt.name
			}
			bp := &blueprint.Blueprint{Name: tableName}
			command := &blueprint.Command{Index: tt.indexName}
			got, err := g.CompileDropIndex(bp, command)
			if tt.wantErr {
				require.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			require.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_CompileDropUnique(t *testing.T) {
	g, err := grammars.NewGrammar("mysql")
	require.NoError(t, err)

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
			want:      "ALTER TABLE `users` DROP INDEX `uk_users_email`",
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
			tableName := tt.table
			if tableName == "" && !tt.wantErr {
				tableName = tt.name
			}
			bp := &blueprint.Blueprint{Name: tableName}
			command := &blueprint.Command{Index: tt.indexName}
			got, err := g.CompileDropUnique(bp, command)
			if tt.wantErr {
				require.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			require.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_CompileDropFulltext(t *testing.T) {
	g, err := grammars.NewGrammar("mysql")
	require.NoError(t, err)

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
			want:      "ALTER TABLE `articles` DROP INDEX `ft_articles_content`",
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
			tableName := tt.table
			if tableName == "" && !tt.wantErr {
				tableName = tt.name
			}
			bp := &blueprint.Blueprint{Name: tableName}
			command := &blueprint.Command{Index: tt.indexName}
			got, err := g.CompileDropFulltext(bp, command)
			if tt.wantErr {
				require.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			require.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_CompileDropPrimary(t *testing.T) {
	g, err := grammars.NewGrammar("mysql")
	require.NoError(t, err)

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
			want:      "ALTER TABLE `users` DROP PRIMARY KEY",
			wantErr:   false,
		},
		{
			name:      "drop primary key with underscore name",
			table:     "user_profiles",
			indexName: "pk_user_profiles",
			want:      "ALTER TABLE `user_profiles` DROP PRIMARY KEY",
			wantErr:   false,
		},
		{
			name:      "drop primary key with numeric name",
			table:     "table123",
			indexName: "pk_123",
			want:      "ALTER TABLE `table123` DROP PRIMARY KEY",
			wantErr:   false,
		},
		{
			name:      "drop primary key with mixed case name",
			table:     "UserTable",
			indexName: "PkUserTable",
			want:      "ALTER TABLE `UserTable` DROP PRIMARY KEY",
			wantErr:   false,
		},
		{
			name:      "drop primary key with special characters in name",
			table:     "orders",
			indexName: "pk_order$id",
			want:      "ALTER TABLE `orders` DROP PRIMARY KEY",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tableName := tt.table
			if tableName == "" && !tt.wantErr {
				tableName = tt.name
			}
			bp := &blueprint.Blueprint{Name: tableName}
			command := &blueprint.Command{Index: tt.indexName}
			got, err := g.CompileDropPrimary(bp, command)
			if tt.wantErr {
				require.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			require.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_CompileRenameIndex(t *testing.T) {
	g, err := grammars.NewGrammar("mysql")
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
			name:    "rename index with valid names",
			table:   "users",
			oldName: "idx_users_email",
			newName: "idx_users_email_address",
			want:    "ALTER TABLE `users` RENAME INDEX `idx_users_email` TO `idx_users_email_address`",
			wantErr: false,
		},
		{
			name:    "rename index with underscore names",
			table:   "user_profiles",
			oldName: "idx_user_profiles_name",
			newName: "idx_user_profiles_full_name",
			want:    "ALTER TABLE `user_profiles` RENAME INDEX `idx_user_profiles_name` TO `idx_user_profiles_full_name`",
			wantErr: false,
		},
		{
			name:    "rename index with numeric names",
			table:   "orders",
			oldName: "idx_123",
			newName: "idx_456",
			want:    "ALTER TABLE `orders` RENAME INDEX `idx_123` TO `idx_456`",
			wantErr: false,
		},
		{
			name:    "rename index with mixed case names",
			table:   "Products",
			oldName: "IdxProductSku",
			newName: "IdxProductCode",
			want:    "ALTER TABLE `Products` RENAME INDEX `IdxProductSku` TO `IdxProductCode`",
			wantErr: false,
		},
		{
			name:    "rename index with special characters",
			table:   "logs",
			oldName: "idx_log$date",
			newName: "idx_log$timestamp",
			want:    "ALTER TABLE `logs` RENAME INDEX `idx_log$date` TO `idx_log$timestamp`",
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
			tableName := tt.table
			if tableName == "" && !tt.wantErr {
				tableName = tt.name
			}
			bp := &blueprint.Blueprint{Name: tableName}
			command := &blueprint.Command{From: tt.oldName, To: tt.newName}
			got, err := g.CompileRenameIndex(bp, command)
			if tt.wantErr {
				require.Error(t, err, "Expected error for test case: %s", tt.name)
				return
			}
			require.NoError(t, err, "Did not expect error for test case: %s", tt.name)
			assert.Equal(t, tt.want, got, "Expected SQL to match for test case: %s", tt.name)
		})
	}
}

func TestMysqlGrammar_GetType(t *testing.T) {
	g, err := grammars.NewGrammar("mysql")
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
			want: "TINYINT(1)",
		},
		{
			name: "char column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Char("code", 10)
			},
			want: "CHAR(10)",
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
			want: "DOUBLE",
		},
		{
			name: "float column type with precision",
			blueprint: func(table *blueprint.Blueprint) {
				table.Float("value", 6)
			},
			want: "FLOAT(6)",
		},
		{
			name: "float column type without precision",
			blueprint: func(table *blueprint.Blueprint) {
				table.Float("value")
			},
			want: "FLOAT(53)",
		},
		{
			name: "big integer column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.BigInteger("id")
			},
			want: "BIGINT",
		},
		{
			name: "integer column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Integer("count")
			},
			want: "INT",
		},
		{
			name: "small integer column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.SmallInteger("status")
			},
			want: "SMALLINT",
		},
		{
			name: "medium integer column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.MediumInteger("value")
			},
			want: "MEDIUMINT",
		},
		{
			name: "small integer column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.SmallInteger("level")
			},
			want: "SMALLINT",
		},
		{
			name: "tiny integer column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.TinyInteger("flag")
			},
			want: "TINYINT",
		},
		{
			name: "time column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Time("created_at")
			},
			want: "TIME",
		},
		{
			name: "datetime column type with precision",
			blueprint: func(table *blueprint.Blueprint) {
				table.DateTime("created_at", 6)
			},
			want: "DATETIME(6)",
		},
		{
			name: "datetime column type without precision",
			blueprint: func(table *blueprint.Blueprint) {
				table.DateTime("created_at", 0)
			},
			want: "DATETIME",
		},
		{
			name: "datetime tz column type with precision",
			blueprint: func(table *blueprint.Blueprint) {
				table.DateTimeTz("created_at", 3)
			},
			want: "DATETIME(3)",
		},
		{
			name: "datetime tz column type without precision",
			blueprint: func(table *blueprint.Blueprint) {
				table.DateTimeTz("created_at", 0)
			},
			want: "DATETIME",
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
			want: "TIMESTAMP",
		},
		{
			name: "timestamp tz column type with precision",
			blueprint: func(table *blueprint.Blueprint) {
				table.TimestampTz("created_at", 3)
			},
			want: "TIMESTAMP(3)",
		},
		{
			name: "timestamp tz column type without precision",
			blueprint: func(table *blueprint.Blueprint) {
				table.TimestampTz("created_at", 0)
			},
			want: "TIMESTAMP",
		},
		{
			name: "time tz column type with precision",
			blueprint: func(table *blueprint.Blueprint) {
				table.TimeTz("created_at", 3)
			},
			want: "TIME(3)",
		},
		{
			name: "time tz column type without precision",
			blueprint: func(table *blueprint.Blueprint) {
				table.TimeTz("created_at", 0)
			},
			want: "TIME",
		},
		{
			name: "set column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Set("tags", []string{"a", "b", "c"})
			},
			want: "SET('a', 'b', 'c')",
		},
		{
			name: "ulid column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.ULID("ulid")
			},
			want: "CHAR(26)",
		},
		{
			name: "ip address column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.IPAddress("ip")
			},
			want: "VARCHAR(45)",
		},
		{
			name: "mac address column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.MacAddress("mac")
			},
			want: "VARCHAR(17)",
		},
		{
			name: "vector column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Vector("embedding")
			},
			want: "VECTOR",
		},
		{
			name: "enum column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Enum("status", []string{"active", "inactive", "pending"})
			},
			want: "ENUM('active', 'inactive', 'pending')",
		},
		{
			name: "`long` text column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.LongText("content")
			},
			want: "LONGTEXT",
		},
		{
			name: "text column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Text("description")
			},
			want: "TEXT",
		},
		{
			name: "`medium` text column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.MediumText("summary")
			},
			want: "MEDIUMTEXT",
		},
		{
			name: "`tiny` text column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.TinyText("notes")
			},
			want: "TINYTEXT",
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
			want: "YEAR",
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
			want: "JSON",
		},
		{
			name: "uuid column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.UUID("uuid")
			},
			want: "CHAR(36)",
		},
		{
			name: "binary column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Binary("data")
			},
			want: "BLOB",
		},
		{
			name: "geography column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Geography("location", "LINESTRING", 4326)
			},
			want: "LINESTRING SRID 4326",
		},
		{
			name: "geometry column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Geometry("shape", "", 4326)
			},
			want: "GEOMETRY SRID 4326",
		},
		{
			name: "point column type",
			blueprint: func(table *blueprint.Blueprint) {
				table.Point("location")
			},
			want: "POINT SRID 4326",
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

func TestMysqlGrammar_CompileCreate_Temporary(t *testing.T) {
	g, err := grammars.NewGrammar("mysql")
	require.NoError(t, err)

	bp := &blueprint.Blueprint{Name: "temp_logs", TemporaryVal: true}
	bp.String("message")
	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)
	assert.Contains(t, sql, "CREATE TEMPORARY TABLE")
	assert.Contains(t, sql, "`temp_logs`")
}

func TestMysqlGrammar_CompileSpatialIndex(t *testing.T) {
	g, err := grammars.NewGrammar("mysql")
	require.NoError(t, err)

	tests := []struct {
		name    string
		setup   func(*blueprint.Blueprint)
		want    string
		wantErr bool
	}{
		{
			name: "basic spatial index",
			setup: func(bp *blueprint.Blueprint) {
				bp.SpatialIndex("location")
			},
			want: "CREATE SPATIAL INDEX `locations_location_spatialindex` ON `locations` (`location`)",
		},
		{
			name: "spatial index with custom name",
			setup: func(bp *blueprint.Blueprint) {
				bp.SpatialIndex("location").Name("idx_loc")
			},
			want: "CREATE SPATIAL INDEX `idx_loc` ON `locations` (`location`)",
		},
		{
			name:    "spatial index with empty column",
			setup:   func(bp *blueprint.Blueprint) { bp.SpatialIndex("") },
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &blueprint.Blueprint{Name: "locations"}
			tt.setup(bp)
			got, err := g.CompileSpatialIndex(bp, bp.Commands[0])
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMysqlGrammar_CompileDropSpatialIndex(t *testing.T) {
	g, err := grammars.NewGrammar("mysql")
	require.NoError(t, err)

	bp := &blueprint.Blueprint{Name: "locations"}
	bp.DropSpatialIndex("idx_loc")
	got, err := g.CompileDropSpatialIndex(bp, bp.Commands[0])
	require.NoError(t, err)
	assert.Equal(t, "ALTER TABLE `locations` DROP INDEX `idx_loc`", got)
}

func TestMysqlGrammar_CompileTableComment(t *testing.T) {
	g, err := grammars.NewGrammar("mysql")
	require.NoError(t, err)

	t.Run("alter table comment", func(t *testing.T) {
		bp := &blueprint.Blueprint{Name: "users"}
		bp.Comment("User accounts")
		got, err := g.CompileTableComment(bp, bp.Commands[0])
		require.NoError(t, err)
		assert.Equal(t, "ALTER TABLE `users` COMMENT = 'User accounts'", got)
	})

	t.Run("skip comment during create", func(t *testing.T) {
		bp := &blueprint.Blueprint{Name: "users"}
		bp.Create()
		bp.Comment("User accounts")
		// find tableComment command
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

func TestMysqlGrammar_CompileAutoIncrementStartingValues(t *testing.T) {
	g, err := grammars.NewGrammar("mysql")
	require.NoError(t, err)

	t.Run("sets auto increment start", func(t *testing.T) {
		bp := &blueprint.Blueprint{Name: "orders"}
		bp.AutoIncrementStartingValues(1000)
		var cmd *blueprint.Command
		for _, c := range bp.Commands {
			if c.Name == blueprint.CommandAutoIncrementStartingValues {
				cmd = c
			}
		}
		require.NotNil(t, cmd)
		got, err := g.CompileAutoIncrementStartingValues(bp, cmd)
		require.NoError(t, err)
		assert.Equal(t, "ALTER TABLE `orders` AUTO_INCREMENT = 1000", got)
	})

	t.Run("zero value produces empty string", func(t *testing.T) {
		bp := &blueprint.Blueprint{Name: "orders"}
		cmd := &blueprint.Command{Name: blueprint.CommandAutoIncrementStartingValues, Value: 0}
		got, err := g.CompileAutoIncrementStartingValues(bp, cmd)
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestMysqlGrammar_GetType_NewTypes(t *testing.T) {
	g, err := grammars.NewGrammar("mysql")
	require.NoError(t, err)

	tests := []struct {
		name      string
		blueprint func(*blueprint.Blueprint)
		want      string
	}{
		{
			name: "binary without length returns BLOB",
			blueprint: func(bp *blueprint.Blueprint) {
				bp.Binary("data")
			},
			want: "BLOB",
		},
		{
			name: "binary with length returns VARBINARY",
			blueprint: func(bp *blueprint.Blueprint) {
				bp.Binary("data", 255)
			},
			want: "VARBINARY(255)",
		},
		{
			name: "binary with length and Fixed returns BINARY",
			blueprint: func(bp *blueprint.Blueprint) {
				bp.Binary("hash", 32).Fixed()
			},
			want: "BINARY(32)",
		},
		{
			name: "float without explicit precision defaults to FLOAT(53)",
			blueprint: func(bp *blueprint.Blueprint) {
				bp.Float("score")
			},
			want: "FLOAT(53)",
		},
		{
			name: "float with precision returns FLOAT(n)",
			blueprint: func(bp *blueprint.Blueprint) {
				bp.Float("score", 6)
			},
			want: "FLOAT(6)",
		},
		{
			name: "float with zero precision returns bare FLOAT",
			blueprint: func(bp *blueprint.Blueprint) {
				bp.Float("score", 0)
			},
			want: "FLOAT",
		},
		{
			name: "vector without dimensions returns VECTOR",
			blueprint: func(bp *blueprint.Blueprint) {
				bp.Vector("embedding")
			},
			want: "VECTOR",
		},
		{
			name: "vector with dimensions returns VECTOR(n)",
			blueprint: func(bp *blueprint.Blueprint) {
				bp.Vector("embedding", 1536)
			},
			want: "VECTOR(1536)",
		},
		{
			name: "raw column returns definition",
			blueprint: func(bp *blueprint.Blueprint) {
				bp.RawColumn("payload", "MEDIUMBLOB NOT NULL")
			},
			want: "MEDIUMBLOB NOT NULL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &blueprint.Blueprint{Name: "t"}
			tt.blueprint(bp)
			got := g.GetType(bp.Columns[0])
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMysqlGrammar_CompileAdd_Modifiers(t *testing.T) {
	g, err := grammars.NewGrammar("mysql")
	require.NoError(t, err)

	tests := []struct {
		name      string
		table     string
		blueprint func(*blueprint.Blueprint)
		want      string
		wantErr   bool
	}{
		{
			name:  "on update current timestamp",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.Timestamp("updated_at").UseCurrentOnUpdate()
			},
			want: "ALTER TABLE `users` ADD COLUMN `updated_at` TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP",
		},
		{
			name:  "column charset",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.String("name", 255).Charset("utf8mb4")
			},
			want: "ALTER TABLE `users` ADD COLUMN `name` VARCHAR(255) CHARACTER SET utf8mb4 NOT NULL",
		},
		{
			name:  "column collation",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.String("name", 255).Collation("utf8mb4_unicode_ci")
			},
			want: "ALTER TABLE `users` ADD COLUMN `name` VARCHAR(255) COLLATE utf8mb4_unicode_ci NOT NULL",
		},
		{
			name:  "column charset and collation",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.String("bio", 500).Charset("utf8mb4").Collation("utf8mb4_unicode_ci")
			},
			want: "ALTER TABLE `users` ADD COLUMN `bio` VARCHAR(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL",
		},
		{
			name:  "invisible column",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.String("secret", 100).Invisible()
			},
			want: "ALTER TABLE `users` ADD COLUMN `secret` VARCHAR(100) NOT NULL INVISIBLE",
		},
		{
			name:  "raw column definition",
			table: "events",
			blueprint: func(table *blueprint.Blueprint) {
				table.RawColumn("payload", "MEDIUMBLOB")
			},
			want: "ALTER TABLE `events` ADD COLUMN `payload` MEDIUMBLOB NOT NULL",
		},
		{
			name:  "invisible nullable column",
			table: "users",
			blueprint: func(table *blueprint.Blueprint) {
				table.String("token", 255).Nullable().Invisible()
			},
			want: "ALTER TABLE `users` ADD COLUMN `token` VARCHAR(255) NULL INVISIBLE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &blueprint.Blueprint{Name: tt.table}
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

func TestMysqlGrammar_CompileCreate_RawColumn(t *testing.T) {
	g, err := grammars.NewGrammar("mysql")
	require.NoError(t, err)

	bp := &blueprint.Blueprint{Name: "events"}
	bp.RawColumn("payload", "MEDIUMBLOB")
	bp.String("source", 100)
	sql, err := g.CompileCreate(bp)
	require.NoError(t, err)
	assert.Equal(t, "CREATE TABLE `events` (`payload` MEDIUMBLOB NOT NULL, `source` VARCHAR(100) NOT NULL)", sql)
}
