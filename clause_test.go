package kingpin

import (
	"io/ioutil"
	"net/url"
	"os"

	"github.com/stretchr/testify/assert"

	"testing"
)

func TestParseStrings(t *testing.T) {
	p := Clause{}
	v := p.Strings()
	p.value.Set("a")
	p.value.Set("b")
	assert.Equal(t, []string{"a", "b"}, *v)
}

func TestStringsStringer(t *testing.T) {
	target := []string{}
	v := newAccumulator(&target, func(v interface{}) Value { return newStringValue(v.(*string)) })
	v.Set("hello")
	v.Set("world")
	assert.Equal(t, "hello,world", v.String())
}

func TestParseStringMap(t *testing.T) {
	p := Clause{}
	v := p.StringMap()
	p.value.Set("a:b")
	p.value.Set("b:c")
	assert.Equal(t, map[string]string{"a": "b", "b": "c"}, *v)
}

func TestParseURL(t *testing.T) {
	p := Clause{}
	v := p.URL()
	p.value.Set("http://w3.org")
	u, err := url.Parse("http://w3.org")
	assert.NoError(t, err)
	assert.Equal(t, *u, **v)
}

func TestParseExistingFile(t *testing.T) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	defer os.Remove(f.Name())

	p := Clause{}
	v := p.ExistingFile()
	err = p.value.Set(f.Name())
	assert.NoError(t, err)
	assert.Equal(t, f.Name(), *v)
	err = p.value.Set("/etc/hostsDEFINITELYMISSING")
	assert.Error(t, err)
}

func TestFloat32(t *testing.T) {
	p := Clause{}
	v := p.Float32()
	err := p.value.Set("123.45")
	assert.NoError(t, err)
	assert.InEpsilon(t, 123.45, *v, 0.001)
}

func TestUnicodeShortFlag(t *testing.T) {
	app := newTestApp()
	f := app.Flag("long", "").Short('ä').Bool()
	_, err := app.Parse([]string{"-ä"})
	assert.NoError(t, err)
	assert.True(t, *f)
}

type TestResolver struct {
	vals map[string]string
}

func (r *TestResolver) Resolve(key string, context *ParseContext) string {
	return r.vals[key]
}

func TestResolverSimple(t *testing.T) {
	app := newTestApp()
	app.ConfigResolver(&TestResolver{vals: map[string]string{"hello": "world"}})
	f := app.Flag("hello", "help").String()
	_, err := app.Parse([]string{})
	assert.NoError(t, err)
	assert.Equal(t, "world", *f)
}

func TestResolverSatisfiesRequired(t *testing.T) {
	app := newTestApp()
	app.ConfigResolver(&TestResolver{vals: map[string]string{"hello": "world"}})
	f := app.Flag("hello", "help").Required().String()
	_, err := app.Parse([]string{})
	assert.NoError(t, err)
	assert.Equal(t, "world", *f)
}

func TestResolverKeyOverride(t *testing.T) {
	app := newTestApp()
	app.ConfigResolver(&TestResolver{vals: map[string]string{"foo": "world"}})
	f := app.Flag("hello", "help").ConfigResolverKey("foo").String()
	_, err := app.Parse([]string{})
	assert.NoError(t, err)
	assert.Equal(t, "world", *f)
}

func TestResolverDisable(t *testing.T) {
	app := newTestApp()
	app.ConfigResolver(&TestResolver{vals: map[string]string{"hello": "world"}})
	f := app.Flag("hello", "help").NoConfigResolver().String()
	_, err := app.Parse([]string{})
	assert.NoError(t, err)
	assert.Equal(t, "", *f)
}

func TestResolverLowerPriorityThanFlag(t *testing.T) {
	app := newTestApp()
	app.ConfigResolver(&TestResolver{vals: map[string]string{"hello": "world"}})
	f := app.Flag("hello", "help").String()
	_, err := app.Parse([]string{"--hello", "there"})
	assert.NoError(t, err)
	assert.Equal(t, "there", *f)
}

func TestResolverLowerPriorityThanEnvar(t *testing.T) {
	os.Setenv("TEST_RESOLVER", "foo")
	app := newTestApp()
	app.ConfigResolver(&TestResolver{vals: map[string]string{"hello": "world"}})
	f := app.Flag("hello", "help").Envar("TEST_RESOLVER").String()
	_, err := app.Parse([]string{})
	assert.NoError(t, err)
	assert.Equal(t, "foo", *f)
}

func TestResolverFallbackWithMultipleResolvers(t *testing.T) {
	app := newTestApp()
	app.ConfigResolver(&TestResolver{vals: map[string]string{"hello": "world"}})
	app.ConfigResolver(&TestResolver{vals: map[string]string{"hello": "there", "foo": "bar"}})
	f1 := app.Flag("hello", "help").String()
	f2 := app.Flag("foo", "help").String()
	_, err := app.Parse([]string{})
	assert.NoError(t, err)
	assert.Equal(t, "world", *f1)
	assert.Equal(t, "bar", *f2)
}
