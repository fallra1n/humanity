package economy

import (
	"sync"

	"github.com/fallra1n/humanity/internal/shared"
)

// ...
type EconomicSystem struct {
	mu sync.RWMutex

	Locations []shared.Location
	Jobs      []shared.Job
	Vacancies []shared.Vacancy
}

func NewEconomicSystem() *EconomicSystem {
	return &EconomicSystem{
		Locations: make([]shared.Location, 0),
		Jobs:      make([]shared.Job, 0),
		Vacancies: make([]shared.Vacancy, 0),
	}
}
