package schema_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/afkdevs/go-schema"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
)

func TestPostgresBuilderSuite(t *testing.T) {
	suite.Run(t, new(postgresBuilderSuite))
}

type dbConfig struct {
	Database string
	Username string
	Password string
}

func getStringFromEnv(envVar string, defaultValue string) string {
	if value, exists := os.LookupEnv(envVar); exists && value != "" {
		return value
	}
	return defaultValue
}

func parseTestConfig() dbConfig {
	return dbConfig{
		Database: getStringFromEnv("DB_NAME", "db_test"),
		Username: getStringFromEnv("DB_USER", "root"),
		Password: getStringFromEnv("DB_PASSWORD", "password"),
	}
}

type postgresBuilderSuite struct {
	suite.Suite
	ctx     context.Context
	db      *sql.DB
	builder schema.Builder
}

func (s *postgresBuilderSuite) SetupSuite() {
	s.ctx = context.Background()

	config := parseTestConfig()

	dsn := fmt.Sprintf("host=localhost port=5432 user=%s password=%s dbname=%s sslmode=disable", config.Username, config.Password, config.Database)

	db, err := sql.Open("postgres", dsn)
	s.Require().NoError(err)

	err = db.Ping()
	s.Require().NoError(err)

	s.db = db
	s.builder, err = schema.NewBuilder("postgres")
	s.Require().NoError(err)
}

func (s *postgresBuilderSuite) TearDownSuite() {
	_ = s.db.Close()
}

func (s *postgresBuilderSuite) TestCreate() {
	builder := s.builder
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	s.Run("when tx is nil, should return error", func() {
		err := builder.Create(s.ctx, nil, "test_table", func(table *schema.Blueprint) {})
		s.Error(err, "expected error when transaction is nil")
	})
	s.Run("when table name is empty, should return error", func() {
		err := builder.Create(s.ctx, tx, "", func(table *schema.Blueprint) {})
		s.Error(err, "expected error when table name is empty")
	})
	s.Run("when blueprint is nil, should return error", func() {
		err := builder.Create(s.ctx, tx, "test_table", nil)
		s.Error(err, "expected error when blueprint is nil")
	})
	s.Run("when all parameters are valid, should create table successfully", func() {
		err = builder.Create(context.Background(), tx, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name")
			table.String("email").Unique()
			table.String("password").Nullable()
			table.Timestamps()
		})
		s.NoError(err, "expected no error when creating table with valid parameters")
	})
	s.Run("when use custom schema should create it successfully", func() {
		_, err = tx.Exec("CREATE SCHEMA IF NOT EXISTS custom_public")
		s.NoError(err, "expected no error when creating custom schema")
		err = builder.Create(context.Background(), tx, "custom_public.users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name")
			table.String("email").Unique()
			table.String("password").Nullable()
			table.TimestampsTz()
		})
		s.NoError(err, "expected no error when creating table with custom schema")
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
			table.BigInteger("user_id")
			table.String("order_id").Unique()
			table.Decimal("amount", 10, 2)
			table.Timestamp("created_at").UseCurrent()

			table.Foreign("user_id").References("id").On("users").OnDelete("CASCADE").OnUpdate("CASCADE")
		})
		s.NoError(err, "expected no error when creating table with foreign key")
	})
	s.Run("when have custom index should create it successfully", func() {
		err = builder.Create(context.Background(), tx, "orders_2", func(table *schema.Blueprint) {
			table.ID()
			table.String("order_id").Unique("uk_orders_2_order_id")
			table.Decimal("amount", 10, 2)
			table.Timestamp("created_at").UseCurrent()

			table.Index("created_at").Name("idx_orders_created_at").Algorithm("BTREE")
		})
		s.NoError(err, "expected no error when creating table with custom index")
	})
	s.Run("when table already exists, should return error", func() {
		err = builder.Create(context.Background(), tx, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name")
			table.String("email").Unique()
		})
		s.Error(err, "expected error when creating table that already exists")
	})
}

func (s *postgresBuilderSuite) TestDrop() {
	builder := s.builder
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	s.Run("when table name is empty, should return error", func() {
		err := builder.Drop(s.ctx, nil, "")
		s.Error(err, "expected error when table name is empty")
	})
	s.Run("when all parameters are valid, should drop table successfully", func() {
		err = builder.Create(s.ctx, tx, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name")
			table.String("email").Unique()
			table.String("password").Nullable()
			table.Timestamps()
		})
		s.NoError(err, "expected no error when creating table before dropping it")
		err = builder.Drop(s.ctx, tx, "users")
		s.NoError(err, "expected no error when dropping table with valid parameters")
	})
	s.Run("when table does not exist, should return error", func() {
		err = builder.Drop(s.ctx, tx, "non_existent_table")
		s.Error(err, "expected error when dropping table that does not exist")
	})
}

func (s *postgresBuilderSuite) TestDropIfExists() {
	builder := s.builder
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	s.Run("when table name is empty, should return error", func() {
		err := builder.DropIfExists(s.ctx, nil, "")
		s.Error(err, "expected error when table name is empty")
	})
	s.Run("when tx is nil, should return error", func() {
		err = builder.DropIfExists(s.ctx, nil, "test_table")
		s.Error(err, "expected error when transaction is nil")
	})
	s.Run("when all parameters are valid, should drop table successfully", func() {
		err = builder.Create(s.ctx, tx, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name")
			table.String("email").Unique()
			table.String("password").Nullable()
			table.Timestamps()
		})
		s.NoError(err, "expected no error when creating table before dropping it")
		err = builder.DropIfExists(s.ctx, tx, "users")
		s.NoError(err, "expected no error when dropping table with valid parameters")
	})
	s.Run("when table does not exist, should not return error", func() {
		err = builder.DropIfExists(s.ctx, tx, "non_existent_table")
		s.NoError(err, "expected no error when dropping non-existent table with IF EXISTS clause")
	})
}

func (s *postgresBuilderSuite) TestRename() {
	builder := s.builder
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	s.Run("when tx is nil, should return error", func() {
		err := builder.Rename(s.ctx, nil, "old_table", "new_table")
		s.Error(err, "expected error when transaction is nil")
	})
	s.Run("when old table name is empty, should return error", func() {
		err := builder.Rename(s.ctx, tx, "", "new_table")
		s.Error(err, "expected error when old table name is empty")
	})
	s.Run("when new table name is empty, should return error", func() {
		err := builder.Rename(s.ctx, tx, "old_table", "")
		s.Error(err, "expected error when new table name is empty")
	})
	s.Run("when all parameters are valid, should rename table successfully", func() {
		err = builder.Create(s.ctx, tx, "old_table", func(table *schema.Blueprint) {
			table.ID()
			table.String("name", 255)
		})
		s.NoError(err, "expected no error when creating table before renaming it")
		err = builder.Rename(s.ctx, tx, "old_table", "new_table")
		s.NoError(err, "expected no error when renaming table with valid parameters")
	})
	s.Run("when renaming non-existent table, should return error", func() {
		err = builder.Rename(s.ctx, tx, "non_existent_table", "new_table")
		s.Error(err, "expected error when renaming non-existent table")
		s.ErrorContains(err, "does not exist", "expected error message to contain 'does not exist'")
	})
}

func (s *postgresBuilderSuite) TestTable() {
	builder := s.builder
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	s.Run("when tx is nil, should return error", func() {
		err := builder.Table(s.ctx, nil, "test_table", func(table *schema.Blueprint) {})
		s.Error(err, "expected error when transaction is nil")
	})
	s.Run("when table name is empty, should return error", func() {
		err := builder.Table(s.ctx, tx, "", func(table *schema.Blueprint) {})
		s.Error(err, "expected error when table name is empty")
	})
	s.Run("when blueprint is nil, should return error", func() {
		err := builder.Table(s.ctx, tx, "test_table", nil)
		s.Error(err, "expected error when blueprint is nil")
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
				table.DropUnique([]string{"email"})
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
				table.ID()
				table.String("role_name", 255).Unique("uk_roles_role_name")
			})
			s.NoError(err, "expected no error when creating roles table before adding foreign key")
			err = builder.Table(s.ctx, tx, "users", func(table *schema.Blueprint) {
				table.Integer("role_id").Nullable()
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
				table.DropPrimary("pk_users")
			})
			s.NoError(err, "expected no error when dropping primary key from users table")
		})
	})
}

func (s *postgresBuilderSuite) TestGetColumns() {
	builder := s.builder
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	s.Run("when tx is nil, should return error", func() {
		_, err := builder.GetColumns(s.ctx, nil, "users")
		s.Error(err, "expected error when transaction is nil")
	})
	s.Run("when table name is empty, should return error", func() {
		_, err := builder.GetColumns(s.ctx, tx, "")
		s.Error(err, "expected error when table name is empty")
	})
	s.Run("when all parameters are valid", func() {
		err = builder.Create(s.ctx, tx, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name", 255)
			table.String("email", 255).Unique()
			table.String("password", 255).Nullable()
			table.Timestamps()
		})
		s.NoError(err, "expected no error when creating table before getting columns")

		columns, err := builder.GetColumns(s.ctx, tx, "users")
		s.NoError(err, "expected no error when getting columns with valid parameters")
		s.Len(columns, 6, "expected 6 columns to be returned")
	})
	s.Run("when table does not exist, should return empty columns", func() {
		columns, err := builder.GetColumns(s.ctx, tx, "non_existent_table")
		s.NoError(err, "expected no error when getting columns of non-existent table")
		s.Empty(columns, "expected empty columns for non-existent table")
	})
}

func (s *postgresBuilderSuite) TestGetIndexes() {
	builder := s.builder
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	s.Run("when tx is nil, should return error", func() {
		_, err := builder.GetIndexes(s.ctx, nil, "users")
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

func (s *postgresBuilderSuite) TestGetTables() {
	builder := s.builder
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	s.Run("when tx is nil, should return error", func() {
		_, err := builder.GetTables(s.ctx, nil)
		s.Error(err, "expected error when transaction is nil")
	})
	s.Run("when all parameters are valid", func() {
		err = builder.Create(s.ctx, tx, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name", 255)
			table.String("email", 255).Unique()
			table.String("password", 255).Nullable()
			table.Timestamps()
		})
		s.NoError(err, "expected no error when creating table before getting tables")

		tables, err := builder.GetTables(s.ctx, tx)
		s.NoError(err, "expected no error when getting tables with valid parameters")
		s.Len(tables, 1, "expected 1 table to be returned")
		userTable := tables[0]
		s.Equal("users", userTable.Name, "expected table name to be 'users'")
		s.Equal("public", userTable.Schema, "expected table schema to be 'public'")
		s.False(userTable.Comment.Valid, "expected table comment to be invalid (nil)")
	})
}

func (s *postgresBuilderSuite) TestHasColumn() {
	builder := s.builder
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	s.Run("when tx is nil, should return error", func() {
		exists, err := builder.HasColumn(s.ctx, nil, "users", "name")
		s.Error(err, "expected error when transaction is nil")
		s.False(exists, "expected exists to be false when transaction is nil")
	})
	s.Run("when table name is empty, should return error", func() {
		exists, err := builder.HasColumn(s.ctx, tx, "", "name")
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
		s.NoError(err, "expected no error when creating table before checking column existence")

		exists, err := builder.HasColumn(s.ctx, tx, "users", "name")
		s.NoError(err, "expected no error when checking if column exists with valid parameters")
		s.True(exists, "expected exists to be true for existing column")

		exists, err = builder.HasColumn(s.ctx, tx, "users", "non_existent_column")
		s.NoError(err, "expected no error when checking non-existent column")
		s.False(exists, "expected exists to be false for non-existent column")
	})
}

func (s *postgresBuilderSuite) TestHasColumns() {
	builder := s.builder
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	s.Run("when tx is nil, should return error", func() {
		exists, err := builder.HasColumns(s.ctx, nil, "users", []string{"name"})
		s.Error(err, "expected error when transaction is nil")
		s.False(exists, "expected exists to be false when transaction is nil")
	})
	s.Run("when table name is empty, should return error", func() {
		exists, err := builder.HasColumns(s.ctx, tx, "", []string{"name"})
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
		s.NoError(err, "expected no error when creating table before checking column existence")

		exists, err := builder.HasColumns(s.ctx, tx, "users", []string{"name", "email"})
		s.NoError(err, "expected no error when checking if columns exist with valid parameters")
		s.True(exists, "expected exists to be true for existing columns")

		exists, err = builder.HasColumns(s.ctx, tx, "users", []string{"name", "non_existent_column"})
		s.NoError(err, "expected no error when checking mixed existing and non-existent columns")
		s.False(exists, "expected exists to be false for mixed existing and non-existent columns")
	})
}

func (s *postgresBuilderSuite) TestHasIndex() {
	builder := s.builder
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback() //nolint:errcheck

	s.Run("when tx is nil, should return error", func() {
		exists, err := builder.HasIndex(s.ctx, nil, "users", []string{"idx_users_name"})
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
			table.Timestamp("created_at").UseCurrent()

			table.Index("company_id", "user_id")
			table.Unique("order_id").Name("uk_orders_order_id")
		})
		s.Require().NoError(err, "expected no error when creating table with index")

		exists, err := builder.HasIndex(s.ctx, tx, "orders", []string{"uk_orders_order_id"})
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

func (s *postgresBuilderSuite) TestHasTable() {
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
		s.NoError(err, "expected no error when creating table before checking existence")

		exists, err := builder.HasTable(s.ctx, tx, "users")
		s.NoError(err, "expected no error when checking if table exists with valid parameters")
		s.True(exists, "expected exists to be true for existing table")

		exists, err = builder.HasTable(s.ctx, tx, "non_existent_table")
		s.NoError(err, "expected no error when checking non-existent table")
		s.False(exists, "expected exists to be false for non-existent table")
	})
	s.Run("when checking table with custom schema, should return true if exists", func() {
		_, err = tx.Exec("CREATE SCHEMA IF NOT EXISTS custom_publics")
		s.NoError(err, "expected no error when creating custom schema")

		err = builder.Create(s.ctx, tx, "custom_publics.users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name", 255)
			table.String("email", 255).Unique()
			table.String("password", 255).Nullable()
			table.Timestamps()
		})
		s.NoError(err, "expected no error when creating table with custom schema")

		exists, err := builder.HasTable(s.ctx, tx, "custom_publics.users")
		s.NoError(err, "expected no error when checking if table with custom schema exists")
		s.True(exists, "expected exists to be true for existing table with custom schema")
		exists, err = builder.HasTable(s.ctx, tx, "custom_publics.non_existent_table")
		s.NoError(err, "expected no error when checking non-existent table with custom schema")
		s.False(exists, "expected exists to be false for non-existent table with custom schema")
	})
}
