package components

import (
	"fmt"
	"math"
	"strings"

	"github.com/fallra1n/humanity/utils"
)

// Human represents a person in the simulation
type Human struct {
	Age                    float64
	Dead                   bool
	BusyHours              uint64
	Money                  int64
	Job                    *Vacancy
	JobTime                uint64
	HomeLocation           *Location
	Parents                map[*Human]float64
	Family                 map[*Human]float64
	Children               map[*Human]float64
	Friends                map[*Human]float64
	Splashes               []*Splash
	GlobalTargets          map[*GlobalTarget]bool
	CompletedGlobalTargets map[*GlobalTarget]bool
	Items                  map[string]int64
}

// NewHuman creates a new human
func NewHuman(parents map[*Human]bool, homeLocation *Location, globalTargets []*GlobalTarget) *Human {
	// Generate age with normal distribution
	age := math.Max(20, math.Min(80, utils.GlobalRandom.NextNormal(25.0, 10.0)))

	human := &Human{
		Age:                    age,
		Dead:                   false,
		BusyHours:              0,
		Money:                  7000,
		Job:                    nil,
		JobTime:                720,
		HomeLocation:           homeLocation,
		Parents:                make(map[*Human]float64),
		Family:                 make(map[*Human]float64),
		Children:               make(map[*Human]float64),
		Friends:                make(map[*Human]float64),
		Splashes:               make([]*Splash, 0),
		GlobalTargets:          make(map[*GlobalTarget]bool),
		CompletedGlobalTargets: make(map[*GlobalTarget]bool),
		Items:                  make(map[string]int64),
	}

	// Set parents
	for parent := range parents {
		human.Parents[parent] = 0.0
	}

	// Assign random global targets
	numTargets := 2 + utils.GlobalRandom.NextInt(2) // 2-3 targets
	if numTargets > len(globalTargets) {
		numTargets = len(globalTargets)
	}

	selectedTargets := make(map[string]bool)
	for len(human.GlobalTargets) < numTargets {
		target := globalTargets[utils.GlobalRandom.NextInt(len(globalTargets))]
		if !selectedTargets[target.Name] {
			// Create a copy of the global target for this human
			newTarget := &GlobalTarget{
				Name:            target.Name,
				Tags:            make(map[string]bool),
				Power:           target.Power,
				TargetsPossible: make(map[*LocalTarget]bool),
				TargetsExecuted: make(map[*LocalTarget]bool),
			}

			// Copy tags
			for tag := range target.Tags {
				newTarget.Tags[tag] = true
			}

			// Copy possible targets
			for localTarget := range target.TargetsPossible {
				newTarget.TargetsPossible[localTarget] = true
			}

			human.GlobalTargets[newTarget] = true
			selectedTargets[target.Name] = true

			// DebugLogger.Printf("Human #%d got Global target: %s", GlobalHumanStorage.Get(human), target.Name)
		}
	}

	GlobalHumanStorage.Append(human)
	return human
}

// findJob attempts to find a new job
func (h *Human) findJob() {
	var possibleJobs []*Vacancy

	for job := range h.HomeLocation.Jobs {
		for vacancy, count := range job.VacantPlaces {
			if count > 0 && (h.Job == nil || vacancy.Payment > h.Job.Payment) {
				suitable := true
				for tag := range vacancy.RequiredTags {
					if h.Items[tag] <= 0 {
						suitable = false
						break
					}
				}
				if suitable {
					possibleJobs = append(possibleJobs, vacancy)
				}
			}
		}
	}

	if len(possibleJobs) > 0 {
		chosen := possibleJobs[utils.GlobalRandom.NextInt(len(possibleJobs))]

		// Quit old job
		if h.Job != nil {
			h.Job.Parent.VacantPlaces[h.Job]++
		}

		// Take new job
		h.Job = chosen
		chosen.Parent.VacantPlaces[chosen]--
		h.JobTime = 0
	}
}

// IterateHour processes one hour of the human's life
func (h *Human) IterateHour() {
	if h.Money <= 0 {
		splash := NewSplash("need_money", []string{"money", "well-being", "career"}, 24)
		h.Splashes = append(h.Splashes, splash)
		// DebugLogger.Printf("Hour %d: human #%d, splash: need_money", GlobalTick.Get(), GlobalHumanStorage.Get(h))
	}

	// Age relationships
	for parent := range h.Parents {
		h.Parents[parent] += 1.0 / (24 * 365)
	}
	for family := range h.Family {
		h.Family[family] += 1.0 / (24 * 365)
	}
	for child := range h.Children {
		h.Children[child] += 1.0 / (24 * 365)
	}
	for friend := range h.Friends {
		h.Friends[friend] += 1.0 / (24 * 365)
	}

	// Job time management
	if h.Job == nil {
		h.JobTime = 721
	} else {
		h.JobTime++
	}

	// Remove expired splashes
	validSplashes := make([]*Splash, 0)
	for _, splash := range h.Splashes {
		if !splash.IsExpired() {
			validSplashes = append(validSplashes, splash)
		}
	}
	h.Splashes = validSplashes

	// Handle death
	if h.Age > 80.0 {
		if !h.Dead {
			h.redistributeWealth()
			h.Money = 0
		}
		h.Dead = true
	}

	if h.Dead {
		return
	}

	// Age the human
	h.Age += 1.0 / (24 * 365)

	// Daily expenses
	if utils.GlobalTick.Get()%24 == 0 {
		h.Money -= 500
		// DebugLogger.Printf("Hour %d: human #%d spent 500 rub on base daily expenses. Money left: %d",
		//	GlobalTick.Get(), GlobalHumanStorage.Get(h), h.Money)

		// Monthly salary
		if h.Job != nil && utils.GlobalTick.Get()%(30*24) == 0 {
			h.Money += int64(h.Job.Payment)
			// DebugLogger.Printf("Hour %d: human #%d got payment. Money balance: %d",
			//	GlobalTick.Get(), GlobalHumanStorage.Get(h), h.Money)
		}
	}

	// Redistribute money within family if needed
	if h.Money < 0 {
		h.redistributeMoneyInFamily()
	}

	// Main activity logic
	if h.BusyHours > 0 {
		h.BusyHours--
	} else {
		h.performActions()
	}
}

// redistributeWealth distributes money to family when dying
func (h *Human) redistributeWealth() {
	if h.Money <= 0 {
		return
	}

	var candidates []*Human
	for family := range h.Family {
		if !family.Dead {
			candidates = append(candidates, family)
		}
	}
	for child := range h.Children {
		if !child.Dead {
			candidates = append(candidates, child)
		}
	}
	for parent := range h.Parents {
		if !parent.Dead {
			candidates = append(candidates, parent)
		}
	}

	if len(candidates) > 0 {
		share := h.Money / int64(len(candidates))
		for _, candidate := range candidates {
			candidate.Money += share
		}
	}
}

// redistributeMoneyInFamily tries to get money from family members
func (h *Human) redistributeMoneyInFamily() {
	for family := range h.Family {
		if h.Money < 0 && family.Money > 0 {
			transfer := int64(math.Min(float64(-h.Money), float64(family.Money)))
			h.Money += transfer
			family.Money -= transfer
		}
	}

	for child := range h.Children {
		if h.Money < 0 && child.Money > 0 {
			transfer := int64(math.Min(float64(-h.Money), float64(child.Money)))
			h.Money += transfer
			child.Money -= transfer
		}
	}

	for parent := range h.Parents {
		if h.Money < 0 && parent.Money > 0 {
			transfer := int64(math.Min(float64(-h.Money), float64(parent.Money)))
			h.Money += transfer
			parent.Money -= transfer
		}
	}
}

// performActions handles the main decision-making logic
func (h *Human) performActions() {
	if len(h.GlobalTargets) == 0 {
		return
	}

	rating := make(map[float64][]*GlobalTarget)

	if len(h.Splashes) > 0 {
		// Rate targets based on splashes
		for target := range h.GlobalTargets {
			var counter uint64 = 0
			for _, splash := range h.Splashes {
				if len(utils.IntersectSlices(getKeysFromMap(target.Tags), getKeysFromMap(splash.Tags))) > 0 {
					counter++
				}
			}
			rate := (target.Power * float64(counter)) / float64(len(h.Splashes))
			rating[rate] = append(rating[rate], target)
		}
	} else {
		// Rate targets based on executability and power
		for target := range h.GlobalTargets {
			var executable float64 = 0
			if target.Executable(h) {
				executable = 1
			}
			rate := executable * target.Power
			rating[rate] = append(rating[rate], target)
		}
	}

	if len(rating) == 0 {
		return
	}

	// Get highest rated targets
	var maxRate float64 = -1
	for rate := range rating {
		if rate > maxRate {
			maxRate = rate
		}
	}

	candidates := rating[maxRate]
	selectedGlobalTarget := candidates[utils.GlobalRandom.NextInt(len(candidates))]

	selectedLocalTarget := selectedGlobalTarget.ChooseTarget(h)
	if selectedLocalTarget != nil {
		selectedAction := selectedLocalTarget.ChooseAction(h)
		if selectedAction != nil {
			// DebugLogger.Printf("Human #%d chosen main global target \"%s\", local target \"%s\", action \"%s\"",
			//	GlobalHumanStorage.Get(h), selectedGlobalTarget.Name, selectedLocalTarget.Name, selectedAction.Name)

			selectedAction.Apply(h)
			selectedLocalTarget.MarkAsExecuted(selectedAction)

			if selectedLocalTarget.IsExecutedFull() {
				// DebugLogger.Printf("Local target \"%s\" reached", selectedLocalTarget.Name)
				selectedGlobalTarget.MarkAsExecuted(selectedLocalTarget)

				if selectedGlobalTarget.IsExecutedFull() {
					// DebugLogger.Printf("Global target \"%s\" reached", selectedGlobalTarget.Name)
					h.CompletedGlobalTargets[selectedGlobalTarget] = true
					delete(h.GlobalTargets, selectedGlobalTarget)
				}
			}
		}
	}
}

// PrintInitialInfo prints detailed information about human at start
func (h *Human) PrintInitialInfo(id int) {
	fmt.Printf("=== Human #%d Initial State ===\n", id)
	fmt.Printf("Age: %.1f years\n", h.Age)
	fmt.Printf("Money: %d rubles\n", h.Money)
	fmt.Printf("Job: %s\n", h.getJobStatus())

	fmt.Printf("Global Targets (%d):\n", len(h.GlobalTargets))
	for target := range h.GlobalTargets {
		fmt.Printf("  - %s (power: %.1f) [%s]\n",
			target.Name, target.Power, h.getTagsString(target.Tags))
	}

	if len(h.Items) > 0 {
		fmt.Printf("Items: %s\n", h.getItemsString())
	}

	fmt.Println()
}

// PrintFinalInfo prints detailed information about human at end
func (h *Human) PrintFinalInfo(id int) {
	fmt.Printf("=== Human #%d Final State ===\n", id)
	fmt.Printf("Status: %s\n", h.getLifeStatus())
	fmt.Printf("Age: %.1f years\n", h.Age)
	fmt.Printf("Money: %d rubles\n", h.Money)
	fmt.Printf("Job: %s\n", h.getJobStatus())

	fmt.Printf("Completed Global Targets (%d):\n", len(h.CompletedGlobalTargets))
	for target := range h.CompletedGlobalTargets {
		fmt.Printf("  ✓ %s (power: %.1f)\n", target.Name, target.Power)
	}

	fmt.Printf("Remaining Global Targets (%d):\n", len(h.GlobalTargets))
	for target := range h.GlobalTargets {
		progress := h.getTargetProgress(target)
		fmt.Printf("  - %s (power: %.1f) - Progress: %.1f%%\n",
			target.Name, target.Power, progress)
	}

	if len(h.Items) > 0 {
		fmt.Printf("Items: %s\n", h.getItemsString())
	}

	if len(h.Family) > 0 || len(h.Children) > 0 {
		fmt.Printf("Family: %d family members, %d children\n",
			len(h.Family), len(h.Children))
	}

	fmt.Println()
}

// Helper methods for formatting output
func (h *Human) getLifeStatus() string {
	if h.Dead {
		return "Dead"
	}
	return "Alive"
}

func (h *Human) getJobStatus() string {
	if h.Job == nil {
		return "Unemployed"
	}
	return fmt.Sprintf("Employed (salary: %d rubles/month, experience: %d hours)",
		h.Job.Payment, h.JobTime)
}

func (h *Human) getTagsString(tags map[string]bool) string {
	var tagList []string
	for tag := range tags {
		tagList = append(tagList, tag)
	}
	return strings.Join(tagList, ", ")
}

func (h *Human) getItemsString() string {
	var items []string
	for item, count := range h.Items {
		if count > 1 {
			items = append(items, fmt.Sprintf("%s x%d", item, count))
		} else {
			items = append(items, item)
		}
	}
	return strings.Join(items, ", ")
}

func (h *Human) getTargetProgress(target *GlobalTarget) float64 {
	totalTags := len(target.Tags)
	if totalTags == 0 {
		return 100.0
	}

	completedTags := 0
	for executedTarget := range target.TargetsExecuted {
		for tag := range executedTarget.Tags {
			if target.Tags[tag] {
				completedTags++
				break // Count each executed target only once
			}
		}
	}

	return float64(completedTags) / float64(totalTags) * 100.0
}

// Helper function to get keys from a map[string]bool
func getKeysFromMap(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}