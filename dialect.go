package schema

import "errors"

var ErrDialectNotSet = errors.New("dialect not set")
var ErrUnknownDialect = errors.New("unknown dialect")

var dialect string
var debug bool = false

var supportedDialects = map[string]bool{
	"postgres": true,
	"pgx":      true,
	"mysql":    true,
}

// SetDialect sets the current dialect for the schema package.
func SetDialect(d string) error {
	_, ok := supportedDialects[d]
	if !ok {
		return ErrUnknownDialect
	}
	dialect = d

	return nil
}

// SetDebug enables or disables debug mode for the schema package.
func SetDebug(d bool) {
	debug = d
}
