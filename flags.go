package kingpin

import (
	"fmt"
	"os"
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

func (f *flagGroup) init() error {
	for _, flag := range f.long {
		if err := flag.init(); err != nil {
			return err
		}
		if flag.shorthand != 0 {
			f.short[string(flag.shorthand)] = flag
		}
	}
	return nil
}

func (f *flagGroup) parse(tokens tokens, ignoreRequired bool) (tokens, error) {
	// Track how many required flags we've seen.
	remaining := make(map[string]struct{})
	for k, flag := range f.long {
		if !ignoreRequired && flag.needsValue() {
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
			invert := false

			name := token.Value
			if token.Type == TokenLong {
				if strings.HasPrefix(name, "no-") {
					name = name[3:]
					invert = true
				}
				flag, ok = f.long[name]
				if !ok {
					return nil, fmt.Errorf("unknown long flag '%s'", flagToken)
				}
			} else {
				flag, ok = f.short[name]
				if !ok {
					return nil, fmt.Errorf("unknown short flag '%s", flagToken)
				}
			}

			delete(remaining, flag.name)

			fb, ok := flag.value.(boolFlag)
			if ok && fb.IsBoolFlag() {
				if invert {
					defaultValue = "false"
				} else {
					defaultValue = "true"
				}
			} else {
				if invert {
					return nil, fmt.Errorf("unknown long flag '%s'", flagToken)
				}
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
	shorthand    byte
	help         string
	envar        string
	defaultValue string
	placeholder  string
	dispatch     Dispatch
}

func newFlag(name, help string) *FlagClause {
	f := &FlagClause{
		name: name,
		help: help,
	}
	return f
}

func (f *FlagClause) needsValue() bool {
	return f.required && f.defaultValue == ""
}

func (f *FlagClause) formatPlaceHolder() string {
	if f.placeholder != "" {
		return f.placeholder
	}
	if f.defaultValue != "" {
		if _, ok := f.value.(*stringValue); ok {
			return fmt.Sprintf("%q", f.value)
		}
		return f.value.String()
	}
	return strings.ToUpper(f.name)
}

func (f *FlagClause) init() error {
	if f.required && f.defaultValue != "" {
		return fmt.Errorf("required flag '--%s' with default value that will never be used", f.name)
	}
	if f.value == nil {
		return fmt.Errorf("no value defined for --%s", f.name)
	}
	if f.envar != "" {
		if v := os.Getenv(f.envar); v != "" {
			f.defaultValue = v
		}
	}
	if f.defaultValue != "" {
		if err := f.value.Set(f.defaultValue); err != nil {
			return fmt.Errorf("default value for --%s is invalid: %s", f.name, err)
		}
	}
	return nil
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

// OverrideDefaultFromEnvar overrides the default value for a flag from an
// environment variable, if available.
func (f *FlagClause) OverrideDefaultFromEnvar(envar string) *FlagClause {
	f.envar = envar
	return f
}

// PlaceHolder sets the place-holder string used for flag values in the help. The
// default behaviour is to use the value provided by Default() if provided,
// then fall back on the capitalized flag name.
func (f *FlagClause) PlaceHolder(placeholder string) *FlagClause {
	f.placeholder = placeholder
	return f
}

// Required makes the flag required. You can not provide a Default() value to a Required() flag.
func (f *FlagClause) Required() *FlagClause {
	f.required = true
	return f
}

// Short sets the short flag name.
func (f *FlagClause) Short(name byte) *FlagClause {
	f.shorthand = name
	return f
}

// Bool makes this flag a boolean flag.
func (f *FlagClause) Bool() (target *bool) {
	target = new(bool)
	f.SetValue(newBoolValue(false, target))
	return
}
