package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	beeutils "github.com/astaxie/beego/utils"
	"github.com/qiniu/log"
	"github.com/shxsun/go-sh"
	"github.com/shxsun/go-uuid"
	"github.com/shxsun/gobuild/database"
	"github.com/shxsun/gobuild/utils"
)

var history = make(map[string]string)

type Builder struct {
	wbc      *utils.WriteBroadcaster
	sh       *sh.Session
	project  string //
	ref      string
	os       string
	arch     string
	fullname string // p + ref + os + arch

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
		wbc:      wbc,
		sh:       sh.NewSession(),
		project:  project,
		ref:      ref,
		os:       goos,
		arch:     arch,
		fullname: strings.Join([]string{project, ref, goos, arch}, "-"),
	}
	b.sh.ShowCMD = true
	if wbc != nil {
		b.sh.Stdout = wbc
		b.sh.Stderr = wbc
	}
	selfbin := beeutils.SelfDir() + "/bin"
	env := map[string]string{
		"PATH":    strings.Join([]string{"/bin:/usr/bin", selfbin, os.Getenv("PATH")}, ":"),
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
func (this *Builder) build(os, arch string) (file string, err error) {
	this.sh.Env["GOOS"] = os
	this.sh.Env["GOARCH"] = arch
	this.sh.SetDir(this.srcDir)

	// switch framework
	this.framework = this.rc.Framework
	switch this.rc.Framework {
	case "beego":
		err = this.sh.SetDir(this.srcDir).Call("bee", "pack", "-f", "zip")
		file = filepath.Join(this.srcDir, filepath.Base(this.project)) + ".zip"
		return
	case "revel":
		err = this.sh.SetDir(this.srcDir).Call("revel", "package", this.project)
		file = filepath.Join(this.srcDir, filepath.Base(this.project)) + ".tar.gz"
		return
	default:
		this.framework = ""
	}

	//if this.sh.Test("f", ".gopmfile") {
	//	this.sh.Alias("go", "gopm")
	//}
	/* // close godep now
	if this.sh.Test("d", "Godeps") {
		err = this.sh.Call("godep", "go", "install")
		return
	}
	*/

	err = this.sh.Call("go", "get", "-u", "-v", ".")
	if err != nil {
		return
	}
	// find binary
	target := filepath.Base(this.project)
	if os == "windows" {
		target += ".exe"
	}
	return beeutils.SearchFile(target, this.gobin, filepath.Join(this.gobin, os+"_"+arch))
}

// achieve and upload
func (b *Builder) publish(file string) (addr string, err error) {
	var path string
	if b.framework == "" {
		path, err = b.pack([]string{file}, filepath.Join(b.srcDir, ".gobuild.yml"))
	} else {
		path, err = utils.TempFile("files", "tmp-", "-"+filepath.Base(file))
		if err != nil {
			return
		}
		err = sh.Command("mv", "-v", file, path).Run()
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
			log.Debug("delete history:", b.tag)
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
			name := fmt.Sprintf("%s-%s-%s-%s",
				filepath.Base(b.project),
				b.os, b.arch, b.ref) + suffix
			cdnAddr, err = UploadFile(path, uuid.New()+"/"+name)
		}
		if err != nil {
			return
		}
		log.Debug("upload ok:", cdnAddr)
		output := ""
		if b.wbc != nil {
			output = string(b.wbc.Bytes())
		}
		err = database.AddFile(b.pid, b.tag, cdnAddr, output)
		if err != nil {
			log.Error(err)
		}
	}()
	tmpAddr := "http://" + opts.Hostname + "/" + path
	history[b.tag] = tmpAddr
	return tmpAddr, nil
}

// remove tmp file
func (b *Builder) clean() (err error) {
	b.sh.Call("echo", "cleaning...")
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
			log.Warn(er)
		}
		if err != nil && j.wbc != nil {
			io.WriteString(j.wbc, err.Error())
		}
		// FIXME: delete WriteBroadcaster after finish
		// if build error, output will not saved, it is not a good idea
		// better to change database func to -..
		//		SearchProject(project) AddProject(project)
		//		SearchFile(pid, ref, os, arch)
		//		AddFile(pid, ref, os, arch, sha)
	}()
	// download src (in order to get sha)
	err = j.get()
	if err != nil {
		return
	}
	// search db for history project record
	log.Info("request project:", j.project)
	log.Info("current sha:", j.sha)
	p, err := database.SearchProject(j.project, j.sha)
	if err != nil {
		log.Info("exists in db", j.project, j.ref, j.sha)
		pid, er := database.AddProject(j.project, j.ref, j.sha)
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
	f, err := database.SearchFile(j.pid, j.tag)
	log.Debugf("search db: %v", f)
	if err == nil {
		addr = f.Addr
		return
	}

	// build xc
	// file maybe empty
	file, err := j.build(j.os, j.arch)
	if err != nil {
		return
	}
	// package build file(include upload)
	addr, err = j.publish(file)
	if err != nil {
		return
	}
	return
}
