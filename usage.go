package kingpin

import (
	"bytes"
	"fmt"
	"go/doc"
	"io"
	"strings"
)

func formatTwoColumns(w io.Writer, indent, padding, width int, rows [][2]string) {
	// Find size of first column.
	s := 0
	for _, row := range rows {
		if c := len(row[0]); c > s && c < 20 {
			s = c
		}
	}

	indentStr := strings.Repeat(" ", indent)
	offsetStr := strings.Repeat(" ", s+padding)

	for _, row := range rows {
		buf := bytes.NewBuffer(nil)
		doc.ToText(buf, row[1], "", "", width-s-padding-indent)
		lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
		fmt.Fprintf(w, "%s%-*s%*s", indentStr, s, row[0], padding, "")
		if len(row[0]) > 20 {
			fmt.Fprintf(w, "\n%s%s", indentStr, offsetStr)
		}
		fmt.Fprintf(w, "%s\n", lines[0])
		for _, line := range lines[1:] {
			fmt.Fprintf(w, "%s%s%s\n", indentStr, offsetStr, line)
		}
	}
}

func (c *Application) Usage(w io.Writer) {
	c.writeHelp(guessWidth(w), w)
}

func (c *Application) CommandUsage(w io.Writer, command string) {
	cmd, ok := c.commands[command]
	if !ok {
		Fatalf("unknown command '%s'", command)
	}
	s := []string{formatArgsAndFlags(c.Name, c.argGroup, c.flagGroup)}
	s = append(s, formatArgsAndFlags(cmd.name, cmd.argGroup, cmd.flagGroup))
	fmt.Fprintf(w, "usage: %s\n", strings.Join(s, " "))
	if cmd.help != "" {
		fmt.Fprintf(w, "\n%s\n", cmd.help)
	}
	cmd.writeHelp(guessWidth(w), w)
}

func (c *Application) writeHelp(width int, w io.Writer) {
	s := []string{formatArgsAndFlags(c.Name, c.argGroup, c.flagGroup)}
	if len(c.commands) > 0 {
		s = append(s, "<command>", "[<flags>]", "[<args> ...]")
	}

	prefix := "usage: "
	usage := strings.Join(s, " ")
	buf := bytes.NewBuffer(nil)
	doc.ToText(buf, usage, "", "", width-len(prefix))
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")

	fmt.Fprintf(w, "%s%s\n", prefix, lines[0])
	for _, l := range lines[1:] {
		fmt.Fprintf(w, "%*s%s\n", len(prefix), "", l)
	}
	if c.Help != "" {
		fmt.Fprintf(w, "\n")
		doc.ToText(w, c.Help, "", "", width)
	}

	c.flagGroup.writeHelp(width, w)
	c.argGroup.writeHelp(width, w)

	if len(c.commands) > 0 {
		fmt.Fprintf(w, "\nCommands:\n")
		c.helpCommands(width, w)
	}
}

func (c *Application) helpCommands(width int, w io.Writer) {
	for _, cmd := range c.commandOrder {
		fmt.Fprintf(w, "  %s\n", formatArgsAndFlags(cmd.name, cmd.argGroup, cmd.flagGroup))
		buf := bytes.NewBuffer(nil)
		doc.ToText(buf, cmd.help, "", "", width-4)
		lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
		for _, line := range lines {
			fmt.Fprintf(w, "    %s\n", line)
		}
		fmt.Fprintf(w, "\n")
	}
}

func (f *flagGroup) writeHelp(width int, w io.Writer) {
	if f.visibleFlags() == 0 {
		return
	}

	fmt.Fprintf(w, "\nFlags:\n")

	rows := [][2]string{}
	for _, flag := range f.flagOrder {
		rows = append(rows, [2]string{formatFlag(flag), flag.help})
	}
	formatTwoColumns(w, 2, 2, width, rows)
}

func (f *flagGroup) gatherFlagSummary() (out []string) {
	for _, flag := range f.flagOrder {
		if flag.required {
			fb, ok := flag.value.(boolFlag)
			if ok && fb.IsBoolFlag() {
				out = append(out, fmt.Sprintf("--%s", flag.name))
			} else {
				out = append(out, fmt.Sprintf("--%s=%s", flag.name, flag.formatPlaceHolder()))
			}
		}
	}
	if len(f.long) != len(out) {
		out = append(out, "[<flags>]")
	}
	return
}

func (a *argGroup) writeHelp(width int, w io.Writer) {
	if len(a.args) == 0 {
		return
	}

	fmt.Fprintf(w, "\nArgs:\n")

	rows := [][2]string{}
	for _, arg := range a.args {
		s := "<" + arg.name + ">"
		if !arg.required {
			s = "[" + s + "]"
		}
		rows = append(rows, [2]string{s, arg.help})
	}

	formatTwoColumns(w, 2, 2, width, rows)
}

func (c *CmdClause) writeHelp(width int, w io.Writer) {
	c.flagGroup.writeHelp(width, w)
	c.argGroup.writeHelp(width, w)
}

func formatArgsAndFlags(name string, args *argGroup, flags *flagGroup) string {
	s := []string{name}
	s = append(s, flags.gatherFlagSummary()...)
	depth := 0
	for _, arg := range args.args {
		h := "<" + arg.name + ">"
		if !arg.required {
			h = "[" + h
			depth++
		}
		s = append(s, h)
	}
	s[len(s)-1] = s[len(s)-1] + strings.Repeat("]", depth)
	return strings.Join(s, " ")
}

func formatFlag(flag *FlagClause) string {
	flagString := ""
	if flag.shorthand != 0 {
		flagString += fmt.Sprintf("-%c, ", flag.shorthand)
	}
	flagString += fmt.Sprintf("--%s", flag.name)
	fb, ok := flag.value.(boolFlag)
	if !ok || !fb.IsBoolFlag() {
		flagString += fmt.Sprintf("=%s", flag.formatPlaceHolder())
	}
	return flagString
}
