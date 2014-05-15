package kingpin

type CmdClause struct {
	*flagGroup
	*argGroup
	name     string
	help     string
	dispatch Dispatch
}

func newCommand(name, help string) *CmdClause {
	return &CmdClause{
		flagGroup: newFlagGroup(),
		argGroup:  newArgGroup(),
		name:      name,
		help:      help,
	}
}

func (c *CmdClause) Dispatch(dispatch Dispatch) *CmdClause {
	c.dispatch = dispatch
	return c
}

func (c *CmdClause) init() {
	c.flagGroup.init()
	c.argGroup.init()
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
