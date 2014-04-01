package main

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"github.com/shxsun/go-sh"
	"github.com/shxsun/goyaml"
)
import beeutils "github.com/astaxie/beego/utils"

// download src
func (b *Builder) get() (err error) {
	exists := beeutils.FileExists(b.srcDir)
	b.sh.Command("go", "version").Run()
	if !exists {
		err = b.sh.Command("go", "get", "-v", "-d", b.project).Run()
		if err != nil {
			return
		}
	}
	b.sh.SetDir(b.srcDir)
	if b.ref == "-" {
		b.ref = "master"
	}
	if err = b.sh.Command("git", "fetch", "origin").Run(); err != nil {
		return
	}
	if err = b.sh.Command("git", "checkout", "-q", b.ref).Run(); err != nil {
		return
	}
	//if err = b.sh.Command("git", "merge", "origin/"+b.ref).Run(); err != nil {
	//	return
	//}
	out, err := sh.Command("git", "rev-parse", "HEAD", sh.Dir(b.srcDir)).Output()
	if err != nil {
		return
	}
	b.sha = strings.TrimSpace(string(out))

	// parse .gobuild
	b.rc = new(Assembly)
	rcfile := "public/gobuildrc"
	if b.sh.Test("f", ".gobuild") {
		rcfile = filepath.Join(b.srcDir, ".gobuild")
	}
	data, err := ioutil.ReadFile(rcfile)
	if err != nil {
		return
	}
	err = goyaml.Unmarshal(data, b.rc)
	return
}
