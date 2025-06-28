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

	t.Run("when table name is empty, should return error", func(t *testing.T) {
		err := schema.Create(ctx, nil, "", func(table *schema.Blueprint) {})
		assert.Error(t, err, "expected error when table name is empty")
		assert.ErrorIs(t, err, schema.ErrTableIsNotSet)
	})
	t.Run("when blueprint is nil, should return error", func(t *testing.T) {
		err = schema.Create(ctx, nil, "test_table", nil)
		assert.Error(t, err, "expected error when blueprint is nil")
		assert.ErrorIs(t, err, schema.ErrBlueprintIsNil)
	})
	t.Run("when tx is nil, should return error", func(t *testing.T) {
		err = schema.Create(ctx, nil, "test_table", func(table *schema.Blueprint) {})
		assert.Error(t, err, "expected error when transaction is nil")
		assert.ErrorIs(t, err, schema.ErrTxIsNil)
	})
	t.Run("when all parameters are valid, should create table successfully", func(t *testing.T) {
		tx, mock := createMockTx(t)

		ddl := "CREATE TABLE users (" +
			"id BIGSERIAL NOT NULL PRIMARY KEY, " +
			"name VARCHAR(255) NOT NULL, " +
			"email VARCHAR(255) NOT NULL, " +
			"created_at TIMESTAMP(0) DEFAULT CURRENT_TIMESTAMP NOT NULL)"
		mock.ExpectExec(toRegex(ddl)).
			WithoutArgs().
			WillReturnResult(sqlmock.NewResult(1, 0))

		err = schema.Create(ctx, tx, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name")
			table.String("email")
			table.Timestamp("created_at").Default("CURRENT_TIMESTAMP")
		})
		assert.NoError(t, err, "expected no error when creating table with valid parameters")
	})
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
			"name VARCHAR(255) NOT NULL)"
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

func TestTable(t *testing.T) {
	ctx := context.Background()

	err := schema.SetDialect("postgres")
	assert.NoError(t, err, "expected no error when setting dialect")

	t.Run("when table name is empty, should return error", func(t *testing.T) {
		err := schema.Table(ctx, nil, "", func(table *schema.Blueprint) {})
		assert.Error(t, err, "expected error when table name is empty")
		assert.ErrorIs(t, err, schema.ErrTableIsNotSet)
	})
	t.Run("when blueprint is nil, should return error", func(t *testing.T) {
		err = schema.Table(ctx, nil, "test_table", nil)
		assert.Error(t, err, "expected error when blueprint is nil")
		assert.ErrorIs(t, err, schema.ErrBlueprintIsNil)
	})
	t.Run("when tx is nil, should return error", func(t *testing.T) {
		err = schema.Table(ctx, nil, "test_table", func(table *schema.Blueprint) {})
		assert.Error(t, err, "expected error when transaction is nil")
		assert.ErrorIs(t, err, schema.ErrTxIsNil)
	})
	t.Run("when all parameters are valid", func(t *testing.T) {
		tx, mock := createMockTx(t)

		sql := "ALTER TABLE users ADD COLUMN age INT"
		mock.ExpectExec(toRegex(sql)).
			WillReturnResult(sqlmock.NewResult(1, 0))

		err = schema.Table(ctx, tx, "users", func(table *schema.Blueprint) {
			table.Integer("age")
		})
		assert.NoError(t, err, "expected no error when altering table with valid parameters")
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
