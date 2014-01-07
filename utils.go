package main

import (
	"errors"
	"fmt"
	"time"
)
import "log"

func Debugf(format string, a ...interface{}) {
	if false {
		log.Println("DEBUG: ", fmt.Sprintf(format, a...))
	}
}

func GoTimeout(f func() error, timeout time.Duration) (err error) {
	done := make(chan bool)
	go func() {
		err = f()
		done <- true
	}()
	select {
	case <-time.After(timeout):
		return errors.New("timeout")
	case <-done:
		return
	}
}
