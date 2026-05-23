package blueprint

import (
	"fmt"

	"github.com/akfaiz/migris/internal/dialect"
	"github.com/akfaiz/migris/internal/util"
	"github.com/akfaiz/migris/schema/core"
)

const (
	ColumnTypeBoolean       string = "boolean"
	ColumnTypeChar          string = "char"
	ColumnTypeString        string = "string"
	ColumnTypeLongText      string = "longText"
	ColumnTypeMediumText    string = "mediumText"
	ColumnTypeText          string = "text"
	ColumnTypeTinyText      string = "tinyText"
	ColumnTypeBigInteger    string = "bigInteger"
	ColumnTypeInteger       string = "integer"
	ColumnTypeMediumInteger string = "mediumInteger"
	ColumnTypeSmallInteger  string = "smallInteger"
	ColumnTypeTinyInteger   string = "tinyInteger"
	ColumnTypeDecimal       string = "decimal"
	ColumnTypeDouble        string = "double"
	ColumnTypeFloat         string = "float"
	ColumnTypeDateTime      string = "dateTime"
	ColumnTypeDateTimeTz    string = "dateTimeTz"
	ColumnTypeDate          string = "date"
	ColumnTypeTime          string = "time"
	ColumnTypeTimeTz        string = "timeTz"
	ColumnTypeTimestamp     string = "timestamp"
	ColumnTypeTimestampTz   string = "timestampTz"
	ColumnTypeYear          string = "year"
	ColumnTypeBinary        string = "binary"
	ColumnTypeJSON          string = "json"
	ColumnTypeJSONB         string = "jsonb"
	ColumnTypeGeography     string = "geography"
	ColumnTypeGeometry      string = "geometry"
	ColumnTypePoint         string = "point"
	ColumnTypeUUID          string = "uuid"
	ColumnTypeULID          string = "ulid"
	ColumnTypeEnum          string = "enum"
	ColumnTypeSet           string = "set"
	ColumnTypeIPAddress     string = "ipAddress"
	ColumnTypeMacAddress    string = "macAddress"
	ColumnTypeVector        string = "vector"
	ColumnTypeTSVector      string = "tsvector"
	ColumnTypeCidr          string = "cidr"
	ColumnTypeInet          string = "inet"
	ColumnTypeMacaddr       string = "macaddr"
	ColumnTypeMacaddr8      string = "macaddr8"
	ColumnTypeRaw           string = "raw"
)

const (
	defaultStringLength  int = 255
	defaultTimePrecision int = 0
)

// DefaultMorphKeyType controls the key type used by Morphs/NullableMorphs ("int", "uuid", "ulid").
var DefaultMorphKeyType = "int"

// Blueprint represents a schema blueprint for creating or altering a database table.
type Blueprint struct {
	Dialect                        dialect.Dialect
	Columns                        []*Column
	Commands                       []*Command
	Grammar                        Grammar
	Builder                        Builder
	Name                           string
	CharsetVal                     string
	CollationVal                   string
	EngineVal                      string
	CommentVal                     string
	TemporaryVal                   bool
	AutoIncrementStartingValuesVal *int
	afterColumn                    string
}

// Builder interface.
type Builder interface {
	GetColumns(ctx core.Context, table string) ([]*core.Column, error)
}

// NewBlueprintForTesting creates a new blueprint for testing purposes.
func NewBlueprintForTesting(name string, g Grammar) *Blueprint {
	return &Blueprint{Name: name, Grammar: g}
}

func (b *Blueprint) SetDialect(d dialect.Dialect) {
	b.Dialect = d
}

func (b *Blueprint) SetBuilder(builder Builder) {
	b.Builder = builder
}

func (b *Blueprint) SetGrammar(g Grammar) {
	b.Grammar = g
}

func (b *Blueprint) SetName(name string) {
	b.Name = name
}

func (b *Blueprint) Charset(charset string) {
	b.CharsetVal = charset
}

func (b *Blueprint) Collation(collation string) {
	b.CollationVal = collation
}

func (b *Blueprint) Engine(engine string) {
	b.EngineVal = engine
}

func (b *Blueprint) Comment(comment string) {
	b.CommentVal = comment
	b.addCommand(CommandTableComment, &Command{Comment: comment})
}

func (b *Blueprint) AutoIncrementStartingValues(value int) {
	b.AutoIncrementStartingValuesVal = &value
	b.addCommand(CommandAutoIncrementStartingValues, &Command{Value: value})
}

func (b *Blueprint) Column(name string, columnType string) ColumnDefinition {
	return b.addColumn(columnType, name)
}

func (b *Blueprint) Boolean(name string) ColumnDefinition {
	return b.addColumn(ColumnTypeBoolean, name)
}

func (b *Blueprint) Char(name string, length ...int) ColumnDefinition {
	return b.addColumn(ColumnTypeChar, name, &Column{
		Length: util.OptionalPtr(defaultStringLength, length...),
	})
}

func (b *Blueprint) String(name string, length ...int) ColumnDefinition {
	return b.addColumn(ColumnTypeString, name, &Column{
		Length: util.OptionalPtr(defaultStringLength, length...),
	})
}

func (b *Blueprint) LongText(name string) ColumnDefinition {
	return b.addColumn(ColumnTypeLongText, name)
}

func (b *Blueprint) Text(name string) ColumnDefinition {
	return b.addColumn(ColumnTypeText, name)
}

func (b *Blueprint) MediumText(name string) ColumnDefinition {
	return b.addColumn(ColumnTypeMediumText, name)
}

func (b *Blueprint) TinyText(name string) ColumnDefinition {
	return b.addColumn(ColumnTypeTinyText, name)
}

func (b *Blueprint) BigIncrements(name string) ColumnDefinition {
	return b.UnsignedBigInteger(name).AutoIncrement()
}

func (b *Blueprint) BigInteger(name string) ColumnDefinition {
	return b.addColumn(ColumnTypeBigInteger, name)
}

func (b *Blueprint) Decimal(name string, params ...int) ColumnDefinition {
	defaultPlaces := 2
	if len(params) > 1 {
		defaultPlaces = params[1]
	}
	return b.addColumn(ColumnTypeDecimal, name, &Column{
		Total:  util.OptionalPtr(8, params...),
		Places: util.PtrOf(defaultPlaces),
	})
}

func (b *Blueprint) Double(name string) ColumnDefinition {
	return b.addColumn(ColumnTypeDouble, name)
}

func (b *Blueprint) Float(name string, precision ...int) ColumnDefinition {
	return b.addColumn(ColumnTypeFloat, name, &Column{
		Precision: util.OptionalPtr(53, precision...),
	})
}

func (b *Blueprint) ID(name ...string) ColumnDefinition {
	return b.BigIncrements(util.Optional("id", name...)).Primary()
}

func (b *Blueprint) Increments(name string) ColumnDefinition {
	return b.UnsignedInteger(name).AutoIncrement()
}

func (b *Blueprint) Integer(name string) ColumnDefinition {
	return b.addColumn(ColumnTypeInteger, name)
}

func (b *Blueprint) MediumIncrements(name string) ColumnDefinition {
	return b.UnsignedMediumInteger(name).AutoIncrement()
}

func (b *Blueprint) MediumInteger(name string) ColumnDefinition {
	return b.addColumn(ColumnTypeMediumInteger, name)
}

func (b *Blueprint) SmallIncrements(name string) ColumnDefinition {
	return b.UnsignedSmallInteger(name).AutoIncrement()
}

func (b *Blueprint) SmallInteger(name string) ColumnDefinition {
	return b.addColumn(ColumnTypeSmallInteger, name)
}

func (b *Blueprint) TinyIncrements(name string) ColumnDefinition {
	return b.UnsignedTinyInteger(name).AutoIncrement()
}

func (b *Blueprint) TinyInteger(name string) ColumnDefinition {
	return b.addColumn(ColumnTypeTinyInteger, name)
}

func (b *Blueprint) UnsignedBigInteger(name string) ColumnDefinition {
	return b.BigInteger(name).Unsigned()
}

func (b *Blueprint) UnsignedInteger(name string) ColumnDefinition {
	return b.Integer(name).Unsigned()
}

func (b *Blueprint) UnsignedMediumInteger(name string) ColumnDefinition {
	return b.MediumInteger(name).Unsigned()
}

func (b *Blueprint) UnsignedSmallInteger(name string) ColumnDefinition {
	return b.SmallInteger(name).Unsigned()
}

func (b *Blueprint) UnsignedTinyInteger(name string) ColumnDefinition {
	return b.TinyInteger(name).Unsigned()
}

func (b *Blueprint) DateTime(name string, precision ...int) ColumnDefinition {
	return b.addColumn(ColumnTypeDateTime, name, &Column{
		Precision: util.OptionalPtr(defaultTimePrecision, precision...),
	})
}

func (b *Blueprint) DateTimeTz(name string, precision ...int) ColumnDefinition {
	return b.addColumn(ColumnTypeDateTimeTz, name, &Column{
		Precision: util.OptionalPtr(defaultTimePrecision, precision...),
	})
}

func (b *Blueprint) Date(name string) ColumnDefinition {
	return b.addColumn(ColumnTypeDate, name)
}

func (b *Blueprint) Time(name string, precision ...int) ColumnDefinition {
	return b.addColumn(ColumnTypeTime, name, &Column{
		Precision: util.OptionalPtr(defaultTimePrecision, precision...),
	})
}

func (b *Blueprint) TimeTz(name string, precision ...int) ColumnDefinition {
	return b.addColumn(ColumnTypeTimeTz, name, &Column{
		Precision: util.OptionalPtr(defaultTimePrecision, precision...),
	})
}

func (b *Blueprint) Timestamp(name string, precision ...int) ColumnDefinition {
	return b.addColumn(ColumnTypeTimestamp, name, &Column{
		Precision: util.OptionalPtr(defaultTimePrecision, precision...),
	})
}

func (b *Blueprint) TimestampTz(name string, precision ...int) ColumnDefinition {
	return b.addColumn(ColumnTypeTimestampTz, name, &Column{
		Precision: util.OptionalPtr(defaultTimePrecision, precision...),
	})
}

func (b *Blueprint) Timestamps(precision ...int) {
	b.Timestamp("created_at", precision...).Nullable()
	b.Timestamp("updated_at", precision...).Nullable()
}

func (b *Blueprint) TimestampsTz(precision ...int) {
	b.TimestampTz("created_at", precision...).Nullable()
	b.TimestampTz("updated_at", precision...).Nullable()
}

func (b *Blueprint) Year(name string) ColumnDefinition {
	return b.addColumn(ColumnTypeYear, name)
}

func (b *Blueprint) Binary(name string, length ...int) ColumnDefinition {
	return b.addColumn(ColumnTypeBinary, name, &Column{
		Length: util.OptionalNil(length...),
	})
}

func (b *Blueprint) JSON(name string) ColumnDefinition {
	return b.addColumn(ColumnTypeJSON, name)
}

func (b *Blueprint) JSONB(name string) ColumnDefinition {
	return b.addColumn(ColumnTypeJSONB, name)
}

func (b *Blueprint) UUID(name string) ColumnDefinition {
	return b.addColumn(ColumnTypeUUID, name)
}

func (b *Blueprint) ULID(name string) ColumnDefinition {
	return b.addColumn(ColumnTypeULID, name)
}

func (b *Blueprint) Geography(name string, subtype string, srid ...int) ColumnDefinition {
	return b.addColumn(ColumnTypeGeography, name, &Column{
		Subtype: util.OptionalPtr("", subtype),
		Srid:    util.OptionalPtr(4326, srid...),
	})
}

func (b *Blueprint) Geometry(name string, subtype string, srid ...int) ColumnDefinition {
	return b.addColumn(ColumnTypeGeometry, name, &Column{
		Subtype: util.OptionalPtr("", subtype),
		Srid:    util.OptionalNil(srid...),
	})
}

func (b *Blueprint) Point(name string, srid ...int) ColumnDefinition {
	return b.addColumn(ColumnTypePoint, name, &Column{
		Srid: util.OptionalPtr(4326, srid...),
	})
}

func (b *Blueprint) Enum(name string, allowed []string) ColumnDefinition {
	return b.addColumn(ColumnTypeEnum, name, &Column{
		Allowed: allowed,
	})
}

func (b *Blueprint) Set(name string, allowed []string) ColumnDefinition {
	return b.addColumn(ColumnTypeSet, name, &Column{
		Allowed: allowed,
	})
}

func (b *Blueprint) IPAddress(name string) ColumnDefinition {
	return b.addColumn(ColumnTypeIPAddress, name)
}

func (b *Blueprint) MacAddress(name string) ColumnDefinition {
	return b.addColumn(ColumnTypeMacAddress, name)
}

func (b *Blueprint) Vector(name string, dimensions ...int) ColumnDefinition {
	return b.addColumn(ColumnTypeVector, name, &Column{
		Places: util.OptionalNil(dimensions...),
	})
}

func (b *Blueprint) TSVector(name string) ColumnDefinition {
	return b.addColumn(ColumnTypeTSVector, name)
}

func (b *Blueprint) Cidr(name string) ColumnDefinition {
	return b.addColumn(ColumnTypeCidr, name)
}

func (b *Blueprint) Inet(name string) ColumnDefinition {
	return b.addColumn(ColumnTypeInet, name)
}

func (b *Blueprint) MacAddr(name string) ColumnDefinition {
	return b.addColumn(ColumnTypeMacaddr, name)
}

func (b *Blueprint) MacAddr8(name string) ColumnDefinition {
	return b.addColumn(ColumnTypeMacaddr8, name)
}

func (b *Blueprint) Temporary() {
	b.TemporaryVal = true
}

func (b *Blueprint) InnoDB() {
	b.Engine("InnoDB")
}

func (b *Blueprint) IntegerIncrements(name string) ColumnDefinition {
	return b.Increments(name)
}

func (b *Blueprint) SoftDeletes(name string, precision ...int) ColumnDefinition {
	return b.Timestamp(util.Optional("deleted_at", name), precision...).Nullable()
}

func (b *Blueprint) SoftDeletesTz(name string, precision ...int) ColumnDefinition {
	return b.TimestampTz(util.Optional("deleted_at", name), precision...).Nullable()
}

func (b *Blueprint) SoftDeletesDatetime(name string, precision ...int) ColumnDefinition {
	return b.DateTime(util.Optional("deleted_at", name), precision...).Nullable()
}

func (b *Blueprint) DropSoftDeletes(name ...string) {
	b.DropColumn(util.Optional("deleted_at", name...))
}

func (b *Blueprint) DropSoftDeletesTz(name ...string) {
	b.DropSoftDeletes(name...)
}

func (b *Blueprint) NullableTimestamps(precision ...int) {
	b.Timestamps(precision...)
}

func (b *Blueprint) NullableTimestampsTz(precision ...int) {
	b.TimestampsTz(precision...)
}

func (b *Blueprint) Datetimes(precision ...int) {
	b.DateTime("created_at", precision...).Nullable()
	b.DateTime("updated_at", precision...).Nullable()
}

func (b *Blueprint) RawColumn(name string, definition string) ColumnDefinition {
	return b.addColumn(ColumnTypeRaw, name, &Column{RawDefinition: &definition})
}

func (b *Blueprint) RawIndex(expression string, name string) IndexDefinition {
	return b.indexCommand(CommandIndex, expression).Name(name)
}

func (b *Blueprint) After(column string, callback func(*Blueprint)) {
	b.afterColumn = column
	callback(b)
	b.afterColumn = ""
}

func (b *Blueprint) RemoveColumn(name string) {
	filtered := b.Columns[:0]
	for _, col := range b.Columns {
		if col.Name != name {
			filtered = append(filtered, col)
		}
	}
	b.Columns = filtered
}

func (b *Blueprint) SpatialIndex(column string, otherColumns ...string) IndexDefinition {
	return b.indexCommand(CommandSpatialIndex, append([]string{column}, otherColumns...)...)
}

func (b *Blueprint) VectorIndex(column string) IndexDefinition {
	return b.indexCommand(CommandVectorIndex, column)
}

func (b *Blueprint) DropSpatialIndex(index any) {
	b.dropIndexCommand(CommandDropSpatialIndex, CommandSpatialIndex, index)
}

func (b *Blueprint) Morphs(name string, indexName ...string) {
	switch DefaultMorphKeyType {
	case "uuid":
		b.UUIDMorphs(name, indexName...)
	case "ulid":
		b.ULIDMorphs(name, indexName...)
	default:
		b.NumericMorphs(name, indexName...)
	}
}

func (b *Blueprint) NullableMorphs(name string, indexName ...string) {
	switch DefaultMorphKeyType {
	case "uuid":
		b.NullableUUIDMorphs(name, indexName...)
	case "ulid":
		b.NullableULIDMorphs(name, indexName...)
	default:
		b.NullableNumericMorphs(name, indexName...)
	}
}

func (b *Blueprint) NumericMorphs(name string, indexName ...string) {
	b.String(name + "_type")
	b.UnsignedBigInteger(name + "_id")
	b.Index(name+"_type", name+"_id").Name(util.Optional("", indexName...))
}

func (b *Blueprint) NullableNumericMorphs(name string, indexName ...string) {
	b.String(name + "_type").Nullable()
	b.UnsignedBigInteger(name + "_id").Nullable()
	b.Index(name+"_type", name+"_id").Name(util.Optional("", indexName...))
}

func (b *Blueprint) UUIDMorphs(name string, indexName ...string) {
	b.String(name + "_type")
	b.UUID(name + "_id")
	b.Index(name+"_type", name+"_id").Name(util.Optional("", indexName...))
}

func (b *Blueprint) NullableUUIDMorphs(name string, indexName ...string) {
	b.String(name + "_type").Nullable()
	b.UUID(name + "_id").Nullable()
	b.Index(name+"_type", name+"_id").Name(util.Optional("", indexName...))
}

func (b *Blueprint) ULIDMorphs(name string, indexName ...string) {
	b.String(name + "_type")
	b.ULID(name + "_id")
	b.Index(name+"_type", name+"_id").Name(util.Optional("", indexName...))
}

func (b *Blueprint) NullableULIDMorphs(name string, indexName ...string) {
	b.String(name + "_type").Nullable()
	b.ULID(name + "_id").Nullable()
	b.Index(name+"_type", name+"_id").Name(util.Optional("", indexName...))
}

func (b *Blueprint) DropMorphs(name string, indexName ...string) {
	idx := util.Optional("", indexName...)
	if idx == "" {
		idx = b.Grammar.CreateIndexName(b, "index", name+"_type", name+"_id")
	}
	b.DropIndex(idx)
	b.DropColumn(name+"_type", name+"_id")
}

func (b *Blueprint) ForeignID(name string) ForeignIDColumnDefinition {
	col := b.addColumn(ColumnTypeBigInteger, name)
	col.UnsignedVal = util.PtrOf(true)
	return &foreignIDColumnDefinition{column: col, blueprint: b}
}

func (b *Blueprint) ForeignUUID(name string) ForeignIDColumnDefinition {
	col := b.addColumn(ColumnTypeUUID, name)
	return &foreignIDColumnDefinition{column: col, blueprint: b}
}

func (b *Blueprint) ForeignULID(name string, length ...int) ForeignIDColumnDefinition {
	col := b.addColumn(ColumnTypeChar, name, &Column{
		Length: util.OptionalPtr(26, length...),
	})
	return &foreignIDColumnDefinition{column: col, blueprint: b}
}

func (b *Blueprint) DropConstrainedForeignID(name string) {
	b.DropForeign([]string{name})
	b.DropColumn(name)
}

func (b *Blueprint) DropTimestamps() {
	b.DropColumn("created_at", "updated_at")
}

func (b *Blueprint) DropTimestampsTz() {
	b.DropTimestamps()
}

func (b *Blueprint) Index(column string, otherColumns ...string) IndexDefinition {
	return b.indexCommand(CommandIndex, append([]string{column}, otherColumns...)...)
}

func (b *Blueprint) Unique(column string, otherColumns ...string) IndexDefinition {
	return b.indexCommand(CommandUnique, append([]string{column}, otherColumns...)...)
}

func (b *Blueprint) Primary(column string, otherColumns ...string) IndexDefinition {
	return b.indexCommand(CommandPrimary, append([]string{column}, otherColumns...)...)
}

func (b *Blueprint) FullText(column string, otherColumns ...string) IndexDefinition {
	return b.indexCommand(CommandFullText, append([]string{column}, otherColumns...)...)
}

func (b *Blueprint) Foreign(column string) ForeignKeyDefinition {
	return b.ForeignColumns(column)
}

func (b *Blueprint) ForeignColumns(columns ...string) ForeignKeyDefinition {
	command := b.addCommand(CommandForeign, &Command{
		Columns: columns,
	})
	return &foreignKeyDefinition{command: command}
}

func (b *Blueprint) DropColumn(column string, otherColumns ...string) {
	b.addCommand(CommandDropColumn, &Command{
		Columns: append([]string{column}, otherColumns...),
	})
}

func (b *Blueprint) RenameColumn(oldColumn string, newColumn string) {
	b.addCommand(CommandRenameColumn, &Command{
		From: oldColumn,
		To:   newColumn,
	})
}

func (b *Blueprint) DropIndex(index any) {
	b.dropIndexCommand(CommandDropIndex, CommandIndex, index)
}

func (b *Blueprint) DropForeign(index any) {
	b.dropIndexCommand(CommandDropForeign, CommandForeign, index)
}

func (b *Blueprint) DropPrimary(index any) {
	b.dropIndexCommand(CommandDropPrimary, CommandPrimary, index)
}

func (b *Blueprint) DropUnique(index any) {
	b.dropIndexCommand(CommandDropUnique, CommandUnique, index)
}

func (b *Blueprint) DropFulltext(index any) {
	b.dropIndexCommand(CommandDropFullText, CommandFullText, index)
}

func (b *Blueprint) RenameIndex(oldIndexName string, newIndexName string) {
	b.addCommand(CommandRenameIndex, &Command{
		From: oldIndexName,
		To:   newIndexName,
	})
}

func (b *Blueprint) GetAddedColumns() []*Column {
	var addedColumns []*Column
	for _, col := range b.Columns {
		if !col.ChangeVal {
			addedColumns = append(addedColumns, col)
		}
	}
	return addedColumns
}

func (b *Blueprint) GetChangedColumns() []*Column {
	var changedColumns []*Column
	for _, col := range b.Columns {
		if col.ChangeVal {
			changedColumns = append(changedColumns, col)
		}
	}
	return changedColumns
}

func (b *Blueprint) Create() {
	b.addCommand(CommandCreate)
}

func (b *Blueprint) IsCreating() bool {
	for _, command := range b.Commands {
		if command.Name == CommandCreate {
			return true
		}
	}
	return false
}

func (b *Blueprint) Drop() {
	b.addCommand(CommandDrop)
}

func (b *Blueprint) DropIfExists() {
	b.addCommand(CommandDropIfExists)
}

func (b *Blueprint) Rename(to string) {
	b.addCommand(CommandRename, &Command{
		To: to,
	})
}

func (b *Blueprint) AddImpliedCommands() {
	b.addFluentIndexes()

	if !b.IsCreating() {
		if len(b.GetAddedColumns()) > 0 {
			b.Commands = append([]*Command{{Name: CommandAdd}}, b.Commands...)
		}
		if len(b.GetChangedColumns()) > 0 {
			changedCommands := make([]*Command, 0, len(b.GetChangedColumns()))
			for _, col := range b.GetChangedColumns() {
				changedCommands = append(changedCommands, &Command{Name: CommandChange, Column: col})
			}
			b.Commands = append(changedCommands, b.Commands...)
		}
	}
}

func (b *Blueprint) addFluentIndexes() {
	for _, col := range b.Columns {
		skipped := b.addFluentIndexPrimary(col)
		if skipped {
			continue
		}
		b.addFluentIndexIndex(col)
		b.addFluentIndexUnique(col)
	}
}

func (b *Blueprint) addFluentIndexPrimary(col *Column) bool {
	if col.PrimaryVal != nil {
		if b.Dialect == dialect.MySQL {
			return true
		}
		if !*col.PrimaryVal && col.ChangeVal {
			b.DropPrimary([]string{col.Name})
			col.PrimaryVal = nil
		}
	}
	return false
}

func (b *Blueprint) addFluentIndexIndex(col *Column) {
	if col.IndexVal != nil {
		if *col.IndexVal {
			b.Index(col.Name).Name(col.IndexName)
			col.IndexVal = nil
		} else if !*col.IndexVal && col.ChangeVal {
			b.DropIndex([]string{col.Name})
			col.IndexVal = nil
		}
	}
}

func (b *Blueprint) addFluentIndexUnique(col *Column) {
	if col.UniqueVal != nil {
		if *col.UniqueVal {
			b.Unique(col.Name).Name(col.UniqueName)
			col.UniqueVal = nil
		} else if !*col.UniqueVal && col.ChangeVal {
			b.DropUnique([]string{col.Name})
			col.UniqueVal = nil
		}
	}
}

func (b *Blueprint) GetFluentStatements() []string {
	var statements []string
	for _, column := range b.Columns {
		for _, fluentCommand := range b.Grammar.GetFluentCommands() {
			if statement := fluentCommand(b, &Command{Column: column}); statement != "" {
				statements = append(statements, statement)
			}
		}
	}

	for _, tableFluentCommand := range b.Grammar.GetTableFluentCommands() {
		if statement := tableFluentCommand(b); statement != "" {
			statements = append(statements, statement)
		}
	}

	return statements
}

func (b *Blueprint) Build(ctx core.Context) error {
	if err := b.hydrate(ctx); err != nil {
		return err
	}

	statements, err := b.ToSQL()
	if err != nil {
		return err
	}

	for _, statement := range statements {
		if _, execErr := ctx.Exec(statement); execErr != nil {
			return execErr
		}
	}

	return nil
}

func (b *Blueprint) hydrate(ctx core.Context) error {
	changedColumns := b.GetChangedColumns()
	if len(changedColumns) == 0 {
		return nil
	}

	if b.Builder == nil {
		return nil
	}

	existingColumns, err := b.Builder.GetColumns(ctx, b.Name)
	if err != nil {
		return err
	}

	columnsMap := make(map[string]*core.Column)
	for _, col := range existingColumns {
		columnsMap[col.Name] = col
	}

	for _, colDef := range changedColumns {
		if existing, ok := columnsMap[colDef.Name]; ok {
			b.mergeColumnMetadata(colDef, existing)
		}
	}
	return nil
}

func (b *Blueprint) mergeColumnMetadata(colDef *Column, existing *core.Column) {
	if !colDef.HasCommand("nullable") {
		colDef.Nullable(existing.Nullable)
	}

	if !colDef.HasCommand("default") && existing.DefaultVal.Valid {
		colDef.Default(Expression(existing.DefaultVal.String))
	}

	if !colDef.HasCommand("comment") && existing.Comment.Valid {
		colDef.Comment(existing.Comment.String)
	}
}

func (b *Blueprint) ToSQL() ([]string, error) {
	b.AddImpliedCommands()

	var statements []string

	mainCommandMap := map[string]func(blueprint *Blueprint) (string, error){
		CommandCreate:       b.Grammar.CompileCreate,
		CommandAdd:          b.Grammar.CompileAdd,
		CommandDrop:         b.Grammar.CompileDrop,
		CommandDropIfExists: b.Grammar.CompileDropIfExists,
	}
	secondaryCommandMap := map[string]func(blueprint *Blueprint, command *Command) (string, error){
		CommandChange:                      b.Grammar.CompileChange,
		CommandDropColumn:                  b.Grammar.CompileDropColumn,
		CommandDropIndex:                   b.Grammar.CompileDropIndex,
		CommandDropForeign:                 b.Grammar.CompileDropForeign,
		CommandDropFullText:                b.Grammar.CompileDropFulltext,
		CommandDropPrimary:                 b.Grammar.CompileDropPrimary,
		CommandDropUnique:                  b.Grammar.CompileDropUnique,
		CommandDropSpatialIndex:            b.Grammar.CompileDropSpatialIndex,
		CommandForeign:                     b.Grammar.CompileForeign,
		CommandFullText:                    b.Grammar.CompileFullText,
		CommandIndex:                       b.Grammar.CompileIndex,
		CommandPrimary:                     b.Grammar.CompilePrimary,
		CommandRename:                      b.Grammar.CompileRename,
		CommandRenameColumn:                b.Grammar.CompileRenameColumn,
		CommandRenameIndex:                 b.Grammar.CompileRenameIndex,
		CommandUnique:                      b.Grammar.CompileUnique,
		CommandSpatialIndex:                b.Grammar.CompileSpatialIndex,
		CommandVectorIndex:                 b.Grammar.CompileVectorIndex,
		CommandTableComment:                b.Grammar.CompileTableComment,
		CommandAutoIncrementStartingValues: b.Grammar.CompileAutoIncrementStartingValues,
	}
	for _, cmd := range b.Commands {
		if compileFunc, exists := mainCommandMap[cmd.Name]; exists {
			sql, err := compileFunc(b)
			if err != nil {
				return nil, err
			}
			if sql != "" {
				statements = append(statements, sql)
			}
			continue
		}
		if compileFunc, exists := secondaryCommandMap[cmd.Name]; exists {
			sql, err := compileFunc(b, cmd)
			if err != nil {
				return nil, err
			}
			if sql != "" {
				statements = append(statements, sql)
			}
			continue
		}
		return nil, fmt.Errorf("unknown command: %s", cmd.Name)
	}

	statements = append(statements, b.GetFluentStatements()...)

	return statements, nil
}

func (b *Blueprint) addColumn(colType string, name string, columnDefs ...*Column) *Column {
	var col *Column
	if len(columnDefs) > 0 {
		col = columnDefs[0]
	} else {
		col = &Column{}
	}
	col.ColumnType = colType
	col.Name = name

	return b.addColumnDefinition(col)
}

func (b *Blueprint) addColumnDefinition(col *Column) *Column {
	b.Columns = append(b.Columns, col)
	if b.afterColumn != "" {
		after := b.afterColumn
		col.AfterVal = &after
		b.afterColumn = col.Name
	}
	return col
}

func (b *Blueprint) indexCommand(name string, columns ...string) IndexDefinition {
	command := b.addCommand(name, &Command{
		Columns: columns,
	})
	return &indexDefinition{command: command}
}

func (b *Blueprint) dropIndexCommand(name string, indexType string, index any) {
	switch index := index.(type) {
	case string:
		b.addCommand(name, &Command{
			Index: index,
		})
	case []string:
		indexName := b.Grammar.CreateIndexName(b, indexType, index...)
		b.addCommand(name, &Command{
			Index: indexName,
		})
	default:
		panic(fmt.Sprintf("unsupported index type: %T", index))
	}
}

func (b *Blueprint) addCommand(name string, parameters ...*Command) *Command {
	var parameter *Command
	if len(parameters) > 0 {
		parameter = parameters[0]
	} else {
		parameter = &Command{}
	}
	parameter.Name = name
	b.Commands = append(b.Commands, parameter)

	return parameter
}
