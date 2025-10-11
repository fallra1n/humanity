package shared

import (
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

type Agent interface{}

type Action interface{}

// ...
type Target interface {
}

// ...
type LocalTarget interface {
	Target
}

// ...
type GlobalTarget interface {
	Target
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
