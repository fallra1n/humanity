package utils

import (
	"math/rand"
	"sync"
	"time"
)

// Random обеспечивает потокобезопасную генерацию случайных чисел
type Random struct {
	mu   sync.Mutex
	rand *rand.Rand
}

var GlobalRandom = &Random{
	rand: rand.New(rand.NewSource(time.Now().UnixNano())),
}

// Next возвращает случайное uint64
func (r *Random) Next() uint64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return uint64(r.rand.Int63())<<1 | uint64(r.rand.Int63n(2))
}

// NextFloat возвращает случайное float64 между 0 и 1
func (r *Random) NextFloat() float64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rand.Float64()
}

// NextInt возвращает случайное int в диапазоне [0, max)
func (r *Random) NextInt(max int) int {
	if max <= 0 {
		return 0
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rand.Intn(max)
}

// NextNormal возвращает нормально распределенное случайное число
func (r *Random) NextNormal(mean, stddev float64) float64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rand.NormFloat64()*stddev + mean
}
