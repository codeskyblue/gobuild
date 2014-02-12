package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	beeutils "github.com/astaxie/beego/utils"
	"github.com/shxsun/go-sh"
	"github.com/shxsun/go-uuid"
	"github.com/shxsun/gobuild/models"
	"github.com/shxsun/gobuild/utils"
)

var history = make(map[string]string)

type Builder struct {
	wbc     *utils.WriteBroadcaster
	cmd     *exec.Cmd
	sh      *sh.Session
	project string //
	ref     string
	os      string
	arch    string

	gopath    string    // init
	gobin     string    // init
	srcDir    string    // init
	sha       string    // get
	rc        *Assembly // get
	framework string    // build

	pid int64  // db
	tag string // tag = pid + str(os-arch), set after pid set
}

func NewBuilder(project, ref string, goos, arch string, wbc *utils.WriteBroadcaster) *Builder {
	b := &Builder{
		wbc:     wbc,
		sh:      sh.NewSession(),
		project: project,
		ref:     ref,
		os:      goos,
		arch:    arch,
	}
	if wbc != nil {
		b.sh.Stdout = wbc
		b.sh.Stderr = wbc
	}
	selfbin := beeutils.SelfDir() + "/bin"
	env := map[string]string{
		"PATH":    "/bin:/usr/bin:/usr/local/bin:" + selfbin,
		"PROJECT": project,
		"GOROOT":  opts.GOROOT,
	}
	// enable cgo on current os-arch
	if goos == runtime.GOOS && arch == runtime.GOARCH {
		env["CGO_ENABLED"] = "1"
	}

	b.sh.Env = env
	return b
}

// prepare environ
func (b *Builder) init() (err error) {
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
func (j *Builder) build(os, arch string) (file string, err error) {
	fmt.Println(j.sh.Env)
	j.sh.Env["GOOS"] = os
	j.sh.Env["GOARCH"] = arch

	// switch framework
	j.framework = j.rc.Framework
	switch j.rc.Framework {
	case "beego":
		err = j.sh.Set(sh.Dir(j.srcDir)).Call("bee", []string{"pack", "-f", "zip"})
		file = filepath.Join(j.srcDir, filepath.Base(j.project)) + ".zip"
		return
	case "revel":
		err = j.sh.Set(sh.Dir(j.srcDir)).Call("revel", []string{"package", j.project})
		file = filepath.Join(j.srcDir, filepath.Base(j.project)) + ".tar.gz"
		return
	default:
		j.framework = ""
	}

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
func (b *Builder) publish(file string) (addr string, err error) {
	var path string
	if b.framework == "" {
		path, err = b.pack([]string{file}, filepath.Join(b.srcDir, ".gobuild"))
	} else {
		path, err = utils.TempFile("files", "tmp-", "-"+filepath.Base(file))
		if err != nil {
			return
		}
		_, err = sh.Capture("mv", []string{"-v", file, path})
	}
	if err != nil {
		return
	}

	// file ext<zip|tar.gz>
	suffix := ".zip"
	if strings.HasSuffix(path, ".tar.gz") {
		suffix = ".tar.gz"
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
			cdnAddr, err = UploadFile(path, uuid.New()+"/"+filepath.Base(b.project)+suffix)
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
func (b *Builder) clean() (err error) {
	b.sh.Call("echo", []string{"cleaning..."})
	err = os.RemoveAll(b.gobin)
	return
}

// init + build + publish + clean
func (j *Builder) Auto() (addr string, err error) {
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
	// search db for history project record
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
	//}
	// package build file(include upload)
	addr, err = j.publish(file)
	if err != nil {
		return
	}
	return
}
