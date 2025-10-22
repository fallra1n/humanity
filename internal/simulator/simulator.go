package simulator

import (
	"sync"
	"time"

	"github.com/fallra1n/humanity/internal/agent"
	"github.com/fallra1n/humanity/internal/config"
	"github.com/fallra1n/humanity/internal/economy"
	"github.com/fallra1n/humanity/internal/shared"
)

// ...
type EmploymentStats struct {
	TotalVacancies     int
	AvailableVacancies int
	FilledVacancies    int
	UnemploymentRate   float64
}

type SimulationStats struct {
	Time         shared.SimulationTime
	TotalAgents  int
	AliveAgents  int
	DeadAgents   int
	AverageAge   float64
	AverageMoney float64
	TotalMoney   shared.Money
	Employment   EmploymentStats
	Timestamp    time.Time
}

// ...
type EventBusImpl struct {
	mu          sync.RWMutex
	subscribers map[string][]func(shared.Event)
}

func NewEventBus() *EventBusImpl {
	return &EventBusImpl{
		subscribers: make(map[string][]func(shared.Event)),
	}
}

// ...
type BasicSimulator struct {
	mu sync.RWMutex

	Time      shared.SimulationTime
	Agents    []*agent.Agent
	Locations []shared.Location
	Economy   *economy.EconomicSystem
	EventBus  shared.EventBus
	Config    *config.Config

	Running  bool
	StopChan chan struct{}
	DoneChan chan struct{}

	Stats     SimulationStats
	StatsChan chan SimulationStats

	SpeedMultiplier time.Duration
	MaxAgents       int
}

func NewBasicSimulator(cfg *config.Config) *BasicSimulator {
	return &BasicSimulator{
		Agents:          make([]*agent.Agent, 0),
		Locations:       make([]shared.Location, 0),
		Economy:         economy.NewEconomicSystem(),
		EventBus:        NewEventBus(),
		Config:          cfg,
		Running:         false,
		StopChan:        make(chan struct{}),
		DoneChan:        make(chan struct{}),
		StatsChan:       make(chan SimulationStats, 100),
		SpeedMultiplier: time.Second,
		MaxAgents:       1000,
	}
}
