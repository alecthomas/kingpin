package kingpin

import (
	"fmt"
	"strings"
)

type flagGroup struct {
	short     map[string]*Clause
	long      map[string]*Clause
	flagOrder []*Clause
}

func newFlagGroup() *flagGroup {
	return &flagGroup{
		short: map[string]*Clause{},
		long:  map[string]*Clause{},
	}
}

// GetFlag gets a flag definition.
//
// This allows existing flags to be modified after definition but before parsing. Useful for
// modular applications.
func (f *flagGroup) GetFlag(name string) *Clause {
	return f.long[name]
}

// Flag defines a new flag with the given long name and help.
func (f *flagGroup) Flag(name, help string) *Clause {
	flag := NewClause(name, help)
	f.long[name] = flag
	f.flagOrder = append(f.flagOrder, flag)
	return flag
}

func (f *flagGroup) init(defaultEnvarPrefix string) error {
	if err := f.checkDuplicates(); err != nil {
		return err
	}
	for _, flag := range f.long {
		if defaultEnvarPrefix != "" && !flag.noEnvar && flag.envar == "" {
			flag.envar = envarTransform(defaultEnvarPrefix + "_" + flag.name)
		}
		if err := flag.init(); err != nil {
			return err
		}
		if flag.shorthand != 0 {
			f.short[string(flag.shorthand)] = flag
		}
	}
	return nil
}

func (f *flagGroup) checkDuplicates() error {
	seenShort := map[rune]bool{}
	seenLong := map[string]bool{}
	for _, flag := range f.flagOrder {
		if flag.shorthand != 0 {
			if _, ok := seenShort[flag.shorthand]; ok {
				return fmt.Errorf("duplicate short flag -%c", flag.shorthand)
			}
			seenShort[flag.shorthand] = true
		}
		if _, ok := seenLong[flag.name]; ok {
			return fmt.Errorf("duplicate long flag --%s", flag.name)
		}
		seenLong[flag.name] = true
	}
	return nil
}

func (f *flagGroup) parse(context *ParseContext) (*Clause, error) {
	var token *Token

loop:
	for {
		token = context.Peek()
		switch token.Type {
		case TokenEOL:
			break loop

		case TokenLong, TokenShort:
			flagToken := token
			var flag *Clause
			var ok bool
			invert := false

			name := token.Value
			if token.Type == TokenLong {
				flag, ok = f.long[name]
				if !ok {
					if strings.HasPrefix(name, "no-") {
						name = name[3:]
						invert = true
					}
					flag, ok = f.long[name]
				} else if strings.HasPrefix(name, "no-") {
					invert = true
				}
				if !ok {
					return nil, fmt.Errorf("unknown long flag '%s'", flagToken)
				}
			} else {
				flag, ok = f.short[name]
				if !ok {
					return nil, fmt.Errorf("unknown short flag '%s'", flagToken)
				}
			}

			context.Next()

			var defaultValue string
			if fb, ok := flag.value.(boolFlag); ok && fb.IsBoolFlag() {
				if invert {
					defaultValue = "false"
				} else {
					defaultValue = "true"
				}
			} else {
				if invert {
					context.Push(token)
					return nil, fmt.Errorf("unknown long flag '%s'", flagToken)
				}
				token = context.Peek()
				if token.Type != TokenArg {
					context.Push(token)
					return nil, fmt.Errorf("expected argument for flag '%s'", flagToken)
				}
				context.Next()
				defaultValue = token.Value
			}

			context.matchedFlag(flag, defaultValue)
			return flag, nil

		default:
			break loop
		}
	}
	return nil, nil
}
