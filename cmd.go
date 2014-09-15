package kingpin

import "fmt"

type cmdGroup struct {
	commands     map[string]*CmdClause
	commandOrder []*CmdClause
}

func newCmdGroup() *cmdGroup {
	return &cmdGroup{
		commands: make(map[string]*CmdClause),
	}
}

// Command adds a new top-level command to the application.
func (c *cmdGroup) Command(name, help string) *CmdClause {
	cmd := newCommand(name, help)
	c.commands[name] = cmd
	c.commandOrder = append(c.commandOrder, cmd)
	return cmd
}

func (c *cmdGroup) init() error {
	seen := map[string]bool{}
	for _, cmd := range c.commandOrder {
		if seen[cmd.name] {
			return fmt.Errorf("duplicate command '%s'", cmd.name)
		}
		seen[cmd.name] = true
	}
	return nil
}

func (c *cmdGroup) parse(context *ParseContext) (selected []string, _ error) {
	token := context.Next()
	if token.Type != TokenArg {
		return nil, fmt.Errorf("expected command but got '%s'", token)
	}
	cmd, ok := c.commands[token.String()]
	if !ok {
		return nil, fmt.Errorf("no such command '%s'", token)
	}
	context.SelectedCommand = cmd.name
	selected, err := cmd.parse(context)
	if err == nil {
		selected = append([]string{token.String()}, selected...)
	}
	return selected, err
}

func (c *cmdGroup) have() bool {
	return len(c.commands) > 0
}

// A CmdClause is a single top-level command. It encapsulates a set of flags
// and either subcommands or positional arguments.
type CmdClause struct {
	*flagGroup
	*argGroup
	*cmdGroup
	name     string
	help     string
	dispatch Dispatch
}

func newCommand(name, help string) *CmdClause {
	c := &CmdClause{
		flagGroup: newFlagGroup(),
		argGroup:  newArgGroup(),
		cmdGroup:  newCmdGroup(),
		name:      name,
		help:      help,
	}
	return c
}

func (c *CmdClause) Dispatch(dispatch Dispatch) *CmdClause {
	c.dispatch = dispatch
	return c
}

func (c *CmdClause) init() error {
	if err := c.flagGroup.init(); err != nil {
		return err
	}
	if c.argGroup.have() && c.cmdGroup.have() {
		return fmt.Errorf("can't mix Arg()s with Command()s")
	}
	if err := c.argGroup.init(); err != nil {
		return err
	}
	if err := c.cmdGroup.init(); err != nil {
		return err
	}
	return nil
}

func (c *CmdClause) parse(context *ParseContext) (selected []string, _ error) {
	err := c.flagGroup.parse(context, false)
	if err != nil {
		return nil, err
	}
	if c.cmdGroup.have() {
		selected, err = c.cmdGroup.parse(context)
	} else if c.argGroup.have() {
		err = c.argGroup.parse(context)
	}
	if err == nil && c.dispatch != nil {
		err = c.dispatch()
	}
	return selected, err
}
