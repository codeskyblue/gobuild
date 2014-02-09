package utils

import (
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func randString() string {
	return strconv.Itoa(rand.Int() % 100000)
}
func reseed() {
	seed := uint32(time.Now().UnixNano() + int64(os.Getpid()))
	rand.Seed(int64(seed))
}

func TempFile(dir, prefix, suffix string) (name string, err error) {
	if dir == "" {
		dir = os.TempDir()
	}

	var f *os.File
	nconflict := 0
	for i := 0; i < 10000; i++ {
		name = filepath.Join(dir, prefix+randString()+suffix)
		f, err = os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
		if os.IsExist(err) {
			if nconflict++; nconflict > 10 {
				reseed()
			}
			continue
		}
		f.Close()
		break
	}
	return
}
