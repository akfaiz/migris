package blueprint_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/akfaiz/migris/internal/dialect"
	"github.com/akfaiz/migris/schema/blueprint"
	"github.com/akfaiz/migris/schema/core"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockGrammar struct {
	blueprint.BaseGrammar
}

func (m *mockGrammar) CompileTableExists(_ string, _ string) (string, error) { return "", nil }
func (m *mockGrammar) CompileTables(_ string) (string, error)                { return "", nil }
func (m *mockGrammar) CompileColumns(_, _ string) (string, error)            { return "", nil }
func (m *mockGrammar) CompileIndexes(_, _ string) (string, error)            { return "", nil }
func (m *mockGrammar) CompileCreate(_ *blueprint.Blueprint) (string, error)  { return "SELECT 1", nil }
func (m *mockGrammar) CompileAdd(_ *blueprint.Blueprint) (string, error)     { return "SELECT 1", nil }
func (m *mockGrammar) CompileDrop(_ *blueprint.Blueprint) (string, error)    { return "SELECT 1", nil }
func (m *mockGrammar) CompileDropIfExists(_ *blueprint.Blueprint) (string, error) {
	return "SELECT 1", nil
}
func (m *mockGrammar) CompileRename(_ *blueprint.Blueprint, _ *blueprint.Command) (string, error) {
	return "SELECT 1", nil
}
func (m *mockGrammar) CompileChange(_ *blueprint.Blueprint, _ *blueprint.Command) (string, error) {
	return "SELECT 1", nil
}
func (m *mockGrammar) GetType(_ *blueprint.Column) string { return "TEXT" }

func TestBlueprint_TableSettings(t *testing.T) {
	bp := blueprint.NewBlueprintForTesting("users", &mockGrammar{})

	bp.Charset("utf8mb4")
	bp.Collation("utf8mb4_unicode_ci")
	bp.Engine("InnoDB")
	bp.Comment("User table")
	bp.AutoIncrementStartingValues(100)

	assert.Equal(t, "utf8mb4", bp.CharsetVal)
	assert.Equal(t, "utf8mb4_unicode_ci", bp.CollationVal)
	assert.Equal(t, "InnoDB", bp.EngineVal)
	assert.Equal(t, "User table", bp.CommentVal)
	assert.Equal(t, 100, *bp.AutoIncrementStartingValuesVal)
}

func TestBlueprint_ColumnAddition(t *testing.T) {
	bp := blueprint.NewBlueprintForTesting("users", &mockGrammar{})

	bp.ID()
	bp.String("email", 100).Unique().Nullable()
	bp.Integer("age").Unsigned().Default(18)
	bp.Boolean("active").Default(true)
	bp.Timestamps()

	columns := bp.Columns
	require.Len(t, columns, 6)

	assert.Equal(t, "id", columns[0].Name)
	assert.Equal(t, blueprint.ColumnTypeBigInteger, columns[0].ColumnType)
	assert.True(t, *columns[0].AutoIncrementVal)
	assert.True(t, *columns[0].PrimaryVal)

	assert.Equal(t, "email", columns[1].Name)
	assert.Equal(t, blueprint.ColumnTypeString, columns[1].ColumnType)
	assert.Equal(t, 100, *columns[1].Length)
	assert.True(t, *columns[1].UniqueVal)
	assert.True(t, *columns[1].NullableVal)

	assert.Equal(t, "age", columns[2].Name)
	assert.Equal(t, blueprint.ColumnTypeInteger, columns[2].ColumnType)
	assert.True(t, *columns[2].UnsignedVal)
	assert.Equal(t, 18, columns[2].DefaultValue)

	assert.Equal(t, "active", columns[3].Name)
	assert.Equal(t, blueprint.ColumnTypeBoolean, columns[3].ColumnType)
	assert.Equal(t, true, columns[3].DefaultValue)

	assert.Equal(t, "created_at", columns[4].Name)
	assert.Equal(t, blueprint.ColumnTypeTimestamp, columns[4].ColumnType)
	assert.True(t, columns[4].UseCurrentVal)

	assert.Equal(t, "updated_at", columns[5].Name)
	assert.Equal(t, blueprint.ColumnTypeTimestamp, columns[5].ColumnType)
	assert.True(t, columns[5].UseCurrentVal)
	assert.True(t, columns[5].UseCurrentOnUpdateVal)
}

func TestBlueprint_Indexes(t *testing.T) {
	bp := blueprint.NewBlueprintForTesting("users", &mockGrammar{})

	bp.Index("email", "username").Name("idx_email_username").Algorithm("btree")
	bp.Unique("phone").Name("uk_phone")
	bp.Primary("id")
	bp.FullText("bio")

	commands := bp.Commands
	require.Len(t, commands, 4)

	assert.Equal(t, blueprint.CommandIndex, commands[0].Name)
	assert.Equal(t, []string{"email", "username"}, commands[0].Columns)
	assert.Equal(t, "idx_email_username", commands[0].Index)
	assert.Equal(t, "btree", commands[0].Algorithm)

	assert.Equal(t, blueprint.CommandUnique, commands[1].Name)
	assert.Equal(t, []string{"phone"}, commands[1].Columns)
	assert.Equal(t, "uk_phone", commands[1].Index)

	assert.Equal(t, blueprint.CommandPrimary, commands[2].Name)
	assert.Equal(t, []string{"id"}, commands[2].Columns)

	assert.Equal(t, blueprint.CommandFullText, commands[3].Name)
	assert.Equal(t, []string{"bio"}, commands[3].Columns)
}

func TestBlueprint_ForeignKeys(t *testing.T) {
	bp := blueprint.NewBlueprintForTesting("posts", &mockGrammar{})

	bp.Foreign("user_id").References("id").On("users").OnDelete("CASCADE").OnUpdate("RESTRICT")

	commands := bp.Commands
	require.Len(t, commands, 1)

	assert.Equal(t, blueprint.CommandForeign, commands[0].Name)
	assert.Equal(t, []string{"user_id"}, commands[0].Columns)
	assert.Equal(t, []string{"id"}, commands[0].References)
	assert.Equal(t, "users", commands[0].On)
	assert.Equal(t, "CASCADE", commands[0].OnDelete)
	assert.Equal(t, "RESTRICT", commands[0].OnUpdate)
}

func TestBlueprint_DropOperations(t *testing.T) {
	bp := blueprint.NewBlueprintForTesting("users", &mockGrammar{})

	bp.DropColumn("age", "active")
	bp.DropIndex("idx_email")
	bp.DropUnique("uk_phone")
	bp.DropPrimary([]string{"id"})
	bp.DropForeign("fk_user_id")

	commands := bp.Commands
	require.Len(t, commands, 5)

	assert.Equal(t, blueprint.CommandDropColumn, commands[0].Name)
	assert.Equal(t, []string{"age", "active"}, commands[0].Columns)

	assert.Equal(t, blueprint.CommandDropIndex, commands[1].Name)
	assert.Equal(t, "idx_email", commands[1].Index)

	assert.Equal(t, blueprint.CommandDropUnique, commands[2].Name)
	assert.Equal(t, "uk_phone", commands[2].Index)

	assert.Equal(t, blueprint.CommandDropPrimary, commands[3].Name)
	assert.Equal(t, "users_id_primary", commands[3].Index)

	assert.Equal(t, blueprint.CommandDropForeign, commands[4].Name)
	assert.Equal(t, "fk_user_id", commands[4].Index)
}

func TestBlueprint_ImpliedCommands(t *testing.T) {
	t.Run("creates add command when not creating", func(t *testing.T) {
		bp := blueprint.NewBlueprintForTesting("users", &mockGrammar{})
		bp.String("new_col")
		bp.AddImpliedCommands()

		assert.Equal(t, blueprint.CommandAdd, bp.Commands[0].Name)
	})

	t.Run("does not create add command when creating", func(t *testing.T) {
		bp := blueprint.NewBlueprintForTesting("users", &mockGrammar{})
		bp.Create()
		bp.String("new_col")
		bp.AddImpliedCommands()

		for _, cmd := range bp.Commands {
			assert.NotEqual(t, blueprint.CommandAdd, cmd.Name)
		}
	})

	t.Run("handles fluent indexes for MySQL", func(t *testing.T) {
		bp := blueprint.NewBlueprintForTesting("users", &mockGrammar{})
		bp.SetDialect(dialect.MySQL)
		bp.Integer("id").Primary()
		bp.AddImpliedCommands()

		// MySQL primary key should be handled in-line, not as a command
		for _, cmd := range bp.Commands {
			assert.NotEqual(t, blueprint.CommandPrimary, cmd.Name)
		}
	})

	t.Run("handles fluent indexes for non-MySQL", func(t *testing.T) {
		bp := blueprint.NewBlueprintForTesting("users", &mockGrammar{})
		bp.SetDialect(dialect.Postgres)
		bp.String("email").Unique()
		bp.AddImpliedCommands()

		found := false
		for _, cmd := range bp.Commands {
			if cmd.Name == blueprint.CommandUnique {
				found = true
				assert.Equal(t, []string{"email"}, cmd.Columns)
			}
		}
		assert.True(t, found)
	})
}

func TestBlueprint_AllColumnTypes(t *testing.T) {
	bp := blueprint.NewBlueprintForTesting("all_types", &mockGrammar{})

	bp.Char("char_col", 10)
	bp.LongText("long_text")
	bp.Text("text_col")
	bp.MediumText("medium_text")
	bp.TinyText("tiny_text")
	bp.Decimal("decimal_col", 10, 3)
	bp.Double("double_col")
	bp.Float("float_col", 20)
	bp.Increments("inc")
	bp.MediumIncrements("med_inc")
	bp.SmallIncrements("small_inc")
	bp.TinyIncrements("tiny_inc")
	bp.Integer("int_col")
	bp.MediumInteger("med_int")
	bp.SmallInteger("small_int")
	bp.TinyInteger("tiny_int")
	bp.UnsignedInteger("u_int")
	bp.UnsignedMediumInteger("u_med_int")
	bp.UnsignedSmallInteger("u_small_int")
	bp.UnsignedTinyInteger("u_tiny_int")
	bp.DateTime("dt", 6)
	bp.DateTimeTz("dt_tz", 6)
	bp.Date("d")
	bp.Time("t", 6)
	bp.TimeTz("t_tz", 6)
	bp.TimestampTz("ts_tz", 6)
	bp.Year("y")
	bp.Binary("bin", 100)
	bp.JSON("j")
	bp.JSONB("jb")
	bp.UUID("u")
	bp.ULID("ul")
	bp.Geography("geog", "POINT", 4326)
	bp.Geometry("geom", "POLYGON", 4326)
	bp.Point("pt", 4326)
	bp.Enum("e", []string{"a", "b"})
	bp.Set("s", []string{"x", "y"})
	bp.IPAddress("ip")
	bp.MacAddress("mac")
	bp.Vector("v", 128)
	bp.TSVector("tsv")
	bp.Cidr("cid")
	bp.Inet("ine")
	bp.MacAddr("ma")
	bp.MacAddr8("ma8")
	bp.Column("custom", "CUSTOM_TYPE")

	assert.Len(t, bp.Columns, 46)
}

func TestBlueprint_Altering(t *testing.T) {
	bp := blueprint.NewBlueprintForTesting("users", &mockGrammar{})
	bp.Drop()
	bp.DropIfExists()
	bp.Rename("new_users")
	bp.RenameColumn("old", "new")
	bp.DropTimestamps()
	bp.DropTimestampsTz()
	bp.RenameIndex("old_idx", "new_idx")
	bp.DropFulltext("idx")

	assert.Len(t, bp.Commands, 8)
}

type mockBuilder struct{}

func (m *mockBuilder) GetColumns(_ core.Context, _ string) ([]*core.Column, error) {
	return []*core.Column{
		{
			Name:       "bio",
			Nullable:   false,
			DefaultVal: sql.NullString{String: "Hello", Valid: true},
			Comment:    sql.NullString{String: "World", Valid: true},
		},
	}, nil
}

func TestBlueprint_BuildAndHydrate(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()
	ctx := context.Background()
	tx, _ := db.BeginTx(ctx, nil)
	defer tx.Rollback()
	c := core.NewContext(ctx, tx, core.WithDialect("postgres"))

	bp := blueprint.NewBlueprintForTesting("users", &mockGrammar{})
	bp.SetBuilder(&mockBuilder{})
	bp.String("bio").Change()

	err := bp.Build(c)
	require.NoError(t, err)

	col := bp.Columns[0]
	assert.False(t, *col.NullableVal)
	assert.Equal(t, "Hello", string(col.DefaultValue.(blueprint.Expression)))
	assert.Equal(t, "World", *col.CommentVal)
}

func TestBlueprint_DropIndex_Panic(t *testing.T) {
	bp := blueprint.NewBlueprintForTesting("users", &mockGrammar{})
	assert.Panics(t, func() {
		bp.DropIndex(123)
	})
}

func (m *mockGrammar) GetFluentCommands() []func(blueprint *blueprint.Blueprint, command *blueprint.Command) string {
	return []func(blueprint *blueprint.Blueprint, command *blueprint.Command) string{
		func(_ *blueprint.Blueprint, cmd *blueprint.Command) string {
			if cmd.Column != nil && cmd.Column.Name == "fluent" {
				return "SELECT 1"
			}
			return ""
		},
	}
}

func (m *mockGrammar) GetTableFluentCommands() []func(blueprint *blueprint.Blueprint) string {
	return []func(blueprint *blueprint.Blueprint) string{
		func(_ *blueprint.Blueprint) string {
			return "SELECT 1"
		},
	}
}

func TestBlueprint_FluentStatements(t *testing.T) {
	bp := blueprint.NewBlueprintForTesting("users", &mockGrammar{})
	bp.String("fluent")
	bp.String("normal")

	statements := bp.GetFluentStatements()
	assert.Contains(t, statements, "SELECT 1")
}

func TestBlueprint_ToSQL_Errors(t *testing.T) {
	bp := blueprint.NewBlueprintForTesting("users", &mockGrammar{})
	bp.Commands = append(bp.Commands, &blueprint.Command{Name: "unknown"})
	_, err := bp.ToSQL()
	assert.Error(t, err)
}
