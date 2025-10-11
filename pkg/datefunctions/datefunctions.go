package datefunctions

import "time"

// Симуляция течения времени
type SimulationTime struct {
	startTime   time.Time
	currentTime time.Time
}

func NewSimulationTime(startTime time.Time) *SimulationTime {
	return &SimulationTime{
		startTime:   startTime,
		currentTime: startTime,
	}
}

// Продвижение времени на один час
func (st *SimulationTime) Tick() {
	st.currentTime = st.currentTime.Add(time.Hour)
}

// Возвращает текущее время
func (st *SimulationTime) Now() time.Time {
	return st.currentTime
}

// Возвращает время начала симуляции
func (st *SimulationTime) StartTime() time.Time {
	return st.startTime
}

// Возвращает разницу между текущим временем и временем начала симуляции в часах
func (st *SimulationTime) ElapsedTime() time.Duration {
	return st.currentTime.Sub(st.startTime) / time.Hour
}
