package lock

import "sync"

type DoneStore struct {
	sync.RWMutex
	// loclName --> doneChannel
	inner map[string]chan struct{}
}

func NewDoneStore() *DoneStore {
	var d DoneStore
	d.inner = make(map[string]chan struct{})
	return &d
}

func (d *DoneStore) CreateDoneChan(lockName string) {
	d.Lock()
	defer d.Unlock()

	d.inner[lockName] = make(chan struct{})
}

func (d *DoneStore) CLoseDoneChan(lockName string) {
	d.Lock()
	defer d.Unlock()

	_, ok := d.inner[lockName]
	if ok {
		close(d.inner[lockName])
	}
}

func (d *DoneStore) GetDoneChan(lockName string) chan struct{} {
	d.RLock()
	defer d.RUnlock()

	ch := d.inner[lockName]
	return ch
}
