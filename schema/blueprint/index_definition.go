package blueprint

type IndexDefinition interface {
	Name(name string) IndexDefinition
	Algorithm(algorithm string) IndexDefinition
	Deferrable(deferrable ...bool) IndexDefinition
	InitiallyImmediate(immediate ...bool) IndexDefinition
	Language(language string) IndexDefinition
}

type indexDefinition struct {
	command *Command
}

func (i *indexDefinition) Name(name string) IndexDefinition {
	i.command.Index = name
	return i
}

func (i *indexDefinition) Algorithm(algorithm string) IndexDefinition {
	i.command.Algorithm = algorithm
	return i
}

func (i *indexDefinition) Deferrable(deferrable ...bool) IndexDefinition {
	val := true
	if len(deferrable) > 0 {
		val = deferrable[0]
	}
	i.command.Deferrable = &val
	return i
}

func (i *indexDefinition) InitiallyImmediate(immediate ...bool) IndexDefinition {
	val := true
	if len(immediate) > 0 {
		val = immediate[0]
	}
	i.command.InitiallyImmediate = &val
	return i
}

func (i *indexDefinition) Language(language string) IndexDefinition {
	i.command.Language = language
	return i
}
