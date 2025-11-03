package main

import (
	"fmt"
	"strconv"
	"strings"
)

// LoadActions loads actions from configuration file
func LoadActions(filename string) ([]*Action, error) {
	sequences, err := LoadSequencesFromFile(filename)
	if err != nil {
		return nil, err
	}

	var actions []*Action

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
				// Rule
				if len(word) > 1 && word[1] == '-' {
					// Removable item
					key := word[2:]
					if count, exists := removableItems[key]; exists {
						removableItems[key] = count + 1
					} else {
						removableItems[key] = 1
					}
				} else {
					// Regular rule
					key := word[1:]
					if count, exists := rules[key]; exists {
						rules[key] = count + 1
					} else {
						rules[key] = 1
					}
				}
			} else if strings.HasPrefix(word, "@") {
				// Item to add
				key := word[1:]
				if count, exists := items[key]; exists {
					items[key] = count + 1
				} else {
					items[key] = 1
				}
			} else if strings.HasPrefix(word, "+") {
				// Bonus money
				bonus, err := strconv.ParseInt(word[1:], 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid bonus for action %s: %v", name, err)
				}
				bonusMoney = bonus
			} else {
				// Tag
				tags = append(tags, word)
			}
		}

		action := NewAction(name, price, timeToExecute, tags, rules, items, removableItems, bonusMoney)
		actions = append(actions, action)
	}

	return actions, nil
}

// LoadLocalTargets loads local targets from configuration file
func LoadLocalTargets(filename string, allActions []*Action) ([]*LocalTarget, error) {
	sequences, err := LoadSequencesFromFile(filename)
	if err != nil {
		return nil, err
	}

	var targets []*LocalTarget

	for _, words := range sequences {
		if len(words) < 2 {
			return nil, fmt.Errorf("invalid local target format in %s", filename)
		}

		name := words[0]
		tags := words[1:]

		target := NewLocalTarget(name, tags, allActions)
		targets = append(targets, target)
	}

	return targets, nil
}

// LoadGlobalTargets loads global targets from configuration file
func LoadGlobalTargets(filename string, allLocalTargets []*LocalTarget) ([]*GlobalTarget, error) {
	sequences, err := LoadSequencesFromFile(filename)
	if err != nil {
		return nil, err
	}

	var targets []*GlobalTarget

	for _, words := range sequences {
		if len(words) < 2 {
			return nil, fmt.Errorf("invalid global target format in %s", filename)
		}

		name := words[0]
		var power float64 = 1.0
		var tags []string

		// Try to parse second word as power (float)
		if len(words) > 2 {
			if parsedPower, err := strconv.ParseFloat(words[1], 64); err == nil {
				power = parsedPower
				tags = words[2:]
			} else {
				// Second word is not a number, treat as tag
				tags = words[1:]
			}
		} else {
			tags = words[1:]
		}

		target := NewGlobalTarget(name, tags, power, allLocalTargets)
		targets = append(targets, target)
	}

	return targets, nil
}

// CreateNameMaps creates lookup maps for actions, local targets, and global targets
func CreateNameMaps(actions []*Action, localTargets []*LocalTarget, globalTargets []*GlobalTarget) (
	map[string]*Action, map[string]*LocalTarget, map[string]*GlobalTarget, error) {

	actionMap := make(map[string]*Action)
	localMap := make(map[string]*LocalTarget)
	globalMap := make(map[string]*GlobalTarget)

	// Check for duplicates and create action map
	for _, action := range actions {
		if _, exists := actionMap[action.Name]; exists {
			return nil, nil, nil, fmt.Errorf("action name duplication: %s", action.Name)
		}
		actionMap[action.Name] = action
	}

	// Check for duplicates and create local target map
	for _, target := range localTargets {
		if _, exists := localMap[target.Name]; exists {
			return nil, nil, nil, fmt.Errorf("local target name duplication: %s", target.Name)
		}
		localMap[target.Name] = target
	}

	// Check for duplicates and create global target map
	for _, target := range globalTargets {
		if _, exists := globalMap[target.Name]; exists {
			return nil, nil, nil, fmt.Errorf("global target name duplication: %s", target.Name)
		}
		globalMap[target.Name] = target
	}

	return actionMap, localMap, globalMap, nil
}
