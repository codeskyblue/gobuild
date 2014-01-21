package utils

import (
	"errors"
	"testing"
	"time"
)

func TestGoTimeout(t *testing.T) {
	err := GoTimeout(func() error {
		return nil
	}, time.Second*1)
	if err != nil {
		t.Error(err)
	}
}

func TestGoTimeoutErr(t *testing.T) {
	var Err = errors.New("SampleErr")
	err := GoTimeout(func() error {
		return Err
	}, time.Second*1)
	if err != Err {
		t.Error(err)
	}
}

func TestGoTimeoutTimeout(t *testing.T) {
	err := GoTimeout(func() error {
		time.Sleep(1e8 * 2)
		return nil
	}, 1e8)
	if err != ErrTimeout {
		t.Error(err)
	}
}
