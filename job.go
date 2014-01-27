package main

import (
	"io"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/shxsun/gobuild/utils"
	"github.com/shxsun/gobuild/xsh"
)

var GOPATH, GOBIN string

func init() {
	var err error
	GOPATH, err = filepath.Abs("project")
	if err != nil {
		lg.Fatal(err)
	}
	GOBIN, err = filepath.Abs("files")
	if err != nil {
		lg.Fatal(err)
	}
}

type Job struct {
	wbc     *utils.WriteBroadcaster
	cmd     *exec.Cmd
	sh      *xsh.Session
	project string
	sync.Mutex
}

func NewJob(addr string, wbc *utils.WriteBroadcaster) *Job {
	env := map[string]string{
		"PATH":   "/bin:/usr/bin:/usr/local/bin",
		"GOPATH": GOPATH,
		"GOBIN":  GOBIN,
	}
	b := &Job{
		wbc:     wbc,
		sh:      xsh.NewSession("./autobuild", []string{addr}, env),
		project: addr,
	}
	b.sh.Output = wbc
	return b
}

// parse .gobuild, prepare environ
func (j *Job) init() (err error) {
	//err = j.sh.Call("echo", []string{"xyz"})
	err = j.sh.Call("go", []string{"get", "-v", "-d", j.project})
	return
}

// build src
func (j *Job) build() error {
	return j.sh.Call("echo", []string{"1234"})
}

// achieve and upload
func (j *Job) pkg() {
}

// remove tmp file
func (j *Job) clean() {
	j.sh.Call("echo", []string{"cleaning..."})
}

// init + build + pkg + clean
func (j *Job) Auto() (err error) {
	if err = j.init(); err != nil {
		return
	}
	defer func() {
		j.clean()
		j.wbc.CloseWriters()
	}()
	if err = j.build(); err != nil {
		return
	}
	j.pkg()

	err = j.sh.Call("./autobuild", []string{j.project})
	if err != nil {
		utils.Debugf("start cmd error: %v", err)
		io.WriteString(j.wbc, "\nERROR: "+err.Error())
	}
	return
}
