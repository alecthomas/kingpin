package kingpin

import (
	"fmt"
	"strings"
)

type flagGroup struct {
	short map[string]*FlagClause
	long  map[string]*FlagClause
}

func newFlagGroup() *flagGroup {
	return &flagGroup{
		short: make(map[string]*FlagClause),
		long:  make(map[string]*FlagClause),
	}
}

// Flag defines a new flag with the given long name and help.
func (f *flagGroup) Flag(name, help string) *FlagClause {
	flag := newFlag(name, help)
	f.long[name] = flag
	return flag
}

func (f *flagGroup) init() {
	for _, flag := range f.long {
		flag.init()
		if flag.Shorthand != 0 {
			f.short[string(flag.Shorthand)] = flag
		}
	}
}

func (f *flagGroup) parse(tokens tokens, ignoreRequired bool) (tokens, error) {
	// Track how many required flags we've seen.
	remaining := make(map[string]struct{})
	for k, flag := range f.long {
		if flag.required && !ignoreRequired {
			remaining[k] = struct{}{}
		}
	}

	var token *token

loop:
	for {
		token, tokens = tokens.Next()
		switch token.Type {
		case TokenEOF:
			break loop

		case TokenLong, TokenShort:
			flagToken := token
			defaultValue := ""
			var flag *FlagClause
			var ok bool

			if token.Type == TokenLong {
				flag, ok = f.long[token.Value]
				if !ok {
					flag, ok = f.long["no-"+token.Value]
					if !ok {
						return nil, fmt.Errorf("unknown long flag '%s'", flagToken)
					}
					defaultValue = "false"
				}
			} else {
				flag, ok = f.short[token.Value]
				if !ok {
					return nil, fmt.Errorf("unknown short flag '%s", flagToken)
				}
			}

			delete(remaining, flag.name)

			if !flag.boolean {
				token, tokens = tokens.Next()
				if token.Type != TokenArg {
					return nil, fmt.Errorf("expected argument for flag '%s'", flagToken)
				}
				defaultValue = token.Value
			}

			if err := flag.value.Set(defaultValue); err != nil {
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

// FlagClause is a fluid interface used to build flags.
type FlagClause struct {
	parserMixin
	name      string
	Shorthand byte
	help      string
	DefValue  string
	metavar   string
	boolean   bool
	dispatch  Dispatch
}

func newFlag(name, help string) *FlagClause {
	f := &FlagClause{
		name: name,
		help: help,
	}
	return f
}

func (f *FlagClause) formatMetaVar() string {
	if f.metavar != "" {
		return f.metavar
	}
	return strings.ToUpper(f.name)
}

func (f *FlagClause) init() {
	if f.required && f.DefValue != "" {
		panic(fmt.Sprintf("required flag '%s' with unusable default value", f.name))
	}
	if f.value == nil {
		panic(fmt.Sprintf("no value defined for --%s", f.name))
	}
	if f.DefValue != "" {
		if err := f.value.Set(f.DefValue); err != nil {
			panic(fmt.Sprintf("default value for --%s is invalid: %s", f.name, err))
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
	target = new(bool)
	f.boolean = true
	f.SetValue(newBoolValue(false, target))
	return
}
