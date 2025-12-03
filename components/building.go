package components

import "sync"

// BuildingType represents the type of building
type BuildingType string

const (
	Hospital         BuildingType = "hospital"
	School           BuildingType = "school"
	Workplace        BuildingType = "workplace"
	Entertainment    BuildingType = "entertainment"
	Cafe             BuildingType = "cafe"
	Shop             BuildingType = "shop"
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

	// For residential buildings - apartment sales
	ApartmentsForSale []*Human // List of apartments available for sale (previous owners)
	ApartmentPrice    int64    // Price per apartment in rubles

	// Thread safety
	Mu sync.RWMutex
}

// NewBuilding creates a new building
func NewBuilding(id int, buildingType BuildingType, name string, capacity int, location *Location) *Building {
	building := &Building{
		ID:        id,
		Type:      buildingType,
		Name:      name,
		Location:  location,
		Jobs:      make(map[*Job]bool),
		Residents: make(map[*Human]bool),
		Capacity:  capacity,
		Occupied:  0,
	}

	// Initialize apartment sales for residential buildings
	if buildingType == ResidentialHouse {
		// Different prices for small and large cities
		if location.Name == "Greenville" {
			building.ApartmentPrice = 2000000 // 2 million rubles for small city
		} else {
			building.ApartmentPrice = 3000000 // 3 million rubles for large city
		}
		building.ApartmentsForSale = make([]*Human, 0) // Initially no apartments for sale
	}

	return building
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
	human.CurrentBuilding = b // Start at home
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

// PutApartmentForSale puts an apartment up for sale when a resident moves out
func (b *Building) PutApartmentForSale(previousOwner *Human) {
	if b.Type != ResidentialHouse {
		return
	}

	b.Mu.Lock()
	defer b.Mu.Unlock()

	b.ApartmentsForSale = append(b.ApartmentsForSale, previousOwner)
}

// BuyApartment allows a human to buy an apartment
func (b *Building) BuyApartment(human *Human) bool {
	if b.Type != ResidentialHouse {
		return false
	}

	b.Mu.Lock()
	defer b.Mu.Unlock()

	if len(b.ApartmentsForSale) <= 0 {
		return false
	}

	// Check if human has enough money
	if human.Money < b.ApartmentPrice {
		return false
	}

	// Process the purchase
	human.Money -= b.ApartmentPrice
	// Remove the first apartment from sale list
	b.ApartmentsForSale = b.ApartmentsForSale[1:]
	b.Residents[human] = true
	b.Occupied++
	human.ResidentialBuilding = b
	human.CurrentBuilding = b

	return true
}

// MoveToSpouse moves a person to their spouse's residential building
func (b *Building) MoveToSpouse(human *Human, spouse *Human) bool {
	if b.Type != ResidentialHouse {
		return false
	}

	spouseBuilding := spouse.ResidentialBuilding
	if spouseBuilding == nil || spouseBuilding.Type != ResidentialHouse {
		return false
	}

	// Lock current building first
	b.Mu.Lock()

	// Remove from current building and put apartment for sale
	if b.Residents[human] {
		delete(b.Residents, human)
		b.Occupied--
		b.ApartmentsForSale = append(b.ApartmentsForSale, human) // Put apartment up for sale directly
	}
	b.Mu.Unlock()

	// Lock spouse's building
	spouseBuilding.Mu.Lock()
	defer spouseBuilding.Mu.Unlock()

	// Add to spouse's building (no capacity check since it's a move within existing capacity)
	spouseBuilding.Residents[human] = true
	human.ResidentialBuilding = spouseBuilding
	human.CurrentBuilding = spouseBuilding

	return true
}
