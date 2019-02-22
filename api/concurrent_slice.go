package api

import (
	"sync"
)

type Item = []byte


type ConcurrentSlice struct {
	sync.Mutex
	Items []Item
}

func (cs *ConcurrentSlice) append(items ...Item) {
	cs.Lock()
	defer cs.Unlock()

	cs.Items = append(cs.Items, items...)
}

func (cs *ConcurrentSlice) dumpSlice() []Item {
	cs.Lock()
	defer cs.Unlock()

	cpy := cs.Items
	cs.Items = []Item{}
	return cpy
}