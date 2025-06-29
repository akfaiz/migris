package schema_test

import (
	"context"
	"database/sql"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ahmadfaizk/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockTx(t *testing.T) (*sql.Tx, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err, "failed to create sqlmock database")
	t.Cleanup(func() { db.Close() })

	mock.ExpectBegin()
	tx, err := db.Begin()
	require.NoError(t, err, "failed to begin transaction")

	return tx, mock
}

func toRegex(ddl string) string {
	// Escape all regex metacharacters
	replacer := strings.NewReplacer(
		`(`, `\(`,
		`)`, `\)`,
		`[`, `\[`,
		`]`, `\]`,
		`{`, `\{`,
		`}`, `\}`,
		`+`, `\+`,
		`*`, `\*`,
		`?`, `\?`,
		`.`, `\.`,
		`^`, `\^`,
		`$`, `\$`,
		`|`, `\|`,
	)

	escaped := replacer.Replace(ddl)

	// Replace multiple spaces and newlines with `\s+` for flexibility
	escaped = regexp.MustCompile(`\s+`).ReplaceAllString(escaped, `\s+`)

	return escaped
}

func TestCreate(t *testing.T) {
	ctx := context.Background()
	err := schema.SetDialect("postgres")
	assert.NoError(t, err)

	t.Run("when tx is nil, should return error", func(t *testing.T) {
		err = schema.Create(ctx, nil, "test_table", func(table *schema.Blueprint) {})
		assert.Error(t, err, "expected error when transaction is nil")
		assert.ErrorIs(t, err, schema.ErrTxIsNil)
	})

	testCases := []struct {
		name       string
		table      string
		blueprint  func(table *schema.Blueprint)
		wantErr    bool
		statements []string
	}{
		{
			name:    "when table name is empty, should return error",
			table:   "",
			wantErr: true,
		},
		{
			name:      "when blueprint is nil, should return error",
			table:     "test_table",
			blueprint: nil,
			wantErr:   true,
		},
		{
			name:  "when all parameters are valid, should create table successfully",
			table: "users",
			blueprint: func(table *schema.Blueprint) {
				table.ID()
				table.String("name", 255)
				table.String("email", 255)
				table.Timestamp("created_at").Default("CURRENT_TIMESTAMP")
			},
			statements: []string{
				"CREATE TABLE users (" +
					"id BIGSERIAL NOT NULL PRIMARY KEY, " +
					"name VARCHAR(255) NOT NULL, " +
					"email VARCHAR(255) NOT NULL, " +
					"created_at TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP NOT NULL)",
			},
		},
		{
			name:  "when have composite primary key should create it successfully",
			table: "user_roles",
			blueprint: func(table *schema.Blueprint) {
				table.Integer("user_id")
				table.Integer("role_id")

				table.Primary("user_id", "role_id")
			},
			statements: []string{
				"CREATE TABLE user_roles (" +
					"user_id INTEGER NOT NULL, " +
					"role_id INTEGER NOT NULL)",
				"ALTER TABLE user_roles ADD CONSTRAINT pk_user_roles PRIMARY KEY (user_id, role_id)",
			},
		},
		{
			name:  "when have composite unique index should create it successfully",
			table: "orders",
			blueprint: func(table *schema.Blueprint) {
				table.ID()
				table.String("order_id", 255)
				table.Integer("user_id")

				table.Unique("order_id", "user_id")
			},
			statements: []string{
				"CREATE TABLE orders (" +
					"id BIGSERIAL NOT NULL PRIMARY KEY, " +
					"order_id VARCHAR(255) NOT NULL, " +
					"user_id INTEGER NOT NULL)",
				"CREATE UNIQUE INDEX uk_orders_order_id_user_id ON orders(order_id, user_id)",
			},
		},
		{
			name:  "when have custom index should create it successfully",
			table: "orders",
			blueprint: func(table *schema.Blueprint) {
				table.ID()
				table.String("order_id", 255).Unique("uk_orders_order_id")
				table.Decimal("amount", 10, 2)
				table.Timestamp("created_at").Default("CURRENT_TIMESTAMP")

				table.Index("created_at").Name("idx_orders_created_at").Algorithm("btree")
			},
			statements: []string{
				"CREATE TABLE orders (" +
					"id BIGSERIAL NOT NULL PRIMARY KEY, " +
					"order_id VARCHAR(255) NOT NULL, " +
					"amount DECIMAL(10, 2) NOT NULL, " +
					"created_at TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP NOT NULL)",
				"CREATE UNIQUE INDEX uk_orders_order_id ON orders(order_id)",
				"CREATE INDEX idx_orders_created_at ON orders(created_at) USING btree",
			},
		},
		{
			name:  "when have foreign key should create it successfully",
			table: "orders",
			blueprint: func(table *schema.Blueprint) {
				table.ID()
				table.BigInteger("user_id")
				table.String("order_id", 255).Unique("uk_orders_order_id")
				table.Decimal("amount", 10, 2)
				table.Timestamp("created_at").Default("CURRENT_TIMESTAMP")

				table.Foreign("user_id").References("id").On("users").OnDelete("CASCADE").OnUpdate("CASCADE")
			},
			statements: []string{
				"CREATE TABLE orders (" +
					"id BIGSERIAL NOT NULL PRIMARY KEY, " +
					"user_id BIGINT NOT NULL, " +
					"order_id VARCHAR(255) NOT NULL, " +
					"amount DECIMAL(10, 2) NOT NULL, " +
					"created_at TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP NOT NULL)",
				"CREATE UNIQUE INDEX uk_orders_order_id ON orders(order_id)",
				"ALTER TABLE orders ADD CONSTRAINT fk_orders_users FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tx, mock := createMockTx(t)

			for _, stmt := range tc.statements {
				mock.ExpectExec(toRegex(stmt)).
					WillReturnResult(sqlmock.NewResult(1, 0))
			}

			err = schema.Create(ctx, tx, tc.table, tc.blueprint)
			if tc.wantErr {
				assert.Error(t, err, "expected error for test case: "+tc.name)
			} else {
				assert.NoError(t, err, "expected no error for test case: "+tc.name)
			}

			assert.NoError(t, mock.ExpectationsWereMet(), "expected no unfulfilled expectations")
		})
	}
}

func TestCreateIfNotExists(t *testing.T) {
	ctx := context.Background()

	err := schema.SetDialect("postgres")
	assert.NoError(t, err, "expected no error when creating schema")

	t.Run("when table name is empty, should return error", func(t *testing.T) {
		err := schema.CreateIfNotExists(ctx, nil, "", func(table *schema.Blueprint) {})
		assert.Error(t, err, "expected error when table name is empty")
		assert.ErrorIs(t, err, schema.ErrTableIsNotSet)
	})
	t.Run("when blueprint is nil, should return error", func(t *testing.T) {
		err = schema.CreateIfNotExists(ctx, nil, "test_table", nil)
		assert.Error(t, err, "expected error when blueprint is nil")
		assert.ErrorIs(t, err, schema.ErrBlueprintIsNil)
	})
	t.Run("when tx is nil, should return error", func(t *testing.T) {
		err = schema.CreateIfNotExists(ctx, nil, "test_table", func(table *schema.Blueprint) {})
		assert.Error(t, err, "expected error when transaction is nil")
		assert.ErrorIs(t, err, schema.ErrTxIsNil)
	})
	t.Run("when all parameters are valid", func(t *testing.T) {
		tx, mock := createMockTx(t)

		ddl := "CREATE TABLE IF NOT EXISTS users (" +
			"id BIGSERIAL NOT NULL PRIMARY KEY, " +
			"name VARCHAR NOT NULL)"
		mock.ExpectExec(toRegex(ddl)).
			WithoutArgs().
			WillReturnResult(sqlmock.NewResult(1, 0))

		err = schema.CreateIfNotExists(ctx, tx, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name")
		})
		assert.NoError(t, err, "expected no error when creating table with valid parameters")
	})
}

func TestDrop(t *testing.T) {
	ctx := context.Background()

	err := schema.SetDialect("postgres")
	assert.NoError(t, err, "expected no error when setting dialect")

	t.Run("when table name is empty, should return error", func(t *testing.T) {
		err := schema.Drop(ctx, nil, "")
		assert.Error(t, err, "expected error when table name is empty")
		assert.ErrorIs(t, err, schema.ErrTableIsNotSet)
	})
	t.Run("when tx is nil, should return error", func(t *testing.T) {
		err = schema.Drop(ctx, nil, "test_table")
		assert.Error(t, err, "expected error when transaction is nil")
		assert.ErrorIs(t, err, schema.ErrTxIsNil)
	})
	t.Run("when all parameters are valid", func(t *testing.T) {
		tx, mock := createMockTx(t)

		sql := "DROP TABLE users"
		mock.ExpectExec(toRegex(sql)).
			WillReturnResult(sqlmock.NewResult(1, 0))

		err = schema.Drop(ctx, tx, "users")
		assert.NoError(t, err, "expected no error when dropping table with valid parameters")
	})
}

func TestDropIfExists(t *testing.T) {
	ctx := context.Background()

	err := schema.SetDialect("postgres")
	assert.NoError(t, err, "expected no error when setting dialect")

	t.Run("when table name is empty, should return error", func(t *testing.T) {
		err := schema.DropIfExists(ctx, nil, "")
		assert.Error(t, err, "expected error when table name is empty")
		assert.ErrorIs(t, err, schema.ErrTableIsNotSet)
	})
	t.Run("when tx is nil, should return error", func(t *testing.T) {
		err = schema.DropIfExists(ctx, nil, "test_table")
		assert.Error(t, err, "expected error when transaction is nil")
		assert.ErrorIs(t, err, schema.ErrTxIsNil)
	})
	t.Run("when all parameters are valid", func(t *testing.T) {
		tx, mock := createMockTx(t)

		sql := "DROP TABLE IF EXISTS users"
		mock.ExpectExec(toRegex(sql)).
			WillReturnResult(sqlmock.NewResult(1, 0))

		err = schema.DropIfExists(ctx, tx, "users")
		assert.NoError(t, err, "expected no error when dropping table with valid parameters")
	})
}

func TestHasTable(t *testing.T) {
	ctx := context.Background()

	err := schema.SetDialect("postgres")
	assert.NoError(t, err, "expected no error when setting dialect")

	t.Run("when table name is empty, should return error", func(t *testing.T) {
		exists, err := schema.HasTable(ctx, nil, "")
		assert.Error(t, err, "expected error when table name is empty")
		assert.ErrorIs(t, err, schema.ErrTableIsNotSet)
		assert.False(t, exists, "expected exists to be false when table name is empty")
	})
	t.Run("when tx is nil, should return error", func(t *testing.T) {
		exists, err := schema.HasTable(ctx, nil, "test_table")
		assert.Error(t, err, "expected error when transaction is nil")
		assert.ErrorIs(t, err, schema.ErrTxIsNil)
		assert.False(t, exists, "expected exists to be false when transaction is nil")
	})
	t.Run("when all parameters are valid", func(t *testing.T) {
		tx, mock := createMockTx(t)

		sql := "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'users' AND table_type = 'BASE TABLE'"
		mock.ExpectQuery(toRegex(sql)).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		exists, err := schema.HasTable(ctx, tx, "users")
		assert.NoError(t, err, "expected no error when checking if table exists with valid parameters")
		assert.True(t, exists, "expected exists to be true for existing table")
	})
}

func TestRename(t *testing.T) {
	ctx := context.Background()

	err := schema.SetDialect("postgres")
	assert.NoError(t, err, "expected no error when setting dialect")

	t.Run("when old name is empty, should return error", func(t *testing.T) {
		err := schema.Rename(ctx, nil, "", "new_name")
		assert.Error(t, err, "expected error when old name is empty")
		assert.ErrorIs(t, err, schema.ErrTableIsNotSet)
	})
	t.Run("when new name is empty, should return error", func(t *testing.T) {
		err = schema.Rename(ctx, nil, "old_name", "")
		assert.Error(t, err, "expected error when new name is empty")
		assert.ErrorIs(t, err, schema.ErrTableIsNotSet)
	})
	t.Run("when tx is nil", func(t *testing.T) {
		err = schema.Rename(ctx, nil, "old_name", "new_name")
		assert.Error(t, err, "expected error when transaction is nil")
		assert.ErrorIs(t, err, schema.ErrTxIsNil)
	})
	t.Run("when all parameters are valid", func(t *testing.T) {
		tx, mock := createMockTx(t)

		sql := "ALTER TABLE old_name RENAME TO new_name"
		mock.ExpectExec(toRegex(sql)).
			WillReturnResult(sqlmock.NewResult(1, 0))

		err = schema.Rename(ctx, tx, "old_name", "new_name")
		assert.NoError(t, err, "expected no error when renaming table with valid parameters")
	})
}

func TestTable(t *testing.T) {
	ctx := context.Background()

	err := schema.SetDialect("postgres")
	assert.NoError(t, err, "expected no error when setting dialect")

	t.Run("when tx is nil, should return error", func(t *testing.T) {
		err = schema.Table(ctx, nil, "test_table", func(table *schema.Blueprint) {})
		assert.Error(t, err, "expected error when transaction is nil")
		assert.ErrorIs(t, err, schema.ErrTxIsNil)
	})

	testCases := []struct {
		name       string
		table      string
		blueprint  func(table *schema.Blueprint)
		statements []string
		wantErr    bool
	}{
		{
			name:    "when table name is empty, should return error",
			table:   "",
			wantErr: true,
		},
		{
			name:      "when blueprint is nil, should return error",
			table:     "test_table",
			blueprint: nil,
			wantErr:   true,
		},
		{
			name:  "when all parameters are valid, should execute statements",
			table: "users",
			blueprint: func(table *schema.Blueprint) {
				table.Integer("age")
				table.String("name", 255)
			},
			statements: []string{"ALTER TABLE users ADD COLUMN age INTEGER NOT NULL, ADD COLUMN name VARCHAR(255) NOT NULL"},
		},
		{
			name:  "when have drop column, should drop it successfully",
			table: "users",
			blueprint: func(table *schema.Blueprint) {
				table.DropColumn("age")
			},
			statements: []string{"ALTER TABLE users DROP COLUMN age"},
		},
		{
			name:  "when have rename column, should rename it successfully",
			table: "users",
			blueprint: func(table *schema.Blueprint) {
				table.RenameColumn("name", "full_name")
			},
			statements: []string{"ALTER TABLE users RENAME COLUMN name TO full_name"},
		},
		{
			name:  "when have drop index, should drop it successfully",
			table: "users",
			blueprint: func(table *schema.Blueprint) {
				table.DropIndex("idx_users_name")
			},
			statements: []string{"DROP INDEX idx_users_name"},
		},
		{
			name:  "when have drop unique index, should drop it successfully",
			table: "users",
			blueprint: func(table *schema.Blueprint) {
				table.DropUnique("uk_users_email")
			},
			statements: []string{"DROP INDEX uk_users_email"},
		},
		{
			name:  "when have rename index, should rename it successfully",
			table: "users",
			blueprint: func(table *schema.Blueprint) {
				table.RenameIndex("idx_users_name", "idx_users_full_name")
			},
			statements: []string{"ALTER INDEX idx_users_name RENAME TO idx_users_full_name"},
		},
		{
			name:  "when have drop primary key, should drop it successfully",
			table: "users",
			blueprint: func(table *schema.Blueprint) {
				table.DropPrimary("users_pkey")
			},
			statements: []string{"ALTER TABLE users DROP CONSTRAINT users_pkey"},
		},
		{
			name:  "when have drop foreign key, should drop it successfully",
			table: "orders",
			blueprint: func(table *schema.Blueprint) {
				table.DropForeign("fk_orders_users")
			},
			statements: []string{"ALTER TABLE orders DROP CONSTRAINT fk_orders_users"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tx, mock := createMockTx(t)

			for _, stmt := range tc.statements {
				mock.ExpectExec(toRegex(stmt)).
					WillReturnResult(sqlmock.NewResult(1, 0))
			}

			err := schema.Table(ctx, tx, tc.table, tc.blueprint)
			if tc.wantErr {
				assert.Error(t, err, "expected error for test case: "+tc.name)
			} else {
				assert.NoError(t, err, "expected no error for test case: "+tc.name)
			}

			assert.NoError(t, mock.ExpectationsWereMet(), "expected no unfulfilled expectations")
		})
	}
}
