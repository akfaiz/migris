package schema

import "errors"

var ErrDialectNotSet = errors.New("dialect not set")
var ErrUnknownDialect = errors.New("unknown dialect")

var dialect string = ""

var supportedDialects = map[string]bool{
	"postgres": true,
	"pgx":      true,
}

func SetDialect(d string) error {
	if _, ok := supportedDialects[d]; !ok {
		return ErrUnknownDialect
	}

	dialect = d
	return nil
}

func newGrammar() (grammar, error) {
	switch dialect {
	case "postgres", "pgx":
		return newPgGrammar(), nil
	default:
		return nil, ErrDialectNotSet
	}
}
