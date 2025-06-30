package schema

import "slices"

type columnType uint8

const (
	columnTypeBoolean columnType = iota
	columnTypeChar
	columnTypeString
	columnTypeText
	columnTypeIncrements
	columnTypeBigIncrements
	columnTypeSmallIncrements
	columnTypeInteger
	columnTypeBigInteger
	columnTypeSmallInteger
	columnTypeDecimal
	columnTypeDouble
	columnTypeFloat
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
	columnTypeUUID
	columnTypeEnum
)

type indexType int

const (
	indexTypeIndex indexType = iota
	indexTypeUnique
	indexTypePrimary
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

// Boolean creates a new boolean column definition in the blueprint.
func (b *Blueprint) Boolean(name string) ColumnDefinition {
	col := &columnDefinition{
		name:       name,
		columnType: columnTypeBoolean,
	}
	b.columns = append(b.columns, col)
	return col
}

// Char creates a new char column definition in the blueprint.
func (b *Blueprint) Char(name string, length ...int) ColumnDefinition {
	col := &columnDefinition{
		name:       name,
		columnType: columnTypeChar,
		length:     optionalInt(0, length...),
	}
	b.columns = append(b.columns, col)
	return col
}

// String creates a new string column definition in the blueprint.
func (b *Blueprint) String(name string, length ...int) ColumnDefinition {
	col := &columnDefinition{
		name:       name,
		columnType: columnTypeString,
		length:     optionalInt(0, length...),
	}
	b.columns = append(b.columns, col)
	return col
}

// Text creates a new text column definition in the blueprint.
func (b *Blueprint) Text(name string) ColumnDefinition {
	col := &columnDefinition{
		name:       name,
		columnType: columnTypeText,
	}
	b.columns = append(b.columns, col)
	return col
}

// BigIncrements creates a new big increments column definition in the blueprint.
func (b *Blueprint) BigIncrements(name string) ColumnDefinition {
	col := &columnDefinition{
		name:       name,
		columnType: columnTypeBigIncrements,
	}
	b.columns = append(b.columns, col)
	return col
}

// BigInteger creates a new big integer column definition in the blueprint.
func (b *Blueprint) BigInteger(name string) ColumnDefinition {
	col := &columnDefinition{
		name:       name,
		columnType: columnTypeBigInteger,
	}
	b.columns = append(b.columns, col)
	return col
}

// Decimal creates a new decimal column definition in the blueprint.
func (b *Blueprint) Decimal(name string, precision, scale int) ColumnDefinition {
	col := &columnDefinition{
		name:       name,
		columnType: columnTypeDecimal,
		precision:  precision,
		scale:      scale,
	}
	b.columns = append(b.columns, col)
	return col
}

// Double creates a new double column definition in the blueprint.
func (b *Blueprint) Double(name string) ColumnDefinition {
	col := &columnDefinition{
		name:       name,
		columnType: columnTypeDouble,
	}
	b.columns = append(b.columns, col)
	return col
}

// Float creates a new float column definition in the blueprint.
func (b *Blueprint) Float(name string) ColumnDefinition {
	col := &columnDefinition{
		name:       name,
		columnType: columnTypeFloat,
	}
	b.columns = append(b.columns, col)
	return col
}

// ID creates a new big increments column definition with the name "id" in the blueprint.
func (b *Blueprint) ID() ColumnDefinition {
	return b.BigIncrements("id").Primary()
}

// Increments creates a new increments column definition in the blueprint.
func (b *Blueprint) Increments(name string) ColumnDefinition {
	col := &columnDefinition{
		name:       name,
		columnType: columnTypeIncrements,
	}
	b.columns = append(b.columns, col)
	return col
}

// Integer creates a new integer column definition in the blueprint.
func (b *Blueprint) Integer(name string) ColumnDefinition {
	col := &columnDefinition{
		name:       name,
		columnType: columnTypeInteger,
	}
	b.columns = append(b.columns, col)
	return col
}

// SmallIncrements creates a new small increments column definition in the blueprint.
func (b *Blueprint) SmallIncrements(name string) ColumnDefinition {
	col := &columnDefinition{
		name:       name,
		columnType: columnTypeSmallIncrements,
	}
	b.columns = append(b.columns, col)
	return col
}

// SmallInteger creates a new small integer column definition in the blueprint.
func (b *Blueprint) SmallInteger(name string) ColumnDefinition {
	col := &columnDefinition{
		name:       name,
		columnType: columnTypeSmallInteger,
	}
	b.columns = append(b.columns, col)
	return col
}

// Date creates a new date column definition in the blueprint.
func (b *Blueprint) Date(name string) ColumnDefinition {
	col := &columnDefinition{
		name:       name,
		columnType: columnTypeDate,
	}
	b.columns = append(b.columns, col)
	return col
}

// Time creates a new time column definition in the blueprint.
func (b *Blueprint) Time(name string) ColumnDefinition {
	col := &columnDefinition{
		name:       name,
		columnType: columnTypeTime,
	}
	b.columns = append(b.columns, col)
	return col
}

// Timestamp creates a new timestamp column definition in the blueprint.
// The precision parameter is optional and defaults to 0 if not provided.
func (b *Blueprint) Timestamp(name string, precission ...int) ColumnDefinition {
	col := &columnDefinition{
		name:       name,
		columnType: columnTypeTimestamp,
		precision:  optionalInt(0, precission...),
	}
	b.columns = append(b.columns, col)
	return col
}

// TimestampTz creates a new timestamp with time zone column definition in the blueprint.
// The precision parameter is optional and defaults to 0 if not provided.
func (b *Blueprint) TimestampTz(name string, precission ...int) ColumnDefinition {
	col := &columnDefinition{
		name:       name,
		columnType: columnTypeTimestampTz,
		precision:  optionalInt(0, precission...),
	}
	b.columns = append(b.columns, col)
	return col
}

// Year creates a new year column definition in the blueprint.
func (b *Blueprint) Year(name string) ColumnDefinition {
	col := &columnDefinition{
		name:       name,
		columnType: columnTypeYear,
	}
	b.columns = append(b.columns, col)
	return col
}

// Binary creates a new binary column definition in the blueprint.
func (b *Blueprint) Binary(name string) ColumnDefinition {
	col := &columnDefinition{
		name:       name,
		columnType: columnTypeBinary,
	}
	b.columns = append(b.columns, col)
	return col
}

// JSON creates a new JSON column definition in the blueprint.
func (b *Blueprint) JSON(name string) ColumnDefinition {
	col := &columnDefinition{
		name:       name,
		columnType: columnTypeJSON,
	}
	b.columns = append(b.columns, col)
	return col
}

// JSONB creates a new JSONB column definition in the blueprint.
func (b *Blueprint) JSONB(name string) ColumnDefinition {
	col := &columnDefinition{
		name:       name,
		columnType: columnTypeJSONB,
	}
	b.columns = append(b.columns, col)
	return col
}

// UUID creates a new UUID column definition in the blueprint.
func (b *Blueprint) UUID(name string) ColumnDefinition {
	col := &columnDefinition{
		name:       name,
		columnType: columnTypeUUID,
	}
	b.columns = append(b.columns, col)
	return col
}

// Geography creates a new geography column definition in the blueprint.
// The subType parameter is optional and can be used to specify the type of geography (e.g., "Point", "LineString", "Polygon").
// The srid parameter is optional and specifies the Spatial Reference Identifier (SRID) for the geography type.
func (b *Blueprint) Geography(name string, subType string, srid ...int) ColumnDefinition {
	col := &columnDefinition{
		name:       name,
		columnType: columnTypeGeography,
		subType:    subType,
		srid:       optionalInt(4326, srid...),
	}
	b.columns = append(b.columns, col)
	return col
}

// Geometry creates a new geometry column definition in the blueprint.
// The subType parameter is optional and can be used to specify the type of geometry (e.g., "Point", "LineString", "Polygon").
// The srid parameter is optional and specifies the Spatial Reference Identifier (SRID) for the geometry type.
func (b *Blueprint) Geometry(name string, subType string, srid int) ColumnDefinition {
	col := &columnDefinition{
		name:       name,
		columnType: columnTypeGeometry,
		subType:    subType,
		srid:       srid,
	}
	b.columns = append(b.columns, col)
	return col
}

// Enum creates a new enum column definition in the blueprint.
// The allowedEnums parameter is a slice of strings that defines the allowed values for the enum column.
//
// Example:
//
//	table.Enum("status", []string{"active", "inactive", "pending"})
//	table.Enum("role", []string{"admin", "user", "guest"}).Comment("User role in the system")
func (b *Blueprint) Enum(name string, allowedEnums []string) ColumnDefinition {
	col := &columnDefinition{
		name:         name,
		columnType:   columnTypeEnum,
		allowedEnums: allowedEnums,
	}
	b.columns = append(b.columns, col)
	return col
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

	return index
}

// Primary creates a new primary key index definition in the blueprint.
//
// Example:
//
//	table.Primary("id")
//	table.Primary("id", "email") // creates a composite primary key
func (b *Blueprint) Primary(column string, otherColumns ...string) {
	index := &indexDefinition{
		indexType: indexTypePrimary,
		columns:   append([]string{column}, otherColumns...),
	}
	b.indexes = append(b.indexes, index)
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

func (b *Blueprint) getAddeddColumns() []*columnDefinition {
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

func (b *Blueprint) addCommand(command string) {
	if command == "" {
		return
	}
	if !slices.Contains(b.commands, command) {
		b.commands = append(b.commands, command)
	}
}

func (b *Blueprint) addImpliedCommands() {
	if len(b.getAddeddColumns()) > 0 && !b.creating() {
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
			for _, col := range b.getAddeddColumns() {
				if col.unique && col.uniqueIndexName != "" {
					sql, err := grammar.compileIndexSql(b, &indexDefinition{
						name:      col.uniqueIndexName,
						indexType: indexTypeUnique,
						columns:   []string{col.name},
					})
					if err != nil {
						return nil, err
					}
					statements = append(statements, sql)
				}
				if col.index {
					sql, err := grammar.compileIndexSql(b, &indexDefinition{
						name:      col.indexName,
						indexType: indexTypeIndex,
						columns:   []string{col.name},
					})
					if err != nil {
						return nil, err
					}
					statements = append(statements, sql)
				}
			}

			for _, index := range b.indexes {
				sql, err := grammar.compileIndexSql(b, index)
				if err != nil {
					return nil, err
				}
				statements = append(statements, sql)
			}

			for _, foreignKey := range b.foreignKeys {
				sql, err := grammar.compileForeignKeySql(b, foreignKey)
				if err != nil {
					return nil, err
				}
				if sql != "" {
					statements = append(statements, sql)
				}
			}
		case "change":
			sqls, err := grammar.compileChange(b)
			if err != nil {
				return nil, err
			}
			statements = append(statements, sqls...)
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
				sql, err := grammar.compileDropIndex(indexName)
				if err != nil {
					return nil, err
				}
				if sql != "" {
					statements = append(statements, sql)
				}
			}
		case "dropUnique":
			for _, uniqueKeyName := range b.dropUniqueKeys {
				sql, err := grammar.compileDropUnique(uniqueKeyName)
				if err != nil {
					return nil, err
				}
				if sql != "" {
					statements = append(statements, sql)
				}
			}
		case "dropPrimary":
			for _, primaryKeyName := range b.dropPrimaryKeys {
				sql, err := grammar.compileDropPrimaryKey(b, primaryKeyName)
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
				sql, err := grammar.compileDropForeignKey(b, foreignKeyName)
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

type columnDefinition struct {
	name            string
	columnType      columnType
	comment         string
	defaultVal      any
	nullable        bool
	primary         bool
	index           bool
	indexName       string
	unique          bool
	uniqueIndexName string
	length          int
	precision       int
	scale           int
	changed         bool
	allowedEnums    []string // for enum type columns
	subType         string   // for geography and geometry types
	srid            int      // for geography and geometry types

}

func (c *columnDefinition) Comment(comment string) ColumnDefinition {
	c.comment = comment
	return c
}

func (c *columnDefinition) Default(value any) ColumnDefinition {
	c.defaultVal = value
	return c
}

func (c *columnDefinition) Index(indexName ...string) ColumnDefinition {
	c.index = true
	c.indexName = optionalString("", indexName...)
	return c
}

func (c *columnDefinition) Nullable(value ...bool) ColumnDefinition {
	c.nullable = optionalBool(true, value...)
	return c
}

func (c *columnDefinition) Primary() ColumnDefinition {
	c.primary = true
	return c
}

func (c *columnDefinition) Unique(indexName ...string) ColumnDefinition {
	c.unique = true
	c.uniqueIndexName = optionalString("", indexName...)
	return c
}

func (c *columnDefinition) Change() ColumnDefinition {
	c.changed = true
	return c
}

type indexDefinition struct {
	name       string
	indexType  indexType
	algorithmn string
	columns    []string
}

func (id *indexDefinition) Algorithm(algorithm string) IndexDefinition {
	id.algorithmn = algorithm
	return id
}

func (id *indexDefinition) Name(name string) IndexDefinition {
	id.name = name
	return id
}

type foreignKeyDefinition struct {
	tableName  string
	column     string
	references string
	on         string
	onDelete   string
	onUpdate   string
}

func (fk *foreignKeyDefinition) References(column string) ForeignKeyDefinition {
	fk.references = column
	return fk
}

func (fk *foreignKeyDefinition) On(table string) ForeignKeyDefinition {
	fk.on = table
	return fk
}

func (fk *foreignKeyDefinition) OnDelete(action string) ForeignKeyDefinition {
	fk.onDelete = action
	return fk
}

func (fk *foreignKeyDefinition) OnUpdate(action string) ForeignKeyDefinition {
	fk.onUpdate = action
	return fk
}
