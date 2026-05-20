package dialect

import "github.com/pressly/goose/v3/database"

// Dialect is the type of database dialect.
type Dialect string

const (
	MySQL    Dialect = "mysql"
	MariaDB  Dialect = "mariadb"
	Postgres Dialect = "postgres"
	SQLite3  Dialect = "sqlite3"
	Unknown  Dialect = ""
)

func (d Dialect) String() string {
	return string(d)
}

func (d Dialect) GooseDialect() database.Dialect {
	switch d {
	case MySQL, MariaDB:
		return database.DialectMySQL
	case Postgres:
		return database.DialectPostgres
	case SQLite3:
		return database.DialectSQLite3
	case Unknown:
		return database.DialectCustom
	default:
		return database.DialectCustom
	}
}

func FromString(dialect string) Dialect {
	switch dialect {
	case "mysql":
		return MySQL
	case "mariadb":
		return MariaDB
	case "postgres", "pgx":
		return Postgres
	case "sqlite3", "sqlite":
		return SQLite3
	default:
		return Unknown
	}
}

// DriverName returns the Go sql driver name to use with sql.Open.
// originalValue is the raw string passed by the caller; it is preserved for
// postgres so that "pgx" and "postgres" (lib/pq) remain distinct.
func DriverName(d Dialect, originalValue string) string {
	switch d {
	case MySQL, MariaDB:
		return "mysql"
	case SQLite3:
		return "sqlite3"
	case Postgres:
		return originalValue // preserve "pgx" vs "postgres"
	case Unknown:
		return originalValue
	}
	return originalValue
}
