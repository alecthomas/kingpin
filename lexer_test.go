package kingpin

import (
	"testing"

	"github.com/stretchrcom/testify/assert"
)

func TestLexer(t *testing.T) {
	tokens := Tokenize([]string{"-abc", "foo", "--foo=bar", "--bar", "foo"}).Tokens
	assert.Equal(t, 8, len(tokens))
	tok, tokens := tokens.Next()
	assert.Equal(t, &Token{TokenShort, "a"}, tok)
	tok, tokens = tokens.Next()
	assert.Equal(t, &Token{TokenShort, "b"}, tok)
	tok, tokens = tokens.Next()
	assert.Equal(t, &Token{TokenShort, "c"}, tok)
	tok, tokens = tokens.Next()
	assert.Equal(t, &Token{TokenArg, "foo"}, tok)
	tok, tokens = tokens.Next()
	assert.Equal(t, &Token{TokenLong, "foo"}, tok)
	tok, tokens = tokens.Next()
	assert.Equal(t, &Token{TokenArg, "bar"}, tok)
	tok, tokens = tokens.Next()
	assert.Equal(t, &Token{TokenLong, "bar"}, tok)
	tok, tokens = tokens.Next()
	assert.Equal(t, &Token{TokenArg, "foo"}, tok)
}
