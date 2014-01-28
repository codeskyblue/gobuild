package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Unknwon/cae/zip"
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
func uploadFile(file string) (addr string, err error) {
	f, err := ioutil.TempFile("files", "upload-")
	if err != nil {
		return
	}
	err = f.Close()
	if err != nil {
		return
	}
	exec.Command("cp", "-f", file, f.Name()).Run()
	addr = "http://" + filepath.Join(opts.Hostname, f.Name())
	return
}

// basic regrex match
func match(bre string, str string) bool {
	if bre == str { // FIXME: use re
		return true
	}
	return false
}

func pkgZip(root string, files []string) (addr string, err error) {
	tmpFile, err := ioutil.TempFile("files", "upload-")
	if err != nil {
		return
	}
	err = tmpFile.Close()
	if err != nil {
		return
	}
	defer os.Remove(tmpFile.Name())

	z, err := zip.Create(tmpFile.Name())
	if err != nil {
		return
	}
	for _, f := range files {
		var save string
		// binary file use abspath
		if strings.HasSuffix(root, f) {
			save = f[len(root):]
		} else {
			save = filepath.Base(f)
		}
		info, er := os.Stat(f)
		if er != nil {
			continue
		}
		lg.Debug("add", save, f)
		if info.IsDir() {
			if err = z.AddDir(save, f); err != nil {
				return
			}
		} else {
			if err = z.AddFile(save, f); err != nil {
				return
			}
		}
	}
	if err = z.Close(); err != nil {
		lg.Error(err)
		return
	}
	return uploadFile(tmpFile.Name())
}

// package according .gobuild, return a download url
// format: <tgz|zip>
var defaultRc = `---
filesets:
    includes:
        - static
        - README.*
        - LICENCE
    excludes:
        - .svn
`

func Package(bins []string, rcfile string) (addr string, err error) {
	lg.Debug(bins)
	data, err := ioutil.ReadFile(rcfile)
	if err != nil {
		lg.Debug("use default rc")
		data = []byte(defaultRc)
	}
	ass := new(Assembly)
	err = goyaml.Unmarshal(data, ass)
	if err != nil {
		return
	}
	dir := filepath.Dir(rcfile)
	//fmt.Println(dir, ass)
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}
	var includes = bins // this may change slice bins
	for _, f := range fs {
		var ok = false
		for _, patten := range ass.Includes {
			if match(patten, f.Name()) {
				ok = true
				break
			}
		}

		if ok {
			includes = append(includes, filepath.Join(dir, f.Name()))
		}
	}
	return pkgZip(dir, includes)
}
