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
	Gender                 Gender
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

	// Randomly assign gender (50/50 chance)
	var gender Gender
	if utils.GlobalRandom.NextFloat() < 0.5 {
		gender = Male
	} else {
		gender = Female
	}

	human := &Human{
		Age:                    age,
		Gender:                 gender,
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
		}
	}

	GlobalHumanStorage.Append(human)
	return human
}

// findJob attempts to find a new job
func (h *Human) findJob() {
	var possibleJobs []*Vacancy

	h.HomeLocation.Mu.RLock()
	for job := range h.HomeLocation.Jobs {
		job.Mu.RLock()
		for vacancy, count := range job.VacantPlaces {
			if count > 0 {
				// Balanced job requirements
				requiredSkills := 0
				hasSkills := 0
				
				for tag := range vacancy.RequiredTags {
					requiredSkills++
					if h.Items[tag] > 0 {
						hasSkills++
					}
				}
				
				// Allow job if:
				// 1. No requirements at all
				// 2. Has at least 80% of required skills
				// 3. Unemployed and very desperate (money < -10000)
				if requiredSkills == 0 ||
				   float64(hasSkills)/float64(requiredSkills) >= 0.8 ||
				   (h.Job == nil && h.Money < -10000) {
					
					// If employed, only consider better paying jobs
					if h.Job == nil || vacancy.Payment > h.Job.Payment {
						possibleJobs = append(possibleJobs, vacancy)
					}
				}
			}
		}
		job.Mu.RUnlock()
	}
	h.HomeLocation.Mu.RUnlock()

	if len(possibleJobs) > 0 {
		chosen := possibleJobs[utils.GlobalRandom.NextInt(len(possibleJobs))]

		// Quit old job
		if h.Job != nil {
			h.Job.Parent.Mu.Lock()
			h.Job.Parent.VacantPlaces[h.Job]++
			h.Job.Parent.Mu.Unlock()
		}

		// Take new job
		chosen.Parent.Mu.Lock()
		h.Job = chosen
		chosen.Parent.VacantPlaces[chosen]--
		h.JobTime = 0
		chosen.Parent.Mu.Unlock()
	}
}

// checkJobMarket periodically checks for better job opportunities
func (h *Human) checkJobMarket() {
	// Check job market more frequently and with less experience required
	if h.Job == nil {
		// If unemployed, search for job but not every hour - every 24 hours to maintain unemployment rate
		if utils.GlobalTick.Get()%24 == 0 {
			h.findJob()
		}
		return
	}

	// If employed, check for better opportunities every week (168 hours)
	if h.JobTime < 168 || h.JobTime%168 != 0 {
		return
	}

	var betterJobs []*Vacancy
	currentSalary := h.Job.Payment

	h.HomeLocation.Mu.RLock()
	for job := range h.HomeLocation.Jobs {
		job.Mu.RLock()
		for vacancy, count := range job.VacantPlaces {
			// Look for jobs with moderate pay increase (10% or more)
			minSalaryIncrease := int(float64(currentSalary) * 1.10)
			if count > 0 && vacancy.Payment >= minSalaryIncrease {
				// Balanced requirements for job switching
				requiredSkills := 0
				hasSkills := 0
				
				for tag := range vacancy.RequiredTags {
					requiredSkills++
					if h.Items[tag] > 0 {
						hasSkills++
					}
				}
				
				// Accept job if has at least 70% of required skills or no requirements
				if requiredSkills == 0 || float64(hasSkills)/float64(requiredSkills) >= 0.7 {
					betterJobs = append(betterJobs, vacancy)
				}
			}
		}
		job.Mu.RUnlock()
	}
	h.HomeLocation.Mu.RUnlock()

	// Consider job change with moderate probability
	if len(betterJobs) > 0 {
		bestJob := betterJobs[0]
		for _, job := range betterJobs {
			if job.Payment > bestJob.Payment {
				bestJob = job
			}
		}

		// Moderate probability of job change (20-60% chance)
		salaryIncrease := float64(bestJob.Payment-currentSalary) / float64(currentSalary)
		changeProb := math.Max(0.2, math.Min(0.6, salaryIncrease)) // 20-60% chance

		if utils.GlobalRandom.NextFloat() < changeProb {
			// Quit current job
			h.Job.Parent.Mu.Lock()
			h.Job.Parent.VacantPlaces[h.Job]++
			h.Job.Parent.Mu.Unlock()
			
			// Take new job
			bestJob.Parent.Mu.Lock()
			h.Job = bestJob
			bestJob.Parent.VacantPlaces[bestJob]--
			h.JobTime = 0 // Reset job experience
			bestJob.Parent.Mu.Unlock()

			// Add a splash about career advancement
			splash := NewSplash("career_advancement", []string{"career", "money", "well-being"}, 48)
			h.Splashes = append(h.Splashes, splash)
		}
	}
}

// FireEmployee fires the human from their current job
func (h *Human) FireEmployee(reason string) {
	if h.Job == nil {
		return
	}

	// Return the vacant position
	h.Job.Parent.Mu.Lock()
	h.Job.Parent.VacantPlaces[h.Job]++
	h.Job.Parent.Mu.Unlock()

	// Remove job from human
	h.Job = nil
	h.JobTime = 721 // Set to unemployed state

	// Add splash about job loss
	splash := NewSplash("job_loss", []string{"money", "stress", "career"}, 72)
	h.Splashes = append(h.Splashes, splash)
}

// CanBeFired determines if a human can be fired based on various factors
func (h *Human) CanBeFired() (bool, string) {
	if h.Job == nil {
		return false, ""
	}

	// Balanced firing probability to maintain natural unemployment
	var fireProb float64 = 0.0
	var reason string

	// 1. Poor performance (new employees with low experience)
	if h.JobTime < 168 { // Less than 168 hours (1 week) experience
		fireProb += 0.01 // 1% chance
		reason = "poor_performance"
	}

	// 2. Economic downturn (moderate chance)
	if utils.GlobalRandom.NextFloat() < 0.0005 { // 0.05% chance per hour
		fireProb += 0.03 // 3% additional chance
		reason = "economic_downturn"
	}

	// 3. Company restructuring (for high salary employees)
	if h.Job.Payment > 60000 { // High salary employees
		fireProb += 0.0003 // 0.03% chance
		reason = "restructuring"
	}

	// 4. Behavioral issues (moderate impact)
	negativeSpashes := 0
	for _, splash := range h.Splashes {
		if splash.Name == "stress" || splash.Name == "job_loss" {
			negativeSpashes++
		}
	}
	if negativeSpashes > 1 { // If any negative splashes
		fireProb += 0.005 // 0.5% additional chance
		reason = "behavioral_issues"
	}

	// 5. Age discrimination (small chance for older workers)
	if h.Age > 55 { // Age threshold
		fireProb += 0.0001 // 0.01% chance
		reason = "age_discrimination"
	}

	// 6. Random layoffs to maintain unemployment rate
	if utils.GlobalRandom.NextFloat() < 0.00001 { // Very small base chance
		fireProb += 0.001 // 0.1% chance
		reason = "random_layoff"
	}

	return utils.GlobalRandom.NextFloat() < fireProb, reason
}

// IterateHour processes one hour of the human's life
func (h *Human) IterateHour() {
	if h.Money <= 0 {
		splash := NewSplash("need_money", []string{"money", "well-being", "career"}, 24)
		h.Splashes = append(h.Splashes, splash)
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

		// Monthly salary
		if h.Job != nil && utils.GlobalTick.Get()%(30*24) == 0 {
			h.Money += int64(h.Job.Payment)
		}
	}

	// Redistribute money within family if needed
	if h.Money < 0 {
		h.redistributeMoneyInFamily()
	}

	// Check job market for better opportunities
	h.checkJobMarket()

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
			selectedAction.Apply(h)
			selectedLocalTarget.MarkAsExecuted(selectedAction)

			if selectedLocalTarget.IsExecutedFull() {
				selectedGlobalTarget.MarkAsExecuted(selectedLocalTarget)

				if selectedGlobalTarget.IsExecutedFull() {
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
	fmt.Printf("Gender: %s\n", h.Gender)
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
	fmt.Printf("Gender: %s\n", h.Gender)
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
