package kingpin

type ParseContext struct {
	Tokens          Tokens
	SelectedCommand string
	flags           *flagGroup
}

func newParseContext(tokens Tokens) *ParseContext {
	return &ParseContext{
		Tokens: tokens,
		flags:  newFlagGroup(),
	}
}

func (p *ParseContext) mergeFlags(flags *flagGroup) error {
	for _, flag := range flags.flagOrder {
		if flag.shorthand != 0 {
			p.flags.short[string(flag.shorthand)] = flag
		}
		p.flags.long[flag.name] = flag
		p.flags.flagOrder = append(p.flags.flagOrder, flag)
	}
	return nil
}

func (p *ParseContext) Next() {
	p.Tokens = p.Tokens.Next()
}

func (p *ParseContext) Peek() *Token {
	return p.Tokens.Peek()
}

func (p *ParseContext) Return(token *Token) {
	p.Tokens = p.Tokens.Return(token)
}

func (p *ParseContext) String() string {
	return p.SelectedCommand + ": " + p.Tokens.String()
}
