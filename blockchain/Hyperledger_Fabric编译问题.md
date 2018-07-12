# Hyperledger Fabric(v1.1.0)编译时遇到的问题

##  0. 编译过程的坑

编译时，按照如下顺序编译

+ 1. make release，编译源码生成二进制文件
+ 2. make docker，生成一系列的docker镜像

第一步没什么问题，第二部时会遇到一些问题，比如：

+ `go get`不能获取golang.org包的问题
+ docker中ubuntu官方镜像源不能访问或者太慢问题

## 1. `go get`不能获取golang.org包

在编译过程中，生成docker镜像时需要go相关的工具，github上的可以正常，但是golang.org的包由于墙的原因不能获取。尝试过设置全局代理、设置git代理，不过试了半天都没成功。

我的解决方法是:

+ 1. 首先在github上获取包的源码
+ 2. 然后只编译gotools，`make gotools`
+ 3. 将生成的可执行文件复制到`fabric/build/docker/gotools/bin`目录下

在fabric/gotools/Makefile文件中，可以看到编译时需要的golang依赖包。
```go
go.fqp.govendor  := github.com/kardianos/govendor
go.fqp.golint    := golang.org/x/lint/golint
go.fqp.goimports := golang.org/x/tools/cmd/goimports
go.fqp.ginkgo    := github.com/onsi/ginkgo/ginkgo
go.fqp.gocov     := github.com/axw/gocov/...
go.fqp.misspell  := github.com/client9/misspell/cmd/misspell
go.fqp.gocov-xml:= github.com/AlekSi/gocov-xml
go.fqp.manifest-tool := github.com/estesp/manifest-tool
```
### 1.1 github上获取包的源码

在目录`fabric/build/gopath/src`下分别建立两个目录`github.com`和`golang.org`。

进入`golang.org`目录获取包源码

```go
mkdir x && cd x
git clone https://github.com/golang/tools.git
git clone https://github.com/golang/lint.git
```

进入`github.com`目录获取包源码

```go
git clone https://github.com/kardianos/govendor.git
git clone https://github.com/onsi/ginkgo.git
git clone https://github.com/axw/gocov.git
git clone https://github.com/client9/misspell.git
git clone https://github.com/AlekSi/gocov-xml.git
git clone https://github.com/estesp/manifest-tool.git
```

### 1.2 编译gotools并复制可执行文件

`make gotools`编译go的工具。然后将将生成的可执行文件复制到`fabric/build/docker/gotools/bin`目录下。


## 2. docker官方镜像源修改

解决完上一个问题后，会卡在生成testenv的docker地方，原因是原有docker中的ubuntu镜像源是Ubuntu官方的，速度很慢或者根本不能访问，直接替换为阿里云的镜像源就好了。

首先定位生成docker的目录位置，是在`fabric/build/testenv/`目录下。该目录下的Dockerfile文件描述了该docker要做的东西。我们需要在这个文件中修改镜像源。

在文件中加入这三行:

```
# add by gexin, change the sources.list
COPY payload/sources.list  /etc/apt/
RUN sudo apt-get update
```
修改后的文件内容如下：
```
# Copyright Greg Haskins All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#
FROM hyperledger/fabric-buildenv:x86_64-1.1.1-snapshot-feed00ad6

# fabric configuration locations
ENV FABRIC_CFG_PATH /etc/hyperledger/fabric

# add by gexin, change the sources.list
COPY payload/sources.list  /etc/apt/
RUN sudo apt-get update

# create needed directories
RUN mkdir -p \
  $FABRIC_CFG_PATH \
  /var/hyperledger/production

# fabric configuration files
ADD payload/sampleconfig.tar.bz2 $FABRIC_CFG_PATH

# fabric binaries
COPY payload/orderer /usr/local/bin
COPY payload/peer /usr/local/bin

# softhsm2
COPY payload/install-softhsm2.sh /tmp
RUN bash /tmp/install-softhsm2.sh && rm -f install-softhsm2.sh

# typically, this is mapped to a developer's dev environment
WORKDIR /opt/gopath/src/github.com/hyperledger/fabric
LABEL org.hyperledger.fabric.version=1.1.1-snapshot-feed00ad6 \
      org.hyperledger.fabric.base.version=0.4.6

```

加入的三行就是我们用 自己的sources.list文件替换原有的文件。需要注意的是，这时我们还没有需要替换的文件呢。

在payload目录下，新建一个sources.list文件，在文件内容为阿里云的镜像源，如下所示:

```
deb http://mirrors.aliyun.com/ubuntu/ xenial main restricted universe multiverse
deb http://mirrors.aliyun.com/ubuntu/ xenial-updates main restricted universe multiverse
deb http://mirrors.aliyun.com/ubuntu/ xenial-backports main restricted universe multiverse
deb http://mirrors.aliyun.com/ubuntu/ xenial-security main restricted universe multiverse
deb http://mirrors.aliyun.com/ubuntu/ xenial-proposed main restricted universe multiverse
```

## 总结

先是golang包的问题折腾了一下午，解决之后又遇到了镜像源的问题。

对golang以及docker的理解还不深入，需要继续在解决问题的过程中提高多这两个部分的认知。






