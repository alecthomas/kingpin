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

func (c *cmdGroup) parse(tokens tokens) (selected []string, _ tokens, _ error) {
	token, tokens := tokens.Next()
	if token.Type != TokenArg {
		return nil, nil, fmt.Errorf("expected command but got '%s'", token)
	}
	cmd, ok := c.commands[token.String()]
	if !ok {
		return nil, nil, fmt.Errorf("no such command '%s'", token)
	}
	selected, tokens, err := cmd.parse(tokens)
	if err == nil {
		selected = append([]string{token.String()}, selected...)
	}
	return selected, tokens, err
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

func (c *CmdClause) parse(tokens tokens) (selected []string, _ tokens, _ error) {
	tokens, err := c.flagGroup.parse(tokens, false)
	if err != nil {
		return nil, nil, err
	}
	if c.cmdGroup.have() {
		selected, tokens, err = c.cmdGroup.parse(tokens)
	} else if c.argGroup.have() {
		tokens, err = c.argGroup.parse(tokens)
	}
	if err == nil && c.dispatch != nil {
		err = c.dispatch()
	}
	return selected, tokens, err
}
