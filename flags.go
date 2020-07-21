package kingpin

import (
	"fmt"
	"strings"
)

// FlagGroup provides access to defining and retrieving flags
type FlagGroup interface {
	// GetFlag gets a flag definition.
	//
	// This allows existing flags to be modified after definition but before parsing. Useful for
	// modular applications.
	GetFlag(name string) *FlagClause

	// Flag defines a new flag with the given long name and help.
	Flag(name, help string) *FlagClause

	// GetFlagGroup returns a child by the given name of this FlagGroup
	GetFlagGroup(prefix string) *SubFlagGroup

	// FlagGroup create a new sub group at this FlagGroup with the given name.
	//
	// This allows grouping flags and prevents duplication of namings.
	FlagGroup(prefix string) *SubFlagGroup

	// RegisterFlagsOf will execute a registrar for flags for this FlagGroup.
	RegisterFlagsOf(registrars ...FlagRegistrar)
}

type flagGroup struct {
	short     map[string]*FlagClause
	long      map[string]*FlagClause
	flagOrder []*FlagClause

	subGroups       map[string]*SubFlagGroup
	envarNamePrefix string
}

func newFlagGroup() *flagGroup {
	return &flagGroup{
		short:     map[string]*FlagClause{},
		long:      map[string]*FlagClause{},
		subGroups: map[string]*SubFlagGroup{},
	}
}

// GetFlag gets a flag definition.
//
// This allows existing flags to be modified after definition but before parsing. Useful for
// modular applications.
func (f *flagGroup) GetFlag(name string) *FlagClause {
	return f.long[name]
}

func (f *flagGroup) newFlag(name, help string, holder FlagGroup, envarNameResolver envarNameResolver) *FlagClause {
	flag := newFlag(name, help, holder, envarNameResolver)
	f.long[name] = flag
	f.flagOrder = append(f.flagOrder, flag)
	return flag
}

// GetFlagGroup returns a child by the given name of this FlagGroup
func (f *flagGroup) GetFlagGroup(prefix string) *SubFlagGroup {
	return f.subGroups[prefix]
}

// FlagGroup create a new sub group at this FlagGroup with the given name.
//
// This allows grouping flags and prevents duplication of namings.
func (f *flagGroup) newFlagGroup(prefix string, holder FlagGroup, parentEnvarPrefixProvider envarPrefixProvider) *SubFlagGroup {
	fg := newFlagGroup()
	s := &SubFlagGroup{
		flagGroup:                 *fg,
		parentEnvarPrefixProvider: parentEnvarPrefixProvider,
		holder:                    holder,
		prefix:                    prefix,
	}
	f.subGroups[prefix] = s
	return s
}

func (f *flagGroup) init(defaultEnvarPrefix string) error {
	if err := f.checkDuplicates(); err != nil {
		return err
	}
	return f.initRecursively(defaultEnvarPrefix)
}

func (f *flagGroup) initRecursively(defaultEnvarPrefix string) error {
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
	for name, subGroup := range f.subGroups {
		subPrefix := ""
		if defaultEnvarPrefix != "" {
			subPrefix = defaultEnvarPrefix + "_" + name
		}
		if err := subGroup.initRecursively(subPrefix); err != nil {
			return err
		}
	}
	return nil
}

func (f *flagGroup) checkDuplicates() error {
	seenShort := map[rune]bool{}
	seenLong := map[string]bool{}
	return f.checkDuplicatesRecursively("", seenShort, seenLong)
}

func (f *flagGroup) checkDuplicatesRecursively(prefix string, seenShort map[rune]bool, seenLong map[string]bool) error {
	for _, flag := range f.flagOrder {
		if flag.shorthand != 0 {
			if _, ok := seenShort[flag.shorthand]; ok {
				return fmt.Errorf("duplicate short flag -%c", flag.shorthand)
			}
			seenShort[flag.shorthand] = true
		}
		if _, ok := seenLong[prefix+flag.name]; ok {
			return fmt.Errorf("duplicate long flag --%s", prefix+flag.name)
		}
		seenLong[prefix+flag.name] = true
	}
	for name, subGroup := range f.subGroups {
		if err := subGroup.checkDuplicatesRecursively(prefix+name, seenShort, seenLong); err != nil {
			return err
		}
	}
	return nil
}

func (f *flagGroup) getShort(name string) (*FlagClause, bool) {
	if v, ok := f.short[name]; ok {
		return v, true
	}
	for _, subGroup := range f.subGroups {
		if v, ok := subGroup.getShort(name); ok {
			return v, true
		}
	}
	return nil, false
}

func (f *flagGroup) getLong(name string) (*FlagClause, bool) {
	if v, ok := f.long[name]; ok {
		return v, true
	}
	for subGroupName, subGroup := range f.subGroups {
		if strings.HasPrefix(name, subGroupName) {
			if v, ok := subGroup.getLong(name[len(subGroupName):]); ok {
				return v, true
			}
		}
	}
	return nil, false
}

func (f *flagGroup) parse(context *ParseContext) (*FlagClause, error) {
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
				flag, ok = f.getLong(name)
				if !ok {
					if strings.HasPrefix(name, "no-") {
						name = name[3:]
						invert = true
					}
					flag, ok = f.long[name]
				}
				if !ok {
					return nil, fmt.Errorf("unknown long flag '%s'", flagToken)
				}
			} else {
				flag, ok = f.getShort(name)
				if !ok {
					return nil, fmt.Errorf("unknown short flag '%s'", flagToken)
				}
			}

			context.Next()

			flag.isSetByUser()

			fb, ok := flag.value.(boolFlag)
			if ok && fb.IsBoolFlag() {
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

// GetEnvarPrefix will return the actual environment variable prefix.
func (f *flagGroup) GetEnvarPrefix() string {
	return f.envarNamePrefix
}

// FlagClause is a fluid interface used to build flags.
type FlagClause struct {
	parserMixin
	actionMixin
	completionsMixin
	envarMixin
	name          string
	shorthand     rune
	help          string
	defaultValues []string
	placeholder   string
	hidden        bool
	setByUser     *bool
	holder        FlagGroup
}

func newFlag(name, help string, holder FlagGroup, envarNameResolver envarNameResolver) *FlagClause {
	f := &FlagClause{
		name:   name,
		help:   help,
		holder: holder,
	}
	f.nameResolver = envarNameResolver
	return f
}

func (f *FlagClause) getName() string {
	name := f.name
	nextHolder := f.holder
	for {
		if subGroup, ok := nextHolder.(*SubFlagGroup); ok {
			name = subGroup.prefix + name
			nextHolder = subGroup.holder
		} else {
			break
		}
	}
	return name
}

func (f *FlagClause) setDefault() error {
	if f.HasEnvarValue() {
		if v, ok := f.value.(repeatableFlag); !ok || !v.IsCumulative() {
			// Use the value as-is
			return f.value.Set(f.GetEnvarValue())
		} else {
			for _, value := range f.GetSplitEnvarValue() {
				if err := f.value.Set(value); err != nil {
					return err
				}
			}
			return nil
		}
	}

	if len(f.defaultValues) > 0 {
		for _, defaultValue := range f.defaultValues {
			if err := f.value.Set(defaultValue); err != nil {
				return err
			}
		}
		return nil
	}

	return nil
}

func (f *FlagClause) isSetByUser() {
	if f.setByUser != nil {
		*f.setByUser = true
	}
}

func (f *FlagClause) needsValue() bool {
	haveDefault := len(f.defaultValues) > 0
	return f.required && !(haveDefault || f.HasEnvarValue())
}

func (f *FlagClause) init() error {
	if f.required && len(f.defaultValues) > 0 {
		return fmt.Errorf("required flag '--%s' with default value that will never be used", f.getName())
	}
	if f.value == nil {
		return fmt.Errorf("no type defined for --%s (eg. .String())", f.getName())
	}
	if v, ok := f.value.(repeatableFlag); (!ok || !v.IsCumulative()) && len(f.defaultValues) > 1 {
		return fmt.Errorf("invalid default for '--%s', expecting single value", f.getName())
	}
	return nil
}

// Action dispatch to the given function after the flag is parsed and validated.
func (f *FlagClause) Action(action Action) *FlagClause {
	f.addAction(action)
	return f
}

// AddAction works like Action but does not return the current instance.
// This will fulfill the common interface ActionGroup
func (f *FlagClause) AddAction(action Action) {
	f.Action(action)
}

// PreAction called after parsing completes but before validation and execution.
func (f *FlagClause) PreAction(action Action) *FlagClause {
	f.addPreAction(action)
	return f
}

// AddPreAction works like PreAction but does not return the current instance.
// This will fulfill the common interface ActionGroup
func (f *FlagClause) AddPreAction(action Action) {
	f.PreAction(action)
}

// HintAction registers a HintAction (function) for the flag to provide completions
func (f *FlagClause) HintAction(action HintAction) *FlagClause {
	f.addHintAction(action)
	return f
}

// HintOptions registers any number of options for the flag to provide completions
func (f *FlagClause) HintOptions(options ...string) *FlagClause {
	f.addHintAction(func() []string {
		return options
	})
	return f
}

func (f *FlagClause) EnumVar(target *string, options ...string) {
	f.parserMixin.EnumVar(target, options...)
	f.addHintActionBuiltin(func() []string {
		return options
	})
}

func (f *FlagClause) Enum(options ...string) (target *string) {
	f.addHintActionBuiltin(func() []string {
		return options
	})
	return f.parserMixin.Enum(options...)
}

// IsSetByUser let to know if the flag was set by the user
func (f *FlagClause) IsSetByUser(setByUser *bool) *FlagClause {
	if setByUser != nil {
		*setByUser = false
	}
	f.setByUser = setByUser
	return f
}

// Default values for this flag. They *must* be parseable by the value of the flag.
func (f *FlagClause) Default(values ...string) *FlagClause {
	f.defaultValues = values
	return f
}

// DEPRECATED: Use Envar(name) instead.
func (f *FlagClause) OverrideDefaultFromEnvar(envar string) *FlagClause {
	return f.Envar(envar)
}

// Envar overrides the default value(s) for a flag from an environment variable,
// if it is set. Several default values can be provided by using new lines to
// separate them.
func (f *FlagClause) Envar(name string) *FlagClause {
	f.envar = name
	f.noEnvar = false
	return f
}

// NoEnvar forces environment variable defaults to be disabled for this flag.
// Most useful in conjunction with app.DefaultEnvars().
func (f *FlagClause) NoEnvar() *FlagClause {
	f.envar = ""
	f.noEnvar = true
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
func (f *FlagClause) Short(name rune) *FlagClause {
	f.shorthand = name
	return f
}

// Help sets the help message.
func (f *FlagClause) Help(help string) *FlagClause {
	f.help = help
	return f
}

// Bool makes this flag a boolean flag.
func (f *FlagClause) Bool() (target *bool) {
	target = new(bool)
	f.SetValue(newBoolValue(target))
	return
}

type SubFlagGroup struct {
	flagGroup
	parentEnvarPrefixProvider envarPrefixProvider

	holder FlagGroup
	prefix string
}

type envarPrefixProvider func() string

// EnvarNamePrefix will set a prefix for environment variables.
// This will prevent for duplicating the prefixes for all environment variables.
func (f *SubFlagGroup) EnvarNamePrefix(prefix string) *SubFlagGroup {
	f.envarNamePrefix = prefix
	return f
}

// Flag defines a new flag with the given long name and help.
func (f *SubFlagGroup) Flag(name, help string) *FlagClause {
	return f.newFlag(name, help, f, f.resolveEnvarName)
}

func (f *SubFlagGroup) resolveEnvarName(name string) string {
	return f.GetEnvarPrefix() + name
}

// GetEnvarPrefix will return the actual environment variable prefix.
func (f *SubFlagGroup) GetEnvarPrefix() string {
	return f.parentEnvarPrefixProvider() + f.flagGroup.GetEnvarPrefix()
}

// FlagGroup create a new sub group at this FlagGroup with the given name.
//
// This allows grouping flags and prevents duplication of namings.
func (f *SubFlagGroup) FlagGroup(prefix string) *SubFlagGroup {
	return f.newFlagGroup(prefix, f, f.GetEnvarPrefix)
}

// RegisterFlagsOf will execute a registrar for flags.
// which will be executed before pre-actions and validation to register its flags.
func (f *SubFlagGroup) RegisterFlagsOf(registrars ...FlagRegistrar) {
	registerFlagsOf(f, registrars...)
}

// FlagsOf will execute a registrar for flags for this CmdClause
// which will be executed before pre-actions and validation to register its flags.
func (f *SubFlagGroup) FlagsOf(registrars ...FlagRegistrar) *SubFlagGroup {
	f.RegisterFlagsOf(registrars...)
	return f
}
