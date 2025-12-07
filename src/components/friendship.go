package components

import (
	"github.com/fallra1n/humanity/src/utils"
)

// ProcessFriendships обрабатывает формирование дружбы между людьми в одном здании
func ProcessFriendships(people []*Human) {
	// Группировать людей по их текущему зданию
	buildingGroups := make(map[*Building][]*Human)

	for _, person := range people {
		if person.Dead || person.CurrentBuilding == nil {
			continue
		}
		buildingGroups[person.CurrentBuilding] = append(buildingGroups[person.CurrentBuilding], person)
	}

	// Обработать дружбу в каждом здании
	for _, group := range buildingGroups {
		if len(group) < 2 {
			continue // Нужно как минимум 2 человека для формирования дружбы
		}

		// Проверить все пары людей в здании
		for i := 0; i < len(group); i++ {
			for j := i + 1; j < len(group); j++ {
				person1 := group[i]
				person2 := group[j]

				// Пропустить если уже друзья
				if _, alreadyFriends := person1.Friends[person2]; alreadyFriends {
					continue
				}

				// 25% шанс стать друзьями
				if utils.GlobalRandom.NextFloat() < 0.25 {
					// Создать двустороннюю дружбу
					person1.Friends[person2] = 0.0 // Начать с 0 силы отношений
					person2.Friends[person1] = 0.0
				}
			}
		}
	}
}
