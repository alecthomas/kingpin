package kingpin

import "strings"

type tokenType int

// Token types.
const (
	TokenShort tokenType = iota
	TokenLong
	TokenArg
	TokenEOL
)

var (
	TokenEOLMarker = token{TokenEOL, ""}
)

type token struct {
	Type  tokenType
	Value string
}

func (t *token) IsFlag() bool {
	return t.Type == TokenShort || t.Type == TokenLong
}

func (t *token) IsEOF() bool {
	return t.Type == TokenEOL
}

func (t *token) String() string {
	switch t.Type {
	case TokenShort:
		return "-" + t.Value
	case TokenLong:
		return "--" + t.Value
	case TokenArg:
		return t.Value
	case TokenEOL:
		return "<EOL>"
	default:
		panic("unhandled type")
	}
}

type tokens []*token

func (t tokens) String() string {
	out := []string{}
	for _, tok := range t {
		out = append(out, tok.String())
	}
	return strings.Join(out, " ")
}

func (t tokens) Next() (*token, tokens) {
	if len(t) == 0 {
		return &TokenEOLMarker, nil
	}
	return t[0], t[1:]
}

func (t tokens) Return(token *token) tokens {
	if token.Type == TokenEOL {
		return t
	}
	return append(tokens{token}, t...)
}

func (t tokens) Peek() *token {
	if len(t) == 0 {
		return &TokenEOLMarker
	}
	return t[0]
}

func Tokenize(args []string) (tokens tokens) {
	for _, arg := range args {
		if strings.HasPrefix(arg, "--") {
			parts := strings.SplitN(arg[2:], "=", 2)
			tokens = append(tokens, &token{TokenLong, parts[0]})
			if len(parts) == 2 {
				tokens = append(tokens, &token{TokenArg, parts[1]})
			}
		} else if strings.HasPrefix(arg, "-") {
			for _, a := range arg[1:] {
				tokens = append(tokens, &token{TokenShort, string(a)})
			}
		} else {
			tokens = append(tokens, &token{TokenArg, arg})
		}
	}
	return
}
