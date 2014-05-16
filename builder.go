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

type Commander struct {
	*flagGroup
	*argGroup
	name         string
	help         string
	commands     map[string]*CmdClause
	commandOrder []*CmdClause
	commandHelp  *string
}

func New(name, help string) *Commander {
	c := &Commander{
		flagGroup: newFlagGroup(),
		argGroup:  newArgGroup(),
		name:      name,
		help:      help,
		commands:  make(map[string]*CmdClause),
	}
	c.Flag("help", "Show help.").Dispatch(c.onFlagHelp).Bool()
	return c
}

func (c *Commander) onFlagHelp() error {
	c.Usage(os.Stderr)
	os.Exit(0)
	return nil
}

func (c *Commander) Command(name, help string) *CmdClause {
	cmd := newCommand(name, help)
	c.commands[name] = cmd
	c.commandOrder = append(c.commandOrder, cmd)
	return cmd
}

func (c *Commander) Parse(args []string) (command string, err error) {
	c.init()
	tokens := Tokenize(args)
	return c.parse(tokens)
}

func (c *Commander) init() {
	if len(c.commands) > 0 && len(c.args) > 0 {
		panic("can't mix top-level Arg()s with Command()s")
	}
	if len(c.commands) > 0 {
		cmd := c.Command("help", "Show help for a command.")
		c.commandHelp = cmd.Arg("command", "Command name.").Required().Dispatch(c.onCommandHelp).String()
		// Make "help" command first in order. Also, Go's slice operations are woeful.
		c.commandOrder = append(c.commandOrder[len(c.commandOrder)-1:], c.commandOrder[:len(c.commandOrder)-1]...)
	}
	c.flagGroup.init()
	c.argGroup.init()
	for _, cmd := range c.commands {
		cmd.init()
	}
}

func (c *Commander) onCommandHelp() error {
	c.CommandUsage(os.Stderr, *c.commandHelp)
	os.Exit(0)
	return nil
}

func (c *Commander) parse(tokens tokens) (command string, err error) {
	// Special-case "help" to avoid issues with required flags.
	runHelp := (tokens.Peek().Value == "help")

	tokens, err = c.flagGroup.parse(tokens, runHelp)
	if err != nil {
		return "", err
	}

	if len(c.args) > 0 {
		tokens, err = c.argGroup.parse(tokens)
		if err != nil {
			return "", err
		}
	} else {
		token, tokens := tokens.Next()
		switch token.Type {
		case TokenArg:
			cmd, ok := c.commands[token.Value]
			if !ok {
				return "", fmt.Errorf("unknown command '%s'", token)
			}
			tokens, err = cmd.parse(tokens)
			if err != nil {
				return "", err
			}
			return cmd.name, nil

		case TokenEOF:
		default:
			return "", fmt.Errorf("unexpected '%s'", token)
		}
	}
	return "", nil
}
