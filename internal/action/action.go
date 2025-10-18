package action

import (
	"context"

	"github.com/fallra1n/humanity/internal/shared"
)

// ...
type BasicAction struct {
	Id            shared.ID
	Name          string
	Cost          shared.Money
	Duration      shared.Hours
	RequiredTags  shared.Tags
	RequiredItems map[string]int64
	ConsumedItems map[string]int64
	ProducedItems map[string]int64
	BonusMoney    shared.Money
	Rules         map[string]string
}

// ...
func NewBasicAction(id shared.ID, name string, cost shared.Money, duration shared.Hours) *BasicAction {
	return &BasicAction{
		Id:            id,
		Name:          name,
		Cost:          cost,
		Duration:      duration,
		RequiredTags:  make(shared.Tags),
		RequiredItems: make(map[string]int64),
		ConsumedItems: make(map[string]int64),
		ProducedItems: make(map[string]int64),
		BonusMoney:    0,
		Rules:         make(map[string]string),
	}
}

func (a *BasicAction) CanExecute(agent shared.Agent) bool {
	return false
}

func (a *BasicAction) Execute(ctx context.Context, agent shared.Agent) error {
	return nil
}

func (a *BasicAction) GetRequiredTags() shared.Tags {
	return a.RequiredTags
}
