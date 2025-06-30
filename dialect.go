package schema

import "errors"

var ErrDialectNotSet = errors.New("dialect not set")
var ErrUnknownDialect = errors.New("unknown dialect")

type dialectType uint8

const (
	unknown dialectType = iota // Represents no dialect set
	postgres
	mysql
)

var dialect dialectType
var debug bool = false

var supportedDialects = map[string]dialectType{
	"postgres": postgres,
	"pgx":      postgres,
	"mysql":    mysql,
}

// SetDialect sets the current dialect for the schema package.
func SetDialect(d string) error {
	v, ok := supportedDialects[d]
	if !ok {
		return ErrUnknownDialect
	}
	if v == unknown {
		return ErrDialectNotSet
	}
	dialect = v

	return nil
}

// SetDebug enables or disables debug mode for the schema package.
func SetDebug(d bool) {
	debug = d
}
