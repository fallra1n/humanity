package components

import "github.com/fallra1n/humanity/src/config"

type Gender string

const (
	Male   Gender = "male"
	Female Gender = "female"
)

func (g Gender) GetDeathAge() float64 {
	if g == Male {
		return config.DeathAgeMale
	}

	return config.DeathAgeFemale
}
