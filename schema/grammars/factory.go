package grammars

import (
	"fmt"

	"github.com/akfaiz/migris/internal/dialect"
	"github.com/akfaiz/migris/schema/blueprint"
)

// NewGrammar creates a new grammar instance for the given dialect.
func NewGrammar(dialectValue string) (blueprint.Grammar, error) {
	dialectVal := dialect.FromString(dialectValue)
	switch dialectVal {
	case dialect.MySQL:
		return newMysqlGrammar(), nil
	case dialect.MariaDB:
		return newMariadbGrammar(), nil
	case dialect.Postgres:
		return newPostgresGrammar(), nil
	case dialect.SQLite3:
		return newSqliteGrammar(), nil
	case dialect.Unknown:
		return nil, fmt.Errorf("unsupported Dialect: %s", dialectValue)
	default:
		return nil, fmt.Errorf("unsupported Dialect: %s", dialectValue)
	}
}
