package schema_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/afkdevs/go-schema"
	"github.com/stretchr/testify/suite"
)

func TestSchema(t *testing.T) {
	suite.Run(t, new(schemaTestSuite))
}

type schemaTestSuite struct {
	suite.Suite
	ctx context.Context
	db  *sql.DB
}

func (s *schemaTestSuite) SetupSuite() {
	ctx := context.Background()
	s.ctx = ctx

	config := parseTestConfig()

	dsn := fmt.Sprintf("host=localhost port=5432 user=%s password=%s dbname=%s sslmode=disable", config.Username, config.Password, config.Database)

	db, err := sql.Open("postgres", dsn)
	s.Require().NoError(err)

	err = db.Ping()
	s.Require().NoError(err)

	s.db = db

	s.Run("when dialect is not set should return error", func() {
		builderFuncs := []func() error{
			func() error { return schema.Create(ctx, nil, "", nil) },
			func() error { return schema.Drop(ctx, nil, "") },
			func() error { return schema.DropIfExists(ctx, nil, "") },
			func() error { _, err := schema.GetColumns(ctx, nil, ""); return err },
			func() error { _, err := schema.GetIndexes(ctx, nil, ""); return err },
			func() error { _, err := schema.GetTables(ctx, nil); return err },
			func() error { _, err := schema.HasColumn(ctx, nil, "", ""); return err },
			func() error { _, err := schema.HasColumns(ctx, nil, "", nil); return err },
			func() error { _, err := schema.HasIndex(ctx, nil, "", nil); return err },
			func() error { _, err := schema.HasTable(ctx, nil, ""); return err },
			func() error { return schema.Rename(ctx, nil, "", "") },
			func() error { return schema.Table(ctx, nil, "", nil) },
		}
		for _, fn := range builderFuncs {
			s.Error(fn(), "Expected error when dialect is not set")
		}
	})
	schema.SetDialect("postgres")
}

func (s *schemaTestSuite) TearDownSuite() {
	_ = s.db.Close()
}

func (s *schemaTestSuite) TestCreate() {
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	s.Run("when parameters are valid should create table", func() {
		err := schema.Create(s.ctx, tx, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name")
			table.String("email").Unique()
			table.String("password")
			table.Timestamp("created_at").UseCurrent()
			table.Timestamp("updated_at").UseCurrent()
		})
		s.NoError(err)
	})
	s.Run("when table already exists should return error", func() {
		err := schema.Create(s.ctx, tx, "users", func(table *schema.Blueprint) {
			table.ID()
		})
		s.Error(err)
		s.ErrorContains(err, "\"users\" already exists")
	})
	s.Run("when table name is empty should return error", func() {
		err := schema.Create(s.ctx, tx, "", func(table *schema.Blueprint) {
			table.ID()
		})
		s.Error(err)
		s.ErrorContains(err, "invalid arguments")
	})
	s.Run("when blueprint function is nil should return error", func() {
		err := schema.Create(s.ctx, tx, "test", nil)
		s.Error(err)
		s.ErrorContains(err, "invalid arguments")
	})
	s.Run("when transaction is nil should return error", func() {
		err := schema.Create(s.ctx, nil, "test", func(table *schema.Blueprint) {
			table.ID()
		})
		s.Error(err)
		s.ErrorContains(err, "invalid arguments")
	})
}

func (s *schemaTestSuite) TestDrop() {
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	s.Run("when parameters are valid should drop table", func() {
		err := schema.Create(s.ctx, tx, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name")
			table.String("email").Unique()
			table.String("password")
			table.Timestamp("created_at").UseCurrent()
			table.Timestamp("updated_at").UseCurrent()
		})
		s.Require().NoError(err)
		err = schema.Drop(s.ctx, tx, "users")
		s.NoError(err)
	})
	s.Run("when table does not exist should return error", func() {
		err := schema.Drop(s.ctx, tx, "non_existing_table")
		s.Error(err)
	})
	s.Run("when table name is empty should return error", func() {
		err := schema.Drop(s.ctx, tx, "")
		s.Error(err)
		s.ErrorContains(err, "invalid arguments")
	})
	s.Run("when transaction is nil should return error", func() {
		err := schema.Drop(s.ctx, nil, "test")
		s.Error(err)
		s.ErrorContains(err, "invalid arguments")
	})
}

func (s *schemaTestSuite) TestDropIfExists() {
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	s.Run("when parameters are valid should drop table", func() {
		err := schema.Create(s.ctx, tx, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name")
			table.String("email").Unique()
			table.String("password")
			table.Timestamp("created_at").UseCurrent()
			table.Timestamp("updated_at").UseCurrent()
		})
		s.Require().NoError(err)
		err = schema.DropIfExists(s.ctx, tx, "users")
		s.NoError(err)
	})
	s.Run("when table does not exist should return no error", func() {
		err := schema.DropIfExists(s.ctx, tx, "non_existing_table")
		s.NoError(err)
	})
	s.Run("when table name is empty should return error", func() {
		err := schema.DropIfExists(s.ctx, tx, "")
		s.Error(err)
		s.ErrorContains(err, "invalid arguments")
	})
	s.Run("when transaction is nil should return error", func() {
		err := schema.DropIfExists(s.ctx, nil, "test")
		s.Error(err)
		s.ErrorContains(err, "invalid arguments")
	})
}

func (s *schemaTestSuite) TestGetColumns() {
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	s.Run("when parameters are valid should return columns", func() {
		err := schema.Create(s.ctx, tx, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name")
			table.String("email").Unique()
			table.String("password")
			table.Timestamp("created_at").UseCurrent()
			table.Timestamp("updated_at").UseCurrent()
		})
		s.Require().NoError(err)
		columns, err := schema.GetColumns(s.ctx, tx, "users")
		s.NoError(err)
		s.NotEmpty(columns)
		s.Len(columns, 6)
	})
	s.Run("when table does not exist should empty columns", func() {
		columns, err := schema.GetColumns(s.ctx, tx, "non_existing_table")
		s.NoError(err)
		s.Nil(columns)
	})
	s.Run("when table name is empty should return error", func() {
		columns, err := schema.GetColumns(s.ctx, tx, "")
		s.Error(err)
		s.Nil(columns)
	})
	s.Run("when transaction is nil should return error", func() {
		columns, err := schema.GetColumns(s.ctx, nil, "test")
		s.Error(err)
		s.Nil(columns)
	})
}

func (s *schemaTestSuite) TestGetIndexes() {
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	s.Run("when parameters are valid should return indexes", func() {
		err := schema.Create(s.ctx, tx, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name")
			table.String("email").Unique()
			table.String("password")
			table.Timestamp("created_at").UseCurrent()
			table.Timestamp("updated_at").UseCurrent()
		})
		s.Require().NoError(err)
		indexes, err := schema.GetIndexes(s.ctx, tx, "users")
		s.NoError(err)
		s.NotEmpty(indexes)
		s.Len(indexes, 2) // Expecting the unique index on email and the primary key index on id
	})
	s.Run("when table does not exist should return empty indexes", func() {
		indexes, err := schema.GetIndexes(s.ctx, tx, "non_existing_table")
		s.NoError(err)
		s.Nil(indexes)
	})
	s.Run("when table name is empty should return error", func() {
		indexes, err := schema.GetIndexes(s.ctx, tx, "")
		s.Error(err)
		s.Nil(indexes)
	})
	s.Run("when transaction is nil should return error", func() {
		indexes, err := schema.GetIndexes(s.ctx, nil, "test")
		s.Error(err)
		s.Nil(indexes)
	})
}

func (s *schemaTestSuite) TestGetTables() {
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	s.Run("when no tables exist should return empty", func() {
		tables, err := schema.GetTables(s.ctx, tx)
		s.NoError(err)
		s.Empty(tables)
	})
	s.Run("when transaction is valid should return tables", func() {
		err := schema.Create(s.ctx, tx, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name")
			table.String("email").Unique()
			table.String("password")
			table.Timestamp("created_at").UseCurrent()
			table.Timestamp("updated_at").UseCurrent()
		})
		s.Require().NoError(err)
		tables, err := schema.GetTables(s.ctx, tx)
		s.NoError(err)
		s.NotEmpty(tables)
		s.Len(tables, 1) // Expecting at least the 'users' table created in previous tests
	})
	s.Run("when transaction is nil should return error", func() {
		tables, err := schema.GetTables(s.ctx, nil)
		s.Error(err)
		s.Nil(tables)
	})
}

func (s *schemaTestSuite) TestHasColumn() {
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	s.Run("when column exists should return true", func() {
		err := schema.Create(s.ctx, tx, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name")
			table.String("email").Unique()
			table.String("password")
			table.Timestamp("created_at").UseCurrent()
			table.Timestamp("updated_at").UseCurrent()
		})
		s.Require().NoError(err)

		exists, err := schema.HasColumn(s.ctx, tx, "users", "email")
		s.NoError(err)
		s.True(exists)
	})

	s.Run("when column does not exist should return false", func() {
		exists, err := schema.HasColumn(s.ctx, tx, "users", "non_existing_column")
		s.NoError(err)
		s.False(exists)
	})

	s.Run("when transaction is nil should return error", func() {
		exists, err := schema.HasColumn(s.ctx, nil, "users", "email")
		s.Error(err)
		s.False(exists)
	})
}

func (s *schemaTestSuite) TestHasColumns() {
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	s.Run("when all columns exist should return true", func() {
		err := schema.Create(s.ctx, tx, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name")
			table.String("email").Unique()
			table.String("password")
			table.Timestamp("created_at").UseCurrent()
			table.Timestamp("updated_at").UseCurrent()
		})
		s.Require().NoError(err)

		exists, err := schema.HasColumns(s.ctx, tx, "users", []string{"email", "name"})
		s.NoError(err)
		s.True(exists)
	})
	s.Run("when some columns do not exist should return false", func() {
		exists, err := schema.HasColumns(s.ctx, tx, "users", []string{"email", "non_existing_column"})
		s.NoError(err)
		s.False(exists)
	})
	s.Run("when no columns are provided should return error", func() {
		exists, err := schema.HasColumns(s.ctx, tx, "users", []string{})
		s.Error(err)
		s.False(exists)
	})
}

func (s *schemaTestSuite) TestHasIndex() {
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	s.Run("when index exists should return true", func() {
		err := schema.Create(s.ctx, tx, "users", func(table *schema.Blueprint) {
			table.ID()
			table.Integer("company_id")
			table.String("name")
			table.String("email").Unique("uk_users_email")
			table.String("password")
			table.Timestamp("created_at").UseCurrent()
			table.Timestamp("updated_at").UseCurrent()

			table.Index("company_id", "id")
		})
		s.Require().NoError(err)

		exists, err := schema.HasIndex(s.ctx, tx, "users", []string{"email"})
		s.NoError(err)
		s.True(exists)

		exists, err = schema.HasIndex(s.ctx, tx, "users", []string{"company_id", "id"})
		s.NoError(err)
		s.True(exists)

		exists, err = schema.HasIndex(s.ctx, tx, "users", []string{"uk_users_email"})
		s.NoError(err)
		s.True(exists)
	})
	s.Run("when index does not exist should return false", func() {
		exists, err := schema.HasIndex(s.ctx, tx, "users", []string{"non_existing_index"})
		s.NoError(err)
		s.False(exists)
	})
}

func (s *schemaTestSuite) TestHasTable() {
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	s.Run("when table exists should return true", func() {
		err := schema.Create(s.ctx, tx, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name")
			table.String("email").Unique()
			table.String("password")
			table.Timestamp("created_at").UseCurrent()
			table.Timestamp("updated_at").UseCurrent()
		})
		s.Require().NoError(err)

		exists, err := schema.HasTable(s.ctx, tx, "users")
		s.NoError(err)
		s.True(exists)
	})
	s.Run("when table does not exist should return false", func() {
		exists, err := schema.HasTable(s.ctx, tx, "non_existing_table")
		s.NoError(err)
		s.False(exists)
	})
	s.Run("when table name is empty should return error", func() {
		exists, err := schema.HasTable(s.ctx, tx, "")
		s.Error(err)
		s.False(exists)
	})
	s.Run("when transaction is nil should return error", func() {
		exists, err := schema.HasTable(s.ctx, nil, "users")
		s.Error(err)
		s.False(exists)
	})
}

func (s *schemaTestSuite) TestRename() {
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	s.Run("when parameters are valid should rename table", func() {
		err := schema.Create(s.ctx, tx, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name")
			table.String("email").Unique()
			table.String("password")
			table.Timestamp("created_at").UseCurrent()
			table.Timestamp("updated_at").UseCurrent()
		})
		s.Require().NoError(err)

		err = schema.Rename(s.ctx, tx, "users", "members")
		s.NoError(err)

		columns, err := schema.GetColumns(s.ctx, tx, "members")
		s.NoError(err)
		s.NotEmpty(columns)
		s.Len(columns, 6)
	})

	s.Run("when table does not exist should return error", func() {
		err := schema.Rename(s.ctx, tx, "non_existing_table", "new_name")
		s.Error(err)
	})

	s.Run("when new name is empty should return error", func() {
		err := schema.Rename(s.ctx, tx, "users", "")
		s.Error(err)
	})

	s.Run("when transaction is nil should return error", func() {
		err := schema.Rename(s.ctx, nil, "users", "new_name")
		s.Error(err)
	})
}

func (s *schemaTestSuite) TestTable() {
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	s.Run("when parameters are valid should alter table", func() {
		err := schema.Create(s.ctx, tx, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name")
			table.String("email").Unique()
			table.String("password")
			table.Timestamp("created_at").UseCurrent()
			table.Timestamp("updated_at").UseCurrent()
		})
		s.Require().NoError(err)

		err = schema.Table(s.ctx, tx, "users", func(table *schema.Blueprint) {
			table.String("phone").Nullable()
		})
		s.NoError(err)

		columns, err := schema.GetColumns(s.ctx, tx, "users")
		s.NoError(err)
		s.Len(columns, 7) // Expecting the new 'phone' column to be added
	})

	s.Run("when table does not exist should return error", func() {
		err := schema.Table(s.ctx, tx, "non_existing_table", func(table *schema.Blueprint) {
			table.String("new_column")
		})
		s.Error(err)
	})

	s.Run("when transaction is nil should return error", func() {
		err := schema.Table(s.ctx, nil, "users", func(table *schema.Blueprint) {
			table.String("new_column")
		})
		s.Error(err)
	})
}
