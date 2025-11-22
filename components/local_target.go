package components

import (
	"sync"
	"github.com/fallra1n/humanity/utils"
)

// LocalTarget represents a short-term goal
type LocalTarget struct {
	Name            string
	Tags            map[string]bool
	ActionsPossible map[*Action]bool
	ActionsExecuted map[*Action]bool
	Mu              sync.RWMutex
}

// NewLocalTarget creates a new LocalTarget
func NewLocalTarget(name string, tags []string, allActions []*Action) *LocalTarget {
	tagSet := make(map[string]bool)
	for _, tag := range tags {
		tagSet[tag] = true
	}

	actionsPossible := make(map[*Action]bool)
	for _, action := range allActions {
		if len(utils.IntersectSlices(tags, getKeysFromMap(action.Tags))) > 0 {
			actionsPossible[action] = true
		}
	}

	return &LocalTarget{
		Name:            name,
		Tags:            tagSet,
		ActionsPossible: actionsPossible,
		ActionsExecuted: make(map[*Action]bool),
	}
}

// MarkAsExecuted marks an action as executed
func (lt *LocalTarget) MarkAsExecuted(action *Action) {
	lt.Mu.Lock()
	defer lt.Mu.Unlock()
	delete(lt.ActionsPossible, action)
	lt.ActionsExecuted[action] = true
}

// IsExecutedFull checks if all tags are covered by executed actions
func (lt *LocalTarget) IsExecutedFull() bool {
	lt.Mu.RLock()
	defer lt.Mu.RUnlock()
	
	remainingTags := make(map[string]bool)
	for tag := range lt.Tags {
		remainingTags[tag] = true
	}

	for action := range lt.ActionsExecuted {
		for tag := range action.Tags {
			delete(remainingTags, tag)
		}
	}

	return len(remainingTags) == 0
}

// Executable checks if the local target can be executed
func (lt *LocalTarget) Executable(person *Human) bool {
	lt.Mu.RLock()
	defer lt.Mu.RUnlock()
	
	unclosedTags := make(map[string]bool)
	for tag := range lt.Tags {
		unclosedTags[tag] = true
	}

	// Remove executed tags
	for action := range lt.ActionsExecuted {
		for tag := range action.Tags {
			delete(unclosedTags, tag)
		}
	}

	// Check if remaining tags can be closed
	for action := range lt.ActionsPossible {
		if action.Executable(person) {
			for tag := range action.Tags {
				delete(unclosedTags, tag)
			}
		}
	}

	return len(unclosedTags) == 0
}

// ChooseAction selects the best action for this target
func (lt *LocalTarget) ChooseAction(person *Human) *Action {
	lt.Mu.RLock()
	defer lt.Mu.RUnlock()
	
	leftTags := make(map[string]bool)
	for tag := range lt.Tags {
		leftTags[tag] = true
	}

	for action := range lt.ActionsExecuted {
		for tag := range action.Tags {
			delete(leftTags, tag)
		}
	}

	rating := make(map[uint64][]*Action)
	for action := range lt.ActionsPossible {
		if action.Executable(person) {
			var rate uint64 = 0
			for tag := range action.Tags {
				if leftTags[tag] {
					rate++
				}
			}
			rating[rate] = append(rating[rate], action)
		}
	}

	if len(rating) > 0 {
		// Get highest rated actions
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