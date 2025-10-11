package location

import (
	"sync"

	"github.com/fallra1n/humanity/internal/shared"
)

// ...
type BasicLocation struct {
	mu sync.RWMutex

	id     shared.ID
	name   string
	jobs   []shared.Job
	agents []shared.Agent
}
