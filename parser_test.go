package kingpin

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParserArgsOnly(t *testing.T) {

	for _, tt := range []struct {
		name     string
		when     func(app *Application)
		parse    []string
		wantArgs []string
		wantFlag string
	}{
		{
			name:     "flag before terminator",
			parse:    []string{"--flag=flag", "a1", "--", "a2"},
			wantArgs: []string{"a1", "a2"},
			wantFlag: "flag",
		},
		{
			name:     "flag after terminator",
			parse:    []string{"--", "--flag=flag", "a1", "a2"},
			wantArgs: []string{"--flag=flag", "a1", "a2"},
			wantFlag: "",
		},
		{
			name:     "multiple terminators",
			parse:    []string{"a0", "--", "--flag=flag", "--", "a1", "a2"},
			wantArgs: []string{"a0", "--flag=flag", "a1", "a2"},
			wantFlag: "",
		},
		{
			name:     "not interspersed: flag before terminator",
			when:     func(app *Application) { app.Interspersed(false) },
			parse:    []string{"--flag=flag", "a1", "--", "a2"},
			wantArgs: []string{"a1", "a2"},
			wantFlag: "flag",
		},
		{
			name:     "not interspersed: flag after terminator",
			when:     func(app *Application) { app.Interspersed(false) },
			parse:    []string{"--", "--flag=flag", "a1", "a2"},
			wantArgs: []string{"--flag=flag", "a1", "a2"},
			wantFlag: "",
		},
		{
			name:     "not interspersed: multiple terminators",
			when:     func(app *Application) { app.Interspersed(false) },
			parse:    []string{"a0", "--", "--flag=flag", "--", "a1", "a2"},
			wantArgs: []string{"a0", "--flag=flag", "a1", "a2"},
			wantFlag: "",
		},
		{
			name:     "not interspersed: flag before multiple terminators",
			when:     func(app *Application) { app.Interspersed(false) },
			parse:    []string{"--flag=flag", "a0", "--", "a1", "--", "a2"},
			wantArgs: []string{"a0", "a1", "a2"},
			wantFlag: "flag",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var (
				app  = New("test", "")
				args []string
				flag string
			)
			app.Arg("arg", "").StringsVar(&args)
			app.Flag("flag", "").StringVar(&flag)

			if tt.when != nil {
				tt.when(app)
			}

			_, err := app.Parse(tt.parse)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantArgs, args)
			assert.Equal(t, tt.wantFlag, flag)
		})
	}
}

func TestParserExpandFromFile(t *testing.T) {
	f, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	defer os.Remove(f.Name())
	f.WriteString("hello\nworld\n")
	f.Close()

	app := New("test", "")
	arg0 := app.Arg("arg0", "").String()
	arg1 := app.Arg("arg1", "").String()

	_, err = app.Parse([]string{"@" + f.Name()})
	assert.NoError(t, err)
	assert.Equal(t, "hello", *arg0)
	assert.Equal(t, "world", *arg1)
}

func TestParserExpandFromFileLeadingArg(t *testing.T) {
	f, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	defer os.Remove(f.Name())
	f.WriteString("hello\nworld\n")
	f.Close()

	app := New("test", "")
	arg0 := app.Arg("arg0", "").String()
	arg1 := app.Arg("arg1", "").String()
	arg2 := app.Arg("arg2", "").String()

	_, err = app.Parse([]string{"prefix", "@" + f.Name()})
	assert.NoError(t, err)
	assert.Equal(t, "prefix", *arg0)
	assert.Equal(t, "hello", *arg1)
	assert.Equal(t, "world", *arg2)
}

func TestParserExpandFromFileTrailingArg(t *testing.T) {
	f, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	defer os.Remove(f.Name())
	f.WriteString("hello\nworld\n")
	f.Close()

	app := New("test", "")
	arg0 := app.Arg("arg0", "").String()
	arg1 := app.Arg("arg1", "").String()
	arg2 := app.Arg("arg2", "").String()

	_, err = app.Parse([]string{"@" + f.Name(), "suffix"})
	assert.NoError(t, err)
	assert.Equal(t, "hello", *arg0)
	assert.Equal(t, "world", *arg1)
	assert.Equal(t, "suffix", *arg2)
}

func TestParserExpandFromFileMultipleSurroundingArgs(t *testing.T) {
	f, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	defer os.Remove(f.Name())
	f.WriteString("hello\nworld\n")
	f.Close()

	app := New("test", "")
	arg0 := app.Arg("arg0", "").String()
	arg1 := app.Arg("arg1", "").String()
	arg2 := app.Arg("arg2", "").String()
	arg3 := app.Arg("arg3", "").String()

	_, err = app.Parse([]string{"prefix", "@" + f.Name(), "suffix"})
	assert.NoError(t, err)
	assert.Equal(t, "prefix", *arg0)
	assert.Equal(t, "hello", *arg1)
	assert.Equal(t, "world", *arg2)
	assert.Equal(t, "suffix", *arg3)
}

func TestParserExpandFromFileMultipleFlags(t *testing.T) {
	f, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	defer os.Remove(f.Name())
	f.WriteString("--flag1=f1\n--flag2=f2\n")
	f.Close()

	app := New("test", "")
	flag0 := app.Flag("flag0", "").String()
	flag1 := app.Flag("flag1", "").String()
	flag2 := app.Flag("flag2", "").String()
	flag3 := app.Flag("flag3", "").String()

	_, err = app.Parse([]string{"--flag0=f0", "@" + f.Name(), "--flag3=f3"})
	assert.NoError(t, err)
	assert.Equal(t, "f0", *flag0)
	assert.Equal(t, "f1", *flag1)
	assert.Equal(t, "f2", *flag2)
	assert.Equal(t, "f3", *flag3)
}

func TestParseContextPush(t *testing.T) {
	app := New("test", "")
	app.Command("foo", "").Command("bar", "")
	c := tokenize([]string{"foo", "bar"}, false)
	a := c.Next()
	assert.Equal(t, TokenArg, a.Type)
	b := c.Next()
	assert.Equal(t, TokenArg, b.Type)
	c.Push(b)
	c.Push(a)
	a = c.Next()
	assert.Equal(t, "foo", a.Value)
	b = c.Next()
	assert.Equal(t, "bar", b.Value)
}
