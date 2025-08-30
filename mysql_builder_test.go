package schema_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/afkdevs/go-schema"
	_ "github.com/go-sql-driver/mysql" // MySQL driver
	"github.com/stretchr/testify/suite"
)

func TestMysqlBuilderSuite(t *testing.T) {
	suite.Run(t, new(mysqlBuilderSuite))
}

type mysqlBuilderSuite struct {
	suite.Suite
	ctx     context.Context
	db      *sql.DB
	builder schema.Builder
}

func (s *mysqlBuilderSuite) SetupSuite() {
	s.ctx = context.Background()

	config := parseTestConfig()

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local",
		config.Username,
		config.Password,
		"localhost",
		3306,
		config.Database,
	)

	db, err := sql.Open("mysql", dsn)
	s.Require().NoError(err)

	err = db.Ping()
	s.Require().NoError(err)

	s.db = db
	s.builder, err = schema.NewBuilder("mysql")
	s.Require().NoError(err)
}

func (s *mysqlBuilderSuite) TearDownSuite() {
	_ = s.db.Close()
}

func (s *mysqlBuilderSuite) AfterTest(_, _ string) {
	builder := s.builder
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	tables, err := builder.GetTables(s.ctx, tx)
	s.Require().NoError(err)
	for _, table := range tables {
		err := builder.DropIfExists(s.ctx, tx, table.Name)
		if err != nil {
			s.T().Logf("error dropping table %s: %v", table.Name, err)
		}
	}
	err = tx.Commit()
	s.Require().NoError(err, "expected no error when committing transaction after dropping tables")
}

func (s *mysqlBuilderSuite) TestCreate() {
	builder := s.builder
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	s.Run("when tx is nil, should return error", func() {
		err := builder.Create(s.ctx, nil, "test_table", func(table *schema.Blueprint) {
			table.String("name")
		})
		s.Error(err)
	})
	s.Run("when table name is empty, should return error", func() {
		err := builder.Create(s.ctx, tx, "", func(table *schema.Blueprint) {
			table.String("name")
		})
		s.Error(err)
	})
	s.Run("when blueprint is nil, should return error", func() {
		err := builder.Create(s.ctx, tx, "test_table", nil)
		s.Error(err)
	})
	s.Run("when all parameters are valid, should create table successfully", func() {
		err = builder.Create(context.Background(), tx, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name", 255)
			table.String("email", 255).Unique()
			table.String("password", 255).Nullable()
			table.Timestamps()
		})
		s.NoError(err, "expected no error when creating table with valid parameters")
	})
	s.Run("when have composite primary key should create it successfully", func() {
		err = builder.Create(context.Background(), tx, "user_roles", func(table *schema.Blueprint) {
			table.Integer("user_id")
			table.Integer("role_id")

			table.Primary("user_id", "role_id")
		})
		s.NoError(err, "expected no error when creating table with composite primary key")
	})
	s.Run("when have foreign key should create it successfully", func() {
		err = builder.Create(context.Background(), tx, "orders", func(table *schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("user_id")
			table.String("order_id", 255).Unique()
			table.Decimal("amount", 10, 2)
			table.Timestamp("created_at").UseCurrent()

			table.Foreign("user_id").References("id").On("users").OnDelete("CASCADE").OnUpdate("CASCADE")
		})
		s.NoError(err, "expected no error when creating table with foreign key")
	})
	s.Run("when have custom index should create it successfully", func() {
		err = builder.Create(context.Background(), tx, "orders_2", func(table *schema.Blueprint) {
			table.ID()
			table.String("order_id", 255).Unique("uk_orders_2_order_id")
			table.Decimal("amount", 10, 2)
			table.Timestamp("created_at").UseCurrent()

			table.Index("created_at").Name("idx_orders_created_at").Algorithm("BTREE")
		})
		s.NoError(err, "expected no error when creating table with custom index")
	})
	s.Run("when table already exists, should return error", func() {
		err = builder.Create(context.Background(), tx, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name", 255)
			table.String("email", 255).Unique()
		})
		s.Error(err, "expected error when creating table that already exists")
	})
}

func (s *mysqlBuilderSuite) TestDrop() {
	builder := s.builder
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	s.Run("when tx is nil, should return error", func() {
		err := builder.Drop(s.ctx, nil, "test_table")
		s.Error(err)
	})
	s.Run("when table name is empty, should return error", func() {
		err := builder.Drop(s.ctx, tx, "")
		s.Error(err)
	})
	s.Run("when all parameters are valid, should drop table successfully", func() {
		err = builder.Create(context.Background(), tx, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name", 255)
			table.String("email", 255).Unique()
			table.String("password", 255).Nullable()
			table.Timestamps()
		})
		s.NoError(err, "expected no error when creating table before dropping it")
		err = builder.Drop(context.Background(), tx, "users")
		s.NoError(err, "expected no error when dropping table with valid parameters")
	})
	s.Run("when table does not exist, should return error", func() {
		err = builder.Drop(context.Background(), tx, "non_existent_table")
		s.Error(err, "expected error when dropping a table that does not exist")
	})
}

func (s *mysqlBuilderSuite) TestDropIfExists() {
	builder := s.builder
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	s.Run("when tx is nil, should return error", func() {
		err := builder.DropIfExists(s.ctx, nil, "test_table")
		s.Error(err)
	})
	s.Run("when table name is empty, should return error", func() {
		err := builder.DropIfExists(s.ctx, tx, "")
		s.Error(err)
	})
	s.Run("when all parameters are valid, should drop table successfully", func() {
		err = builder.Create(context.Background(), tx, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name", 255)
			table.String("email", 255).Unique()
			table.String("password", 255).Nullable()
			table.Timestamps()
		})
		s.NoError(err, "expected no error when creating table before dropping it")
		err = builder.DropIfExists(context.Background(), tx, "users")
		s.NoError(err, "expected no error when dropping table with valid parameters")
	})
	s.Run("when table does not exist, should not return error", func() {
		err = builder.DropIfExists(context.Background(), tx, "non_existent_table")
		s.NoError(err, "expected no error when dropping a table that does not exist with DropIfExists")
	})
}

func (s *mysqlBuilderSuite) TestRename() {
	builder := s.builder
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	s.Run("when tx is nil, should return error", func() {
		err := builder.Rename(s.ctx, nil, "old_table", "new_table")
		s.Error(err)
	})
	s.Run("when old table name is empty, should return error", func() {
		err := builder.Rename(s.ctx, tx, "", "new_table")
		s.Error(err)
	})
	s.Run("when new table name is empty, should return error", func() {
		err := builder.Rename(s.ctx, tx, "old_table", "")
		s.Error(err)
	})
	s.Run("when all parameters are valid, should rename table successfully", func() {
		err = builder.Create(context.Background(), tx, "old_table", func(table *schema.Blueprint) {
			table.ID()
			table.String("name", 255)
		})
		s.NoError(err, "expected no error when creating old_table before renaming it")
		err = builder.Rename(context.Background(), tx, "old_table", "new_table")
		s.NoError(err, "expected no error when renaming table with valid parameters")
	})
}

func (s *mysqlBuilderSuite) TestTable() {
	builder := s.builder
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	s.Run("when tx is nil, should return error", func() {
		err := builder.Table(s.ctx, nil, "test_table", func(table *schema.Blueprint) {
			table.String("name")
		})
		s.Error(err)
	})
	s.Run("when table name is empty, should return error", func() {
		err := builder.Table(s.ctx, tx, "", func(table *schema.Blueprint) {
			table.String("name")
		})
		s.Error(err)
	})
	s.Run("when blueprint is nil, should return error", func() {
		err := builder.Table(s.ctx, tx, "test_table", nil)
		s.Error(err)
	})
	s.Run("when all parameters are valid, should modify table successfully", func() {
		err = builder.Create(s.ctx, tx, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name", 255)
			table.String("email", 255).Unique("uk_users_email")
			table.String("password", 255).Nullable()
			table.Text("bio").Nullable()
			table.Timestamps()

			table.FullText("bio")
		})
		s.NoError(err, "expected no error when creating table before modifying it")

		s.Run("should add new columns and modify existing ones", func() {
			err = builder.Table(s.ctx, tx, "users", func(table *schema.Blueprint) {
				table.String("address", 255).Nullable()
				table.String("phone", 20).Nullable().Unique("uk_users_phone")
			})
			s.NoError(err, "expected no error when modifying table with valid parameters")
		})
		s.Run("should modify existing column", func() {
			err = builder.Table(s.ctx, tx, "users", func(table *schema.Blueprint) {
				table.String("email", 255).Nullable().Change()
			})
			s.NoError(err, "expected no error when modifying existing column")
		})
		s.Run("should drop column and rename existing one", func() {
			err = builder.Table(s.ctx, tx, "users", func(table *schema.Blueprint) {
				table.DropColumn("password")
				table.RenameColumn("name", "full_name")
			})
			s.NoError(err, "expected no error when dropping column and renaming existing one")
		})
		s.Run("should add index", func() {
			err = builder.Table(s.ctx, tx, "users", func(table *schema.Blueprint) {
				table.Index("phone").Name("idx_users_phone").Algorithm("BTREE")
			})
			s.NoError(err, "expected no error when adding index to table")
		})
		s.Run("should rename index", func() {
			err = builder.Table(s.ctx, tx, "users", func(table *schema.Blueprint) {
				table.RenameIndex("idx_users_phone", "idx_users_contact")
			})
			s.NoError(err, "expected no error when renaming index in table")
		})
		s.Run("should drop index", func() {
			err = builder.Table(s.ctx, tx, "users", func(table *schema.Blueprint) {
				table.DropIndex("idx_users_contact")
			})
			s.NoError(err, "expected no error when dropping index from table")
		})
		s.Run("should drop unique constraint", func() {
			err = builder.Table(s.ctx, tx, "users", func(table *schema.Blueprint) {
				table.DropUnique("uk_users_email")
			})
			s.NoError(err, "expected no error when dropping unique constraint from table")
		})
		s.Run("should drop fulltext index", func() {
			err = builder.Table(s.ctx, tx, "users", func(table *schema.Blueprint) {
				table.DropFulltext("ft_users_bio")
			})
			s.NoError(err, "expected no error when dropping fulltext index from table")
		})
		s.Run("should add foreign key", func() {
			err = builder.Create(s.ctx, tx, "roles", func(table *schema.Blueprint) {
				table.UnsignedInteger("id").Primary()
				table.String("role_name", 255).Unique("uk_roles_role_name")
			})
			s.NoError(err, "expected no error when creating roles table before adding foreign key")
			err = builder.Table(s.ctx, tx, "users", func(table *schema.Blueprint) {
				table.UnsignedInteger("role_id").Nullable()
				table.Foreign("role_id").References("id").On("roles").OnDelete("SET NULL").OnUpdate("CASCADE")
			})
			s.NoError(err, "expected no error when adding foreign key to users table")
		})
		s.Run("should drop foreign key", func() {
			err = builder.Table(s.ctx, tx, "users", func(table *schema.Blueprint) {
				table.DropForeign("fk_users_roles")
			})
			s.NoError(err, "expected no error when dropping foreign key from users table")
		})
		s.Run("should drop primary key", func() {
			err = builder.Table(s.ctx, tx, "users", func(table *schema.Blueprint) {
				table.UnsignedBigInteger("id").Change()
				table.DropPrimary("users_pkey")
			})
			s.NoError(err, "expected no error when dropping primary key from users table")
		})
	})
}

func (s *mysqlBuilderSuite) TestGetColumns() {
	builder := s.builder
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	s.Run("when tx is nil, should return error", func() {
		columns, err := builder.GetColumns(s.ctx, nil, "test_table")
		s.Error(err)
		s.Nil(columns)
	})
	s.Run("when table name is empty, should return error", func() {
		columns, err := builder.GetColumns(s.ctx, tx, "")
		s.Error(err)
		s.Nil(columns)
	})
	s.Run("when table does not exist, should return empty slice", func() {
		columns, err := builder.GetColumns(s.ctx, tx, "non_existent_table")
		s.NoError(err)
		s.Empty(columns)
	})
	s.Run("when table exists, should return columns successfully", func() {
		err = builder.Create(context.Background(), tx, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name", 255)
			table.String("email", 255).Unique()
			table.String("password", 255).Nullable()
			table.Timestamps()
		})
		s.NoError(err, "expected no error when creating table before getting columns")

		columns, err := builder.GetColumns(context.Background(), tx, "users")
		s.NoError(err, "expected no error when getting columns from existing table")
		s.NotEmpty(columns)
		s.Len(columns, 6, "expected 6 columns in the users table")
	})
}

func (s *mysqlBuilderSuite) TestGetIndexes() {
	builder := s.builder
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	s.Run("when tx is nil, should return error", func() {
		_, err := builder.GetIndexes(s.ctx, nil, "users_indexes")
		s.Error(err, "expected error when transaction is nil")
	})
	s.Run("when table name is empty, should return error", func() {
		_, err := builder.GetIndexes(s.ctx, tx, "")
		s.Error(err, "expected error when table name is empty")
	})
	s.Run("when all parameters are valid", func() {
		err = builder.Create(s.ctx, tx, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name", 255)
			table.String("email", 255).Unique()
			table.String("password", 255).Nullable()
			table.Timestamps()

			table.Index("name").Name("idx_users_name")
		})
		s.NoError(err, "expected no error when creating table before getting indexes")

		indexes, err := builder.GetIndexes(s.ctx, tx, "users")
		s.NoError(err, "expected no error when getting indexes with valid parameters")
		s.Len(indexes, 3, "expected 3 index to be returned")

	})
	s.Run("when table does not exist, should return empty indexes", func() {
		indexes, err := builder.GetIndexes(s.ctx, tx, "non_existent_table")
		s.NoError(err, "expected no error when getting indexes of non-existent table")
		s.Empty(indexes, "expected empty indexes for non-existent table")
	})
}

func (s *mysqlBuilderSuite) TestGetTables() {
	builder := s.builder
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	s.Run("when tx is nil, should return error", func() {
		tables, err := builder.GetTables(s.ctx, nil)
		s.Error(err, "expected error when transaction is nil")
		s.Nil(tables)
	})
	s.Run("when all parameters are valid", func() {
		err = builder.Create(context.Background(), tx, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name", 255)
			table.String("email", 255).Unique()
			table.String("password", 255).Nullable()
			table.Timestamps()
		})
		s.NoError(err, "expected no error when creating table before getting tables")

		tables, err := builder.GetTables(context.Background(), tx)
		s.NoError(err, "expected no error when getting tables after creating one")
		s.NotEmpty(tables, "expected non-empty tables slice after creating a table")
		found := false
		for _, table := range tables {
			if table.Name == "users" {
				found = true
				break
			}
		}
		s.True(found, "expected users to be in the list of tables")
	})
}

func (s *mysqlBuilderSuite) TestHasColumn() {
	builder := s.builder
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	s.Run("when tx is nil, should return error", func() {
		exists, err := builder.HasColumn(s.ctx, nil, "users", "name")
		s.Error(err, "expected error when transaction is nil")
		s.False(exists)
	})
	s.Run("when table name is empty, should return error", func() {
		exists, err := builder.HasColumn(s.ctx, tx, "", "name")
		s.Error(err, "expected error when table name is empty")
		s.False(exists)
	})
	s.Run("when column name is empty, should return error", func() {
		exists, err := builder.HasColumn(s.ctx, tx, "users", "")
		s.Error(err, "expected error when column name is empty")
		s.False(exists)
	})
	s.Run("when all parameters are valid", func() {
		err = builder.Create(context.Background(), tx, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name", 255)
			table.String("email", 255).Unique()
			table.String("password", 255).Nullable()
			table.Timestamps()
		})
		s.NoError(err, "expected no error when creating table before checking for column existence")

		exists, err := builder.HasColumn(context.Background(), tx, "users", "name")
		s.NoError(err, "expected no error when checking for existing column")
		s.True(exists, "expected 'name' column to exist in users table")

		exists, err = builder.HasColumn(context.Background(), tx, "users", "non_existent_column")
		s.NoError(err, "expected no error when checking for non-existing column")
		s.False(exists, "expected 'non_existent_column' to not exist in users table")
	})
}

func (s *mysqlBuilderSuite) TestHasColumns() {
	builder := s.builder
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	s.Run("when tx is nil, should return error", func() {
		exists, err := builder.HasColumns(s.ctx, nil, "users", []string{"name"})
		s.Error(err, "expected error when transaction is nil")
		s.False(exists)
	})
	s.Run("when table name is empty, should return error", func() {
		exists, err := builder.HasColumns(s.ctx, tx, "", []string{"name"})
		s.Error(err, "expected error when table name is empty")
		s.False(exists)
	})
	s.Run("when column names are empty, should return error", func() {
		exists, err := builder.HasColumns(s.ctx, tx, "users", []string{})
		s.Error(err, "expected error when column names are empty")
		s.False(exists)
	})
	s.Run("when all parameters are valid", func() {
		err = builder.Create(context.Background(), tx, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name", 255)
			table.String("email", 255).Unique()
			table.String("password", 255).Nullable()
			table.Timestamps()
		})
		s.NoError(err, "expected no error when creating table before checking for columns existence")

		exists, err := builder.HasColumns(context.Background(), tx, "users", []string{"name", "email"})
		s.NoError(err, "expected no error when checking for existing columns")
		s.True(exists, "expected 'name' and 'email' columns to exist in users_has_columns table")

		exists, err = builder.HasColumns(context.Background(), tx, "users", []string{"name", "non_existent_column"})
		s.NoError(err, "expected no error when checking for mixed existing and non-existing columns")
		s.False(exists, "expected 'non_existent_column' to not exist in users_has_columns table")
	})
}

func (s *mysqlBuilderSuite) TestHasIndex() {
	builder := s.builder
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	s.Run("when tx is nil, should return error", func() {
		exists, err := builder.HasIndex(s.ctx, nil, "orders", []string{"idx_users_name"})
		s.Error(err, "expected error when transaction is nil")
		s.False(exists, "expected exists to be false when transaction is nil")
	})
	s.Run("when table name is empty, should return error", func() {
		exists, err := builder.HasIndex(s.ctx, tx, "", []string{"idx_users_name"})
		s.Error(err, "expected error when table name is empty")
		s.False(exists, "expected exists to be false when table name is empty")
	})
	s.Run("when all parameters are valid", func() {
		err = builder.Create(s.ctx, tx, "orders", func(table *schema.Blueprint) {
			table.ID()
			table.Integer("company_id")
			table.Integer("user_id")
			table.String("order_id", 255)
			table.Decimal("amount", 10, 2)
			table.Timestamps()

			table.Index("company_id", "user_id")
			table.Unique("order_id").Name("uk_orders3_order_id").Algorithm("BTREE")
		})
		s.NoError(err, "expected no error when creating table with index")

		exists, err := builder.HasIndex(s.ctx, tx, "orders", []string{"uk_orders3_order_id"})
		s.NoError(err, "expected no error when checking if index exists with valid parameters")
		s.True(exists, "expected exists to be true for existing index")

		exists, err = builder.HasIndex(s.ctx, tx, "orders", []string{"company_id", "user_id"})
		s.NoError(err, "expected no error when checking non-existent index")
		s.True(exists, "expected exists to be true for existing composite index")

		exists, err = builder.HasIndex(s.ctx, tx, "orders", []string{"non_existent_index"})
		s.NoError(err, "expected no error when checking non-existent index")
		s.False(exists, "expected exists to be false for non-existent index")
	})
}

func (s *mysqlBuilderSuite) TestHasTable() {
	builder := s.builder
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	s.Run("when tx is nil, should return error", func() {
		exists, err := builder.HasTable(s.ctx, nil, "users")
		s.Error(err, "expected error when transaction is nil")
		s.False(exists, "expected exists to be false when transaction is nil")
	})
	s.Run("when table name is empty, should return error", func() {
		exists, err := builder.HasTable(s.ctx, tx, "")
		s.Error(err, "expected error when table name is empty")
		s.False(exists, "expected exists to be false when table name is empty")
	})
	s.Run("when all parameters are valid", func() {
		err = builder.Create(s.ctx, tx, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name", 255)
			table.String("email", 255).Unique()
			table.String("password", 255).Nullable()
			table.Timestamps()
		})
		s.NoError(err, "expected no error when creating table before checking if it exists")

		exists, err := builder.HasTable(s.ctx, tx, "users")
		s.NoError(err, "expected no error when checking if table exists with valid parameters")
		s.True(exists, "expected exists to be true for existing table")

		exists, err = builder.HasTable(s.ctx, tx, "non_existent_table")
		s.NoError(err, "expected no error when checking non-existent table")
		s.False(exists, "expected exists to be false for non-existent table")
	})
}
