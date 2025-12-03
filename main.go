package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/fallra1n/humanity/components"
	"github.com/fallra1n/humanity/config"
	"github.com/fallra1n/humanity/utils"
)

// processFriendships handles friendship formation between people in the same building
func processFriendships(people []*components.Human) {
	// Group people by their current building
	buildingGroups := make(map[*components.Building][]*components.Human)

	for _, person := range people {
		if person.Dead || person.CurrentBuilding == nil {
			continue
		}
		buildingGroups[person.CurrentBuilding] = append(buildingGroups[person.CurrentBuilding], person)
	}

	// Process friendships within each building
	for _, group := range buildingGroups {
		if len(group) < 2 {
			continue // Need at least 2 people to form friendships
		}

		// Check all pairs of people in the building
		for i := 0; i < len(group); i++ {
			for j := i + 1; j < len(group); j++ {
				person1 := group[i]
				person2 := group[j]

				// Skip if already friends
				if _, alreadyFriends := person1.Friends[person2]; alreadyFriends {
					continue
				}

				// 25% chance to become friends
				if utils.GlobalRandom.NextFloat() < 0.25 {
					// Make bidirectional friendship
					person1.Friends[person2] = 0.0 // Start with 0 relationship strength
					person2.Friends[person1] = 0.0
				}
			}
		}
	}
}

// processMarriages handles marriage formation between compatible people
func processMarriages(people []*components.Human) {
	// Group single people by their current building
	buildingGroups := make(map[*components.Building][]*components.Human)

	for _, person := range people {
		if person.Dead || person.CurrentBuilding == nil || person.MaritalStatus != components.Single {
			continue
		}
		buildingGroups[person.CurrentBuilding] = append(buildingGroups[person.CurrentBuilding], person)
	}

	// Process potential marriages within each building
	for _, group := range buildingGroups {
		if len(group) < 2 {
			continue // Need at least 2 people to form marriages
		}

		// Check all pairs of people in the building
		for i := 0; i < len(group); i++ {
			for j := i + 1; j < len(group); j++ {
				person1 := group[i]
				person2 := group[j]

				// Check compatibility
				if person1.IsCompatibleWith(person2) {
					// Small chance to get married (5% per hour when compatible)
					if utils.GlobalRandom.NextFloat() < 0.05 {
						person1.MarryWith(person2)
					}
				}
			}
		}
	}
}

// processBirths handles pregnancy progression and births
func processBirths(people []*components.Human, globalTargets []*components.GlobalTarget) []*components.Human {
	var newChildren []*components.Human

	for _, person := range people {
		if person.Gender == components.Female && person.IsPregnant {
			newChild := person.ProcessPregnancy(people, globalTargets)
			if newChild != nil {
				// Child was born!
				newChildren = append(newChildren, newChild)

				// Add child to the city
				person.HomeLocation.Humans[newChild] = true
			}
		}
	}

	return newChildren
}

// logToCSV writes the current state of all people to a CSV file
func logToCSV(people []*components.Human, hour uint64) error {
	filename := "log.csv"
	
	// Check if file exists to determine if we need to write headers
	fileExists := false
	if _, err := os.Stat(filename); err == nil {
		fileExists = true
	}
	
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	
	writer := csv.NewWriter(file)
	defer writer.Flush()
	
	// Write header if this is the first time
	if !fileExists {
		header := []string{"hour", "agent_id", "age", "gender", "alive", "money", "location", "building_type", "job_status", "marital_status"}
		if err := writer.Write(header); err != nil {
			return err
		}
	}
	
	// Write data for each person
	for _, person := range people {
		location := "unknown"
		buildingType := "unknown"
		
		if person.CurrentBuilding != nil {
			location = person.CurrentBuilding.Name
			buildingType = string(person.CurrentBuilding.Type)
		}
		
		jobStatus := "unemployed"
		if person.Job != nil {
			jobStatus = "employed"
		}
		
		row := []string{
			strconv.FormatUint(hour, 10),
			strconv.Itoa(components.GlobalHumanStorage.Get(person)),
			fmt.Sprintf("%.2f", person.Age),
			string(person.Gender),
			fmt.Sprintf("%t", !person.Dead),
			strconv.FormatInt(person.Money, 10),
			location,
			buildingType,
			jobStatus,
			string(person.MaritalStatus),
		}
		
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	
	return nil
}

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

	// Create two cities
	smallCity := components.CreateSmallCity("Greenville")
	largeCity := components.CreateLargeCity("Metropolis")

	// Print city information
	components.PrintCityInfo(smallCity)
	components.PrintCityInfo(largeCity)

	// Create people for each city separately to ensure employment rate in each
	var people []*components.Human

	// Small city population and employment
	smallCityEmployed := int(float64(config.SmallCityPopulation) * config.EmploymentRate)
	smallCityResidential := components.GetResidentialBuildings(smallCity)
	for i := 0; i < config.SmallCityPopulation; i++ {
		human := components.NewHuman(make(map[*components.Human]bool), smallCity, globalTargets)
		human.Money = config.StartingMoney

		// Assign residential building
		assigned := false
		for _, building := range smallCityResidential {
			if building.AddResident(human) {
				assigned = true
				break
			}
		}
		if !assigned {
			fmt.Printf("Warning: Could not assign residential building to small city human %d\n", i+1)
		}

		// Employment based on config rate
		if i < smallCityEmployed {
			var availableVacancies []*components.Vacancy

			// Look for jobs in workplace buildings
			for building := range smallCity.Buildings {
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
				human.WorkBuilding = chosenVacancy.Parent.Building // Set work building
				chosenVacancy.Parent.VacantPlaces[chosenVacancy]--
			}
		}

		people = append(people, human)
		smallCity.Humans[human] = true
	}

	// Large city population and employment
	largeCityEmployed := int(float64(config.LargeCityPopulation) * config.EmploymentRate)
	largeCityResidential := components.GetResidentialBuildings(largeCity)
	for i := 0; i < config.LargeCityPopulation; i++ {
		human := components.NewHuman(make(map[*components.Human]bool), largeCity, globalTargets)
		human.Money = config.StartingMoney

		// Assign residential building
		assigned := false
		for _, building := range largeCityResidential {
			if building.AddResident(human) {
				assigned = true
				break
			}
		}
		if !assigned {
			fmt.Printf("Warning: Could not assign residential building to large city human %d\n", i+1)
		}

		// Employment based on config rate
		if i < largeCityEmployed {
			var availableVacancies []*components.Vacancy

			// Look for jobs in workplace buildings
			for building := range largeCity.Buildings {
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
				human.WorkBuilding = chosenVacancy.Parent.Building // Set work building
				chosenVacancy.Parent.VacantPlaces[chosenVacancy]--
			}
		}

		people = append(people, human)
		largeCity.Humans[human] = true
	}

	// Count actual employment statistics
	actualSmallCityEmployed := 0
	actualLargeCityEmployed := 0
	totalEmployed := 0

	for _, person := range people {
		if person.Job != nil {
			totalEmployed++
			if person.HomeLocation == smallCity {
				actualSmallCityEmployed++
			} else {
				actualLargeCityEmployed++
			}
		}
	}

	fmt.Printf("\nCreated %d people total:\n", len(people))
	fmt.Printf("  %s: %d residents, %d employed (%.1f%%)\n",
		smallCity.Name, config.SmallCityPopulation, actualSmallCityEmployed,
		float64(actualSmallCityEmployed)/float64(config.SmallCityPopulation)*100)
	fmt.Printf("  %s: %d residents, %d employed (%.1f%%)\n",
		largeCity.Name, config.LargeCityPopulation, actualLargeCityEmployed,
		float64(actualLargeCityEmployed)/float64(config.LargeCityPopulation)*100)
	fmt.Printf("Total employment: %d employed, %d unemployed (%.1f%% employment rate)\n",
		totalEmployed, len(people)-totalEmployed, float64(totalEmployed)/float64(len(people))*100)

	// Print initial state of all humans
	fmt.Println("========================================")
	fmt.Println("INITIAL STATE OF ALL HUMANS")
	fmt.Println("========================================")
	for i, person := range people {
		person.PrintInitialInfo(i + 1)
	}

	// Simulation parameters from config
	var iterateTimer time.Duration

	// Main simulation loop
	for hour := uint64(0); hour < config.TotalSimulationHours; hour++ {
		startTime := time.Now()

		wg := sync.WaitGroup{}

		// Process each person
		for _, person := range people {
			wg.Add(1)

			go func(person *components.Human) {
				defer wg.Done()
				if !person.Dead {
					person.IterateHour()
				}
			}(person)
		}

		wg.Wait()

		// Process friendships after all humans have acted (single-threaded for safety)
		// Only during non-sleep hours
		if !utils.IsSleepTime(utils.GlobalTick.Get()) {
			processFriendships(people)
			processMarriages(people)
		}

		// Process births (children born during this hour)
		newChildren := processBirths(people, globalTargets)
		if len(newChildren) > 0 {
			people = append(people, newChildren...)
		}

		// Process potential layoffs after all humans have acted
		for _, person := range people {
			if !person.Dead {
				if canFire, reason := person.CanBeFired(); canFire {
					person.FireEmployee(reason)
				}
			}
		}
		
		// Log current state to CSV
		if err := logToCSV(people, hour); err != nil {
			log.Printf("Warning: Failed to write to log.csv: %v", err)
		}
		
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
	maleCount := 0
	femaleCount := 0
	marriedCount := 0
	childrenCount := 0
	pregnantCount := 0
	totalChildren := 0
	moveCount := 0
	apartmentsForSaleSmallCity := 0
	apartmentsForSaleLargeCity := 0
	peopleWithoutHousing := 0

	for _, person := range people {
		if !person.Dead {
			aliveCount++
		}
		if person.Job != nil {
			employedCount++
		}
		if person.Gender == components.Male {
			maleCount++
		} else {
			femaleCount++
		}
		if person.MaritalStatus == components.Married {
			marriedCount++
		}
		if person.Age < 18.0 {
			childrenCount++
		}
		if person.IsPregnant {
			pregnantCount++
		}
		totalChildren += len(person.Children)
		completedTargetsCount += len(person.CompletedGlobalTargets)
		totalMoney += person.Money
		totalItems += len(person.Items)
		
		// Count people without housing
		if person.ResidentialBuilding == nil {
			peopleWithoutHousing++
		}

		// Count moves due to marriage
		if person.MaritalStatus == components.Married && person.Gender == components.Female {
			// Check if the woman moved to her husband's building
			if person.Spouse != nil && person.ResidentialBuilding != nil &&
				person.Spouse.ResidentialBuilding != nil &&
				person.ResidentialBuilding == person.Spouse.ResidentialBuilding {
				moveCount++
			}
		}
	}

	// Count apartments for sale in each city
	for building := range smallCity.Buildings {
		if building.Type == components.ResidentialHouse {
			apartmentsForSaleSmallCity += len(building.ApartmentsForSale)
		}
	}
	for building := range largeCity.Buildings {
		if building.Type == components.ResidentialHouse {
			apartmentsForSaleLargeCity += len(building.ApartmentsForSale)
		}
	}

	fmt.Printf("Population: %d humans (%d male, %d female)\n",
		len(people), maleCount, femaleCount)
	fmt.Printf("Survival Rate: %d/%d humans alive (%.1f%%)\n",
		aliveCount, len(people), float64(aliveCount)/float64(len(people))*100)
	fmt.Printf("Employment Rate: %d/%d humans employed (%.1f%%)\n",
		employedCount, aliveCount, float64(employedCount)/float64(aliveCount)*100)
	fmt.Printf("Marriage Rate: %d/%d humans married (%.1f%%)\n",
		marriedCount, aliveCount, float64(marriedCount)/float64(aliveCount)*100)
	fmt.Printf("Children: %d children under 18 (%.1f%% of population)\n",
		childrenCount, float64(childrenCount)/float64(len(people))*100)
	fmt.Printf("Pregnancies: %d women currently pregnant\n", pregnantCount)
	fmt.Printf("Total Children Born: %d children (average %.1f per adult)\n",
		totalChildren/2, float64(totalChildren)/float64(len(people)-childrenCount)) // Divide by 2 since both parents count the same child
	fmt.Printf("Marriage Moves: %d women moved to husband's building\n", moveCount)
	fmt.Printf("Apartments for Sale: %d in %s, %d in %s (total: %d)\n",
		apartmentsForSaleSmallCity, smallCity.Name, apartmentsForSaleLargeCity, largeCity.Name,
		apartmentsForSaleSmallCity+apartmentsForSaleLargeCity)
	fmt.Printf("People without housing: %d/%d (%.1f%%)\n",
		peopleWithoutHousing, len(people), float64(peopleWithoutHousing)/float64(len(people))*100)
	fmt.Printf("Total Completed Global Targets: %d\n", completedTargetsCount)
	fmt.Printf("Average Completed Targets per Person: %.1f\n",
		float64(completedTargetsCount)/float64(len(people)))
	fmt.Printf("Total Money in Economy: %d rubles\n", totalMoney)
	fmt.Printf("Average Money per Person: %d rubles\n", totalMoney/int64(len(people)))
	fmt.Printf("Total Items Acquired: %d\n", totalItems)

	// New functionality statistics
	totalFriends := 0
	peopleWithFriends := 0
	peopleAtWork := 0
	peopleAtHome := 0

	for _, person := range people {
		if !person.Dead {
			totalFriends += len(person.Friends)
			if len(person.Friends) > 0 {
				peopleWithFriends++
			}

			// Count current locations
			if person.CurrentBuilding != nil {
				if person.CurrentBuilding == person.WorkBuilding {
					peopleAtWork++
				} else if person.CurrentBuilding == person.ResidentialBuilding {
					peopleAtHome++
				}
			}
		}
	}

	fmt.Printf("\n=== NEW FEATURES STATISTICS ===\n")
	fmt.Printf("Social Connections:\n")
	fmt.Printf("  Total Friendships: %d\n", totalFriends)
	fmt.Printf("  People with Friends: %d/%d (%.1f%%)\n",
		peopleWithFriends, aliveCount, float64(peopleWithFriends)/float64(aliveCount)*100)
	fmt.Printf("  Average Friends per Person: %.1f\n",
		float64(totalFriends)/float64(aliveCount))

	fmt.Printf("Location Distribution:\n")
	fmt.Printf("  People at Work: %d (%.1f%%)\n",
		peopleAtWork, float64(peopleAtWork)/float64(aliveCount)*100)
	fmt.Printf("  People at Home: %d (%.1f%%)\n",
		peopleAtHome, float64(peopleAtHome)/float64(aliveCount)*100)
	fmt.Printf("  Other Locations: %d (%.1f%%)\n",
		aliveCount-peopleAtWork-peopleAtHome,
		float64(aliveCount-peopleAtWork-peopleAtHome)/float64(aliveCount)*100)

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
