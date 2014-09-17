package kingpin

import (
	"testing"

	"github.com/stretchrcom/testify/assert"
)

func TestLexer(t *testing.T) {
	tokens := Tokenize([]string{"-abc", "foo", "--foo=bar", "--bar", "foo"}).Tokens
	assert.Equal(t, 8, len(tokens))
	tok := tokens.Peek()
	assert.Equal(t, &Token{TokenShort, "a"}, tok)
	tokens = tokens.Next()
	tok = tokens.Peek()
	assert.Equal(t, &Token{TokenShort, "b"}, tok)
	tokens = tokens.Next()
	tok = tokens.Peek()
	assert.Equal(t, &Token{TokenShort, "c"}, tok)
	tokens = tokens.Next()
	tok = tokens.Peek()
	assert.Equal(t, &Token{TokenArg, "foo"}, tok)
	tokens = tokens.Next()
	tok = tokens.Peek()
	assert.Equal(t, &Token{TokenLong, "foo"}, tok)
	tokens = tokens.Next()
	tok = tokens.Peek()
	assert.Equal(t, &Token{TokenArg, "bar"}, tok)
	tokens = tokens.Next()
	tok = tokens.Peek()
	assert.Equal(t, &Token{TokenLong, "bar"}, tok)
	tokens = tokens.Next()
	tok = tokens.Peek()
	assert.Equal(t, &Token{TokenArg, "foo"}, tok)
	tokens = tokens.Next()
}
