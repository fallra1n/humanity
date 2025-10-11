package localtarget

import "github.com/fallra1n/humanity/internal/shared"

// ...
type BasicLocalTarget struct {
	id               shared.ID
	name             string
	requiredTags     shared.Tags
	availableActions []shared.Action
	completedActions []shared.Action
}
