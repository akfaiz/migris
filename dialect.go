package schema

import (
	"fmt"
)

type dialect uint8

const (
	dialectUnknown dialect = iota
	dialectMySQL
	dialectPostgres
)

func (d dialect) String() string {
	switch d {
	case dialectMySQL:
		return "mysql"
	case dialectPostgres:
		return "postgres"
	default:
		return ""
	}
}

var dialectValue = dialectUnknown
var cfg = &config{
	debug: false, // default value
}

// Init initializes the schema package with the given dialect and options.
func Init(dialect string, options ...Option) error {
	dialectValue = dialectFromString(dialect)
	if dialectValue == dialectUnknown {
		return fmt.Errorf("unknown dialect: %s", dialect)
	}
	cfg = applyOptions(options...)

	return nil
}

func dialectFromString(d string) dialect {
	switch d {
	case "mysql", "mariadb":
		return dialectMySQL
	case "postgres", "pgx":
		return dialectPostgres
	default:
		return dialectUnknown
	}
}
