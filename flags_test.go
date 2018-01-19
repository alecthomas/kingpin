package kingpin

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBool(t *testing.T) {
	app := newTestApp()
	b := app.Flag("b", "").Bool()
	_, err := app.Parse([]string{"--b"})
	assert.NoError(t, err)
	assert.True(t, *b)
}

func TestNegatableBool(t *testing.T) {
	app := newTestApp()
	f := app.Flag("b", "").Default("true")
	b := f.NegatableBool()
	_, err := app.Parse([]string{})
	assert.NoError(t, err)
	assert.True(t, *b)
	_, err = app.Parse([]string{"--no-b"})
	assert.NoError(t, err)
	assert.False(t, *b)
}

func TestNegateNonBool(t *testing.T) {
	app := newTestApp()
	f := app.Flag("b", "")
	f.Int()
	_, err := app.Parse([]string{"--no-b"})
	assert.Error(t, err)
}

func TestNegativePrefixLongFlag(t *testing.T) {
	app := newTestApp()
	f := app.Flag("no-comment", "")
	b := f.Bool()
	_, err := app.Parse([]string{})
	assert.NoError(t, err)
	assert.False(t, *b)
	_, err = app.Parse([]string{"--no-comment"})
	assert.NoError(t, err)
	assert.False(t, *b)
}

func TestNonNegatableBoolFlagCanNotBeNegated(t *testing.T) {
	app := newTestApp()
	nonNegatable := app.Flag("nonneg", "").Bool()

	_, err := app.Parse([]string{})
	assert.NoError(t, err)
	assert.False(t, *nonNegatable)

	_, err = app.Parse([]string{"--nonneg"})
	assert.NoError(t, err)
	assert.True(t, *nonNegatable)

	_, err = app.Parse([]string{"--no-nonneg"})
	assert.Error(t, err)
}

func TestHelpForBoolFlags(t *testing.T) {
	app := newTestApp()
	app.Flag("yes", "").Default("true").NegatableBool()
	app.Flag("no", "").Default("false").NegatableBool()
	app.Flag("nonneg", "").Bool()

	w := bytes.NewBuffer(nil)
	app.Writers(w, w).Usage(nil)
	assert.Contains(t, w.String(), "--[no-]yes")
	assert.Contains(t, w.String(), "--[no-]no")
	assert.Contains(t, w.String(), "--nonneg")
}

func TestInvalidFlagDefaultCanBeOverridden(t *testing.T) {
	app := newTestApp()
	app.Flag("a", "").Default("invalid").Bool()
	_, err := app.Parse([]string{})
	assert.Error(t, err)
}

func TestRequiredFlag(t *testing.T) {
	app := newTestApp().Version("0.0.0")
	exits := 0
	app.Terminate(func(int) { exits++ })
	app.Flag("a", "").Required().Bool()
	_, err := app.Parse([]string{"--a"})
	assert.NoError(t, err)
	_, err = app.Parse([]string{})
	assert.Error(t, err)
	// This will error because the Terminate() function doesn't actually terminate.
	_, _ = app.Parse([]string{"--version"})
	assert.Equal(t, 1, exits)
}

func TestShortFlag(t *testing.T) {
	app := newTestApp()
	f := app.Flag("long", "").Short('s').Bool()
	_, err := app.Parse([]string{"-s"})
	assert.NoError(t, err)
	assert.True(t, *f)
}

func TestCombinedShortFlags(t *testing.T) {
	app := newTestApp()
	a := app.Flag("short0", "").Short('0').Bool()
	b := app.Flag("short1", "").Short('1').Bool()
	c := app.Flag("short2", "").Short('2').Bool()
	_, err := app.Parse([]string{"-01"})
	assert.NoError(t, err)
	assert.True(t, *a)
	assert.True(t, *b)
	assert.False(t, *c)
}

func TestCombinedShortFlagArg(t *testing.T) {
	a := newTestApp()
	n := a.Flag("short", "").Short('s').Int()
	_, err := a.Parse([]string{"-s10"})
	assert.NoError(t, err)
	assert.Equal(t, 10, *n)
}

func TestEmptyShortFlagIsAnError(t *testing.T) {
	_, err := newTestApp().Parse([]string{"-"})
	assert.Error(t, err)
}

func TestRequiredWithEnvarMissingErrors(t *testing.T) {
	app := newTestApp()
	app.Flag("t", "").Envar("TEST_ENVAR").Required().Int()
	_, err := app.Parse([]string{})
	assert.Error(t, err)
}

func TestRequiredWithEnvar(t *testing.T) {
	os.Setenv("TEST_ENVAR", "123")
	app := newTestApp()
	flag := app.Flag("t", "").Envar("TEST_ENVAR").Required().Int()
	_, err := app.Parse([]string{})
	assert.NoError(t, err)
	assert.Equal(t, 123, *flag)
}

func TestSubcommandFlagRequiredWithEnvar(t *testing.T) {
	os.Setenv("TEST_ENVAR", "123")
	app := newTestApp()
	cmd := app.Command("command", "")
	flag := cmd.Flag("t", "").Envar("TEST_ENVAR").Required().Int()
	_, err := app.Parse([]string{"command"})
	assert.NoError(t, err)
	assert.Equal(t, 123, *flag)
}

func TestRegexp(t *testing.T) {
	app := newTestApp()
	flag := app.Flag("reg", "").Regexp()
	_, err := app.Parse([]string{"--reg", "^abc$"})
	assert.NoError(t, err)
	assert.NotNil(t, *flag)
	assert.Equal(t, "^abc$", (*flag).String())
	assert.Regexp(t, *flag, "abc")
	assert.NotRegexp(t, *flag, "abcd")
}

func TestDuplicateShortFlag(t *testing.T) {
	app := newTestApp()
	app.Flag("a", "").Short('a').String()
	app.Flag("b", "").Short('a').String()
	_, err := app.Parse([]string{})
	assert.Error(t, err)
}

func TestDuplicateLongFlag(t *testing.T) {
	app := newTestApp()
	app.Flag("a", "").String()
	app.Flag("a", "").String()
	_, err := app.Parse([]string{})
	assert.Error(t, err)
}

func TestGetFlagAndOverrideDefault(t *testing.T) {
	app := newTestApp()
	a := app.Flag("a", "").Default("default").String()
	_, err := app.Parse([]string{})
	assert.NoError(t, err)
	assert.Equal(t, "default", *a)
	app.GetFlag("a").Default("new")
	_, err = app.Parse([]string{})
	assert.NoError(t, err)
	assert.Equal(t, "new", *a)
}

func TestEnvarOverrideDefault(t *testing.T) {
	os.Setenv("TEST_ENVAR", "123")
	app := newTestApp()
	flag := app.Flag("t", "").Default("default").Envar("TEST_ENVAR").String()
	_, err := app.Parse([]string{})
	assert.NoError(t, err)
	assert.Equal(t, "123", *flag)
}

func TestFlagMultipleValuesDefault(t *testing.T) {
	app := newTestApp()
	a := app.Flag("a", "").Default("default1", "default2").Strings()
	_, err := app.Parse([]string{})
	assert.NoError(t, err)
	assert.Equal(t, []string{"default1", "default2"}, *a)
}

func TestFlagMultipleValuesDefaultNonRepeatable(t *testing.T) {
	c := newTestApp()
	c.Flag("foo", "foo").Default("a", "b").String()
	_, err := c.Parse([]string{})
	assert.Error(t, err)
}

func TestFlagMultipleValuesDefaultEnvarUnix(t *testing.T) {
	app := newTestApp()
	a := app.Flag("a", "").Envar("TEST_MULTIPLE_VALUES").Strings()
	os.Setenv("TEST_MULTIPLE_VALUES", "123:456")
	_, err := app.Parse([]string{})
	assert.NoError(t, err)
	assert.Equal(t, []string{"123", "456"}, *a)
}

func TestFlagMultipleValuesDefaultEnvarNonRepeatable(t *testing.T) {
	c := newTestApp()
	a := c.Flag("foo", "foo").Envar("TEST_MULTIPLE_VALUES_NON_REPEATABLE").String()
	os.Setenv("TEST_MULTIPLE_VALUES_NON_REPEATABLE", "123\n456")
	_, err := c.Parse([]string{})
	assert.NoError(t, err)
	assert.Equal(t, "123\n456", *a)
}

func TestFlagHintAction(t *testing.T) {
	c := newTestApp()

	action := func() []string {
		return []string{"opt1", "opt2"}
	}

	a := c.Flag("foo", "foo").HintAction(action)
	args := a.resolveCompletion()
	assert.False(t, args.Directories)
	assert.False(t, args.Files)
	assert.Equal(t, []string{"opt1", "opt2"}, args.resolveWords())
}

func TestFlagHintOptions(t *testing.T) {
	c := newTestApp()

	a := c.Flag("foo", "foo").HintOptions("opt1", "opt2")
	args := a.resolveCompletion()
	assert.False(t, args.Directories)
	assert.False(t, args.Files)
	assert.Equal(t, []string{"opt1", "opt2"}, args.resolveWords())
}

func TestFlagEnumVar(t *testing.T) {
	c := newTestApp()
	var bar string

	a := c.Flag("foo", "foo")
	a.Enum("opt1", "opt2")
	b := c.Flag("bar", "bar")
	b.EnumVar(&bar, "opt3", "opt4")

	args := a.resolveCompletion()
	assert.False(t, args.Directories)
	assert.False(t, args.Files)
	assert.Equal(t, []string{"opt1", "opt2"}, args.resolveWords())

	args = b.resolveCompletion()
	assert.False(t, args.Directories)
	assert.False(t, args.Files)
	assert.Equal(t, []string{"opt3", "opt4"}, args.resolveWords())
}

func TestMultiHintOptions(t *testing.T) {
	c := newTestApp()

	a := c.Flag("foo", "foo").HintOptions("opt1").HintOptions("opt2")
	args := a.resolveCompletion()
	assert.False(t, args.Directories)
	assert.False(t, args.Files)
	assert.Equal(t, []string{"opt1", "opt2"}, args.resolveWords())
}
func TestMultiHintActions(t *testing.T) {
	c := newTestApp()

	a := c.Flag("foo", "foo").
		HintAction(func() []string {
			return []string{"opt1"}
		}).
		HintAction(func() []string {
			return []string{"opt2"}
		})
	args := a.resolveCompletion()
	assert.False(t, args.Directories)
	assert.False(t, args.Files)
	assert.Equal(t, []string{"opt1", "opt2"}, args.resolveWords())
}

func TestCombinationHintActionsOptions(t *testing.T) {
	c := newTestApp()

	a := c.Flag("foo", "foo").HintAction(func() []string {
		return []string{"opt1"}
	}).HintOptions("opt2")
	args := a.resolveCompletion()
	assert.False(t, args.Directories)
	assert.False(t, args.Files)
	assert.Equal(t, []string{"opt1", "opt2"}, args.resolveWords())
}

func TestCombinationEnumActions(t *testing.T) {
	c := newTestApp()
	var foo string

	a := c.Flag("foo", "foo").
		HintAction(func() []string {
			return []string{"opt1", "opt2"}
		})
	a.Enum("opt3", "opt4")

	b := c.Flag("bar", "bar").
		HintAction(func() []string {
			return []string{"opt5", "opt6"}
		})
	b.EnumVar(&foo, "opt3", "opt4")

	// Provided HintActions should override automatically generated Enum options.
	args := a.resolveCompletion()
	assert.False(t, args.Directories)
	assert.False(t, args.Files)
	assert.Equal(t, []string{"opt1", "opt2"}, args.resolveWords())

	args = b.resolveCompletion()
	assert.False(t, args.Directories)
	assert.False(t, args.Files)
	assert.Equal(t, []string{"opt5", "opt6"}, args.resolveWords())
}

func TestCombinationEnumOptions(t *testing.T) {
	c := newTestApp()
	var foo string

	a := c.Flag("foo", "foo").HintOptions("opt1", "opt2")
	a.Enum("opt3", "opt4")

	b := c.Flag("bar", "bar").HintOptions("opt5", "opt6")
	b.EnumVar(&foo, "opt3", "opt4")

	// Provided HintOptions should override automatically generated Enum options.
	args := a.resolveCompletion()
	assert.False(t, args.Directories)
	assert.False(t, args.Files)
	assert.Equal(t, []string{"opt1", "opt2"}, args.resolveWords())

	args = b.resolveCompletion()
	assert.False(t, args.Directories)
	assert.False(t, args.Files)
	assert.Equal(t, []string{"opt5", "opt6"}, args.resolveWords())
}
