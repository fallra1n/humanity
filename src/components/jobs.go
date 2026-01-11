package components

import "github.com/fallra1n/humanity/src/utils"

// findJob пытается найти новую работу
func findJob(h *Human) {
	var possibleJobs []*Vacancy

	// Искать работу в рабочих зданиях в том же городе
	h.HomeLocation.Mu.RLock()
	for building := range h.HomeLocation.Buildings {
		if building.Type == Workplace {
			building.Mu.RLock()
			for job := range building.Jobs {
				job.Mu.RLock()
				for vacancy, count := range job.VacantPlaces {
					if count > 0 {
						// Сбалансированные требования к работе
						requiredSkills := 0
						hasSkills := 0

						for tag := range vacancy.RequiredTags {
							requiredSkills++
							if h.Items[tag] > 0 {
								hasSkills++
							}
						}

						// Разрешить работу если:
						// 1. Нет требований вообще
						// 2. Имеет как минимум 80% требуемых навыков
						// 3. Безработный и очень отчаянный (деньги < -10000)
						if requiredSkills == 0 ||
							float64(hasSkills)/float64(requiredSkills) >= 0.8 ||
							(h.Job == nil && h.Money < -10000) {

							// Если трудоустроен, рассматривать только более высокооплачиваемые работы
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

		// Уволиться со старой работы
		if h.Job != nil {
			h.Job.Parent.Mu.Lock()
			h.Job.Parent.VacantPlaces[h.Job]++
			h.Job.Parent.Mu.Unlock()
		}

		// Взять новую работу
		chosen.Parent.Mu.Lock()
		h.Job = chosen
		chosen.Parent.VacantPlaces[chosen]--
		h.JobTime = 0
		// Установить рабочее здание в здание, где находится работа
		h.WorkBuilding = chosen.Parent.Building
		chosen.Parent.Mu.Unlock()
	}
}
