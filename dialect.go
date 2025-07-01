package schema

import (
	"fmt"
	"strings"
)

var dialect string
var debug = false

var supportedDialects = map[string]bool{
	"postgres": true,
	"pgx":      true,
	"mysql":    true,
	"mariadb":  true,
}

// SetDialect sets the current dialect for the schema package.
func SetDialect(d string) error {
	_, ok := supportedDialects[d]
	if !ok {
		supportedDialectList := make([]string, 0, len(supportedDialects))
		for key := range supportedDialects {
			supportedDialectList = append(supportedDialectList, key)
		}
		return fmt.Errorf("unknown dialect: %s, supported dialects are: %s", d, strings.Join(supportedDialectList, ", "))
	}
	dialect = d

	return nil
}

// SetDebug enables or disables debug mode for the schema package.
func SetDebug(d bool) {
	debug = d
}
