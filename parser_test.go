package kingpin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseContextPush(t *testing.T) {
	c := tokenize([]string{"foo", "bar"}, false)
	a := c.Next()
	assert.Equal(t, TokenArg, a.Type)
	b := c.Next()
	assert.Equal(t, TokenArg, b.Type)
	c.Push(b)
	c.Push(a)
	a = c.Next()
	assert.Equal(t, "foo", a.Value)
	b = c.Next()
	assert.Equal(t, "bar", b.Value)
}
