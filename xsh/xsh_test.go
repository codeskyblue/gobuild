package xsh

import "testing"

func TestCall(t *testing.T) {
	err := Call("echo", []string{"a", "b"})
	if err != nil {
		t.Error(err)
	}
}

func TestSession(t *testing.T) {
	session := NewSession("pwd")
	err := session.Call()
	if err != nil {
		t.Error(err)
	}
}
