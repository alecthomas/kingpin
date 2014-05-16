package kingpin

import (
	"fmt"
	"strings"
)

type flagGroup struct {
	short     map[string]*FlagClause
	long      map[string]*FlagClause
	flagOrder []*FlagClause
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
	f.flagOrder = append(f.flagOrder, flag)
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

			fb, ok := flag.value.(boolFlag)
			if !ok || !fb.IsBoolFlag() {
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
	name         string
	Shorthand    byte
	help         string
	defaultValue string
	metavar      string
	dispatch     Dispatch
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
		if f.metavar == "%DEFAULT%" {
			return f.defaultValue
		}
		return f.metavar
	}
	return strings.ToUpper(f.name)
}

func (f *FlagClause) init() {
	if f.required && f.defaultValue != "" {
		panic(fmt.Sprintf("required flag '%s' with unusable default value", f.name))
	}
	if f.value == nil {
		panic(fmt.Sprintf("no value defined for --%s", f.name))
	}
	if f.defaultValue != "" {
		if err := f.value.Set(f.defaultValue); err != nil {
			panic(fmt.Sprintf("default value for --%s is invalid: %s", f.name, err))
		}
	}
}

// Dispatch to the given function when the flag is parsed.
func (f *FlagClause) Dispatch(dispatch Dispatch) *FlagClause {
	f.dispatch = dispatch
	return f
}

// Default value for this flag. It *must* be parseable by the value of the flag.
func (f *FlagClause) Default(value string) *FlagClause {
	f.defaultValue = value
	return f
}

// MetaVar sets the placeholder string used for flag values in the help. If
// the magic string "%DEFAULT%" is used, the Default() value of the flag will
// be used.
func (f *FlagClause) MetaVar(metavar string) *FlagClause {
	f.metavar = metavar
	return f
}

// MetaVarFromDefault uses the default value for the flag as the MetaVar.
func (f *FlagClause) MetaVarFromDefault() *FlagClause {
	f.metavar = "%DEFAULT%"
	return f
}

// Required makes the flag required. You can not provide a Default() value to a Required() flag.
func (f *FlagClause) Required() *FlagClause {
	f.required = true
	return f
}

// Short sets the short flag name.
func (f *FlagClause) Short(name byte) *FlagClause {
	f.Shorthand = name
	return f
}

// Bool makes this flag a boolean flag.
func (f *FlagClause) Bool() (target *bool) {
	target = new(bool)
	f.SetValue(newBoolValue(false, target))
	return
}
