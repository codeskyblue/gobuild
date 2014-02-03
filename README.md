## gobuild
[![Build Status](https://drone.io/github.com/shxsun/gobuild/status.png)](https://drone.io/github.com/shxsun/gobuild/latest)
[![Total views](https://sourcegraph.com/api/repos/github.com/shxsun/gobuild/counters/views.png)](https://sourcegraph.com/github.com/shxsun/gobuild)

Online website: <http://gobuild.io>

Go build + pacakge + distributions

There are a lot of golang open souce project, sometime we want to share code, sometimes we want to share binary file to friends.
But few website offers golang binary shares. So I created one.

### How to use
	-still developing wget gobuild.io/github.com/shxsun/fswatch/v1.0/linux/amd64 -O fswatch.zip-

### For developers
#### Prepare dependencies
	go get github.com/mitchellh/gox
	
	# build toolchain
	gox -build-toolchain

#### setup
	go get github.com/shxsun/gobuild
	
	# update config.yaml
	go build && ./gobuild
	

### related package
* <https://github.com/codegangsta/martini>
* <https://github.com/codegangsta/inject>
* [gox](https://github.com/mitchellh/gox) 
* ...

### Q/A(knownen issues)
##### not support os/user

	golang's cross compile not support CGO, but package os/user use CGO.
	sulutions: use environment variables to get use-name <http://stackoverflow.com/questions/7922270/obtain-users-home-directory>

### Contributers
* [skyblue](https://github.com/shxsun)
* [Codefor](https://github.com/Codefor)
* ...
