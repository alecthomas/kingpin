// Package commander is used to manage a set of command-line "commands", with
// per-command flags and arguments.
//
// Supports command like so:
//
//   <command> <required> [<optional> [<optional> ...]]
//   <command> <remainder...>
//
// eg.
//
//   register [--name <name>] <nick>|<id>
//   post --channel|-a <channel> [--image <image>] [<text>]
//
// var (
//   chat = commander.New()
//   debug = chat.Flag("debug", "enable debug mode").Default("false").Bool()
//
//   register = chat.Command("register", "Register a new user.")
//   registerName = register.Flag("name", "name of user").Required().String()
//   registerNick = register.Arg("nick", "nickname for user").Required().String()
//
//   post = chat.Command("post", "Post a message to a channel.")
//   postChannel = post.Flag("channel", "channel to post to").Short('a').Required().String()
//   postImage = post.Flag("image", "image to post").String()
// )
//

package kingpin

import (
	"fmt"
	"io"
	"os"
)

type Dispatch func() error

// An Application contains the definitions of flags, arguments and commands
// for an application.
type Application struct {
	*flagGroup
	*argGroup
	Name         string
	Help         string
	commands     map[string]*CmdClause
	commandOrder []*CmdClause
	commandHelp  *string
}

// New creates a new Kingpin application instance.
func New(name, help string) *Application {
	a := &Application{
		flagGroup: newFlagGroup(),
		argGroup:  newArgGroup(),
		Name:      name,
		Help:      help,
		commands:  make(map[string]*CmdClause),
	}
	a.Flag("help", "Show help.").Dispatch(a.onFlagHelp).Bool()
	return a
}

func (a *Application) onFlagHelp() error {
	a.Usage(os.Stderr)
	os.Exit(0)
	return nil
}

// Command adds a new top-level command to the application.
func (a *Application) Command(name, help string) *CmdClause {
	cmd := newCommand(a, name, help)
	a.commands[name] = cmd
	a.commandOrder = append(a.commandOrder, cmd)
	return cmd
}

// Parse parses command-line arguments.
func (a *Application) Parse(args []string) (command string, err error) {
	if err := a.init(); err != nil {
		return "", err
	}
	tokens := Tokenize(args)
	tokens, command, err = a.parse(tokens)

	if len(tokens) == 1 {
		return "", fmt.Errorf("unexpected argument '%s'", tokens)
	} else if len(tokens) > 0 {
		return "", fmt.Errorf("unexpected arguments '%s'", tokens)
	}

	return command, err
}

// Version adds a --version flag for displaying the application version.
func (a *Application) Version(version string) *Application {
	a.Flag("version", "Show application version.").Dispatch(func() error {
		fmt.Println(version)
		os.Exit(0)
		return nil
	}).Bool()
	return a
}

func (a *Application) init() error {
	if len(a.commands) > 0 && len(a.args) > 0 {
		return fmt.Errorf("can't mix top-level Arg()s with Command()s")
	}
	if len(a.commands) > 0 {
		cmd := a.Command("help", "Show help for a command.")
		a.commandHelp = cmd.Arg("command", "Command name.").Required().Dispatch(a.onCommandHelp).String()
		// Make "help" command first in order. Also, Go's slice operations are woeful.
		a.commandOrder = append(a.commandOrder[len(a.commandOrder)-1:], a.commandOrder[:len(a.commandOrder)-1]...)
	}
	if err := a.flagGroup.init(); err != nil {
		return err
	}
	if err := a.argGroup.init(); err != nil {
		return err
	}
	for _, cmd := range a.commands {
		if err := cmd.init(); err != nil {
			return err
		}
	}
	return nil
}

func (a *Application) onCommandHelp() error {
	a.CommandUsage(os.Stderr, *a.commandHelp)
	os.Exit(0)
	return nil
}

func (a *Application) parse(tokens tokens) (tokens, string, error) {
	// Special-case "help" to avoid issues with required flags.
	runHelp := (tokens.Peek().Value == "help")

	var err error
	var token *token
	tokens, err = a.flagGroup.parse(tokens, runHelp)
	if err != nil {
		return tokens, "", err
	}

	selected := ""

	// Parse arguments or commands.
	if len(a.args) > 0 {
		tokens, err = a.argGroup.parse(tokens)
	} else {
		token, tokens = tokens.Next()
		switch token.Type {
		case TokenArg:
			cmd, ok := a.commands[token.Value]
			if !ok {
				return tokens, "", fmt.Errorf("unknown command '%s'", token)
			}
			tokens, err = cmd.parse(tokens)
			if err != nil {
				return tokens, "", err
			}
			selected = cmd.name

		default:
		}
	}
	return tokens, selected, nil
}

// Errorf prints an error message to w.
func (a *Application) Errorf(w io.Writer, format string, args ...interface{}) {
	fmt.Fprintf(w, a.Name+": error: "+format+"\n", args...)
}

// UsageErrorf prints an error message followed by usage information, then
// exits with a non-zero status.
func (a *Application) UsageErrorf(w io.Writer, format string, args ...interface{}) {
	a.Errorf(w, format, args...)
	a.Usage(w)
	os.Exit(1)
}

// FatalIfError prints an error and exits if err is not nil. The error is printed
// with the given prefix.
func (a *Application) FatalIfError(w io.Writer, err error, prefix string) {
	if err != nil {
		if prefix != "" {
			prefix += ": "
		}
		a.Errorf(w, prefix+"%s", err)
		os.Exit(1)
	}
}
