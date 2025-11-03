package utils

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

// Tick represents global time counter
type Tick struct {
	tick uint64
	mu   sync.RWMutex
}

var GlobalTick = &Tick{tick: 0}

func (t *Tick) Get() uint64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.tick
}

func (t *Tick) Increment() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.tick++
}

// IsNatural checks if string represents a natural number
func IsNatural(s string) bool {
	if _, err := strconv.Atoi(s); err != nil {
		return false
	}
	return true
}

// Split splits string by delimiter and trims whitespace
func Split(s, delimiter string) []string {
	if delimiter == "" {
		return []string{s}
	}

	parts := strings.Split(s, delimiter)
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// Shuffle shuffles a slice of strings
func Shuffle(slice []string) []string {
	result := make([]string, len(slice))
	copy(result, slice)

	// Use the existing GlobalRandom instance instead of deprecated rand.Seed
	GlobalRandom.mu.Lock()
	defer GlobalRandom.mu.Unlock()
	GlobalRandom.rand.Shuffle(len(result), func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})

	return result
}

// Intersect returns intersection of two string sets
func Intersect(a, b map[string]bool) map[string]bool {
	result := make(map[string]bool)
	for key := range a {
		if b[key] {
			result[key] = true
		}
	}
	return result
}

// IntersectSlices returns intersection of two string slices as a map
func IntersectSlices(a, b []string) map[string]bool {
	setA := make(map[string]bool)
	setB := make(map[string]bool)

	for _, item := range a {
		setA[item] = true
	}
	for _, item := range b {
		setB[item] = true
	}

	return Intersect(setA, setB)
}

// Compare performs comparison operations
func Compare(a int64, op string, b int64) bool {
	switch op {
	case "=":
		return a == b
	case ">":
		return a > b
	case "<":
		return a < b
	case "<>":
		return a != b
	case ">=":
		return a >= b
	case "<=":
		return a <= b
	default:
		return false
	}
}

// LoadSequencesFromFile loads configuration from file
func LoadSequencesFromFile(filename string) ([][]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %v", filename, err)
	}
	defer file.Close()

	var result [][]string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		words := Split(line, " ")
		if len(words) > 0 {
			result = append(result, words)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file %s: %v", filename, err)
	}

	return result, nil
}

// Logger setup for debugging
var (
	DebugLogger *log.Logger
	InfoLogger  *log.Logger
)

const DEBUG_ENABLED = false // Set to true to enable debug logging

func init() {
	if DEBUG_ENABLED {
		// Create log files only if debug is enabled
		debugFile, err := os.OpenFile("logs_2.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalln("Failed to open debug log file:", err)
		}
		
		infoFile, err := os.OpenFile("logs_1.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalln("Failed to open info log file:", err)
		}
		
		DebugLogger = log.New(debugFile, "", log.LstdFlags)
		InfoLogger = log.New(infoFile, "", log.LstdFlags)
	} else {
		// Create no-op loggers that discard output
		DebugLogger = log.New(os.Stderr, "", 0)
		DebugLogger.SetOutput(log.Writer()) // This will discard output
		InfoLogger = log.New(os.Stderr, "", 0)
		InfoLogger.SetOutput(log.Writer()) // This will discard output
	}
}
