package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"

	"launchpad.net/goyaml"
)

type FileSet struct {
	Includes []string `yaml:"includes"`
	Excludes []string `yaml:"excludes"`
}

type Assembly struct {
	FileSet `yaml:"filesets"`
}

// upload a file and return a address
// FIXME: need to support qiniu
func uploadFile(reader io.Reader) (addr string, err error) {
	f, err := ioutil.TempFile("files", "upload-")
	if err != nil {
		return
	}
	_, err = io.Copy(f, reader)
	if err != nil {
		return
	}
	addr = "http://" + filepath.Join(opts.Hostname, f.Name())
	return
}

func achieveZip(target string, files ...string) (err error) {
	fmt.Println("target =", target)
	fmt.Println(files)
	return nil
}

// basic regrex match
func match(bre string, str string) bool {
	if bre == str { // FIXME: use re
		return true
	}
	return false
}

// package according .gobuild, return a download url
func Package(rc string) (addr string, err error) {
	data, err := ioutil.ReadFile(rc)
	if err != nil {
		return
	}
	fmt.Println(string(data))
	ass := new(Assembly)
	err = goyaml.Unmarshal(data, ass)
	if err != nil {
		return
	}
	fmt.Println(ass)
	return
}
