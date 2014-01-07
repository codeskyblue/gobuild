package main

import "fmt"
import "log"

func Debugf(format string, a ...interface{}) {
	if false {
		log.Println("DEBUG: ", fmt.Sprintf(format, a...))
	}
}
