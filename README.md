# gobuild
[![Build Status](https://drone.io/github.com/shxsun/gobuild/status.png)](https://drone.io/github.com/shxsun/gobuild/latest)
[![Total views](https://sourcegraph.com/api/repos/github.com/shxsun/gobuild/counters/views.png)](https://sourcegraph.com/github.com/shxsun/gobuild)

在线服务的网站: <http://build.golangtc.com>

## 安装说明
#### 安装依赖

下载 [gox](https://github.com/mitchellh/gox) 

	go get github.com/mitchellh/gox
	
	# build toolchain
	gox -build-toolchain

#### 安装主程序
	go get github.com/shxsun/gobuild
	
	# $ cd到程序所在目录, 修改app.ini
	go build
	./gobuild
	
	
#### 使用到的资源
* <https://github.com/codegangsta/martini>
* [gox](https://github.com/mitchellh/gox) 
* ...

#### 已知的问题
golang的交叉编译不支持CGO

系统包中的os/user包中用到了CGO，所以CGO不支持，不过有解决办法，可以参考这个帖子，[使用环境变量获取用户名](http://stackoverflow.com/questions/7922270/obtain-users-home-directory)

#### Contributers
* [skyblue](https://github.com/shxsun)
* [Codefor](https://github.com/Codefor)
* ...
