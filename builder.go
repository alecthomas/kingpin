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
	"strings"
)

type Dispatch func() error

type Flags struct {
	short map[string]*FlagClause
	long  map[string]*FlagClause
}

// Flag defines a new flag with the given long name and help.
func (f *Flags) Flag(name, help string) *FlagClause {
	flag := newFlag(name, help)
	f.long[name] = flag
	return flag
}

func (f *Flags) init() {
	for _, flag := range f.long {
		flag.init()
		if flag.Shorthand != 0 {
			f.short[string(flag.Shorthand)] = flag
		}
	}
}

func (f *Flags) parse(tokens Tokens) (Tokens, error) {
	remaining := make(map[string]struct{})
	for k, flag := range f.long {
		if flag.required {
			remaining[k] = struct{}{}
		}
	}

	var token *Token

loop:
	for {
		token, tokens = tokens.Next()
		switch token.Type {
		case TokenEOF:
			break loop

		case TokenLong, TokenShort:
			flagToken := token
			DefValue := ""
			var flag *FlagClause
			var ok bool

			if token.Type == TokenLong {
				flag, ok = f.long[token.Value]
				if !ok {
					flag, ok = f.long["no-"+token.Value]
					if !ok {
						return nil, fmt.Errorf("unknown long flag '%s'", flagToken)
					}
					DefValue = "false"
				}
			} else {
				flag, ok = f.short[token.Value]
				if !ok {
					return nil, fmt.Errorf("unknown short flag '%s", flagToken)
				}
			}

			delete(remaining, flag.Name)

			if !flag.boolean {
				token, tokens = tokens.Next()
				if token.Type != TokenArg {
					return nil, fmt.Errorf("expected argument for flag '%s'", flagToken)
				}
				DefValue = token.Value
			}

			if err := flag.parser(DefValue); err != nil {
				return nil, err
			}

			if flag.dispatch != nil {
				if err := flag.dispatch(); err != nil {
					return nil, err
				}
			}

		default:
			tokens = tokens.Return(token)
			break loop
		}
	}

	// Check that required flags were provided.
	if len(remaining) == 1 {
		for k := range remaining {
			return nil, fmt.Errorf("required flag --%s not provided", k)
		}
	} else if len(remaining) > 1 {
		flags := make([]string, 0, len(remaining))
		for k := range remaining {
			flags = append(flags, "--"+k)
		}
		return nil, fmt.Errorf("required flags %s not provided", strings.Join(flags, ", "))
	}
	return tokens, nil
}

func makeFlags() Flags {
	return Flags{
		short: make(map[string]*FlagClause),
		long:  make(map[string]*FlagClause),
	}
}

type Commander struct {
	Flags
	Name        string
	Help        string
	commands    map[string]*CmdClause
	commandHelp *string
}

func New(name, help string) *Commander {
	c := &Commander{
		Flags:    makeFlags(),
		Name:     name,
		Help:     help,
		commands: make(map[string]*CmdClause),
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
	return cmd
}

func (c *Commander) Parse(args []string) (command string, err error) {
	c.init()
	tokens := Tokenize(args)
	return c.parse(tokens)
}

func (c *Commander) init() {
	if len(c.commands) > 0 {
		cmd := c.Command("help", "Show help for a command.")
		c.commandHelp = cmd.Arg("command", "Command name.").Required().Dispatch(c.onCommandHelp).String()
	}
	c.Flags.init()
	for _, cmd := range c.commands {
		cmd.init()
	}
}

func (c *Commander) onCommandHelp() error {
	c.CommandUsage(os.Stderr, *c.commandHelp)
	os.Exit(0)
	return nil
}

func (c *Commander) parse(tokens Tokens) (command string, err error) {
	tokens, err = c.Flags.parse(tokens)
	if err != nil {
		return "", err
	}

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
		return cmd.Name, nil

	case TokenEOF:
		return "", nil

	default:
		return "", fmt.Errorf("unexpected '%s'", token)
	}
}

// FlagClause is a fluid interface used to build flags.
type FlagClause struct {
	parserMixin
	Name      string
	Shorthand byte
	Help      string
	DefValue  string
	metavar   string
	boolean   bool
	dispatch  Dispatch
}

func newFlag(name, help string) *FlagClause {
	f := &FlagClause{
		Name: name,
		Help: help,
	}
	return f
}

func (f *FlagClause) formatMetaVar() string {
	if f.metavar != "" {
		return f.metavar
	}
	return strings.ToUpper(f.Name)
}

func (f *FlagClause) init() {
	if f.parser == nil {
		panic(fmt.Sprintf("no parser defined for --%s", f.Name))
	}
	if f.DefValue != "" {
		if err := f.parser(f.DefValue); err != nil {
			panic(fmt.Sprintf("default value for --%s is invalid: %s", f.Name, err))
		}
	}
}

// Dispatch to the given function when the flag is parsed.
func (f *FlagClause) Dispatch(dispatch Dispatch) *FlagClause {
	f.dispatch = dispatch
	return f
}

// Default value for this flag.
func (f *FlagClause) Default(value string) *FlagClause {
	f.DefValue = value
	return f
}

// MetaVar sets the placeholder string used for flag values in the help.
func (f *FlagClause) MetaVar(metavar string) *FlagClause {
	f.metavar = metavar
	return f
}

// Required makes the flag required.
func (f *FlagClause) Required() *FlagClause {
	f.required = true
	return f
}

// Short sets the short flag name.
func (f *FlagClause) Short(name byte) *FlagClause {
	f.Shorthand = name
	return f
}

func (f *FlagClause) Bool() (target *bool) {
	return BoolParser(f)
}

// SetIsBoolean tells the parser that this is a boolean flag. Typically only
// used by Parser implementations.
func (f *FlagClause) SetIsBoolean() {
	f.boolean = true
}

type ArgClause struct {
	parserMixin
	name     string
	help     string
	DefValue string
	required bool
	dispatch Dispatch
}

func newArg(name, help string) *ArgClause {
	a := &ArgClause{
		name: name,
		help: help,
	}
	return a
}

func (a *ArgClause) Required() *ArgClause {
	a.required = true
	return a
}

func (a *ArgClause) Default(value string) *ArgClause {
	a.DefValue = value
	return a
}

func (a *ArgClause) Dispatch(dispatch Dispatch) *ArgClause {
	a.dispatch = dispatch
	return a
}

func (a *ArgClause) init() {
	if a.parser == nil {
		panic(fmt.Sprintf("no parser defined for arg '%s'", a.name))
	}
	if a.DefValue != "" {
		if err := a.parser(a.DefValue); err != nil {
			panic(fmt.Sprintf("invalid default value '%s' for argument '%s'", a.DefValue, a.name))
		}
	}
}

func (a *ArgClause) parse(tokens Tokens) (Tokens, error) {
	if token, tokens := tokens.Next(); token.Type == TokenArg {
		if err := a.parser(token.Value); err != nil {
			return nil, err
		}
		if a.dispatch != nil {
			if err := a.dispatch(); err != nil {
				return nil, err
			}
		}
		return tokens, nil
	} else {
		return tokens.Return(token), nil
	}
}

type CmdClause struct {
	Flags
	Name     string
	Help     string
	args     []*ArgClause
	dispatch Dispatch
}

func newCommand(name, help string) *CmdClause {
	return &CmdClause{
		Flags: makeFlags(),
		Name:  name,
		Help:  help,
	}
}

func (c *CmdClause) Dispatch(dispatch Dispatch) *CmdClause {
	c.dispatch = dispatch
	return c
}

func (c *CmdClause) Arg(name, help string) *ArgClause {
	arg := newArg(name, help)
	c.args = append(c.args, arg)
	return arg
}

func (c *CmdClause) init() {
	c.Flags.init()
	required := 0
	seen := map[string]struct{}{}
	for i, arg := range c.args {
		if _, ok := seen[arg.name]; ok {
			panic(fmt.Sprintf("duplicate argument '%s'", arg.name))
		}
		seen[arg.name] = struct{}{}
		if arg.required && required != i {
			panic("required arguments found after non-required")
		}
		if arg.required {
			required++
		}
		arg.init()
	}
}

func (c *CmdClause) parse(tokens Tokens) (Tokens, error) {
	tokens, err := c.Flags.parse(tokens)
	if err != nil {
		return nil, err
	}
	for _, arg := range c.args {
		token := tokens.Peek()
		if token.IsFlag() {
			return nil, fmt.Errorf("unknown flag '%s'", token)
		}
		if token.Type != TokenArg {
			if arg.required {
				return nil, fmt.Errorf("'%s' is required", arg.name)
			}
			break
		}

		tokens, err = arg.parse(tokens)
		if err != nil {
			return nil, err
		}
	}

	if c.dispatch != nil {
		return tokens, c.dispatch()
	}
	return tokens, nil
}
