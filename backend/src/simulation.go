package src

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/fallra1n/humanity/src/components"
	"github.com/fallra1n/humanity/src/config"
	"github.com/fallra1n/humanity/src/utils"
)

// Simulation represents the main simulation structure
type Simulation struct {
	AgentCount int    // количество агентов
	Duration   uint64 // длительность симуляции в часах
	ShowStats  bool   // показывать ли подробную статистику

	// Internal fields
	actions       []*components.Action
	localTargets  []*components.LocalTarget
	globalTargets []*components.GlobalTarget
	people        []*components.Human
	smallCity     *components.Location
	largeCity     *components.Location
}

// NewSimulation creates a new simulation instance
func NewSimulation(agentCount int, duration uint64, showStats bool) *Simulation {
	return &Simulation{
		AgentCount: agentCount,
		Duration:   duration,
		ShowStats:  showStats,
	}
}

// NewDefaultSimulation creates a simulation with default parameters from config
func NewDefaultSimulation(showStats bool) *Simulation {
	return &Simulation{
		AgentCount: config.TotalPopulation,
		Duration:   config.TotalSimulationHours,
		ShowStats:  showStats,
	}
}

// loadInitData loads actions, local targets, and global targets from configuration files
func (s *Simulation) loadInitData() error {
	// Загрузить действия
	actions, err := LoadActions("actions.ini")
	if err != nil {
		return fmt.Errorf("failed to load actions: %v", err)
	}
	s.actions = actions

	// Загрузить локальные цели
	localTargets, err := LoadLocalTargets("local.ini", actions)
	if err != nil {
		return fmt.Errorf("failed to load local targets: %v", err)
	}
	s.localTargets = localTargets

	// Загрузить глобальные цели
	globalTargets, err := LoadGlobalTargets("global.ini", localTargets)
	if err != nil {
		return fmt.Errorf("failed to load global targets: %v", err)
	}
	s.globalTargets = globalTargets

	return nil
}

// initializeCities creates and initializes the cities for the simulation
func (s *Simulation) initializeCities() {
	// Создать два города
	s.smallCity = components.CreateSmallCity("City 1")
	s.largeCity = components.CreateLargeCity("City 2")

	// Вывести информацию о городах (только если включен флаг --stat)
	if s.ShowStats {
		components.PrintCityInfo(s.smallCity)
		components.PrintCityInfo(s.largeCity)
	}
}

// initializePopulation creates the initial population for the simulation
func (s *Simulation) initializePopulation() {
	// Создать популяцию для симуляции
	people, populationStats := CreatePopulation(s.smallCity, s.largeCity, s.globalTargets)
	s.people = people

	// Вывести статистику популяции (только если включен флаг --stat)
	if s.ShowStats {
		PrintPopulationStats(populationStats, s.smallCity, s.largeCity)
		PrintInitialStatistics(s.people)
	}
}

// runSimulationLoop executes the main simulation loop
func (s *Simulation) runSimulationLoop() error {
	var iterateTimer time.Duration

	// Основной цикл симуляции
	for hour := uint64(0); hour < s.Duration; hour++ {
		startTime := time.Now()

		wg := sync.WaitGroup{}

		// Обработать каждого человека
		for _, person := range s.people {
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
			components.ProcessFriendships(s.people)
			components.ProcessMarriages(s.people)
		}

		// Обработать роды (дети, рожденные в течение этого часа)
		newChildren := components.ProcessBirths(s.people, s.globalTargets)
		if len(newChildren) > 0 {
			s.people = append(s.people, newChildren...)
		}

		// Обработать потенциальные увольнения после того, как все люди действовали
		for _, person := range s.people {
			if !person.Dead {
				if canFire, reason := person.CanBeFired(); canFire {
					person.FireEmployee(reason)
				}
			}
		}

		// Записать текущее состояние в CSV
		if err := logToCSV(s.people, hour); err != nil {
			log.Printf("Warning: Failed to write to log.csv: %v", err)
		}

		iterateTimer += time.Since(startTime)

		// Увеличить глобальное время
		utils.GlobalTick.Increment()
	}

	fmt.Printf("Simulation completed. Total iteration time: %v\n", iterateTimer)
	return nil
}

// printResults outputs the final simulation results
func (s *Simulation) printResults() {
	if s.ShowStats {
		PrintFinalStatistics(s.people)
		PrintSimulationSummary(s.people, s.smallCity, s.largeCity)
	} else {
		// Краткая статистика без флага --stat
		stats := CalculateStatistics(s.people, s.smallCity, s.largeCity)
		fmt.Printf("Simulation completed. Population: %d (%d alive, %d employed)\n",
			len(s.people), stats.AliveCount, stats.EmployedCount)
	}
}

// Run executes the complete simulation
func (s *Simulation) Run() error {
	// Загрузить начальные данные
	if err := s.loadInitData(); err != nil {
		return err
	}

	// Создать карты имен для поиска (пока не используются, но могут понадобиться)
	actionMap, localMap, globalMap, err := CreateNameMaps(s.actions, s.localTargets, s.globalTargets)
	if err != nil {
		return fmt.Errorf("failed to create name maps: %v", err)
	}

	// Инициализировать города
	s.initializeCities()

	// Инициализировать популяцию
	s.initializePopulation()

	// Запустить основной цикл симуляции
	if err := s.runSimulationLoop(); err != nil {
		return err
	}

	// Вывести результаты
	s.printResults()

	// Подавить предупреждения о неиспользуемых переменных
	_ = actionMap
	_ = localMap
	_ = globalMap

	return nil
}
