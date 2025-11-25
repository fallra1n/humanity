package config

// Population and Employment Constants
const (
	// Total population in simulation
	TotalPopulation = 200
	
	// Employment rate (90% employed, 10% unemployed - typical for Russia)
	EmploymentRate = 0.9
	
	// City population distribution
	SmallCityPopulation = 0.4 * TotalPopulation  // 40% in small city
	LargeCityPopulation = 0.6 * TotalPopulation  // 60% in large city
)

// Economic Constants
const (
	// Starting capital for each person
	StartingMoney = 10000 // rubles
	
	// Daily living expenses
	DailyExpenses = 500 // rubles per day
	
	// Salary ranges for small city
	SmallCityJuniorSalaryMin = 30000 // rubles/month
	SmallCityJuniorSalaryMax = 45000 // rubles/month
	SmallCitySeniorSalaryMin = 50000 // rubles/month
	SmallCitySeniorSalaryMax = 75000 // rubles/month
	
	// Salary ranges for large city (higher cost of living)
	LargeCityJuniorSalaryMin = 35000 // rubles/month
	LargeCityJuniorSalaryMax = 55000 // rubles/month
	LargeCitySeniorSalaryMin = 60000 // rubles/month
	LargeCitySeniorSalaryMax = 90000 // rubles/month
	
	// Job vacancy counts
	SmallCityVacanciesMin = 3 // minimum positions per vacancy
	SmallCityVacanciesMax = 8 // maximum positions per vacancy (3 + 5)
	LargeCityVacanciesMin = 5 // minimum positions per vacancy
	LargeCityVacanciesMax = 11 // maximum positions per vacancy (5 + 6)
)

// Job Market Constants
const (
	// Job search frequency for unemployed (every 24 hours)
	UnemployedJobSearchInterval = 24
	
	// Job market check frequency for employed (every week)
	EmployedJobSearchInterval = 168 // hours in a week
	
	// Minimum work experience before job switching
	MinJobExperienceForSwitch = 168 // hours (1 week)
	
	// Minimum salary increase to consider job change
	MinSalaryIncreasePercent = 1.10 // 10% increase
	
	// Skill requirements
	MinSkillMatchForJob = 0.8      // 80% of required skills for initial job
	MinSkillMatchForSwitch = 0.7   // 70% of required skills for job switching
	
	// Job change probability range
	MinJobChangeProbability = 0.2  // 20% minimum chance
	MaxJobChangeProbability = 0.6  // 60% maximum chance
)

// Age and Life Constants
const (
	// Age generation parameters
	MinAge = 20.0
	MaxAge = 80.0
	MeanAge = 25.0
	AgeStdDev = 10.0
	
	// Death age threshold
	DeathAge = 80.0
	
	// Work experience range for initial job assignment
	MaxInitialWorkExperience = 2000 // hours
)

// Firing and Layoff Constants
const (
	// Poor performance (new employees)
	NewEmployeePeriod = 168        // hours (1 week)
	PoorPerformanceFireRate = 0.01 // 1% chance per hour
	
	// Economic downturn
	EconomicDownturnRate = 0.0005  // 0.05% chance per hour
	EconomicDownturnFireRate = 0.03 // 3% additional fire chance
	
	// High salary restructuring threshold and rate
	HighSalaryThreshold = 60000    // rubles/month
	RestructuringFireRate = 0.0003 // 0.03% chance per hour
	
	// Behavioral issues fire rate
	BehavioralIssuesFireRate = 0.005 // 0.5% additional chance
	
	// Age discrimination
	AgeDiscriminationThreshold = 55    // years
	AgeDiscriminationFireRate = 0.0001 // 0.01% chance per hour
	
	// Random layoffs
	RandomLayoffBaseRate = 0.00001 // very small base chance
	RandomLayoffFireRate = 0.001   // 0.1% chance
)

// Building Capacity Constants
const (
	// Small city building capacities
	SmallCityHospitalCapacity = 50
	SmallCitySchoolCapacity = 200
	SmallCityWorkplaceCapacity = 100
	SmallCityEntertainmentCapacity = 150
	SmallCityCafeCapacity = 30
	SmallCityShopCapacity = 40
	SmallCityHouseCapacity = 15
	
	// Large city building capacities
	LargeCityHospitalCapacity = 75
	LargeCitySchoolCapacity = 300
	LargeCityWorkplaceCapacity = 150
	LargeCityEntertainmentCapacity = 200
	LargeCityCafeCapacity = 40
	LargeCityShopCapacity = 50
	LargeCityHouseCapacity = 25
)

// Simulation Time Constants
const (
	// Time units
	HoursPerDay = 24
	HoursPerWeek = 168
	HoursPerMonth = 30 * HoursPerDay
	HoursPerYear = 365 * HoursPerDay
	
	// Simulation duration
	SimulationYears = 3
	TotalSimulationHours = SimulationYears * HoursPerYear
)

// Gender Distribution
const (
	// Gender probability (50/50 split)
	MaleGenderProbability = 0.5
)

// Global Target Assignment
const (
	// Number of global targets per person
	MinGlobalTargets = 2
	MaxGlobalTargets = 4 // 2 + 2
)

// Splash (temporary needs) Constants
const (
	// Splash lifetimes in hours
	NeedMoneyLifetime = 24
	JobLossLifetime = 72
	CareerAdvancementLifetime = 48
)

// Sleep Schedule Constants
const (
	// Sleep hours (23:00 to 07:00)
	SleepStartHour = 23  // 23:00
	SleepEndHour = 7     // 07:00
	SleepDuration = 8    // hours of sleep
)

// Work Schedule Constants
const (
	// Work hours (09:00 to 18:00)
	WorkStartHour = 9   // 09:00
	WorkEndHour = 18    // 18:00
	WorkDuration = 9    // hours of work
)

// Family and Birth Constants
const (
	// Marriage and family planning
	MinMarriageDurationForChildren = 8760 // hours (1 year)
	MinFamilyIncomeForChildren = 50000    // rubles combined monthly income
	MaxChildrenPerFamily = 4              // maximum children per couple
	
	// Birth probability and timing
	BaseBirthPlanningProbability = 0.25   // 25% chance per month for eligible couples
	PregnancyDurationHours = 6480         // hours (9 months)
	ChildExpensesPerDay = 300             // rubles per child per day
	
	// Age restrictions for having children
	MinMotherAge = 18.0
	MaxMotherAge = 45.0
	MinFatherAge = 18.0
	MaxFatherAge = 60.0
	
	// City family-friendly coefficients
	BaseFamilyCoefficient = 1.0           // base multiplier
	HospitalFamilyBonus = 0.2             // +20% per hospital
	SchoolFamilyBonus = 0.15              // +15% per school
	MaxFamilyCoefficient = 2.5           // maximum multiplier
)