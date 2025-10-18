package registry

import "github.com/fallra1n/humanity/internal/shared"

// Управляет регистрацией действий
type Registry struct {
	actions map[shared.ID]shared.Action
}

func NewRegistry() *Registry {
	return &Registry{
		actions: make(map[shared.ID]shared.Action),
	}
}

func (r *Registry) Register(actionId shared.ID, action shared.Action) {
	r.actions[actionId] = action
}

// Возвращает все действия
func (r *Registry) All() []shared.Action {
	actions := make([]shared.Action, 0, len(r.actions))
	for _, action := range r.actions {
		actions = append(actions, action)
	}
	return actions
}
