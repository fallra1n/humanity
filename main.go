package main

import (
	"flag"
	"log"

	"github.com/fallra1n/humanity/src"
)

func main() {
	// Парсинг аргументов командной строки
	var showStats bool
	flag.BoolVar(&showStats, "stat", false, "Показать подробную статистику")
	flag.Parse()

	// Создать и запустить симуляцию
	simulation := src.NewDefaultSimulation(showStats)
	if err := simulation.Run(); err != nil {
		log.Fatalf("Simulation failed: %v", err)
	}
}
