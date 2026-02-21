package components

import (
	"github.com/fallra1n/humanity/src/utils"
	"sync"
)

// GlobalTarget представляет долгосрочную цель
type GlobalTarget struct {
	Name            string
	Tags            map[string]bool
	Power           float64
	TargetsPossible map[*LocalTarget]bool
	TargetsExecuted map[*LocalTarget]bool
	Mu              sync.RWMutex
}

// NewGlobalTarget создает новую глобальную цель
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

// MarkAsExecuted отмечает локальную цель как выполненную
func (gt *GlobalTarget) MarkAsExecuted(target *LocalTarget) {
	gt.Mu.Lock()
	defer gt.Mu.Unlock()
	delete(gt.TargetsPossible, target)
	gt.TargetsExecuted[target] = true
}

// IsExecutedFull проверяет, покрыты ли все теги
func (gt *GlobalTarget) IsExecutedFull() bool {
	gt.Mu.RLock()
	defer gt.Mu.RUnlock()

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

// Executable проверяет, может ли глобальная цель быть выполнена
func (gt *GlobalTarget) Executable(person *Human) bool {
	gt.Mu.RLock()
	defer gt.Mu.RUnlock()

	unclosedTags := make(map[string]bool)
	for tag := range gt.Tags {
		unclosedTags[tag] = true
	}

	// Удалить выполненные теги
	for target := range gt.TargetsExecuted {
		for tag := range target.Tags {
			delete(unclosedTags, tag)
		}
	}

	// Проверить, могут ли оставшиеся теги быть закрыты
	for target := range gt.TargetsPossible {
		if target.Executable(person) {
			for tag := range target.Tags {
				delete(unclosedTags, tag)
			}
		}
	}

	return len(unclosedTags) == 0
}

// ChooseTarget выбирает лучшую локальную цель для этой глобальной цели
func (gt *GlobalTarget) ChooseTarget(person *Human) *LocalTarget {
	gt.Mu.RLock()
	defer gt.Mu.RUnlock()

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
		// Получить цели с наивысшим рейтингом
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
