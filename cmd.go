package kingpin

import "os"

type CmdClause struct {
	app *Application
	*flagGroup
	*argGroup
	name     string
	help     string
	dispatch Dispatch
}

func newCommand(app *Application, name, help string) *CmdClause {
	c := &CmdClause{
		app:       app,
		flagGroup: newFlagGroup(),
		argGroup:  newArgGroup(),
		name:      name,
		help:      help,
	}
	c.Flag("help", "Show help.").Dispatch(c.onFlagHelp).Hidden().Bool()
	return c
}

func (c *CmdClause) onFlagHelp() error {
	c.app.CommandUsage(os.Stderr, c.name)
	os.Exit(0)
	return nil
}

func (c *CmdClause) Dispatch(dispatch Dispatch) *CmdClause {
	c.dispatch = dispatch
	return c
}

func (c *CmdClause) init() error {
	if err := c.flagGroup.init(); err != nil {
		return err
	}
	if err := c.argGroup.init(); err != nil {
		return err
	}
	return nil
}

func (c *CmdClause) parse(tokens tokens) (tokens, error) {
	tokens, err := c.flagGroup.parse(tokens, false)
	if err != nil {
		return nil, err
	}
	if tokens, err = c.argGroup.parse(tokens); err != nil {
		return nil, err
	}
	if c.dispatch != nil {
		return tokens, c.dispatch()
	}
	return tokens, nil
}
