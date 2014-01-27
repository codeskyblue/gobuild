package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"sync"

	beeutils "github.com/astaxie/beego/utils"
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
	GOPATH  string
	project string
	srcDir  string
	sync.Mutex
}

func NewJob(p string, wbc *utils.WriteBroadcaster) *Job {
	env := map[string]string{
		"PATH":    "/bin:/usr/bin:/usr/local/bin",
		"GOPATH":  GOPATH,
		"PROJECT": p,
	}
	b := &Job{
		wbc:     wbc,
		sh:      xsh.NewSession(),
		project: p,
		GOPATH:  GOPATH,
		srcDir:  filepath.Join(GOPATH, "src", p),
	}
	b.sh.Output = wbc
	b.sh.Env = env
	return b
}

// parse .gobuild, prepare environ
func (j *Job) init() (err error) {
	err = j.sh.Call("go", []string{"get", "-v", "-d", j.project})
	return
}

// build src
func (j *Job) build(ref, os, arch string) (file string, err error) {
	srcDir := filepath.Join(j.GOPATH, "src", j.project)
	fmt.Println(j.sh.Env)
	j.sh.Env["GOOS"] = os
	j.sh.Env["GOARCH"] = arch

	// fetch branch
	if err = j.sh.Call("git", []string{"fetch", "origin"}, xsh.Dir(srcDir)); err != nil {
		return
	}
	if ref == "-" {
		ref = "master"
	}
	if err = j.sh.Call("git", []string{"checkout", ref}); err != nil {
		return
	}

	err = j.sh.Call("go", []string{"get", "-v", "."})
	if err != nil {
		return
	}
	// find binary
	target := filepath.Base(j.project)
	if os == "windows" {
		target += ".exe"
	}
	gobin := filepath.Join(j.GOPATH, "bin")
	return beeutils.SearchFile(target, gobin, filepath.Join(gobin, os+"_"+arch))
}

// achieve and upload
func (j *Job) pkg() error {
	//args := []string{"-os=linux windows darwin", "-arch=amd64 386"}
	//args = append(args, "-output="+"$CURDIR/files/$PROJECT/$SHA/{{.OS}}_{{.Arch}}/{{.Dir}}")
	//return j.sh.Call("gox", args)
	return nil
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
	file, err := j.build("-", "linux", "amd64")
	if err != nil {
		return
	}
	fmt.Println(file)
	return j.pkg()

	//err = j.sh.Call("./autobuild", []string{j.project}, xsh.Dir(""))
	//if err != nil {
	//	utils.Debugf("start cmd error: %v", err)
	//	io.WriteString(j.wbc, "\nERROR: "+err.Error())
	//}
	//return
}
