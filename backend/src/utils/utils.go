package utils

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Tick представляет глобальный счетчик времени
type Tick struct {
	tick uint64
}

var GlobalTick = &Tick{tick: 0}

func (t *Tick) Get() uint64 {
	return t.tick
}

func (t *Tick) Increment() {
	t.tick++
}

// IsNatural проверяет, представляет ли строка натуральное число
func IsNatural(s string) bool {
	if _, err := strconv.Atoi(s); err != nil {
		return false
	}
	return true
}

// Split разделяет строку по разделителю и обрезает пробелы
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

// Shuffle перемешивает срез строк
func Shuffle(slice []string) []string {
	result := make([]string, len(slice))
	copy(result, slice)

	// Использовать существующий экземпляр GlobalRandom вместо устаревшего rand.Seed
	GlobalRandom.mu.Lock()
	defer GlobalRandom.mu.Unlock()
	GlobalRandom.rand.Shuffle(len(result), func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})

	return result
}

// Intersect возвращает пересечение двух строковых множеств
func Intersect(a, b map[string]bool) map[string]bool {
	result := make(map[string]bool)
	for key := range a {
		if b[key] {
			result[key] = true
		}
	}
	return result
}

// IntersectSlices возвращает пересечение двух строковых срезов как карту
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

// Compare выполняет операции сравнения
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

// LoadSequencesFromFile загружает конфигурацию из файла
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

		// Пропустить пустые строки и комментарии
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

// IsSleepTime проверяет, находится ли текущий час в пределах времени сна (23:00 до 07:00)
func IsSleepTime(currentHour uint64) bool {
	hourOfDay := currentHour % 24
	// Сон с 23:00 до 07:00
	return hourOfDay >= 23 || hourOfDay < 7
}

// GetHourOfDay возвращает час дня (0-23) из глобального времени
func GetHourOfDay(globalHour uint64) uint64 {
	return globalHour % 24
}

// IsWorkTime проверяет, находится ли текущий час в пределах рабочего времени (09:00 до 18:00)
func IsWorkTime(currentHour uint64) bool {
	hourOfDay := currentHour % 24
	// Работа с 09:00 до 18:00
	return hourOfDay >= 9 && hourOfDay < 18
}

// IsWorkDay проверяет, является ли это рабочим днем (понедельник-пятница)
func IsWorkDay(currentHour uint64) bool {
	// Предполагая, что симуляция начинается в понедельник (день 0)
	dayOfWeek := (currentHour / 24) % 7
	// Понедельник=0, Вторник=1, ..., Пятница=4, Суббота=5, Воскресенье=6
	return dayOfWeek < 5
}
