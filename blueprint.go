package schema

import (
	"fmt"
	"slices"
)

type columnType uint8

const (
	columnTypeBoolean columnType = iota
	columnTypeChar
	columnTypeString
	columnTypeLongText
	columnTypeMediumText
	columnTypeText
	columnTypeTinyText
	columnTypeBigInteger
	columnTypeInteger
	columnTypeMediumInteger
	columnTypeSmallInteger
	columnTypeTinyInteger
	columnTypeDecimal
	columnTypeDouble
	columnTypeFloat
	columnTypeDateTime
	columnTypeDateTimeTz
	columnTypeDate
	columnTypeTime
	columnTypeTimestamp
	columnTypeTimestampTz
	columnTypeYear
	columnTypeBinary
	columnTypeJSON
	columnTypeJSONB
	columnTypeGeography
	columnTypeGeometry
	columnTypePoint
	columnTypeUUID
	columnTypeEnum
	columnTypeCustom // Custom type for user-defined types
)

type indexType int

const (
	indexTypeIndex indexType = iota
	indexTypeUnique
	indexTypePrimary
	indexTypeFulltext
)

// Blueprint represents a schema blueprint for creating or altering a database table.
type Blueprint struct {
	commands        []string // commands to be executed
	name            string
	newName         string
	charset         string
	collation       string
	engine          string
	columns         []*columnDefinition
	indexes         []*indexDefinition
	foreignKeys     []*foreignKeyDefinition
	dropColumns     []string
	renameColumns   map[string]string // old column name to new column name
	dropIndexes     []string          // indexes to be dropped
	dropForeignKeys []string          // foreign keys to be dropped
	dropPrimaryKeys []string          // primary keys to be dropped
	dropUniqueKeys  []string          // unique keys to be dropped
	dropFulltext    []string          // fulltext indexes to be dropped
	renameIndexes   map[string]string // old index name to new index name
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
	return b.addColumn(&columnDefinition{
		name:             name,
		columnType:       columnTypeCustom,
		customColumnType: columnType,
	})
}

// Boolean creates a new boolean column definition in the blueprint.
func (b *Blueprint) Boolean(name string) ColumnDefinition {
	return b.addColumn(&columnDefinition{
		name:       name,
		columnType: columnTypeBoolean,
	})
}

// Char creates a new char column definition in the blueprint.
func (b *Blueprint) Char(name string, length ...int) ColumnDefinition {
	return b.addColumn(&columnDefinition{
		name:       name,
		columnType: columnTypeChar,
		length:     optional(0, length...),
	})
}

// String creates a new string column definition in the blueprint.
func (b *Blueprint) String(name string, length ...int) ColumnDefinition {
	return b.addColumn(&columnDefinition{
		name:       name,
		columnType: columnTypeString,
		length:     optional(0, length...),
	})
}

// LongText creates a new long text column definition in the blueprint.
func (b *Blueprint) LongText(name string) ColumnDefinition {
	return b.addColumn(&columnDefinition{
		name:       name,
		columnType: columnTypeLongText,
	})
}

// Text creates a new text column definition in the blueprint.
func (b *Blueprint) Text(name string) ColumnDefinition {
	return b.addColumn(&columnDefinition{
		name:       name,
		columnType: columnTypeText,
	})
}

// MediumText creates a new medium text column definition in the blueprint.
func (b *Blueprint) MediumText(name string) ColumnDefinition {
	return b.addColumn(&columnDefinition{
		name:       name,
		columnType: columnTypeMediumText,
	})
}

// TinyText creates a new tiny text column definition in the blueprint.
func (b *Blueprint) TinyText(name string) ColumnDefinition {
	return b.addColumn(&columnDefinition{
		name:       name,
		columnType: columnTypeTinyText,
	})
}

// BigIncrements creates a new big increments column definition in the blueprint.
func (b *Blueprint) BigIncrements(name string) ColumnDefinition {
	return b.UnsignedBigInteger(name, true)
}

// BigInteger creates a new big integer column definition in the blueprint.
func (b *Blueprint) BigInteger(name string, autoIncrement ...bool) ColumnDefinition {
	return b.addColumn(&columnDefinition{
		name:          name,
		columnType:    columnTypeBigInteger,
		autoIncrement: optional(false, autoIncrement...),
	})
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
	return b.addColumn(&columnDefinition{
		name:       name,
		columnType: columnTypeDecimal,
		total:      optional(8, params...),
		places:     optionalAtIndex(1, 2, params...),
	})
}

// Double creates a new double column definition in the blueprint.
func (b *Blueprint) Double(name string, params ...int) ColumnDefinition {
	return b.addColumn(&columnDefinition{
		name:       name,
		columnType: columnTypeDouble,
		total:      optional(0, params...),
		places:     optionalAtIndex(1, 0, params...),
	})
}

// Float creates a new float column definition in the blueprint.
//
// The total and places parameters are optional.
//
// Example:
//
//	table.Float("price", 10, 2) // creates a float column with total 10 and places 2
//
//	table.Float("price") // creates a float column with default total 8 and places 2
func (b *Blueprint) Float(name string, params ...int) ColumnDefinition {
	return b.addColumn(&columnDefinition{
		name:       name,
		columnType: columnTypeFloat,
		total:      optional(8, params...),
		places:     optionalAtIndex(1, 2, params...),
	})
}

// ID creates a new big increments column definition in the blueprint with the name "id" or a custom name.
//
// If a name is provided, it will be used as the column name; otherwise, "id" will be used.
func (b *Blueprint) ID(name ...string) ColumnDefinition {
	return b.BigIncrements(optional("id", name...)).Primary()
}

// Increments create a new increment column definition in the blueprint.
func (b *Blueprint) Increments(name string) ColumnDefinition {
	return b.UnsignedInteger(name, true)
}

// Integer creates a new integer column definition in the blueprint.
func (b *Blueprint) Integer(name string, autoIncrement ...bool) ColumnDefinition {
	return b.addColumn(&columnDefinition{
		name:          name,
		columnType:    columnTypeInteger,
		autoIncrement: optional(false, autoIncrement...),
	})
}

// MediumIncrements creates a new medium increments column definition in the blueprint.
func (b *Blueprint) MediumIncrements(name string) ColumnDefinition {
	return b.UnsignedMediumInteger(name, true)
}

func (b *Blueprint) MediumInteger(name string, autoIncrement ...bool) ColumnDefinition {
	return b.addColumn(&columnDefinition{
		name:          name,
		columnType:    columnTypeMediumInteger,
		autoIncrement: optional(false, autoIncrement...),
	})
}

// SmallIncrements creates a new small increments column definition in the blueprint.
func (b *Blueprint) SmallIncrements(name string) ColumnDefinition {
	return b.UnsignedSmallInteger(name, true)
}

// SmallInteger creates a new small integer column definition in the blueprint.
func (b *Blueprint) SmallInteger(name string, autoIncrement ...bool) ColumnDefinition {
	return b.addColumn(&columnDefinition{
		name:          name,
		columnType:    columnTypeSmallInteger,
		autoIncrement: optional(false, autoIncrement...),
	})
}

// TinyIncrements creates a new tiny increments column definition in the blueprint.
func (b *Blueprint) TinyIncrements(name string) ColumnDefinition {
	return b.UnsignedTinyInteger(name, true)
}

// TinyInteger creates a new tiny integer column definition in the blueprint.
func (b *Blueprint) TinyInteger(name string, autoIncrement ...bool) ColumnDefinition {
	return b.addColumn(&columnDefinition{
		name:          name,
		columnType:    columnTypeTinyInteger,
		autoIncrement: optional(false, autoIncrement...),
	})
}

// UnsignedBigInteger creates a new unsigned big integer column definition in the blueprint.
func (b *Blueprint) UnsignedBigInteger(name string, autoIncrement ...bool) ColumnDefinition {
	return b.addColumn(&columnDefinition{
		name:          name,
		columnType:    columnTypeBigInteger,
		autoIncrement: optional(false, autoIncrement...),
		unsigned:      true,
	})
}

// UnsignedInteger creates a new unsigned integer column definition in the blueprint.
func (b *Blueprint) UnsignedInteger(name string, autoIncrement ...bool) ColumnDefinition {
	return b.addColumn(&columnDefinition{
		name:          name,
		columnType:    columnTypeInteger,
		autoIncrement: optional(false, autoIncrement...),
		unsigned:      true,
	})
}

// UnsignedMediumInteger creates a new unsigned medium integer column definition in the blueprint.
func (b *Blueprint) UnsignedMediumInteger(name string, autoIncrement ...bool) ColumnDefinition {
	return b.addColumn(&columnDefinition{
		name:          name,
		columnType:    columnTypeMediumInteger,
		autoIncrement: optional(false, autoIncrement...),
		unsigned:      true,
	})
}

// UnsignedSmallInteger creates a new unsigned small integer column definition in the blueprint.
func (b *Blueprint) UnsignedSmallInteger(name string, autoIncrement ...bool) ColumnDefinition {
	return b.addColumn(&columnDefinition{
		name:          name,
		columnType:    columnTypeSmallInteger,
		autoIncrement: optional(false, autoIncrement...),
		unsigned:      true,
	})
}

// UnsignedTinyInteger creates a new unsigned tiny integer column definition in the blueprint.
func (b *Blueprint) UnsignedTinyInteger(name string, autoIncrement ...bool) ColumnDefinition {
	return b.addColumn(&columnDefinition{
		name:          name,
		columnType:    columnTypeTinyInteger,
		autoIncrement: optional(false, autoIncrement...),
		unsigned:      true,
	})
}

// DateTime creates a new date time column definition in the blueprint.
func (b *Blueprint) DateTime(name string, precision ...int) ColumnDefinition {
	return b.addColumn(&columnDefinition{
		name:       name,
		columnType: columnTypeDateTime,
		precision:  optional(0, precision...),
	})
}

// DateTimeTz creates a new date time with a time zone column definition in the blueprint.
func (b *Blueprint) DateTimeTz(name string, precision ...int) ColumnDefinition {
	return b.addColumn(&columnDefinition{
		name:       name,
		columnType: columnTypeDateTimeTz,
		precision:  optional(0, precision...),
	})
}

// Date creates a new date column definition in the blueprint.
func (b *Blueprint) Date(name string) ColumnDefinition {
	return b.addColumn(&columnDefinition{
		name:       name,
		columnType: columnTypeDate,
	})
}

// Time creates a new time column definition in the blueprint.
func (b *Blueprint) Time(name string, precission ...int) ColumnDefinition {
	return b.addColumn(&columnDefinition{
		name:       name,
		columnType: columnTypeTime,
		precision:  optional(0, precission...),
	})
}

// Timestamp creates a new timestamp column definition in the blueprint.
// The precision parameter is optional and defaults to 0 if not provided.
func (b *Blueprint) Timestamp(name string, precision ...int) ColumnDefinition {
	return b.addColumn(&columnDefinition{
		name:       name,
		columnType: columnTypeTimestamp,
		precision:  optional(0, precision...),
	})
}

// TimestampTz creates a new timestamp with time zone column definition in the blueprint.
// The precision parameter is optional and defaults to 0 if not provided.
func (b *Blueprint) TimestampTz(name string, precision ...int) ColumnDefinition {
	return b.addColumn(&columnDefinition{
		name:       name,
		columnType: columnTypeTimestampTz,
		precision:  optional(0, precision...),
	})
}

// Timestamps adds created_at and updated_at timestamp columns to the blueprint.
func (b *Blueprint) Timestamps() {
	b.Timestamp("created_at").Nullable(false).UseCurrent()
	b.Timestamp("updated_at").Nullable(false).UseCurrent().UseCurrentOnUpdate()
}

// TimestampsTz adds created_at and updated_at timestamp with time zone columns to the blueprint.
func (b *Blueprint) TimestampsTz() {
	b.TimestampTz("created_at").Nullable(false).UseCurrent()
	b.TimestampTz("updated_at").Nullable(false).UseCurrent().UseCurrentOnUpdate()
}

// Year creates a new year column definition in the blueprint.
func (b *Blueprint) Year(name string) ColumnDefinition {
	return b.addColumn(&columnDefinition{
		name:       name,
		columnType: columnTypeYear,
	})
}

// Binary creates a new binary column definition in the blueprint.
func (b *Blueprint) Binary(name string) ColumnDefinition {
	return b.addColumn(&columnDefinition{
		name:       name,
		columnType: columnTypeBinary,
	})
}

// JSON creates a new JSON column definition in the blueprint.
func (b *Blueprint) JSON(name string) ColumnDefinition {
	return b.addColumn(&columnDefinition{
		name:       name,
		columnType: columnTypeJSON,
	})
}

// JSONB creates a new JSONB column definition in the blueprint.
func (b *Blueprint) JSONB(name string) ColumnDefinition {
	return b.addColumn(&columnDefinition{
		name:       name,
		columnType: columnTypeJSONB,
	})
}

// UUID creates a new UUID column definition in the blueprint.
func (b *Blueprint) UUID(name string) ColumnDefinition {
	return b.addColumn(&columnDefinition{
		name:       name,
		columnType: columnTypeUUID,
	})
}

// Geography creates a new geography column definition in the blueprint.
// The subType parameter is optional and can be used to specify the type of geography (e.g., "Point", "LineString", "Polygon").
// The srid parameter is optional and specifies the Spatial Reference Identifier (SRID) for the geography type.
func (b *Blueprint) Geography(name string, subType string, srid ...int) ColumnDefinition {
	return b.addColumn(&columnDefinition{
		name:       name,
		columnType: columnTypeGeography,
		subType:    subType,
		srid:       optional(4326, srid...),
	})
}

// Geometry creates a new geometry column definition in the blueprint.
// The subType parameter is optional and can be used to specify the type of geometry (e.g., "Point", "LineString", "Polygon").
// The srid parameter is optional and specifies the Spatial Reference Identifier (SRID) for the geometry type.
func (b *Blueprint) Geometry(name string, subType string, srid ...int) ColumnDefinition {
	return b.addColumn(&columnDefinition{
		name:       name,
		columnType: columnTypeGeometry,
		subType:    subType,
		srid:       optional(0, srid...),
	})
}

// Point creates a new point column definition in the blueprint.
func (b *Blueprint) Point(name string, srid ...int) ColumnDefinition {
	return b.addColumn(&columnDefinition{
		name:       name,
		columnType: columnTypePoint,
		srid:       optional(4326, srid...),
	})
}

// Enum creates a new enum column definition in the blueprint.
// The allowedEnums parameter is a slice of strings that defines the allowed values for the enum column.
//
// Example:
//
//	table.Enum("status", []string{"active", "inactive", "pending"})
//	table.Enum("role", []string{"admin", "user", "guest"}).Comment("User role in the system")
func (b *Blueprint) Enum(name string, allowedEnums []string) ColumnDefinition {
	return b.addColumn(&columnDefinition{
		name:         name,
		columnType:   columnTypeEnum,
		allowedEnums: allowedEnums,
	})
}

// Index creates a new index definition in the blueprint.
//
// Example:
//
//	table.Index("email")
//	table.Index("email", "username") // creates a composite index
//	table.Index("email").Algorithm("btree") // creates a btree index
func (b *Blueprint) Index(column string, otherColumns ...string) IndexDefinition {
	index := &indexDefinition{
		indexType: indexTypeIndex,
		columns:   append([]string{column}, otherColumns...),
	}
	b.indexes = append(b.indexes, index)
	b.addCommand("index")

	return index
}

// Unique creates a new unique index definition in the blueprint.
//
// Example:
//
//	table.Unique("email")
//	table.Unique("email", "username") // creates a composite unique index
func (b *Blueprint) Unique(column string, otherColumns ...string) IndexDefinition {
	index := &indexDefinition{
		indexType: indexTypeUnique,
		columns:   append([]string{column}, otherColumns...),
	}
	b.indexes = append(b.indexes, index)
	b.addCommand("unique")

	return index
}

// Primary creates a new primary key index definition in the blueprint.
//
// Example:
//
//	table.Primary("id")
//	table.Primary("id", "email") // creates a composite primary key
func (b *Blueprint) Primary(column string, otherColumns ...string) IndexDefinition {
	index := &indexDefinition{
		indexType: indexTypePrimary,
		columns:   append([]string{column}, otherColumns...),
	}
	b.indexes = append(b.indexes, index)
	b.addCommand("primary")
	return index
}

// Fulltext creates a new fulltext index definition in the blueprint.
func (b *Blueprint) Fulltext(column string, otherColumns ...string) IndexDefinition {
	index := &indexDefinition{
		indexType: indexTypeFulltext,
		columns:   append([]string{column}, otherColumns...),
	}
	b.indexes = append(b.indexes, index)
	b.addCommand("fulltext")

	return index
}

// Foreign creates a new foreign key definition in the blueprint.
//
// Example:
//
//	table.Foreign("user_id").References("id").On("users").OnDelete("CASCADE").OnUpdate("CASCADE")
func (b *Blueprint) Foreign(column string) ForeignKeyDefinition {
	fk := &foreignKeyDefinition{
		tableName: b.name,
		column:    column,
	}
	b.foreignKeys = append(b.foreignKeys, fk)
	b.addCommand("foreign")
	return fk
}

// DropColumn adds a column to be dropped from the table.
//
// Example:
//
//	table.DropColumn("old_column")
//	table.DropColumn("old_column", "another_old_column") // drops multiple columns
func (b *Blueprint) DropColumn(column string, otherColumns ...string) {
	b.dropColumns = append(b.dropColumns, append([]string{column}, otherColumns...)...)
	b.addCommand("dropColumn")
}

// RenameColumn changes the name of the table in the blueprint.
//
// Example:
//
//	table.RenameColumn("old_table_name", "new_table_name")
func (b *Blueprint) RenameColumn(oldColumn string, newColumn string) {
	if b.renameColumns == nil {
		b.renameColumns = make(map[string]string)
	}
	b.renameColumns[oldColumn] = newColumn
	b.addCommand("renameColumn")
}

// DropIndex adds an index to be dropped from the table.
func (b *Blueprint) DropIndex(indexName string) {
	b.dropIndexes = append(b.dropIndexes, indexName)
	b.addCommand("dropIndex")
}

// DropForeign adds a foreign key to be dropped from the table.
func (b *Blueprint) DropForeign(foreignKeyName string) {
	b.dropForeignKeys = append(b.dropForeignKeys, foreignKeyName)
	b.addCommand("dropForeign")
}

// DropPrimary adds a primary key to be dropped from the table.
func (b *Blueprint) DropPrimary(primaryKeyName string) {
	b.dropPrimaryKeys = append(b.dropPrimaryKeys, primaryKeyName)
	b.addCommand("dropPrimary")
}

// DropUnique adds a unique key to be dropped from the table.
func (b *Blueprint) DropUnique(uniqueKeyName string) {
	b.dropUniqueKeys = append(b.dropUniqueKeys, uniqueKeyName)
	b.addCommand("dropUnique")
}

func (b *Blueprint) DropFulltext(fulltextIndexName string) {
	b.dropFulltext = append(b.dropFulltext, fulltextIndexName)
	b.addCommand("dropFulltext")
}

// RenameIndex changes the name of an index in the blueprint.
// Example:
//
//	table.RenameIndex("old_index_name", "new_index_name")
func (b *Blueprint) RenameIndex(oldIndexName string, newIndexName string) {
	if b.renameIndexes == nil {
		b.renameIndexes = make(map[string]string)
	}
	b.renameIndexes[oldIndexName] = newIndexName
	b.addCommand("renameIndex")
}

func (b *Blueprint) getAddedColumns() []*columnDefinition {
	var addedColumns []*columnDefinition
	for _, col := range b.columns {
		if !col.changed {
			addedColumns = append(addedColumns, col)
		}
	}
	return addedColumns
}

func (b *Blueprint) getChangedColumns() []*columnDefinition {
	var changedColumns []*columnDefinition
	for _, col := range b.columns {
		if col.changed {
			changedColumns = append(changedColumns, col)
		}
	}
	return changedColumns
}

func (b *Blueprint) create() {
	b.addCommand("create")
}

func (b *Blueprint) creating() bool {
	for _, command := range b.commands {
		if command == "create" || command == "createIfNotExists" {
			return true
		}
	}
	return false
}

func (b *Blueprint) createIfNotExists() {
	b.addCommand("createIfNotExists")
}

func (b *Blueprint) drop() {
	b.addCommand("drop")
}

func (b *Blueprint) dropIfExists() {
	b.addCommand("dropIfNotExists")
}

func (b *Blueprint) rename() {
	b.addCommand("rename")
}

func (b *Blueprint) addImpliedCommands() {
	if len(b.getAddedColumns()) > 0 && !b.creating() {
		b.commands = append([]string{"add"}, b.commands...)
	}
	if len(b.getChangedColumns()) > 0 && !b.creating() {
		b.commands = append([]string{"change"}, b.commands...)
	}
}

func (b *Blueprint) toSql(grammar grammar) ([]string, error) {
	b.addImpliedCommands()

	var statements []string

	commandMap := map[string]func(*Blueprint) (string, error){
		"create":            grammar.compileCreate,
		"createIfNotExists": grammar.compileCreateIfNotExists,
		"add":               grammar.compileAdd,
		"drop":              grammar.compileDrop,
		"dropIfNotExists":   grammar.compileDropIfExists,
		"rename":            grammar.compileRename,
	}
	for _, command := range b.commands {
		if compileFunc, exists := commandMap[command]; exists {
			sql, err := compileFunc(b)
			if err != nil {
				return nil, err
			}
			if sql != "" {
				statements = append(statements, sql)
			}
		}
		switch command {
		case "create", "createIfNotExists", "add":
			for _, col := range b.getAddedColumns() {
				if col.index {
					indexDef := &indexDefinition{
						indexType: indexTypeIndex,
						columns:   []string{col.name},
						name:      col.indexName,
					}
					sql, err := grammar.compileIndex(b, indexDef)
					if err != nil {
						return nil, err
					}
					if sql != "" {
						statements = append(statements, sql)
					}
				}
			}
		case "change":
			changedStatements, err := grammar.compileChange(b)
			if err != nil {
				return nil, err
			}
			statements = append(statements, changedStatements...)
		case "index":
			indexStatements, err := b.getIndexStatements(grammar, indexTypeIndex)
			if err != nil {
				return nil, err
			}
			statements = append(statements, indexStatements...)
		case "unique":
			uniqueStatements, err := b.getIndexStatements(grammar, indexTypeUnique)
			if err != nil {
				return nil, err
			}
			statements = append(statements, uniqueStatements...)
		case "primary":
			primaryStatements, err := b.getIndexStatements(grammar, indexTypePrimary)
			if err != nil {
				return nil, err
			}
			statements = append(statements, primaryStatements...)
		case "fulltext":
			fulltextStatements, err := b.getIndexStatements(grammar, indexTypeFulltext)
			if err != nil {
				return nil, err
			}
			statements = append(statements, fulltextStatements...)
		case "foreign":
			for _, foreignKey := range b.foreignKeys {
				sql, err := grammar.compileForeign(b, foreignKey)
				if err != nil {
					return nil, err
				}
				if sql != "" {
					statements = append(statements, sql)
				}
			}
		case "dropColumn":
			sql, err := grammar.compileDropColumn(b)
			if err != nil {
				return nil, err
			}
			if sql != "" {
				statements = append(statements, sql)
			}
		case "renameColumn":
			for oldName, newName := range b.renameColumns {
				sql, err := grammar.compileRenameColumn(b, oldName, newName)
				if err != nil {
					return nil, err
				}
				if sql != "" {
					statements = append(statements, sql)
				}
			}
		case "dropIndex":
			for _, indexName := range b.dropIndexes {
				sql, err := grammar.compileDropIndex(b, indexName)
				if err != nil {
					return nil, err
				}
				if sql != "" {
					statements = append(statements, sql)
				}
			}
		case "dropUnique":
			for _, uniqueKeyName := range b.dropUniqueKeys {
				sql, err := grammar.compileDropUnique(b, uniqueKeyName)
				if err != nil {
					return nil, err
				}
				if sql != "" {
					statements = append(statements, sql)
				}
			}
		case "dropFulltext":
			for _, fulltextIndexName := range b.dropFulltext {
				sql, err := grammar.compileDropFulltext(b, fulltextIndexName)
				if err != nil {
					return nil, err
				}
				if sql != "" {
					statements = append(statements, sql)
				}
			}
		case "dropPrimary":
			for _, primaryKeyName := range b.dropPrimaryKeys {
				sql, err := grammar.compileDropPrimary(b, primaryKeyName)
				if err != nil {
					return nil, err
				}
				if sql != "" {
					statements = append(statements, sql)
				}
			}
		case "renameIndex":
			for oldName, newName := range b.renameIndexes {
				sql, err := grammar.compileRenameIndex(b, oldName, newName)
				if err != nil {
					return nil, err
				}
				if sql != "" {
					statements = append(statements, sql)
				}
			}
		case "dropForeign":
			for _, foreignKeyName := range b.dropForeignKeys {
				sql, err := grammar.compileDropForeign(b, foreignKeyName)
				if err != nil {
					return nil, err
				}
				if sql != "" {
					statements = append(statements, sql)
				}
			}
		}
	}

	return statements, nil
}

func (b *Blueprint) getIndexStatements(grammar grammar, idxType indexType) ([]string, error) {
	indexCommandMap := map[indexType]func(*Blueprint, *indexDefinition) (string, error){
		indexTypeIndex:    grammar.compileIndex,
		indexTypeUnique:   grammar.compileUnique,
		indexTypePrimary:  grammar.compilePrimary,
		indexTypeFulltext: grammar.compileFullText,
	}
	var statements []string
	for _, index := range b.indexes {
		if index.indexType == idxType {
			compileFunc, exists := indexCommandMap[idxType]
			if !exists {
				return nil, fmt.Errorf("unsupported index type: %d", idxType)
			}
			sql, err := compileFunc(b, index)
			if err != nil {
				return nil, err
			}
			if sql != "" {
				statements = append(statements, sql)
			}
		}
	}
	return statements, nil
}

func (b *Blueprint) addColumn(col *columnDefinition) *columnDefinition {
	b.columns = append(b.columns, col)
	return col
}

func (b *Blueprint) addCommand(command string) {
	if command == "" {
		return
	}
	if !slices.Contains(b.commands, command) {
		b.commands = append(b.commands, command)
	}
}
