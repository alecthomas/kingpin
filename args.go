package kingpin

import "fmt"

type argGroup struct {
	args []*ArgClause
}

func newArgGroup() *argGroup {
	return &argGroup{}
}

func (a *argGroup) Arg(name, help string) *ArgClause {
	arg := newArg(name, help)
	a.args = append(a.args, arg)
	return arg
}

func (a *argGroup) parse(tokens tokens) (tokens, error) {
	for _, arg := range a.args {
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

		var err error
		tokens, err = arg.parse(tokens)
		if err != nil {
			return nil, err
		}
	}
	return tokens, nil
}

func (a *argGroup) init() {
	required := 0
	seen := map[string]struct{}{}
	for i, arg := range a.args {
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
	if a.required && a.DefValue != "" {
		panic(fmt.Sprintf("required argument '%s' with unusable default value", a.name))
	}
	if a.value == nil {
		panic(fmt.Sprintf("no parser defined for arg '%s'", a.name))
	}
	if a.DefValue != "" {
		if err := a.value.Set(a.DefValue); err != nil {
			panic(fmt.Sprintf("invalid default value '%s' for argument '%s'", a.DefValue, a.name))
		}
	}
}

func (a *ArgClause) parse(tokens tokens) (tokens, error) {
	if token, tokens := tokens.Next(); token.Type == TokenArg {
		if err := a.value.Set(token.Value); err != nil {
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
