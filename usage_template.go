package kingpin

import (
	"bytes"
	"strings"
	"text/template"
)

// templateRenderFunc executes usage rendering via text/template.
// This function is stored in Application.templateRenderer when
// UsageTemplate() or UsageFuncs() is called, ensuring that text/template
// is only linked into the binary when templates are actually used.
// Programs that only call UsageRenderer will not reference this function,
// allowing the linker to eliminate text/template and its reflect.MethodByName dependency.
func templateRenderFunc(a *Application, context *ParseContext, indent int, tmpl string) error {
	width := guessWidth(a.usageWriter)
	var selectedCommand *CmdModel
	if context.SelectedCommand != nil {
		selectedCommand = context.SelectedCommand.Model()
	}
	ctx := &UsageContext{
		App:   a.Model(),
		Width: width,
		Context: &UsageParseContext{
			SelectedCommand: selectedCommand,
			FlagGroupModel:  context.flags.Model(),
			ArgGroupModel:   context.arguments.Model(),
		},
	}

	funcs := template.FuncMap{
		"Indent": func(level int) string {
			return strings.Repeat(" ", level*indent)
		},
		"Wrap": func(indent int, s string) string {
			return Wrap(width, indent, s)
		},
		"FormatFlag":        FormatFlag,
		"FormatFlagCompact": FormatFlagCompact,
		"FlagsToTwoColumnsCompact": func(f []*FlagModel) [][2]string {
			var rows [][2]string
			haveShort := ShortFlagsPresent(f)
			for _, flag := range f {
				if !flag.Hidden {
					rows = append(rows, [2]string{FormatFlagCompact(haveShort, flag), flag.Help})
				}
			}
			return rows
		},
		"FlagsToTwoColumns": FlagsToTwoColumns,
		"RequiredFlags": func(f []*FlagModel) []*FlagModel {
			var out []*FlagModel
			for _, flag := range f {
				if flag.Required {
					out = append(out, flag)
				}
			}
			return out
		},
		"OptionalFlags": func(f []*FlagModel) []*FlagModel {
			var out []*FlagModel
			for _, flag := range f {
				if !flag.Required {
					out = append(out, flag)
				}
			}
			return out
		},
		"ArgsToTwoColumns": ArgsToTwoColumns,
		"FormatTwoColumns": func(rows [][2]string) string {
			var buf bytes.Buffer
			FormatTwoColumns(&buf, indent, 2, width, rows)
			return buf.String()
		},
		"FormatTwoColumnsWithIndent": func(rows [][2]string, indent, padding int) string {
			var buf bytes.Buffer
			FormatTwoColumns(&buf, indent, padding, width, rows)
			return buf.String()
		},
		"FormatAppUsage":     formatAppUsage,
		"FormatCommandUsage": formatCmdUsage,
		"IsCumulative": func(value Value) bool {
			r, ok := value.(repeatableFlag)
			return ok && r.IsCumulative()
		},
		"Char": func(c rune) string {
			return string(c)
		},
	}
	for k, v := range a.usageFuncs {
		funcs[k] = v
	}

	t, err := template.New("usage").Funcs(funcs).Parse(tmpl)
	if err != nil {
		return err
	}

	return t.Execute(a.usageWriter, ctx)
}
