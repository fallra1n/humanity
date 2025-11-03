package types

import "sync"

// HumanStorage manages human ID mapping
type HumanStorage struct {
	links map[interface{}]int
	count int
	mu    sync.RWMutex
}

var GlobalHumanStorage = &HumanStorage{
	links: make(map[interface{}]int),
	count: 0,
}

func (hs *HumanStorage) Append(human interface{}) {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	if _, exists := hs.links[human]; !exists {
		hs.links[human] = hs.count
		hs.count++
	}
}

func (hs *HumanStorage) Get(human interface{}) int {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	if id, exists := hs.links[human]; exists {
		return id
	}
	return -1
}
