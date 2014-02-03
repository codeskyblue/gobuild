package main

import (
	"io/ioutil"
	"os/exec"
	"path/filepath"

	"github.com/qiniu/api/io"
	"github.com/qiniu/api/rs"
)

const SCOPE = "gobuild-io"

func UploadFile(localFile string, destName string) (addr string, err error) {
	policy := new(rs.PutPolicy)
	policy.Scope = "gobuild-io"
	uptoken := policy.Token(nil)

	var ret io.PutRet
	var extra = new(io.PutExtra)
	err = io.PutFile(nil, &ret, uptoken, destName, localFile, extra)
	addr = "http://" + SCOPE + ".qiniudn.com/" + destName
	return
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
