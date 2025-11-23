package components

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fallra1n/humanity/utils"
)

// Action represents a concrete action that can be performed
type Action struct {
	Name           string
	Price          int64
	TimeToExecute  int64
	BonusMoney     int64
	Tags           map[string]bool
	Rules          map[string]int64
	Items          map[string]int64
	RemovableItems map[string]int64
}

// NewAction creates a new Action from configuration data
func NewAction(name string, price, timeToExecute int64, tags []string, rules, items, removableItems map[string]int64, bonusMoney int64) *Action {
	tagSet := make(map[string]bool)
	for _, tag := range tags {
		tagSet[tag] = true
	}

	return &Action{
		Name:           name,
		Price:          price,
		TimeToExecute:  timeToExecute,
		BonusMoney:     bonusMoney,
		Tags:           tagSet,
		Rules:          rules,
		Items:          items,
		RemovableItems: removableItems,
	}
}

// Executable checks if action can be executed by the human
func (a *Action) Executable(person *Human) bool {
	comparisons1 := []string{">", "<", "="}
	comparisons2 := []string{"<>", ">=", "<="}

	for rule, value := range a.Rules {
		comparison := ""

		// Find comparison operator
		for _, op := range comparisons2 {
			if strings.Contains(rule, op) {
				comparison = op
				break
			}
		}
		if comparison == "" {
			for _, op := range comparisons1 {
				if strings.Contains(rule, op) {
					comparison = op
					break
				}
			}
		}

		if comparison != "" {
			parts := strings.Split(rule, comparison)
			if len(parts) != 2 {
				continue
			}

			var values [2]int64
			for i, part := range parts {
				if utils.IsNatural(part) {
					val, _ := strconv.ParseInt(part, 10, 64)
					values[i] = val
				} else {
					// Handle metrics
					switch part {
					case "cash":
						values[i] = person.Money
						for family := range person.Family {
							values[i] += family.Money
						}
					case "job_time":
						values[i] = int64(person.JobTime)
					}
				}
			}

			if !utils.Compare(values[0], comparison, values[1]) {
				return false
			}
		} else {
			// Check item availability
			if person.Items[rule] < value {
				return false
			}
		}
	}

	// Check removable items availability
	for item, count := range a.RemovableItems {
		if person.Items[item] < count {
			return false
		}
	}

	return true
}

// Apply executes the action on the human
func (a *Action) Apply(person *Human) {
	person.Money -= a.Price

	// Add items
	for item, count := range a.Items {
		person.Items[item] += count
	}

	// Remove items
	for item, count := range a.RemovableItems {
		person.Items[item] -= count
		if person.Items[item] <= 0 {
			delete(person.Items, item)
		}
	}

	person.Money += a.BonusMoney

	// Special case: job finding
	if a.Name == "find_job" {
		person.findJob()
	}
	
	// Special case: meeting new people
	if a.Name == "meet_new_person" {
		person.meetNewPerson()
	}
}

// String method for Action
func (a *Action) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Action \"%s\":\n", a.Name))
	sb.WriteString(fmt.Sprintf("  Price: %d\n", a.Price))
	sb.WriteString(fmt.Sprintf("  Time to execute: %d\n", a.TimeToExecute))

	if len(a.Rules) > 0 {
		sb.WriteString("  Rules:\n")
		for rule, value := range a.Rules {
			sb.WriteString(fmt.Sprintf("    %s [%d]\n", rule, value))
		}
	}

	if len(a.Tags) > 0 {
		sb.WriteString("  Tags:\n")
		for tag := range a.Tags {
			sb.WriteString(fmt.Sprintf("    %s\n", tag))
		}
	}

	if len(a.RemovableItems) > 0 {
		sb.WriteString("  Removable items:\n")
		for item, count := range a.RemovableItems {
			sb.WriteString(fmt.Sprintf("    %s [%d]\n", item, count))
		}
	}

	if len(a.Items) > 0 {
		sb.WriteString("  Items:\n")
		for item, count := range a.Items {
			sb.WriteString(fmt.Sprintf("    %s [%d]\n", item, count))
		}
	}

	return sb.String()
}