package kingpin

// Action callback triggered during parsing.
//
// "element" is the flag, argument or command associated with the callback. It contains the Clause
// and the string value.
//
// "context" contains the full parse context, including all other elements that have been parsed.
type Action func(element *ParseElement, context *ParseContext) error

type actionApplier interface {
	applyActions(*ParseElement, *ParseContext) error
	applyPreActions(*ParseElement, *ParseContext) error
}

type actionMixin struct {
	actions    []Action
	preActions []Action
}

func (h *actionMixin) addAction(action Action) {
	h.actions = append(h.actions, action)
}

func (h *actionMixin) addPreAction(action Action) {
	h.actions = append(h.actions, action)
}

func (a *actionMixin) applyActions(element *ParseElement, context *ParseContext) error {
	for _, action := range a.actions {
		if err := action(element, context); err != nil {
			return err
		}
	}
	return nil
}

func (a *actionMixin) applyPreActions(element *ParseElement, context *ParseContext) error {
	for _, preAction := range a.preActions {
		if err := preAction(element, context); err != nil {
			return err
		}
	}
	return nil
}
