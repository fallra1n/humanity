package config

// BuildingCoordinate представляет координаты здания
type BuildingCoordinate struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// CityCoordinates содержит координаты для всех типов зданий в городе
type CityCoordinates struct {
	Hospitals     []BuildingCoordinate `json:"hospitals"`
	Schools       []BuildingCoordinate `json:"schools"`
	Workplaces    []BuildingCoordinate `json:"workplaces"`
	Entertainment []BuildingCoordinate `json:"entertainment"`
	Cafes         []BuildingCoordinate `json:"cafes"`
	Shops         []BuildingCoordinate `json:"shops"`
	Residential   []BuildingCoordinate `json:"residential"`
}

// (малый город - City 1)
var SmallCityCoordinates = CityCoordinates{
	// 1 больница
	Hospitals: []BuildingCoordinate{
		{Lat: 55.5681, Lon: 39.4260},
	},

	// 1 школа
	Schools: []BuildingCoordinate{
		{Lat: 55.5656, Lon: 39.4241},
	},

	// 2 офиса
	Workplaces: []BuildingCoordinate{
		{Lat: 55.5658, Lon: 39.4342},
		{Lat: 55.5664, Lon: 39.4254},
	},

	// 1 развлекательный центр
	Entertainment: []BuildingCoordinate{
		{Lat: 55.5575, Lon: 39.4315},
	},

	// 1 кафе
	Cafes: []BuildingCoordinate{
		{Lat: 55.5651, Lon: 39.4268},
	},

	// 1 магазин
	Shops: []BuildingCoordinate{
		{Lat: 55.5685, Lon: 39.4185},
	},

	// 3 жилых дома
	Residential: []BuildingCoordinate{
		{Lat: 55.5670, Lon: 39.4258},
		{Lat: 55.5670, Lon: 39.4258},
		{Lat: 55.5665, Lon: 39.4264},
	},
}

// (большой город - City 2)
var BigCityCoordinates = CityCoordinates{
	// 2 больницы
	Hospitals: []BuildingCoordinate{
		{Lat: 55.581565, Lon: 39.541703},
		{Lat: 55.581960, Lon: 39.540375},
	},

	// 2 школы
	Schools: []BuildingCoordinate{
		{Lat: 55.581436, Lon: 39.533431},
		{Lat: 55.577821, Lon: 39.527903},
	},

	// 3 офиса
	Workplaces: []BuildingCoordinate{
		{Lat: 55.582656, Lon: 39.533124},
		{Lat: 55.572849, Lon: 39.531288},
		{Lat: 55.579003, Lon: 39.541344},
	},

	// 1 развлекательный центр
	Entertainment: []BuildingCoordinate{
		{Lat: 55.582575, Lon: 39.522617},
	},

	// 2 кафе
	Cafes: []BuildingCoordinate{
		{Lat: 55.575140, Lon: 39.525415},
		{Lat: 55.575915, Lon: 39.528772},
	},

	// 2 магазина
	Shops: []BuildingCoordinate{
		{Lat: 55.576933, Lon: 39.521386},
		{Lat: 55.577918, Lon: 39.526411},
	},

	// 3 жилых дома
	Residential: []BuildingCoordinate{
		{Lat: 55.578274, Lon: 39.521161},
		{Lat: 55.579445, Lon: 39.527625},
		{Lat: 55.579132, Lon: 39.524888},
	},
}

// GetCoordinatesForCity возвращает координаты для указанного города
func GetCoordinatesForCity(cityName string) *CityCoordinates {
	switch cityName {
	case "City 1":
		return &SmallCityCoordinates
	case "City 2":
		return &BigCityCoordinates
	default:
		return &SmallCityCoordinates
	}
}

// GetCoordinateForBuilding возвращает координаты для здания определенного типа и индекса в указанном городе
func GetCoordinateForBuilding(cityName, buildingType string, index int) (float64, float64) {
	cityCoords := GetCoordinatesForCity(cityName)

	var coords []BuildingCoordinate

	switch buildingType {
	case "hospital":
		coords = cityCoords.Hospitals
	case "school":
		coords = cityCoords.Schools
	case "workplace":
		coords = cityCoords.Workplaces
	case "entertainment":
		coords = cityCoords.Entertainment
	case "cafe":
		coords = cityCoords.Cafes
	case "shop":
		coords = cityCoords.Shops
	case "residential_house":
		coords = cityCoords.Residential
	default:
		// По умолчанию возвращаем центр города
		if cityName == "City 1" {
			return 55.563289, 39.427530
		} else {
			return 55.579629, 39.531722
		}
	}

	if len(coords) == 0 {
		// Если нет координат для типа здания, возвращаем центр города
		if cityName == "City 1" {
			return 55.563289, 39.427530
		} else {
			return 55.579629, 39.531722
		}
	}

	// Используем индекс с циклическим повтором если зданий больше чем координат
	coordIndex := index % len(coords)
	coord := coords[coordIndex]

	return coord.Lat, coord.Lon
}
