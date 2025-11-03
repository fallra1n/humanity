package components

import (
	"github.com/fallra1n/humanity/utils"
)

// GlobalTarget represents a long-term goal
type GlobalTarget struct {
	Name            string
	Tags            map[string]bool
	Power           float64
	TargetsPossible map[*LocalTarget]bool
	TargetsExecuted map[*LocalTarget]bool
}

// NewGlobalTarget creates a new GlobalTarget
func NewGlobalTarget(name string, tags []string, power float64, allTargets []*LocalTarget) *GlobalTarget {
	tagSet := make(map[string]bool)
	for _, tag := range tags {
		tagSet[tag] = true
	}

	targetsPossible := make(map[*LocalTarget]bool)
	for _, target := range allTargets {
		if len(utils.IntersectSlices(tags, getKeysFromMap(target.Tags))) > 0 {
			targetsPossible[target] = true
		}
	}

	return &GlobalTarget{
		Name:            name,
		Tags:            tagSet,
		Power:           power,
		TargetsPossible: targetsPossible,
		TargetsExecuted: make(map[*LocalTarget]bool),
	}
}

// MarkAsExecuted marks a local target as executed
func (gt *GlobalTarget) MarkAsExecuted(target *LocalTarget) {
	delete(gt.TargetsPossible, target)
	gt.TargetsExecuted[target] = true
}

// IsExecutedFull checks if all tags are covered
func (gt *GlobalTarget) IsExecutedFull() bool {
	remainingTags := make(map[string]bool)
	for tag := range gt.Tags {
		remainingTags[tag] = true
	}

	for target := range gt.TargetsExecuted {
		for tag := range target.Tags {
			delete(remainingTags, tag)
		}
	}

	return len(remainingTags) == 0
}

// Executable checks if the global target can be executed
func (gt *GlobalTarget) Executable(person *Human) bool {
	unclosedTags := make(map[string]bool)
	for tag := range gt.Tags {
		unclosedTags[tag] = true
	}

	// Remove executed tags
	for target := range gt.TargetsExecuted {
		for tag := range target.Tags {
			delete(unclosedTags, tag)
		}
	}

	// Check if remaining tags can be closed
	for target := range gt.TargetsPossible {
		if target.Executable(person) {
			for tag := range target.Tags {
				delete(unclosedTags, tag)
			}
		}
	}

	return len(unclosedTags) == 0
}

// ChooseTarget selects the best local target for this global target
func (gt *GlobalTarget) ChooseTarget(person *Human) *LocalTarget {
	leftTags := make(map[string]bool)
	for tag := range gt.Tags {
		leftTags[tag] = true
	}

	for target := range gt.TargetsExecuted {
		for tag := range target.Tags {
			delete(leftTags, tag)
		}
	}

	rating := make(map[uint64][]*LocalTarget)
	for target := range gt.TargetsPossible {
		if target.Executable(person) {
			var rate uint64 = 0
			for tag := range target.Tags {
				if leftTags[tag] {
					rate++
				}
			}
			rating[rate] = append(rating[rate], target)
		}
	}

	if len(rating) > 0 {
		// Get highest rated targets
		var maxRate uint64 = 0
		for rate := range rating {
			if rate > maxRate {
				maxRate = rate
			}
		}

		candidates := rating[maxRate]
		if len(candidates) > 0 {
			return candidates[utils.GlobalRandom.NextInt(len(candidates))]
		}
	}

	return nil
}