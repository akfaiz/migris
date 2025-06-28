package schema

type grammar interface {
	compileCreate(bp *Blueprint) ([]string, error)
	compileCreateIfNotExists(bp *Blueprint) ([]string, error)
	compileAlter(bp *Blueprint) ([]string, error)
	compileDrop(bp *Blueprint) (string, error)
	compileDropIfExists(bp *Blueprint) (string, error)
	compileRename(bp *Blueprint) (string, error)
}
