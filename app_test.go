package kingpin

import (
	"errors"
	"io/ioutil"

	"github.com/stretchr/testify/assert"

	"sort"
	"strings"
	"testing"
	"time"
)

func newTestApp() *Application {
	return New("test", "").Terminate(nil)
}

func TestCommander(t *testing.T) {
	c := newTestApp()
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
	c := newTestApp()
	c.Flag("a", "a").String()
	c.Flag("b", "b").Required().String()

	_, err := c.Parse([]string{"--a=foo"})
	assert.Error(t, err)
	_, err = c.Parse([]string{"--b=foo"})
	assert.NoError(t, err)
}

func TestRepeatableFlags(t *testing.T) {
	c := newTestApp()
	c.Flag("a", "a").String()
	c.Flag("b", "b").Strings()
	_, err := c.Parse([]string{"--a=foo", "--a=bar"})
	assert.Error(t, err)
	_, err = c.Parse([]string{"--b=foo", "--b=bar"})
	assert.NoError(t, err)
}

func TestInvalidDefaultFlagValueErrors(t *testing.T) {
	c := newTestApp()
	c.Flag("foo", "foo").Default("a").Int()
	_, err := c.Parse([]string{})
	assert.Error(t, err)
}

func TestInvalidDefaultArgValueErrors(t *testing.T) {
	c := newTestApp()
	cmd := c.Command("cmd", "cmd")
	cmd.Arg("arg", "arg").Default("one").Int()
	_, err := c.Parse([]string{"cmd"})
	assert.Error(t, err)
}

func TestArgsRequiredAfterNonRequiredErrors(t *testing.T) {
	c := newTestApp()
	cmd := c.Command("cmd", "")
	cmd.Arg("a", "a").String()
	cmd.Arg("b", "b").Required().String()
	_, err := c.Parse([]string{"cmd"})
	assert.Error(t, err)
}

func TestArgsMultipleRequiredThenNonRequired(t *testing.T) {
	c := newTestApp().Writer(ioutil.Discard)
	cmd := c.Command("cmd", "")
	cmd.Arg("a", "a").Required().String()
	cmd.Arg("b", "b").Required().String()
	cmd.Arg("c", "c").String()
	cmd.Arg("d", "d").String()
	_, err := c.Parse([]string{"cmd", "a", "b"})
	assert.NoError(t, err)
	_, err = c.Parse([]string{})
	assert.Error(t, err)
}

func TestDispatchCallbackIsCalled(t *testing.T) {
	dispatched := false
	c := newTestApp()
	c.Command("cmd", "").Action(func(*ParseContext) error {
		dispatched = true
		return nil
	})

	_, err := c.Parse([]string{"cmd"})
	assert.NoError(t, err)
	assert.True(t, dispatched)
}

func TestTopLevelArgWorks(t *testing.T) {
	c := newTestApp()
	s := c.Arg("arg", "help").String()
	_, err := c.Parse([]string{"foo"})
	assert.NoError(t, err)
	assert.Equal(t, "foo", *s)
}

func TestTopLevelArgCantBeUsedWithCommands(t *testing.T) {
	c := newTestApp()
	c.Arg("arg", "help").String()
	c.Command("cmd", "help")
	_, err := c.Parse([]string{})
	assert.Error(t, err)
}

func TestTooManyArgs(t *testing.T) {
	a := newTestApp()
	a.Arg("a", "").String()
	_, err := a.Parse([]string{"a", "b"})
	assert.Error(t, err)
}

func TestTooManyArgsAfterCommand(t *testing.T) {
	a := newTestApp()
	a.Command("a", "")
	assert.NoError(t, a.init())
	_, err := a.Parse([]string{"a", "b"})
	assert.Error(t, err)
}

func TestArgsLooksLikeFlagsWithConsumeRemainder(t *testing.T) {
	a := newTestApp()
	a.Arg("opts", "").Required().Strings()
	_, err := a.Parse([]string{"hello", "-world"})
	assert.Error(t, err)
}

func TestCommandParseDoesNotResetFlagsToDefault(t *testing.T) {
	app := newTestApp()
	flag := app.Flag("flag", "").Default("default").String()
	app.Command("cmd", "")

	_, err := app.Parse([]string{"--flag=123", "cmd"})
	assert.NoError(t, err)
	assert.Equal(t, "123", *flag)
}

func TestCommandParseDoesNotFailRequired(t *testing.T) {
	app := newTestApp()
	flag := app.Flag("flag", "").Required().String()
	app.Command("cmd", "")

	_, err := app.Parse([]string{"cmd", "--flag=123"})
	assert.NoError(t, err)
	assert.Equal(t, "123", *flag)
}

func TestSelectedCommand(t *testing.T) {
	app := newTestApp()
	c0 := app.Command("c0", "")
	c0.Command("c1", "")
	s, err := app.Parse([]string{"c0", "c1"})
	assert.NoError(t, err)
	assert.Equal(t, "c0 c1", s)
}

func TestSubCommandRequired(t *testing.T) {
	app := newTestApp()
	c0 := app.Command("c0", "")
	c0.Command("c1", "")
	_, err := app.Parse([]string{"c0"})
	assert.Error(t, err)
}

func TestInterspersedFalse(t *testing.T) {
	app := newTestApp().Interspersed(false)
	a1 := app.Arg("a1", "").String()
	a2 := app.Arg("a2", "").String()
	f1 := app.Flag("flag", "").String()

	_, err := app.Parse([]string{"a1", "--flag=flag"})
	assert.NoError(t, err)
	assert.Equal(t, "a1", *a1)
	assert.Equal(t, "--flag=flag", *a2)
	assert.Equal(t, "", *f1)
}

func TestInterspersedTrue(t *testing.T) {
	// test once with the default value and once with explicit true
	for i := 0; i < 2; i++ {
		app := newTestApp()
		if i != 0 {
			t.Log("Setting explicit")
			app.Interspersed(true)
		} else {
			t.Log("Using default")
		}
		a1 := app.Arg("a1", "").String()
		a2 := app.Arg("a2", "").String()
		f1 := app.Flag("flag", "").String()

		_, err := app.Parse([]string{"a1", "--flag=flag"})
		assert.NoError(t, err)
		assert.Equal(t, "a1", *a1)
		assert.Equal(t, "", *a2)
		assert.Equal(t, "flag", *f1)
	}
}

func TestDefaultEnvars(t *testing.T) {
	a := New("some-app", "").Terminate(nil).DefaultEnvars()
	f0 := a.Flag("some-flag", "")
	f0.Bool()
	f1 := a.Flag("some-other-flag", "").NoEnvar()
	f1.Bool()
	f2 := a.Flag("a-1-flag", "")
	f2.Bool()
	_, err := a.Parse([]string{})
	assert.NoError(t, err)
	assert.Equal(t, "SOME_APP_SOME_FLAG", f0.getEnvar())
	assert.Equal(t, "", f1.getEnvar())
	assert.Equal(t, "SOME_APP_A_1_FLAG", f2.getEnvar())
}

func TestEnvarNamePrefixOnCommands(t *testing.T) {
	app := New("some-app", "").Terminate(nil).EnvarNamePrefix("FOO_")

	c1 := app.Command("c1", "").EnvarNamePrefix("C1_")
	f11 := c1.Flag("f11", "").Envar("F11")
	f12 := c1.Flag("f12", "")
	c11 := c1.Command("c11", "").EnvarNamePrefix("C11_")
	f111 := c11.Flag("f111", "").Envar("F111")
	f112 := c11.Flag("f112", "")
	c12 := c1.Command("c12", "")
	f121 := c12.Flag("f121", "").Envar("F121")
	f122 := c12.Flag("f122", "")

	c2 := app.Command("c2", "")
	f21 := c2.Flag("f21", "").Envar("F21")
	f22 := c2.Flag("f22", "")
	c21 := c2.Command("c21", "").EnvarNamePrefix("C21_")
	f211 := c21.Flag("f211", "").Envar("F211")
	f212 := c21.Flag("f212", "")
	c22 := c2.Command("c22", "")
	f221 := c22.Flag("f221", "").Envar("F221")
	f222 := c22.Flag("f222", "")

	for _, f := range []*FlagClause{f11, f12, f111, f112, f121, f122, f21, f22, f211, f212, f221, f222} {
		f.Bool()
	}

	_, err := app.Parse([]string{"c1", "c11"})
	assert.NoError(t, err)
	assert.Equal(t, "FOO_C1_F11", f11.getEnvar())
	assert.Equal(t, "", f12.getEnvar())
	assert.Equal(t, "FOO_C1_C11_F111", f111.getEnvar())
	assert.Equal(t, "", f112.getEnvar())
	assert.Equal(t, "FOO_C1_F121", f121.getEnvar())
	assert.Equal(t, "", f122.getEnvar())

	assert.Equal(t, "FOO_F21", f21.getEnvar())
	assert.Equal(t, "", f22.getEnvar())
	assert.Equal(t, "FOO_C21_F211", f211.getEnvar())
	assert.Equal(t, "", f212.getEnvar())
	assert.Equal(t, "FOO_F221", f221.getEnvar())
	assert.Equal(t, "", f222.getEnvar())
}

func TestEnvarNamePrefixOnFlagGroups(t *testing.T) {
	app := New("some-app", "").Terminate(nil).EnvarNamePrefix("FOO_")

	g1 := app.FlagGroup("g1").EnvarNamePrefix("G1_")
	f11 := g1.Flag("f11", "").Envar("F11")
	f12 := g1.Flag("f12", "")
	g11 := g1.FlagGroup("g11").EnvarNamePrefix("G11_")
	f111 := g11.Flag("f111", "").Envar("F111")
	f112 := g11.Flag("f112", "")
	g12 := g1.FlagGroup("g12")
	f121 := g12.Flag("f121", "").Envar("F121")
	f122 := g12.Flag("f122", "")

	g2 := app.FlagGroup("g2")
	f21 := g2.Flag("f21", "").Envar("F21")
	f22 := g2.Flag("f22", "")
	g21 := g2.FlagGroup("g21").EnvarNamePrefix("G21_")
	f211 := g21.Flag("f211", "").Envar("F211")
	f212 := g21.Flag("f212", "")
	g22 := g2.FlagGroup("g22")
	f221 := g22.Flag("f221", "").Envar("F221")
	f222 := g22.Flag("f222", "")

	for _, f := range []*FlagClause{f11, f12, f111, f112, f121, f122, f21, f22, f211, f212, f221, f222} {
		f.Bool()
	}

	_, err := app.Parse([]string{})
	assert.NoError(t, err)
	assert.Equal(t, "FOO_G1_F11", f11.getEnvar())
	assert.Equal(t, "", f12.getEnvar())
	assert.Equal(t, "FOO_G1_G11_F111", f111.getEnvar())
	assert.Equal(t, "", f112.getEnvar())
	assert.Equal(t, "FOO_G1_F121", f121.getEnvar())
	assert.Equal(t, "", f122.getEnvar())

	assert.Equal(t, "FOO_F21", f21.getEnvar())
	assert.Equal(t, "", f22.getEnvar())
	assert.Equal(t, "FOO_G21_F211", f211.getEnvar())
	assert.Equal(t, "", f212.getEnvar())
	assert.Equal(t, "FOO_F221", f221.getEnvar())
	assert.Equal(t, "", f222.getEnvar())
}

func TestBashCompletionOptionsWithEmptyApp(t *testing.T) {
	a := newTestApp()
	context, err := a.ParseContext([]string{"--completion-bash"})
	if err != nil {
		t.Errorf("Unexpected error whilst parsing context: [%v]", err)
	}
	args := a.completionOptions(context)
	assert.Equal(t, []string(nil), args)
}

func TestBashCompletionOptions(t *testing.T) {
	a := newTestApp()
	a.Command("one", "")
	a.Flag("flag-0", "").String()
	a.Flag("flag-1", "").HintOptions("opt1", "opt2", "opt3").String()

	two := a.Command("two", "")
	two.Flag("flag-2", "").String()
	two.Flag("flag-3", "").HintOptions("opt4", "opt5", "opt6").String()

	three := a.Command("three", "")
	three.Flag("flag-4", "").String()
	three.Arg("arg-1", "").String()
	three.Arg("arg-2", "").HintOptions("arg-2-opt-1", "arg-2-opt-2").String()
	three.Arg("arg-3", "").String()
	three.Arg("arg-4", "").HintAction(func() []string {
		return []string{"arg-4-opt-1", "arg-4-opt-2"}
	}).String()

	cases := []struct {
		Args            string
		ExpectedOptions []string
	}{
		{
			Args:            "--completion-bash",
			ExpectedOptions: []string{"help", "one", "three", "two"},
		},
		{
			Args:            "--completion-bash --",
			ExpectedOptions: []string{"--flag-0", "--flag-1", "--help"},
		},
		{
			Args:            "--completion-bash --fla",
			ExpectedOptions: []string{"--flag-0", "--flag-1", "--help"},
		},
		{
			// No options available for flag-0, return to cmd completion
			Args:            "--completion-bash --flag-0",
			ExpectedOptions: []string{"help", "one", "three", "two"},
		},
		{
			Args:            "--completion-bash --flag-0 --",
			ExpectedOptions: []string{"--flag-0", "--flag-1", "--help"},
		},
		{
			Args:            "--completion-bash --flag-1",
			ExpectedOptions: []string{"opt1", "opt2", "opt3"},
		},
		{
			Args:            "--completion-bash --flag-1 opt",
			ExpectedOptions: []string{"opt1", "opt2", "opt3"},
		},
		{
			Args:            "--completion-bash --flag-1 opt1",
			ExpectedOptions: []string{"help", "one", "three", "two"},
		},
		{
			Args:            "--completion-bash --flag-1 opt1 --",
			ExpectedOptions: []string{"--flag-0", "--flag-1", "--help"},
		},

		// Try Subcommand
		{
			Args:            "--completion-bash two",
			ExpectedOptions: []string(nil),
		},
		{
			Args:            "--completion-bash two --",
			ExpectedOptions: []string{"--help", "--flag-2", "--flag-3", "--flag-0", "--flag-1"},
		},
		{
			Args:            "--completion-bash two --flag",
			ExpectedOptions: []string{"--help", "--flag-2", "--flag-3", "--flag-0", "--flag-1"},
		},
		{
			Args:            "--completion-bash two --flag-2",
			ExpectedOptions: []string(nil),
		},
		{
			// Top level flags carry downwards
			Args:            "--completion-bash two --flag-1",
			ExpectedOptions: []string{"opt1", "opt2", "opt3"},
		},
		{
			// Top level flags carry downwards
			Args:            "--completion-bash two --flag-1 opt",
			ExpectedOptions: []string{"opt1", "opt2", "opt3"},
		},
		{
			// Top level flags carry downwards
			Args:            "--completion-bash two --flag-1 opt1",
			ExpectedOptions: []string(nil),
		},
		{
			Args:            "--completion-bash two --flag-3",
			ExpectedOptions: []string{"opt4", "opt5", "opt6"},
		},
		{
			Args:            "--completion-bash two --flag-3 opt",
			ExpectedOptions: []string{"opt4", "opt5", "opt6"},
		},
		{
			Args:            "--completion-bash two --flag-3 opt4",
			ExpectedOptions: []string(nil),
		},
		{
			Args:            "--completion-bash two --flag-3 opt4 --",
			ExpectedOptions: []string{"--help", "--flag-2", "--flag-3", "--flag-0", "--flag-1"},
		},

		// Args complete
		{
			// After a command with an arg with no options, nothing should be
			// shown
			Args:            "--completion-bash three ",
			ExpectedOptions: []string(nil),
		},
		{
			// After a command with an arg, explicitly starting a flag should
			// complete flags
			Args:            "--completion-bash three --",
			ExpectedOptions: []string{"--flag-0", "--flag-1", "--flag-4", "--help"},
		},
		{
			// After a command with an arg that does have completions, they
			// should be shown
			Args:            "--completion-bash three arg1 ",
			ExpectedOptions: []string{"arg-2-opt-1", "arg-2-opt-2"},
		},
		{
			// After a command with an arg that does have completions, but a
			// flag is started, flag options should be completed
			Args:            "--completion-bash three arg1 --",
			ExpectedOptions: []string{"--flag-0", "--flag-1", "--flag-4", "--help"},
		},
		{
			// After a command with an arg that has no completions, and isn't first,
			// nothing should be shown
			Args:            "--completion-bash three arg1 arg2 ",
			ExpectedOptions: []string(nil),
		},
		{
			// After a command with a different arg that also has completions,
			// those different options should be shown
			Args:            "--completion-bash three arg1 arg2 arg3 ",
			ExpectedOptions: []string{"arg-4-opt-1", "arg-4-opt-2"},
		},
		{
			// After a command with all args listed, nothing should complete
			Args:            "--completion-bash three arg1 arg2 arg3 arg4",
			ExpectedOptions: []string(nil),
		},
		{
			// After a -- argument, no more flags should be suggested
			Args:            "--completion-bash three --flag-0 -- --",
			ExpectedOptions: []string(nil),
		},
		{
			// After a -- argument, argument options should still be suggested
			Args:            "--completion-bash three -- arg1 ",
			ExpectedOptions: []string{"arg-2-opt-1", "arg-2-opt-2"},
		},
	}

	for _, c := range cases {
		context, _ := a.ParseContext(strings.Split(c.Args, " "))
		args := a.completionOptions(context)

		sort.Strings(args)
		sort.Strings(c.ExpectedOptions)

		assert.Equal(t, c.ExpectedOptions, args, "Expected != Actual: [%v] != [%v]. \nInput was: [%v]", c.ExpectedOptions, args, c.Args)
	}

}

func TestCmdValidation(t *testing.T) {
	c := newTestApp()
	cmd := c.Command("cmd", "")

	var a, b string
	cmd.Flag("a", "a").StringVar(&a)
	cmd.Flag("b", "b").StringVar(&b)
	cmd.Validate(func(*CmdClause) error {
		if a == "" && b == "" {
			return errors.New("must specify either a or b")
		}
		return nil
	})

	_, err := c.Parse([]string{"cmd"})
	assert.Error(t, err)

	_, err = c.Parse([]string{"cmd", "--a", "A"})
	assert.NoError(t, err)
}

func TestSubFlagGroup(t *testing.T) {
	app := newTestApp()
	f1 := app.Flag("f1", "")

	g1 := app.FlagGroup("g1.")
	f11 := g1.Flag("f11", "")

	g11 := g1.FlagGroup("g11.")
	f111 := g11.Flag("f111", "").Required()

	for _, f := range []*FlagClause{f1, f11, f111} {
		f.String()
	}

	context, err := app.ParseContext([]string{})
	assert.NoError(t, err)

	assert.Equal(t, f1, context.flags.long["f1"])
	assert.Equal(t, f11, context.flags.long["g1.f11"])
	assert.Equal(t, f111, context.flags.long["g1.g11.f111"])

	_, err = app.Parse([]string{})

	assert.EqualError(t, err, "required flag --g1.g11.f111 not provided")
}

func TestFlagsOf(t *testing.T) {
	app := newTestApp().FlagsOf(FlagRegistrarFunc(func(fg FlagGroup) {
		fg.Flag("app1", "").Bool()
		fg.Flag("app2", "").Bool()
	}), FlagRegistrarFunc(func(fg FlagGroup) {
		fg.Flag("app3", "").Bool()
		fg.Flag("app4", "").Bool()
	}))

	cmd1 := app.Command("cmd1", "").FlagsOf(FlagRegistrarFunc(func(fg FlagGroup) {
		fg.Flag("cmd11", "").Bool()
		fg.Flag("cmd12", "").Bool()
	}), FlagRegistrarFunc(func(fg FlagGroup) {
		fg.Flag("cmd13", "").Bool()
		fg.Flag("cmd14", "").Bool()
	}))

	cmd11 := cmd1.Command("cmd11", "").FlagsOf(FlagRegistrarFunc(func(fg FlagGroup) {
		fg.Flag("cmd111", "").Bool()
		fg.Flag("cmd112", "").Bool()
	}), FlagRegistrarFunc(func(fg FlagGroup) {
		fg.Flag("cmd113", "").Bool()
		fg.Flag("cmd114", "").Bool()
	}))

	_, err := app.Parse([]string{"cmd1", "cmd11"})
	assert.NoError(t, err)
	assert.IsType(t, &boolValue{}, app.GetFlag("app1").value)
	assert.IsType(t, &boolValue{}, app.GetFlag("app2").value)
	assert.IsType(t, &boolValue{}, app.GetFlag("app3").value)
	assert.IsType(t, &boolValue{}, app.GetFlag("app4").value)

	assert.IsType(t, &boolValue{}, cmd1.GetFlag("cmd11").value)
	assert.IsType(t, &boolValue{}, cmd1.GetFlag("cmd12").value)
	assert.IsType(t, &boolValue{}, cmd1.GetFlag("cmd13").value)
	assert.IsType(t, &boolValue{}, cmd1.GetFlag("cmd14").value)

	assert.IsType(t, &boolValue{}, cmd11.GetFlag("cmd111").value)
	assert.IsType(t, &boolValue{}, cmd11.GetFlag("cmd112").value)
	assert.IsType(t, &boolValue{}, cmd11.GetFlag("cmd113").value)
	assert.IsType(t, &boolValue{}, cmd11.GetFlag("cmd114").value)
}

func TestCmdsOf(t *testing.T) {
	app := newTestApp().CmdsOf(CmdRegistrarFunc(func(c Cmd) {
		c.Command("app-1", "h app-1")
		c.Command("app-2", "h app-2")
	}), CmdRegistrarFunc(func(c Cmd) {
		c.Command("app-3", "h app-3")
		c.Command("app-4", "h app-4")
	}))

	cmd1 := app.Command("cmd1", "").CmdsOf(CmdRegistrarFunc(func(c Cmd) {
		c.Command("cmd1-1", "h cmd1-1")
		c.Command("cmd1-2", "h cmd1-2")
	}), CmdRegistrarFunc(func(c Cmd) {
		c.Command("cmd1-3", "h cmd1-3")
		c.Command("cmd1-4", "h cmd1-4")
	}))

	cmd11 := cmd1.Command("cmd11", "").CmdsOf(CmdRegistrarFunc(func(c Cmd) {
		c.Command("cmd11-1", "h cmd11-1")
		c.Command("cmd11-2", "h cmd11-2")
	}), CmdRegistrarFunc(func(c Cmd) {
		c.Command("cmd11-3", "h cmd11-3")
		c.Command("cmd11-4", "h cmd11-4")
	}))

	_, err := app.Parse([]string{"cmd1", "cmd11", "cmd11-1"})
	assert.NoError(t, err)
	assert.Equal(t, "h app-1", app.GetCommand("app-1").help)
	assert.Equal(t, "h app-2", app.GetCommand("app-2").help)
	assert.Equal(t, "h app-3", app.GetCommand("app-3").help)
	assert.Equal(t, "h app-4", app.GetCommand("app-4").help)

	assert.Equal(t, "h cmd1-1", cmd1.GetCommand("cmd1-1").help)
	assert.Equal(t, "h cmd1-2", cmd1.GetCommand("cmd1-2").help)
	assert.Equal(t, "h cmd1-3", cmd1.GetCommand("cmd1-3").help)
	assert.Equal(t, "h cmd1-4", cmd1.GetCommand("cmd1-4").help)

	assert.Equal(t, "h cmd11-1", cmd11.GetCommand("cmd11-1").help)
	assert.Equal(t, "h cmd11-2", cmd11.GetCommand("cmd11-2").help)
	assert.Equal(t, "h cmd11-3", cmd11.GetCommand("cmd11-3").help)
	assert.Equal(t, "h cmd11-4", cmd11.GetCommand("cmd11-4").help)
}
