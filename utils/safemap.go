package utils

import "sync"

type SafeMap struct {
	db map[string]interface{}
	sync.Mutex
}

func NewSafeMap() *SafeMap {
	return &SafeMap{
		db: make(map[string]interface{}),
	}
}

func (this *SafeMap) Set(key string, v interface{}) {
	this.Lock()
	defer this.Unlock()
	this.db[key] = v
}

func (this *SafeMap) Del(key string) {
	this.Lock()
	defer this.Unlock()
	delete(this.db, key)
}

func (this *SafeMap) Get(key string) (v interface{}) {
	this.Lock()
	defer this.Unlock()
	return this.db[key]
}
