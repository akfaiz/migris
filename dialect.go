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
	switch d {
	case "mysql", "mariadb":
		dialectValue = dialectMySQL
	case "postgres", "pgx":
		dialectValue = dialectPostgres
	default:
		return fmt.Errorf("unknown dialect: %s", d)
	}

	return nil
}

// SetDebug enables or disables debug mode for the schema package.
func SetDebug(d bool) {
	debug = d
}
