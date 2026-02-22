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

	// Для жилых зданий - цена квартиры (покупка/продажа через администрацию)
	ApartmentPrice int64 // Цена за квартиру в рублях

	// Координаты (широта и долгота)
	Lat, Lon float64

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

	// Инициализировать цены квартир для жилых зданий
	if buildingType == ResidentialHouse {
		// Разные цены для малых и больших городов
		if location.Name == "City 1" {
			building.ApartmentPrice = 2000000 // 2 миллиона рублей для малого города
		} else {
			building.ApartmentPrice = 3000000 // 3 миллиона рублей для большого города
		}
	}

	return building
}

// NewBuildingWithCoordinates создает новое здание с координатами
func NewBuildingWithCoordinates(id int, buildingType BuildingType, name string, capacity int, location *Location, lat, lon float64) *Building {
	building := NewBuilding(id, buildingType, name, capacity, location)
	building.SetCoordinates(lat, lon)
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
	human.CurrentBuilding = b
	return true
}

// SellApartmentToAdmin продает квартиру администрации (мгновенная продажа)
func (b *Building) SellApartmentToAdmin(seller *Human) bool {
	if b.Type != ResidentialHouse {
		return false
	}

	b.Mu.Lock()
	defer b.Mu.Unlock()

	// Удалить жителя из здания
	delete(b.Residents, seller)
	b.Occupied--
	seller.ResidentialBuilding = nil

	// Выплатить деньги за квартиру (администрация покупает по полной стоимости)
	seller.Money += b.ApartmentPrice

	return true
}

// BuyApartmentFromAdmin позволяет человеку купить квартиру у администрации
func (b *Building) BuyApartmentFromAdmin(buyer *Human) bool {
	if b.Type != ResidentialHouse {
		return false
	}

	b.Mu.Lock()
	defer b.Mu.Unlock()

	// Проверить, есть ли свободные места в здании
	if b.Occupied >= b.Capacity {
		return false
	}

	// Проверить, есть ли у человека достаточно денег
	if buyer.Money < b.ApartmentPrice {
		return false
	}

	// Обработать покупку у администрации
	buyer.Money -= b.ApartmentPrice
	b.Residents[buyer] = true
	b.Occupied++
	buyer.ResidentialBuilding = b
	buyer.CurrentBuilding = b

	return true
}

// SetCoordinates устанавливает координаты здания
func (b *Building) SetCoordinates(lat, lon float64) {
	b.Mu.Lock()
	defer b.Mu.Unlock()
	
	b.Lat = lat
	b.Lon = lon
}

// GetCoordinates возвращает координаты здания
func (b *Building) GetCoordinates() (float64, float64) {
	b.Mu.RLock()
	defer b.Mu.RUnlock()
	
	return b.Lat, b.Lon
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

	// Удалить из текущего здания и продать квартиру администрации
	if b.Residents[human] {
		delete(b.Residents, human)
		b.Occupied--
		// Продать квартиру администрации (получить деньги)
		human.Money += b.ApartmentPrice
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
