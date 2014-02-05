#!/bin/bash -x
#
gopath=$(go env GOPATH)
gobin=${GOBIN:-"$gopath/bin"}
cd $gopath/src/github.com/shxsun/gobuild
git pull
go install
cd -
mv -v $gobin/gobuild ./gobuild

