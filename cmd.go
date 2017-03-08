package kingpin

import "strings"

type cmdMixin struct {
	actionMixin
	*flagGroup
	*argGroup
	*cmdGroup
}

// CmdCompletion returns completion options for arguments, if that's where
// parsing left off, or commands if there aren't any unsatisfied args.
func (c *cmdMixin) CmdCompletion(context *ParseContext) []string {
	var options []string

	// Count args already satisfied - we won't complete those, and add any
	// default commands' alternatives, since they weren't listed explicitly
	// and the user may want to explicitly list something else.
	argsSatisfied := 0
	for _, el := range context.Elements {
		switch {
		case el.OneOf.Arg != nil:
			if el.Value != nil && *el.Value != "" {
				argsSatisfied++
			}
		case el.OneOf.Cmd != nil:
			options = append(options, el.OneOf.Cmd.completionAlts...)
		default:
		}
	}

	if argsSatisfied < len(c.argGroup.args) {
		// Since not all args have been satisfied, show options for the current one
		options = append(options, c.argGroup.args[argsSatisfied].resolveCompletions()...)
	} else {
		// If all args are satisfied, then go back to completing commands
		for _, cmd := range c.cmdGroup.commandOrder {
			if !cmd.hidden {
				options = append(options, cmd.name)
			}
		}
	}

	return options
}

func (c *cmdMixin) FlagCompletion(flagName string, flagValue string) (choices []string, flagMatch bool, optionMatch bool) {
	// Check if flagName matches a known flag.
	// If it does, show the options for the flag
	// Otherwise, show all flags

	options := []string{}

	for _, flag := range c.flagGroup.flagOrder {
		// Loop through each flag and determine if a match exists
		if flag.name == flagName {
			// User typed entire flag. Need to look for flag options.
			options = flag.resolveCompletions()
			if len(options) == 0 {
				// No Options to Choose From, Assume Match.
				return options, true, true
			}

			// Loop options to find if the user specified value matches
			isPrefix := false
			matched := false

			for _, opt := range options {
				if flagValue == opt {
					matched = true
				} else if strings.HasPrefix(opt, flagValue) {
					isPrefix = true
				}
			}

			// Matched Flag Directly
			// Flag Value Not Prefixed, and Matched Directly
			return options, true, !isPrefix && matched
		}

		if !flag.hidden {
			options = append(options, "--"+flag.name)
		}
	}
	// No Flag directly matched.
	return options, false, false

}

type cmdGroup struct {
	app          *Application
	parent       *CmdClause
	commands     map[string]*CmdClause
	commandOrder []*CmdClause
}

func (c *cmdGroup) defaultSubcommand() *CmdClause {
	for _, cmd := range c.commandOrder {
		if cmd.isDefault {
			return cmd
		}
	}
	return nil
}

func (c *cmdGroup) cmdNames() []string {
	names := make([]string, 0, len(c.commandOrder))
	for _, cmd := range c.commandOrder {
		names = append(names, cmd.name)
	}
	return names
}

// GetArg gets a command definition.
//
// This allows existing commands to be modified after definition but before parsing. Useful for
// modular applications.
func (c *cmdGroup) GetCommand(name string) *CmdClause {
	return c.commands[name]
}

func newCmdGroup(app *Application) *cmdGroup {
	return &cmdGroup{
		app:      app,
		commands: make(map[string]*CmdClause),
	}
}

func (c *cmdGroup) flattenedCommands() (out []*CmdClause) {
	for _, cmd := range c.commandOrder {
		if len(cmd.commands) == 0 {
			out = append(out, cmd)
		}
		out = append(out, cmd.flattenedCommands()...)
	}
	return
}

func (c *cmdGroup) addCommand(name, help string) *CmdClause {
	cmd := newCommand(c.app, name, help)
	c.commands[name] = cmd
	c.commandOrder = append(c.commandOrder, cmd)
	return cmd
}

func (c *cmdGroup) init() error {
	seen := map[string]bool{}
	if c.defaultSubcommand() != nil && !c.have() {
		return TError("default subcommand {{.Arg0}} provided but no subcommands defined", V{"Arg0": c.defaultSubcommand().name})
	}
	defaults := []string{}
	for _, cmd := range c.commandOrder {
		if cmd.isDefault {
			defaults = append(defaults, cmd.name)
		}
		if seen[cmd.name] {
			return TError("duplicate command {{.Arg0}}", V{"Arg0": cmd.name})
		}
		seen[cmd.name] = true
		for _, alias := range cmd.aliases {
			if seen[alias] {
				return TError("alias duplicates existing command {{.Arg0}}", V{"Arg0": alias})
			}
			c.commands[alias] = cmd
		}
		if err := cmd.init(); err != nil {
			return err
		}
	}
	if len(defaults) > 1 {
		return TError("more than one default subcommand exists: {{.Arg0}}", V{"Arg0": strings.Join(defaults, ", ")})
	}
	return nil
}

func (c *cmdGroup) have() bool {
	return len(c.commands) > 0
}

type CmdClauseValidator func(*CmdClause) error

// A CmdClause is a single top-level command. It encapsulates a set of flags
// and either subcommands or positional arguments.
type CmdClause struct {
	cmdMixin
	app            *Application
	name           string
	aliases        []string
	help           string
	isDefault      bool
	validator      CmdClauseValidator
	hidden         bool
	completionAlts []string
}

func newCommand(app *Application, name, help string) *CmdClause {
	c := &CmdClause{
		app:  app,
		name: name,
		help: help,
	}
	c.flagGroup = newFlagGroup()
	c.argGroup = newArgGroup()
	c.cmdGroup = newCmdGroup(app)
	return c
}

// Struct allows applications to define flags with struct tags.
//
// Supported struct tags are: help, placeholder, default, short, long, required, hidden, enum, and
// arg.
//
// The name of the flag will default to the CamelCase name transformed to camel-case. This can
// be overridden with the "long" tag.
//
// All basic Go types are supported including floats, ints, strings, time.Duration,
// and slices of same.
//
// For compatibility, also supports the tags used by https://github.com/jessevdk/go-flags
func (c *CmdClause) Struct(v interface{}) error {
	return c.fromStruct(c, v)
}

// Add an Alias for this command.
func (c *CmdClause) Alias(name string) *CmdClause {
	c.aliases = append(c.aliases, name)
	return c
}

// Validate sets a validation function to run when parsing.
func (c *CmdClause) Validate(validator CmdClauseValidator) *CmdClause {
	c.validator = validator
	return c
}

func (c *CmdClause) FullCommand() string {
	out := []string{c.name}
	for p := c.parent; p != nil; p = p.parent {
		out = append([]string{p.name}, out...)
	}
	return strings.Join(out, " ")
}

// Command adds a new sub-command.
func (c *CmdClause) Command(name, help string) *CmdClause {
	cmd := c.addCommand(name, help)
	cmd.parent = c
	return cmd
}

// Default makes this command the default if commands don't match.
func (c *CmdClause) Default() *CmdClause {
	c.isDefault = true
	return c
}

func (c *CmdClause) Action(action Action) *CmdClause {
	c.addAction(action)
	return c
}

func (c *CmdClause) PreAction(action Action) *CmdClause {
	c.addPreAction(action)
	return c
}

func (c *CmdClause) init() error {
	if err := c.flagGroup.init(c.app.defaultEnvarPrefix()); err != nil {
		return err
	}
	if c.argGroup.have() && c.cmdGroup.have() {
		return TError("can't mix Arg()s with Command()s")
	}
	if err := c.argGroup.init(); err != nil {
		return err
	}
	if err := c.cmdGroup.init(); err != nil {
		return err
	}
	return nil
}

func (c *CmdClause) Hidden() *CmdClause {
	c.hidden = true
	return c
}
