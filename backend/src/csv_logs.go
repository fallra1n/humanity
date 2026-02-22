package src

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"github.com/fallra1n/humanity/src/components"
)

// logToCSV записывает текущее состояние всех людей в CSV файл
func logToCSV(people []*components.Human, hour uint64) error {
	filename := "log.csv"

	// Проверить, существует ли файл, чтобы определить, нужно ли писать заголовки
	fileExists := false
	if _, err := os.Stat(filename); err == nil {
		fileExists = true
	}

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Записать заголовок если это первый раз
	if !fileExists {
		header := []string{"hour", "agent_id", "age", "gender", "alive", "money", "location", "building_type", "job_status", "marital_status", "geo"}
		if err := writer.Write(header); err != nil {
			return err
		}
	}

	// Записать данные для каждого человека
	for _, person := range people {
		location := "unknown"
		buildingType := "unknown"

		if person.CurrentBuilding != nil {
			location = person.CurrentBuilding.Name
			buildingType = string(person.CurrentBuilding.Type)
		}

		jobStatus := "unemployed"
		if person.Job != nil {
			jobStatus = "employed"
		}

		// Получаем координаты здания
		geoCoords := ""
		if person.CurrentBuilding != nil {
			lat, lon := person.CurrentBuilding.GetCoordinates()
			geoCoords = fmt.Sprintf("%.6f,%.6f", lat, lon)
		}

		row := []string{
			strconv.FormatUint(hour, 10),
			strconv.Itoa(components.GlobalHumanStorage.Get(person)),
			fmt.Sprintf("%.2f", person.Age),
			string(person.Gender),
			fmt.Sprintf("%t", !person.Dead),
			strconv.FormatInt(person.Money, 10),
			location,
			buildingType,
			jobStatus,
			string(person.MaritalStatus),
			geoCoords,
		}

		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}
