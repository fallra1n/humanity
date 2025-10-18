package localtarget

import "github.com/fallra1n/humanity/internal/shared"

// ...
type BasicLocalTarget struct {
	Id               shared.ID
	Name             string
	RequiredTags     shared.Tags
	AvailableActions []shared.Action
	CompletedActions []shared.Action
}

func NewBasicLocalTarget(id shared.ID, name string, requiredTags shared.Tags) *BasicLocalTarget {
	return &BasicLocalTarget{
		Id:               id,
		Name:             name,
		RequiredTags:     requiredTags,
		AvailableActions: make([]shared.Action, 0),
		CompletedActions: make([]shared.Action, 0),
	}
}

func (lt *BasicLocalTarget) GetRequiredTags() shared.Tags {
	return lt.RequiredTags
}

func (lt *BasicLocalTarget) IsCompleted() bool {
	return len(lt.CompletedActions) == len(lt.AvailableActions)
}

func (lt *BasicLocalTarget) Progress() float64 {
	return .0
}

func (lt *BasicLocalTarget) CanAchieve(agent shared.Agent) bool {
	return false
}

func (lt *BasicLocalTarget) ChooseAction(agent shared.Agent) shared.Action {
	return nil
}

func (lt *BasicLocalTarget) MarkActionCompleted(action shared.Action) {
	lt.CompletedActions = append(lt.CompletedActions, action)
}
