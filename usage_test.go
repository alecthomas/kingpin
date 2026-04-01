package kingpin

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatTwoColumns(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	FormatTwoColumns(buf, 2, 2, 20, [][2]string{
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
	FormatTwoColumns(buf, 0, 0, 200, samples)
	expected := `xxxxxxxxxxxxxxxxxxxxxxxxxxxxx29 chars
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
                             30 chars
`
	assert.Equal(t, expected, buf.String())
}

func TestHiddenCommand(t *testing.T) {
	templates := []struct {
		name     string
		template string
		renderer UsageRenderer
	}{
		{
			name:     "default template",
			template: DefaultUsageTemplate,
		},
		{
			name:     "default renderer",
			renderer: RenderDefault,
		},
		{
			name:     "Compact template",
			template: CompactUsageTemplate,
		},
		{
			name:     "Compact renderer",
			renderer: RenderCompact,
		},
		{
			name:     "Long template",
			template: LongHelpTemplate,
		},
		{
			name:     "Long renderer",
			renderer: RenderLongHelp,
		},
		{
			name:     "Man template",
			template: ManPageTemplate,
		},
		{
			name:     "Man renderer",
			template: ManPageTemplate,
		},
	}

	for _, tp := range templates {
		t.Run(tp.name, func(t *testing.T) {
			var buf bytes.Buffer
			a := New("test", "Test").Writer(&buf).Terminate(nil)
			a.Command("visible", "visible")
			a.Command("hidden", "hidden").Hidden()
			if tp.template != "" {
				a.UsageTemplate(tp.template)
			}
			if tp.renderer != nil {
				a.UsageRenderer(tp.renderer)
			}
			_, err := a.Parse(nil)
			require.ErrorIs(t, err, ErrCommandNotSpecified)
			usage := buf.String()
			t.Logf("Usage for %s is:\n%s\n", tp.name, usage)

			assert.NotContains(t, usage, "hidden")
			assert.Contains(t, usage, "visible")
		})
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
	_, err := a.Parse([]string{"--help"})
	require.NoError(t, err)
	usage := buf.String()
	assert.Equal(t, "3", usage)

	buf.Reset()
	a = New("test", "Test help").UsageWriter(&buf).Terminate(nil)
	a.UsageFuncs(map[string]interface{}{
		"Wrap": func(indent int, s string) string { return "OVERRIDDEN\n" },
	})

	_, err = a.Parse([]string{"--help"})
	require.NoError(t, err)
	usage = buf.String()
	assert.Contains(t, usage, "OVERRIDDEN")
}

func TestUsageFuncsApplyToHiddenFlagPaths(t *testing.T) {
	for _, tp := range []struct {
		name string
		flag string
	}{
		{
			name: "help-long",
			flag: "--help-long",
		},
		{
			name: "help-man",
			flag: "--help-man",
		},
	} {
		t.Run(tp.name, func(t *testing.T) {
			var buf bytes.Buffer
			a := New("test", "Test help").HiddenHelpWriter(&buf).Terminate(nil)
			a.Flag("verbose", "Verbose output.").Short('v').Bool()
			a.UsageFuncs(map[string]interface{}{
				"Wrap": func(indent int, s string) string { return "OVERRIDDEN\n" },
				"Char": func(c rune) string { return "OVERRIDDEN" },
			})

			_, err := a.Parse([]string{tp.flag})
			require.NoError(t, err)

			assert.Contains(t, buf.String(), "OVERRIDDEN")
		})
	}
}

func TestUsageRenderer(t *testing.T) {
	var buf bytes.Buffer
	a := New("test", "Test").Writer(&buf).Terminate(nil)
	a.UsageRenderer(func(w io.Writer, ctx *UsageContext) error {
		_, err := fmt.Fprintf(w, "custom: %s", ctx.App.Name)
		return err
	})

	_, err := a.Parse([]string{"--help"})
	require.NoError(t, err)
	usage := buf.String()
	assert.Equal(t, "custom: test", usage)
}

func TestCmdClause_HelpLong(t *testing.T) {
	var buf bytes.Buffer
	tpl := `{{define "FormatUsage"}}{{.HelpLong}}{{end -}}
{{template "FormatUsage" .Context.SelectedCommand}}`

	a := New("test", "Test").Writer(&buf).Terminate(nil)
	a.UsageTemplate(tpl)
	a.Command("command", "short help text").HelpLong("long help text")

	_, err := a.Parse([]string{"command", "--help"})
	require.NoError(t, err)
	usage := buf.String()
	assert.Equal(t, "long help text", usage)
}

func TestArgEnvVar(t *testing.T) {
	var buf bytes.Buffer

	a := New("test", "Test").Writer(&buf).Terminate(nil)
	a.Arg("arg", "Enable arg").Envar("ARG").String()
	a.Flag("flag", "Enable flag").Envar("FLAG").String()

	_, err := a.Parse([]string{"command", "--help"})
	require.NoError(t, err)
	usage := buf.String()
	assert.Contains(t, usage, "($ARG)")
	assert.Contains(t, usage, "($FLAG)")
}

func TestRendererPrioritizedWhenUsageFuncsSet(t *testing.T) {
	var buf bytes.Buffer
	a := New("test", "Test").UsageWriter(&buf).Terminate(nil)
	a.UsageFuncs(map[string]interface{}{
		"Wrap": func(indent int, s string) string { return "FROM_TEMPLATE\n" },
	})
	a.UsageRenderer(func(w io.Writer, ctx *UsageContext) error {
		_, err := fmt.Fprint(w, "FROM_RENDERER")
		return err
	})

	_, err := a.Parse([]string{"--help"})
	require.NoError(t, err)

	assert.Equal(t, "FROM_RENDERER", buf.String())

}

func TestUsageRendererDoesNotOverrideHiddenFlags(t *testing.T) {
	flags := []string{"--help-long", "--help-man", "--completion-script-bash", "--completion-script-zsh", "--completion-script-fish"}

	for _, flag := range flags {
		t.Run(flag, func(t *testing.T) {
			var buf bytes.Buffer
			a := New("test", "Test help").HiddenHelpWriter(&buf).Terminate(nil)
			a.UsageRenderer(func(w io.Writer, ctx *UsageContext) error {
				_, err := fmt.Fprint(w, "CUSTOM_HELP")
				return err
			})

			_, err := a.Parse([]string{flag})
			require.NoError(t, err)

			assert.NotContains(t, buf.String(), "CUSTOM_HELP")
			assert.NotEmpty(t, buf.String())
		})
	}
}

func TestUsageTemplatePreferredOverUsageRenderer(t *testing.T) {
	var buf bytes.Buffer
	a := New("test", "Test").UsageWriter(&buf).Terminate(nil)
	a.UsageTemplate("TEMPLATE:{{ .App.Name }}")
	a.UsageRenderer(func(w io.Writer, ctx *UsageContext) error {
		_, err := fmt.Fprint(w, "FROM_RENDERER")
		return err
	})

	_, err := a.Parse([]string{"--help"})
	require.NoError(t, err)

	assert.Equal(t, "TEMPLATE:test", buf.String())
}

func TestRenderersMatchTemplates(t *testing.T) {
	tests := []struct {
		name     string
		renderer UsageRenderer
		template string
	}{
		{
			name:     "default",
			renderer: RenderDefault,
			template: DefaultUsageTemplate,
		},
		{
			name:     "separate-optional-flags",
			renderer: RenderSeparateOptionalFlags,
			template: SeparateOptionalFlagsUsageTemplate,
		},
		{
			name:     "compact",
			renderer: RenderCompact,
			template: CompactUsageTemplate,
		},
		{
			name:     "long-help",
			renderer: RenderLongHelp,
			template: LongHelpTemplate,
		},
		{
			name:     "man-page",
			renderer: RenderManPage,
			template: ManPageTemplate,
		},
		{
			name:     "bash-completion",
			renderer: RenderBashCompletion,
			template: BashCompletionTemplate,
		},
		{
			name:     "zsh-completion",
			renderer: RenderZshCompletion,
			template: ZshCompletionTemplate,
		},
		{
			name:     "fish-completion",
			renderer: RenderFishCompletion,
			template: FishCompletionTemplate,
		},
	}

	contexts := []struct {
		name string
		args []string
	}{
		{
			name: "root",
			args: nil,
		},
		{
			name: "nested-command",
			args: []string{"sub", "nested"},
		},
	}

	for _, test := range tests {
		for _, context := range contexts {
			t.Run(test.name+"/"+context.name, func(t *testing.T) {
				var templateBuf, rendererBuf bytes.Buffer
				makeApp := func(w *bytes.Buffer) *Application {
					a := New("test", "A test application.").UsageWriter(w).Terminate(nil).Version("1.0").Author("Test Author")
					a.Flag("verbose", "Enable verbose output.").Short('v').Bool()
					a.Flag("config", "Path to config file.").String()

					sub := a.Command("sub", "A subcommand.")
					sub.Flag("count", "Number of items.").Int()

					nested := sub.Command("nested", "A nested command.")
					nested.Arg("file", "File to process.").String()

					a.Command("other", "Another command.")
					return a
				}

				appT := makeApp(&templateBuf)
				ctx, err := appT.ParseContext(context.args)
				require.NoError(t, err)
				err = appT.UsageForContextWithTemplate(ctx, 2, test.template)
				require.NoError(t, err)

				appR := makeApp(&rendererBuf)
				ctx, err = appR.ParseContext(context.args)
				require.NoError(t, err)
				err = appR.usageForContextWithUsageRenderer(ctx, 2, test.renderer)
				require.NoError(t, err)

				assert.Equal(t, templateBuf.String(), rendererBuf.String())
			})
		}
	}
}
