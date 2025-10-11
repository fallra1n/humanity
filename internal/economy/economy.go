package economy

import (
	"sync"

	"github.com/fallra1n/humanity/internal/shared"
)

// ...
type EconomicSystem struct {
	mu sync.RWMutex

	locations []shared.Location
	jobs      []shared.Job
	vacancies []shared.Vacancy
}
