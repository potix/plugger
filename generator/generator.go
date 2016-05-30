package generator

import (
	"sync"
)

type IdGenerator struct {
	currentId uint64
	mutex *sync.Mutex
}

func (ig *IdGenerator) Get() uint64 {
	ig.mutex.Lock()
	defer ig.mutex.Unlock()
	ig.currentId += 1
	if ig.currentId == 0 {
		ig.currentId += 1
	}
	return ig.currentId
}

func NewIdGenerator() *IdGenerator {
	return &IdGenerator {
		mutex : new(sync.Mutex),
	}
}

