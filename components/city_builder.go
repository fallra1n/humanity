package components

import (
	"fmt"
	"github.com/fallra1n/humanity/utils"
)

// CreateSmallCity creates a small city with 10 buildings
// 1 hospital, 1 school, 2 workplaces, 1 entertainment, 1 cafe, 1 shop, 3 residential houses
func CreateSmallCity(name string) *Location {
	city := &Location{
		Name:      name,
		Buildings: make(map[*Building]bool),
		Jobs:      make(map[*Job]bool),
		Humans:    make(map[*Human]bool),
		Paths:     make(map[*Path]bool),
	}

	buildingID := 1

	// 1 Hospital
	hospital := NewBuilding(buildingID, Hospital, fmt.Sprintf("%s Hospital", name), 50, city)
	city.Buildings[hospital] = true
	buildingID++

	// 1 School
	school := NewBuilding(buildingID, School, fmt.Sprintf("%s School", name), 200, city)
	city.Buildings[school] = true
	buildingID++

	// 2 Workplaces
	for i := 1; i <= 2; i++ {
		workplace := NewBuilding(buildingID, Workplace, fmt.Sprintf("%s Office %d", name, i), 100, city)
		
		// Create jobs for this workplace
		job := &Job{
			VacantPlaces: make(map[*Vacancy]uint64),
			HomeLocation: city,
			Building:     workplace,
		}

		// Create junior and senior positions
		juniorVacancy := &Vacancy{
			Parent:       job,
			RequiredTags: make(map[string]bool),
			Payment:      30000 + utils.GlobalRandom.NextInt(15000), // 30-45k rubles
		}

		seniorVacancy := &Vacancy{
			Parent:       job,
			RequiredTags: make(map[string]bool),
			Payment:      50000 + utils.GlobalRandom.NextInt(25000), // 50-75k rubles
		}

		// Some senior positions require education
		if i == 1 {
			seniorVacancy.RequiredTags["engineer_diploma"] = true
		}

		// Each vacancy has 3-7 positions in small city
		job.VacantPlaces[juniorVacancy] = uint64(3 + utils.GlobalRandom.NextInt(5))
		job.VacantPlaces[seniorVacancy] = uint64(3 + utils.GlobalRandom.NextInt(5))

		workplace.AddJob(job)
		city.Jobs[job] = true
		city.Buildings[workplace] = true
		buildingID++
	}

	// 1 Entertainment center
	entertainment := NewBuilding(buildingID, Entertainment, fmt.Sprintf("%s Entertainment Center", name), 150, city)
	city.Buildings[entertainment] = true
	buildingID++

	// 1 Cafe
	cafe := NewBuilding(buildingID, Cafe, fmt.Sprintf("%s Cafe", name), 30, city)
	city.Buildings[cafe] = true
	buildingID++

	// 1 Shop
	shop := NewBuilding(buildingID, Shop, fmt.Sprintf("%s Shop", name), 40, city)
	city.Buildings[shop] = true
	buildingID++

	// 3 Residential houses
	for i := 1; i <= 3; i++ {
		house := NewBuilding(buildingID, ResidentialHouse, fmt.Sprintf("%s House %d", name, i), 15, city)
		city.Buildings[house] = true
		buildingID++
	}

	return city
}

// CreateLargeCity creates a large city with 15 buildings
// 2 hospitals, 2 schools, 3 workplaces, 1 entertainment, 2 cafes, 2 shops, 3 residential houses
func CreateLargeCity(name string) *Location {
	city := &Location{
		Name:      name,
		Buildings: make(map[*Building]bool),
		Jobs:      make(map[*Job]bool),
		Humans:    make(map[*Human]bool),
		Paths:     make(map[*Path]bool),
	}

	buildingID := 1

	// 2 Hospitals
	for i := 1; i <= 2; i++ {
		hospital := NewBuilding(buildingID, Hospital, fmt.Sprintf("%s Hospital %d", name, i), 75, city)
		city.Buildings[hospital] = true
		buildingID++
	}

	// 2 Schools
	for i := 1; i <= 2; i++ {
		school := NewBuilding(buildingID, School, fmt.Sprintf("%s School %d", name, i), 300, city)
		city.Buildings[school] = true
		buildingID++
	}

	// 3 Workplaces
	for i := 1; i <= 3; i++ {
		workplace := NewBuilding(buildingID, Workplace, fmt.Sprintf("%s Office %d", name, i), 150, city)
		
		// Create jobs for this workplace
		job := &Job{
			VacantPlaces: make(map[*Vacancy]uint64),
			HomeLocation: city,
			Building:     workplace,
		}

		// Create junior and senior positions
		juniorVacancy := &Vacancy{
			Parent:       job,
			RequiredTags: make(map[string]bool),
			Payment:      35000 + utils.GlobalRandom.NextInt(20000), // 35-55k rubles (higher in large city)
		}

		seniorVacancy := &Vacancy{
			Parent:       job,
			RequiredTags: make(map[string]bool),
			Payment:      60000 + utils.GlobalRandom.NextInt(30000), // 60-90k rubles (higher in large city)
		}

		// Some senior positions require education
		if i <= 2 {
			seniorVacancy.RequiredTags["engineer_diploma"] = true
		}

		// Each vacancy has 5-10 positions in large city
		job.VacantPlaces[juniorVacancy] = uint64(5 + utils.GlobalRandom.NextInt(6))
		job.VacantPlaces[seniorVacancy] = uint64(5 + utils.GlobalRandom.NextInt(6))

		workplace.AddJob(job)
		city.Jobs[job] = true
		city.Buildings[workplace] = true
		buildingID++
	}

	// 1 Entertainment center
	entertainment := NewBuilding(buildingID, Entertainment, fmt.Sprintf("%s Entertainment Center", name), 200, city)
	city.Buildings[entertainment] = true
	buildingID++

	// 2 Cafes
	for i := 1; i <= 2; i++ {
		cafe := NewBuilding(buildingID, Cafe, fmt.Sprintf("%s Cafe %d", name, i), 40, city)
		city.Buildings[cafe] = true
		buildingID++
	}

	// 2 Shops
	for i := 1; i <= 2; i++ {
		shop := NewBuilding(buildingID, Shop, fmt.Sprintf("%s Shop %d", name, i), 50, city)
		city.Buildings[shop] = true
		buildingID++
	}

	// 3 Residential houses
	for i := 1; i <= 3; i++ {
		house := NewBuilding(buildingID, ResidentialHouse, fmt.Sprintf("%s House %d", name, i), 25, city)
		city.Buildings[house] = true
		buildingID++
	}

	return city
}

// GetResidentialBuildings returns all residential buildings in a location
func GetResidentialBuildings(location *Location) []*Building {
	location.Mu.RLock()
	defer location.Mu.RUnlock()

	var residential []*Building
	for building := range location.Buildings {
		if building.Type == ResidentialHouse {
			residential = append(residential, building)
		}
	}
	return residential
}

// GetWorkplaceBuildings returns all workplace buildings in a location
func GetWorkplaceBuildings(location *Location) []*Building {
	location.Mu.RLock()
	defer location.Mu.RUnlock()

	var workplaces []*Building
	for building := range location.Buildings {
		if building.Type == Workplace {
			workplaces = append(workplaces, building)
		}
	}
	return workplaces
}

// PrintCityInfo prints detailed information about a city
func PrintCityInfo(city *Location) {
	city.Mu.RLock()
	defer city.Mu.RUnlock()

	fmt.Printf("\n=== %s City Information ===\n", city.Name)
	fmt.Printf("Total buildings: %d\n", len(city.Buildings))

	buildingCounts := make(map[BuildingType]int)
	for building := range city.Buildings {
		buildingCounts[building.Type]++
	}

	for buildingType, count := range buildingCounts {
		fmt.Printf("  %s: %d\n", buildingType, count)
	}

	fmt.Printf("Total jobs: %d\n", len(city.Jobs))
	fmt.Printf("Total residents: %d\n", len(city.Humans))
}