#!/bin/bash -
#
WORKSPACE=$(cd $(dirname $0)/..; pwd)
export GOPATH=$WORKSPACE/gopath
export GOBIN=$WORKSPACE/bin

go get -v github.com/beego/bee
go get -v github.com/robfig/revel

