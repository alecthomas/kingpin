package kingpin

import (
	"bytes"
	"fmt"
	"go/doc"
	"io"
	"strings"
)

// UsageContext contains all data needed to render usage output.
type UsageContext struct {
	App     *ApplicationModel
	Width   int
	Indent  int
	Context *UsageParseContext
}

// UsageParseContext contains parsed context for usage rendering.
type UsageParseContext struct {
	SelectedCommand *CmdModel
	*FlagGroupModel
	*ArgGroupModel
}

// UsageRenderer is a function that renders usage output programmatically.
// Setting a UsageRenderer on an Application bypasses template-based rendering,
// which avoids importing text/template and enables dead code elimination.
type UsageRenderer func(w io.Writer, ctx *UsageContext) error

// isCumulative checks if a Value is cumulative (repeatable).
func isCumulative(value Value) bool {
	r, ok := value.(interface{ IsCumulative() bool })
	return ok && r.IsCumulative()
}

// WriteFormatCommand writes the flag summary and args portion of a command usage line.
func WriteFormatCommand(w io.Writer, fg *FlagGroupModel, ag *ArgGroupModel) {
	if fg != nil && len(fg.Flags) > 0 {
		summary := fg.FlagSummary()
		if summary != "" {
			fmt.Fprintf(w, " %s", summary)
		}
	}

	if ag == nil {
		return
	}

	for _, arg := range ag.Args {
		if arg.Hidden {
			continue
		}

		name := "<" + arg.Name + ">"
		if arg.PlaceHolder != "" {
			name = arg.PlaceHolder
		}
		if isCumulative(arg.Value) {
			name += "..."
		}
		if !arg.Required {
			fmt.Fprintf(w, " [%s]", name)
		} else {
			fmt.Fprintf(w, " %s", name)
		}
	}
}

// WriteFormatUsage writes the usage line for a command or app, including the
// flag summary, args, optional "<command> [<args> ...]" suffix, and help text.
func WriteFormatUsage(w io.Writer, fg *FlagGroupModel, ag *ArgGroupModel, cg *CmdGroupModel, help string, wrapWidth int) {
	WriteFormatCommand(w, fg, ag)
	if cg != nil && len(cg.Commands) > 0 {
		fmt.Fprint(w, " <command> [<args> ...]")
	}
	fmt.Fprintln(w)
	if help != "" {
		fmt.Fprintln(w)
		buf := bytes.NewBuffer(nil)
		doc.ToText(buf, help, "", preIndent, wrapWidth)
		fmt.Fprint(w, buf.String())
	}
}

// RenderDefault renders the standard usage output.
func RenderDefault(w io.Writer, ctx *UsageContext) error {
	width := ctx.Width

	fmtTwoColumns := func(rows [][2]string) string {
		var buf bytes.Buffer
		FormatTwoColumns(&buf, ctx.Indent, ctx.Indent, width, rows)
		return buf.String()
	}

	if ctx.Context.SelectedCommand != nil {
		cmd := ctx.Context.SelectedCommand
		fmt.Fprintf(w, "usage: %s %s", ctx.App.Name, cmd.String())
		WriteFormatUsage(w, cmd.FlagGroupModel, cmd.ArgGroupModel, cmd.CmdGroupModel, cmd.Help, width)
	} else {
		fmt.Fprintf(w, "usage: %s", ctx.App.Name)
		WriteFormatUsage(w, ctx.App.FlagGroupModel, ctx.App.ArgGroupModel, ctx.App.CmdGroupModel, ctx.App.Help, width)
	}

	fmt.Fprint(w, "\n\n")

	if len(ctx.Context.Flags) > 0 {
		fmt.Fprintln(w, "Flags:")
		fmt.Fprint(w, fmtTwoColumns(FlagsToTwoColumns(ctx.Context.Flags)))
		fmt.Fprintln(w)
	}
	if len(ctx.Context.Args) > 0 {
		fmt.Fprintln(w, "Args:")
		fmt.Fprint(w, fmtTwoColumns(ArgsToTwoColumns(ctx.Context.Args)))
		fmt.Fprintln(w)
	}

	writeCommands := func(commands []*CmdModel) {
		for _, cmd := range commands {
			if cmd.Hidden {
				continue
			}

			fmt.Fprint(w, cmd.FullCommand)
			if cmd.Default {
				fmt.Fprint(w, "*")
			}
			WriteFormatCommand(w, cmd.FlagGroupModel, cmd.ArgGroupModel)
			fmt.Fprintln(w)
			fmt.Fprint(w, Wrap(width, 4, cmd.Help))
			fmt.Fprintln(w)
		}
	}

	if ctx.Context.SelectedCommand != nil {
		if ctx.Context.SelectedCommand.CmdGroupModel != nil && len(ctx.Context.SelectedCommand.Commands) > 0 {
			fmt.Fprintln(w, "Subcommands:")
			writeCommands(ctx.Context.SelectedCommand.FlattenedCommands())
			fmt.Fprintln(w)
		}
	} else if ctx.App.CmdGroupModel != nil && len(ctx.App.Commands) > 0 {
		fmt.Fprintln(w, "Commands:")
		writeCommands(ctx.App.FlattenedCommands())
		fmt.Fprintln(w)
	}

	return nil
}

// RenderSeparateOptionalFlags renders usage with required and optional flags listed separately.
func RenderSeparateOptionalFlags(w io.Writer, ctx *UsageContext) error {
	width := ctx.Width

	fmtTwoColumns := func(rows [][2]string) string {
		var buf bytes.Buffer
		FormatTwoColumns(&buf, ctx.Indent, ctx.Indent, width, rows)
		return buf.String()
	}

	requiredFlags := func(f []*FlagModel) []*FlagModel {
		var out []*FlagModel
		for _, flag := range f {
			if flag.Required {
				out = append(out, flag)
			}
		}
		return out
	}

	optionalFlags := func(f []*FlagModel) []*FlagModel {
		var out []*FlagModel
		for _, flag := range f {
			if !flag.Required {
				out = append(out, flag)
			}
		}
		return out
	}

	if ctx.Context.SelectedCommand != nil {
		cmd := ctx.Context.SelectedCommand
		fmt.Fprintf(w, "usage: %s %s", ctx.App.Name, cmd.String())
		WriteFormatUsage(w, cmd.FlagGroupModel, cmd.ArgGroupModel, cmd.CmdGroupModel, cmd.Help, width)
	} else {
		fmt.Fprintf(w, "usage: %s", ctx.App.Name)
		WriteFormatUsage(w, ctx.App.FlagGroupModel, ctx.App.ArgGroupModel, ctx.App.CmdGroupModel, ctx.App.Help, width)
	}
	fmt.Fprintln(w)

	if ctx.Context.Flags != nil {
		if rf := requiredFlags(ctx.Context.Flags); len(rf) > 0 {
			fmt.Fprintln(w, "Required flags:")
			fmt.Fprint(w, fmtTwoColumns(FlagsToTwoColumns(rf)))
			fmt.Fprintln(w)
		}
		if of := optionalFlags(ctx.Context.Flags); len(of) > 0 {
			fmt.Fprintln(w, "Optional flags:")
			fmt.Fprint(w, fmtTwoColumns(FlagsToTwoColumns(of)))
			fmt.Fprintln(w)
		}
	}
	if len(ctx.Context.Args) > 0 {
		fmt.Fprintln(w, "Args:")
		fmt.Fprint(w, fmtTwoColumns(ArgsToTwoColumns(ctx.Context.Args)))
		fmt.Fprintln(w)
	}

	writeCommands := func(commands []*CmdModel) {
		for _, cmd := range commands {
			if cmd.Hidden {
				continue
			}

			fmt.Fprint(w, cmd.FullCommand)
			if cmd.Default {
				fmt.Fprint(w, "*")
			}
			WriteFormatCommand(w, cmd.FlagGroupModel, cmd.ArgGroupModel)
			fmt.Fprintln(w)
			fmt.Fprint(w, Wrap(width, 4, cmd.Help))
			fmt.Fprintln(w)
		}
	}

	if ctx.Context.SelectedCommand != nil {
		fmt.Fprintln(w, "Subcommands:")
		if ctx.Context.SelectedCommand.CmdGroupModel != nil && len(ctx.Context.SelectedCommand.Commands) > 0 {
			writeCommands(ctx.Context.SelectedCommand.FlattenedCommands())
			fmt.Fprintln(w)
		}
	} else if ctx.App.CmdGroupModel != nil && len(ctx.App.Commands) > 0 {
		fmt.Fprintln(w, "Commands:")
		writeCommands(ctx.App.FlattenedCommands())
		fmt.Fprintln(w)
	}

	return nil
}

// RenderCompact renders usage with a hierarchical command list.
func RenderCompact(w io.Writer, ctx *UsageContext) error {
	width := ctx.Width

	fmtTwoColumns := func(rows [][2]string) string {
		var buf bytes.Buffer
		FormatTwoColumns(&buf, ctx.Indent, ctx.Indent, width, rows)
		return buf.String()
	}

	var writeCommandList func(commands []*CmdModel)
	writeCommandList = func(commands []*CmdModel) {
		for _, cmd := range commands {
			if cmd.Hidden {
				continue
			}

			fmt.Fprintf(w, "%s%s", strings.Repeat(" ", cmd.Depth*ctx.Indent), cmd.Name)
			if cmd.Default {
				fmt.Fprint(w, "*")
			}
			WriteFormatCommand(w, cmd.FlagGroupModel, cmd.ArgGroupModel)
			fmt.Fprintln(w)
			if cmd.CmdGroupModel != nil {
				writeCommandList(cmd.Commands)
			}
		}
	}

	if ctx.Context.SelectedCommand != nil {
		cmd := ctx.Context.SelectedCommand
		fmt.Fprintf(w, "usage: %s %s", ctx.App.Name, cmd.String())
		WriteFormatUsage(w, cmd.FlagGroupModel, cmd.ArgGroupModel, cmd.CmdGroupModel, cmd.Help, width)
	} else {
		fmt.Fprintf(w, "usage: %s", ctx.App.Name)
		WriteFormatUsage(w, ctx.App.FlagGroupModel, ctx.App.ArgGroupModel, ctx.App.CmdGroupModel, ctx.App.Help, width)
	}

	fmt.Fprintln(w)

	if len(ctx.Context.Flags) > 0 {
		fmt.Fprintln(w, "Flags:")
		fmt.Fprint(w, fmtTwoColumns(FlagsToTwoColumns(ctx.Context.Flags)))
		fmt.Fprintln(w)
	}
	if len(ctx.Context.Args) > 0 {
		fmt.Fprintln(w, "Args:")
		fmt.Fprint(w, fmtTwoColumns(ArgsToTwoColumns(ctx.Context.Args)))
		fmt.Fprintln(w)
	}

	if ctx.Context.SelectedCommand != nil {
		if ctx.Context.SelectedCommand.CmdGroupModel != nil && len(ctx.Context.SelectedCommand.Commands) > 0 {
			fmt.Fprintln(w, "Commands:")
			fmt.Fprintf(w, "  %s\n", ctx.Context.SelectedCommand.String())
			writeCommandList(ctx.Context.SelectedCommand.Commands)
			fmt.Fprintln(w)
		}
	} else if ctx.App.CmdGroupModel != nil && len(ctx.App.Commands) > 0 {
		fmt.Fprintln(w, "Commands:")
		writeCommandList(ctx.App.Commands)
		fmt.Fprintln(w)
	}

	return nil
}

// RenderManPage renders usage as a Unix man page.
func RenderManPage(w io.Writer, ctx *UsageContext) error {
	writeFormatFlags := func(fg *FlagGroupModel) {
		if fg == nil {
			return
		}
		for _, flag := range fg.Flags {
			if flag.Hidden {
				continue
			}

			fmt.Fprintln(w, ".TP")
			fmt.Fprint(w, `\fB`)
			if flag.Short != 0 {
				fmt.Fprintf(w, "-%s, ", string(flag.Short))
			}
			fmt.Fprintf(w, "--%s", flag.Name)
			if !flag.IsBoolFlag() {
				fmt.Fprintf(w, "=%s", flag.FormatPlaceHolder())
			}
			fmt.Fprintln(w, `\fR`)
			fmt.Fprintln(w, flag.Help)
		}
	}

	// Header
	fmt.Fprintf(w, ".TH %s 1 %s \"%s\"\n", ctx.App.Name, ctx.App.Version, ctx.App.Author)
	fmt.Fprintln(w, `.SH "NAME"`)
	fmt.Fprintln(w, ctx.App.Name)
	fmt.Fprintln(w, `.SH "SYNOPSIS"`)
	fmt.Fprintln(w, ".TP")
	fmt.Fprintf(w, `\fB%s`, ctx.App.Name)
	WriteFormatCommand(w, ctx.App.FlagGroupModel, ctx.App.ArgGroupModel)
	if ctx.App.CmdGroupModel != nil && len(ctx.App.Commands) > 0 {
		fmt.Fprint(w, " <command> [<args> ...]")
	}
	fmt.Fprintln(w, `\fR`)
	fmt.Fprintln(w)
	fmt.Fprintln(w, `.SH "DESCRIPTION"`)
	fmt.Fprintln(w, ctx.App.Help)
	fmt.Fprintln(w, `.SH "OPTIONS"`)
	writeFormatFlags(ctx.App.FlagGroupModel)

	if ctx.App.CmdGroupModel != nil && len(ctx.App.Commands) > 0 {
		fmt.Fprintln(w, `.SH "COMMANDS"`)
		for _, cmd := range ctx.App.FlattenedCommands() {
			if cmd.Hidden {
				continue
			}

			fmt.Fprintln(w, ".SS")
			fmt.Fprintf(w, `\fB%s`, cmd.FullCommand)
			WriteFormatCommand(w, cmd.FlagGroupModel, cmd.ArgGroupModel)
			fmt.Fprintln(w, `\fR`)
			fmt.Fprintln(w, ".PP")
			fmt.Fprintln(w, cmd.Help)
			writeFormatFlags(cmd.FlagGroupModel)
		}
	}

	return nil
}

// RenderLongHelp renders verbose usage with per-command flags.
func RenderLongHelp(w io.Writer, ctx *UsageContext) error {
	width := ctx.Width

	fmtTwoColumns := func(rows [][2]string) string {
		var buf bytes.Buffer
		FormatTwoColumns(&buf, ctx.Indent, ctx.Indent, width, rows)
		return buf.String()
	}

	fmt.Fprintf(w, "usage: %s", ctx.App.Name)
	WriteFormatUsage(w, ctx.App.FlagGroupModel, ctx.App.ArgGroupModel, ctx.App.CmdGroupModel, ctx.App.Help, width)
	fmt.Fprintln(w)

	if len(ctx.Context.Flags) > 0 {
		fmt.Fprintln(w, "Flags:")
		fmt.Fprint(w, fmtTwoColumns(FlagsToTwoColumns(ctx.Context.Flags)))
		fmt.Fprintln(w)
	}
	if len(ctx.Context.Args) > 0 {
		fmt.Fprintln(w, "Args:")
		fmt.Fprint(w, fmtTwoColumns(ArgsToTwoColumns(ctx.Context.Args)))
		fmt.Fprintln(w)
	}

	if ctx.App.CmdGroupModel != nil && len(ctx.App.Commands) > 0 {
		fmt.Fprintln(w, "Commands:")
		for _, cmd := range ctx.App.FlattenedCommands() {
			if cmd.Hidden {
				continue
			}

			fmt.Fprint(w, cmd.FullCommand)
			WriteFormatCommand(w, cmd.FlagGroupModel, cmd.ArgGroupModel)
			fmt.Fprintln(w)
			fmt.Fprint(w, Wrap(width, 4, cmd.Help))
			if cmd.FlagGroupModel != nil {
				rows := FlagsToTwoColumns(cmd.Flags)
				if len(rows) > 0 {
					var buf bytes.Buffer
					FormatTwoColumns(&buf, 4, 2, width, rows)
					fmt.Fprint(w, buf.String())
				}
			}
			fmt.Fprint(w, "\n\n")
		}
		fmt.Fprint(w, "\n")
	}

	return nil
}

// RenderBashCompletion renders a bash completion script.
func RenderBashCompletion(w io.Writer, ctx *UsageContext) error {
	name := ctx.App.Name
	_, err := fmt.Fprintf(w, `
_%[1]s_bash_autocomplete() {
    local cur prev opts base
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    opts=$( ${COMP_WORDS[0]} --completion-bash "${COMP_WORDS[@]:1:$COMP_CWORD}" )
    COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
    return 0
}
complete -F _%[1]s_bash_autocomplete -o default %[1]s

`, name)
	if err != nil {
		return err
	}

	return nil
}

// RenderZshCompletion renders a zsh completion script.
func RenderZshCompletion(w io.Writer, ctx *UsageContext) error {
	name := ctx.App.Name
	_, err := fmt.Fprintf(w, `#compdef %[1]s

_%[1]s() {
    local matches=($(${words[1]} --completion-bash "${(@)words[2,$CURRENT]}"))
    compadd -a matches

    if [[ $compstate[nmatches] -eq 0 && $words[$CURRENT] != -* ]]; then
        _files
    fi
}

if [[ "$(basename -- ${(%%):-%%x})" != "_%[1]s" ]]; then
    compdef _%[1]s %[1]s
fi
`, name)
	if err != nil {
		return err
	}

	return nil
}

func RenderFishCompletion(w io.Writer, ctx *UsageContext) error {
	name := ctx.App.Name
	_, err := fmt.Fprintf(w, `complete -c %[1]s -f -a '(
    set -l tokens (commandline -xpc)
    set -l current (commandline -ct)
    set -l completions (%[1]s --completion-bash $tokens[2..] $current)
    if test -n "$completions"
        printf "%%s\n" $completions
    else
        __fish_complete_path $current
    end
)'
`, name)
	if err != nil {
		return err
	}

	return nil
}

// FlagsToTwoColumns converts a slice of flags into two-column row data
// suitable for FormatTwoColumns. Hidden flags are excluded.
func FlagsToTwoColumns(f []*FlagModel) [][2]string {
	var rows [][2]string
	haveShort := ShortFlagsPresent(f)
	for _, flag := range f {
		if !flag.Hidden {
			rows = append(rows, [2]string{FormatFlag(haveShort, flag), flag.HelpWithEnvar()})
		}
	}
	return rows
}

// ArgsToTwoColumns converts a slice of args into two-column row data
// suitable for FormatTwoColumns. Hidden args are excluded.
func ArgsToTwoColumns(a []*ArgModel) [][2]string {
	var rows [][2]string
	for _, arg := range a {
		if arg.Hidden {
			continue
		}

		s := "<" + arg.Name + ">"
		if arg.PlaceHolder != "" {
			s = arg.PlaceHolder
		}
		if !arg.Required {
			s = "[" + s + "]"
		}
		rows = append(rows, [2]string{s, arg.HelpWithEnvar()})
	}
	return rows
}

// Wrap wraps text at the given width with the given indent level, using go/doc formatting.
func Wrap(width, indent int, s string) string {
	var buf bytes.Buffer
	indentText := strings.Repeat(" ", indent)
	doc.ToText(&buf, s, indentText, "  "+indentText, width-indent)
	return buf.String()
}

// ShortFlagsPresent checks if any flags in the slice have a short form.
func ShortFlagsPresent(f []*FlagModel) bool {
	for _, flag := range f {
		if flag.Short != 0 {
			return true
		}
	}
	return false
}
