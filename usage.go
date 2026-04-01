package kingpin

import (
	"bytes"
	"errors"
	"fmt"
	"go/doc"
	"io"
	"strings"
)

var preIndent = "  "

// FormatTwoColumns formats rows of two-column data (e.g., flags and their descriptions)
// into aligned output written to w, with the given indent, padding, and total width.
func FormatTwoColumns(w io.Writer, indent, padding, width int, rows [][2]string) {
	// Find size of first column.
	s := 0
	for _, row := range rows {
		if c := len(row[0]); c > s && c < 30 {
			s = c
		}
	}

	indentStr := strings.Repeat(" ", indent)
	offsetStr := strings.Repeat(" ", s+padding)

	for _, row := range rows {
		buf := bytes.NewBuffer(nil)
		doc.ToText(buf, row[1], "", preIndent, width-s-padding-indent)
		lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
		fmt.Fprintf(w, "%s%-*s%*s", indentStr, s, row[0], padding, "")
		if len(row[0]) >= 30 {
			fmt.Fprintf(w, "\n%s%s", indentStr, offsetStr)
		}
		fmt.Fprintf(w, "%s\n", lines[0])
		for _, line := range lines[1:] {
			fmt.Fprintf(w, "%s%s%s\n", indentStr, offsetStr, line)
		}
	}
}

// Usage writes application usage to w. It parses args to determine
// appropriate help context, such as which command to show help for.
func (a *Application) Usage(args []string) {
	context, err := a.parseContext(true, args)
	a.FatalIfError(err, "")
	if err := a.UsageForContext(context); err != nil {
		panic(err)
	}
}

func formatAppUsage(app *ApplicationModel) string {
	s := []string{app.Name}
	if len(app.Flags) > 0 {
		s = append(s, app.FlagSummary())
	}
	if len(app.Args) > 0 {
		s = append(s, app.ArgSummary())
	}
	return strings.Join(s, " ")
}

func formatCmdUsage(app *ApplicationModel, cmd *CmdModel) string {
	s := []string{app.Name, cmd.String()}
	if len(cmd.Flags) > 0 {
		s = append(s, cmd.FlagSummary())
	}
	if len(cmd.Args) > 0 {
		s = append(s, cmd.ArgSummary())
	}
	return strings.Join(s, " ")
}

// FormatFlagCompact formats a flag name without placeholder value for compact display.
func FormatFlagCompact(haveShort bool, flag *FlagModel) string {
	flagString := ""
	flagName := flag.Name
	if flag.IsBoolFlag() {
		flagName = "[no-]" + flagName
	}
	if flag.Short != 0 {
		flagString += fmt.Sprintf("-%c, --%s", flag.Short, flagName)
	} else {
		if haveShort {
			flagString += fmt.Sprintf("    --%s", flagName)
		} else {
			flagString += fmt.Sprintf("--%s", flagName)
		}
	}
	return flagString
}

// FormatFlag formats a flag with its placeholder value and cumulative indicator.
func FormatFlag(haveShort bool, flag *FlagModel) string {
	flagString := FormatFlagCompact(haveShort, flag)
	if !flag.IsBoolFlag() {
		flagString += fmt.Sprintf("=%s", flag.FormatPlaceHolder())
	}
	if v, ok := flag.Value.(repeatableFlag); ok && v.IsCumulative() {
		flagString += " ..."
	}
	return flagString
}

// UsageForContext displays usage information from a ParseContext (obtained from
// Application.ParseContext() or Action(f) callbacks).
func (a *Application) UsageForContext(context *ParseContext) error {
	if a.usageTemplate != "" && a.templateRenderer != nil {
		return a.templateRenderer(a, context, 2, a.usageTemplate)
	}

	if a.usageRenderer != nil {
		return a.usageForContextWithUsageRenderer(context, 2, a.usageRenderer)
	}

	if a.usageFuncs != nil && a.templateRenderer != nil {
		return a.templateRenderer(a, context, 2, DefaultUsageTemplate)
	}

	return a.usageForContextWithUsageRenderer(context, 2, RenderDefault)
}

// UsageForContextWithUsageRenderer renders usage without using text/template.
// This is the fallback path when no UsageRenderer is set and the template doesn't match
// a built-in renderer. Callers who want to avoid pulling in text/template should
// use UsageRenderer or one of the built-in renderer constants.
func (a *Application) UsageForContextWithUsageRenderer(context *ParseContext, indent int) error {
	return a.usageForContextWithUsageRenderer(context, indent, a.usageRenderer)
}

func (a *Application) usageForContextWithUsageRenderer(context *ParseContext, indent int, fn UsageRenderer) error {
	if fn == nil {
		return errors.New("no usage renderer provided")
	}

	width := guessWidth(a.usageWriter)
	var selectedCommand *CmdModel
	if context.SelectedCommand != nil {
		selectedCommand = context.SelectedCommand.Model()
	}

	return fn(a.usageWriter, &UsageContext{
		App:    a.Model(),
		Indent: indent,
		Width:  width,
		Context: &UsageParseContext{
			SelectedCommand: selectedCommand,
			FlagGroupModel:  context.flags.Model(),
			ArgGroupModel:   context.arguments.Model(),
		},
	})
}

// UsageForContextWithTemplate renders usage using text/template for custom template strings.
// This is the fallback path when no UsageRenderer is set. Callers who want to avoid pulling in text/template
// should specify a UsageRenderer and call UsageForContextWithUsageRenderer instead.
//
// Note: calling this method directly will cause text/template to be linked into the binary.
// For dead code elimination, prefer UsageRenderer with a UsageRenderer.
func (a *Application) UsageForContextWithTemplate(context *ParseContext, indent int, tmpl string) error {

	return templateRenderFunc(a, context, indent, tmpl)
}
