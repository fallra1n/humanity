package config

// Константы населения и занятости
const (
	// Общее население в симуляции
	TotalPopulation = 10

	// Уровень занятости (90% трудоустроены, 10% безработные - типично для России)
	EmploymentRate = 0.9

	// Распределение населения по городам
	SmallCityPopulation = 0.4 * TotalPopulation // 40% в малом городе
	LargeCityPopulation = 0.6 * TotalPopulation // 60% в большом городе
)

// Экономические константы
const (
	// Стартовый капитал для каждого человека
	StartingMoney = 10000 // рубли

	// Ежедневные расходы на жизнь
	DailyExpenses = 500 // рубли в день

	// Диапазоны зарплат для малого города
	SmallCityJuniorSalaryMin = 30000 // рубли/месяц
	SmallCityJuniorSalaryMax = 45000 // рубли/месяц
	SmallCitySeniorSalaryMin = 50000 // рубли/месяц
	SmallCitySeniorSalaryMax = 75000 // рубли/месяц

	// Диапазоны зарплат для большого города (более высокая стоимость жизни)
	LargeCityJuniorSalaryMin = 35000 // рубли/месяц
	LargeCityJuniorSalaryMax = 55000 // рубли/месяц
	LargeCitySeniorSalaryMin = 60000 // рубли/месяц
	LargeCitySeniorSalaryMax = 90000 // рубли/месяц

	// Количество вакансий
	SmallCityVacanciesMin = 3  // минимум позиций на вакансию
	SmallCityVacanciesMax = 8  // максимум позиций на вакансию (3 + 5)
	LargeCityVacanciesMin = 5  // минимум позиций на вакансию
	LargeCityVacanciesMax = 11 // максимум позиций на вакансию (5 + 6)
)

// Константы рынка труда
const (
	// Частота поиска работы для безработных (каждые 24 часа)
	UnemployedJobSearchInterval = 24

	// Частота проверки рынка труда для трудоустроенных (каждую неделю)
	EmployedJobSearchInterval = 168 // часов в неделе

	// Минимальный опыт работы перед сменой работы
	MinJobExperienceForSwitch = 168 // часов (1 неделя)

	// Минимальное увеличение зарплаты для рассмотрения смены работы
	MinSalaryIncreasePercent = 1.10 // 10% увеличение

	// Требования к навыкам
	MinSkillMatchForJob    = 0.8 // 80% требуемых навыков для первоначальной работы
	MinSkillMatchForSwitch = 0.7 // 70% требуемых навыков для смены работы

	// Диапазон вероятности смены работы
	MinJobChangeProbability = 0.2 // 20% минимальный шанс
	MaxJobChangeProbability = 0.6 // 60% максимальный шанс
)

// Константы возраста и жизни
const (
	// Параметры генерации возраста
	MinAge    = 20.0
	MaxAge    = 80.0
	MeanAge   = 25.0
	AgeStdDev = 10.0

	// Порог возраста смерти
	DeathAgeFemale = 78.0
	DeathAgeMale = 68.0

	// Диапазон опыта работы для первоначального назначения работы
	MaxInitialWorkExperience = 2000 // часы
)

// Константы увольнений и сокращений
const (
	// Плохая производительность (новые сотрудники)
	NewEmployeePeriod       = 168  // часов (1 неделя)
	PoorPerformanceFireRate = 0.01 // 1% шанс в час

	// Экономический спад
	EconomicDownturnRate     = 0.0005 // 0.05% шанс в час
	EconomicDownturnFireRate = 0.03   // 3% дополнительный шанс увольнения

	// Порог и коэффициент реструктуризации высоких зарплат
	HighSalaryThreshold   = 60000  // рубли/месяц
	RestructuringFireRate = 0.0003 // 0.03% шанс в час

	// Коэффициент увольнения за поведенческие проблемы
	BehavioralIssuesFireRate = 0.005 // 0.5% дополнительный шанс

	// Возрастная дискриминация
	AgeDiscriminationThreshold = 55     // лет
	AgeDiscriminationFireRate  = 0.0001 // 0.01% шанс в час

	// Случайные увольнения
	RandomLayoffBaseRate = 0.00001 // очень маленький базовый шанс
	RandomLayoffFireRate = 0.001   // 0.1% шанс
)

// Константы вместимости зданий
const (
	// Вместимость зданий малого города
	SmallCityHospitalCapacity      = 50
	SmallCitySchoolCapacity        = 200
	SmallCityWorkplaceCapacity     = 100
	SmallCityEntertainmentCapacity = 150
	SmallCityCafeCapacity          = 30
	SmallCityShopCapacity          = 40
	SmallCityHouseCapacity         = 30

	// Вместимость зданий большого города
	LargeCityHospitalCapacity      = 75
	LargeCitySchoolCapacity        = 300
	LargeCityWorkplaceCapacity     = 150
	LargeCityEntertainmentCapacity = 200
	LargeCityCafeCapacity          = 40
	LargeCityShopCapacity          = 50
	LargeCityHouseCapacity         = 45
)

// Константы времени симуляции
const (
	// Единицы времени
	HoursPerDay   = 24
	HoursPerWeek  = 168
	HoursPerMonth = 30 * HoursPerDay
	HoursPerYear  = 365 * HoursPerDay

	// Продолжительность симуляции
	SimulationYears      = 2
	TotalSimulationHours = SimulationYears * HoursPerYear
)

// Распределение по полу
const (
	// Вероятность мужского пола (разделение 50/50)
	MaleGenderProbability = 0.5
)

// Назначение глобальных целей
const (
	// Количество глобальных целей на человека
	MinGlobalTargets = 2
	MaxGlobalTargets = 4 // 2 + 2
)

// Константы всплесков (временных потребностей)
const (
	// Время жизни всплесков в часах
	NeedMoneyLifetime         = 24
	JobLossLifetime           = 72
	CareerAdvancementLifetime = 48
)

// Константы расписания сна
const (
	// Часы сна (23:00 до 07:00)
	SleepStartHour = 23 // 23:00
	SleepEndHour   = 7  // 07:00
	SleepDuration  = 8  // часов сна
)

// Константы рабочего расписания
const (
	// Рабочие часы (09:00 до 18:00)
	WorkStartHour = 9  // 09:00
	WorkEndHour   = 18 // 18:00
	WorkDuration  = 9  // часов работы
)

// Константы семьи и рождения
const (
	// Брак и планирование семьи
	MinMarriageDurationForChildren = 8760  // часов (1 год)
	MinFamilyIncomeForChildren     = 50000 // рубли совокупный месячный доход
	MaxChildrenPerFamily           = 4     // максимум детей на пару

	// Вероятность рождения и время
	BaseBirthPlanningProbability = 0.25 // 25% шанс в месяц для подходящих пар
	PregnancyDurationHours       = 6480 // часов (9 месяцев)
	ChildExpensesPerDay          = 300  // рубли на ребенка в день

	// Возрастные ограничения для рождения детей
	MinMotherAge = 18.0
	MaxMotherAge = 45.0
	MinFatherAge = 18.0
	MaxFatherAge = 60.0

	// Коэффициенты дружелюбности города к семьям
	BaseFamilyCoefficient = 1.0  // базовый множитель
	HospitalFamilyBonus   = 0.2  // +20% за больницу
	SchoolFamilyBonus     = 0.15 // +15% за школу
	MaxFamilyCoefficient  = 2.5  // максимальный множитель
)
