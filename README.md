# gobuild

## 安装说明
#### 安装依赖

下载 [gox](https://github.com/mitchellh/gox) 

	go get github.com/mitchellh/gox
	
	# build toolchain
	gox -build-toolchain

#### 安装主程序
	go get github.com/shxsun/gobuild
	
	# $ cd到程序所在目录
	go build
	./gobuild --addr=localhost:3000 --cdn=http://gobuild.qiniudn.com/files
	
	