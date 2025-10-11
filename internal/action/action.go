package action

import "github.com/fallra1n/humanity/internal/shared"

// ...
type BasicAction struct {
	id            shared.ID
	name          string
	cost          shared.Money
	duration      shared.Hours
	requiredTags  shared.Tags
	requiredItems map[string]int64
	consumedItems map[string]int64
	producedItems map[string]int64
	bonusMoney    shared.Money
	rules         map[string]string
}
