package main

import (
	"fmt"
	"log"
	"sync"
	"time"
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

	// Create city
	city := &Location{
		Jobs:   make(map[*Job]bool),
		Humans: make(map[*Human]bool),
		Paths:  make(map[*Path]bool),
	}

	// Create job and vacancy
	company := &Job{
		VacantPlaces: make(map[*Vacancy]uint64),
		HomeLocation: city,
	}

	vacancy := &Vacancy{
		Parent:       company,
		RequiredTags: make(map[string]bool),
		Payment:      50000, // 50,000 rubles per month
	}

	company.VacantPlaces[vacancy] = 10
	city.Jobs[company] = true

	// Create 10 humans
	var people []*Human
	for i := 0; i < 1000; i++ {
		human := NewHuman(make(map[*Human]bool), city, globalTargets)
		human.Money = 10000 // Starting capital
		people = append(people, human)
		city.Humans[human] = true
	}

	// Simulation parameters
	const hoursPerMonth = 30 * 24       // 1 month in hours
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

			go func(person *Human) {
				defer wg.Done()
				if !person.Dead {
					person.IterateHour()
				}
			}(person)
		}

		wg.Wait()

		iterateTimer += time.Since(startTime)

		// Increment global time
		GlobalTick.Increment()
	}

	fmt.Printf("Simulation completed. Total iteration time: %v\n", iterateTimer)

	// Print final statistics
	fmt.Println("\nFinal Statistics:")
	aliveCount := 0
	completedTargetsCount := 0

	for _, person := range people {
		if !person.Dead {
			aliveCount++
		}
		completedTargetsCount += len(person.CompletedGlobalTargets)
	}

	fmt.Printf("\nSummary: %d/%d humans alive, %d total completed global targets\n",
		aliveCount, len(people), completedTargetsCount)

	// Suppress unused variable warnings
	_ = actionMap
	_ = localMap
	_ = globalMap
}
