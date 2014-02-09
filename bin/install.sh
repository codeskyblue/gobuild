#!/bin/bash -
#
WORKSPACE=$(cd $(dirname $0)/..; pwd)
export GOPATH=$WORKSPACE/gopath
export GOBIN=$WORKSPACE/bin
export PATH=$GOBIN:$PATH

go get -v github.com/mitchellh/gox

go get -v github.com/beego/bee
go get -v github.com/robfig/revel/revel

echo build toolchain
gox -build-toolchain

go build
