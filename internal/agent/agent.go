package agent

import (
	"sync"
	"time"

	"github.com/fallra1n/humanity/internal/shared"
)

type Agent struct {
	mu sync.RWMutex

	Id        shared.ID
	Age       shared.Age
	Money     shared.Money
	Items     map[string]int64
	Tags      shared.Tags
	Dead      bool
	BusyUntil time.Time

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
