package vacancy

import (
	"sync"

	"github.com/fallra1n/humanity/internal/shared"
)

// ...
type BasicVacancy struct {
	mu sync.RWMutex

	id           shared.ID
	job          shared.Job
	requiredTags shared.Tags
	salary       shared.Money
	employee     shared.Agent
	available    bool
}
