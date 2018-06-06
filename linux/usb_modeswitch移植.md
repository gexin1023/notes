## usb_modeswitch移植

- 交叉工具链安装
- 交叉编译安装libsub库
- 交叉编译安装lib-compat-x.x.x
- 交叉编译安装usb_modeswitch

### 交叉编译工具链

为了使编译的程序可以在嵌入式开发板上运行，需要使用交叉编译链进行编译。交叉工具链可以直接下载二进制来用，下载[linaro-arm-linux-gnueabihf](https://releases.linaro.org/components/toolchain/binaries/latest-5/arm-linux-gnueabihf/)。交叉工具链放在“~/BBB/toolchain/gcc-linaro-6.2.1-2016.11-x86_64_arm-linux-gnueabihf”的地方，将arm-linux-gnueabihf-gcc所在目录加入环境变量，即可直接使用arm-linux-gnueabihf-gcc进行交叉编译。交叉编译链配置成功后，可以用 “arm-linux-gnueabihf-gcc -v” 命令查看交叉编译工具的信息。

```
$ arm-linux-gnueabihf-gcc -v

Using built-in specs.
COLLECT_GCC=arm-linux-gnueabihf-gcc
COLLECT_LTO_WRAPPER=/home/gexin/BBB/toolchain/gcc-linaro-6.2.1-2016.11-x86_64_arm-linux-gnueabihf/bin/../libexec/gcc/arm-linux-gnueabihf/6.2.1/lto-wrapper
Target: arm-linux-gnueabihf
Configured with: /home/tcwg-buildslave/workspace/tcwg-make-release/label/docker-trusty-amd64-tcwg-build/target/arm-linux-gnueabihf/snapshots/gcc-linaro-6.2-2016.11/configure SHELL=/bin/bash --with-mpc=/home/tcwg-buildslave/workspace/tcwg-make-release/label/docker-trusty-amd64-tcwg-build/target/arm-linux-gnueabihf/_build/builds/destdir/x86_64-unknown-linux-gnu --with-mpfr=/home/tcwg-buildslave/workspace/tcwg-make-release/label/docker-trusty-amd64-tcwg-build/target/arm-linux-gnueabihf/_build/builds/destdir/x86_64-unknown-linux-gnu --with-gmp=/home/tcwg-buildslave/workspace/tcwg-make-release/label/docker-trusty-amd64-tcwg-build/target/arm-linux-gnueabihf/_build/builds/destdir/x86_64-unknown-linux-gnu --with-gnu-as --with-gnu-ld --disable-libstdcxx-pch --disable-libmudflap --with-cloog=no --with-ppl=no --with-isl=no --disable-nls --enable-c99 --enable-gnu-indirect-function --disable-multilib --with-tune=cortex-a9 --with-arch=armv7-a --with-fpu=vfpv3-d16 --with-float=hard --with-mode=thumb --enable-multiarch --with-build-sysroot=/home/tcwg-buildslave/workspace/tcwg-make-release/label/docker-trusty-amd64-tcwg-build/target/arm-linux-gnueabihf/_build/sysroots/arm-linux-gnueabihf --enable-lto --enable-linker-build-id --enable-long-long --enable-shared --with-sysroot=/home/tcwg-buildslave/workspace/tcwg-make-release/label/docker-trusty-amd64-tcwg-build/target/arm-linux-gnueabihf/_build/builds/destdir/x86_64-unknown-linux-gnu/arm-linux-gnueabihf/libc --enable-languages=c,c++,fortran,lto --enable-checking=release --disable-bootstrap --build=x86_64-unknown-linux-gnu --host=x86_64-unknown-linux-gnu --target=arm-linux-gnueabihf --prefix=/home/tcwg-buildslave/workspace/tcwg-make-release/label/docker-trusty-amd64-tcwg-build/target/arm-linux-gnueabihf/_build/builds/destdir/x86_64-unknown-linux-gnu
Thread model: posix
gcc version 6.2.1 20161016 (Linaro GCC 6.2-2016.11) 

```

### 下载软件源码

1. mkdir USB-4G && cd USB-4G 创建文件夹并进入

1. 下载[libusb-1.0.21](https://sourceforge.net/projects/libusb/files/libusb-1.0/libusb-1.0.21/) ,并解压 tar jvxf libusb-1.0.21.tar.bz2

2. 下载[libusb-compat-0.1.5](https://sourceforge.net/projects/libusb/files/libusb-compat-0.1/libusb-compat-0.1.5/),并解压tar jvxf libusb-compat-0.1.5.tar.bz2

3. 下载[usb_modeswitch-2.5.1](http://www.draisberghof.de/usb_modeswitch/#download), 并解压tar jvxf usb-modeswitch-2.5.1.tar.bz2


### libusb交叉编译

1. 进入解压的包， cd ./libusb-1.0.21

2. 配置Makefile文件，使用./configure工具，可以使用./configure -h 查看该工具支持的选项。--prefix选项表示安装工具的位置，将工具都安装在“/home/username/USB-4G/install”的位置。

```shell
./configure  \
	--build = arm-linux \
	--host  = arm-linux-gnueabihf  \
	--prefix=/home/username/USB-4G/install \
	--disable-shared  \
	--enable-static  \
	--disable-udev
```

3. 编译 make

4. 安装 make install，安装完毕后在目录/home/username/USB-4G/install中可以看到多了lib和include两个文件夹。

### libusb-compat

1. 进入解压目录， cd libusb-compat-0.1.5

2. 配置Makefile文件，./configure工具。libusb-compat的安装依赖于libusb的库，而我们是将libusb安装在了自定义位置，因此在configure之前需要设置两个环境变量。



```
$ export LIBUSB_1_0_LIBS=/home/gexin/USB-4G/install/lib
$ export LIBUSB_1_0_CFLAGS = /home/gexin/USB-4G/install/

./configure  \
	--build=arm-linux \
	--host=arm-linux-gnueabihf \
	--prefix=/home/gexin/USB-4G/install \
	--disable-shared  \
	--enable-static --disable-udev \	
```

3. make

4. make install

### usb-modeswitch

1. 进入解压目录，cd usb-modeswitch-2.5.1

2. 该文件夹没有提供，./configure工具，因此需要我们自己手动对Makefile文件进行修改。以下几项需要修改。

```
CC		   = arm-linux-gnueabihf-gcc  使用交叉编译工具
INCLUDEDIR = /home/gexin/USB-4G/install/include  包含目录
LDFLAGS    = /home/gexin/USB-4G/install/libc	 库连接目录

下面这一行是编译命令，添加-I、-L选项
$(CC) -o $(PROG) $(OBJS) $(CFLAGS) $(LIBS) -I $(INCLUDEDIR) -L $(LDFLAGS)  -static  -pthread

```

3. 下面可以通过make编译文件，但是在编译之前，需要修改一个环境变量。

```
$ export PKG_CONFIG_PATH=/home/gexin/USB-4G/install/lib/pkgconfig	//	修改环境变量

$ make	// 编译

arm-linux-gnueabihf-gcc -o usb_modeswitch usb_modeswitch.c -Wall `pkg-config --libs --cflags libusb-1.0` -I /home/gexin/USB-4G/install/include -L /home/gexin/USB-4G/install/lib  -static  -pthread
sed 's_!/usr/bin/tclsh_!'"/usr/bin/tclsh"'_' < usb_modeswitch.tcl > usb_modeswitch_dispatcher

$ sudo make install	// 安装

```

安装之后，在当前文件夹下的usb-modeswitch可执行文件可以移植到目标板上的。




