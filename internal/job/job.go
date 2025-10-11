package job

import (
	"sync"

	"github.com/fallra1n/humanity/internal/shared"
)

// ...
type BasicJob struct {
	mu sync.RWMutex

	id        shared.ID
	location  shared.Location
	vacancies []shared.Vacancy
}
