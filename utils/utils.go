package utils

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/shxsun/goyaml"
)

func Debugf(format string, a ...interface{}) {
	if false {
		log.Println("DEBUG: ", fmt.Sprintf(format, a...))
	}
}

var ErrTimeout = errors.New("timeout")

func GoTimeout(f func() error, timeout time.Duration) (err error) {
	done := make(chan bool)
	go func() {
		err = f()
		done <- true
	}()
	select {
	case <-time.After(timeout):
		return ErrTimeout
	case <-done:
		return
	}
}

func Dump(a interface{}) {
	out, _ := goyaml.Marshal(a)
	fmt.Println(string(out))
}
