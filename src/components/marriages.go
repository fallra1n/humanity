package components

import "github.com/fallra1n/humanity/src/utils"

// ProcessMarriages обрабатывает формирование браков между совместимыми людьми
func ProcessMarriages(people []*Human) {
	// Группировать одиноких людей по их текущему зданию
	buildingGroups := make(map[*Building][]*Human)

	for _, person := range people {
		if person.Dead || person.CurrentBuilding == nil || person.MaritalStatus != Single {
			continue
		}
		buildingGroups[person.CurrentBuilding] = append(buildingGroups[person.CurrentBuilding], person)
	}

	// Обработать потенциальные браки в каждом здании
	for _, group := range buildingGroups {
		if len(group) < 2 {
			continue // Нужно как минимум 2 человека для формирования браков
		}

		// Проверить все пары людей в здании
		for i := 0; i < len(group); i++ {
			for j := i + 1; j < len(group); j++ {
				person1 := group[i]
				person2 := group[j]

				// Проверить совместимость
				if person1.IsCompatibleWith(person2) {
					// Небольшой шанс пожениться (5% в час при совместимости)
					if utils.GlobalRandom.NextFloat() < 0.05 {
						person1.MarryWith(person2)
					}
				}
			}
		}
	}
}
