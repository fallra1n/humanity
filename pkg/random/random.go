// Быстрый генератор случайных чисел для многопоточного приложения
package random

import (
	"math/rand"
	"runtime"
	"time"
	"unsafe"
)

var (
	generators    []*rand.Rand
	numGenerators int
)

func init() {
	numGenerators = runtime.GOMAXPROCS(0) // Количество ядер
	generators = make([]*rand.Rand, numGenerators)

	// Создаем отдельный генератор для каждого ядра
	for i := 0; i < numGenerators; i++ {
		generators[i] = rand.New(rand.NewSource(time.Now().UnixNano() + int64(i*1000)))
	}
}

// Получение хэша от ID горутины
func getGoroutineID() int {
	// Используем указатель на стек как уникальный ID горутины
	var dummy byte
	return int(uintptr(unsafe.Pointer(&dummy)))
}

func getRandom() *rand.Rand {
	// Берем хэш от ID горутины и находим соответствующий генератор
	gid := getGoroutineID()
	idx := gid % numGenerators
	if idx < 0 {
		idx = -idx
	}
	return generators[idx]
}

func FastFloat64() float64 {
	return getRandom().Float64()
}

func FastIntn(n int) int {
	return getRandom().Intn(n)
}

func FastUint64() uint64 {
	return getRandom().Uint64()
}

func FastInt63() int64 {
	return getRandom().Int63()
}
