package agent

import (
	"sync"
	"time"

	"github.com/fallra1n/humanity/internal/shared"
)

type Agent struct {
	mu sync.RWMutex

	id        shared.ID
	age       shared.Age
	money     shared.Money
	items     map[string]int64
	tags      shared.Tags
	dead      bool
	busyUntil time.Time

	// Информация о работе
	job     shared.Vacancy
	jobTime shared.Hours

	// Локация
	location shared.Location

	// Семья
	parents  map[shared.ID]*Agent
	children map[shared.ID]*Agent
	family   map[shared.ID]*Agent
	friends  map[shared.ID]*Agent

	// Цели и нужды
	globalTargets    []shared.GlobalTarget
	completedTargets []shared.GlobalTarget
	splashes         []*shared.Splash

	// Communication channels
	actionChan chan actionRequest
	stopChan   chan struct{}
	doneChan   chan struct{}
}

// ...
type actionRequest struct {
	action   shared.Action
	response chan error
}
