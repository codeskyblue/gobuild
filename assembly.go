package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Unknwon/cae/zip"
	"github.com/shxsun/gobuild/utils"
	"github.com/shxsun/goyaml"
)

// package according .gobuild, return a download url
// format: <tgz|zip>
/*
var defaultRc = `---
filesets:
    includes:
        - static
        - README.*
        - LICENSE
    excludes:
        - .svn
`
*/

type FileSet struct {
	Includes []string `yaml:"includes"`
	Excludes []string `yaml:"excludes"`
}

type Assembly struct {
	FileSet `yaml:"filesets"`
}

// basic regrex match
func match(bre string, str string) bool {
	if bre == str { // FIXME: use re
		return true
	}
	return false
}

func pkgZip(root string, files []string) (path string, err error) {
	tmpFile, err := utils.TempFile("files", "tmp-", "-"+filepath.Base(root)+".zip")
	if err != nil {
		return
	}

	z, err := zip.Create(tmpFile)
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
	return tmpFile, nil

}

func Package(bins []string, rcfile string) (path string, err error) {
	lg.Debug(bins)
	lg.Debug(rcfile)
	data, err := ioutil.ReadFile(rcfile)
	if err != nil {
		lg.Error(err)
		lg.Debug("use default rc")
		data, err = ioutil.ReadFile("public/gobuildrc")
		if err != nil {
			lg.Error(err)
		}
		////data = []byte(defaultRc)
	}
	ass := new(Assembly)
	err = goyaml.Unmarshal(data, ass)
	if err != nil {
		return
	}
	dir := filepath.Dir(rcfile)
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
