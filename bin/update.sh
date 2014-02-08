#!/bin/bash -x
#
gopath=$(go env GOPATH)
gobin=${GOBIN:-"$gopath/bin"}
cd $gopath/src/github.com/shxsun/gobuild
git pull
go get -v
cd -
mv -v $gobin/gobuild ./gobuild

