package builders_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/akfaiz/migris/internal/testutil"
	"github.com/akfaiz/migris/schema"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
)

func TestMariadbBuilderSuite(t *testing.T) {
	suite.Run(t, new(mariadbBuilderSuite))
}

type mariadbBuilderSuite struct {
	suite.Suite

	ctx     context.Context
	db      *sql.DB
	builder schema.Builder
	tc      testcontainers.Container
}

func (s *mariadbBuilderSuite) SetupSuite() {
	s.ctx = context.Background()

	container, db, err := testutil.StartMariaDBTestDB(s.ctx)
	s.Require().NoError(err)

	s.tc = container
	s.db = db
	s.builder, err = schema.NewBuilder("mariadb")
	s.Require().NoError(err)
}

func (s *mariadbBuilderSuite) TearDownSuite() {
	_ = s.db.Close()
	if s.tc != nil {
		_ = s.tc.Terminate(s.ctx)
	}
}

func (s *mariadbBuilderSuite) AfterTest(_, _ string) {
	builder := s.builder
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	c := schema.NewContext(s.ctx, tx)
	tables, err := builder.GetTables(c)
	s.Require().NoError(err)
	for _, table := range tables {
		err := builder.DropIfExists(c, table.Name)
		if err != nil {
			s.T().Logf("error dropping table %s: %v", table.Name, err)
		}
	}
	err = tx.Commit()
	s.Require().NoError(err, "expected no error when committing transaction after dropping tables")
}

func (s *mariadbBuilderSuite) TestCreate() {
	builder := s.builder
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback()

	c := schema.NewContext(s.ctx, tx)

	s.Run("when all parameters are valid, should create table successfully", func() {
		err = builder.Create(c, "users", func(table *schema.Blueprint) {
			table.ID()
			table.String("name", 255)
			table.String("email", 255).Unique()
			table.UUID("uuid").Nullable()
			table.Timestamps()
		})
		s.Require().NoError(err)
	})
}

func (s *mariadbBuilderSuite) TestGetColumns() {
	builder := s.builder
	tx, err := s.db.BeginTx(s.ctx, nil)
	s.Require().NoError(err)
	defer tx.Rollback()

	c := schema.NewContext(s.ctx, tx)

	s.Run("when table exists, should return columns successfully", func() {
		tableName := "all_types"
		err = builder.Create(c, tableName, func(table *schema.Blueprint) {
			table.ID()
			table.String("name", 255)
			table.UUID("guid")
			table.Point("location")
		})
		s.Require().NoError(err)

		columns, err := builder.GetColumns(c, tableName)
		s.Require().NoError(err)
		s.Len(columns, 4)

		columnMap := make(map[string]*schema.Column)
		for _, col := range columns {
			columnMap[col.Name] = col
		}

		s.Contains(columnMap, "id")
		s.Contains(columnMap, "name")
		s.Contains(columnMap, "guid")
		s.Contains(columnMap, "location")

		s.Equal("uuid", columnMap["guid"].TypeName)
	})
}
