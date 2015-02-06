package kingpin

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

type flagGroup struct {
	short     map[string]*FlagClause
	long      map[string]*FlagClause
	pattern   map[string]*FlagClause
	flagOrder []*FlagClause
}

func newFlagGroup() *flagGroup {
	return &flagGroup{
		short:   make(map[string]*FlagClause),
		long:    make(map[string]*FlagClause),
		pattern: make(map[string]*FlagClause),
	}
}

// Flag defines a new flag with the given long name and help.
func (f *flagGroup) Flag(name, help string) *FlagClause {
	flag := newFlag(name, help)
	f.long[name] = flag
	f.flagOrder = append(f.flagOrder, flag)
	return flag
}

// FlagPattern defines a new flag pattern with the given regular expression and help.
func (f *flagGroup) FlagPattern(regex, help string) *FlagClause {
	flag := newFlagPattern(regex, help)
	f.pattern[regex] = flag
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
	for _, flag := range f.pattern {
		if err := flag.init(); err != nil {
			return err
		}
	}
	return nil
}

func (f *flagGroup) parse(context *ParseContext, ignoreRequired bool) error {
	// Track how many required flags we've seen.
	required := make(map[string]bool)
	// Keep track of any flags that we need to initialise with defaults.
	defaults := make(map[string]bool)
	for k, flag := range f.long {
		defaults[k] = true
		if !ignoreRequired && flag.needsValue() {
			required[k] = true
		}
	}

	var token *Token

loop:
	for {
		token = context.Peek()
		switch token.Type {
		case TokenEOL:
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
					for _, p := range f.pattern {
						if m := p.regex.FindStringSubmatch(name); m != nil {
							flag = p
							if flag.submatches != nil {
								*flag.submatches = append(*flag.submatches, m[1:]...)
							}
							ok = true
							break
						}
					}
					if !ok {
						return fmt.Errorf("unknown long flag '%s'", flagToken)
					}
				}
			} else {
				flag, ok = f.short[name]
				if !ok {
					return fmt.Errorf("unknown short flag '%s'", flagToken)
				}
			}

			delete(required, flag.name)
			delete(defaults, flag.name)

			context.Next()

			fb, ok := flag.value.(boolFlag)
			if ok && fb.IsBoolFlag() {
				if invert {
					defaultValue = "false"
				} else {
					defaultValue = "true"
				}
			} else {
				if invert {
					return fmt.Errorf("unknown long flag '%s'", flagToken)
				}
				token = context.Peek()
				if token.Type != TokenArg {
					return fmt.Errorf("expected argument for flag '%s'", flagToken)
				}
				context.Next()
				defaultValue = token.Value
			}

			if err := flag.value.Set(defaultValue); err != nil {
				return err
			}

			if flag.dispatch != nil {
				if err := flag.dispatch(context); err != nil {
					return err
				}
			}

		default:
			break loop
		}
	}

	// Check that required flags were provided.
	if len(required) == 1 {
		for k := range required {
			return fmt.Errorf("required flag --%s not provided", k)
		}
	} else if len(required) > 1 {
		flags := make([]string, 0, len(required))
		for k := range required {
			flags = append(flags, "--"+k)
		}
		return fmt.Errorf("required flags %s not provided", strings.Join(flags, ", "))
	}

	// Apply defaults to all unprocessed flags.
	for k := range defaults {
		flag := f.long[k]
		if flag.defaultValue != "" {
			if err := flag.value.Set(flag.defaultValue); err != nil {
				return fmt.Errorf("default value for --%s is invalid: %s", flag.name, err)
			}
		}
	}
	return nil
}

func (f *flagGroup) visibleFlags() int {
	count := 0
	for _, flag := range f.long {
		if !flag.hidden {
			count++
		}
	}
	return count
}

// FlagClause is a fluid interface used to build flags.
type FlagClause struct {
	parserMixin
	name         string
	pattern      string
	regex        *regexp.Regexp
	submatches   *[]string
	shorthand    byte
	help         string
	envar        string
	defaultValue string
	placeholder  string
	dispatch     Dispatch
	hidden       bool
}

func newFlag(name, help string) *FlagClause {
	f := &FlagClause{
		name: name,
		help: help,
	}
	return f
}

func newFlagPattern(pattern, help string) *FlagClause {
	f := &FlagClause{
		name:    fmt.Sprintf("/%s/", pattern),
		pattern: pattern,
		help:    help,
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
			return fmt.Sprintf("%q", f.defaultValue)
		}
		return f.defaultValue
	}
	return strings.ToUpper(f.name)
}

func (f *FlagClause) init() error {
	if f.required && f.defaultValue != "" {
		return fmt.Errorf("required flag '--%s' with default value that will never be used", f.name)
	}
	if f.value == nil {
		return fmt.Errorf("no type defined for --%s (eg. .String())", f.name)
	}
	if f.envar != "" {
		if v := os.Getenv(f.envar); v != "" {
			f.defaultValue = v
		}
	}
	if f.pattern != "" {
		if regex, err := regexp.Compile(f.pattern); err == nil {
			f.regex = regex
		} else {
			return fmt.Errorf("invalid pattern definition - %s", err.Error())
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

// Hidden hides a flag from usage but still allows it to be used.
func (f *FlagClause) Hidden() *FlagClause {
	f.hidden = true
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

// Capture specifies the variable that holds all flag pattern regexp submatches`.
func (f *FlagClause) Capture(submatches *[]string) *FlagClause {
	f.submatches = submatches
	return f
}
