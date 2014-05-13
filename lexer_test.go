package kingpin

import (
	"testing"

	"github.com/stretchrcom/testify/assert"
)

func TestLexer(t *testing.T) {
	tokens := Tokenize([]string{"-abc", "foo", "--foo=bar", "--bar", "foo"})
	assert.Equal(t, 8, len(tokens))
	token, tokens := tokens.Next()
	assert.Equal(t, &Token{TokenShort, "a"}, token)
	token, tokens = tokens.Next()
	assert.Equal(t, &Token{TokenShort, "b"}, token)
	token, tokens = tokens.Next()
	assert.Equal(t, &Token{TokenShort, "c"}, token)
	token, tokens = tokens.Next()
	assert.Equal(t, &Token{TokenArg, "foo"}, token)
	token, tokens = tokens.Next()
	assert.Equal(t, &Token{TokenLong, "foo"}, token)
	token, tokens = tokens.Next()
	assert.Equal(t, &Token{TokenArg, "bar"}, token)
	token, tokens = tokens.Next()
	assert.Equal(t, &Token{TokenLong, "bar"}, token)
	token, tokens = tokens.Next()
	assert.Equal(t, &Token{TokenArg, "foo"}, token)
}
