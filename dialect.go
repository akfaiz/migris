package schema

import "errors"

var ErrDialectNotSet = errors.New("dialect not set")
var ErrUnknownDialect = errors.New("unknown dialect")

type dialectType uint8

const (
	emptyDialect dialectType = iota // Represents no dialect set
	postgres
)

var dialect dialectType
var debug bool = false

var supportedDialects = map[string]dialectType{
	"postgres": postgres,
	"pgx":      postgres,
}

// SetDialect sets the current dialect for the schema package.
func SetDialect(d string) error {
	v, ok := supportedDialects[d]
	if !ok {
		return ErrUnknownDialect
	}
	if v == emptyDialect {
		return ErrDialectNotSet
	}
	dialect = v

	return nil
}

// SetDebug enables or disables debug mode for the schema package.
func SetDebug(d bool) {
	debug = d
}

func newGrammar() (grammar, error) {
	switch dialect {
	case postgres:
		return newPgGrammar(), nil
	default:
		return nil, ErrDialectNotSet
	}
}
