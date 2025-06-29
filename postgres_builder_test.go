package schema

import (
	"context"
	"database/sql"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
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

func TestPgBuilderCreate(t *testing.T) {
	ctx := context.Background()

	builder, err := newBuilder(postgres)
	assert.NoError(t, err)

	t.Run("when tx is nil, should return error", func(t *testing.T) {
		err = builder.Create(ctx, nil, "test_table", func(table *Blueprint) {})
		assert.Error(t, err, "expected error when transaction is nil")
		assert.ErrorIs(t, err, ErrTxIsNil)
	})

	testCases := []struct {
		name       string
		table      string
		blueprint  func(table *Blueprint)
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
			blueprint: func(table *Blueprint) {
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
			blueprint: func(table *Blueprint) {
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
			blueprint: func(table *Blueprint) {
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
			blueprint: func(table *Blueprint) {
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
			blueprint: func(table *Blueprint) {
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

			err = builder.Create(ctx, tx, tc.table, tc.blueprint)
			if tc.wantErr {
				assert.Error(t, err, "expected error for test case: "+tc.name)
			} else {
				assert.NoError(t, err, "expected no error for test case: "+tc.name)
			}

			assert.NoError(t, mock.ExpectationsWereMet(), "expected no unfulfilled expectations")
		})
	}
}

func TestPgBuilderCreateIfNotExists(t *testing.T) {
	ctx := context.Background()

	builder, err := newBuilder(postgres)
	assert.NoError(t, err)

	t.Run("when table name is empty, should return error", func(t *testing.T) {
		err := builder.CreateIfNotExists(ctx, nil, "", func(table *Blueprint) {})
		assert.Error(t, err, "expected error when table name is empty")
		assert.ErrorIs(t, err, ErrTableIsNotSet)
	})
	t.Run("when all parameters are valid", func(t *testing.T) {
		tx, mock := createMockTx(t)

		ddl := "CREATE TABLE IF NOT EXISTS users (" +
			"id BIGSERIAL NOT NULL PRIMARY KEY, " +
			"name VARCHAR NOT NULL)"
		mock.ExpectExec(toRegex(ddl)).
			WithoutArgs().
			WillReturnResult(sqlmock.NewResult(1, 0))

		err = builder.CreateIfNotExists(ctx, tx, "users", func(table *Blueprint) {
			table.ID()
			table.String("name")
		})
		assert.NoError(t, err, "expected no error when creating table with valid parameters")
	})
}

func TestPgBuilderDrop(t *testing.T) {
	ctx := context.Background()

	builder, err := newBuilder(postgres)
	assert.NoError(t, err)

	t.Run("when table name is empty, should return error", func(t *testing.T) {
		err := builder.Drop(ctx, nil, "")
		assert.Error(t, err, "expected error when table name is empty")
		assert.ErrorIs(t, err, ErrTableIsNotSet)
	})
	t.Run("when all parameters are valid", func(t *testing.T) {
		tx, mock := createMockTx(t)

		sql := "DROP TABLE users"
		mock.ExpectExec(toRegex(sql)).
			WillReturnResult(sqlmock.NewResult(1, 0))

		err = builder.Drop(ctx, tx, "users")
		assert.NoError(t, err, "expected no error when dropping table with valid parameters")
	})
}

func TestPgBuilderDropIfExists(t *testing.T) {
	ctx := context.Background()

	builder, err := newBuilder(postgres)
	assert.NoError(t, err)

	t.Run("when table name is empty, should return error", func(t *testing.T) {
		err := builder.DropIfExists(ctx, nil, "")
		assert.Error(t, err, "expected error when table name is empty")
		assert.ErrorIs(t, err, ErrTableIsNotSet)
	})
	t.Run("when tx is nil, should return error", func(t *testing.T) {
		err = builder.DropIfExists(ctx, nil, "test_table")
		assert.Error(t, err, "expected error when transaction is nil")
		assert.ErrorIs(t, err, ErrTxIsNil)
	})
	t.Run("when all parameters are valid", func(t *testing.T) {
		tx, mock := createMockTx(t)

		sql := "DROP TABLE IF EXISTS users"
		mock.ExpectExec(toRegex(sql)).
			WillReturnResult(sqlmock.NewResult(1, 0))

		err = builder.DropIfExists(ctx, tx, "users")
		assert.NoError(t, err, "expected no error when dropping table with valid parameters")
	})
}

func TestPgBuilderGetColumns(t *testing.T) {
	ctx := context.Background()

	builder, err := newBuilder(postgres)
	assert.NoError(t, err)

	t.Run("when tx is nil, should return error", func(t *testing.T) {
		_, err := builder.GetColumns(ctx, nil, "users")
		assert.Error(t, err, "expected error when transaction is nil")
		assert.ErrorIs(t, err, ErrTxIsNil)
	})

	t.Run("when table name is empty, should return error", func(t *testing.T) {
		tx, _ := createMockTx(t)
		_, err := builder.GetColumns(ctx, tx, "")
		assert.Error(t, err, "expected error when table name is empty")
		assert.ErrorIs(t, err, ErrTableIsNotSet)
	})

	t.Run("when all parameters are valid", func(t *testing.T) {
		tx, mock := createMockTx(t)

		sql := "select a.attname as name, t.typname as type_name, format_type(a.atttypid, a.atttypmod) as type"
		mock.ExpectQuery(toRegex(sql)).
			WillReturnRows(
				sqlmock.NewRows([]string{"name", "type_name", "type", "collation", "nullable", "default", "comment"}).
					AddRow("id", "bigint", "bigint", nil, false, nil, nil),
			)

		columns, err := builder.GetColumns(ctx, tx, "users")
		assert.NoError(t, err, "expected no error when getting columns with valid parameters")
		require.Len(t, columns, 1, "expected one column to be returned")
	})
}

func TestPgBuilderGetIndexes(t *testing.T) {
	ctx := context.Background()

	builder, err := newBuilder(postgres)
	assert.NoError(t, err)

	t.Run("when tx is nil, should return error", func(t *testing.T) {
		_, err := builder.GetIndexes(ctx, nil, "users")
		assert.Error(t, err, "expected error when transaction is nil")
		assert.ErrorIs(t, err, ErrTxIsNil)
	})

	t.Run("when table name is empty, should return error", func(t *testing.T) {
		tx, _ := createMockTx(t)
		_, err := builder.GetIndexes(ctx, tx, "")
		assert.Error(t, err, "expected error when table name is empty")
		assert.ErrorIs(t, err, ErrTableIsNotSet)
	})

	t.Run("when all parameters are valid", func(t *testing.T) {
		tx, mock := createMockTx(t)

		sql := "select ic.relname as name, string_agg(a.attname, ',' order by indseq.ord) as columns"
		mock.ExpectQuery(toRegex(sql)).
			WillReturnRows(
				sqlmock.NewRows([]string{"name", "columns", "type", "unique", "primary"}).
					AddRow("idx_users_name", "name", "btree", true, false),
			)

		indexes, err := builder.GetIndexes(ctx, tx, "users")
		assert.NoError(t, err, "expected no error when getting indexes with valid parameters")
		require.Len(t, indexes, 1, "expected one index to be returned")
	})
}

func TestPgBuilderGetTables(t *testing.T) {
	ctx := context.Background()

	builder, err := newBuilder(postgres)
	assert.NoError(t, err)

	t.Run("when tx is nil, should return error", func(t *testing.T) {
		_, err := builder.GetTables(ctx, nil)
		assert.Error(t, err, "expected error when transaction is nil")
		assert.ErrorIs(t, err, ErrTxIsNil)
	})

	t.Run("when all parameters are valid", func(t *testing.T) {
		tx, mock := createMockTx(t)

		sql := "select c.relname as name, n.nspname as schema, pg_total_relation_size(c.oid) as size, obj_description(c.oid, 'pg_class') as comment from pg_class c"
		mock.ExpectQuery(toRegex(sql)).
			WillReturnRows(sqlmock.NewRows([]string{"name", "schema", "size", "comment"}).AddRow("users", "public", 0, nil))

		tables, err := builder.GetTables(ctx, tx)
		assert.NoError(t, err, "expected no error when getting tables with valid parameters")
		require.Len(t, tables, 1, "expected one table to be returned")
		user := tables[0]
		assert.Equal(t, "users", user.Name, "expected table name to be 'users'")
		assert.Equal(t, "public", user.Schema, "expected table schema to be 'public'")
		assert.Equal(t, int64(0), user.Size, "expected table size to be 0")
		assert.False(t, user.Comment.Valid, "expected table comment to be invalid (nil)")
	})
}

func TestPgBuilderHasColumns(t *testing.T) {
	ctx := context.Background()

	builder, err := newBuilder(postgres)
	assert.NoError(t, err)

	t.Run("when tx is nil, should return error", func(t *testing.T) {
		exists, err := builder.HasColumns(ctx, nil, "users", "name")
		assert.Error(t, err, "expected error when transaction is nil")
		assert.ErrorIs(t, err, ErrTxIsNil)
		assert.False(t, exists, "expected exists to be false when transaction is nil")
	})

	t.Run("when table name is empty, should return error", func(t *testing.T) {
		tx, _ := createMockTx(t)
		exists, err := builder.HasColumns(ctx, tx, "", "name")
		assert.Error(t, err, "expected error when table name is empty")
		assert.ErrorIs(t, err, ErrTableIsNotSet)
		assert.False(t, exists, "expected exists to be false when table name is empty")
	})

	t.Run("when all parameters are valid", func(t *testing.T) {
		tx, mock := createMockTx(t)

		sql := "select a.attname as name, t.typname as type_name, format_type(a.atttypid, a.atttypmod) as type"
		mock.ExpectQuery(toRegex(sql)).
			WillReturnRows(
				sqlmock.NewRows([]string{"name", "type_name", "type", "collation", "nullable", "default", "comment"}).
					AddRow("id", "bigint", "bigint", nil, false, nil, nil),
			)

		exists, err := builder.HasColumns(ctx, tx, "users", "id")
		assert.NoError(t, err, "expected no error when checking if columns exist with valid parameters")
		assert.True(t, exists, "expected exists to be true for existing column")
	})
}

func TestPgBuilderHasIndex(t *testing.T) {
	ctx := context.Background()

	builder, err := newBuilder(postgres)
	assert.NoError(t, err)

	t.Run("when tx is nil, should return error", func(t *testing.T) {
		exists, err := builder.HasIndex(ctx, nil, "users", []string{"idx_users_name"})
		assert.Error(t, err, "expected error when transaction is nil")
		assert.ErrorIs(t, err, ErrTxIsNil)
		assert.False(t, exists, "expected exists to be false when transaction is nil")
	})

	t.Run("when table name is empty, should return error", func(t *testing.T) {
		tx, _ := createMockTx(t)
		exists, err := builder.HasIndex(ctx, tx, "", []string{"idx_users_name"})
		assert.Error(t, err, "expected error when table name is empty")
		assert.ErrorIs(t, err, ErrTableIsNotSet)
		assert.False(t, exists, "expected exists to be false when table name is empty")
	})

	t.Run("when all parameters are valid", func(t *testing.T) {
		tx, mock := createMockTx(t)

		sql := "select ic.relname as name, string_agg(a.attname, ',' order by indseq.ord) as columns"
		mock.ExpectQuery(toRegex(sql)).
			WillReturnRows(
				sqlmock.NewRows([]string{"name", "columns", "type", "unique", "primary"}).
					AddRow("idx_users_name", "name", "btree", true, false),
			)

		exists, err := builder.HasIndex(ctx, tx, "users", []string{"idx_users_name"})
		assert.NoError(t, err, "expected no error when checking if index exists with valid parameters")
		assert.True(t, exists, "expected exists to be true for existing index")
	})

	t.Run("when use composite index, should return true", func(t *testing.T) {
		tx, mock := createMockTx(t)

		sql := "select ic.relname as name, string_agg(a.attname, ',' order by indseq.ord) as columns"
		mock.ExpectQuery(toRegex(sql)).
			WillReturnRows(
				sqlmock.NewRows([]string{"name", "columns", "type", "unique", "primary"}).
					AddRow("idx_users_name_email", "name,email", "btree", true, false),
			)

		exists, err := builder.HasIndex(ctx, tx, "users", []string{"name", "email"})
		assert.NoError(t, err, "expected no error when checking if composite index exists with valid parameters")
		assert.True(t, exists, "expected exists to be true for existing composite index")
	})
}

func TestPgBuilderHasTable(t *testing.T) {
	ctx := context.Background()

	builder, err := newBuilder(postgres)
	assert.NoError(t, err)

	t.Run("when table name is empty, should return error", func(t *testing.T) {
		exists, err := builder.HasTable(ctx, nil, "")
		assert.Error(t, err, "expected error when table name is empty")
		assert.ErrorIs(t, err, ErrTableIsNotSet)
		assert.False(t, exists, "expected exists to be false when table name is empty")
	})
	t.Run("when all parameters are valid", func(t *testing.T) {
		tx, mock := createMockTx(t)

		sql := "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'users' AND table_type = 'BASE TABLE'"
		mock.ExpectQuery(toRegex(sql)).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		exists, err := builder.HasTable(ctx, tx, "users")
		assert.NoError(t, err, "expected no error when checking if table exists with valid parameters")
		assert.True(t, exists, "expected exists to be true for existing table")
	})
}

func TestPgBuilderRename(t *testing.T) {
	ctx := context.Background()

	builder, err := newBuilder(postgres)
	assert.NoError(t, err)

	t.Run("when old name is empty, should return error", func(t *testing.T) {
		err := builder.Rename(ctx, nil, "", "new_name")
		assert.Error(t, err, "expected error when old name is empty")
		assert.ErrorIs(t, err, ErrTableIsNotSet)
	})
	t.Run("when new name is empty, should return error", func(t *testing.T) {
		err = builder.Rename(ctx, nil, "old_name", "")
		assert.Error(t, err, "expected error when new name is empty")
		assert.ErrorIs(t, err, ErrTableIsNotSet)
	})
	t.Run("when tx is nil", func(t *testing.T) {
		err = builder.Rename(ctx, nil, "old_name", "new_name")
		assert.Error(t, err, "expected error when transaction is nil")
		assert.ErrorIs(t, err, ErrTxIsNil)
	})
	t.Run("when all parameters are valid", func(t *testing.T) {
		tx, mock := createMockTx(t)

		sql := "ALTER TABLE old_name RENAME TO new_name"
		mock.ExpectExec(toRegex(sql)).
			WillReturnResult(sqlmock.NewResult(1, 0))

		err = builder.Rename(ctx, tx, "old_name", "new_name")
		assert.NoError(t, err, "expected no error when renaming table with valid parameters")
	})
}

func TestPgBuilderTable(t *testing.T) {
	ctx := context.Background()

	builder, err := newBuilder(postgres)
	assert.NoError(t, err)

	t.Run("when tx is nil, should return error", func(t *testing.T) {
		err = builder.Table(ctx, nil, "test_table", func(table *Blueprint) {})
		assert.Error(t, err, "expected error when transaction is nil")
		assert.ErrorIs(t, err, ErrTxIsNil)
	})

	testCases := []struct {
		name       string
		table      string
		blueprint  func(table *Blueprint)
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
			blueprint: func(table *Blueprint) {
				table.Integer("age")
				table.String("name", 255)
			},
			statements: []string{"ALTER TABLE users ADD COLUMN age INTEGER NOT NULL, ADD COLUMN name VARCHAR(255) NOT NULL"},
		},
		{
			name:  "when have drop column, should drop it successfully",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.DropColumn("age")
			},
			statements: []string{"ALTER TABLE users DROP COLUMN age"},
		},
		{
			name:  "when have rename column, should rename it successfully",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.RenameColumn("name", "full_name")
			},
			statements: []string{"ALTER TABLE users RENAME COLUMN name TO full_name"},
		},
		{
			name:  "when have drop index, should drop it successfully",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.DropIndex("idx_users_name")
			},
			statements: []string{"DROP INDEX idx_users_name"},
		},
		{
			name:  "when have drop unique index, should drop it successfully",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.DropUnique("uk_users_email")
			},
			statements: []string{"DROP INDEX uk_users_email"},
		},
		{
			name:  "when have rename index, should rename it successfully",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.RenameIndex("idx_users_name", "idx_users_full_name")
			},
			statements: []string{"ALTER INDEX idx_users_name RENAME TO idx_users_full_name"},
		},
		{
			name:  "when have drop primary key, should drop it successfully",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.DropPrimary("users_pkey")
			},
			statements: []string{"ALTER TABLE users DROP CONSTRAINT users_pkey"},
		},
		{
			name:  "when have drop foreign key, should drop it successfully",
			table: "orders",
			blueprint: func(table *Blueprint) {
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

			err := builder.Table(ctx, tx, tc.table, tc.blueprint)
			if tc.wantErr {
				assert.Error(t, err, "expected error for test case: "+tc.name)
			} else {
				assert.NoError(t, err, "expected no error for test case: "+tc.name)
			}

			assert.NoError(t, mock.ExpectationsWereMet(), "expected no unfulfilled expectations")
		})
	}
}
