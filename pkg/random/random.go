package random

import (
	"math/rand"
	"sync"
	"time"
)

var randPool = sync.Pool{
	New: func() interface{} {
		return rand.New(rand.NewSource(time.Now().UnixNano()))
	},
}

func FastFloat64() float64 {
	r := randPool.Get().(*rand.Rand)
	defer randPool.Put(r)
	return r.Float64()
}

func FastIntn(n int) int {
	r := randPool.Get().(*rand.Rand)
	defer randPool.Put(r)
	return r.Intn(n)
}

func FastUint64() uint64 {
	r := randPool.Get().(*rand.Rand)
	defer randPool.Put(r)
	return r.Uint64()
}

func FastInt63() int64 {
	r := randPool.Get().(*rand.Rand)
	defer randPool.Put(r)
	return r.Int63()
}
