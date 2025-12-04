package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fallra1n/humanity/components"
	"github.com/fallra1n/humanity/utils"
)

// LoadActions загружает действия из конфигурационного файла
func LoadActions(filename string) ([]*components.Action, error) {
	sequences, err := utils.LoadSequencesFromFile(filename)
	if err != nil {
		return nil, err
	}

	var actions []*components.Action

	for _, words := range sequences {
		if len(words) < 4 {
			return nil, fmt.Errorf("invalid action format in %s", filename)
		}

		name := words[0]
		price, err := strconv.ParseInt(words[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid price for action %s: %v", name, err)
		}

		timeToExecute, err := strconv.ParseInt(words[2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid time for action %s: %v", name, err)
		}

		var bonusMoney int64 = 0
		var tags []string
		rules := make(map[string]int64)
		items := make(map[string]int64)
		removableItems := make(map[string]int64)

		for i := 3; i < len(words); i++ {
			word := words[i]

			if strings.HasPrefix(word, "$") {
				// Правило
				if len(word) > 1 && word[1] == '-' {
					// Удаляемый предмет
					key := word[2:]
					if count, exists := removableItems[key]; exists {
						removableItems[key] = count + 1
					} else {
						removableItems[key] = 1
					}
				} else {
					// Обычное правило
					key := word[1:]
					if count, exists := rules[key]; exists {
						rules[key] = count + 1
					} else {
						rules[key] = 1
					}
				}
			} else if strings.HasPrefix(word, "@") {
				// Предмет для добавления
				key := word[1:]
				if count, exists := items[key]; exists {
					items[key] = count + 1
				} else {
					items[key] = 1
				}
			} else if strings.HasPrefix(word, "+") {
				// Бонусные деньги
				bonus, err := strconv.ParseInt(word[1:], 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid bonus for action %s: %v", name, err)
				}
				bonusMoney = bonus
			} else {
				// Тег
				tags = append(tags, word)
			}
		}

		action := components.NewAction(name, price, timeToExecute, tags, rules, items, removableItems, bonusMoney)
		actions = append(actions, action)
	}

	return actions, nil
}

// LoadLocalTargets загружает локальные цели из конфигурационного файла
func LoadLocalTargets(filename string, allActions []*components.Action) ([]*components.LocalTarget, error) {
	sequences, err := utils.LoadSequencesFromFile(filename)
	if err != nil {
		return nil, err
	}

	var targets []*components.LocalTarget

	for _, words := range sequences {
		if len(words) < 2 {
			return nil, fmt.Errorf("invalid local target format in %s", filename)
		}

		name := words[0]
		tags := words[1:]

		target := components.NewLocalTarget(name, tags, allActions)
		targets = append(targets, target)
	}

	return targets, nil
}

// LoadGlobalTargets загружает глобальные цели из конфигурационного файла
func LoadGlobalTargets(filename string, allLocalTargets []*components.LocalTarget) ([]*components.GlobalTarget, error) {
	sequences, err := utils.LoadSequencesFromFile(filename)
	if err != nil {
		return nil, err
	}

	var targets []*components.GlobalTarget

	for _, words := range sequences {
		if len(words) < 2 {
			return nil, fmt.Errorf("invalid global target format in %s", filename)
		}

		name := words[0]
		var power float64 = 1.0
		var tags []string

		// Попытаться разобрать второе слово как силу (float)
		if len(words) > 2 {
			if parsedPower, err := strconv.ParseFloat(words[1], 64); err == nil {
				power = parsedPower
				tags = words[2:]
			} else {
				// Второе слово не число, рассматривать как тег
				tags = words[1:]
			}
		} else {
			tags = words[1:]
		}

		target := components.NewGlobalTarget(name, tags, power, allLocalTargets)
		targets = append(targets, target)
	}

	return targets, nil
}

// CreateNameMaps создает карты поиска для действий, локальных целей и глобальных целей
func CreateNameMaps(actions []*components.Action, localTargets []*components.LocalTarget, globalTargets []*components.GlobalTarget) (
	map[string]*components.Action, map[string]*components.LocalTarget, map[string]*components.GlobalTarget, error) {

	actionMap := make(map[string]*components.Action)
	localMap := make(map[string]*components.LocalTarget)
	globalMap := make(map[string]*components.GlobalTarget)

	// Проверить дубликаты и создать карту действий
	for _, action := range actions {
		if _, exists := actionMap[action.Name]; exists {
			return nil, nil, nil, fmt.Errorf("action name duplication: %s", action.Name)
		}
		actionMap[action.Name] = action
	}

	// Проверить дубликаты и создать карту локальных целей
	for _, target := range localTargets {
		if _, exists := localMap[target.Name]; exists {
			return nil, nil, nil, fmt.Errorf("local target name duplication: %s", target.Name)
		}
		localMap[target.Name] = target
	}

	// Проверить дубликаты и создать карту глобальных целей
	for _, target := range globalTargets {
		if _, exists := globalMap[target.Name]; exists {
			return nil, nil, nil, fmt.Errorf("global target name duplication: %s", target.Name)
		}
		globalMap[target.Name] = target
	}

	return actionMap, localMap, globalMap, nil
}
