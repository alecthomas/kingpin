package kingpin

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"go/doc"

	"github.com/kr/pty"
)

func guessWidth(w io.Writer) int {
	width := 80
	if width == 0 {
		if t, ok := w.(*os.File); ok {
			if _, cols, err := pty.Getsize(t); err == nil {
				width = cols
			}
		}
	}
	return width
}

func (c *Commander) Usage(w io.Writer) {
	c.writeHelp(guessWidth(w), w)
}

func (c *Commander) CommandUsage(w io.Writer, command string) {
	cmd, ok := c.commands[command]
	if !ok {
		UsageErrorf("unknown command '%s'", command)
	}
	s := []string{formatArgsAndFlags(c.name, c.argGroup, c.flagGroup, true)}
	s = append(s, formatArgsAndFlags(cmd.name, cmd.argGroup, cmd.flagGroup, true))
	fmt.Fprintf(w, "usage: %s\n", strings.Join(s, " "))
	if cmd.help != "" {
		fmt.Fprintf(w, "\n%s\n", cmd.help)
	}
	cmd.writeHelp(guessWidth(w), w)
}

func (c *Commander) writeHelp(width int, w io.Writer) {
	s := []string{formatArgsAndFlags(c.name, c.argGroup, c.flagGroup, true)}
	if len(c.commands) > 0 {
		s = append(s, "<command>", "[<flags>]", "[<args> ...]")
	}

	helpSummary := ""
	if c.help != "" {
		helpSummary = "\n\n" + c.help
	}
	fmt.Fprintf(w, "usage: %s%s\n", strings.Join(s, " "), helpSummary)

	c.flagGroup.writeHelp(2, width, w)
	c.argGroup.writeHelp(2, width, w)

	if len(c.commands) > 0 {
		fmt.Fprintf(w, "\nCommands:\n")
		c.helpCommands(width, w)
	}
}

func (c *Commander) helpCommands(width int, w io.Writer) {
	l := 0
	commands := []string{}
	for _, cmd := range c.commands {
		if cl := len(formatArgsAndFlags(cmd.name, cmd.argGroup, cmd.flagGroup, false)); cl > l {
			l = cl
		}
		commands = append(commands, cmd.name)
	}

	sort.Strings(commands)

	l += 5
	indentStr := strings.Repeat(" ", l)

	for _, name := range commands {
		cmd := c.commands[name]
		prefix := fmt.Sprintf("  %-*s", l-2, formatArgsAndFlags(cmd.name, cmd.argGroup, cmd.flagGroup, false))
		buf := bytes.NewBuffer(nil)
		doc.ToText(buf, cmd.help, "", "", width-l)
		lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
		fmt.Fprintf(w, "%s%s\n", prefix, lines[0])
		for _, line := range lines[1:] {
			fmt.Fprintf(w, "%s%s\n", indentStr, line)
		}
	}
}

func (f *flagGroup) writeHelp(indent, width int, w io.Writer) {
	if len(f.long) == 0 {
		return
	}

	fmt.Fprintf(w, "\nFlags:\n")
	l := 0
	for _, flag := range f.long {
		if fl := len(formatFlag(flag)); fl > l {
			l = fl
		}
	}

	l += 3 + indent

	indentStr := strings.Repeat(" ", l)

	for _, flag := range f.long {
		prefix := fmt.Sprintf("  %-*s", l-2, formatFlag(flag))
		buf := bytes.NewBuffer(nil)
		doc.ToText(buf, flag.help, "", "", width-l)
		lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
		fmt.Fprintf(w, "%s%s\n", prefix, lines[0])
		for _, line := range lines[1:] {
			fmt.Fprintf(w, "%s%s\n", indentStr, line)
		}
	}
}

func (f *flagGroup) gatherFlagSummary() (out []string) {
	for _, flag := range f.long {
		if flag.required {
			if flag.boolean {
				out = append(out, fmt.Sprintf("--%s", flag.name))
			} else {
				out = append(out, fmt.Sprintf("--%s=%s", flag.name, flag.formatMetaVar()))
			}
		}
	}
	if len(f.long) != len(out) {
		out = append(out, "[<flags>]")
	}
	return
}

func (a *argGroup) writeHelp(indent, width int, w io.Writer) {
	if len(a.args) == 0 {
		return
	}

	fmt.Fprintf(w, "\nArgs:\n")
	l := 0
	for _, arg := range a.args {
		if al := len(arg.name) + 2; al > l {
			l = al
			if !arg.required {
				l += 2
			}
		}
	}

	l += 3 + indent

	indentStr := strings.Repeat(" ", l)

	for _, arg := range a.args {
		argString := "<" + arg.name + ">"
		if !arg.required {
			argString = "[" + argString + "]"
		}
		prefix := fmt.Sprintf("  %-*s", l-2, argString)
		buf := bytes.NewBuffer(nil)
		doc.ToText(buf, arg.help, "", "", width-l)
		lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
		fmt.Fprintf(w, "%s%s\n", prefix, lines[0])
		for _, line := range lines[1:] {
			fmt.Fprintf(w, "%s%s\n", indentStr, line)
		}
	}

}

func (c *CmdClause) writeHelp(width int, w io.Writer) {
	c.flagGroup.writeHelp(2, width, w)
	c.argGroup.writeHelp(2, width, w)
}

func formatArgsAndFlags(name string, args *argGroup, flags *flagGroup, showOptionalArgs bool) string {
	s := []string{name}
	s = append(s, flags.gatherFlagSummary()...)
	depth := 0
	for _, arg := range args.args {
		h := "<" + arg.name + ">"
		if !arg.required {
			if showOptionalArgs {
				h = "[" + h
				depth++
			} else {
				break
			}
		}
		s = append(s, h)
	}
	s[len(s)-1] = s[len(s)-1] + strings.Repeat("]", depth)
	return strings.Join(s, " ")
}

func formatFlag(flag *FlagClause) string {
	flagString := ""
	if flag.Shorthand != 0 {
		flagString += fmt.Sprintf("-%c, ", flag.Shorthand)
	}
	flagString += fmt.Sprintf("--%s", flag.name)
	if !flag.boolean {
		flagString += fmt.Sprintf("=%s", flag.formatMetaVar())
	}
	return flagString
}
