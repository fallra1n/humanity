package types

import (
	"fmt"
	"math"
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
}

// LocalTarget represents a short-term goal
type LocalTarget struct {
	Name            string
	Tags            map[string]bool
	ActionsPossible map[*Action]bool
	ActionsExecuted map[*Action]bool
}

// NewLocalTarget creates a new LocalTarget
func NewLocalTarget(name string, tags []string, allActions []*Action) *LocalTarget {
	tagSet := make(map[string]bool)
	for _, tag := range tags {
		tagSet[tag] = true
	}

	actionsPossible := make(map[*Action]bool)
	for _, action := range allActions {
		if len(utils.IntersectSlices(tags, getKeysFromMap(action.Tags))) > 0 {
			actionsPossible[action] = true
		}
	}

	return &LocalTarget{
		Name:            name,
		Tags:            tagSet,
		ActionsPossible: actionsPossible,
		ActionsExecuted: make(map[*Action]bool),
	}
}

// MarkAsExecuted marks an action as executed
func (lt *LocalTarget) MarkAsExecuted(action *Action) {
	delete(lt.ActionsPossible, action)
	lt.ActionsExecuted[action] = true
}

// IsExecutedFull checks if all tags are covered by executed actions
func (lt *LocalTarget) IsExecutedFull() bool {
	remainingTags := make(map[string]bool)
	for tag := range lt.Tags {
		remainingTags[tag] = true
	}

	for action := range lt.ActionsExecuted {
		for tag := range action.Tags {
			delete(remainingTags, tag)
		}
	}

	return len(remainingTags) == 0
}

// Executable checks if the local target can be executed
func (lt *LocalTarget) Executable(person *Human) bool {
	unclosedTags := make(map[string]bool)
	for tag := range lt.Tags {
		unclosedTags[tag] = true
	}

	// Remove executed tags
	for action := range lt.ActionsExecuted {
		for tag := range action.Tags {
			delete(unclosedTags, tag)
		}
	}

	// Check if remaining tags can be closed
	for action := range lt.ActionsPossible {
		if action.Executable(person) {
			for tag := range action.Tags {
				delete(unclosedTags, tag)
			}
		}
	}

	return len(unclosedTags) == 0
}

// ChooseAction selects the best action for this target
func (lt *LocalTarget) ChooseAction(person *Human) *Action {
	leftTags := make(map[string]bool)
	for tag := range lt.Tags {
		leftTags[tag] = true
	}

	for action := range lt.ActionsExecuted {
		for tag := range action.Tags {
			delete(leftTags, tag)
		}
	}

	rating := make(map[uint64][]*Action)
	for action := range lt.ActionsPossible {
		if action.Executable(person) {
			var rate uint64 = 0
			for tag := range action.Tags {
				if leftTags[tag] {
					rate++
				}
			}
			rating[rate] = append(rating[rate], action)
		}
	}

	if len(rating) > 0 {
		// Get highest rated actions
		var maxRate uint64 = 0
		for rate := range rating {
			if rate > maxRate {
				maxRate = rate
			}
		}

		candidates := rating[maxRate]
		if len(candidates) > 0 {
			return candidates[utils.GlobalRandom.NextInt(len(candidates))]
		}
	}

	return nil
}

// GlobalTarget represents a long-term goal
type GlobalTarget struct {
	Name            string
	Tags            map[string]bool
	Power           float64
	TargetsPossible map[*LocalTarget]bool
	TargetsExecuted map[*LocalTarget]bool
}

// NewGlobalTarget creates a new GlobalTarget
func NewGlobalTarget(name string, tags []string, power float64, allTargets []*LocalTarget) *GlobalTarget {
	tagSet := make(map[string]bool)
	for _, tag := range tags {
		tagSet[tag] = true
	}

	targetsPossible := make(map[*LocalTarget]bool)
	for _, target := range allTargets {
		if len(utils.IntersectSlices(tags, getKeysFromMap(target.Tags))) > 0 {
			targetsPossible[target] = true
		}
	}

	return &GlobalTarget{
		Name:            name,
		Tags:            tagSet,
		Power:           power,
		TargetsPossible: targetsPossible,
		TargetsExecuted: make(map[*LocalTarget]bool),
	}
}

// MarkAsExecuted marks a local target as executed
func (gt *GlobalTarget) MarkAsExecuted(target *LocalTarget) {
	delete(gt.TargetsPossible, target)
	gt.TargetsExecuted[target] = true
}

// IsExecutedFull checks if all tags are covered
func (gt *GlobalTarget) IsExecutedFull() bool {
	remainingTags := make(map[string]bool)
	for tag := range gt.Tags {
		remainingTags[tag] = true
	}

	for target := range gt.TargetsExecuted {
		for tag := range target.Tags {
			delete(remainingTags, tag)
		}
	}

	return len(remainingTags) == 0
}

// Executable checks if the global target can be executed
func (gt *GlobalTarget) Executable(person *Human) bool {
	unclosedTags := make(map[string]bool)
	for tag := range gt.Tags {
		unclosedTags[tag] = true
	}

	// Remove executed tags
	for target := range gt.TargetsExecuted {
		for tag := range target.Tags {
			delete(unclosedTags, tag)
		}
	}

	// Check if remaining tags can be closed
	for target := range gt.TargetsPossible {
		if target.Executable(person) {
			for tag := range target.Tags {
				delete(unclosedTags, tag)
			}
		}
	}

	return len(unclosedTags) == 0
}

// ChooseTarget selects the best local target for this global target
func (gt *GlobalTarget) ChooseTarget(person *Human) *LocalTarget {
	leftTags := make(map[string]bool)
	for tag := range gt.Tags {
		leftTags[tag] = true
	}

	for target := range gt.TargetsExecuted {
		for tag := range target.Tags {
			delete(leftTags, tag)
		}
	}

	rating := make(map[uint64][]*LocalTarget)
	for target := range gt.TargetsPossible {
		if target.Executable(person) {
			var rate uint64 = 0
			for tag := range target.Tags {
				if leftTags[tag] {
					rate++
				}
			}
			rating[rate] = append(rating[rate], target)
		}
	}

	if len(rating) > 0 {
		// Get highest rated targets
		var maxRate uint64 = 0
		for rate := range rating {
			if rate > maxRate {
				maxRate = rate
			}
		}

		candidates := rating[maxRate]
		if len(candidates) > 0 {
			return candidates[utils.GlobalRandom.NextInt(len(candidates))]
		}
	}

	return nil
}

// Splash represents a temporary thought or need
type Splash struct {
	Name       string
	Tags       map[string]bool
	AppearTime uint64
	LifeLength uint64
}

// NewSplash creates a new splash
func NewSplash(name string, tags []string, lifeLength uint64) *Splash {
	tagSet := make(map[string]bool)
	for _, tag := range tags {
		tagSet[tag] = true
	}

	return &Splash{
		Name:       name,
		Tags:       tagSet,
		AppearTime: utils.GlobalTick.Get(),
		LifeLength: lifeLength,
	}
}

// IsExpired checks if splash has expired
func (s *Splash) IsExpired() bool {
	return utils.GlobalTick.Get()-s.AppearTime > s.LifeLength
}

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

// Vacancy represents a job opening
type Vacancy struct {
	Parent       *Job
	RequiredTags map[string]bool
	Payment      int
}

// Job represents a workplace
type Job struct {
	VacantPlaces map[*Vacancy]uint64
	HomeLocation *Location
}

// Location represents a place where humans live and work
type Location struct {
	Jobs   map[*Job]bool
	Humans map[*Human]bool
	Paths  map[*Path]bool
}

// Path represents a connection between locations
type Path struct {
	From  *Location
	To    *Location
	Price uint64
	Time  uint64
}

// Helper function to get keys from a map[string]bool
func getKeysFromMap(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
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
