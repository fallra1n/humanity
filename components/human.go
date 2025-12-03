package components

import (
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/fallra1n/humanity/config"
	"github.com/fallra1n/humanity/utils"
)

// Human represents a person in the simulation
type Human struct {
	Age                    float64
	Gender                 Gender
	MaritalStatus          MaritalStatus
	Spouse                 *Human // Reference to spouse if married
	IsPregnant             bool   // True if currently pregnant
	PregnancyTime          uint64 // Hours since pregnancy started
	Dead                   bool
	BusyHours              uint64
	Money                  int64
	Job                    *Vacancy
	JobTime                uint64
	HomeLocation           *Location
	CurrentBuilding        *Building // Where the person currently is
	WorkBuilding           *Building // Where the person works (can be nil if unemployed)
	ResidentialBuilding    *Building
	Parents                map[*Human]float64
	Family                 map[*Human]float64
	Children               map[*Human]float64
	Friends                map[*Human]float64
	Splashes               []*Splash
	GlobalTargets          map[*GlobalTarget]bool
	CompletedGlobalTargets map[*GlobalTarget]bool
	Items                  map[string]int64

	// Mutex for thread-safe access to relationships
	Mu sync.RWMutex
}

// NewHuman creates a new human
func NewHuman(parents map[*Human]bool, homeLocation *Location, globalTargets []*GlobalTarget) *Human {
	// Generate age with normal distribution
	age := math.Max(config.MinAge, math.Min(config.MaxAge, utils.GlobalRandom.NextNormal(config.MeanAge, config.AgeStdDev)))

	// Randomly assign gender
	var gender Gender
	if utils.GlobalRandom.NextFloat() < config.MaleGenderProbability {
		gender = Male
	} else {
		gender = Female
	}

	human := &Human{
		Age:                    age,
		Gender:                 gender,
		MaritalStatus:          Single, // Start as single
		Spouse:                 nil,    // No spouse initially
		IsPregnant:             false,  // Not pregnant initially
		PregnancyTime:          0,      // No pregnancy time
		Dead:                   false,
		BusyHours:              0,
		Money:                  7000,
		Job:                    nil,
		JobTime:                720,
		HomeLocation:           homeLocation,
		CurrentBuilding:        nil, // Will be set when assigned to residential building
		WorkBuilding:           nil, // Will be set when getting a job
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

	// Look for jobs in workplace buildings in the same city
	h.HomeLocation.Mu.RLock()
	for building := range h.HomeLocation.Buildings {
		if building.Type == Workplace {
			building.Mu.RLock()
			for job := range building.Jobs {
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
			building.Mu.RUnlock()
		}
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
		// Set work building to the building where the job is
		h.WorkBuilding = chosen.Parent.Building
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

	// Look for better jobs in workplace buildings in the same city
	h.HomeLocation.Mu.RLock()
	for building := range h.HomeLocation.Buildings {
		if building.Type == Workplace {
			building.Mu.RLock()
			for job := range building.Jobs {
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
			building.Mu.RUnlock()
		}
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
		dailyExpenses := int64(config.DailyExpenses)

		// Additional expenses for children
		dailyExpenses += int64(len(h.Children)) * config.ChildExpensesPerDay

		h.Money -= dailyExpenses

		// Monthly salary
		if h.Job != nil && utils.GlobalTick.Get()%(30*24) == 0 {
			h.Money += int64(h.Job.Payment)
		}
	}

	// Process pregnancy and child planning (only for women)
	if h.Gender == Female {
		// Try to plan child every hour
		h.PlanChild()

		// Process ongoing pregnancy
		// Note: ProcessPregnancy returns a new child if birth occurs
		// This will be handled in main.go to add the child to the people slice
	}

	// Redistribute money within family if needed
	if h.Money < 0 {
		h.redistributeMoneyInFamily()
	}

	// Check job market for better opportunities
	h.checkJobMarket()

	// Handle movement between buildings
	h.handleMovement()

	// Friendship processing moved to main.go for thread safety

	// Main activity logic - check if it's sleep time
	if utils.IsSleepTime(utils.GlobalTick.Get()) {
		// During sleep hours (23:00 to 07:00), humans don't perform actions
		// They just rest and recover
		return
	}

	if h.BusyHours > 0 {
		h.BusyHours--
	} else {
		h.performActions()
	}
}

// handleMovement manages movement between buildings based on time of day
func (h *Human) handleMovement() {
	currentHour := utils.GetHourOfDay(utils.GlobalTick.Get())

	// Go to work during work hours (9:00-17:59) if employed and it's a work day
	if currentHour >= 9 && currentHour < 18 && h.Job != nil && h.WorkBuilding != nil && utils.IsWorkDay(utils.GlobalTick.Get()) {
		if h.CurrentBuilding != h.WorkBuilding {
			h.CurrentBuilding = h.WorkBuilding
		}
	}

	// Go home after work (18:00+) or during non-work hours
	if (currentHour >= 18 || currentHour < 9 || !utils.IsWorkDay(utils.GlobalTick.Get())) && h.ResidentialBuilding != nil {
		if h.CurrentBuilding != h.ResidentialBuilding {
			h.CurrentBuilding = h.ResidentialBuilding
		}
	}

	// Stay home during sleep hours (23:00-07:00)
	if utils.IsSleepTime(utils.GlobalTick.Get()) && h.ResidentialBuilding != nil {
		if h.CurrentBuilding != h.ResidentialBuilding {
			h.CurrentBuilding = h.ResidentialBuilding
		}
	}
}

// MarryWith creates a marriage between two humans
func (h *Human) MarryWith(other *Human) {
	if h.MaritalStatus == Married || other.MaritalStatus == Married {
		return // One of them is already married
	}

	if h == other {
		return // Can't marry yourself
	}

	// Determine who moves to whom (bride moves to groom)
	var bride, groom *Human
	if h.Gender == Female {
		bride = h
		groom = other
	} else {
		bride = other
		groom = h
	}

	// Create bidirectional marriage
	h.MaritalStatus = Married
	h.Spouse = other
	other.MaritalStatus = Married
	other.Spouse = h

	// Add to family relationships if not already there
	if _, exists := h.Family[other]; !exists {
		h.Family[other] = 0.0
	}
	if _, exists := other.Family[h]; !exists {
		other.Family[h] = 0.0
	}

	// Bride moves to groom's residential building (always, even if in same building)
	if bride.ResidentialBuilding != nil && groom.ResidentialBuilding != nil {
		bride.ResidentialBuilding.MoveToSpouse(bride, groom)
	}
}

// Divorce ends the marriage between two humans
func (h *Human) Divorce() {
	if h.MaritalStatus != Married || h.Spouse == nil {
		return // Not married
	}

	spouse := h.Spouse

	// End bidirectional marriage
	h.MaritalStatus = Single
	h.Spouse = nil
	spouse.MaritalStatus = Single
	spouse.Spouse = nil
}

// IsCompatibleWith checks if two humans are compatible for marriage
func (h *Human) IsCompatibleWith(other *Human) bool {
	// Check if both are single
	if h.MaritalStatus != Single || other.MaritalStatus != Single {
		return false
	}

	// Check if genders are opposite
	if h.Gender == other.Gender {
		return false
	}

	// Check age difference (max 10 years)
	ageDiff := h.Age - other.Age
	if ageDiff < 0 {
		ageDiff = -ageDiff
	}
	if ageDiff > 10.0 {
		return false
	}

	// Check if they know each other for at least 6 months (0.5 years)
	friendship, exists := h.Friends[other]
	if !exists || friendship < 0.5 {
		return false
	}

	// Check if they have at least 3 common global target types
	commonTargets := 0
	for hTarget := range h.GlobalTargets {
		for otherTarget := range other.GlobalTargets {
			if hTarget.Name == otherTarget.Name {
				commonTargets++
				break
			}
		}
	}

	return commonTargets >= 3
}

// CanHaveChildren checks if a person can have children based on age and marital status
func (h *Human) CanHaveChildren() bool {
	// Must be married
	if h.MaritalStatus != Married || h.Spouse == nil {
		return false
	}

	// Age restrictions based on gender
	if h.Gender == Female {
		return h.Age >= config.MinMotherAge && h.Age <= config.MaxMotherAge
	} else {
		return h.Age >= config.MinFatherAge && h.Age <= config.MaxFatherAge
	}
}

// GetFamilyIncome calculates combined income of married couple
func (h *Human) GetFamilyIncome() int64 {
	income := int64(0)
	if h.Job != nil {
		income += int64(h.Job.Payment)
	}
	if h.Spouse != nil && h.Spouse.Job != nil {
		income += int64(h.Spouse.Job.Payment)
	}
	return income
}

// ShouldPlanChild determines if a couple should plan for a child
func (h *Human) ShouldPlanChild() bool {
	// Only women can get pregnant
	if h.Gender != Female {
		return false
	}

	// Already pregnant
	if h.IsPregnant {
		return false
	}

	// Check basic requirements
	if !h.CanHaveChildren() {
		return false
	}

	// Check if married for required duration
	marriageTime, exists := h.Family[h.Spouse]
	marriageTimeHours := marriageTime * 365 * 24 // Convert years to hours
	if !exists || marriageTimeHours < float64(config.MinMarriageDurationForChildren) {
		return false
	}

	// Check if couple has happy_family goal
	hasHappyFamilyGoal := false
	for target := range h.GlobalTargets {
		if target.Name == "happy_family" {
			hasHappyFamilyGoal = true
			break
		}
	}
	if !hasHappyFamilyGoal {
		return false
	}

	// Financial stability check
	if h.GetFamilyIncome() < int64(config.MinFamilyIncomeForChildren) {
		return false
	}

	// Limit number of children
	if len(h.Children) >= config.MaxChildrenPerFamily {
		return false
	}

	return true
}

// PlanChild starts pregnancy process
func (h *Human) PlanChild() {
	if h.ShouldPlanChild() {
		// Calculate city family coefficient
		cityCoefficient := CalculateFamilyFriendlyCoefficient(h.HomeLocation)

		// Use configurable base probability, modified by city coefficient
		baseProb := config.BaseBirthPlanningProbability / (30 * 24) // Per hour probability
		adjustedProb := baseProb * cityCoefficient

		if utils.GlobalRandom.NextFloat() < adjustedProb {
			h.IsPregnant = true
			h.PregnancyTime = 0

			// Add pregnancy splash
			splash := NewSplash("pregnancy", []string{"family", "health", "responsibility"}, config.PregnancyDurationHours)
			h.Splashes = append(h.Splashes, splash)
		}
	}
}

// ProcessPregnancy handles pregnancy progression and birth
func (h *Human) ProcessPregnancy(people []*Human, globalTargets []*GlobalTarget) *Human {
	if !h.IsPregnant {
		return nil
	}

	h.PregnancyTime++

	// Check if pregnancy duration is complete
	if h.PregnancyTime >= config.PregnancyDurationHours {
		// Give birth!
		return h.GiveBirth(people, globalTargets)
	}

	return nil
}

// GiveBirth creates a new child
func (h *Human) GiveBirth(people []*Human, globalTargets []*GlobalTarget) *Human {
	// Reset pregnancy status
	h.IsPregnant = false
	h.PregnancyTime = 0

	// Create child with parents
	parents := make(map[*Human]bool)
	parents[h] = true
	parents[h.Spouse] = true

	child := NewHuman(parents, h.HomeLocation, globalTargets)
	child.Age = 0.0 // Newborn
	child.Money = 0 // Children don't have money
	child.ResidentialBuilding = h.ResidentialBuilding
	child.CurrentBuilding = h.ResidentialBuilding

	// Add child to parents' children
	h.Children[child] = 0.0
	h.Spouse.Children[child] = 0.0

	// Add parents to child's parents
	child.Parents[h] = 0.0
	child.Parents[h.Spouse] = 0.0

	// Add birth splash to parents
	birthSplash := NewSplash("child_birth", []string{"family", "happiness", "responsibility"}, 168) // 1 week
	h.Splashes = append(h.Splashes, birthSplash)
	h.Spouse.Splashes = append(h.Spouse.Splashes, birthSplash)

	return child
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

	if len(h.Family) > 0 || len(h.Children) > 0 || len(h.Friends) > 0 {
		fmt.Printf("Relationships: %d family members, %d children, %d friends\n",
			len(h.Family), len(h.Children), len(h.Friends))
	}

	// Show current location and work info
	fmt.Printf("Current Location: %s\n", h.getCurrentLocationString())
	if h.WorkBuilding != nil {
		fmt.Printf("Work Location: %s\n", h.WorkBuilding.Name)
	} else {
		fmt.Printf("Work Location: Unemployed\n")
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

// getCurrentLocationString returns a string representation of current location
func (h *Human) getCurrentLocationString() string {
	if h.CurrentBuilding == nil {
		return "Unknown"
	}
	return h.CurrentBuilding.Name
}
