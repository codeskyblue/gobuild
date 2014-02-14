package main

import (
	"io/ioutil"
	"path/filepath"
	"github.com/shxsun/go-sh"
	"github.com/shxsun/goyaml"
)
import beeutils "github.com/astaxie/beego/utils"

// download src
func (b *Builder) get() (err error) {
	exists := beeutils.FileExists(b.srcDir)
	if !exists {
		b.sh.Call("echo", []string{"downloading src"})
		err = b.sh.Call("go", []string{"get", "-v", "-d", b.project})
		if err != nil {
			return
		}
	}
	err = b.sh.Call("echo", []string{"fetch", b.ref}, sh.Dir(b.srcDir))
	if err != nil {
		return
	}
	if b.ref == "-" {
		b.ref = "master"
	}
	if err = b.sh.Call("git", []string{"fetch", "origin"}); err != nil {
		return
	}
	if err = b.sh.Call("git", []string{"checkout", "-q", b.ref}); err != nil {
		return
	}
	if err = b.sh.Call("git", []string{"merge", "origin/" + b.ref}); err != nil {
		return
	}
	r, err := sh.Capture("git", []string{"rev-parse", "HEAD"}, sh.Dir(b.srcDir))
	if err != nil {
		return
	}
	b.sha = r.Trim()

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
