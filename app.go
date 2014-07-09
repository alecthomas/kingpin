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
//   post --channel|-c <channel> [--image <image>] [<text>]
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
//   postChannel = post.Flag("channel", "channel to post to").Short('c').Required().String()
//   postImage = post.Flag("image", "image to post").String()
// )
//

package kingpin

import (
	"fmt"
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
	c := &Application{
		flagGroup: newFlagGroup(),
		argGroup:  newArgGroup(),
		Name:      name,
		Help:      help,
		commands:  make(map[string]*CmdClause),
	}
	c.Flag("help", "Show help.").Dispatch(c.onFlagHelp).Bool()
	return c
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

func (c *Application) init() error {
	if len(c.commands) > 0 && len(c.args) > 0 {
		return fmt.Errorf("can't mix top-level Arg()s with Command()s")
	}
	if len(c.commands) > 0 {
		cmd := c.Command("help", "Show help for a command.")
		c.commandHelp = cmd.Arg("command", "Command name.").Required().Dispatch(c.onCommandHelp).String()
		// Make "help" command first in order. Also, Go's slice operations are woeful.
		c.commandOrder = append(c.commandOrder[len(c.commandOrder)-1:], c.commandOrder[:len(c.commandOrder)-1]...)
	}
	if err := c.flagGroup.init(); err != nil {
		return err
	}
	if err := c.argGroup.init(); err != nil {
		return err
	}
	for _, cmd := range c.commands {
		if err := cmd.init(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Application) onCommandHelp() error {
	c.CommandUsage(os.Stderr, *c.commandHelp)
	os.Exit(0)
	return nil
}

func (c *Application) parse(tokens tokens) (tokens, string, error) {
	// Special-case "help" to avoid issues with required flags.
	runHelp := (tokens.Peek().Value == "help")

	var err error
	var token *token
	tokens, err = c.flagGroup.parse(tokens, runHelp)
	if err != nil {
		return tokens, "", err
	}

	selected := ""

	// Parse arguments or commands.
	if len(c.args) > 0 {
		tokens, err = c.argGroup.parse(tokens)
	} else {
		token, tokens = tokens.Next()
		switch token.Type {
		case TokenArg:
			cmd, ok := c.commands[token.Value]
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
