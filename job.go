package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	beeutils "github.com/astaxie/beego/utils"
	"github.com/shxsun/go-sh"
	"github.com/shxsun/go-uuid"
	"github.com/shxsun/gobuild/models"
	"github.com/shxsun/gobuild/utils"
)

var history = make(map[string]string)

type Job struct {
	wbc     *utils.WriteBroadcaster
	cmd     *exec.Cmd
	sh      *sh.Session
	project string //
	ref     string
	os      string
	arch    string

	gopath string    // init
	gobin  string    // init
	srcDir string    // init
	sha    string    // get
	rc     *Assembly // get

	pid int64  // db
	tag string // tag = pid + str(os-arch), set after pid set
}

func NewJob(project, ref string, os, arch string, wbc *utils.WriteBroadcaster) *Job {
	b := &Job{
		wbc:     wbc,
		sh:      sh.NewSession(),
		project: project,
		ref:     ref,
		os:      os,
		arch:    arch,
	}
	if wbc != nil {
		b.sh.Stdout = wbc
		b.sh.Stderr = wbc
	}
	env := map[string]string{
		"PATH":    "/bin:/usr/bin:/usr/local/bin",
		"PROJECT": project,
		"GOROOT":  opts.GOROOT,
	}
	// enable cgo on current os-arch
	if os == runtime.GOOS && arch == runtime.GOARCH {
		env["CGO_ENABLED"] = "1"
	}

	b.sh.Env = env
	return b
}

// prepare environ
func (b *Job) init() (err error) {
	gobin, err := ioutil.TempDir("tmp", "gobin-")
	if err != nil {
		return
	}
	b.gobin, _ = filepath.Abs(gobin)
	b.gopath, _ = filepath.Abs("gopath")
	b.sh.Env["GOPATH"] = b.gopath
	b.sh.Env["GOBIN"] = b.gobin
	b.srcDir = filepath.Join(b.gopath, "src", b.project)
	return
}

// build src
func (j *Job) build(os, arch string) (file string, err error) {
	fmt.Println(j.sh.Env)
	j.sh.Env["GOOS"] = os
	j.sh.Env["GOARCH"] = arch

	err = j.sh.Call("go", []string{"get", "-u", "-v", "."})
	if err != nil {
		return
	}
	// find binary
	target := filepath.Base(j.project)
	if os == "windows" {
		target += ".exe"
	}
	return beeutils.SearchFile(target, j.gobin, filepath.Join(j.gobin, os+"_"+arch))
}

// achieve and upload
func (b *Job) publish(bins []string) (addr string, err error) {
	path, err := b.pack(bins, filepath.Join(b.srcDir, ".gobuild"))
	if err != nil {
		return
	}
	go func() {
		defer func() {
			lg.Debug("delete history:", b.tag)
			delete(history, b.tag)
			go func() {
				// leave 5min gap for unfinished downloading.
				time.Sleep(time.Minute * 5)
				//time.Sleep(time.Second * 5)
				os.Remove(path)
			}()
		}()
		// upload
		var cdnAddr string
		var err error
		if *environment == "development" {
			cdnAddr, err = UploadLocal(path)
		} else {
			cdnAddr, err = UploadFile(path, uuid.New()+"/"+filepath.Base(b.project)+".zip")
		}
		if err != nil {
			return
		}
		lg.Debug("upload ok:", cdnAddr)
		err = models.AddFile(b.pid, b.tag, cdnAddr, "output-")
		if err != nil {
			lg.Error(err)
		}
	}()
	tmpAddr := "http://" + opts.Hostname + "/" + path
	history[b.tag] = tmpAddr
	return tmpAddr, nil
}

// remove tmp file
func (b *Job) clean() (err error) {
	b.sh.Call("echo", []string{"cleaning..."})
	err = os.RemoveAll(b.gobin)
	return
}

// init + build + publish + clean
func (j *Job) Auto() (addr string, err error) {
	lock := utils.NewNameLock(j.project)
	lock.Lock()
	defer func() {
		lock.Unlock()
		if j.wbc != nil {
			j.wbc.CloseWriters()
		}
	}()
	if err = j.init(); err != nil {
		return
	}
	// defer clean should start when GOPATH success created
	defer func() {
		er := j.clean()
		if er != nil {
			lg.Warn(er)
		}
	}()
	// download src (in order to get sha)
	err = j.get()
	if err != nil {
		return
	}
	// search db for history data
	p, err := models.SearchProject(j.project, j.sha)
	if err != nil {
		pid, er := models.AddProject(j.project, j.ref, j.sha)
		if er != nil {
			err = er
			return
		}
		j.pid = pid // project id
	} else {
		j.pid = p.Id
	}
	// generate tag for builded-file search
	j.tag = fmt.Sprintf("%d-%s-%s", j.pid, j.os, j.arch)

	// search memory history
	hisAddr, ok := history[j.tag]
	if ok {
		return hisAddr, nil
	}
	// search database history
	f, err := models.SearchFile(j.pid, j.tag)
	lg.Debugf("search db: %v", f)
	if err == nil {
		addr = f.Addr
		return
	}
	// build xc
	j.sh.Call("echo", "building")
	file, err := j.build(j.os, j.arch)
	if err != nil {
		return
	}
	// package build file(include upload)
	addr, err = j.publish([]string{file})
	if err != nil {
		return
	}
	return
}
