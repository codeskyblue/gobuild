package main

import (
	"io"
	"os/exec"
	"sync"

	"github.com/shxsun/gobuild/utils"
)

type Job struct {
	wbc *utils.WriteBroadcaster
	cmd *exec.Cmd
	sync.Mutex
}

func NewJob(addr string, wbc *utils.WriteBroadcaster) *Job {
	cmd := exec.Command("./autobuild", addr)
	cmd.Stdout = wbc
	cmd.Stderr = wbc
	return &Job{
		wbc: wbc,
		cmd: cmd,
	}
}

// parse .gobuild, prepare environ
func (j *Job) init() {
}

// build src
func (j *Job) build() {
}

// achieve and upload
func (j *Job) pkg() {
}

// remove tmp file
func (j *Job) clean() {
}

// init + build + pkg + clean
func (j *Job) Auto() {
	j.init()
	j.build()
	j.pkg()
	j.clean()

	err := j.cmd.Run()
	if err != nil {
		utils.Debugf("start cmd error: %v", err)
		io.WriteString(j.wbc, "\nERROR: "+err.Error())
	}
	j.wbc.CloseWriters()
}
