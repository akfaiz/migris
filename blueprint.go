package schema

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
	columnTypeJSON
	columnTypeJSONB
	columnTypeUUID
)

type indexType int

const (
	indexTypeIndex indexType = iota
	indexTypeUnique
	indexTypePrimary
)

// Blueprint represents a schema blueprint for creating or altering a database table.
type Blueprint struct {
	name            string
	newName         string
	columns         []*columnDefinition
	indexes         []*indexDefinition
	foreignKeys     []*foreignKeyDefinition
	dropColumns     []string
	renameColumns   map[string]string // old column name to new column name
	dropIndexes     []string          // indexes to be dropped
	dropForeignKeys []string          // foreign keys to be dropped
	dropPrimaryKeys []string          // primary keys to be dropped
	dropUniqueKeys  []string          // unique keys to be dropped
}

// Boolean creates a new boolean column definition in the blueprint.
func (b *Blueprint) Boolean(name string) ColumnDefinition {
	col := &columnDefinition{
		tableName:  b.name,
		name:       name,
		columnType: columnTypeBoolean,
	}
	b.columns = append(b.columns, col)
	return col
}

// Char creates a new char column definition in the blueprint.
// The length parameter is optional and defaults to 1 if not provided.
func (b *Blueprint) Char(name string, length ...int) ColumnDefinition {
	col := &columnDefinition{
		tableName:  b.name,
		name:       name,
		columnType: columnTypeChar,
		length:     optionalInt(1, length...),
	}
	b.columns = append(b.columns, col)
	return col
}

// String creates a new string column definition in the blueprint.
// The length parameter is optional and defaults to 255 if not provided.
func (b *Blueprint) String(name string, length ...int) ColumnDefinition {
	col := &columnDefinition{
		tableName:  b.name,
		name:       name,
		columnType: columnTypeString,
		length:     optionalInt(255, length...),
	}
	b.columns = append(b.columns, col)
	return col
}

// Text creates a new text column definition in the blueprint.
func (b *Blueprint) Text(name string) ColumnDefinition {
	col := &columnDefinition{
		tableName:  b.name,
		name:       name,
		columnType: columnTypeText,
	}
	b.columns = append(b.columns, col)
	return col
}

// BigIncrements creates a new big increments column definition in the blueprint.
func (b *Blueprint) BigIncrements(name string) ColumnDefinition {
	col := &columnDefinition{
		tableName:  b.name,
		name:       name,
		columnType: columnTypeBigIncrements,
	}
	b.columns = append(b.columns, col)
	return col
}

// BigInteger creates a new big integer column definition in the blueprint.
func (b *Blueprint) BigInteger(name string) ColumnDefinition {
	col := &columnDefinition{
		tableName:  b.name,
		name:       name,
		columnType: columnTypeBigInteger,
	}
	b.columns = append(b.columns, col)
	return col
}

// Decimal creates a new decimal column definition in the blueprint.
func (b *Blueprint) Decimal(name string, precision, scale int) ColumnDefinition {
	col := &columnDefinition{
		tableName:  b.name,
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
		tableName:  b.name,
		name:       name,
		columnType: columnTypeDouble,
	}
	b.columns = append(b.columns, col)
	return col
}

// Float creates a new float column definition in the blueprint.
func (b *Blueprint) Float(name string) ColumnDefinition {
	col := &columnDefinition{
		tableName:  b.name,
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
		tableName:  b.name,
		name:       name,
		columnType: columnTypeIncrements,
	}
	b.columns = append(b.columns, col)
	return col
}

// Integer creates a new integer column definition in the blueprint.
func (b *Blueprint) Integer(name string) ColumnDefinition {
	col := &columnDefinition{
		tableName:  b.name,
		name:       name,
		columnType: columnTypeInteger,
	}
	b.columns = append(b.columns, col)
	return col
}

// SmallIncrements creates a new small increments column definition in the blueprint.
func (b *Blueprint) SmallIncrements(name string) ColumnDefinition {
	col := &columnDefinition{
		tableName:  b.name,
		name:       name,
		columnType: columnTypeSmallIncrements,
	}
	b.columns = append(b.columns, col)
	return col
}

// SmallInteger creates a new small integer column definition in the blueprint.
func (b *Blueprint) SmallInteger(name string) ColumnDefinition {
	col := &columnDefinition{
		tableName:  b.name,
		name:       name,
		columnType: columnTypeSmallInteger,
	}
	b.columns = append(b.columns, col)
	return col
}

// Date creates a new date column definition in the blueprint.
func (b *Blueprint) Date(name string) ColumnDefinition {
	col := &columnDefinition{
		tableName:  b.name,
		name:       name,
		columnType: columnTypeDate,
	}
	b.columns = append(b.columns, col)
	return col
}

// Time creates a new time column definition in the blueprint.
func (b *Blueprint) Time(name string) ColumnDefinition {
	col := &columnDefinition{
		tableName:  b.name,
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
		tableName:  b.name,
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
		tableName:  b.name,
		name:       name,
		columnType: columnTypeTimestamp,
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

// Index creates a new index definition in the blueprint.
func (b *Blueprint) Index(column string, otherColumns ...string) IndexDefinition {
	index := &indexDefinition{
		tableName: b.name,
		indexType: indexTypeIndex,
		columns:   append([]string{column}, otherColumns...),
	}
	b.indexes = append(b.indexes, index)

	return index
}

// Unique creates a new unique index definition in the blueprint.
func (b *Blueprint) Unique(column string, otherColumns ...string) {
	index := &indexDefinition{
		indexType: indexTypeUnique,
		tableName: b.name,
		columns:   append([]string{column}, otherColumns...),
	}
	b.indexes = append(b.indexes, index)
}

// Primary creates a new primary key index definition in the blueprint.
func (b *Blueprint) Primary(column string, otherColumns ...string) {
	index := &indexDefinition{
		indexType: indexTypePrimary,
		tableName: b.name,
		columns:   append([]string{column}, otherColumns...),
	}
	b.indexes = append(b.indexes, index)
}

// Foreign creates a new foreign key definition in the blueprint.
func (b *Blueprint) Foreign(column string) ForeignKeyDefinition {
	fk := &foreignKeyDefinition{
		tableName: b.name,
		column:    column,
	}
	b.foreignKeys = append(b.foreignKeys, fk)
	return fk
}

// DropColumn adds a column to be dropped from the table.
func (b *Blueprint) DropColumn(column string, otherColumns ...string) {
	b.dropColumns = append(b.dropColumns, append([]string{column}, otherColumns...)...)
}

// Rename changes the name of the table in the blueprint.
func (b *Blueprint) RenameColumn(oldColumn string, newColumn string) {
	if b.renameColumns == nil {
		b.renameColumns = make(map[string]string)
	}
	b.renameColumns[oldColumn] = newColumn
}

// DropIndex adds an index to be dropped from the table.
func (b *Blueprint) DropIndex(indexName string) {
	b.dropIndexes = append(b.dropIndexes, indexName)
}

// DropForeign adds a foreign key to be dropped from the table.
func (b *Blueprint) DropForeign(foreignKeyName string) {
	b.dropForeignKeys = append(b.dropForeignKeys, foreignKeyName)
}

// DropPrimary adds a primary key to be dropped from the table.
func (b *Blueprint) DropPrimary(primaryKeyName string) {
	b.dropPrimaryKeys = append(b.dropPrimaryKeys, primaryKeyName)
}

// DropUnique adds a unique key to be dropped from the table.
func (b *Blueprint) DropUnique(uniqueKeyName string) {
	b.dropUniqueKeys = append(b.dropUniqueKeys, uniqueKeyName)
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

type columnDefinition struct {
	tableName       string
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

type indexDefinition struct {
	tableName  string
	indexType  indexType
	algorithmn string
	columns    []string
}

func (id *indexDefinition) Algorithm(algorithm string) IndexDefinition {
	id.algorithmn = algorithm
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
