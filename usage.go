package kingpin

import (
	"bytes"
	"fmt"
	"io"
	"os"
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
	c.help(guessWidth(w), w)
}

func (c *Commander) CommandUsage(w io.Writer, command string) {
	cmd, ok := c.commands[command]
	if !ok {
		UsageErrorf("unknown command '%s'", command)
	}
	s := c.topHelp()
	s = append(s, formatCommand(cmd))
	fmt.Fprintf(w, "usage: %s\n", strings.Join(s, " "))
	if cmd.Help != "" {
		fmt.Fprintf(w, "\n%s\n", cmd.Help)
	}
	cmd.help(guessWidth(w), w)
}

func (c *Commander) topHelp() []string {
	s := []string{c.Name}
	if len(c.long) > 0 {
		s = append(s, c.gatherFlagSummary()...)
	}
	return s
}

func (c *Commander) help(width int, w io.Writer) {
	s := c.topHelp()
	if len(c.commands) > 0 {
		s = append(s, "<command>", "[<flags>]", "[<args> ...]")
	}

	helpSummary := ""
	if c.Help != "" {
		helpSummary = "\n\n" + c.Help
	}
	fmt.Fprintf(w, "usage: %s%s\n", strings.Join(s, " "), helpSummary)

	c.Flags.help(2, width, w)

	if len(c.commands) > 0 {
		fmt.Fprintf(w, "\nCommands:\n")
		c.helpCommands(width, w)
	}
}

func (c *Commander) helpCommands(width int, w io.Writer) {
	l := 0
	for _, cmd := range c.commands {
		if cl := len(formatCommand(cmd)); cl > l {
			l = cl
		}
	}

	l += 5
	indentStr := strings.Repeat(" ", l)

	for _, cmd := range c.commands {
		prefix := fmt.Sprintf("  %-*s", l-2, formatCommand(cmd))
		buf := bytes.NewBuffer(nil)
		doc.ToText(buf, cmd.Help, "", "", width-l)
		lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
		fmt.Fprintf(w, "%s%s\n", prefix, lines[0])
		for _, line := range lines[1:] {
			fmt.Fprintf(w, "%s%s\n", indentStr, line)
		}
	}
}

func (f *Flags) help(indent, width int, w io.Writer) {
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
		doc.ToText(buf, flag.Help, "", "", width-l)
		lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
		fmt.Fprintf(w, "%s%s\n", prefix, lines[0])
		for _, line := range lines[1:] {
			fmt.Fprintf(w, "%s%s\n", indentStr, line)
		}
	}
}

func (f *Flags) gatherFlagSummary() (out []string) {
	for _, flag := range f.long {
		if flag.required {
			if flag.boolean {
				out = append(out, fmt.Sprintf("--%s", flag.Name))
			} else {
				out = append(out, fmt.Sprintf("--%s=%s", flag.Name, flag.formatMetaVar()))
			}
		}
	}
	if len(f.long) != len(out) {
		out = append(out, "[<flags>]")
	}
	return
}

func (c *CmdClause) help(width int, w io.Writer) {
	c.Flags.help(2, width, w)
}

func formatCommand(cmd *CmdClause) string {
	s := []string{cmd.Name}
	s = append(s, cmd.gatherFlagSummary()...)
	for _, arg := range cmd.args {
		h := "<" + arg.name + ">"
		if !arg.required {
			break
		}
		s = append(s, h)
	}
	return strings.Join(s, " ")
}

func formatFlag(flag *FlagClause) string {
	flagString := ""
	if flag.Shorthand != 0 {
		flagString += fmt.Sprintf("-%c, ", flag.Shorthand)
	}
	flagString += fmt.Sprintf("--%s", flag.Name)
	if !flag.boolean {
		flagString += fmt.Sprintf("=%s", flag.formatMetaVar())
	}
	return flagString

}
