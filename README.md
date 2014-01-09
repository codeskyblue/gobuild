# gobuild
[![Build Status](https://drone.io/github.com/shxsun/gobuild/status.png)](https://drone.io/github.com/shxsun/gobuild/latest)
[![Total views](https://sourcegraph.com/api/repos/github.com/shxsun/gobuild/counters/views.png)](https://sourcegraph.com/github.com/shxsun/gobuild)

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

#### Contributers
* [skyblue](https://github.com/shxsun)
* [Codefor](https://github.com/Codefor)
* ...
