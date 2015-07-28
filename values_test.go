package kingpin

import (
	"github.com/stretchr/testify/assert"

	"testing"
)

func TestAccumulatorStrings(t *testing.T) {
	target := []string{}
	acc := newAccumulator(&target, func(v interface{}) Value { return newStringValue(v.(*string)) })
	acc.Set("a")
	assert.Equal(t, []string{"a"}, target)
	acc.Set("b")
	assert.Equal(t, []string{"a", "b"}, target)
}

func TestStrings(t *testing.T) {
	app := New("", "")
	app.Arg("a", "").Required().String()
	app.Arg("b", "").Required().String()
	c := app.Arg("c", "").Required().Strings()
	app.Parse([]string{"a", "b", "a", "b"})
	assert.Equal(t, []string{"a", "b"}, *c)
}
