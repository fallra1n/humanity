package utils

import (
	"math/rand"
	"sync"
	"time"
)

// Random provides thread-safe random number generation
type Random struct {
	mu   sync.Mutex
	rand *rand.Rand
}

var GlobalRandom = &Random{
	rand: rand.New(rand.NewSource(time.Now().UnixNano())),
}

// Next returns a random uint64
func (r *Random) Next() uint64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return uint64(r.rand.Int63())<<1 | uint64(r.rand.Int63n(2))
}

// NextFloat returns a random float64 between 0 and 1
func (r *Random) NextFloat() float64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rand.Float64()
}

// NextInt returns a random int in range [0, max)
func (r *Random) NextInt(max int) int {
	if max <= 0 {
		return 0
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rand.Intn(max)
}

// NextNormal returns a normally distributed random number
func (r *Random) NextNormal(mean, stddev float64) float64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rand.NormFloat64()*stddev + mean
}
