package components

import (
	"github.com/fallra1n/humanity/utils"
	"sync"
)

// Vacancy represents a job opening
type Vacancy struct {
	Parent       *Job
	RequiredTags map[string]bool
	Payment      int
}

// Job represents a workplace
type Job struct {
	VacantPlaces map[*Vacancy]uint64
	HomeLocation *Location
	Building     *Building
	Mu           sync.RWMutex
}

// Location represents a place where humans live and work
type Location struct {
	Name      string
	Buildings map[*Building]bool
	Jobs      map[*Job]bool
	Humans    map[*Human]bool
	Paths     map[*Path]bool
	Mu        sync.RWMutex
}

// Path represents a connection between locations
type Path struct {
	From  *Location
	To    *Location
	Price uint64
	Time  uint64
}

// Splash represents a temporary thought or need
type Splash struct {
	Name       string
	Tags       map[string]bool
	AppearTime uint64
	LifeLength uint64
}

// NewSplash creates a new splash
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

// IsExpired checks if splash has expired
func (s *Splash) IsExpired() bool {
	return utils.GlobalTick.Get()-s.AppearTime > s.LifeLength
}
