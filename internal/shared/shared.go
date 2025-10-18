package shared

import (
	"context"
	"time"
)

// ...
type ID string

// ...
type Money int64

// ...
type Hours int64

// ...
type Age float64

// ...
type Tags map[string]bool

func (t Tags) Add(tag string) {
	t[tag] = true
}

func NewTags(tags ...string) Tags {
	t := make(Tags)
	for _, tag := range tags {
		t[tag] = true
	}
	return t
}

func (t Tags) Intersect(other Tags) Tags {
	result := make(Tags)
	for tag := range t {
		if val, ok := other[tag]; ok && val {
			result.Add(tag)
		}
	}
	return result
}

func (t Tags) ToSlice() []string {
	result := make([]string, 0, len(t))
	for tag := range t {
		result = append(result, tag)
	}
	return result
}

// Сущность, которая может выполнять действия
type Agent interface {
	CanExecute(action Action) bool
	Execute(ctx context.Context, action Action) error
}

// Сущность, которая может быть выполнена агентом
type Action interface {
	CanExecute(agent Agent) bool
	Execute(ctx context.Context, agent Agent) error
	GetRequiredTags() Tags
}

// Цель, которая может быть достигнута агентом
type Target interface {
	IsCompleted() bool
	Progress() float64
	CanAchieve(agent Agent) bool
	GetRequiredTags() Tags
}

// Локальная цель, которая может быть достигнута агентом
type LocalTarget interface {
	Target
	ChooseAction(agent Agent) Action
	MarkActionCompleted(action Action)
}

// Глобальная цель, которая может быть достигнута агентом
type GlobalTarget interface {
	Target
	ChooseLocalTarget(agent Agent) LocalTarget
	MarkLocalTargetCompleted(target LocalTarget)
}

// ...
type Location interface {
}

type Job interface {
}

// ...
type Vacancy interface {
}

// ...
type Splash struct {
}

// ...
func (s *Splash) IsExpired() bool {
	return false
}

// ...
type SimulationTime interface {
	Tick()
	Now() time.Time
	StartTime() time.Time
	ElapsedTime() time.Duration
}

// ...
type Event struct {
}

// ...
type EventBus interface {
}

// ...
type Simulator interface {
}
