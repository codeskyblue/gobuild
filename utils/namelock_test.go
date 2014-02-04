package utils

import (
	"testing"
	"time"
)

func TestNameLock(t *testing.T) {
	n1 := NewNameLock("hello")
	n2 := NewNameLock("hello")
	go func() {
		n1.Lock()
		time.Sleep(1e5)
		t.Log("first")
		n1.Unlock()
	}()
	time.Sleep(1e4)
	n2.Lock()
	t.Log("second")
	n2.Unlock()
}
