package components

import "sync"

// BuildingType представляет тип здания
type BuildingType string

const (
	Hospital         BuildingType = "hospital"
	School           BuildingType = "school"
	Workplace        BuildingType = "workplace"
	Entertainment    BuildingType = "entertainment"
	Cafe             BuildingType = "cafe"
	Shop             BuildingType = "shop"
	ResidentialHouse BuildingType = "residential_house"
)

// Building представляет структуру в локации
type Building struct {
	ID       int
	Type     BuildingType
	Name     string
	Location *Location

	// Для рабочих мест - содержит вакансии
	Jobs map[*Job]bool

	// Для жилых зданий - содержит жителей
	Residents map[*Human]bool

	// Общая вместимость и текущая заполненность
	Capacity int
	Occupied int

	// Для жилых зданий - продажа квартир
	ApartmentsForSale []*Human // Список квартир доступных для продажи (предыдущие владельцы)
	ApartmentPrice    int64    // Цена за квартиру в рублях

	// Потокобезопасность
	Mu sync.RWMutex
}

// NewBuilding создает новое здание
func NewBuilding(id int, buildingType BuildingType, name string, capacity int, location *Location) *Building {
	building := &Building{
		ID:        id,
		Type:      buildingType,
		Name:      name,
		Location:  location,
		Jobs:      make(map[*Job]bool),
		Residents: make(map[*Human]bool),
		Capacity:  capacity,
		Occupied:  0,
	}

	// Инициализировать продажу квартир для жилых зданий
	if buildingType == ResidentialHouse {
		// Разные цены для малых и больших городов
		if location.Name == "Greenville" {
			building.ApartmentPrice = 2000000 // 2 миллиона рублей для малого города
		} else {
			building.ApartmentPrice = 3000000 // 3 миллиона рублей для большого города
		}
		building.ApartmentsForSale = make([]*Human, 0) // Изначально нет квартир на продажу
	}

	return building
}

// AddJob добавляет работу в рабочее здание
func (b *Building) AddJob(job *Job) bool {
	if b.Type != Workplace {
		return false
	}

	b.Mu.Lock()
	defer b.Mu.Unlock()

	b.Jobs[job] = true
	job.Building = b
	return true
}

// AddResident добавляет жителя в жилое здание
func (b *Building) AddResident(human *Human) bool {
	if b.Type != ResidentialHouse {
		return false
	}

	b.Mu.Lock()
	defer b.Mu.Unlock()

	if b.Occupied >= b.Capacity {
		return false
	}

	b.Residents[human] = true
	b.Occupied++
	human.ResidentialBuilding = b
	human.CurrentBuilding = b // Начать дома
	return true
}

// RemoveResident удаляет жителя из жилого здания
func (b *Building) RemoveResident(human *Human) {
	if b.Type != ResidentialHouse {
		return
	}

	b.Mu.Lock()
	defer b.Mu.Unlock()

	if b.Residents[human] {
		delete(b.Residents, human)
		b.Occupied--
		human.ResidentialBuilding = nil
	}
}

// GetAvailableJobs возвращает все доступные работы в этом здании
func (b *Building) GetAvailableJobs() []*Vacancy {
	if b.Type != Workplace {
		return nil
	}

	b.Mu.RLock()
	defer b.Mu.RUnlock()

	var vacancies []*Vacancy
	for job := range b.Jobs {
		job.Mu.RLock()
		for vacancy, count := range job.VacantPlaces {
			if count > 0 {
				vacancies = append(vacancies, vacancy)
			}
		}
		job.Mu.RUnlock()
	}

	return vacancies
}

// HasCapacity проверяет, есть ли у здания доступная вместимость
func (b *Building) HasCapacity() bool {
	b.Mu.RLock()
	defer b.Mu.RUnlock()
	return b.Occupied < b.Capacity
}

// GetOccupancyRate возвращает коэффициент заполненности в процентах
func (b *Building) GetOccupancyRate() float64 {
	b.Mu.RLock()
	defer b.Mu.RUnlock()

	if b.Capacity == 0 {
		return 0
	}
	return float64(b.Occupied) / float64(b.Capacity) * 100
}

// PutApartmentForSale выставляет квартиру на продажу когда житель выезжает
func (b *Building) PutApartmentForSale(previousOwner *Human) {
	if b.Type != ResidentialHouse {
		return
	}

	b.Mu.Lock()
	defer b.Mu.Unlock()

	b.ApartmentsForSale = append(b.ApartmentsForSale, previousOwner)
}

// BuyApartment позволяет человеку купить квартиру
func (b *Building) BuyApartment(human *Human) bool {
	if b.Type != ResidentialHouse {
		return false
	}

	b.Mu.Lock()
	defer b.Mu.Unlock()

	if len(b.ApartmentsForSale) <= 0 {
		return false
	}

	// Проверить, есть ли у человека достаточно денег
	if human.Money < b.ApartmentPrice {
		return false
	}

	// Обработать покупку
	human.Money -= b.ApartmentPrice
	// Удалить первую квартиру из списка продаж
	b.ApartmentsForSale = b.ApartmentsForSale[1:]
	b.Residents[human] = true
	b.Occupied++
	human.ResidentialBuilding = b
	human.CurrentBuilding = b

	return true
}

// MoveToSpouse перемещает человека в жилое здание супруга
func (b *Building) MoveToSpouse(human *Human, spouse *Human) bool {
	if b.Type != ResidentialHouse {
		return false
	}

	spouseBuilding := spouse.ResidentialBuilding
	if spouseBuilding == nil || spouseBuilding.Type != ResidentialHouse {
		return false
	}

	// Заблокировать текущее здание сначала
	b.Mu.Lock()

	// Удалить из текущего здания и выставить квартиру на продажу
	if b.Residents[human] {
		delete(b.Residents, human)
		b.Occupied--
		b.ApartmentsForSale = append(b.ApartmentsForSale, human) // Выставить квартиру на продажу напрямую
	}
	b.Mu.Unlock()

	// Заблокировать здание супруга
	spouseBuilding.Mu.Lock()
	defer spouseBuilding.Mu.Unlock()

	// Добавить в здание супруга (без проверки вместимости, так как это перемещение в рамках существующей вместимости)
	spouseBuilding.Residents[human] = true
	human.ResidentialBuilding = spouseBuilding
	human.CurrentBuilding = spouseBuilding

	return true
}
