package kingpin

import (
	"bufio"
	"os"
	"strings"
)

type TokenType int

// Token types.
const (
	TokenShort TokenType = iota
	TokenLong
	TokenArg
	TokenEOL
)

var (
	TokenEOLMarker = Token{TokenEOL, ""}
)

type Token struct {
	Type  TokenType
	Value string
}

func (t *Token) IsFlag() bool {
	return t.Type == TokenShort || t.Type == TokenLong
}

func (t *Token) IsEOF() bool {
	return t.Type == TokenEOL
}

func (t *Token) String() string {
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

type ParseContext struct {
	SelectedCommand string
	argsOnly        bool
	peek            []*Token
	args            []string
	flags           *flagGroup
}

func tokenize(args []string) *ParseContext {
	return &ParseContext{
		args:  args,
		flags: newFlagGroup(),
	}
}

func (p *ParseContext) mergeFlags(flags *flagGroup) {
	for _, flag := range flags.flagOrder {
		if flag.shorthand != 0 {
			p.flags.short[string(flag.shorthand)] = flag
		}
		p.flags.long[flag.name] = flag
		p.flags.flagOrder = append(p.flags.flagOrder, flag)
	}
}

func (p *ParseContext) EOL() bool {
	return p.Peek().Type == TokenEOL
}

// Next token in the parse context.
func (p *ParseContext) Next() *Token {
	if len(p.peek) > 0 {
		return p.pop()
	}

	// End of tokens.
	if len(p.args) == 0 {
		return &Token{Type: TokenEOL}
	}

	arg := p.args[0]
	p.args = p.args[1:]

	if p.argsOnly {
		return &Token{Type: TokenArg, Value: arg}
	}

	// All remaining args are passed directly.
	if arg == "--" {
		p.argsOnly = true
		return p.Next()
	}

	if strings.HasPrefix(arg, "--") {
		parts := strings.SplitN(arg[2:], "=", 2)
		token := &Token{TokenLong, parts[0]}
		if len(parts) == 2 {
			p.push(&Token{TokenArg, parts[1]})
		}
		return token
	}

	if strings.HasPrefix(arg, "-") {
		if len(arg) == 1 {
			return &Token{Type: TokenShort}
		}
		short := arg[1:2]
		flag, ok := p.flags.short[short]
		// Not a known short flag, we'll just return it anyway.
		if !ok {
		} else if fb, ok := flag.value.(boolFlag); ok && fb.IsBoolFlag() {
			// Bool short flag.
		} else {
			// Short flag with combined argument: -fARG
			token := &Token{TokenShort, short}
			if len(arg) > 2 {
				p.push(&Token{TokenArg, arg[2:]})
			}
			return token
		}

		p.args = append([]string{"-" + arg[2:]}, p.args...)
		return &Token{TokenShort, short}
	}

	return &Token{TokenArg, arg}
}

func (p *ParseContext) Peek() *Token {
	if len(p.peek) == 0 {
		return p.push(p.Next())
	}
	return p.peek[len(p.peek)-1]
}

func (p *ParseContext) push(token *Token) *Token {
	p.peek = append(p.peek, token)
	return token
}

func (p *ParseContext) pop() *Token {
	end := len(p.peek) - 1
	token := p.peek[end]
	p.peek = p.peek[0:end]
	return token
}

func (p *ParseContext) String() string {
	return p.SelectedCommand
}

// ExpandArgsFromFiles expands arguments in the form @<file> into one-arg-per-
// line read from that file.
func ExpandArgsFromFiles(args []string) ([]string, error) {
	out := []string{}
	for _, arg := range args {
		if strings.HasPrefix(arg, "@") {
			r, err := os.Open(arg[1:])
			if err != nil {
				return nil, err
			}
			scanner := bufio.NewScanner(r)
			for scanner.Scan() {
				out = append(out, scanner.Text())
			}
			r.Close()
			if scanner.Err() != nil {
				return nil, scanner.Err()
			}
		} else {
			out = append(out, arg)
		}
	}
	return out, nil
}
