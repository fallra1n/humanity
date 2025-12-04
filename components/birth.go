package components

// ProcessBirths обрабатывает прогресс беременности и роды
func ProcessBirths(people []*Human, globalTargets []*GlobalTarget) []*Human {
	var newChildren []*Human

	for _, person := range people {
		if person.Gender == Female && person.IsPregnant {
			newChild := person.ProcessPregnancy(people, globalTargets)
			if newChild != nil {
				// Ребенок родился!
				newChildren = append(newChildren, newChild)

				// Добавить ребенка в город
				person.HomeLocation.Humans[newChild] = true
			}
		}
	}

	return newChildren
}
