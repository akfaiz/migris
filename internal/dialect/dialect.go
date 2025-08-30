package dialect

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
