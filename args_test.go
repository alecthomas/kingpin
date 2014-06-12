package kingpin

import (
	"github.com/stretchrcom/testify/assert"

	"testing"
)

func TestArgRemainder(t *testing.T) {
	a := newArgGroup()
	v := a.Arg("test", "").Strings()
	a.init()
	args := []string{"hello", "world"}
	tokens := Tokenize(args)
	a.parse(tokens)
	assert.Equal(t, args, *v)
}

func TestArgRemainderPanicsWhenNotLast(t *testing.T) {
	a := newArgGroup()
	a.Arg("test", "").Strings()
	a.Arg("test2", "").String()
	assert.Panics(t, func() { a.init() })
}
