package conf

import (
    "github.com/qiniu/rpc"
)

var UP_HOST  = "http://up.qiniu.com"
var RS_HOST  = "http://rs.qbox.me"
var RSF_HOST = "http://rsf.qbox.me"

var PUB_HOST = "http://pub.qbox.me"
var IO_HOST = "http://iovip.qbox.me"

var ACCESS_KEY string
var SECRET_KEY string

func SetUserAgent(userAgent string) {
    rpc.UserAgent = userAgent
}

func init() {
    SetUserAgent("qiniu go-sdk v6.0.0")
}
