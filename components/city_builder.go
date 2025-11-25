package components

import (
	"fmt"
	"github.com/fallra1n/humanity/config"
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
	hospital := NewBuilding(buildingID, Hospital, fmt.Sprintf("%s Hospital", name), config.SmallCityHospitalCapacity, city)
	city.Buildings[hospital] = true
	buildingID++

	// 1 School
	school := NewBuilding(buildingID, School, fmt.Sprintf("%s School", name), config.SmallCitySchoolCapacity, city)
	city.Buildings[school] = true
	buildingID++

	// 2 Workplaces
	for i := 1; i <= 2; i++ {
		workplace := NewBuilding(buildingID, Workplace, fmt.Sprintf("%s Office %d", name, i), config.SmallCityWorkplaceCapacity, city)
		
		// Create jobs for this workplace
		job := &Job{
			VacantPlaces: make(map[*Vacancy]uint64),
			HomeLocation: city,
			Building:     workplace,
		}

		// Create junior and senior positions
		salaryRange := config.SmallCityJuniorSalaryMax - config.SmallCityJuniorSalaryMin
		juniorVacancy := &Vacancy{
			Parent:       job,
			RequiredTags: make(map[string]bool),
			Payment:      config.SmallCityJuniorSalaryMin + utils.GlobalRandom.NextInt(salaryRange),
		}

		salaryRange = config.SmallCitySeniorSalaryMax - config.SmallCitySeniorSalaryMin
		seniorVacancy := &Vacancy{
			Parent:       job,
			RequiredTags: make(map[string]bool),
			Payment:      config.SmallCitySeniorSalaryMin + utils.GlobalRandom.NextInt(salaryRange),
		}

		// Some senior positions require education
		if i == 1 {
			seniorVacancy.RequiredTags["engineer_diploma"] = true
		}

		// Each vacancy has positions based on config
		vacancyRange := config.SmallCityVacanciesMax - config.SmallCityVacanciesMin
		job.VacantPlaces[juniorVacancy] = uint64(config.SmallCityVacanciesMin + utils.GlobalRandom.NextInt(vacancyRange))
		job.VacantPlaces[seniorVacancy] = uint64(config.SmallCityVacanciesMin + utils.GlobalRandom.NextInt(vacancyRange))

		workplace.AddJob(job)
		city.Jobs[job] = true
		city.Buildings[workplace] = true
		buildingID++
	}

	// 1 Entertainment center
	entertainment := NewBuilding(buildingID, Entertainment, fmt.Sprintf("%s Entertainment Center", name), config.SmallCityEntertainmentCapacity, city)
	city.Buildings[entertainment] = true
	buildingID++

	// 1 Cafe
	cafe := NewBuilding(buildingID, Cafe, fmt.Sprintf("%s Cafe", name), config.SmallCityCafeCapacity, city)
	city.Buildings[cafe] = true
	buildingID++

	// 1 Shop
	shop := NewBuilding(buildingID, Shop, fmt.Sprintf("%s Shop", name), config.SmallCityShopCapacity, city)
	city.Buildings[shop] = true
	buildingID++

	// 3 Residential houses
	for i := 1; i <= 3; i++ {
		house := NewBuilding(buildingID, ResidentialHouse, fmt.Sprintf("%s House %d", name, i), config.SmallCityHouseCapacity, city)
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
		hospital := NewBuilding(buildingID, Hospital, fmt.Sprintf("%s Hospital %d", name, i), config.LargeCityHospitalCapacity, city)
		city.Buildings[hospital] = true
		buildingID++
	}

	// 2 Schools
	for i := 1; i <= 2; i++ {
		school := NewBuilding(buildingID, School, fmt.Sprintf("%s School %d", name, i), config.LargeCitySchoolCapacity, city)
		city.Buildings[school] = true
		buildingID++
	}

	// 3 Workplaces
	for i := 1; i <= 3; i++ {
		workplace := NewBuilding(buildingID, Workplace, fmt.Sprintf("%s Office %d", name, i), config.LargeCityWorkplaceCapacity, city)
		
		// Create jobs for this workplace
		job := &Job{
			VacantPlaces: make(map[*Vacancy]uint64),
			HomeLocation: city,
			Building:     workplace,
		}

		// Create junior and senior positions
		salaryRange := config.LargeCityJuniorSalaryMax - config.LargeCityJuniorSalaryMin
		juniorVacancy := &Vacancy{
			Parent:       job,
			RequiredTags: make(map[string]bool),
			Payment:      config.LargeCityJuniorSalaryMin + utils.GlobalRandom.NextInt(salaryRange),
		}

		salaryRange = config.LargeCitySeniorSalaryMax - config.LargeCitySeniorSalaryMin
		seniorVacancy := &Vacancy{
			Parent:       job,
			RequiredTags: make(map[string]bool),
			Payment:      config.LargeCitySeniorSalaryMin + utils.GlobalRandom.NextInt(salaryRange),
		}

		// Some senior positions require education
		if i <= 2 {
			seniorVacancy.RequiredTags["engineer_diploma"] = true
		}

		// Each vacancy has positions based on config
		vacancyRange := config.LargeCityVacanciesMax - config.LargeCityVacanciesMin
		job.VacantPlaces[juniorVacancy] = uint64(config.LargeCityVacanciesMin + utils.GlobalRandom.NextInt(vacancyRange))
		job.VacantPlaces[seniorVacancy] = uint64(config.LargeCityVacanciesMin + utils.GlobalRandom.NextInt(vacancyRange))

		workplace.AddJob(job)
		city.Jobs[job] = true
		city.Buildings[workplace] = true
		buildingID++
	}

	// 1 Entertainment center
	entertainment := NewBuilding(buildingID, Entertainment, fmt.Sprintf("%s Entertainment Center", name), config.LargeCityEntertainmentCapacity, city)
	city.Buildings[entertainment] = true
	buildingID++

	// 2 Cafes
	for i := 1; i <= 2; i++ {
		cafe := NewBuilding(buildingID, Cafe, fmt.Sprintf("%s Cafe %d", name, i), config.LargeCityCafeCapacity, city)
		city.Buildings[cafe] = true
		buildingID++
	}

	// 2 Shops
	for i := 1; i <= 2; i++ {
		shop := NewBuilding(buildingID, Shop, fmt.Sprintf("%s Shop %d", name, i), config.LargeCityShopCapacity, city)
		city.Buildings[shop] = true
		buildingID++
	}

	// 3 Residential houses
	for i := 1; i <= 3; i++ {
		house := NewBuilding(buildingID, ResidentialHouse, fmt.Sprintf("%s House %d", name, i), config.LargeCityHouseCapacity, city)
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

// CalculateFamilyFriendlyCoefficient calculates how family-friendly a city is
// based on its infrastructure (hospitals, schools, etc.)
func CalculateFamilyFriendlyCoefficient(city *Location) float64 {
	city.Mu.RLock()
	defer city.Mu.RUnlock()
	
	coefficient := 1.0 // Base coefficient
	
	buildingCounts := make(map[BuildingType]int)
	for building := range city.Buildings {
		buildingCounts[building.Type]++
	}
	
	// Hospitals increase family coefficient (medical care for children)
	coefficient += float64(buildingCounts[Hospital]) * 0.3
	
	// Schools increase family coefficient (education for children)
	coefficient += float64(buildingCounts[School]) * 0.4
	
	// Entertainment centers increase coefficient (family activities)
	coefficient += float64(buildingCounts[Entertainment]) * 0.2
	
	// Cafes and shops provide convenience for families
	coefficient += float64(buildingCounts[Cafe]) * 0.1
	coefficient += float64(buildingCounts[Shop]) * 0.1
	
	return coefficient
}