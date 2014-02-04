package utils

import "sync"

var (
	namelock = &sync.Mutex{}
	namelist = make(map[string]*sync.Mutex)
)

type NameLock struct {
	name string
	mu   *sync.Mutex
}

func NewNameLock(name string) *NameLock {
	namelock.Lock()
	defer namelock.Unlock()

	mu := namelist[name]
	if mu == nil {
		mu = &sync.Mutex{}
		namelist[name] = mu
	}
	return &NameLock{
		name: name,
		mu:   mu,
	}
}

func (l *NameLock) Lock() {
	l.mu.Lock()
}

func (l *NameLock) Unlock() {
	//namelock.Lock()
	//defer namelock.Unlock()
	// FIXME: I dont know if need to delete from namelist
	l.mu.Unlock()
}
