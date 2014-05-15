package kingpin

import (
	"testing"

	"github.com/stretchrcom/testify/assert"
)

func TestLexer(t *testing.T) {
	tokens := Tokenize([]string{"-abc", "foo", "--foo=bar", "--bar", "foo"})
	assert.Equal(t, 8, len(tokens))
	tok, tokens := tokens.Next()
	assert.Equal(t, &token{TokenShort, "a"}, tok)
	tok, tokens = tokens.Next()
	assert.Equal(t, &token{TokenShort, "b"}, tok)
	tok, tokens = tokens.Next()
	assert.Equal(t, &token{TokenShort, "c"}, tok)
	tok, tokens = tokens.Next()
	assert.Equal(t, &token{TokenArg, "foo"}, tok)
	tok, tokens = tokens.Next()
	assert.Equal(t, &token{TokenLong, "foo"}, tok)
	tok, tokens = tokens.Next()
	assert.Equal(t, &token{TokenArg, "bar"}, tok)
	tok, tokens = tokens.Next()
	assert.Equal(t, &token{TokenLong, "bar"}, tok)
	tok, tokens = tokens.Next()
	assert.Equal(t, &token{TokenArg, "foo"}, tok)
}
