package kingpin

import (
	"fmt"
	"strconv"
	"strings"
)

// Data model for Kingpin command-line structure.

type FlagGroupModel struct {
	Flags []*FlagModel
}

func (f *FlagGroupModel) FlagSummary() string {
	out := []string{}
	count := 0
	for _, flag := range f.Flags {
		if flag.Name != "help" {
			count++
		}
		if flag.Required {
			if flag.IsBoolFlag() {
				out = append(out, fmt.Sprintf("--[no-]%s", flag.Name))
			} else {
				out = append(out, fmt.Sprintf("--%s=%s", flag.Name, flag.FormatPlaceHolder()))
			}
		}
	}
	if count != len(out) {
		out = append(out, "[<flags>]")
	}
	return strings.Join(out, " ")
}

type FlagModel struct {
	Name        string
	Help        string
	Short       byte
	Default     string
	Envar       string
	PlaceHolder string
	Required    bool
	Hidden      bool
	flag        *FlagClause
}

func (f *FlagModel) String() string {
	return f.flag.value.String()
}

func (f *FlagModel) IsBoolFlag() bool {
	if fl, ok := f.flag.value.(boolFlag); ok {
		return fl.IsBoolFlag()
	}
	return false
}

func (f *FlagModel) FormatPlaceHolder() string {
	if f.PlaceHolder != "" {
		return f.PlaceHolder
	}
	if f.Default != "" {
		if _, ok := f.flag.value.(*stringValue); ok {
			return strconv.Quote(f.Default)
		}
		return f.Default
	}
	return strings.ToUpper(f.Name)
}

type ArgGroupModel struct {
	Args []*ArgModel
}

func (a *ArgGroupModel) ArgSummary() string {
	depth := 0
	out := []string{}
	for _, arg := range a.Args {
		h := "<" + arg.Name + ">"
		if !arg.Required {
			h = "[" + h
			depth++
		}
		out = append(out, h)
	}
	out[len(out)-1] = out[len(out)-1] + strings.Repeat("]", depth)
	return strings.Join(out, " ")
}

type ArgModel struct {
	Name     string
	Help     string
	Default  string
	Required bool
	arg      *ArgClause
}

func (a *ArgModel) String() string {
	return a.arg.value.String()
}

type CmdGroupModel struct {
	Commands []*CmdModel
}

func (c *CmdGroupModel) FlattenedCommands() (out []*CmdModel) {
	for _, cmd := range c.Commands {
		if len(cmd.Commands) == 0 {
			out = append(out, cmd)
		}
		out = append(out, cmd.FlattenedCommands()...)
	}
	return
}

type CmdModel struct {
	Name string
	Help string
	*FlagGroupModel
	*ArgGroupModel
	*CmdGroupModel
	cmd *CmdClause
}

func (c *CmdModel) String() string {
	return c.cmd.FullCommand()
}

type ApplicationModel struct {
	Name string
	Help string
	*ArgGroupModel
	*CmdGroupModel
	*FlagGroupModel
}

func (a *Application) Model() *ApplicationModel {
	return &ApplicationModel{
		Name:           a.Name,
		Help:           a.Help,
		FlagGroupModel: a.flagGroup.Model(),
		ArgGroupModel:  a.argGroup.Model(),
		CmdGroupModel:  a.cmdGroup.Model(),
	}
}

func (a *argGroup) Model() *ArgGroupModel {
	m := &ArgGroupModel{}
	for _, arg := range a.args {
		m.Args = append(m.Args, arg.Model())
	}
	return m
}

func (a *ArgClause) Model() *ArgModel {
	return &ArgModel{
		Name:     a.name,
		Help:     a.help,
		Default:  a.defaultValue,
		Required: a.required,
		arg:      a,
	}
}

func (f *flagGroup) Model() *FlagGroupModel {
	m := &FlagGroupModel{}
	for _, fl := range f.flagOrder {
		m.Flags = append(m.Flags, fl.Model())
	}
	return m
}

func (f *FlagClause) Model() *FlagModel {
	return &FlagModel{
		Name:        f.name,
		Help:        f.help,
		Short:       f.shorthand,
		Default:     f.defaultValue,
		Envar:       f.envar,
		PlaceHolder: f.placeholder,
		Required:    f.required,
		Hidden:      f.hidden,
		flag:        f,
	}
}

func (c *cmdGroup) Model() *CmdGroupModel {
	m := &CmdGroupModel{}
	for _, cm := range c.commandOrder {
		m.Commands = append(m.Commands, cm.Model())
	}
	return m
}

func (c *CmdClause) Model() *CmdModel {
	return &CmdModel{
		Name:           c.name,
		Help:           c.help,
		FlagGroupModel: c.flagGroup.Model(),
		ArgGroupModel:  c.argGroup.Model(),
		CmdGroupModel:  c.cmdGroup.Model(),
		cmd:            c,
	}
}
