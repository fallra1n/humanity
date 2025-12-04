package components

import (
	"github.com/fallra1n/humanity/utils"
	"sync"
)

// Vacancy представляет вакансию
type Vacancy struct {
	Parent       *Job
	RequiredTags map[string]bool
	Payment      int
}

// Job представляет рабочее место
type Job struct {
	VacantPlaces map[*Vacancy]uint64
	HomeLocation *Location
	Building     *Building
	Mu           sync.RWMutex
}

// Location представляет место, где люди живут и работают
type Location struct {
	Name      string
	Buildings map[*Building]bool
	Jobs      map[*Job]bool
	Humans    map[*Human]bool
	Paths     map[*Path]bool
	Mu        sync.RWMutex
}

// Path представляет соединение между локациями
type Path struct {
	From  *Location
	To    *Location
	Price uint64
	Time  uint64
}

// Splash представляет временную мысль или потребность
type Splash struct {
	Name       string
	Tags       map[string]bool
	AppearTime uint64
	LifeLength uint64
}

// NewSplash создает новый всплеск
func NewSplash(name string, tags []string, lifeLength uint64) *Splash {
	tagSet := make(map[string]bool)
	for _, tag := range tags {
		tagSet[tag] = true
	}

	return &Splash{
		Name:       name,
		Tags:       tagSet,
		AppearTime: utils.GlobalTick.Get(),
		LifeLength: lifeLength,
	}
}

// IsExpired проверяет, истек ли всплеск
func (s *Splash) IsExpired() bool {
	return utils.GlobalTick.Get()-s.AppearTime > s.LifeLength
}
