package components

import (
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/fallra1n/humanity/src/config"
	"github.com/fallra1n/humanity/src/utils"
)

// Human представляет человека в симуляции
type Human struct {
	Age                    float64
	Gender                 Gender
	MaritalStatus          MaritalStatus
	Spouse                 *Human // Ссылка на супруга, если женат/замужем
	IsPregnant             bool   // True если в данный момент беременна
	PregnancyTime          uint64 // Часы с начала беременности
	Dead                   bool
	BusyHours              uint64
	Money                  int64
	Job                    *Vacancy
	JobTime                uint64
	HomeLocation           *Location
	CurrentBuilding        *Building // Где человек находится в данный момент
	WorkBuilding           *Building // Где человек работает (может быть nil если безработный)
	ResidentialBuilding    *Building
	Parents                map[*Human]float64
	Family                 map[*Human]float64
	Children               map[*Human]float64
	Friends                map[*Human]float64
	Splashes               []*Splash
	GlobalTargets          map[*GlobalTarget]bool
	CompletedGlobalTargets map[*GlobalTarget]bool
	Items                  map[string]int64

	// Мьютекс для потокобезопасного доступа к отношениям
	Mu sync.RWMutex
}

// NewHuman создает нового человека
func NewHuman(parents map[*Human]bool, homeLocation *Location, globalTargets []*GlobalTarget) *Human {
	// Генерация возраста с нормальным распределением
	age := math.Max(config.MinAge, math.Min(config.MaxAge, utils.GlobalRandom.NextNormal(config.MeanAge, config.AgeStdDev)))

	// Случайное назначение пола
	var gender Gender
	if utils.GlobalRandom.NextFloat() < config.MaleGenderProbability {
		gender = Male
	} else {
		gender = Female
	}

	human := &Human{
		Age:                    age,
		Gender:                 gender,
		MaritalStatus:          Single, // Начинаем как одинокий
		Spouse:                 nil,    // Изначально без супруга
		IsPregnant:             false,  // Изначально не беременна
		PregnancyTime:          0,      // Нет времени беременности
		Dead:                   false,
		BusyHours:              0,
		Money:                  7000,
		Job:                    nil,
		JobTime:                720,
		HomeLocation:           homeLocation,
		CurrentBuilding:        nil, // Будет установлено при назначении в жилое здание
		WorkBuilding:           nil, // Будет установлено при получении работы
		Parents:                make(map[*Human]float64),
		Family:                 make(map[*Human]float64),
		Children:               make(map[*Human]float64),
		Friends:                make(map[*Human]float64),
		Splashes:               make([]*Splash, 0),
		GlobalTargets:          make(map[*GlobalTarget]bool),
		CompletedGlobalTargets: make(map[*GlobalTarget]bool),
		Items:                  make(map[string]int64),
	}

	// Установить родителей
	for parent := range parents {
		human.Parents[parent] = 0.0
	}

	// Назначить случайные глобальные цели
	numTargets := 2 + utils.GlobalRandom.NextInt(2) // 2-3 цели
	if numTargets > len(globalTargets) {
		numTargets = len(globalTargets)
	}

	selectedTargets := make(map[string]bool)
	for len(human.GlobalTargets) < numTargets {
		target := globalTargets[utils.GlobalRandom.NextInt(len(globalTargets))]
		if !selectedTargets[target.Name] {
			// Создать копию глобальной цели для этого человека
			newTarget := &GlobalTarget{
				Name:            target.Name,
				Tags:            make(map[string]bool),
				Power:           target.Power,
				TargetsPossible: make(map[*LocalTarget]bool),
				TargetsExecuted: make(map[*LocalTarget]bool),
			}

			// Копировать теги
			for tag := range target.Tags {
				newTarget.Tags[tag] = true
			}

			// Копировать возможные цели
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

// IterateHour обрабатывает один час жизни человека
func (h *Human) IterateHour() {
	if h.Money <= 0 {
		splash := NewSplash("need_money", []string{"money", "well-being", "career"}, 24)
		h.Splashes = append(h.Splashes, splash)
	}

	// Старение отношений
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

	// Управление рабочим временем
	if h.Job == nil {
		h.JobTime = 721
	} else {
		h.JobTime++
	}

	// Удалить истекшие всплески
	validSplashes := make([]*Splash, 0)
	for _, splash := range h.Splashes {
		if !splash.IsExpired() {
			validSplashes = append(validSplashes, splash)
		}
	}
	h.Splashes = validSplashes

	// Обработка смерти
	if h.Age > h.Gender.GetDeathAge() {
		if !h.Dead {
			h.redistributeWealth()
			h.Money = 0
		}
		h.Dead = true
	}

	if h.Dead {
		return
	}

	// Старение человека
	h.Age += 1.0 / (24 * 365)

	// Ежедневные расходы
	if utils.GlobalTick.Get()%24 == 0 {
		dailyExpenses := int64(config.DailyExpenses)

		// Дополнительные расходы на детей
		dailyExpenses += int64(len(h.Children)) * config.ChildExpensesPerDay

		h.Money -= dailyExpenses

		// Месячная зарплата
		if h.Job != nil && utils.GlobalTick.Get()%(30*24) == 0 {
			h.Money += int64(h.Job.Payment)
		}
	}

	// Обработка беременности и планирования детей (только для женщин)
	if h.Gender == Female {
		// Пытаться планировать ребенка каждый час
		h.PlanChild()

		// Обработка текущей беременности
		// Примечание: ProcessPregnancy возвращает нового ребенка если происходят роды
		// Это будет обработано в main.go для добавления ребенка в список людей
	}

	// Перераспределить деньги в семье при необходимости
	if h.Money < 0 {
		h.redistributeMoneyInFamily()
	}

	// Проверить рынок труда на лучшие возможности
	h.checkJobMarket()

	// Обработка перемещения между зданиями
	h.handleMovement()

	// Обработка дружбы перенесена в main.go для потокобезопасности

	// Основная логика активности - проверить, время ли сна
	if utils.IsSleepTime(utils.GlobalTick.Get()) {
		// Во время сна (23:00 до 07:00), люди не выполняют действия
		// Они просто отдыхают и восстанавливаются
		return
	}

	if h.BusyHours > 0 {
		h.BusyHours--
	} else {
		h.performActions()
	}
}

// checkJobMarket периодически проверяет лучшие возможности трудоустройства
func (h *Human) checkJobMarket() {
	// Проверять рынок труда чаще и с меньшими требованиями к опыту
	if h.Job == nil {
		// Если безработный, искать работу не каждый час - каждые 24 часа для поддержания уровня безработицы
		if utils.GlobalTick.Get()%24 == 0 {
			findJob(h)
		}
		return
	}

	// Если трудоустроен, проверять лучшие возможности каждую неделю (168 часов)
	if h.JobTime < 168 || h.JobTime%168 != 0 {
		return
	}

	var betterJobs []*Vacancy
	currentSalary := h.Job.Payment

	// Искать лучшие работы в рабочих зданиях в том же городе
	h.HomeLocation.Mu.RLock()
	for building := range h.HomeLocation.Buildings {
		if building.Type == Workplace {
			building.Mu.RLock()
			for job := range building.Jobs {
				job.Mu.RLock()
				for vacancy, count := range job.VacantPlaces {
					// Искать работы с умеренным повышением зарплаты (10% или больше)
					minSalaryIncrease := int(float64(currentSalary) * 1.10)
					if count > 0 && vacancy.Payment >= minSalaryIncrease {
						// Сбалансированные требования для смены работы
						requiredSkills := 0
						hasSkills := 0

						for tag := range vacancy.RequiredTags {
							requiredSkills++
							if h.Items[tag] > 0 {
								hasSkills++
							}
						}

						// Принять работу если имеет как минимум 70% требуемых навыков или нет требований
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

	// Рассмотреть смену работы с умеренной вероятностью
	if len(betterJobs) > 0 {
		bestJob := betterJobs[0]
		for _, job := range betterJobs {
			if job.Payment > bestJob.Payment {
				bestJob = job
			}
		}

		// Умеренная вероятность смены работы (20-60% шанс)
		salaryIncrease := float64(bestJob.Payment-currentSalary) / float64(currentSalary)
		changeProb := math.Max(0.2, math.Min(0.6, salaryIncrease)) // 20-60% chance

		if utils.GlobalRandom.NextFloat() < changeProb {
			// Уволиться с текущей работы
			h.Job.Parent.Mu.Lock()
			h.Job.Parent.VacantPlaces[h.Job]++
			h.Job.Parent.Mu.Unlock()

			// Взять новую работу
			bestJob.Parent.Mu.Lock()
			h.Job = bestJob
			bestJob.Parent.VacantPlaces[bestJob]--
			h.JobTime = 0 // Сбросить опыт работы
			bestJob.Parent.Mu.Unlock()

			// Добавить всплеск о карьерном росте
			splash := NewSplash("career_advancement", []string{"career", "money", "well-being"}, 48)
			h.Splashes = append(h.Splashes, splash)
		}
	}
}

// FireEmployee увольняет человека с текущей работы
func (h *Human) FireEmployee(reason string) {
	if h.Job == nil {
		return
	}

	// Вернуть вакантную позицию
	h.Job.Parent.Mu.Lock()
	h.Job.Parent.VacantPlaces[h.Job]++
	h.Job.Parent.Mu.Unlock()

	// Удалить работу у человека
	h.Job = nil
	h.JobTime = 721 // Установить в состояние безработного

	// Добавить всплеск о потере работы
	splash := NewSplash("job_loss", []string{"money", "stress", "career"}, 72)
	h.Splashes = append(h.Splashes, splash)
}

// CanBeFired определяет, может ли человек быть уволен на основе различных факторов
func (h *Human) CanBeFired() (bool, string) {
	if h.Job == nil {
		return false, ""
	}

	// Сбалансированная вероятность увольнения для поддержания естественной безработицы
	var fireProb float64 = 0.0
	var reason string

	// 1. Плохая производительность (новые сотрудники с малым опытом)
	if h.JobTime < 168 { // Менее 168 часов (1 неделя) опыта
		fireProb += 0.01 // 1% шанс
		reason = "poor_performance"
	}

	// 2. Экономический спад (умеренный шанс)
	if utils.GlobalRandom.NextFloat() < 0.0005 { // 0.05% шанс в час
		fireProb += 0.03 // 3% дополнительный шанс
		reason = "economic_downturn"
	}

	// 3. Реструктуризация компании (для высокооплачиваемых сотрудников)
	if h.Job.Payment > 60000 { // Высокооплачиваемые сотрудники
		fireProb += 0.0003 // 0.03% шанс
		reason = "restructuring"
	}

	// 4. Поведенческие проблемы (умеренное влияние)
	negativeSpashes := 0
	for _, splash := range h.Splashes {
		if splash.Name == "stress" || splash.Name == "job_loss" {
			negativeSpashes++
		}
	}
	if negativeSpashes > 1 { // Если есть негативные всплески
		fireProb += 0.005 // 0.5% дополнительный шанс
		reason = "behavioral_issues"
	}

	// 5. Возрастная дискриминация (небольшой шанс для пожилых работников)
	if h.Age > 55 { // Возрастной порог
		fireProb += 0.0001 // 0.01% шанс
		reason = "age_discrimination"
	}

	// 6. Случайные увольнения для поддержания уровня безработицы
	if utils.GlobalRandom.NextFloat() < 0.00001 { // Очень маленький базовый шанс
		fireProb += 0.001 // 0.1% шанс
		reason = "random_layoff"
	}

	return utils.GlobalRandom.NextFloat() < fireProb, reason
}

// handleMovement управляет перемещением между зданиями в зависимости от времени суток
func (h *Human) handleMovement() {
	currentHour := utils.GetHourOfDay(utils.GlobalTick.Get())

	// Идти на работу в рабочие часы (9:00-17:59) если трудоустроен и это рабочий день
	if currentHour >= 9 && currentHour < 18 && h.Job != nil && h.WorkBuilding != nil && utils.IsWorkDay(utils.GlobalTick.Get()) {
		if h.CurrentBuilding != h.WorkBuilding {
			h.CurrentBuilding = h.WorkBuilding
		}
	}

	// Идти домой после работы (18:00+) или в нерабочие часы
	if (currentHour >= 18 || currentHour < 9 || !utils.IsWorkDay(utils.GlobalTick.Get())) && h.ResidentialBuilding != nil {
		if h.CurrentBuilding != h.ResidentialBuilding {
			h.CurrentBuilding = h.ResidentialBuilding
		}
	}

	// Оставаться дома во время сна (23:00-07:00)
	if utils.IsSleepTime(utils.GlobalTick.Get()) && h.ResidentialBuilding != nil {
		if h.CurrentBuilding != h.ResidentialBuilding {
			h.CurrentBuilding = h.ResidentialBuilding
		}
	}
}

// MarryWith создает брак между двумя людьми
func (h *Human) MarryWith(other *Human) {
	if h.MaritalStatus == Married || other.MaritalStatus == Married {
		return // Один из них уже женат/замужем
	}

	if h == other {
		return // Нельзя жениться на себе
	}

	// Определить кто к кому переезжает (невеста переезжает к жениху)
	var bride, groom *Human
	if h.Gender == Female {
		bride = h
		groom = other
	} else {
		bride = other
		groom = h
	}

	// Создать двусторонний брак
	h.MaritalStatus = Married
	h.Spouse = other
	other.MaritalStatus = Married
	other.Spouse = h

	// Добавить к семейным отношениям если еще не там
	if _, exists := h.Family[other]; !exists {
		h.Family[other] = 0.0
	}
	if _, exists := other.Family[h]; !exists {
		other.Family[h] = 0.0
	}

	// Невеста переезжает в жилое здание жениха (всегда, даже если в том же здании)
	if bride.ResidentialBuilding != nil && groom.ResidentialBuilding != nil {
		bride.ResidentialBuilding.MoveToSpouse(bride, groom)
	}
}

// Divorce завершает брак между двумя людьми
func (h *Human) Divorce() {
	if h.MaritalStatus != Married || h.Spouse == nil {
		return // Не женат/замужем
	}

	spouse := h.Spouse

	// Завершить двусторонний брак
	h.MaritalStatus = Single
	h.Spouse = nil
	spouse.MaritalStatus = Single
	spouse.Spouse = nil
}

// IsCompatibleWith проверяет, совместимы ли два человека для брака
func (h *Human) IsCompatibleWith(other *Human) bool {
	// Проверить, что оба одинокие
	if h.MaritalStatus != Single || other.MaritalStatus != Single {
		return false
	}

	// Проверить, что полы противоположные
	if h.Gender == other.Gender {
		return false
	}

	// Проверить разность в возрасте (максимум 10 лет)
	ageDiff := h.Age - other.Age
	if ageDiff < 0 {
		ageDiff = -ageDiff
	}
	if ageDiff > 10.0 {
		return false
	}

	// Проверить, что они знают друг друга как минимум 6 месяцев (0.5 года)
	friendship, exists := h.Friends[other]
	if !exists || friendship < 0.5 {
		return false
	}

	// Проверить, что у них есть как минимум 3 общих типа глобальных целей
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

// CanHaveChildren проверяет, может ли человек иметь детей на основе возраста и семейного положения
func (h *Human) CanHaveChildren() bool {
	// Должен быть женат/замужем
	if h.MaritalStatus != Married || h.Spouse == nil {
		return false
	}

	// Возрастные ограничения на основе пола
	if h.Gender == Female {
		return h.Age >= config.MinMotherAge && h.Age <= config.MaxMotherAge
	} else {
		return h.Age >= config.MinFatherAge && h.Age <= config.MaxFatherAge
	}
}

// GetFamilyIncome вычисляет совокупный доход супружеской пары
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

// ShouldPlanChild определяет, должна ли пара планировать ребенка
func (h *Human) ShouldPlanChild() bool {
	// Только женщины могут забеременеть
	if h.Gender != Female {
		return false
	}

	// Уже беременна
	if h.IsPregnant {
		return false
	}

	// Проверить основные требования
	if !h.CanHaveChildren() {
		return false
	}

	// Проверить, женаты ли требуемое время
	marriageTime, exists := h.Family[h.Spouse]
	marriageTimeHours := marriageTime * 365 * 24 // Convert years to hours
	if !exists || marriageTimeHours < float64(config.MinMarriageDurationForChildren) {
		return false
	}

	// Проверить, есть ли у пары цель счастливой семьи
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

	// Проверка финансовой стабильности
	if h.GetFamilyIncome() < int64(config.MinFamilyIncomeForChildren) {
		return false
	}

	// Ограничить количество детей
	if len(h.Children) >= config.MaxChildrenPerFamily {
		return false
	}

	return true
}

// PlanChild начинает процесс беременности
func (h *Human) PlanChild() {
	if h.ShouldPlanChild() {
		// Вычислить семейный коэффициент города
		cityCoefficient := CalculateFamilyFriendlyCoefficient(h.HomeLocation)

		// Использовать настраиваемую базовую вероятность, модифицированную городским коэффициентом
		baseProb := config.BaseBirthPlanningProbability / (30 * 24) // Per hour probability
		adjustedProb := baseProb * cityCoefficient

		if utils.GlobalRandom.NextFloat() < adjustedProb {
			h.IsPregnant = true
			h.PregnancyTime = 0

			// Добавить всплеск беременности
			splash := NewSplash("pregnancy", []string{"family", "health", "responsibility"}, config.PregnancyDurationHours)
			h.Splashes = append(h.Splashes, splash)
		}
	}
}

// ProcessPregnancy обрабатывает прогресс беременности и роды
func (h *Human) ProcessPregnancy(people []*Human, globalTargets []*GlobalTarget) *Human {
	if !h.IsPregnant {
		return nil
	}

	h.PregnancyTime++

	// Проверить, завершена ли продолжительность беременности
	if h.PregnancyTime >= config.PregnancyDurationHours {
		// Родить!
		return h.GiveBirth(people, globalTargets)
	}

	return nil
}

// GiveBirth создает нового ребенка
func (h *Human) GiveBirth(people []*Human, globalTargets []*GlobalTarget) *Human {
	// Сбросить статус беременности
	h.IsPregnant = false
	h.PregnancyTime = 0

	// Создать ребенка с родителями
	parents := make(map[*Human]bool)
	parents[h] = true
	parents[h.Spouse] = true

	child := NewHuman(parents, h.HomeLocation, globalTargets)
	child.Age = 0.0 // Новорожденный
	child.Money = 0 // Дети не имеют денег
	child.ResidentialBuilding = h.ResidentialBuilding
	child.CurrentBuilding = h.ResidentialBuilding

	// Добавить ребенка к детям родителей
	h.Children[child] = 0.0
	h.Spouse.Children[child] = 0.0

	// Добавить родителей к родителям ребенка
	child.Parents[h] = 0.0
	child.Parents[h.Spouse] = 0.0

	// Добавить всплеск рождения родителям
	birthSplash := NewSplash("child_birth", []string{"family", "happiness", "responsibility"}, 168) // 1 неделя
	h.Splashes = append(h.Splashes, birthSplash)
	h.Spouse.Splashes = append(h.Spouse.Splashes, birthSplash)

	return child
}

// redistributeWealth распределяет деньги семье при смерти
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

// redistributeMoneyInFamily пытается получить деньги от членов семьи
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

// performActions обрабатывает основную логику принятия решений
func (h *Human) performActions() {
	if len(h.GlobalTargets) == 0 {
		return
	}

	rating := make(map[float64][]*GlobalTarget)

	if len(h.Splashes) > 0 {
		// Оценить цели на основе всплесков
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
		// Оценить цели на основе выполнимости и силы
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

	// Получить цели с наивысшим рейтингом
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

// PrintInitialInfo выводит подробную информацию о человеке в начале
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

// PrintFinalInfo выводит подробную информацию о человеке в конце
func (h *Human) PrintFinalInfo(id int) {
	fmt.Printf("=== Human #%d Final State ===\n", id)
	fmt.Printf("Status: %s\n", h.getLifeStatus())
	fmt.Printf("Age: %.1f years\n", h.Age)
	fmt.Printf("Gender: %s\n", h.Gender)
	fmt.Printf("Money: %d rubles\n", h.Money)
	fmt.Printf("Job: %s\n", h.getJobStatus())
	fmt.Printf("Family Status: %s\n", h.MaritalStatus)

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

// Вспомогательные методы для форматирования вывода
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

// Вспомогательная функция для получения ключей из map[string]bool
func getKeysFromMap(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

// getCurrentLocationString возвращает строковое представление текущего местоположения
func (h *Human) getCurrentLocationString() string {
	if h.CurrentBuilding == nil {
		return "Unknown"
	}
	return h.CurrentBuilding.Name
}
