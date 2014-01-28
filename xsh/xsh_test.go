package xsh

import "testing"

/*
func TestCall(t *testing.T) {
	ret, err := Call("echo", "hello")
	if err != nil {
		t.Error(err)
	}
	t.Log(ret.Trim())
}
*/

func TestCapture(t *testing.T) {
	r, err := Capture("echo", []string{"hello"})
	if err != nil {
		t.Error(err)
	}
	_ = r
	if r.Trim() != "hello" {
		t.Errorf("expect hello, but got %s", r.Trim())
	}
}

func TestSession(t *testing.T) {
	session := NewSession("pwd")
	err := session.Call()
	if err != nil {
		t.Error(err)
	}
}
