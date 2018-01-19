package kingpin

import (
	"strings"

	"github.com/stretchr/testify/assert"

	"testing"
)

func parseAndExecute(app *Application, context *ParseContext) (string, error) {
	if err := parse(context, app); err != nil {
		return "", err
	}

	selected, err := app.setValues(context)
	if err != nil {
		return "", err
	}

	return app.execute(context, selected)
}

func complete(t *testing.T, app *Application, args ...string) Completion {
	context, err := app.ParseContext(args)
	assert.NoError(t, err)
	return app.resolveCompletion(context)
}

func TestNestedCommands(t *testing.T) {
	app := New("app", "")
	sub1 := app.Command("sub1", "")
	sub1.Flag("sub1", "")
	subsub1 := sub1.Command("sub1sub1", "")
	subsub1.Command("sub1sub1end", "")

	sub2 := app.Command("sub2", "")
	sub2.Flag("sub2", "")
	sub2.Command("sub2sub1", "")

	context := tokenize([]string{"sub1", "sub1sub1", "sub1sub1end"}, false, nil)
	selected, err := parseAndExecute(app, context)
	assert.NoError(t, err)
	assert.True(t, context.EOL())
	assert.Equal(t, "sub1 sub1sub1 sub1sub1end", selected)
}

func TestNestedCommandsWithArgs(t *testing.T) {
	app := New("app", "")
	cmd := app.Command("a", "").Command("b", "")
	a := cmd.Arg("a", "").String()
	b := cmd.Arg("b", "").String()
	context := tokenize([]string{"a", "b", "c", "d"}, false, nil)
	selected, err := parseAndExecute(app, context)
	assert.NoError(t, err)
	assert.True(t, context.EOL())
	assert.Equal(t, "a b", selected)
	assert.Equal(t, "c", *a)
	assert.Equal(t, "d", *b)
}

func TestNestedCommandsWithFlags(t *testing.T) {
	app := New("app", "")
	cmd := app.Command("a", "").Command("b", "")
	a := cmd.Flag("aaa", "").Short('a').String()
	b := cmd.Flag("bbb", "").Short('b').String()
	err := app.init()
	assert.NoError(t, err)
	context := tokenize(strings.Fields("a b --aaa x -b x"), false, nil)
	selected, err := parseAndExecute(app, context)
	assert.NoError(t, err)
	assert.True(t, context.EOL())
	assert.Equal(t, "a b", selected)
	assert.Equal(t, "x", *a)
	assert.Equal(t, "x", *b)
}

func TestNestedCommandWithMergedFlags(t *testing.T) {
	app := New("app", "")
	cmd0 := app.Command("a", "")
	cmd0f0 := cmd0.Flag("aflag", "").Bool()
	// cmd1 := app.Command("b", "")
	// cmd1f0 := cmd0.Flag("bflag", "").Bool()
	cmd00 := cmd0.Command("aa", "")
	cmd00f0 := cmd00.Flag("aaflag", "").Bool()
	err := app.init()
	assert.NoError(t, err)
	context := tokenize(strings.Fields("a aa --aflag --aaflag"), false, nil)
	selected, err := parseAndExecute(app, context)
	assert.NoError(t, err)
	assert.True(t, *cmd0f0)
	assert.True(t, *cmd00f0)
	assert.Equal(t, "a aa", selected)
}

func TestNestedCommandWithDuplicateFlagErrors(t *testing.T) {
	app := New("app", "")
	app.Flag("test", "").Bool()
	app.Command("cmd0", "").Flag("test", "").Bool()
	err := app.init()
	assert.Error(t, err)
}

func TestNestedCommandWithArgAndMergedFlags(t *testing.T) {
	app := New("app", "")
	cmd0 := app.Command("a", "")
	cmd0f0 := cmd0.Flag("aflag", "").Bool()
	// cmd1 := app.Command("b", "")
	// cmd1f0 := cmd0.Flag("bflag", "").Bool()
	cmd00 := cmd0.Command("aa", "")
	cmd00a0 := cmd00.Arg("arg", "").String()
	cmd00f0 := cmd00.Flag("aaflag", "").Bool()
	err := app.init()
	assert.NoError(t, err)
	context := tokenize(strings.Fields("a aa hello --aflag --aaflag"), false, nil)
	selected, err := parseAndExecute(app, context)
	assert.NoError(t, err)
	assert.True(t, *cmd0f0)
	assert.True(t, *cmd00f0)
	assert.Equal(t, "a aa", selected)
	assert.Equal(t, "hello", *cmd00a0)
}

func TestDefaultSubcommandEOL(t *testing.T) {
	app := newTestApp()
	c0 := app.Command("c0", "").Default()
	c0.Command("c01", "").Default()
	c0.Command("c02", "")

	cmd, err := app.Parse([]string{"c0"})
	assert.NoError(t, err)
	assert.Equal(t, "c0 c01", cmd)
}

func TestDefaultSubcommandWithArg(t *testing.T) {
	app := newTestApp()
	c0 := app.Command("c0", "").Default()
	c01 := c0.Command("c01", "").Default()
	c012 := c01.Command("c012", "").Default()
	a0 := c012.Arg("a0", "").String()
	c0.Command("c02", "")

	cmd, err := app.Parse([]string{"c0", "hello"})
	assert.NoError(t, err)
	assert.Equal(t, "c0 c01 c012", cmd)
	assert.Equal(t, "hello", *a0)
}

func TestDefaultSubcommandWithFlags(t *testing.T) {
	app := newTestApp()
	c0 := app.Command("c0", "").Default()
	_ = c0.Flag("f0", "").Int()
	c0c1 := c0.Command("c1", "").Default()
	c0c1f1 := c0c1.Flag("f1", "").Int()
	selected, err := app.Parse([]string{"--f1=2"})
	assert.NoError(t, err)
	assert.Equal(t, "c0 c1", selected)
	assert.Equal(t, 2, *c0c1f1)
	_, err = app.Parse([]string{"--f2"})
	assert.Error(t, err)
}

func TestMultipleDefaultCommands(t *testing.T) {
	app := newTestApp()
	app.Command("c0", "").Default()
	app.Command("c1", "").Default()
	_, err := app.Parse([]string{})
	assert.Error(t, err)
}

func TestAliasedCommand(t *testing.T) {
	app := newTestApp()
	app.Command("one", "").Alias("two")
	selected, _ := app.Parse([]string{"one"})
	assert.Equal(t, "one", selected)
	selected, _ = app.Parse([]string{"two"})
	assert.Equal(t, "one", selected)
	// 2 due to "help" and "one"
	assert.Equal(t, 2, len(app.Model().FlattenedCommands()))
}

func TestDuplicateAlias(t *testing.T) {
	app := newTestApp()
	app.Command("one", "")
	app.Command("two", "").Alias("one")
	_, err := app.Parse([]string{"one"})
	assert.Error(t, err)
}

func TestFlagCompletion(t *testing.T) {
	app := newTestApp()
	app.Command("one", "")
	two := app.Command("two", "")
	two.Flag("flag-1", "")
	two.Flag("flag-2", "").HintOptions("opt1", "opt2", "opt3")
	two.Flag("flag-3", "")

	cases := []struct {
		target               cmdMixin
		flagName             string
		flagValue            string
		expectedFlagMatch    bool
		expectedOptionMatch  bool
		expectedDirComplete  bool
		expectedFileComplete bool
		expectedWordComplete []string
	}{
		{
			// Test top level flags
			target:               app.cmdMixin,
			flagName:             "",
			flagValue:            "",
			expectedFlagMatch:    false,
			expectedOptionMatch:  false,
			expectedDirComplete:  false,
			expectedFileComplete: false,
			expectedWordComplete: []string{"--help"},
		},
		{
			// Test no flag passed
			target:               two.cmdMixin,
			flagName:             "",
			flagValue:            "",
			expectedFlagMatch:    false,
			expectedOptionMatch:  false,
			expectedDirComplete:  false,
			expectedFileComplete: false,
			expectedWordComplete: []string{"--flag-1", "--flag-2", "--flag-3"},
		},
		{
			// Test an incomplete flag. Should still give all options as if the flag wasn't given at all.
			target:               two.cmdMixin,
			flagName:             "flag-",
			flagValue:            "",
			expectedFlagMatch:    false,
			expectedOptionMatch:  false,
			expectedDirComplete:  false,
			expectedFileComplete: false,
			expectedWordComplete: []string{"--flag-1", "--flag-2", "--flag-3"},
		},
		{
			// Test with a complete flag. Should show available choices for the flag
			// This flag has no options. No options should be produced.
			// Should also report an option was matched
			target:               two.cmdMixin,
			flagName:             "flag-1",
			flagValue:            "",
			expectedFlagMatch:    true,
			expectedOptionMatch:  true,
			expectedDirComplete:  false,
			expectedFileComplete: false,
			expectedWordComplete: []string{},
		},
		{
			// Test with a complete flag. Should show available choices for the flag
			target:               two.cmdMixin,
			flagName:             "flag-2",
			flagValue:            "",
			expectedFlagMatch:    true,
			expectedOptionMatch:  false,
			expectedDirComplete:  false,
			expectedFileComplete: false,
			expectedWordComplete: []string{"opt1", "opt2", "opt3"},
		},
		{
			// Test with a complete flag and complete option for that flag.
			target:               two.cmdMixin,
			flagName:             "flag-2",
			flagValue:            "opt1",
			expectedFlagMatch:    true,
			expectedOptionMatch:  true,
			expectedDirComplete:  false,
			expectedFileComplete: false,
			expectedWordComplete: []string{"opt1", "opt2", "opt3"},
		},
	}

	for i, c := range cases {
		choices, flagMatch, optionMatch := c.target.FlagCompletion(c.flagName, c.flagValue)
		assert.Equal(t, c.expectedDirComplete, choices.Directories, "Test case %d: expectedDirComplete != actual Directories", i+1)
		assert.Equal(t, c.expectedFileComplete, choices.Files, "Test case %d: expectedFileComplete != actual Files", i+1)
		assert.Equal(t, c.expectedWordComplete, choices.resolveWords(), "Test case %d: expectedWordComplete != actual resolved words", i+1)
		assert.Equal(t, c.expectedFlagMatch, flagMatch, "Test case %d: expectedFlagMatch != flagMatch", i+1)
		assert.Equal(t, c.expectedOptionMatch, optionMatch, "Test case %d: expectedOptionMatch != optionMatch", i+1)
	}

}

func TestCmdCompletion(t *testing.T) {
	app := newTestApp()
	app.Command("one", "")
	two := app.Command("two", "")
	two.Command("sub1", "")
	two.Command("sub2", "")

	assert.Equal(t, []string{"help", "one", "two"}, complete(t, app).resolveWords())
	assert.Equal(t, []string{"sub1", "sub2"}, complete(t, app, "two").resolveWords())
}

func TestHiddenCmdCompletion(t *testing.T) {
	app := newTestApp()

	// top level visible & hidden cmds, with no sub-cmds
	app.Command("visible1", "")
	app.Command("hidden1", "").Hidden()

	// visible cmd with visible & hidden sub-cmds
	visible2 := app.Command("visible2", "")
	visible2.Command("visible2-visible", "")
	visible2.Command("visible2-hidden", "").Hidden()

	// hidden cmd with visible & hidden sub-cmds
	hidden2 := app.Command("hidden2", "").Hidden()
	hidden2.Command("hidden2-visible", "")
	hidden2.Command("hidden2-hidden", "").Hidden()

	// Only top level visible cmds should show
	result := complete(t, app)
	assert.False(t, result.Directories)
	assert.False(t, result.Files)
	assert.Equal(t, []string{"help", "visible1", "visible2"}, result.resolveWords())

	// Only visible sub-cmds should show
	result = complete(t, app, "visible2")
	assert.False(t, result.Directories)
	assert.False(t, result.Files)
	assert.Equal(t, []string{"visible2-visible"}, result.resolveWords())

	// Hidden commands should still complete visible sub-cmds
	result = complete(t, app, "hidden2")
	assert.False(t, result.Directories)
	assert.False(t, result.Files)
	assert.Equal(t, []string{"hidden2-visible"}, result.resolveWords())
}

func TestDefaultCmdCompletion(t *testing.T) {
	app := newTestApp()

	cmd1 := app.Command("cmd1", "")

	cmd1Sub1 := cmd1.Command("cmd1-sub1", "")
	cmd1Sub1.Arg("cmd1-sub1-arg1", "").HintOptions("cmd1-arg1").String()

	cmd2 := app.Command("cmd2", "").Default()

	cmd2.Command("cmd2-sub1", "")

	cmd2Sub2 := cmd2.Command("cmd2-sub2", "").Default()

	cmd2Sub2Sub1 := cmd2Sub2.Command("cmd2-sub2-sub1", "").Default()
	cmd2Sub2Sub1.Arg("cmd2-sub2-sub1-arg1", "").HintOptions("cmd2-sub2-sub1-arg1").String()
	cmd2Sub2Sub1.Arg("cmd2-sub2-sub1-arg2", "").HintOptions("cmd2-sub2-sub1-arg2").String()

	// Without args, should get:
	//   - root cmds (including implicit "help")
	//   - thread of default cmds
	//   - first arg hints for the final default cmd
	result := complete(t, app)
	assert.False(t, result.Directories)
	assert.False(t, result.Files)
	assert.Equal(t, []string{"cmd1", "cmd2", "cmd2-sub1", "cmd2-sub2", "cmd2-sub2-sub1", "cmd2-sub2-sub1-arg1", "help"}, result.resolveWords())

	// With a non-default cmd already listed, should get:
	//   - sub cmds of that arg
	result = complete(t, app, "cmd1")
	assert.False(t, result.Directories)
	assert.False(t, result.Files)
	assert.Equal(t, []string{"cmd1-sub1"}, result.resolveWords())

	// With an explicit default cmd listed, should get:
	//   - default child-cmds
	//   - first arg hints for the final default cmd
	result = complete(t, app, "cmd2")
	assert.False(t, result.Directories)
	assert.False(t, result.Files)
	assert.Equal(t, []string{"cmd2-sub1", "cmd2-sub2", "cmd2-sub2-sub1", "cmd2-sub2-sub1-arg1"}, result.resolveWords())

	// Args should be completed when all preceding cmds are explicit, and when
	// any of them are implicit (not listed). Check this by trying all possible
	// combinations of choosing/excluding the three levels of cmds. This tests
	// root-level default, middle default, and end default.
	for i := 0; i < 8; i++ {
		var cmdline []string

		if i&1 != 0 {
			cmdline = append(cmdline, "cmd2")
		}
		if i&2 != 0 {
			cmdline = append(cmdline, "cmd2-sub2")
		}
		if i&4 != 0 {
			cmdline = append(cmdline, "cmd2-sub2-sub1")
		}

		result = complete(t, app, cmdline...)
		assert.False(t, result.Directories)
		assert.False(t, result.Files)
		assert.Contains(t, result.resolveWords(), "cmd2-sub2-sub1-arg1", "with cmdline: %v", cmdline)
	}

	// With both args of a default sub cmd, should get no completions
	result = complete(t, app, "arg1", "arg2")
	assert.False(t, result.Directories)
	assert.False(t, result.Files)
	assert.Empty(t, result.resolveWords())
}

func TestCannotMixOptionalArgWithCommand(t *testing.T) {
	app := newTestApp()
	app.Arg("arg", "").String()
	app.Command("cmd", "")
	_, err := app.Parse([]string{})
	assert.Error(t, err)
}

func TestMixArgWithCommand(t *testing.T) {
	app := newTestApp()
	arg0 := app.Arg("arg0", "").Required().String()
	arg1 := app.Arg("arg1", "").Required().String()
	cmd := app.Command("cmd", "")
	arg2 := cmd.Arg("arg2", "").Required().String()
	_, err := app.Parse([]string{"a", "b", "cmd"})
	assert.Error(t, err)
	selected, err := app.Parse([]string{"a", "b", "cmd", "c"})
	assert.NoError(t, err)
	assert.Equal(t, "a", *arg0)
	assert.Equal(t, "b", *arg1)
	assert.Equal(t, "cmd", selected)
	assert.Equal(t, "c", *arg2)
}

func TestMultiArgCmdCompletion(t *testing.T) {
	app := newTestApp()
	cmd := app.Command("cmd", "")
	cmd.Arg("arg", "").HintOptions("opt1", "opt2", "opt3").Strings()

	result := complete(t, app, "cmd")
	assert.False(t, result.Directories)
	assert.False(t, result.Files)
	assert.Equal(t, []string{"opt1", "opt2", "opt3"}, result.resolveWords())

	result = complete(t, app, "cmd", "opt1")
	assert.False(t, result.Directories)
	assert.False(t, result.Files)
	assert.Equal(t, []string{"opt1", "opt2", "opt3"}, result.resolveWords())

	result = complete(t, app, "cmd", "opt3", "opt2")
	assert.False(t, result.Directories)
	assert.False(t, result.Files)
	assert.Equal(t, []string{"opt1", "opt2", "opt3"}, result.resolveWords())
}
