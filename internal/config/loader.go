package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fallra1n/humanity/internal/action"
	globaltarget "github.com/fallra1n/humanity/internal/global_target"
	localtarget "github.com/fallra1n/humanity/internal/local_target"
	"github.com/fallra1n/humanity/internal/registry"
	"github.com/fallra1n/humanity/internal/shared"
)

const (
	actionsPath       = "data/actions.ini"
	localTargetsPath  = "data/local.ini"
	globalTargetsPath = "data/global.ini"
)

type Config struct {
	Actions       []action.BasicAction
	LocalTargets  []localtarget.BasicLocalTarget
	GlobalTargets []globaltarget.BasicGlobalTarget
}

type Loader struct {
	actionRegistry *registry.Registry
}

func NewLoader() *Loader {
	return &Loader{
		actionRegistry: registry.NewRegistry(),
	}
}

// Загружает конфигурацию из стандартных файлов
func (l *Loader) LoadFromFiles() (*Config, error) {
	config := &Config{}

	actions, err := l.LoadActions(actionsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load actions: %w", err)
	}
	config.Actions = actions

	for _, act := range actions {
		l.actionRegistry.Register(act.Id, &act)
	}

	localTargets, err := l.LoadLocalTargets(localTargetsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load local targets: %w", err)
	}
	config.LocalTargets = localTargets

	globalTargets, err := l.LoadGlobalTargets(globalTargetsPath, localTargets)
	if err != nil {
		return nil, fmt.Errorf("failed to load global targets: %w", err)
	}
	config.GlobalTargets = globalTargets

	return config, nil
}

// Загружает действия из файла
func (l *Loader) LoadActions(filename string) ([]action.BasicAction, error) {
	lines, err := l.loadLinesFromFile(filename)
	if err != nil {
		return nil, err
	}

	var actions []action.BasicAction
	for i, line := range lines {
		if len(line) < 4 {
			return nil, fmt.Errorf("invalid action format at line %d: need at least 4 fields", i+1)
		}

		name := line[0]
		price, err := strconv.ParseInt(line[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid price at line %d: %w", i+1, err)
		}

		duration, err := strconv.ParseInt(line[2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid duration at line %d: %w", i+1, err)
		}

		act := action.NewBasicAction(
			shared.ID(name),
			name,
			shared.Money(price*100),
			shared.Hours(duration),
		)

		// Парсим остальные поля
		tags := make(shared.Tags)
		rules := make(map[string]string)
		requiredItems := make(map[string]int64)
		consumedItems := make(map[string]int64)
		producedItems := make(map[string]int64)
		var bonusMoney shared.Money

		for j := 3; j < len(line); j++ {
			field := line[j]

			if strings.HasPrefix(field, "$") {
				if strings.HasPrefix(field, "$-") {
					key := field[2:]
					consumedItems[key] = 1
				} else {
					key := field[1:]
					rules[key] = "1"
				}
			} else if strings.HasPrefix(field, "@") {
				key := field[1:]
				requiredItems[key] = 1
			} else if strings.HasPrefix(field, "+") {
				bonus, err := strconv.ParseInt(field[1:], 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid bonus money at line %d: %w", i+1, err)
				}
				bonusMoney = shared.Money(bonus * 100)
			} else {
				tags.Add(field)
			}
		}

		act.RequiredTags = tags
		act.Rules = rules
		act.RequiredItems = requiredItems
		act.ConsumedItems = consumedItems
		act.ProducedItems = producedItems
		act.BonusMoney = bonusMoney

		actions = append(actions, *act)
	}

	return actions, nil
}

// Загружает строки из файла
func (l *Loader) loadLinesFromFile(filename string) ([][]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer file.Close()

	var lines [][]string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Пропускаем пустые строки и строки с комментариями
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Разбиваем строку на поля
		fields := strings.Fields(line)
		if len(fields) > 0 {
			lines = append(lines, fields)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", filename, err)
	}

	return lines, nil
}

// Загружает локальные цели из файла
func (l *Loader) LoadLocalTargets(filename string) ([]localtarget.BasicLocalTarget, error) {
	lines, err := l.loadLinesFromFile(filename)
	if err != nil {
		return nil, err
	}

	var localTargets []localtarget.BasicLocalTarget
	for i, line := range lines {
		if len(line) < 2 {
			return nil, fmt.Errorf("invalid local target format at line %d: need at least 2 fields", i+1)
		}

		name := line[0]
		tags := shared.NewTags(line[1:]...)

		lt := localtarget.NewBasicLocalTarget(
			shared.ID(name),
			name,
			tags,
		)

		var relevantActions []shared.Action
		for _, act := range l.actionRegistry.All() {
			if tags.Intersect(act.GetRequiredTags()).ToSlice() != nil {
				relevantActions = append(relevantActions, act)
			}
		}
		lt.AvailableActions = relevantActions

		localTargets = append(localTargets, *lt)
	}

	return localTargets, nil
}

// Загружает глобальные цели из файла
func (l *Loader) LoadGlobalTargets(filename string, localTargets []localtarget.BasicLocalTarget) ([]globaltarget.BasicGlobalTarget, error) {
	lines, err := l.loadLinesFromFile(filename)
	if err != nil {
		return nil, err
	}

	var globalTargets []globaltarget.BasicGlobalTarget
	for i, line := range lines {
		if len(line) < 2 {
			return nil, fmt.Errorf("invalid global target format at line %d: need at least 2 fields", i+1)
		}

		name := line[0]
		var power float64 = 1.0
		var tagStart int = 1

		if len(line) > 2 {
			if p, err := strconv.ParseFloat(line[1], 64); err == nil {
				power = p
				tagStart = 2
			}
		}

		tags := shared.NewTags(line[tagStart:]...)

		gt := globaltarget.NewBasicGlobalTarget(
			shared.ID(name),
			name,
			power,
			tags,
		)

		var relevantLocalTargets []shared.LocalTarget
		for _, lt := range localTargets {
			if tags.Intersect(lt.GetRequiredTags()).ToSlice() != nil {
				relevantLocalTargets = append(relevantLocalTargets, &lt)
			}
		}
		gt.AvailableLocalTargets = relevantLocalTargets

		globalTargets = append(globalTargets, *gt)
	}

	return globalTargets, nil
}
