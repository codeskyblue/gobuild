package main

import (
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
