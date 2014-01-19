// broadcast_test.go
package main

import (
	"io"
	"io/ioutil"
	"testing"
	"time"
)

func TestBroadCast(t *testing.T) {
	broadcast := NewWriteBroadcaster()
	_, r1 := broadcast.NewReader("test")
	go func() {
		//io.Copy(os.Stdout, r1)
		r1.Close()
	}()
	broadcast.Write([]byte("abc"))
	time.Sleep(time.Millisecond * 500)
	bufstr, r2 := broadcast.NewReader("last reader")
	if string(bufstr) != "abc" {
		t.Errorf("expect []byte(abc), but got: []byte(%s)", string(bufstr))
	}
	go io.Copy(ioutil.Discard, r2)
	broadcast.Write([]byte("ABC"))
	broadcast.CloseWriters()
}
