package builders_test

import (
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/akfaiz/migris/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockContext struct {
	dialect string

	execs    []string
	queries  []string
	queryRow []string

	queryResults map[string]queryResult
	row          schema.Row
}

type queryResult struct {
	rows schema.Rows
	err  error
}

func newMockContext(dialect string) *mockContext {
	return &mockContext{
		dialect:      dialect,
		queryResults: make(map[string]queryResult),
		row:          &mockRow{values: []any{1}},
	}
}

func (m *mockContext) Exec(query string, _ ...any) (sql.Result, error) {
	m.execs = append(m.execs, strings.TrimSpace(query))
	return mockResult{}, nil
}

func (m *mockContext) Query(query string, _ ...any) (schema.Rows, error) {
	query = strings.TrimSpace(query)
	m.queries = append(m.queries, query)
	if result, ok := m.queryResults[query]; ok {
		return result.rows, result.err
	}
	return &mockRows{}, nil
}

func (m *mockContext) QueryRow(query string, _ ...any) schema.Row {
	query = strings.TrimSpace(query)
	m.queryRow = append(m.queryRow, query)
	return m.row
}

func (m *mockContext) Dialect() string {
	return m.dialect
}

func (m *mockContext) withRows(query string, columns []string, rows ...[]any) {
	m.queryResults[strings.TrimSpace(query)] = queryResult{
		rows: &mockRows{columns: columns, rows: rows},
	}
}

type mockRows struct {
	columns []string
	rows    [][]any
	idx     int
	closed  bool
	err     error
}

func (m *mockRows) Next() bool {
	return m.idx < len(m.rows)
}

func (m *mockRows) Scan(dest ...any) error {
	if m.idx >= len(m.rows) {
		return errors.New("scan without row")
	}
	row := m.rows[m.idx]
	m.idx++
	if len(dest) != len(row) {
		return errors.New("scan destination count mismatch")
	}
	return assignValues(dest, row)
}

func (m *mockRows) Close() error {
	m.closed = true
	return nil
}

func (m *mockRows) Err() error {
	return m.err
}

func (m *mockRows) Columns() ([]string, error) {
	return m.columns, nil
}

type mockRow struct {
	values []any
	err    error
}

func (m *mockRow) Scan(dest ...any) error {
	if m.err != nil {
		return m.err
	}
	if len(dest) != len(m.values) {
		return errors.New("scan destination count mismatch")
	}
	return assignValues(dest, m.values)
}

type mockResult struct{}

func (mockResult) LastInsertId() (int64, error) { return 0, nil }
func (mockResult) RowsAffected() (int64, error) { return 0, nil }

func assignValues(dest []any, values []any) error {
	for i := range dest {
		switch d := dest[i].(type) {
		case *any:
			*d = values[i]
		case *string:
			v, ok := values[i].(string)
			if !ok {
				return errors.New("expected string value")
			}
			*d = v
		case *int:
			v, ok := values[i].(int)
			if !ok {
				return errors.New("expected int value")
			}
			*d = v
		case *bool:
			v, ok := values[i].(bool)
			if !ok {
				return errors.New("expected bool value")
			}
			*d = v
		case *sql.NullString:
			switch v := values[i].(type) {
			case sql.NullString:
				*d = v
			case string:
				*d = sql.NullString{String: v, Valid: true}
			case nil:
				*d = sql.NullString{}
			default:
				return errors.New("expected null string value")
			}
		default:
			return errors.New("unsupported scan destination")
		}
	}
	return nil
}

func TestBuilderInputValidation(t *testing.T) {
	for _, dialect := range []string{"sqlite3", "mysql", "mariadb", "postgres"} {
		t.Run(dialect, func(t *testing.T) {
			builder, err := schema.NewBuilder(dialect)
			require.NoError(t, err)
			ctx := newMockContext(dialect)

			require.Error(t, builder.Create(nil, "users", func(*schema.Blueprint) {}))
			require.Error(t, builder.Create(ctx, "", func(*schema.Blueprint) {}))
			require.Error(t, builder.Create(ctx, "users", nil))
			require.Error(t, builder.Table(nil, "users", func(*schema.Blueprint) {}))
			require.Error(t, builder.Table(ctx, "", func(*schema.Blueprint) {}))
			require.Error(t, builder.Table(ctx, "users", nil))
			require.Error(t, builder.Drop(nil, "users"))
			require.Error(t, builder.Drop(ctx, ""))
			require.Error(t, builder.DropIfExists(nil, "users"))
			require.Error(t, builder.DropIfExists(ctx, ""))
			require.Error(t, builder.Rename(nil, "users", "accounts"))
			require.Error(t, builder.Rename(ctx, "", "accounts"))
			require.Error(t, builder.Rename(ctx, "users", ""))
			require.Error(t, errOnly(builder.HasTable(nil, "users")))
			require.Error(t, errOnly(builder.HasTable(ctx, "")))
			require.Error(t, errOnly(builder.HasColumn(nil, "users", "id")))
			require.Error(t, errOnly(builder.HasColumn(ctx, "", "id")))
			require.Error(t, errOnly(builder.HasColumn(ctx, "users", "")))
			require.Error(t, errOnly(builder.HasColumns(nil, "users", []string{"id"})))
			require.Error(t, errOnly(builder.HasColumns(ctx, "", []string{"id"})))
			require.Error(t, errOnly(builder.HasColumns(ctx, "users", nil)))
			require.Error(t, errOnly(builder.HasIndex(nil, "users", []string{"id"})))
			require.Error(t, errOnly(builder.HasIndex(ctx, "", []string{"id"})))
			require.Error(t, errOnly(builder.HasIndex(ctx, "users", nil)))
			require.Error(t, errOnly(builder.GetColumns(nil, "users")))
			require.Error(t, errOnly(builder.GetColumns(ctx, "")))
			require.Error(t, errOnly(builder.GetIndexes(nil, "users")))
			require.Error(t, errOnly(builder.GetIndexes(ctx, "")))
			require.Error(t, errOnly(builder.GetTables(nil)))
		})
	}
}

func TestBuilderExecutesCompiledStatements(t *testing.T) {
	tests := []struct {
		dialect   string
		fragments []string
	}{
		{
			dialect: "sqlite3",
			fragments: []string{
				"CREATE TABLE \"users\"",
				"ALTER TABLE \"users\" ADD COLUMN \"email\"",
				"DROP TABLE \"users\"",
				"DROP TABLE IF EXISTS \"users\"",
				"ALTER TABLE \"users\" RENAME TO \"accounts\"",
			},
		},
		{
			dialect: "mysql",
			fragments: []string{
				"CREATE TABLE `users`",
				"ALTER TABLE `users` ADD COLUMN `email`",
				"DROP TABLE `users`",
				"DROP TABLE IF EXISTS `users`",
				"RENAME TABLE `users` TO `accounts`",
			},
		},
		{
			dialect: "mariadb",
			fragments: []string{
				"CREATE TABLE `users`",
				"ALTER TABLE `users` ADD COLUMN `email`",
				"DROP TABLE `users`",
				"DROP TABLE IF EXISTS `users`",
				"RENAME TABLE `users` TO `accounts`",
			},
		},
		{
			dialect: "postgres",
			fragments: []string{
				"CREATE TABLE \"users\"",
				"ALTER TABLE \"users\" ADD COLUMN \"email\"",
				"DROP TABLE \"users\"",
				"DROP TABLE IF EXISTS \"users\"",
				"ALTER TABLE \"users\" RENAME TO \"accounts\"",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.dialect, func(t *testing.T) {
			builder, err := schema.NewBuilder(tt.dialect)
			require.NoError(t, err)
			ctx := newMockContext(tt.dialect)

			require.NoError(t, builder.Create(ctx, "users", func(table *schema.Blueprint) {
				table.ID()
				table.String("name", 100)
			}))
			require.NoError(t, builder.Table(ctx, "users", func(table *schema.Blueprint) {
				table.String("email", 255).Nullable()
			}))
			require.NoError(t, builder.Drop(ctx, "users"))
			require.NoError(t, builder.DropIfExists(ctx, "users"))
			require.NoError(t, builder.Rename(ctx, "users", "accounts"))

			require.Len(t, ctx.execs, len(tt.fragments))
			for i, fragment := range tt.fragments {
				assert.Contains(t, ctx.execs[i], fragment)
			}
		})
	}
}

func TestBuilderGetColumnsMapsRows(t *testing.T) {
	tests := []struct {
		dialect string
		query   string
		columns []string
		rows    [][]any
	}{
		{
			dialect: "sqlite3",
			query:   compiledColumnsQuery("sqlite3"),
			rows: [][]any{
				{0, "id", "INTEGER", 1, nil, 1},
				{1, "email", "varchar(255)", 0, sql.NullString{}, 0},
			},
		},
		{
			dialect: "mysql",
			query:   compiledColumnsQuery("mysql"),
			columns: []string{"Field", "Type", "Collation", "Null", "Key", "Default", "Extra", "Privileges", "Comment"},
			rows: [][]any{
				{"id", "bigint unsigned", nil, "NO", "PRI", nil, "auto_increment", "", ""},
				{"email", "varchar(255)", "utf8mb4_unicode_ci", "YES", "UNI", nil, "", "", "user email"},
			},
		},
		{
			dialect: "mariadb",
			query:   compiledColumnsQuery("mariadb"),
			columns: []string{"Field", "Type", "Collation", "Null", "Key", "Default", "Extra", "Privileges", "Comment"},
			rows: [][]any{
				{"guid", "uuid", nil, "NO", "", nil, "", "", ""},
			},
		},
		{
			dialect: "postgres",
			query:   compiledColumnsQuery("postgres"),
			rows: [][]any{
				{"id", "bigint", "NO", "nextval('users_id_seq'::regclass)", nil},
				{"email", "character varying", "YES", nil, "user email"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.dialect, func(t *testing.T) {
			builder, err := schema.NewBuilder(tt.dialect)
			require.NoError(t, err)
			ctx := newMockContext(tt.dialect)
			ctx.withRows(tt.query, tt.columns, tt.rows...)

			columns, err := builder.GetColumns(ctx, "users")
			require.NoError(t, err)
			require.Len(t, columns, len(tt.rows))
			if tt.dialect == "mysql" {
				assert.Equal(t, "bigint unsigned", columns[0].TypeFull)
				assert.Equal(t, "bigint unsigned", columns[0].TypeName)
				assert.True(t, columns[1].Nullable)
				assert.Equal(t, "varchar", columns[1].TypeName)
				assert.Equal(t, "user email", columns[1].Comment.String)
			}
			if tt.dialect == "mariadb" {
				assert.Equal(t, "uuid", columns[0].TypeName)
			}
			if tt.dialect == "postgres" {
				assert.Equal(t, "bigint", columns[0].TypeName)
				assert.False(t, columns[0].Nullable)
				assert.True(t, columns[1].Nullable)
				assert.Equal(t, "user email", columns[1].Comment.String)
			}
			if tt.dialect == "sqlite3" {
				assert.Equal(t, "INTEGER", columns[0].TypeName)
				assert.False(t, columns[0].Nullable)
				assert.True(t, columns[1].Nullable)
			}
		})
	}
}

func TestBuilderGetColumnsReturnsNilForMissingTable(t *testing.T) {
	for _, dialect := range []string{"sqlite3", "mysql", "mariadb", "postgres"} {
		t.Run(dialect, func(t *testing.T) {
			builder, err := schema.NewBuilder(dialect)
			require.NoError(t, err)
			ctx := newMockContext(dialect)
			ctx.row = &mockRow{err: sql.ErrNoRows}

			columns, err := builder.GetColumns(ctx, "missing")
			require.NoError(t, err)
			assert.Nil(t, columns)
			assert.Empty(t, ctx.queries)
		})
	}
}

func TestBuilderGetIndexesMapsRows(t *testing.T) {
	t.Run("sqlite3", func(t *testing.T) {
		builder, err := schema.NewBuilder("sqlite3")
		require.NoError(t, err)
		ctx := newMockContext("sqlite3")
		ctx.withRows(`PRAGMA index_list("orders")`, nil,
			[]any{0, "idx_orders_company_user", 0, "c", 0},
			[]any{1, "sqlite_autoindex_orders_1", 1, "pk", 0},
		)
		ctx.withRows(`PRAGMA index_info("idx_orders_company_user")`, nil,
			[]any{0, 1, "company_id"},
			[]any{1, 2, "user_id"},
		)
		ctx.withRows(`PRAGMA index_info("sqlite_autoindex_orders_1")`, nil,
			[]any{0, 0, "id"},
		)

		indexes, err := builder.GetIndexes(ctx, "orders")
		require.NoError(t, err)
		require.Len(t, indexes, 2)
		assert.Equal(t, []string{"company_id", "user_id"}, indexes[0].Columns)
		assert.False(t, indexes[0].Unique)
		assert.True(t, indexes[1].Primary)
	})

	for _, dialect := range []string{"mysql", "mariadb"} {
		t.Run(dialect, func(t *testing.T) {
			builder, err := schema.NewBuilder(dialect)
			require.NoError(t, err)
			ctx := newMockContext(dialect)
			ctx.withRows(compiledIndexesQuery(dialect, "orders"), []string{"Non_unique", "Key_name", "Column_name"},
				[]any{0, "PRIMARY", "id"},
				[]any{1, "idx_orders_company_user", "company_id"},
				[]any{1, "idx_orders_company_user", "user_id"},
			)

			indexes, err := builder.GetIndexes(ctx, "orders")
			require.NoError(t, err)
			require.Len(t, indexes, 2)
			assert.True(t, indexes[0].Primary)
			assert.Equal(t, []string{"company_id", "user_id"}, indexes[1].Columns)
			assert.False(t, indexes[1].Unique)
		})
	}

	t.Run("postgres", func(t *testing.T) {
		builder, err := schema.NewBuilder("postgres")
		require.NoError(t, err)
		ctx := newMockContext("postgres")
		ctx.withRows(compiledIndexesQuery("postgres", "orders"), nil,
			[]any{"orders_pkey", true, true, "id"},
			[]any{"idx_orders_company_user", false, false, "company_id"},
			[]any{"idx_orders_company_user", false, false, "user_id"},
		)

		indexes, err := builder.GetIndexes(ctx, "orders")
		require.NoError(t, err)
		require.Len(t, indexes, 2)
		assert.True(t, indexes[0].Primary)
		assert.Equal(t, []string{"company_id", "user_id"}, indexes[1].Columns)
		assert.False(t, indexes[1].Unique)
	})
}

func TestBuilderGetTablesMapsRows(t *testing.T) {
	tests := []struct {
		dialect string
		query   string
		rows    [][]any
		assert  func(*testing.T, []*schema.TableInfo)
	}{
		{
			dialect: "sqlite3",
			query:   compiledTablesQuery("sqlite3"),
			rows:    [][]any{{"users"}},
			assert: func(t *testing.T, tables []*schema.TableInfo) {
				assert.Equal(t, "users", tables[0].Name)
			},
		},
		{
			dialect: "mysql",
			query:   compiledTablesQuery("mysql"),
			rows:    [][]any{{"users", "application users"}},
			assert: func(t *testing.T, tables []*schema.TableInfo) {
				assert.Equal(t, "users", tables[0].Name)
				assert.Equal(t, "application users", tables[0].Comment.String)
			},
		},
		{
			dialect: "mariadb",
			query:   compiledTablesQuery("mariadb"),
			rows:    [][]any{{"users", "application users"}},
			assert: func(t *testing.T, tables []*schema.TableInfo) {
				assert.Equal(t, "users", tables[0].Name)
				assert.Equal(t, "application users", tables[0].Comment.String)
			},
		},
		{
			dialect: "postgres",
			query:   compiledTablesQuery("postgres"),
			rows:    [][]any{{"users", "public", "application users"}},
			assert: func(t *testing.T, tables []*schema.TableInfo) {
				assert.Equal(t, "users", tables[0].Name)
				assert.Equal(t, "public", tables[0].Schema)
				assert.Equal(t, "application users", tables[0].Comment.String)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.dialect, func(t *testing.T) {
			builder, err := schema.NewBuilder(tt.dialect)
			require.NoError(t, err)
			ctx := newMockContext(tt.dialect)
			ctx.withRows(tt.query, nil, tt.rows...)

			tables, err := builder.GetTables(ctx)
			require.NoError(t, err)
			require.Len(t, tables, len(tt.rows))
			tt.assert(t, tables)
		})
	}
}

func TestBuilderHasHelpersUseMappedMetadata(t *testing.T) {
	builder, err := schema.NewBuilder("postgres")
	require.NoError(t, err)
	ctx := newMockContext("postgres")
	ctx.withRows(compiledColumnsQuery("postgres"), nil,
		[]any{"id", "bigint", "NO", nil, nil},
		[]any{"email", "character varying", "YES", nil, nil},
	)
	ctx.withRows(compiledIndexesQuery("postgres", "users"), nil,
		[]any{"idx_users_email", false, false, "email"},
	)

	exists, err := builder.HasColumn(ctx, "users", "email")
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = builder.HasColumns(ctx, "users", []string{"id", "missing"})
	require.NoError(t, err)
	assert.False(t, exists)

	exists, err = builder.HasIndex(ctx, "users", []string{"idx_users_email"})
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = builder.HasTable(ctx, "users")
	require.NoError(t, err)
	assert.True(t, exists)
}

func errOnly(_ any, err error) error {
	return err
}

func compiledColumnsQuery(dialect string) string {
	grammar, err := schema.NewGrammar(dialect)
	if err != nil {
		panic(err)
	}
	query, err := grammar.CompileColumns("", "users")
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(query)
}

func compiledIndexesQuery(dialect, table string) string {
	grammar, err := schema.NewGrammar(dialect)
	if err != nil {
		panic(err)
	}
	query, err := grammar.CompileIndexes("", table)
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(query)
}

func compiledTablesQuery(dialect string) string {
	grammar, err := schema.NewGrammar(dialect)
	if err != nil {
		panic(err)
	}
	query, err := grammar.CompileTables("")
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(query)
}
