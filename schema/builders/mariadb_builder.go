package builders

import (
	"github.com/akfaiz/migris/schema/grammars"
)

type mariadbBuilder struct {
	mysqlBuilder
}

var _ Builder = (*mariadbBuilder)(nil)

func NewMariadbBuilder() Builder {
	grammar, _ := grammars.NewGrammar("mariadb")
	b := &mariadbBuilder{}
	b.baseBuilder = baseBuilder{Grammar: grammar, Outer: b}
	return b
}
