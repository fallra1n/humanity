package src

import (
	"fmt"

	"github.com/fallra1n/humanity/src/components"
)

// SimulationStatistics содержит статистику симуляции
type SimulationStatistics struct {
	AliveCount                 int
	CompletedTargetsCount      int
	TotalMoney                 int64
	EmployedCount              int
	TotalItems                 int
	MaleCount                  int
	FemaleCount                int
	MarriedCount               int
	ChildrenCount              int
	PregnantCount              int
	TotalChildren              int
	MoveCount                  int
	ApartmentsForSaleLargeCity int
	PeopleWithoutHousing       int
	TotalFriends               int
	PeopleWithFriends          int
	PeopleAtWork               int
	PeopleAtHome               int
	TargetStats                map[string]int
}

// CalculateStatistics вычисляет статистику симуляции
func CalculateStatistics(people []*components.Human, smallCity, largeCity *components.Location) SimulationStatistics {
	stats := SimulationStatistics{
		TargetStats: make(map[string]int),
	}

	// Основная статистика по людям
	for _, person := range people {
		if !person.Dead {
			stats.AliveCount++
		}
		if person.Job != nil {
			stats.EmployedCount++
		}
		if person.Gender == components.Male {
			stats.MaleCount++
		} else {
			stats.FemaleCount++
		}
		if person.MaritalStatus == components.Married {
			stats.MarriedCount++
		}
		if person.Age < 18.0 {
			stats.ChildrenCount++
		}
		if person.IsPregnant {
			stats.PregnantCount++
		}
		stats.TotalChildren += len(person.Children)
		stats.CompletedTargetsCount += len(person.CompletedGlobalTargets)
		stats.TotalMoney += person.Money
		stats.TotalItems += len(person.Items)

		// Подсчитать людей без жилья
		if person.ResidentialBuilding == nil {
			stats.PeopleWithoutHousing++
		}

		// Подсчитать переезды из-за брака
		if person.MaritalStatus == components.Married && person.Gender == components.Female {
			// Проверить, переехала ли женщина в здание мужа
			if person.Spouse != nil && person.ResidentialBuilding != nil &&
				person.Spouse.ResidentialBuilding != nil &&
				person.ResidentialBuilding == person.Spouse.ResidentialBuilding {
				stats.MoveCount++
			}
		}

		// Статистика дружбы и местоположения (только для живых)
		if !person.Dead {
			stats.TotalFriends += len(person.Friends)
			if len(person.Friends) > 0 {
				stats.PeopleWithFriends++
			}

			// Подсчитать текущие местоположения
			if person.CurrentBuilding != nil {
				if person.CurrentBuilding == person.WorkBuilding {
					stats.PeopleAtWork++
				} else if person.CurrentBuilding == person.ResidentialBuilding {
					stats.PeopleAtHome++
				}
			}
		}

		// Статистика выполнения целей
		for target := range person.CompletedGlobalTargets {
			stats.TargetStats[target.Name]++
		}
	}

	return stats
}

// PrintSimulationSummary выводит сводную статистику симуляции
func PrintSimulationSummary(people []*components.Human, smallCity, largeCity *components.Location) {
	stats := CalculateStatistics(people, smallCity, largeCity)

	fmt.Println("========================================")
	fmt.Println("SIMULATION SUMMARY")
	fmt.Println("========================================")

	fmt.Printf("Population: %d humans (%d male, %d female)\n",
		len(people), stats.MaleCount, stats.FemaleCount)
	fmt.Printf("Survival Rate: %d/%d humans alive (%.1f%%)\n",
		stats.AliveCount, len(people), float64(stats.AliveCount)/float64(len(people))*100)
	fmt.Printf("Employment Rate: %d/%d humans employed (%.1f%%)\n",
		stats.EmployedCount, stats.AliveCount, float64(stats.EmployedCount)/float64(stats.AliveCount)*100)
	fmt.Printf("Marriage Rate: %d/%d humans married (%.1f%%)\n",
		stats.MarriedCount, stats.AliveCount, float64(stats.MarriedCount)/float64(stats.AliveCount)*100)
	fmt.Printf("Children: %d children under 18 (%.1f%% of population)\n",
		stats.ChildrenCount, float64(stats.ChildrenCount)/float64(len(people))*100)
	fmt.Printf("Pregnancies: %d women currently pregnant\n", stats.PregnantCount)
	fmt.Printf("Total Children Born: %d children (average %.1f per adult)\n",
		stats.TotalChildren/2, float64(stats.TotalChildren)/float64(len(people)-stats.ChildrenCount)) // Divide by 2 since both parents count the same child
	fmt.Printf("Marriage Moves: %d women moved to husband's building\n", stats.MoveCount)
	fmt.Printf("People without housing: %d/%d (%.1f%%)\n",
		stats.PeopleWithoutHousing, len(people), float64(stats.PeopleWithoutHousing)/float64(len(people))*100)
	fmt.Printf("Total Completed Global Targets: %d\n", stats.CompletedTargetsCount)
	fmt.Printf("Average Completed Targets per Person: %.1f\n",
		float64(stats.CompletedTargetsCount)/float64(len(people)))
	fmt.Printf("Total Money in Economy: %d rubles\n", stats.TotalMoney)
	fmt.Printf("Average Money per Person: %d rubles\n", stats.TotalMoney/int64(len(people)))
	fmt.Printf("Total Items Acquired: %d\n", stats.TotalItems)

	// Статистика новой функциональности
	fmt.Printf("\n=== NEW FEATURES STATISTICS ===\n")
	fmt.Printf("Social Connections:\n")
	fmt.Printf("  Total Friendships: %d\n", stats.TotalFriends)
	fmt.Printf("  People with Friends: %d/%d (%.1f%%)\n",
		stats.PeopleWithFriends, stats.AliveCount, float64(stats.PeopleWithFriends)/float64(stats.AliveCount)*100)
	fmt.Printf("  Average Friends per Person: %.1f\n",
		float64(stats.TotalFriends)/float64(stats.AliveCount))

	fmt.Printf("Location Distribution:\n")
	fmt.Printf("  People at Work: %d (%.1f%%)\n",
		stats.PeopleAtWork, float64(stats.PeopleAtWork)/float64(stats.AliveCount)*100)
	fmt.Printf("  People at Home: %d (%.1f%%)\n",
		stats.PeopleAtHome, float64(stats.PeopleAtHome)/float64(stats.AliveCount)*100)
	fmt.Printf("  Other Locations: %d (%.1f%%)\n",
		stats.AliveCount-stats.PeopleAtWork-stats.PeopleAtHome,
		float64(stats.AliveCount-stats.PeopleAtWork-stats.PeopleAtHome)/float64(stats.AliveCount)*100)

	// Статистика выполнения целей
	fmt.Println("\nTarget Completion Statistics:")
	for targetName, count := range stats.TargetStats {
		fmt.Printf("  %s: %d people (%.1f%%)\n",
			targetName, count, float64(count)/float64(len(people))*100)
	}
}

// PrintInitialStatistics выводит начальную статистику всех людей
func PrintInitialStatistics(people []*components.Human) {
	fmt.Println("========================================")
	fmt.Println("INITIAL STATE OF ALL HUMANS")
	fmt.Println("========================================")
	for i, person := range people {
		person.PrintInitialInfo(i + 1)
	}
}

// PrintFinalStatistics выводит финальную статистику всех людей
func PrintFinalStatistics(people []*components.Human) {
	fmt.Println("\n========================================")
	fmt.Println("FINAL STATE OF ALL HUMANS")
	fmt.Println("========================================")
	for i, person := range people {
		person.PrintFinalInfo(i + 1)
	}
}
