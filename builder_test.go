package kingpin

import (
	"github.com/stretchrcom/testify/assert"

	"testing"
	"time"
)

func TestCommander(t *testing.T) {
	c := New("test", "test")
	ping := c.Command("ping", "Ping an IP address.")
	pingTTL := ping.Flag("ttl", "TTL for ICMP packets").Short('t').Default("5s").Duration()

	selected, err := c.Parse([]string{"ping"})
	assert.NoError(t, err)
	assert.Equal(t, "ping", selected)
	assert.Equal(t, 5*time.Second, *pingTTL)

	selected, err = c.Parse([]string{"ping", "--ttl=10s"})
	assert.NoError(t, err)
	assert.Equal(t, "ping", selected)
	assert.Equal(t, 10*time.Second, *pingTTL)
}

func TestRequiredFlags(t *testing.T) {
	c := New("test", "test")
	c.Flag("a", "a").String()
	c.Flag("b", "b").Required().String()

	_, err := c.Parse([]string{"--a=foo"})
	assert.Error(t, err)
	_, err = c.Parse([]string{"--b=foo"})
	assert.NoError(t, err)
}

func TestInvalidDefaultFlagValuePanics(t *testing.T) {
	c := New("test", "test")
	c.Flag("foo", "foo").Default("a").Int()
	assert.Panics(t, func() { c.Parse([]string{}) })
}

func TestInvalidDefaultArgValuePanics(t *testing.T) {
	c := New("test", "test")
	cmd := c.Command("cmd", "cmd")
	cmd.Arg("arg", "arg").Default("one").Int()
	assert.Panics(t, func() { c.Parse([]string{}) })
}

func TestArgsRequiredAfterNonRequiredPanics(t *testing.T) {
	c := New("test", "test")
	cmd := c.Command("cmd", "")
	cmd.Arg("a", "a").String()
	cmd.Arg("b", "b").Required().String()
	assert.Panics(t, func() { c.Parse([]string{}) })
}

func TestArgsMultipleRequiredThenNonRequired(t *testing.T) {
	c := New("test", "test")
	cmd := c.Command("cmd", "")
	cmd.Arg("a", "a").Required().String()
	cmd.Arg("b", "b").Required().String()
	cmd.Arg("c", "c").String()
	cmd.Arg("d", "d").String()

	assert.NotPanics(t, func() { c.Parse([]string{}) })
}

func TestDispatchCallbackIsCalled(t *testing.T) {
	dispatched := false
	c := New("test", "")
	c.Command("cmd", "").Dispatch(func() error {
		dispatched = true
		return nil
	})

	_, err := c.Parse([]string{"cmd"})
	assert.NoError(t, err)
	assert.True(t, dispatched)
}
