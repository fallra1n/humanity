package globaltarget

import "github.com/fallra1n/humanity/internal/shared"

// ...
type BasicGlobalTarget struct {
	Id                    shared.ID
	Name                  string
	Power                 float64
	RequiredTags          shared.Tags
	AvailableLocalTargets []shared.LocalTarget
	CompletedLocalTargets []shared.LocalTarget
}

func NewBasicGlobalTarget(id shared.ID, name string, power float64, requiredTags shared.Tags) *BasicGlobalTarget {
	return &BasicGlobalTarget{
		Id:                    id,
		Name:                  name,
		Power:                 power,
		RequiredTags:          requiredTags,
		AvailableLocalTargets: make([]shared.LocalTarget, 0),
		CompletedLocalTargets: make([]shared.LocalTarget, 0),
	}
}
