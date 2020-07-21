package kingpin

// FlagRegistrar will be executed before pre-actions and validation to register its flags.
type FlagRegistrar interface {
	// RegisterFlags will register flags at the given FlagGroup
	RegisterFlags(FlagGroup)
}

type FlagRegistrarFunc func(FlagGroup)

func (r FlagRegistrarFunc) RegisterFlags(f FlagGroup) {
	r(f)
}

func registerFlagsOf(fg FlagGroup, registrars ...FlagRegistrar) {
	for _, r := range registrars {
		r.RegisterFlags(fg)
	}
}

// CmdRegistrar will be executed before pre-actions and validation to register its commands.
type CmdRegistrar interface {
	// RegisterFlags will register commands at the given Cmd
	RegisterCommands(Cmd)
}

type CmdRegistrarFunc func(Cmd)

func (r CmdRegistrarFunc) RegisterCommands(c Cmd) {
	r(c)
}

func registerCommandsOf(c Cmd, registrars ...CmdRegistrar) {
	for _, r := range registrars {
		r.RegisterCommands(c)
	}
}
