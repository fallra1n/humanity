package components

import (
	"fmt"

	"github.com/fallra1n/humanity/types"
	"github.com/fallra1n/humanity/utils"
)

func CreateVacancies(city *types.Location) []*types.Vacancy {
	// Create 10 companies with diverse vacancies
	var allVacancies []*types.Vacancy
	companyNames := []string{"TechCorp", "FinanceInc", "HealthPlus", "EduCenter", "RetailChain",
		"ManufacturingLtd", "ServicePro", "CreativeStudio", "LogisticsCo", "ConsultingGroup"}

	for i, companyName := range companyNames {
		company := &types.Job{
			VacantPlaces: make(map[*types.Vacancy]uint64),
			HomeLocation: city,
		}

		// Create 2 different vacancies per company
		// Junior position
		juniorVacancy := &types.Vacancy{
			Parent:       company,
			RequiredTags: make(map[string]bool),
			Payment:      30000 + utils.GlobalRandom.NextInt(20000), // 30-50k rubles
		}

		// Senior position (may require education)
		seniorVacancy := &types.Vacancy{
			Parent:       company,
			RequiredTags: make(map[string]bool),
			Payment:      50000 + utils.GlobalRandom.NextInt(30000), // 50-80k rubles
		}

		// Some senior positions require education
		if i%3 == 0 { // Every 3rd company requires education for senior role
			seniorVacancy.RequiredTags["engineer_diploma"] = true
		}

		// Each vacancy has 5-10 positions
		company.VacantPlaces[juniorVacancy] = uint64(5 + utils.GlobalRandom.NextInt(6))
		company.VacantPlaces[seniorVacancy] = uint64(5 + utils.GlobalRandom.NextInt(6))

		allVacancies = append(allVacancies, juniorVacancy, seniorVacancy)
		city.Jobs[company] = true

		fmt.Printf("Created %s: Junior (%d rub, %d positions), Senior (%d rub, %d positions)\n",
			companyName, juniorVacancy.Payment, company.VacantPlaces[juniorVacancy],
			seniorVacancy.Payment, company.VacantPlaces[seniorVacancy])
	}

	return allVacancies
}
