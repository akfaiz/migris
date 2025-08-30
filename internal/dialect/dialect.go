package dialect

import "github.com/pressly/goose/v3/database"

// Dialect is the type of database dialect.
type Dialect string

const (
	MySQL    Dialect = "mysql"
	Postgres Dialect = "postgres"
	Unknown  Dialect = ""
)

func (d Dialect) String() string {
	return string(d)
}

func (d Dialect) GooseDialect() database.Dialect {
	switch d {
	case MySQL:
		return database.DialectMySQL
	case Postgres:
		return database.DialectPostgres
	default:
		return database.DialectCustom
	}
}

func FromString(dialect string) Dialect {
	switch dialect {
	case "mysql", "mariadb":
		return MySQL
	case "postgres", "pgx":
		return Postgres
	default:
		return Unknown
	}
}
