package main

import (
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/fallra1n/humanity/components"
	"github.com/fallra1n/humanity/config"
	"github.com/fallra1n/humanity/utils"
)

func loadInitData() ([]*components.Action, []*components.LocalTarget, []*components.GlobalTarget) {
	// Загрузить действия
	actions, err := LoadActions("actions.ini")
	if err != nil {
		log.Fatalf("Failed to load actions: %v", err)
	}

	// Загрузить локальные цели
	localTargets, err := LoadLocalTargets("local.ini", actions)
	if err != nil {
		log.Fatalf("Failed to load local targets: %v", err)
	}

	// Загрузить глобальные цели
	globalTargets, err := LoadGlobalTargets("global.ini", localTargets)
	if err != nil {
		log.Fatalf("Failed to load global targets: %v", err)
	}

	return actions, localTargets, globalTargets
}

func main() {
	// Парсинг аргументов командной строки
	var showStats bool
	flag.BoolVar(&showStats, "stat", false, "Показать подробную статистику")
	flag.Parse()

	actions, localTargets, globalTargets := loadInitData()

	// Создать карты имен для поиска
	actionMap, localMap, globalMap, err := CreateNameMaps(actions, localTargets, globalTargets)
	if err != nil {
		log.Fatalf("Failed to create name maps: %v", err)
	}

	// Создать два города
	smallCity := components.CreateSmallCity("City 1")
	largeCity := components.CreateLargeCity("City 2")

	// Вывести информацию о городах (только если включен флаг --stat)
	if showStats {
		components.PrintCityInfo(smallCity)
		components.PrintCityInfo(largeCity)
	}

	// Создать популяцию для симуляции
	people, populationStats := CreatePopulation(smallCity, largeCity, globalTargets)

	// Вывести статистику популяции  (только если включен флаг --stat)
	if showStats {
		PrintPopulationStats(populationStats, smallCity, largeCity)
	}

	// Вывести начальное состояние всех людей (только если включен флаг --stat)
	if showStats {
		PrintInitialStatistics(people)
	}

	// Параметры симуляции из конфигурации
	var iterateTimer time.Duration

	// Основной цикл симуляции
	for hour := uint64(0); hour < config.TotalSimulationHours; hour++ {
		startTime := time.Now()

		wg := sync.WaitGroup{}

		// Обработать каждого человека
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

		// Обработать дружбу после того, как все люди действовали (однопоточно для безопасности)
		// Только в нерабочие часы сна
		if !utils.IsSleepTime(utils.GlobalTick.Get()) {
			components.ProcessFriendships(people)
			components.ProcessMarriages(people)
		}

		// Обработать роды (дети, рожденные в течение этого часа)
		newChildren := components.ProcessBirths(people, globalTargets)
		if len(newChildren) > 0 {
			people = append(people, newChildren...)
		}

		// Обработать потенциальные увольнения после того, как все люди действовали
		for _, person := range people {
			if !person.Dead {
				if canFire, reason := person.CanBeFired(); canFire {
					person.FireEmployee(reason)
				}
			}
		}

		// Записать текущее состояние в CSV
		if err := logToCSV(people, hour); err != nil {
			log.Printf("Warning: Failed to write to log.csv: %v", err)
		}

		iterateTimer += time.Since(startTime)

		// Увеличить глобальное время
		utils.GlobalTick.Increment()
	}

	fmt.Printf("Simulation completed. Total iteration time: %v\n", iterateTimer)

	// Вывести финальное состояние всех людей (только если включен флаг --stat)
	if showStats {
		PrintFinalStatistics(people)
		PrintSimulationSummary(people, smallCity, largeCity)
	} else {
		// Краткая статистика без флага --stat
		stats := CalculateStatistics(people, smallCity, largeCity)
		fmt.Printf("Simulation completed. Population: %d (%d alive, %d employed)\n",
			len(people), stats.AliveCount, stats.EmployedCount)
	}

	// Подавить предупреждения о неиспользуемых переменных
	_ = actionMap
	_ = localMap
	_ = globalMap
}
