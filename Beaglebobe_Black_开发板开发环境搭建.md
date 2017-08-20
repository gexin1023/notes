## Beaglebobe Black 开发板开发环境搭建

### 1.交叉编译环境搭建

在嵌入式开发中，首先需要在主机上安装交叉编译工具。开发人员一般是在主机上进行编程，将编译好的程序传输到嵌入式开发板运行，因此需要与嵌入式平台相对应的编译工具，才能使编译好的程序在目标板上运行。

我们使用Linaro GCC 作为交叉编译工具，它可以在[这里](http://releases.linaro.org/components/toolchain/binaries/latest/arm-linux-gnueabihf/)得到，这是已经编译好的二进制文件，解压后直接可以用。官方网站下载很慢，我直接下载ti官网下载了。

```
$ mkdir BBB     这里首先建立一个自己的目录
$ cd BBB		进入BBB目录
$ wget -c http://releases.linaro.org/components/toolchain/binaries/latest/arm-linux-gnueabihf/gcc-linaro-6.3.1-2017.05-x86_64_arm-linux-gnueabihf.tar.xz

$ tar xf gcc-linaro-6.3.1-2017.05-x86_64_arm-linux-gnueabihf.tar.xz	  解压文件
```

解压之后，会当前目录下出现一个文件夹gcc-linaro-x.x.x-2017.x-x86_64_arm-linux-gnueabihf，在这个文件件内的bin子文件夹中是编译好的二进制文件，如下所示，其中的arm-linux-gnueabihf-gcc就是我们需要的交叉编译器。

```
files in gcc-linaro-x.x.x-2017.x-x86_64_arm-linux-gnueabihf/bin：

arm-linux-gnueabihf-addr2line
arm-linux-gnueabihf-ar
arm-linux-gnueabihf-as
arm-linux-gnueabihf-c++
arm-linux-gnueabihf-c++filt
arm-linux-gnueabihf-cpp
arm-linux-gnueabihf-elfedit
arm-linux-gnueabihf-g++
arm-linux-gnueabihf-gcc
arm-linux-gnueabihf-gcc-5.4.1
arm-linux-gnueabihf-gcc-ar
arm-linux-gnueabihf-gcc-nm
arm-linux-gnueabihf-gcc-ranlib
arm-linux-gnueabihf-gcov
arm-linux-gnueabihf-gcov-dump
arm-linux-gnueabihf-gcov-tool
arm-linux-gnueabihf-gdb
arm-linux-gnueabihf-gfortran
arm-linux-gnueabihf-gprof
arm-linux-gnueabihf-ld
arm-linux-gnueabihf-ld.bfd
arm-linux-gnueabihf-nm
arm-linux-gnueabihf-objcopy
arm-linux-gnueabihf-objdump
arm-linux-gnueabihf-ranlib
arm-linux-gnueabihf-readelf
arm-linux-gnueabihf-size
arm-linux-gnueabihf-strings
arm-linux-gnueabihf-strip
gdbserver
runtest
```

在bin文件夹用 ./arm-linux-gnueabihf-gcc --version 命令查看交叉编译器的版本信息，我所用的交叉编译器版本信息的输出如下所示。

```
arm-linux-gnueabihf-gcc (Linaro GCC 5.4-2017.05) 5.4.1 20170404
Copyright (C) 2015 Free Software Foundation, Inc.
This is free software; see the source for copying conditions.  There is NO
warranty; not even for MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.

```

我们将gcc-linaro-x.x.x-2017.x-x86_64_arm-linux-gnueabihf/bin加入环境变量就可以在系统其他地方直接使用其中的可执行文件啦。在CentOS中通过下面的几个步骤将该目录加入到环境变量中。

- vim ~/.bash_profile  打开该文件

- 将 $HOME/BBB/gcc-linaro-x.x.x-2017.x-x86_64_arm-linux-gnueabihf/bin加入到PATH里面，如下所示。

-  reboot重启虚拟机来激活环境变量更改

```
# .bash_profile 文件
# Get the aliases and functions
if [ -f ~/.bashrc ]; then
	. ~/.bashrc
fi

# User specific environment and startup programs

PATH=$PATH:$HOME/.local/bin:$HOME/bin:$HOME/BBB/gcc-linaro-6.3.1-2017.05-x86_64_arm-linux-gnueabi/bin
export PATH

```

### 2. 编译内核树

为了我们所编写的驱动可以在目标板上顺利运行，还需要编译目标板内核树。首先先获取源码，Beagle Black 开发板的源码可以直接去ti官方获取am335x的源码，或者beaglebone通过git命令下载源码,w我是直接下载的ti源码。

```
$ tar  -xvf   ti_am335x_source.tar		// 解压源码文件
$ cd  linux								// 进入源码中的linux文件
$ make ARCH=arm CROSS_COMPILE=arm-linux-gnueabihf- tisdk_am335x-evm_deconfig
$ make ARCH=arm CROSS_COMPILE=arm-linux-gnueabihf-  zImage
$ make ARCH=arm CROSS_COMPILE=arm-linux-gnueabihf- am335x-boneblack.dtb

Kernel modules:
$ make ARCH=arm CROSS_COMPILE=arm-linux-gnueabi-  modules 
```

