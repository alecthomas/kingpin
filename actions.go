package kingpin

// Action callback executed at various stages after all values are populated.
// The application, commands, arguments and flags all have corresponding
// actions.
type Action func(*ParseContext) error

// ActionGroup provides access to define actions and pre-actions.
type ActionGroup interface {
	// AddAction callback to call when all values are populated and parsing is
	// complete, but before any command, flag or argument actions.
	//
	// All Action() callbacks are called in the order they are encountered on the
	// command line.
	AddAction(action Action)

	// AddPreAction called after parsing completes but before validation and execution.
	AddPreAction(action Action)
}

type actionMixin struct {
	actions    []Action
	preActions []Action
}

type actionApplier interface {
	applyActions(*ParseContext) error
	applyPreActions(*ParseContext) error
}

func (a *actionMixin) addAction(action Action) {
	a.actions = append(a.actions, action)
}

func (a *actionMixin) addPreAction(action Action) {
	a.preActions = append(a.preActions, action)
}

func (a *actionMixin) applyActions(context *ParseContext) error {
	for _, action := range a.actions {
		if err := action(context); err != nil {
			return err
		}
	}
	return nil
}

func (a *actionMixin) applyPreActions(context *ParseContext) error {
	for _, preAction := range a.preActions {
		if err := preAction(context); err != nil {
			return err
		}
	}
	return nil
}
