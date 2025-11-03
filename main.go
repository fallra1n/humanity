package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/fallra1n/humanity/components"
	"github.com/fallra1n/humanity/types"
	"github.com/fallra1n/humanity/utils"
)

func main() {
	// Load actions
	actions, err := LoadActions("actions.ini")
	if err != nil {
		log.Fatalf("Failed to load actions: %v", err)
	}

	// Load local targets
	localTargets, err := LoadLocalTargets("local.ini", actions)
	if err != nil {
		log.Fatalf("Failed to load local targets: %v", err)
	}

	// Load global targets
	globalTargets, err := LoadGlobalTargets("global.ini", localTargets)
	if err != nil {
		log.Fatalf("Failed to load global targets: %v", err)
	}

	// Create name maps for lookups
	actionMap, localMap, globalMap, err := CreateNameMaps(actions, localTargets, globalTargets)
	if err != nil {
		log.Fatalf("Failed to create name maps: %v", err)
	}

	// Create city and vacancies
	city := &types.Location{
		Jobs:   make(map[*types.Job]bool),
		Humans: make(map[*types.Human]bool),
		Paths:  make(map[*types.Path]bool),
	}

	allVacancies := components.CreateVacancies(city)

	// Create 100 humans
	var people []*types.Human
	for i := 0; i < 100; i++ {
		human := types.NewHuman(make(map[*types.Human]bool), city, globalTargets)
		human.Money = 10000 // Starting capital

		// 90% of people start with a job (unemployment rate ~7% in Russia)
		if i < 90 { // First 90 out of 100 people get jobs
			// Randomly assign to available vacancies
			availableVacancies := make([]*types.Vacancy, 0)
			for _, vacancy := range allVacancies {
				if vacancy.Parent.VacantPlaces[vacancy] > 0 {
					// Check if human can work (has required skills)
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

			if len(availableVacancies) > 0 {
				chosenVacancy := availableVacancies[utils.GlobalRandom.NextInt(len(availableVacancies))]
				human.Job = chosenVacancy
				human.JobTime = uint64(utils.GlobalRandom.NextInt(2000)) // Random work experience 0-2000 hours
				chosenVacancy.Parent.VacantPlaces[chosenVacancy]--
			}
		}

		people = append(people, human)
		city.Humans[human] = true
	}

	fmt.Printf("\nCreated %d people, %d employed, %d unemployed\n",
		len(people), len(people)-10, 10)

	// Print initial state of all humans
	fmt.Println("========================================")
	fmt.Println("INITIAL STATE OF ALL HUMANS")
	fmt.Println("========================================")
	for i, person := range people {
		person.PrintInitialInfo(i + 1)
	}

	// Simulation parameters
	const hoursPerYear = 365 * 24       // 1 year in hours
	const totalHours = 1 * hoursPerYear // Total simulation time

	var iterateTimer time.Duration

	// Main simulation loop
	for hour := uint64(0); hour < totalHours; hour++ {
		startTime := time.Now()

		wg := sync.WaitGroup{}

		// Process each person
		for _, person := range people {
			wg.Add(1)

			go func(person *types.Human) {
				defer wg.Done()
				if !person.Dead {
					person.IterateHour()
				}
			}(person)
		}

		wg.Wait()

		iterateTimer += time.Since(startTime)

		// Increment global time
		utils.GlobalTick.Increment()
	}

	fmt.Printf("Simulation completed. Total iteration time: %v\n", iterateTimer)

	// Print final state of all humans
	fmt.Println("\n========================================")
	fmt.Println("FINAL STATE OF ALL HUMANS")
	fmt.Println("========================================")
	for i, person := range people {
		person.PrintFinalInfo(i + 1)
	}

	// Print summary statistics
	fmt.Println("========================================")
	fmt.Println("SIMULATION SUMMARY")
	fmt.Println("========================================")

	aliveCount := 0
	completedTargetsCount := 0
	totalMoney := int64(0)
	employedCount := 0
	totalItems := 0

	for _, person := range people {
		if !person.Dead {
			aliveCount++
		}
		if person.Job != nil {
			employedCount++
		}
		completedTargetsCount += len(person.CompletedGlobalTargets)
		totalMoney += person.Money
		totalItems += len(person.Items)
	}

	fmt.Printf("Survival Rate: %d/%d humans alive (%.1f%%)\n",
		aliveCount, len(people), float64(aliveCount)/float64(len(people))*100)
	fmt.Printf("Employment Rate: %d/%d humans employed (%.1f%%)\n",
		employedCount, aliveCount, float64(employedCount)/float64(aliveCount)*100)
	fmt.Printf("Total Completed Global Targets: %d\n", completedTargetsCount)
	fmt.Printf("Average Completed Targets per Person: %.1f\n",
		float64(completedTargetsCount)/float64(len(people)))
	fmt.Printf("Total Money in Economy: %d rubles\n", totalMoney)
	fmt.Printf("Average Money per Person: %d rubles\n", totalMoney/int64(len(people)))
	fmt.Printf("Total Items Acquired: %d\n", totalItems)

	// Target completion statistics
	fmt.Println("\nTarget Completion Statistics:")
	targetStats := make(map[string]int)
	for _, person := range people {
		for target := range person.CompletedGlobalTargets {
			targetStats[target.Name]++
		}
	}

	for targetName, count := range targetStats {
		fmt.Printf("  %s: %d people (%.1f%%)\n",
			targetName, count, float64(count)/float64(len(people))*100)
	}

	// Suppress unused variable warnings
	_ = actionMap
	_ = localMap
	_ = globalMap
}
