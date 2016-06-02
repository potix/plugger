package plugger

import (
	"sync"
)

type idGenerator struct {
	currentId uint64
	mutex *sync.Mutex
}

func (ig *idGenerator) Get() uint64 {
	ig.mutex.Lock()
	defer ig.mutex.Unlock()
	ig.currentId += 1
	if ig.currentId == 0 {
		ig.currentId += 1
	}
	return ig.currentId
}

func newIdGenerator() *idGenerator {
	return &idGenerator {
		mutex : new(sync.Mutex),
	}
}

