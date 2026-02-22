package components

import (
	"fmt"

	"github.com/fallra1n/humanity/src/config"
	"github.com/fallra1n/humanity/src/utils"
)

// CreateSmallCity создает малый город с 10 зданиями
// 1 больница, 1 школа, 2 рабочих места, 1 развлечение, 1 кафе, 1 магазин, 3 жилых дома
func CreateSmallCity(name string) *Location {
	city := &Location{
		Name:      name,
		Buildings: make(map[*Building]bool),
		Jobs:      make(map[*Job]bool),
		Humans:    make(map[*Human]bool),
		Paths:     make(map[*Path]bool),
	}

	buildingID := 1

	// 1 Больница
	lat, lon := config.GetCoordinateForBuilding(name, "hospital", 0)
	hospital := NewBuildingWithCoordinates(buildingID, Hospital, fmt.Sprintf("%s Hospital", name), config.SmallCityHospitalCapacity, city, lat, lon)
	city.Buildings[hospital] = true
	buildingID++

	// 1 Школа
	lat, lon = config.GetCoordinateForBuilding(name, "school", 0)
	school := NewBuildingWithCoordinates(buildingID, School, fmt.Sprintf("%s School", name), config.SmallCitySchoolCapacity, city, lat, lon)
	city.Buildings[school] = true
	buildingID++

	// 2 Рабочих места
	for i := 1; i <= 2; i++ {
		lat, lon := config.GetCoordinateForBuilding(name, "workplace", i-1)
		workplace := NewBuildingWithCoordinates(buildingID, Workplace, fmt.Sprintf("%s Office %d", name, i), config.SmallCityWorkplaceCapacity, city, lat, lon)

		// Создать работы для этого рабочего места
		job := &Job{
			VacantPlaces: make(map[*Vacancy]uint64),
			HomeLocation: city,
			Building:     workplace,
		}

		// Создать младшие и старшие позиции
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

		// Некоторые старшие позиции требуют образования
		if i == 1 {
			seniorVacancy.RequiredTags["engineer_diploma"] = true
		}

		// Каждая вакансия имеет позиции на основе конфигурации
		vacancyRange := config.SmallCityVacanciesMax - config.SmallCityVacanciesMin
		job.VacantPlaces[juniorVacancy] = uint64(config.SmallCityVacanciesMin + utils.GlobalRandom.NextInt(vacancyRange))
		job.VacantPlaces[seniorVacancy] = uint64(config.SmallCityVacanciesMin + utils.GlobalRandom.NextInt(vacancyRange))

		workplace.AddJob(job)
		city.Jobs[job] = true
		city.Buildings[workplace] = true
		buildingID++
	}

	// 1 Развлекательный центр
	lat, lon = config.GetCoordinateForBuilding(name, "entertainment", 0)
	entertainment := NewBuildingWithCoordinates(buildingID, Entertainment, fmt.Sprintf("%s Entertainment Center", name), config.SmallCityEntertainmentCapacity, city, lat, lon)
	city.Buildings[entertainment] = true
	buildingID++

	// 1 Кафе
	lat, lon = config.GetCoordinateForBuilding(name, "cafe", 0)
	cafe := NewBuildingWithCoordinates(buildingID, Cafe, fmt.Sprintf("%s Cafe", name), config.SmallCityCafeCapacity, city, lat, lon)
	city.Buildings[cafe] = true
	buildingID++

	// 1 Магазин
	lat, lon = config.GetCoordinateForBuilding(name, "shop", 0)
	shop := NewBuildingWithCoordinates(buildingID, Shop, fmt.Sprintf("%s Shop", name), config.SmallCityShopCapacity, city, lat, lon)
	city.Buildings[shop] = true
	buildingID++

	// 3 Жилых дома
	for i := 1; i <= 3; i++ {
		lat, lon := config.GetCoordinateForBuilding(name, "residential_house", i-1)
		house := NewBuildingWithCoordinates(buildingID, ResidentialHouse, fmt.Sprintf("%s House %d", name, i), config.SmallCityHouseCapacity, city, lat, lon)
		city.Buildings[house] = true
		buildingID++
	}

	return city
}

// CreateLargeCity создает большой город с 15 зданиями
// 2 больницы, 2 школы, 3 рабочих места, 1 развлечение, 2 кафе, 2 магазина, 3 жилых дома
func CreateLargeCity(name string) *Location {
	city := &Location{
		Name:      name,
		Buildings: make(map[*Building]bool),
		Jobs:      make(map[*Job]bool),
		Humans:    make(map[*Human]bool),
		Paths:     make(map[*Path]bool),
	}

	buildingID := 1

	// 2 Больницы
	for i := 1; i <= 2; i++ {
		lat, lon := config.GetCoordinateForBuilding(name, "hospital", i-1)
		hospital := NewBuildingWithCoordinates(buildingID, Hospital, fmt.Sprintf("%s Hospital %d", name, i), config.LargeCityHospitalCapacity, city, lat, lon)
		city.Buildings[hospital] = true
		buildingID++
	}

	// 2 Школы
	for i := 1; i <= 2; i++ {
		lat, lon := config.GetCoordinateForBuilding(name, "school", i-1)
		school := NewBuildingWithCoordinates(buildingID, School, fmt.Sprintf("%s School %d", name, i), config.LargeCitySchoolCapacity, city, lat, lon)
		city.Buildings[school] = true
		buildingID++
	}

	// 3 Рабочих места
	for i := 1; i <= 3; i++ {
		lat, lon := config.GetCoordinateForBuilding(name, "workplace", i-1)
		workplace := NewBuildingWithCoordinates(buildingID, Workplace, fmt.Sprintf("%s Office %d", name, i), config.LargeCityWorkplaceCapacity, city, lat, lon)

		// Создать работы для этого рабочего места
		job := &Job{
			VacantPlaces: make(map[*Vacancy]uint64),
			HomeLocation: city,
			Building:     workplace,
		}

		// Создать младшие и старшие позиции
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

		// Некоторые старшие позиции требуют образования
		if i <= 2 {
			seniorVacancy.RequiredTags["engineer_diploma"] = true
		}

		// Каждая вакансия имеет позиции на основе конфигурации
		vacancyRange := config.LargeCityVacanciesMax - config.LargeCityVacanciesMin
		job.VacantPlaces[juniorVacancy] = uint64(config.LargeCityVacanciesMin + utils.GlobalRandom.NextInt(vacancyRange))
		job.VacantPlaces[seniorVacancy] = uint64(config.LargeCityVacanciesMin + utils.GlobalRandom.NextInt(vacancyRange))

		workplace.AddJob(job)
		city.Jobs[job] = true
		city.Buildings[workplace] = true
		buildingID++
	}

	// 1 Развлекательный центр
	lat, lon := config.GetCoordinateForBuilding(name, "entertainment", 0)
	entertainment := NewBuildingWithCoordinates(buildingID, Entertainment, fmt.Sprintf("%s Entertainment Center", name), config.LargeCityEntertainmentCapacity, city, lat, lon)
	city.Buildings[entertainment] = true
	buildingID++

	// 2 Кафе
	for i := 1; i <= 2; i++ {
		lat, lon := config.GetCoordinateForBuilding(name, "cafe", i-1)
		cafe := NewBuildingWithCoordinates(buildingID, Cafe, fmt.Sprintf("%s Cafe %d", name, i), config.LargeCityCafeCapacity, city, lat, lon)
		city.Buildings[cafe] = true
		buildingID++
	}

	// 2 Магазина
	for i := 1; i <= 2; i++ {
		lat, lon := config.GetCoordinateForBuilding(name, "shop", i-1)
		shop := NewBuildingWithCoordinates(buildingID, Shop, fmt.Sprintf("%s Shop %d", name, i), config.LargeCityShopCapacity, city, lat, lon)
		city.Buildings[shop] = true
		buildingID++
	}

	// 3 Жилых дома
	for i := 1; i <= 3; i++ {
		lat, lon := config.GetCoordinateForBuilding(name, "residential_house", i-1)
		house := NewBuildingWithCoordinates(buildingID, ResidentialHouse, fmt.Sprintf("%s House %d", name, i), config.LargeCityHouseCapacity, city, lat, lon)
		city.Buildings[house] = true
		buildingID++
	}

	return city
}

// GetResidentialBuildings возвращает все жилые здания в локации
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

// GetWorkplaceBuildings возвращает все рабочие здания в локации
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

// PrintCityInfo выводит подробную информацию о городе
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

// CalculateFamilyFriendlyCoefficient вычисляет насколько город дружелюбен к семьям
// на основе его инфраструктуры (больницы, школы и т.д.)
func CalculateFamilyFriendlyCoefficient(city *Location) float64 {
	city.Mu.RLock()
	defer city.Mu.RUnlock()

	coefficient := 1.0 // Базовый коэффициент

	buildingCounts := make(map[BuildingType]int)
	for building := range city.Buildings {
		buildingCounts[building.Type]++
	}

	// Больницы увеличивают семейный коэффициент (медицинская помощь детям)
	coefficient += float64(buildingCounts[Hospital]) * 0.3

	// Школы увеличивают семейный коэффициент (образование для детей)
	coefficient += float64(buildingCounts[School]) * 0.4

	// Развлекательные центры увеличивают коэффициент (семейные мероприятия)
	coefficient += float64(buildingCounts[Entertainment]) * 0.2

	// Кафе и магазины обеспечивают удобство для семей
	coefficient += float64(buildingCounts[Cafe]) * 0.1
	coefficient += float64(buildingCounts[Shop]) * 0.1

	return coefficient
}
