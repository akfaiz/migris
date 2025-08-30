package schema

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/afkdevs/go-schema/internal/dialect"
	"github.com/afkdevs/go-schema/internal/util"
)

const (
	columnTypeBoolean       string = "boolean"
	columnTypeChar          string = "char"
	columnTypeString        string = "string"
	columnTypeLongText      string = "longText"
	columnTypeMediumText    string = "mediumText"
	columnTypeText          string = "text"
	columnTypeTinyText      string = "tinyText"
	columnTypeBigInteger    string = "bigInteger"
	columnTypeInteger       string = "integer"
	columnTypeMediumInteger string = "mediumInteger"
	columnTypeSmallInteger  string = "smallInteger"
	columnTypeTinyInteger   string = "tinyInteger"
	columnTypeDecimal       string = "decimal"
	columnTypeDouble        string = "double"
	columnTypeFloat         string = "float"
	columnTypeDateTime      string = "dateTime"
	columnTypeDateTimeTz    string = "dateTimeTz"
	columnTypeDate          string = "date"
	columnTypeTime          string = "time"
	columnTypeTimeTz        string = "timeTz"
	columnTypeTimestamp     string = "timestamp"
	columnTypeTimestampTz   string = "timestampTz"
	columnTypeYear          string = "year"
	columnTypeBinary        string = "binary"
	columnTypeJson          string = "json"
	columnTypeJsonb         string = "jsonb"
	columnTypeGeography     string = "geography"
	columnTypeGeometry      string = "geometry"
	columnTypePoint         string = "point"
	columnTypeUuid          string = "uuid"
	columnTypeEnum          string = "enum"
)

const (
	defaultStringLength  int = 255
	defaultTimePrecision int = 0
)

// Blueprint represents a schema blueprint for creating or altering a database table.
type Blueprint struct {
	dialect   dialect.Dialect
	columns   []*columnDefinition
	commands  []*command
	grammar   grammar
	name      string
	charset   string
	collation string
	engine    string
	verbose   bool
}

// Charset sets the character set for the table in the blueprint.
func (b *Blueprint) Charset(charset string) {
	b.charset = charset
}

// Collation sets the collation for the table in the blueprint.
func (b *Blueprint) Collation(collation string) {
	b.collation = collation
}

// Engine sets the storage engine for the table in the blueprint.
func (b *Blueprint) Engine(engine string) {
	b.engine = engine
}

// Column creates a new custom column definition in the blueprint with the specified name and type.
func (b *Blueprint) Column(name string, columnType string) ColumnDefinition {
	return b.addColumn(columnType, name)
}

// Boolean creates a new boolean column definition in the blueprint.
func (b *Blueprint) Boolean(name string) ColumnDefinition {
	return b.addColumn(columnTypeBoolean, name)
}

// Char creates a new char column definition in the blueprint.
func (b *Blueprint) Char(name string, length ...int) ColumnDefinition {
	return b.addColumn(columnTypeChar, name, &columnDefinition{
		length: util.OptionalPtr(defaultStringLength, length...),
	})
}

// String creates a new string column definition in the blueprint.
func (b *Blueprint) String(name string, length ...int) ColumnDefinition {
	return b.addColumn(columnTypeString, name, &columnDefinition{
		length: util.OptionalPtr(defaultStringLength, length...),
	})
}

// LongText creates a new long text column definition in the blueprint.
func (b *Blueprint) LongText(name string) ColumnDefinition {
	return b.addColumn(columnTypeLongText, name)
}

// Text creates a new text column definition in the blueprint.
func (b *Blueprint) Text(name string) ColumnDefinition {
	return b.addColumn(columnTypeText, name)
}

// MediumText creates a new medium text column definition in the blueprint.
func (b *Blueprint) MediumText(name string) ColumnDefinition {
	return b.addColumn(columnTypeMediumText, name)
}

// TinyText creates a new tiny text column definition in the blueprint.
func (b *Blueprint) TinyText(name string) ColumnDefinition {
	return b.addColumn(columnTypeTinyText, name)
}

// BigIncrements creates a new big increments column definition in the blueprint.
func (b *Blueprint) BigIncrements(name string) ColumnDefinition {
	return b.UnsignedBigInteger(name).AutoIncrement()
}

// BigInteger creates a new big integer column definition in the blueprint.
func (b *Blueprint) BigInteger(name string) ColumnDefinition {
	return b.addColumn(columnTypeBigInteger, name)
}

// Decimal creates a new decimal column definition in the blueprint.
//
// The total and places parameters are optional.
//
// Example:
//
//	table.Decimal("price", 10, 2) // creates a decimal column with total 10 and places 2
//
//	table.Decimal("price") // creates a decimal column with default total 8 and places 2
func (b *Blueprint) Decimal(name string, params ...int) ColumnDefinition {
	defaultPlaces := 2
	if len(params) > 1 {
		defaultPlaces = params[1]
	}
	return b.addColumn(columnTypeDecimal, name, &columnDefinition{
		total:  util.OptionalPtr(8, params...),
		places: util.PtrOf(defaultPlaces),
	})
}

// Double creates a new double column definition in the blueprint.
func (b *Blueprint) Double(name string) ColumnDefinition {
	return b.addColumn(columnTypeDouble, name)
}

// Float creates a new float column definition in the blueprint.
func (b *Blueprint) Float(name string, precision ...int) ColumnDefinition {
	return b.addColumn(columnTypeFloat, name, &columnDefinition{
		precision: util.OptionalPtr(53, precision...),
	})
}

// ID creates a new big increments column definition in the blueprint with the name "id" or a custom name.
//
// If a name is provided, it will be used as the column name; otherwise, "id" will be used.
func (b *Blueprint) ID(name ...string) ColumnDefinition {
	return b.BigIncrements(util.Optional("id", name...)).Primary()
}

// Increments create a new increment column definition in the blueprint.
func (b *Blueprint) Increments(name string) ColumnDefinition {
	return b.UnsignedInteger(name).AutoIncrement()
}

// Integer creates a new integer column definition in the blueprint.
func (b *Blueprint) Integer(name string) ColumnDefinition {
	return b.addColumn(columnTypeInteger, name)
}

// MediumIncrements creates a new medium increments column definition in the blueprint.
func (b *Blueprint) MediumIncrements(name string) ColumnDefinition {
	return b.UnsignedMediumInteger(name).AutoIncrement()
}

func (b *Blueprint) MediumInteger(name string) ColumnDefinition {
	return b.addColumn(columnTypeMediumInteger, name)
}

// SmallIncrements creates a new small increments column definition in the blueprint.
func (b *Blueprint) SmallIncrements(name string) ColumnDefinition {
	return b.UnsignedSmallInteger(name).AutoIncrement()
}

// SmallInteger creates a new small integer column definition in the blueprint.
func (b *Blueprint) SmallInteger(name string) ColumnDefinition {
	return b.addColumn(columnTypeSmallInteger, name)
}

// TinyIncrements creates a new tiny increments column definition in the blueprint.
func (b *Blueprint) TinyIncrements(name string) ColumnDefinition {
	return b.UnsignedTinyInteger(name).AutoIncrement()
}

// TinyInteger creates a new tiny integer column definition in the blueprint.
func (b *Blueprint) TinyInteger(name string) ColumnDefinition {
	return b.addColumn(columnTypeTinyInteger, name)
}

// UnsignedBigInteger creates a new unsigned big integer column definition in the blueprint.
func (b *Blueprint) UnsignedBigInteger(name string) ColumnDefinition {
	return b.BigInteger(name).Unsigned()
}

// UnsignedInteger creates a new unsigned integer column definition in the blueprint.
func (b *Blueprint) UnsignedInteger(name string) ColumnDefinition {
	return b.Integer(name).Unsigned()
}

// UnsignedMediumInteger creates a new unsigned medium integer column definition in the blueprint.
func (b *Blueprint) UnsignedMediumInteger(name string) ColumnDefinition {
	return b.MediumInteger(name).Unsigned()
}

// UnsignedSmallInteger creates a new unsigned small integer column definition in the blueprint.
func (b *Blueprint) UnsignedSmallInteger(name string) ColumnDefinition {
	return b.SmallInteger(name).Unsigned()
}

// UnsignedTinyInteger creates a new unsigned tiny integer column definition in the blueprint.
func (b *Blueprint) UnsignedTinyInteger(name string) ColumnDefinition {
	return b.TinyInteger(name).Unsigned()
}

// DateTime creates a new date time column definition in the blueprint.
func (b *Blueprint) DateTime(name string, precision ...int) ColumnDefinition {
	return b.addColumn(columnTypeDateTime, name, &columnDefinition{
		precision: util.OptionalPtr(defaultTimePrecision, precision...),
	})
}

// DateTimeTz creates a new date time with a time zone column definition in the blueprint.
func (b *Blueprint) DateTimeTz(name string, precision ...int) ColumnDefinition {
	return b.addColumn(columnTypeDateTimeTz, name, &columnDefinition{
		precision: util.OptionalPtr(defaultTimePrecision, precision...),
	})
}

// Date creates a new date column definition in the blueprint.
func (b *Blueprint) Date(name string) ColumnDefinition {
	return b.addColumn(columnTypeDate, name)
}

// Time creates a new time column definition in the blueprint.
func (b *Blueprint) Time(name string, precision ...int) ColumnDefinition {
	return b.addColumn(columnTypeTime, name, &columnDefinition{
		precision: util.OptionalPtr(defaultTimePrecision, precision...),
	})
}

// TimeTz creates a new time with time zone column definition in the blueprint.
func (b *Blueprint) TimeTz(name string, precision ...int) ColumnDefinition {
	return b.addColumn(columnTypeTimeTz, name, &columnDefinition{
		precision: util.OptionalPtr(defaultTimePrecision, precision...),
	})
}

// Timestamp creates a new timestamp column definition in the blueprint.
// The precision parameter is optional and defaults to 0 if not provided.
func (b *Blueprint) Timestamp(name string, precision ...int) ColumnDefinition {
	return b.addColumn(columnTypeTimestamp, name, &columnDefinition{
		precision: util.OptionalPtr(defaultTimePrecision, precision...),
	})
}

// TimestampTz creates a new timestamp with time zone column definition in the blueprint.
// The precision parameter is optional and defaults to 0 if not provided.
func (b *Blueprint) TimestampTz(name string, precision ...int) ColumnDefinition {
	return b.addColumn(columnTypeTimestampTz, name, &columnDefinition{
		precision: util.OptionalPtr(defaultTimePrecision, precision...),
	})
}

// Timestamps adds created_at and updated_at timestamp columns to the blueprint.
func (b *Blueprint) Timestamps(precision ...int) {
	b.Timestamp("created_at", precision...).UseCurrent()
	b.Timestamp("updated_at", precision...).UseCurrent().UseCurrentOnUpdate()
}

// TimestampsTz adds created_at and updated_at timestamp with time zone columns to the blueprint.
func (b *Blueprint) TimestampsTz(precision ...int) {
	b.TimestampTz("created_at", precision...).UseCurrent()
	b.TimestampTz("updated_at", precision...).UseCurrent().UseCurrentOnUpdate()
}

// Year creates a new year column definition in the blueprint.
func (b *Blueprint) Year(name string) ColumnDefinition {
	return b.addColumn(columnTypeYear, name)
}

// Binary creates a new binary column definition in the blueprint.
func (b *Blueprint) Binary(name string, length ...int) ColumnDefinition {
	return b.addColumn(columnTypeBinary, name, &columnDefinition{
		length: util.OptionalNil(length...),
	})
}

// JSON creates a new JSON column definition in the blueprint.
func (b *Blueprint) JSON(name string) ColumnDefinition {
	return b.addColumn(columnTypeJson, name)
}

// JSONB creates a new JSONB column definition in the blueprint.
func (b *Blueprint) JSONB(name string) ColumnDefinition {
	return b.addColumn(columnTypeJsonb, name)
}

// UUID creates a new UUID column definition in the blueprint.
func (b *Blueprint) UUID(name string) ColumnDefinition {
	return b.addColumn(columnTypeUuid, name)
}

// Geography creates a new geography column definition in the blueprint.
// The subType parameter is optional and can be used to specify the type of geography (e.g., "Point", "LineString", "Polygon").
// The srid parameter is optional and specifies the Spatial Reference Identifier (SRID) for the geography type.
func (b *Blueprint) Geography(name string, subtype string, srid ...int) ColumnDefinition {
	return b.addColumn(columnTypeGeography, name, &columnDefinition{
		subtype: util.OptionalPtr("", subtype),
		srid:    util.OptionalPtr(4326, srid...),
	})
}

// Geometry creates a new geometry column definition in the blueprint.
// The subType parameter is optional and can be used to specify the type of geometry (e.g., "Point", "LineString", "Polygon").
// The srid parameter is optional and specifies the Spatial Reference Identifier (SRID) for the geometry type.
func (b *Blueprint) Geometry(name string, subtype string, srid ...int) ColumnDefinition {
	return b.addColumn(columnTypeGeometry, name, &columnDefinition{
		subtype: util.OptionalPtr("", subtype),
		srid:    util.OptionalNil(srid...),
	})
}

// Point creates a new point column definition in the blueprint.
func (b *Blueprint) Point(name string, srid ...int) ColumnDefinition {
	return b.addColumn(columnTypePoint, name, &columnDefinition{
		srid: util.OptionalPtr(4326, srid...),
	})
}

// Enum creates a new enum column definition in the blueprint.
// The allowedEnums parameter is a slice of strings that defines the allowed values for the enum column.
//
// Example:
//
//	table.Enum("status", []string{"active", "inactive", "pending"})
//	table.Enum("role", []string{"admin", "user", "guest"}).Comment("User role in the system")
func (b *Blueprint) Enum(name string, allowed []string) ColumnDefinition {
	return b.addColumn(columnTypeEnum, name, &columnDefinition{
		allowed: allowed,
	})
}

// DropTimestamps removes the created_at and updated_at timestamp columns from the blueprint.
func (b *Blueprint) DropTimestamps() {
	b.DropColumn("created_at", "updated_at")
}

// DropTimestampsTz removes the created_at and updated_at timestamp with time zone columns from the blueprint.
func (b *Blueprint) DropTimestampsTz() {
	b.DropTimestamps()
}

// Index creates a new index definition in the blueprint.
//
// Example:
//
//	table.Index("email")
//	table.Index("email", "username") // creates a composite index
//	table.Index("email").Algorithm("btree") // creates a btree index
func (b *Blueprint) Index(column string, otherColumns ...string) IndexDefinition {
	return b.indexCommand(commandIndex, append([]string{column}, otherColumns...)...)
}

// Unique creates a new unique index definition in the blueprint.
//
// Example:
//
//	table.Unique("email")
//	table.Unique("email", "username") // creates a composite unique index
func (b *Blueprint) Unique(column string, otherColumns ...string) IndexDefinition {
	return b.indexCommand(commandUnique, append([]string{column}, otherColumns...)...)
}

// Primary creates a new primary key index definition in the blueprint.
//
// Example:
//
//	table.Primary("id")
//	table.Primary("id", "email") // creates a composite primary key
func (b *Blueprint) Primary(column string, otherColumns ...string) IndexDefinition {
	return b.indexCommand(commandPrimary, append([]string{column}, otherColumns...)...)
}

// FullText creates a new fulltext index definition in the blueprint.
func (b *Blueprint) FullText(column string, otherColumns ...string) IndexDefinition {
	return b.indexCommand(commandFullText, append([]string{column}, otherColumns...)...)
}

// Foreign creates a new foreign key definition in the blueprint.
//
// Example:
//
//	table.Foreign("user_id").References("id").On("users").OnDelete("CASCADE").OnUpdate("CASCADE")
func (b *Blueprint) Foreign(column string) ForeignKeyDefinition {
	command := b.addCommand(commandForeign, &command{
		columns: []string{column},
	})
	return &foreignKeyDefinition{command: command}
}

// DropColumn adds a column to be dropped from the table.
//
// Example:
//
//	table.DropColumn("old_column")
//	table.DropColumn("old_column", "another_old_column") // drops multiple columns
func (b *Blueprint) DropColumn(column string, otherColumns ...string) {
	b.addCommand(commandDropColumn, &command{
		columns: append([]string{column}, otherColumns...),
	})
}

// RenameColumn changes the name of the table in the blueprint.
//
// Example:
//
//	table.RenameColumn("old_table_name", "new_table_name")
func (b *Blueprint) RenameColumn(oldColumn string, newColumn string) {
	b.addCommand(commandRenameColumn, &command{
		from: oldColumn,
		to:   newColumn,
	})
}

// DropIndex adds an index to be dropped from the table.
func (b *Blueprint) DropIndex(index any) {
	b.dropIndexCommand(commandDropIndex, commandIndex, index)
}

// DropForeign adds a foreign key to be dropped from the table.
func (b *Blueprint) DropForeign(index any) {
	b.dropIndexCommand(commandDropForeign, commandForeign, index)
}

// DropPrimary adds a primary key to be dropped from the table.
func (b *Blueprint) DropPrimary(index any) {
	b.dropIndexCommand(commandDropPrimary, commandPrimary, index)
}

// DropUnique adds a unique key to be dropped from the table.
func (b *Blueprint) DropUnique(index any) {
	b.dropIndexCommand(commandDropUnique, commandUnique, index)
}

func (b *Blueprint) DropFulltext(index any) {
	b.dropIndexCommand(commandDropFullText, commandFullText, index)
}

// RenameIndex changes the name of an index in the blueprint.
// Example:
//
//	table.RenameIndex("old_index_name", "new_index_name")
func (b *Blueprint) RenameIndex(oldIndexName string, newIndexName string) {
	b.addCommand(commandRenameIndex, &command{
		from: oldIndexName,
		to:   newIndexName,
	})
}

func (b *Blueprint) getAddedColumns() []*columnDefinition {
	var addedColumns []*columnDefinition
	for _, col := range b.columns {
		if !col.change {
			addedColumns = append(addedColumns, col)
		}
	}
	return addedColumns
}

func (b *Blueprint) getChangedColumns() []*columnDefinition {
	var changedColumns []*columnDefinition
	for _, col := range b.columns {
		if col.change {
			changedColumns = append(changedColumns, col)
		}
	}
	return changedColumns
}

func (b *Blueprint) create() {
	b.addCommand(commandCreate)
}

func (b *Blueprint) creating() bool {
	for _, command := range b.commands {
		if command.name == commandCreate {
			return true
		}
	}
	return false
}

func (b *Blueprint) drop() {
	b.addCommand(commandDrop)
}

func (b *Blueprint) dropIfExists() {
	b.addCommand(commandDropIfExists)
}

func (b *Blueprint) rename(to string) {
	b.addCommand(commandRename, &command{
		to: to,
	})
}

func (b *Blueprint) addImpliedCommands() {
	b.addFluentIndexes()

	if !b.creating() {
		if len(b.getAddedColumns()) > 0 {
			b.commands = append([]*command{{name: commandAdd}}, b.commands...)
		}
		if len(b.getChangedColumns()) > 0 {
			changedCommands := make([]*command, 0, len(b.getChangedColumns()))
			for _, col := range b.getChangedColumns() {
				changedCommands = append(changedCommands, &command{name: commandChange, column: col})
			}
			b.commands = append(changedCommands, b.commands...)
		}
	}
}

func (b *Blueprint) addFluentIndexes() {
	for _, col := range b.columns {
		if col.primary != nil {
			if b.dialect == dialect.MySQL {
				continue
			}
			if !*col.primary && col.change {
				b.DropPrimary([]string{col.name})
				col.primary = nil
			}
		}
		if col.index != nil {
			if *col.index {
				b.Index(col.name).Name(col.indexName)
				col.index = nil
			} else if !*col.index && col.change {
				b.DropIndex([]string{col.name})
				col.index = nil
			}
		}

		if col.unique != nil {
			if *col.unique {
				b.Unique(col.name).Name(col.uniqueName)
				col.unique = nil
			} else if !*col.unique && col.change {
				b.DropUnique([]string{col.name})
				col.unique = nil
			}
		}
	}
}

func (b *Blueprint) getFluentStatements() []string {
	var statements []string
	for _, column := range b.columns {
		for _, fluentCommand := range b.grammar.GetFluentCommands() {
			if statement := fluentCommand(b, &command{column: column}); statement != "" {
				statements = append(statements, statement)
			}
		}
	}
	return statements
}

func (b *Blueprint) build(ctx context.Context, tx *sql.Tx) error {
	statements, err := b.toSql()
	if err != nil {
		return err
	}
	for _, statement := range statements {
		if b.verbose {
			log.Println(statement)
		}
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

func (b *Blueprint) toSql() ([]string, error) {
	b.addImpliedCommands()

	var statements []string

	mainCommandMap := map[string]func(blueprint *Blueprint) (string, error){
		commandCreate:       b.grammar.CompileCreate,
		commandAdd:          b.grammar.CompileAdd,
		commandDrop:         b.grammar.CompileDrop,
		commandDropIfExists: b.grammar.CompileDropIfExists,
	}
	secondaryCommandMap := map[string]func(blueprint *Blueprint, command *command) (string, error){
		commandChange:       b.grammar.CompileChange,
		commandDropColumn:   b.grammar.CompileDropColumn,
		commandDropIndex:    b.grammar.CompileDropIndex,
		commandDropForeign:  b.grammar.CompileDropForeign,
		commandDropFullText: b.grammar.CompileDropFulltext,
		commandDropPrimary:  b.grammar.CompileDropPrimary,
		commandDropUnique:   b.grammar.CompileDropUnique,
		commandForeign:      b.grammar.CompileForeign,
		commandFullText:     b.grammar.CompileFullText,
		commandIndex:        b.grammar.CompileIndex,
		commandPrimary:      b.grammar.CompilePrimary,
		commandRename:       b.grammar.CompileRename,
		commandRenameColumn: b.grammar.CompileRenameColumn,
		commandRenameIndex:  b.grammar.CompileRenameIndex,
		commandUnique:       b.grammar.CompileUnique,
	}
	for _, cmd := range b.commands {
		if compileFunc, exists := mainCommandMap[cmd.name]; exists {
			sql, err := compileFunc(b)
			if err != nil {
				return nil, err
			}
			if sql != "" {
				statements = append(statements, sql)
			}
			continue
		}
		if compileFunc, exists := secondaryCommandMap[cmd.name]; exists {
			sql, err := compileFunc(b, cmd)
			if err != nil {
				return nil, err
			}
			if sql != "" {
				statements = append(statements, sql)
			}
			continue
		}
		return nil, fmt.Errorf("unknown command: %s", cmd.name)
	}

	statements = append(statements, b.getFluentStatements()...)

	return statements, nil
}

func (b *Blueprint) addColumn(colType string, name string, columnDefs ...*columnDefinition) *columnDefinition {
	var col *columnDefinition
	if len(columnDefs) > 0 {
		col = columnDefs[0]
	} else {
		col = &columnDefinition{}
	}
	col.columnType = colType
	col.name = name

	return b.addColumnDefinition(col)
}

func (b *Blueprint) addColumnDefinition(col *columnDefinition) *columnDefinition {
	b.columns = append(b.columns, col)
	return col
}

func (b *Blueprint) indexCommand(name string, columns ...string) IndexDefinition {
	command := b.addCommand(name, &command{
		columns: columns,
	})
	return &indexDefinition{command}
}

func (b *Blueprint) dropIndexCommand(name string, indexType string, index any) {
	switch index := index.(type) {
	case string:
		b.addCommand(name, &command{
			index: index,
		})
	case []string:
		indexName := b.grammar.CreateIndexName(b, indexType, index...)
		b.addCommand(name, &command{
			index: indexName,
		})
	default:
		panic(fmt.Sprintf("unsupported index type: %T", index))
	}
}

func (b *Blueprint) addCommand(name string, parameters ...*command) *command {
	var parameter *command
	if len(parameters) > 0 {
		parameter = parameters[0]
	} else {
		parameter = &command{}
	}
	parameter.name = name
	b.commands = append(b.commands, parameter)

	return parameter
}
