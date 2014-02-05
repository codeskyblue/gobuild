## [gobuild.io](http://gobuild.io)
[![Build Status](https://drone.io/github.com/shxsun/gobuild/status.png)](https://drone.io/github.com/shxsun/gobuild/latest)
[![Go Walker](http://gowalker.org/api/v1/badge)](http://gowalker.org/github.com/shxsun/gobuild)

Go build + pacakge + distributions

There are a lot of golang open souce project, sometime we want to share code, sometimes we want to share binary file to friends.
But few website offers golang binary shares. So I created one.

### How to use
	wget gobuild.io/github.com/shxsun/fswatch/v1.0/linux/amd64 -O fswatch.zip

### .gobuild
add a file `.gobuild` in the root of project. with content like.

	filesets:
		includes:
			- static
			- README.*
			- LICENSE
		excludes:
			- .svn

directory `static README.* LICENSE` will be packaged in <http://gobuild.io>

-------------------
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
* web framework: <https://github.com/codegangsta/martini>
* xsh use: <https://github.com/codegangsta/inject>
* zip archive support: <https://github.com/Unknwon/cae>
* golang cross compile <https://github.com/mitchellh/gox>
* ...

### Q/A(knownen issues)
##### not support os/user

	golang's cross compile not support CGO, but package os/user use CGO.
	sulutions: use environment variables to get use-name <http://stackoverflow.com/questions/7922270/obtain-users-home-directory>

### Contributers
* [skyblue](https://github.com/shxsun)
* [Codefor](https://github.com/Codefor)
* ...
