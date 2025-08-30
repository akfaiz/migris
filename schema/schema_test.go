package schema_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/afkdevs/migris"
	"github.com/afkdevs/migris/schema"
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
	err = migris.SetDialect("postgres")
	s.Require().NoError(err)
}

func (s *schemaTestSuite) TearDownSuite() {
	_ = s.db.Close()
}

func (s *schemaTestSuite) TestCreate() {
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	c := schema.NewContext(s.ctx, tx)

	s.Run("when parameters are valid should create table", func() {
		err := schema.Create(c, "users", func(table *schema.Blueprint) {
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
		err := schema.Create(c, "users", func(table *schema.Blueprint) {
			table.ID()
		})
		s.Error(err)
		s.ErrorContains(err, "\"users\" already exists")
	})
	s.Run("when table name is empty should return error", func() {
		err := schema.Create(c, "", func(table *schema.Blueprint) {
			table.ID()
		})
		s.Error(err)
		s.ErrorContains(err, "invalid arguments")
	})
	s.Run("when blueprint function is nil should return error", func() {
		err := schema.Create(c, "test", nil)
		s.Error(err)
		s.ErrorContains(err, "invalid arguments")
	})
}

func (s *schemaTestSuite) TestDrop() {
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	c := schema.NewContext(s.ctx, tx)

	s.Run("when parameters are valid should drop table", func() {
		err := schema.Create(c, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name")
			table.String("email").Unique()
			table.String("password")
			table.Timestamp("created_at").UseCurrent()
			table.Timestamp("updated_at").UseCurrent()
		})
		s.Require().NoError(err)
		err = schema.Drop(c, "users")
		s.NoError(err)
	})
	s.Run("when table does not exist should return error", func() {
		err := schema.Drop(c, "non_existing_table")
		s.Error(err)
	})
	s.Run("when table name is empty should return error", func() {
		err := schema.Drop(c, "")
		s.Error(err)
		s.ErrorContains(err, "invalid arguments")
	})
	s.Run("when context is nil should return error", func() {
		err := schema.Drop(nil, "test")
		s.Error(err)
		s.ErrorContains(err, "invalid arguments")
	})
}

func (s *schemaTestSuite) TestDropIfExists() {
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	c := schema.NewContext(s.ctx, tx)

	s.Run("when parameters are valid should drop table", func() {
		err := schema.Create(c, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name")
			table.String("email").Unique()
			table.String("password")
			table.Timestamp("created_at").UseCurrent()
			table.Timestamp("updated_at").UseCurrent()
		})
		s.Require().NoError(err)
		err = schema.DropIfExists(c, "users")
		s.NoError(err)
	})
	s.Run("when table does not exist should return no error", func() {
		err := schema.DropIfExists(c, "non_existing_table")
		s.NoError(err)
	})
	s.Run("when table name is empty should return error", func() {
		err := schema.DropIfExists(c, "")
		s.Error(err)
		s.ErrorContains(err, "invalid arguments")
	})
	s.Run("when context is nil should return error", func() {
		err := schema.DropIfExists(nil, "test")
		s.Error(err)
		s.ErrorContains(err, "invalid arguments")
	})
}

func (s *schemaTestSuite) TestGetColumns() {
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	c := schema.NewContext(s.ctx, tx)

	s.Run("when parameters are valid should return columns", func() {
		err := schema.Create(c, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name")
			table.String("email").Unique()
			table.String("password")
			table.Timestamp("created_at").UseCurrent()
			table.Timestamp("updated_at").UseCurrent()
		})
		s.Require().NoError(err)
		columns, err := schema.GetColumns(c, "users")
		s.NoError(err)
		s.NotEmpty(columns)
		s.Len(columns, 6)
	})
	s.Run("when table does not exist should empty columns", func() {
		columns, err := schema.GetColumns(c, "non_existing_table")
		s.NoError(err)
		s.Nil(columns)
	})
	s.Run("when table name is empty should return error", func() {
		columns, err := schema.GetColumns(c, "")
		s.Error(err)
		s.Nil(columns)
	})
	s.Run("when context is nil should return error", func() {
		columns, err := schema.GetColumns(nil, "test")
		s.Error(err)
		s.Nil(columns)
	})
}

func (s *schemaTestSuite) TestGetIndexes() {
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	c := schema.NewContext(s.ctx, tx)

	s.Run("when parameters are valid should return indexes", func() {
		err := schema.Create(c, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name")
			table.String("email").Unique()
			table.String("password")
			table.Timestamp("created_at").UseCurrent()
			table.Timestamp("updated_at").UseCurrent()
		})
		s.Require().NoError(err)
		indexes, err := schema.GetIndexes(c, "users")
		s.NoError(err)
		s.NotEmpty(indexes)
		s.Len(indexes, 2) // Expecting the unique index on email and the primary key index on id
	})
	s.Run("when table does not exist should return empty indexes", func() {
		indexes, err := schema.GetIndexes(c, "non_existing_table")
		s.NoError(err)
		s.Nil(indexes)
	})
	s.Run("when table name is empty should return error", func() {
		indexes, err := schema.GetIndexes(c, "")
		s.Error(err)
		s.Nil(indexes)
	})
	s.Run("when context is nil should return error", func() {
		indexes, err := schema.GetIndexes(nil, "test")
		s.Error(err)
		s.Nil(indexes)
	})
}

func (s *schemaTestSuite) TestGetTables() {
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	c := schema.NewContext(s.ctx, tx)

	s.Run("when no tables exist should return empty", func() {
		tables, err := schema.GetTables(c)
		s.NoError(err)
		s.Empty(tables)
	})
	s.Run("when transaction is valid should return tables", func() {
		err := schema.Create(c, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name")
			table.String("email").Unique()
			table.String("password")
			table.Timestamp("created_at").UseCurrent()
			table.Timestamp("updated_at").UseCurrent()
		})
		s.Require().NoError(err)
		tables, err := schema.GetTables(c)
		s.NoError(err)
		s.NotEmpty(tables)
		s.Len(tables, 1) // Expecting at least the 'users' table created in previous tests
	})
	s.Run("when context is nil should return error", func() {
		tables, err := schema.GetTables(nil)
		s.Error(err)
		s.Nil(tables)
	})
}

func (s *schemaTestSuite) TestHasColumn() {
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	c := schema.NewContext(s.ctx, tx)

	s.Run("when column exists should return true", func() {
		err := schema.Create(c, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name")
			table.String("email").Unique()
			table.String("password")
			table.Timestamp("created_at").UseCurrent()
			table.Timestamp("updated_at").UseCurrent()
		})
		s.Require().NoError(err)

		exists, err := schema.HasColumn(c, "users", "email")
		s.NoError(err)
		s.True(exists)
	})

	s.Run("when column does not exist should return false", func() {
		exists, err := schema.HasColumn(c, "users", "non_existing_column")
		s.NoError(err)
		s.False(exists)
	})

	s.Run("when context is nil should return error", func() {
		exists, err := schema.HasColumn(nil, "users", "email")
		s.Error(err)
		s.False(exists)
	})
}

func (s *schemaTestSuite) TestHasColumns() {
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	c := schema.NewContext(s.ctx, tx)

	s.Run("when all columns exist should return true", func() {
		err := schema.Create(c, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name")
			table.String("email").Unique()
			table.String("password")
			table.Timestamp("created_at").UseCurrent()
			table.Timestamp("updated_at").UseCurrent()
		})
		s.Require().NoError(err)

		exists, err := schema.HasColumns(c, "users", []string{"email", "name"})
		s.NoError(err)
		s.True(exists)
	})
	s.Run("when some columns do not exist should return false", func() {
		exists, err := schema.HasColumns(c, "users", []string{"email", "non_existing_column"})
		s.NoError(err)
		s.False(exists)
	})
	s.Run("when no columns are provided should return error", func() {
		exists, err := schema.HasColumns(c, "users", []string{})
		s.Error(err)
		s.False(exists)
	})
}

func (s *schemaTestSuite) TestHasIndex() {
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	c := schema.NewContext(s.ctx, tx)

	s.Run("when index exists should return true", func() {
		err := schema.Create(c, "users", func(table *schema.Blueprint) {
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

		exists, err := schema.HasIndex(c, "users", []string{"email"})
		s.NoError(err)
		s.True(exists)

		exists, err = schema.HasIndex(c, "users", []string{"company_id", "id"})
		s.NoError(err)
		s.True(exists)

		exists, err = schema.HasIndex(c, "users", []string{"uk_users_email"})
		s.NoError(err)
		s.True(exists)
	})
	s.Run("when index does not exist should return false", func() {
		exists, err := schema.HasIndex(c, "users", []string{"non_existing_index"})
		s.NoError(err)
		s.False(exists)
	})
}

func (s *schemaTestSuite) TestHasTable() {
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	c := schema.NewContext(s.ctx, tx)

	s.Run("when table exists should return true", func() {
		err := schema.Create(c, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name")
			table.String("email").Unique()
			table.String("password")
			table.Timestamp("created_at").UseCurrent()
			table.Timestamp("updated_at").UseCurrent()
		})
		s.Require().NoError(err)

		exists, err := schema.HasTable(c, "users")
		s.NoError(err)
		s.True(exists)
	})
	s.Run("when table does not exist should return false", func() {
		exists, err := schema.HasTable(c, "non_existing_table")
		s.NoError(err)
		s.False(exists)
	})
	s.Run("when table name is empty should return error", func() {
		exists, err := schema.HasTable(c, "")
		s.Error(err)
		s.False(exists)
	})
	s.Run("when context is nil should return error", func() {
		exists, err := schema.HasTable(nil, "users")
		s.Error(err)
		s.False(exists)
	})
}

func (s *schemaTestSuite) TestRename() {
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	c := schema.NewContext(s.ctx, tx)

	s.Run("when parameters are valid should rename table", func() {
		err := schema.Create(c, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name")
			table.String("email").Unique()
			table.String("password")
			table.Timestamp("created_at").UseCurrent()
			table.Timestamp("updated_at").UseCurrent()
		})
		s.Require().NoError(err)

		err = schema.Rename(c, "users", "members")
		s.NoError(err)

		columns, err := schema.GetColumns(c, "members")
		s.NoError(err)
		s.NotEmpty(columns)
		s.Len(columns, 6)
	})

	s.Run("when table does not exist should return error", func() {
		err := schema.Rename(c, "non_existing_table", "new_name")
		s.Error(err)
	})

	s.Run("when new name is empty should return error", func() {
		err := schema.Rename(c, "users", "")
		s.Error(err)
	})

	s.Run("when context is nil should return error", func() {
		err := schema.Rename(nil, "users", "new_name")
		s.Error(err)
	})
}

func (s *schemaTestSuite) TestTable() {
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	c := schema.NewContext(s.ctx, tx)

	s.Run("when parameters are valid should alter table", func() {
		err := schema.Create(c, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name")
			table.String("email").Unique()
			table.String("password")
			table.Timestamp("created_at").UseCurrent()
			table.Timestamp("updated_at").UseCurrent()
		})
		s.Require().NoError(err)

		err = schema.Table(c, "users", func(table *schema.Blueprint) {
			table.String("phone").Nullable()
		})
		s.NoError(err)

		columns, err := schema.GetColumns(c, "users")
		s.NoError(err)
		s.Len(columns, 7) // Expecting the new 'phone' column to be added
	})

	s.Run("when table does not exist should return error", func() {
		err := schema.Table(c, "non_existing_table", func(table *schema.Blueprint) {
			table.String("new_column")
		})
		s.Error(err)
	})

	s.Run("when context is nil should return error", func() {
		err := schema.Table(nil, "users", func(table *schema.Blueprint) {
			table.String("new_column")
		})
		s.Error(err)
	})
}
