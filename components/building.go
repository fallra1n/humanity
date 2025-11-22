package components

import "sync"

// BuildingType represents the type of building
type BuildingType string

const (
	Hospital        BuildingType = "hospital"
	School          BuildingType = "school"
	Workplace       BuildingType = "workplace"
	Entertainment   BuildingType = "entertainment"
	Cafe            BuildingType = "cafe"
	Shop            BuildingType = "shop"
	ResidentialHouse BuildingType = "residential_house"
)

// Building represents a structure in a location
type Building struct {
	ID       int
	Type     BuildingType
	Name     string
	Location *Location
	
	// For workplaces - contains job vacancies
	Jobs map[*Job]bool
	
	// For residential - contains residents
	Residents map[*Human]bool
	
	// General capacity and current occupancy
	Capacity int
	Occupied int
	
	// Thread safety
	Mu sync.RWMutex
}

// NewBuilding creates a new building
func NewBuilding(id int, buildingType BuildingType, name string, capacity int, location *Location) *Building {
	return &Building{
		ID:        id,
		Type:      buildingType,
		Name:      name,
		Location:  location,
		Jobs:      make(map[*Job]bool),
		Residents: make(map[*Human]bool),
		Capacity:  capacity,
		Occupied:  0,
	}
}

// AddJob adds a job to a workplace building
func (b *Building) AddJob(job *Job) bool {
	if b.Type != Workplace {
		return false
	}
	
	b.Mu.Lock()
	defer b.Mu.Unlock()
	
	b.Jobs[job] = true
	job.Building = b
	return true
}

// AddResident adds a resident to a residential building
func (b *Building) AddResident(human *Human) bool {
	if b.Type != ResidentialHouse {
		return false
	}
	
	b.Mu.Lock()
	defer b.Mu.Unlock()
	
	if b.Occupied >= b.Capacity {
		return false
	}
	
	b.Residents[human] = true
	b.Occupied++
	human.ResidentialBuilding = b
	return true
}

// RemoveResident removes a resident from a residential building
func (b *Building) RemoveResident(human *Human) {
	if b.Type != ResidentialHouse {
		return
	}
	
	b.Mu.Lock()
	defer b.Mu.Unlock()
	
	if b.Residents[human] {
		delete(b.Residents, human)
		b.Occupied--
		human.ResidentialBuilding = nil
	}
}

// GetAvailableJobs returns all available jobs in this building
func (b *Building) GetAvailableJobs() []*Vacancy {
	if b.Type != Workplace {
		return nil
	}
	
	b.Mu.RLock()
	defer b.Mu.RUnlock()
	
	var vacancies []*Vacancy
	for job := range b.Jobs {
		job.Mu.RLock()
		for vacancy, count := range job.VacantPlaces {
			if count > 0 {
				vacancies = append(vacancies, vacancy)
			}
		}
		job.Mu.RUnlock()
	}
	
	return vacancies
}

// HasCapacity checks if building has available capacity
func (b *Building) HasCapacity() bool {
	b.Mu.RLock()
	defer b.Mu.RUnlock()
	return b.Occupied < b.Capacity
}

// GetOccupancyRate returns the occupancy rate as a percentage
func (b *Building) GetOccupancyRate() float64 {
	b.Mu.RLock()
	defer b.Mu.RUnlock()
	
	if b.Capacity == 0 {
		return 0
	}
	return float64(b.Occupied) / float64(b.Capacity) * 100
}