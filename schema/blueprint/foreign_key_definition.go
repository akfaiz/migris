package blueprint

type ForeignKeyDefinition interface {
	References(columns ...string) ForeignKeyDefinition
	On(table string) ForeignKeyDefinition
	OnDelete(action string) ForeignKeyDefinition
	OnUpdate(action string) ForeignKeyDefinition
	Name(name string) ForeignKeyDefinition
	CascadeOnDelete() ForeignKeyDefinition
	CascadeOnUpdate() ForeignKeyDefinition
	RestrictOnDelete() ForeignKeyDefinition
	RestrictOnUpdate() ForeignKeyDefinition
	NullOnDelete() ForeignKeyDefinition
	NullOnUpdate() ForeignKeyDefinition
	NoActionOnDelete() ForeignKeyDefinition
	NoActionOnUpdate() ForeignKeyDefinition
	Deferrable(deferrable ...bool) ForeignKeyDefinition
	InitiallyImmediate(immediate ...bool) ForeignKeyDefinition
}

type foreignKeyDefinition struct {
	command *Command
}

func (f *foreignKeyDefinition) References(columns ...string) ForeignKeyDefinition {
	f.command.References = columns
	return f
}

func (f *foreignKeyDefinition) On(table string) ForeignKeyDefinition {
	f.command.On = table
	return f
}

func (f *foreignKeyDefinition) OnDelete(action string) ForeignKeyDefinition {
	f.command.OnDelete = action
	return f
}

func (f *foreignKeyDefinition) OnUpdate(action string) ForeignKeyDefinition {
	f.command.OnUpdate = action
	return f
}

func (f *foreignKeyDefinition) Name(name string) ForeignKeyDefinition {
	f.command.Index = name
	return f
}

func (f *foreignKeyDefinition) CascadeOnDelete() ForeignKeyDefinition {
	return f.OnDelete("CASCADE")
}

func (f *foreignKeyDefinition) CascadeOnUpdate() ForeignKeyDefinition {
	return f.OnUpdate("CASCADE")
}

func (f *foreignKeyDefinition) RestrictOnDelete() ForeignKeyDefinition {
	return f.OnDelete("RESTRICT")
}

func (f *foreignKeyDefinition) RestrictOnUpdate() ForeignKeyDefinition {
	return f.OnUpdate("RESTRICT")
}

func (f *foreignKeyDefinition) NullOnDelete() ForeignKeyDefinition {
	return f.OnDelete("SET NULL")
}

func (f *foreignKeyDefinition) NullOnUpdate() ForeignKeyDefinition {
	return f.OnUpdate("SET NULL")
}

func (f *foreignKeyDefinition) NoActionOnDelete() ForeignKeyDefinition {
	return f.OnDelete("NO ACTION")
}

func (f *foreignKeyDefinition) NoActionOnUpdate() ForeignKeyDefinition {
	return f.OnUpdate("NO ACTION")
}

func (f *foreignKeyDefinition) Deferrable(deferrable ...bool) ForeignKeyDefinition {
	val := true
	if len(deferrable) > 0 {
		val = deferrable[0]
	}
	f.command.Deferrable = &val
	return f
}

func (f *foreignKeyDefinition) InitiallyImmediate(immediate ...bool) ForeignKeyDefinition {
	val := true
	if len(immediate) > 0 {
		val = immediate[0]
	}
	f.command.InitiallyImmediate = &val
	return f
}
