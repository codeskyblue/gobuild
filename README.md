## [gobuild.io](http://gobuild.io)
[![Build Status](https://drone.io/github.com/codeskyblue/gobuild/status.png)](https://drone.io/github.com/codeskyblue/gobuild/latest)
[![Go Walker](http://gowalker.org/api/v1/badge)](http://gowalker.org/github.com/codeskyblue/gobuild)
[![Gobuild Download](http://gobuild.io/badge/github.com/codeskyblue/gobuild/download.png)](http://gobuild.io/github.com/codeskyblue/gobuild)

# News <http://beta.gobuild.io>
[gobuild2](http://github.com/gobuild/gobuild2) beta was released.

QQ Group: [368327021](http://shang.qq.com/wpa/qunwpa?idkey=48933b5fc6e840d4b871feffee4fa05e08e23cd201f7b6c7944cad3faf375565)

*Thanks very much for you guys stars which encourage me to rewrite this website to gobuild2. Thanks very much. Thanks open source.*

Go build + package + distributions

There are a lot of golang open souce project, sometime we want to share code, sometimes we want to share binary file to friends.
But few website offers golang binary shares. So I created one.

### How to use
	wget gobuild.io/github.com/codeskyblue/fswatch/v1.0/linux/amd64 -O fswatch.zip

-- unfinished --
	wget gobuild.io/linux/amd64/v1.0/github.com/codeskyblue/fswatch/fswatch.zip

### .gobuild.yml
use `.gobuild.yml` file, you can use more function with <https://gobuild.io>.

first you need to add a file `.gobuild.yml` into project root.

For beego project: (platform will will invode `bee pack -f zip`)

	framework: beego
	
For revel project: (`revel package`)

	framework: revel

For self define which file should be packaged.(excludes is not working now).
And binary file is defaulted added, you don't need to worry about it.

	filesets:
		includes:
			- static
			- LICENSE
			- README.md
		excludes:
			- CHANGELOG

There is a default for every project: see [default gobuildrc](public/gobuildrc)

### other build tool support
support [gopm](http://gopm.io).

Test is `.gopmfile` exists in project root, then use alias go=gopm instead.
### add badge
[![Gobuild Download](http://gobuild.io/badge/github.com/codeskyblue/gobuild/download.png)](http://gobuild.io/github.com/codeskyblue/gobuild)

assume you project address is github.com/codeskyblue/gobuild

and the png address is: <http://gobuild.io/badge/github.com/codeskyblue/gobuild/download.png>

Markdown link is link below

	[![Gobuild Download](http://gobuild.io/badge/github.com/codeskyblue/gobuild/download.png)](http://gobuild.io/github.com/codeskyblue/gobuild)
-------------------
### For developers
#### Prepare dependencies
	go get -d github.com/codeskyblue/gobuild
	# cd github.com/codeskyblue/gobuild
	bin/install.sh
	# config file: config.yaml
	./gobuild

2 example project, which contains `.gobuild`

* github.com/codeskyblue/gobuild-beegotest
* github.com/codeskyblue/gobuild-reveltest

### related package
* xorm: <https://github.com/lunny/xorm>
* web framework: <https://github.com/codegangsta/martini>
* go-sh: <https://github.com/codeskyblue/go-sh>
* zip archive support: <https://github.com/Unknwon/cae>
* golang cross compile <https://github.com/mitchellh/gox>
* ...

### Q/A(knownen issues)
##### not support os/user
*golang's cross compile not support CGO, but package os/user use CGO.*

solutions: use environment variables to get use-name <http://stackoverflow.com/questions/7922270/obtain-users-home-directory>

### Contributers
* [codeskyblue](https://github.com/codeskyblue)
* [Codefor](https://github.com/Codefor)
* [fuxiaohei](https://github.com/fuxiaohei)
* ...

