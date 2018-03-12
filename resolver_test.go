package kingpin

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolverSimple(t *testing.T) {
	app := newTestApp()
	app.Resolver(MapResolver(map[string][]string{"hello": {"world"}}))
	f := app.Flag("hello", "help").String()
	_, err := app.Parse([]string{})
	require.NoError(t, err)
	require.Equal(t, "world", *f)
}

func TestResolverSatisfiesRequired(t *testing.T) {
	app := newTestApp()
	app.Resolver(MapResolver(map[string][]string{"hello": {"world"}}))
	f := app.Flag("hello", "help").Required().String()
	_, err := app.Parse([]string{})
	require.NoError(t, err)
	require.Equal(t, "world", *f)
}

func TestResolverLowerPriorityThanFlag(t *testing.T) {
	app := newTestApp()
	app.Resolver(MapResolver(map[string][]string{"hello": {"world"}}))
	f := app.Flag("hello", "help").String()
	_, err := app.Parse([]string{"--hello", "there"})
	require.NoError(t, err)
	require.Equal(t, "there", *f)
}

func TestResolverFallbackWithMultipleResolvers(t *testing.T) {
	app := newTestApp()
	app.Resolver(MapResolver(map[string][]string{"hello": {"there"}, "foo": {"bar"}}))
	app.Resolver(MapResolver(map[string][]string{"hello": {"world"}}))
	f1 := app.Flag("hello", "help").String()
	f2 := app.Flag("foo", "help").String()
	_, err := app.Parse([]string{})
	require.NoError(t, err)
	require.Equal(t, "world", *f1)
	require.Equal(t, "bar", *f2)
}

func TestResolverLowerPriorityThanEnvar(t *testing.T) {
	os.Setenv("TEST_RESOLVER", "foo")
	app := newTestApp()
	app.Resolver(MapResolver(map[string][]string{"hello": {"world"}}))
	f := app.Flag("hello", "help").Envar("TEST_RESOLVER").String()
	_, err := app.Parse([]string{})
	require.NoError(t, err)
	require.Equal(t, "foo", *f)
}

func TestEnvarResolverSplitting(t *testing.T) {
	os.Setenv("TEST_RESOLVER", "foo,bar")
	app := newTestApp()
	scalar := app.Flag("scalar", "").Envar("TEST_RESOLVER").String()
	vector := app.Flag("vector", "").Envar("TEST_RESOLVER").Strings(Separator(","))
	_, err := app.Parse(nil)
	require.NoError(t, err)
	require.Equal(t, "foo,bar", *scalar)
	require.Equal(t, []string{"foo", "bar"}, *vector)
}

func TestJSONResolver(t *testing.T) {
	r, err := JSONResolver(strings.NewReader(`{
		"str": "string",
		"num": 1234,
		"bool": true,
		"array": ["a", "b"]
	}`))
	require.NoError(t, err)
	app := newTestApp().Resolver(r)
	strf := app.Flag("str", "").String()
	numf := app.Flag("num", "").Int()
	boolf := app.Flag("bool", "").Bool()
	arrayf := app.Flag("array", "").Strings()

	_, err = app.Parse([]string{})
	require.NoError(t, err)

	require.Equal(t, "string", *strf)
	require.Equal(t, 1234, *numf)
	require.Equal(t, true, *boolf)
	require.Equal(t, []string{"a", "b"}, *arrayf)
}

func TestJSONConfigClause(t *testing.T) {
	w, err := ioutil.TempFile("", "kingpin-test-")
	require.NoError(t, err)
	w.WriteString(`{
		"str": "string",
		"num": 1234,
		"bool": true,
		"array": ["a", "b"]
	}`)
	defer os.Remove(w.Name())

	app := newTestApp()
	JSONConfigClause(app, app.Flag("config", "").Required())
	strf := app.Flag("str", "").String()
	numf := app.Flag("num", "").Int()
	boolf := app.Flag("bool", "").Bool()
	arrayf := app.Flag("array", "").Strings()

	_, err = app.Parse([]string{"--config", w.Name()})
	require.NoError(t, err)

	require.Equal(t, "string", *strf)
	require.Equal(t, 1234, *numf)
	require.Equal(t, true, *boolf)
	require.Equal(t, []string{"a", "b"}, *arrayf)
}
