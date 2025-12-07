package src

import (
	"fmt"

	"github.com/fallra1n/humanity/src/components"
	"github.com/fallra1n/humanity/src/config"
	"github.com/fallra1n/humanity/src/utils"
)

// PopulationStats содержит статистику созданной популяции
type PopulationStats struct {
	TotalPeople         int
	SmallCityEmployed   int
	LargeCityEmployed   int
	TotalEmployed       int
	SmallCityPopulation int
	LargeCityPopulation int
}

// CreateCityPopulation создает популяцию для одного города
func CreateCityPopulation(city *components.Location, targetPopulation int, globalTargets []*components.GlobalTarget) ([]*components.Human, int) {
	var people []*components.Human
	employedCount := int(float64(targetPopulation) * config.EmploymentRate)
	residentialBuildings := components.GetResidentialBuildings(city)

	for i := 0; i < targetPopulation; i++ {
		human := components.NewHuman(make(map[*components.Human]bool), city, globalTargets)
		human.Money = config.StartingMoney

		// Назначить жилое здание
		assigned := false
		for _, building := range residentialBuildings {
			if building.AddResident(human) {
				assigned = true
				break
			}
		}
		if !assigned {
			fmt.Printf("Warning: Could not assign residential building to %s human %d\n", city.Name, i+1)
		}

		// Трудоустройство на основе конфигурационного коэффициента
		actualEmployed := 0
		if i < employedCount {
			if assignJob(human, city) {
				actualEmployed++
			}
		}

		people = append(people, human)
		city.Humans[human] = true
	}

	// Подсчитать фактически трудоустроенных
	actualEmployedCount := 0
	for _, person := range people {
		if person.Job != nil {
			actualEmployedCount++
		}
	}

	return people, actualEmployedCount
}

// assignJob назначает работу человеку в городе
func assignJob(human *components.Human, city *components.Location) bool {
	var availableVacancies []*components.Vacancy

	// Искать работы в рабочих зданиях
	for building := range city.Buildings {
		if building.Type == components.Workplace {
			for job := range building.Jobs {
				for vacancy, count := range job.VacantPlaces {
					if count > 0 {
						canWork := true
						for requiredTag := range vacancy.RequiredTags {
							if human.Items[requiredTag] <= 0 {
								canWork = false
								break
							}
						}
						if canWork {
							availableVacancies = append(availableVacancies, vacancy)
						}
					}
				}
			}
		}
	}

	if len(availableVacancies) > 0 {
		chosenVacancy := availableVacancies[utils.GlobalRandom.NextInt(len(availableVacancies))]
		human.Job = chosenVacancy
		human.JobTime = uint64(utils.GlobalRandom.NextInt(config.MaxInitialWorkExperience))
		human.WorkBuilding = chosenVacancy.Parent.Building // Установить рабочее здание
		chosenVacancy.Parent.VacantPlaces[chosenVacancy]--
		return true
	}

	return false
}

// CreatePopulation создает всю популяцию для симуляции
func CreatePopulation(smallCity, largeCity *components.Location, globalTargets []*components.GlobalTarget) ([]*components.Human, PopulationStats) {
	var allPeople []*components.Human

	// Создать население малого города
	smallCityPeople, smallCityEmployed := CreateCityPopulation(smallCity, config.SmallCityPopulation, globalTargets)
	allPeople = append(allPeople, smallCityPeople...)

	// Создать население большого города
	largeCityPeople, largeCityEmployed := CreateCityPopulation(largeCity, config.LargeCityPopulation, globalTargets)
	allPeople = append(allPeople, largeCityPeople...)

	// Подсчитать общую статистику
	stats := PopulationStats{
		TotalPeople:         len(allPeople),
		SmallCityEmployed:   smallCityEmployed,
		LargeCityEmployed:   largeCityEmployed,
		TotalEmployed:       smallCityEmployed + largeCityEmployed,
		SmallCityPopulation: config.SmallCityPopulation,
		LargeCityPopulation: config.LargeCityPopulation,
	}

	return allPeople, stats
}

// PrintPopulationStats выводит статистику созданной популяции
func PrintPopulationStats(stats PopulationStats, smallCity, largeCity *components.Location) {
	fmt.Printf("\nCreated %d people total:\n", stats.TotalPeople)
	fmt.Printf("  %s: %d residents, %d employed (%.1f%%)\n",
		smallCity.Name, stats.SmallCityPopulation, stats.SmallCityEmployed,
		float64(stats.SmallCityEmployed)/float64(stats.SmallCityPopulation)*100)
	fmt.Printf("  %s: %d residents, %d employed (%.1f%%)\n",
		largeCity.Name, stats.LargeCityPopulation, stats.LargeCityEmployed,
		float64(stats.LargeCityEmployed)/float64(stats.LargeCityPopulation)*100)
	fmt.Printf("Total employment: %d employed, %d unemployed (%.1f%% employment rate)\n",
		stats.TotalEmployed, stats.TotalPeople-stats.TotalEmployed,
		float64(stats.TotalEmployed)/float64(stats.TotalPeople)*100)
}
