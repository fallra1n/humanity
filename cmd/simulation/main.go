package main

import (
	"log"

	"github.com/fallra1n/humanity/internal/config"
)

func main() {
	log.Println("Starting Humanity Simulation...")

	loader := config.NewLoader()
	cfg, err := loader.LoadFromFiles()
	if err != nil {
		log.Printf("Warning: Could not load config files: %v", err)
		log.Println("Using default configuration...")
		return
	}

	log.Printf("Loaded configuration: %d actions, %d local targets, %d global targets",
		len(cfg.Actions), len(cfg.LocalTargets), len(cfg.GlobalTargets))

}
