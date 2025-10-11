package globaltarget

import "github.com/fallra1n/humanity/internal/shared"

// ...
type BasicGlobalTarget struct {
	id                    shared.ID
	name                  string
	power                 float64
	requiredTags          shared.Tags
	availableLocalTargets []shared.LocalTarget
	completedLocalTargets []shared.LocalTarget
}
