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

var dialectValue dialect = dialectUnknown
var debug = false

// SetDialect sets the current dialect for the schema package.
func SetDialect(d string) error {
	dialectValue = dialectFromString(d)
	if dialectValue == dialectUnknown {
		return fmt.Errorf("unknown dialect: %s", d)
	}

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

// SetDebug enables or disables debug mode for the schema package.
func SetDebug(d bool) {
	debug = d
}
