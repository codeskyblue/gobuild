## [gobuild.io](http://gobuild.io)
[![Build Status](https://drone.io/github.com/shxsun/gobuild/status.png)](https://drone.io/github.com/shxsun/gobuild/latest)
[![Go Walker](http://gowalker.org/api/v1/badge)](http://gowalker.org/github.com/shxsun/gobuild)

Go build + pacakge + distributions

There are a lot of golang open souce project, sometime we want to share code, sometimes we want to share binary file to friends.
But few website offers golang binary shares. So I created one.

### How to use
	wget gobuild.io/github.com/shxsun/fswatch/v1.0/linux/amd64 -O fswatch.zip

### .gobuild (optional)
use `.gobuild` file, you can use more function with gobuild.io.

first you need to add a file `.gobuild` into project root.

`.gobuild` is just a yaml file. specified with which file should be included and excluded.

for example. If I want to add static and LICENSE and exclude README.md. `.gobuild` can be write with

	filesets:
		includes:
			- static
			- LICENSE
		excludes:
			- README.md

binary file is defaulted added, you don't need to worry about it.

if no `.gobuild` file found in your project. A default `.gobuild` file will be used.

	# default
	filesets:
		includes:
			- static
			- README.*
			- LICENSE
		excludes:
			- .svn

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
