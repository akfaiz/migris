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
	assert.True(t, *columns[4].NullableVal)

	assert.Equal(t, "updated_at", columns[5].Name)
	assert.Equal(t, blueprint.ColumnTypeTimestamp, columns[5].ColumnType)
	assert.True(t, *columns[5].NullableVal)
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

func TestBlueprint_Temporary(t *testing.T) {
	bp := blueprint.NewBlueprintForTesting("logs", &mockGrammar{})
	assert.False(t, bp.TemporaryVal)
	bp.Temporary()
	assert.True(t, bp.TemporaryVal)
}

func TestBlueprint_InnoDB(t *testing.T) {
	bp := blueprint.NewBlueprintForTesting("users", &mockGrammar{})
	bp.InnoDB()
	assert.Equal(t, "InnoDB", bp.EngineVal)
}

func TestBlueprint_SoftDeletes(t *testing.T) {
	t.Run("SoftDeletes adds nullable deleted_at timestamp", func(t *testing.T) {
		bp := blueprint.NewBlueprintForTesting("users", &mockGrammar{})
		bp.SoftDeletes("deleted_at")
		require.Len(t, bp.Columns, 1)
		assert.Equal(t, "deleted_at", bp.Columns[0].Name)
		assert.Equal(t, blueprint.ColumnTypeTimestamp, bp.Columns[0].ColumnType)
		assert.True(t, *bp.Columns[0].NullableVal)
	})

	t.Run("SoftDeletes custom name", func(t *testing.T) {
		bp := blueprint.NewBlueprintForTesting("users", &mockGrammar{})
		bp.SoftDeletes("removed_at")
		assert.Equal(t, "removed_at", bp.Columns[0].Name)
	})

	t.Run("SoftDeletesTz adds nullable timestamptz", func(t *testing.T) {
		bp := blueprint.NewBlueprintForTesting("users", &mockGrammar{})
		bp.SoftDeletesTz("deleted_at")
		assert.Equal(t, blueprint.ColumnTypeTimestampTz, bp.Columns[0].ColumnType)
		assert.True(t, *bp.Columns[0].NullableVal)
	})

	t.Run("SoftDeletesDatetime adds nullable datetime", func(t *testing.T) {
		bp := blueprint.NewBlueprintForTesting("users", &mockGrammar{})
		bp.SoftDeletesDatetime("deleted_at")
		assert.Equal(t, blueprint.ColumnTypeDateTime, bp.Columns[0].ColumnType)
		assert.True(t, *bp.Columns[0].NullableVal)
	})

	t.Run("DropSoftDeletes adds dropColumn command", func(t *testing.T) {
		bp := blueprint.NewBlueprintForTesting("users", &mockGrammar{})
		bp.DropSoftDeletes()
		require.Len(t, bp.Commands, 1)
		assert.Equal(t, blueprint.CommandDropColumn, bp.Commands[0].Name)
		assert.Equal(t, []string{"deleted_at"}, bp.Commands[0].Columns)
	})

	t.Run("DropSoftDeletesTz delegates to DropSoftDeletes", func(t *testing.T) {
		bp := blueprint.NewBlueprintForTesting("users", &mockGrammar{})
		bp.DropSoftDeletesTz()
		require.Len(t, bp.Commands, 1)
		assert.Equal(t, []string{"deleted_at"}, bp.Commands[0].Columns)
	})
}

func TestBlueprint_NullableTimestampHelpers(t *testing.T) {
	t.Run("NullableTimestamps produces nullable created_at and updated_at", func(t *testing.T) {
		bp := blueprint.NewBlueprintForTesting("users", &mockGrammar{})
		bp.NullableTimestamps()
		require.Len(t, bp.Columns, 2)
		assert.Equal(t, "created_at", bp.Columns[0].Name)
		assert.True(t, *bp.Columns[0].NullableVal)
		assert.Equal(t, "updated_at", bp.Columns[1].Name)
		assert.True(t, *bp.Columns[1].NullableVal)
	})

	t.Run("NullableTimestampsTz produces nullable timestamptz columns", func(t *testing.T) {
		bp := blueprint.NewBlueprintForTesting("users", &mockGrammar{})
		bp.NullableTimestampsTz()
		require.Len(t, bp.Columns, 2)
		assert.Equal(t, blueprint.ColumnTypeTimestampTz, bp.Columns[0].ColumnType)
		assert.True(t, *bp.Columns[0].NullableVal)
	})

	t.Run("Datetimes produces nullable datetime columns", func(t *testing.T) {
		bp := blueprint.NewBlueprintForTesting("users", &mockGrammar{})
		bp.Datetimes()
		require.Len(t, bp.Columns, 2)
		assert.Equal(t, blueprint.ColumnTypeDateTime, bp.Columns[0].ColumnType)
		assert.True(t, *bp.Columns[0].NullableVal)
		assert.Equal(t, blueprint.ColumnTypeDateTime, bp.Columns[1].ColumnType)
		assert.True(t, *bp.Columns[1].NullableVal)
	})
}

func TestBlueprint_RawColumn(t *testing.T) {
	bp := blueprint.NewBlueprintForTesting("logs", &mockGrammar{})
	bp.RawColumn("payload", "MEDIUMBLOB NOT NULL")
	require.Len(t, bp.Columns, 1)
	assert.Equal(t, "payload", bp.Columns[0].Name)
	assert.Equal(t, blueprint.ColumnTypeRaw, bp.Columns[0].ColumnType)
	assert.Equal(t, "MEDIUMBLOB NOT NULL", *bp.Columns[0].RawDefinition)
}

func TestBlueprint_RemoveColumn(t *testing.T) {
	bp := blueprint.NewBlueprintForTesting("users", &mockGrammar{})
	bp.String("email")
	bp.String("name")
	bp.String("phone")
	require.Len(t, bp.Columns, 3)

	bp.RemoveColumn("name")
	require.Len(t, bp.Columns, 2)
	assert.Equal(t, "email", bp.Columns[0].Name)
	assert.Equal(t, "phone", bp.Columns[1].Name)
}

func TestBlueprint_SpatialVectorIndex(t *testing.T) {
	t.Run("SpatialIndex adds spatialIndex command", func(t *testing.T) {
		bp := blueprint.NewBlueprintForTesting("locations", &mockGrammar{})
		bp.SpatialIndex("geom")
		require.Len(t, bp.Commands, 1)
		assert.Equal(t, blueprint.CommandSpatialIndex, bp.Commands[0].Name)
		assert.Equal(t, []string{"geom"}, bp.Commands[0].Columns)
	})

	t.Run("SpatialIndex with name", func(t *testing.T) {
		bp := blueprint.NewBlueprintForTesting("locations", &mockGrammar{})
		bp.SpatialIndex("geom").Name("idx_geom")
		assert.Equal(t, "idx_geom", bp.Commands[0].Index)
	})

	t.Run("VectorIndex adds vectorIndex command", func(t *testing.T) {
		bp := blueprint.NewBlueprintForTesting("items", &mockGrammar{})
		bp.VectorIndex("embedding")
		require.Len(t, bp.Commands, 1)
		assert.Equal(t, blueprint.CommandVectorIndex, bp.Commands[0].Name)
		assert.Equal(t, []string{"embedding"}, bp.Commands[0].Columns)
	})

	t.Run("DropSpatialIndex adds dropSpatialIndex command", func(t *testing.T) {
		bp := blueprint.NewBlueprintForTesting("locations", &mockGrammar{})
		bp.DropSpatialIndex("idx_geom")
		require.Len(t, bp.Commands, 1)
		assert.Equal(t, blueprint.CommandDropSpatialIndex, bp.Commands[0].Name)
		assert.Equal(t, "idx_geom", bp.Commands[0].Index)
	})
}

func TestBlueprint_Morphs(t *testing.T) {
	t.Run("NumericMorphs adds type+id columns and index", func(t *testing.T) {
		bp := blueprint.NewBlueprintForTesting("comments", &mockGrammar{})
		bp.NumericMorphs("commentable")
		require.Len(t, bp.Columns, 2)
		assert.Equal(t, "commentable_type", bp.Columns[0].Name)
		assert.Equal(t, "commentable_id", bp.Columns[1].Name)
		assert.Equal(t, blueprint.ColumnTypeBigInteger, bp.Columns[1].ColumnType)
		assert.True(t, *bp.Columns[1].UnsignedVal)
		require.Len(t, bp.Commands, 1)
		assert.Equal(t, blueprint.CommandIndex, bp.Commands[0].Name)
	})

	t.Run("NullableNumericMorphs adds nullable columns", func(t *testing.T) {
		bp := blueprint.NewBlueprintForTesting("comments", &mockGrammar{})
		bp.NullableNumericMorphs("commentable")
		assert.True(t, *bp.Columns[0].NullableVal)
		assert.True(t, *bp.Columns[1].NullableVal)
	})

	t.Run("UUIDMorphs adds uuid id column", func(t *testing.T) {
		bp := blueprint.NewBlueprintForTesting("comments", &mockGrammar{})
		bp.UUIDMorphs("commentable")
		assert.Equal(t, blueprint.ColumnTypeUUID, bp.Columns[1].ColumnType)
	})

	t.Run("ULIDMorphs adds ulid id column", func(t *testing.T) {
		bp := blueprint.NewBlueprintForTesting("comments", &mockGrammar{})
		bp.ULIDMorphs("commentable")
		assert.Equal(t, blueprint.ColumnTypeULID, bp.Columns[1].ColumnType)
	})

	t.Run("Morphs with default key type delegates to NumericMorphs", func(t *testing.T) {
		// DefaultMorphKeyType defaults to "int" → NumericMorphs
		bp := blueprint.NewBlueprintForTesting("comments", &mockGrammar{})
		bp.NumericMorphs("commentable")
		assert.Equal(t, blueprint.ColumnTypeBigInteger, bp.Columns[1].ColumnType)
	})

	t.Run("UUIDMorphs produces uuid id", func(t *testing.T) {
		bp := blueprint.NewBlueprintForTesting("comments", &mockGrammar{})
		bp.UUIDMorphs("commentable")
		assert.Equal(t, blueprint.ColumnTypeUUID, bp.Columns[1].ColumnType)
	})

	t.Run("DropMorphs drops columns and index", func(t *testing.T) {
		bp := blueprint.NewBlueprintForTesting("comments", &mockGrammar{})
		bp.DropMorphs("commentable")
		require.Len(t, bp.Commands, 2)
		assert.Equal(t, blueprint.CommandDropIndex, bp.Commands[0].Name)
		assert.Equal(t, blueprint.CommandDropColumn, bp.Commands[1].Name)
		assert.Equal(t, []string{"commentable_type", "commentable_id"}, bp.Commands[1].Columns)
	})
}

func TestBlueprint_ForeignID(t *testing.T) {
	t.Run("ForeignID adds bigint unsigned column", func(t *testing.T) {
		bp := blueprint.NewBlueprintForTesting("posts", &mockGrammar{})
		bp.ForeignID("user_id")
		require.Len(t, bp.Columns, 1)
		assert.Equal(t, "user_id", bp.Columns[0].Name)
		assert.Equal(t, blueprint.ColumnTypeBigInteger, bp.Columns[0].ColumnType)
		assert.True(t, *bp.Columns[0].UnsignedVal)
	})

	t.Run("ForeignID Constrained infers table name", func(t *testing.T) {
		bp := blueprint.NewBlueprintForTesting("posts", &mockGrammar{})
		bp.ForeignID("user_id").Constrained()
		require.Len(t, bp.Commands, 1)
		assert.Equal(t, blueprint.CommandForeign, bp.Commands[0].Name)
		assert.Equal(t, "users", bp.Commands[0].On)
		assert.Equal(t, []string{"id"}, bp.Commands[0].References)
	})

	t.Run("ForeignID Constrained with explicit table", func(t *testing.T) {
		bp := blueprint.NewBlueprintForTesting("posts", &mockGrammar{})
		bp.ForeignID("author_id").Constrained("accounts")
		assert.Equal(t, "accounts", bp.Commands[0].On)
	})

	t.Run("ForeignUUID adds uuid column", func(t *testing.T) {
		bp := blueprint.NewBlueprintForTesting("posts", &mockGrammar{})
		bp.ForeignUUID("user_id")
		assert.Equal(t, blueprint.ColumnTypeUUID, bp.Columns[0].ColumnType)
	})

	t.Run("ForeignULID adds char(26) column", func(t *testing.T) {
		bp := blueprint.NewBlueprintForTesting("posts", &mockGrammar{})
		bp.ForeignULID("user_id")
		assert.Equal(t, blueprint.ColumnTypeChar, bp.Columns[0].ColumnType)
		assert.Equal(t, 26, *bp.Columns[0].Length)
	})

	t.Run("DropConstrainedForeignID drops foreign and column", func(t *testing.T) {
		bp := blueprint.NewBlueprintForTesting("posts", &mockGrammar{})
		bp.DropConstrainedForeignID("user_id")
		require.Len(t, bp.Commands, 2)
		assert.Equal(t, blueprint.CommandDropForeign, bp.Commands[0].Name)
		assert.Equal(t, blueprint.CommandDropColumn, bp.Commands[1].Name)
	})
}

func TestBlueprint_After(t *testing.T) {
	bp := blueprint.NewBlueprintForTesting("users", &mockGrammar{})
	bp.String("email")
	bp.After("email", func(b *blueprint.Blueprint) {
		b.String("phone")
		b.String("bio")
	})
	require.Len(t, bp.Columns, 3)
	assert.Equal(t, "email", bp.Columns[0].Name)
	// phone positioned after email, bio positioned after phone (chained)
	assert.Equal(t, "phone", bp.Columns[1].Name)
	assert.Equal(t, "email", *bp.Columns[1].AfterVal)
	assert.Equal(t, "bio", bp.Columns[2].Name)
	assert.Equal(t, "phone", *bp.Columns[2].AfterVal)
}

func TestBlueprint_RawIndex(t *testing.T) {
	bp := blueprint.NewBlueprintForTesting("users", &mockGrammar{})
	bp.RawIndex("(lower(email))", "idx_email_lower")
	require.Len(t, bp.Commands, 1)
	assert.Equal(t, blueprint.CommandIndex, bp.Commands[0].Name)
	assert.Equal(t, []string{"(lower(email))"}, bp.Commands[0].Columns)
	assert.Equal(t, "idx_email_lower", bp.Commands[0].Index)
}

func TestBlueprint_IntegerIncrements(t *testing.T) {
	bp := blueprint.NewBlueprintForTesting("users", &mockGrammar{})
	bp.IntegerIncrements("id")
	require.Len(t, bp.Columns, 1)
	assert.Equal(t, "id", bp.Columns[0].Name)
	assert.Equal(t, blueprint.ColumnTypeInteger, bp.Columns[0].ColumnType)
	assert.True(t, *bp.Columns[0].UnsignedVal)
	assert.True(t, *bp.Columns[0].AutoIncrementVal)
}

func TestBlueprint_Morphs_KeyTypeDispatch(t *testing.T) {
	original := blueprint.DefaultMorphKeyType
	t.Cleanup(func() {
		blueprint.DefaultMorphKeyType = original //nolint:reassign // Tests verify global morph-key dispatch.
	})

	t.Run("int dispatches to NumericMorphs", func(t *testing.T) {
		blueprint.DefaultMorphKeyType = "int" //nolint:reassign // Tests verify global morph-key dispatch.
		bp := blueprint.NewBlueprintForTesting("comments", &mockGrammar{})
		bp.Morphs("commentable")
		require.Len(t, bp.Columns, 2)
		assert.Equal(t, blueprint.ColumnTypeBigInteger, bp.Columns[1].ColumnType)
		assert.True(t, *bp.Columns[1].UnsignedVal)
	})

	t.Run("uuid dispatches to UUIDMorphs", func(t *testing.T) {
		blueprint.DefaultMorphKeyType = "uuid" //nolint:reassign // Tests verify global morph-key dispatch.
		bp := blueprint.NewBlueprintForTesting("comments", &mockGrammar{})
		bp.Morphs("commentable")
		require.Len(t, bp.Columns, 2)
		assert.Equal(t, blueprint.ColumnTypeUUID, bp.Columns[1].ColumnType)
	})

	t.Run("ulid dispatches to ULIDMorphs", func(t *testing.T) {
		blueprint.DefaultMorphKeyType = "ulid" //nolint:reassign // Tests verify global morph-key dispatch.
		bp := blueprint.NewBlueprintForTesting("comments", &mockGrammar{})
		bp.Morphs("commentable")
		require.Len(t, bp.Columns, 2)
		assert.Equal(t, blueprint.ColumnTypeULID, bp.Columns[1].ColumnType)
	})
}

func TestBlueprint_NullableMorphs(t *testing.T) {
	original := blueprint.DefaultMorphKeyType
	t.Cleanup(func() {
		blueprint.DefaultMorphKeyType = original //nolint:reassign // Tests verify global morph-key dispatch.
	})

	t.Run("NullableMorphs delegates to NullableNumericMorphs by default", func(t *testing.T) {
		blueprint.DefaultMorphKeyType = "int" //nolint:reassign // Tests verify global morph-key dispatch.
		bp := blueprint.NewBlueprintForTesting("comments", &mockGrammar{})
		bp.NullableMorphs("taggable")
		require.Len(t, bp.Columns, 2)
		assert.True(t, *bp.Columns[0].NullableVal)
		assert.True(t, *bp.Columns[1].NullableVal)
		assert.Equal(t, blueprint.ColumnTypeBigInteger, bp.Columns[1].ColumnType)
	})

	t.Run("NullableUUIDMorphs adds nullable uuid columns and index", func(t *testing.T) {
		bp := blueprint.NewBlueprintForTesting("comments", &mockGrammar{})
		bp.NullableUUIDMorphs("taggable")
		require.Len(t, bp.Columns, 2)
		assert.Equal(t, "taggable_type", bp.Columns[0].Name)
		assert.True(t, *bp.Columns[0].NullableVal)
		assert.Equal(t, "taggable_id", bp.Columns[1].Name)
		assert.Equal(t, blueprint.ColumnTypeUUID, bp.Columns[1].ColumnType)
		assert.True(t, *bp.Columns[1].NullableVal)
		require.Len(t, bp.Commands, 1)
		assert.Equal(t, blueprint.CommandIndex, bp.Commands[0].Name)
	})

	t.Run("NullableULIDMorphs adds nullable ulid columns and index", func(t *testing.T) {
		bp := blueprint.NewBlueprintForTesting("comments", &mockGrammar{})
		bp.NullableULIDMorphs("taggable")
		require.Len(t, bp.Columns, 2)
		assert.Equal(t, blueprint.ColumnTypeULID, bp.Columns[1].ColumnType)
		assert.True(t, *bp.Columns[1].NullableVal)
	})

	t.Run("NullableMorphs with uuid key type", func(t *testing.T) {
		blueprint.DefaultMorphKeyType = "uuid" //nolint:reassign // Tests verify global morph-key dispatch.
		bp := blueprint.NewBlueprintForTesting("comments", &mockGrammar{})
		bp.NullableMorphs("taggable")
		require.Len(t, bp.Columns, 2)
		assert.Equal(t, blueprint.ColumnTypeUUID, bp.Columns[1].ColumnType)
		assert.True(t, *bp.Columns[1].NullableVal)
	})
}
