package kingpin

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTerminatorAfterArgumentExpansion(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		fileContents string
		wantCount    int
		wantArgs     []string
	}{
		{
			name:      "single short flag",
			args:      []string{"-v", "--", "--name=value", "--"},
			wantCount: 1,
			wantArgs:  []string{"--name=value", "--"},
		},
		{
			name:      "combined short flags",
			args:      []string{"-vvv", "--", "--name=value", "--"},
			wantCount: 3,
			wantArgs:  []string{"--name=value", "--"},
		},
		{
			name:      "interspersed combined short flags",
			args:      []string{"before", "-vvv", "--", "--name=value"},
			wantCount: 3,
			wantArgs:  []string{"before", "--name=value"},
		},
		{
			name:         "terminator inside arguments file",
			args:         []string{"@args", "--name=tail"},
			fileContents: "-v\n--\n--name=value\n--\n",
			wantCount:    1,
			wantArgs:     []string{"--name=value", "--", "--name=tail"},
		},
		{
			name:         "arguments file before terminator",
			args:         []string{"@args", "--", "--name=value"},
			fileContents: "-vvv\n",
			wantCount:    3,
			wantArgs:     []string{"--name=value"},
		},
		{
			name:     "file expansion disabled after terminator",
			args:     []string{"--", "@does-not-exist"},
			wantArgs: []string{"@does-not-exist"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.fileContents != "" {
				path := filepath.Join(t.TempDir(), "args")
				if err := ioutil.WriteFile(path, []byte(tt.fileContents), 0600); err != nil {
					t.Fatal(err)
				}
				tt.args[0] = "@" + path
			}
			app := newTestApp()
			count := app.Flag("verbose", "").Short('v').Counter()
			name := app.Flag("name", "").String()
			args := app.Arg("args", "").Strings()

			_, err := app.Parse(tt.args)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantCount, *count)
			assert.Equal(t, tt.wantArgs, *args)
			assert.Empty(t, *name)
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

func TestBareDashIsPositionalArg(t *testing.T) {
	app := New("test", "")
	arg := app.Arg("input", "filename or - for stdin").String()
	_, err := app.Parse([]string{"-"})
	assert.NoError(t, err)
	assert.Equal(t, "-", *arg)
}
