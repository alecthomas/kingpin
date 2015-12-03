package kingpin

import (
	"net"

	"github.com/alecthomas/assert"

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

func TestEnum(t *testing.T) {
	app := New("", "")
	a := app.Arg("a", "").Enum("one", "two", "three")
	_, err := app.Parse([]string{"moo"})
	assert.Error(t, err)
	_, err = app.Parse([]string{"one"})
	assert.NoError(t, err)
	assert.Equal(t, "one", *a)
}

func TestEnumVar(t *testing.T) {
	app := New("", "")
	var a string
	app.Arg("a", "").EnumVar(&a, "one", "two", "three")
	_, err := app.Parse([]string{"moo"})
	assert.Error(t, err)
	_, err = app.Parse([]string{"one"})
	assert.NoError(t, err)
	assert.Equal(t, "one", a)
}

func TestCounter(t *testing.T) {
	app := New("", "")
	c := app.Flag("f", "").Counter()
	_, err := app.Parse([]string{"--f", "--f", "--f"})
	assert.NoError(t, err)
	assert.Equal(t, 3, *c)
}

func TestIPv4Addr(t *testing.T) {
	app := New("test", "")
	flag := app.Flag("addr", "").ResolvedIP()
	_, err := app.Parse([]string{"--addr", net.IPv4(1, 2, 3, 4).String()})
	assert.NoError(t, err)
	assert.NotNil(t, *flag)
	assert.Equal(t, net.IPv4(1, 2, 3, 4), *flag)
}

func TestInvalidIPv4Addr(t *testing.T) {
	app := New("test", "")
	app.Flag("addr", "").ResolvedIP()
	_, err := app.Parse([]string{"--addr", "1.2.3.256"})
	assert.Error(t, err)
}

func TestIPv6Addr(t *testing.T) {
	app := New("test", "")
	flag := app.Flag("addr", "").ResolvedIP()
	_, err := app.Parse([]string{"--addr", net.IPv6interfacelocalallnodes.String()})
	assert.NoError(t, err)
	assert.NotNil(t, *flag)
	assert.Equal(t, net.IPv6interfacelocalallnodes, *flag)
}
