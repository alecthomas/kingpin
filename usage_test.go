package kingpin

import (
	"bytes"
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

func TestFormatTwoColumns(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	formatTwoColumns(buf, 2, 2, 20, [][2]string{
		{"--hello", "Hello world help with something that is cool."},
	})
	expected := `  --hello  Hello
           world
           help with
           something
           that is
           cool.
`
	assert.Equal(t, expected, buf.String())
}

func TestFormatTwoColumnsWide(t *testing.T) {
	samples := [][2]string{
		{strings.Repeat("x", 29), "29 chars"},
		{strings.Repeat("x", 30), "30 chars"}}
	buf := bytes.NewBuffer(nil)
	formatTwoColumns(buf, 0, 0, 200, samples)
	expected := `xxxxxxxxxxxxxxxxxxxxxxxxxxxxx29 chars
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
                             30 chars
`
	assert.Equal(t, expected, buf.String())
}

func TestHiddenCommand(t *testing.T) {
	templates := []struct{ name, template string }{
		{"default", DefaultUsageTemplate},
		{"Compact", CompactUsageTemplate},
		{"Long", LongHelpTemplate},
		{"Man", ManPageTemplate},
	}

	var buf bytes.Buffer
	t.Log("1")

	a := New("test", "Test").Writer(&buf).Terminate(nil)
	a.Command("visible", "visible")
	a.Command("hidden", "hidden").Hidden()

	for _, tp := range templates {
		buf.Reset()
		a.UsageTemplate(tp.template)
		a.Parse(nil)
		// a.Parse([]string{"--help"})
		usage := buf.String()
		t.Logf("Usage for %s is:\n%s\n", tp.name, usage)

		assert.NotContains(t, usage, "hidden")
		assert.Contains(t, usage, "visible")
	}
}

func TestUsageFuncs(t *testing.T) {
	var buf bytes.Buffer
	a := New("test", "Test").Writer(&buf).Terminate(nil)
	tpl := `{{ add 2 1 }}`
	a.UsageTemplate(tpl)
	a.UsageFuncs(template.FuncMap{
		"add": func(x, y int) int { return x + y },
	})
	a.Parse([]string{"--help"})
	usage := buf.String()
	assert.Equal(t, "3", usage)
}

func TestCmdClause_HelpLong(t *testing.T) {
	var buf bytes.Buffer
	tpl := `{{define "FormatUsage"}}{{.HelpLong}}{{end}}\
{{template "FormatUsage" .Context.SelectedCommand}}`

	a := New("test", "Test").Writer(&buf).Terminate(nil)
	a.UsageTemplate(tpl)
	a.Command("command", "short help text").HelpLong("long help text")

	a.Parse([]string{"command", "--help"})
	usage := buf.String()
	assert.Equal(t, "long help text", usage)
}
