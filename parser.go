package kingpin

type ParseContext struct {
	Tokens          Tokens
	SelectedCommand string
}

func (p *ParseContext) Next() *Token {
	var token *Token
	token, p.Tokens = p.Tokens.Next()
	return token
}

func (p *ParseContext) Peek() *Token {
	return p.Tokens.Peek()
}

func (p *ParseContext) Return(token *Token) {
	p.Tokens = p.Tokens.Return(token)
}
