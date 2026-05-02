package schema

type mariadbBuilder struct {
	mysqlBuilder
}

var _ Builder = (*mariadbBuilder)(nil)

func newMariadbBuilder() Builder {
	grammar := newMariadbGrammar()
	b := &mariadbBuilder{}
	b.baseBuilder = baseBuilder{grammar: grammar, outer: b}
	return b
}
