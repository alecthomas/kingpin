package kingpin

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolverSimple(t *testing.T) {
	app := newTestApp()
	app.Resolver(MapResolver(map[string][]string{"hello": []string{"world"}}))
	f := app.Flag("hello", "help").String()
	_, err := app.Parse([]string{})
	assert.NoError(t, err)
	assert.Equal(t, "world", *f)
}

func TestResolverSatisfiesRequired(t *testing.T) {
	app := newTestApp()
	app.Resolver(MapResolver(map[string][]string{"hello": []string{"world"}}))
	f := app.Flag("hello", "help").Required().String()
	_, err := app.Parse([]string{})
	assert.NoError(t, err)
	assert.Equal(t, "world", *f)
}

func TestResolverLowerPriorityThanFlag(t *testing.T) {
	app := newTestApp()
	app.Resolver(MapResolver(map[string][]string{"hello": []string{"world"}}))
	f := app.Flag("hello", "help").String()
	_, err := app.Parse([]string{"--hello", "there"})
	assert.NoError(t, err)
	assert.Equal(t, "there", *f)
}

func TestResolverFallbackWithMultipleResolvers(t *testing.T) {
	app := newTestApp()
	app.Resolver(MapResolver(map[string][]string{"hello": []string{"there"}, "foo": []string{"bar"}}))
	app.Resolver(MapResolver(map[string][]string{"hello": []string{"world"}}))
	f1 := app.Flag("hello", "help").String()
	f2 := app.Flag("foo", "help").String()
	_, err := app.Parse([]string{})
	assert.NoError(t, err)
	assert.Equal(t, "world", *f1)
	assert.Equal(t, "bar", *f2)
}

func TestResolverLowerPriorityThanEnvar(t *testing.T) {
	os.Setenv("TEST_RESOLVER", "foo")
	app := newTestApp()
	app.Resolver(MapResolver(map[string][]string{"hello": []string{"world"}}))
	f := app.Flag("hello", "help").Envar("TEST_RESOLVER").String()
	_, err := app.Parse([]string{})
	assert.NoError(t, err)
	assert.Equal(t, "foo", *f)
}

func TestJSONResolver(t *testing.T) {
	r, err := JSONResolver([]byte(`{
		"str": "string",
		"num": 1234,
		"bool": true,
		"array": ["a", "b"]
	}`))
	assert.NoError(t, err)

	values, err := r.Resolve("str", nil)
	assert.NoError(t, err)
	assert.Equal(t, []string{"string"}, values)

	values, err = r.Resolve("num", nil)
	assert.NoError(t, err)
	assert.Equal(t, []string{"1234"}, values)

	values, err = r.Resolve("bool", nil)
	assert.NoError(t, err)
	assert.Equal(t, []string{"true"}, values)

	values, err = r.Resolve("array", nil)
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, values)
}
