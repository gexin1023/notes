# ubuntu下Nodic开发环境搭建

## 1.编译环境

ubuntu可直接装gcc编译环境
```
sudo apt install gcc-arm-none-eabi
```

也可以下载可执行文件[download](https://launchpad.net/gcc-arm-embedded/+download)

## 2. 检查make工具是否安装

```
make -v
```
一般来说开发linux上的开发者都会安装，若没有安装的话，执行以下命令安装。

```
sudo apt-get install build-essential checkinstall
```

## 3. nRF5 SDK下载

Nordic官方对nRF51、nRF52系列提供了SDK，SDK中提供了大量的BLE和ANT示例。可以在[这里](https://developer.nordicsemi.com/nRF5_SDK/nRF5_SDK_v12.x.x/)下载。我这边下在的是12.3.0版本的SDK。

下载之后解压到自己习惯的目录中，然后修改Makefile.posix文件。

```
vim  <SDK_PATH>/components/toolchain/gcc/Makefile.posix
```

文件内容修改如下：

```
GNU_INSTALL_ROOT := /usr/
GNU_VERSION := 5.4.1
GNU_PREFIX := arm-none-eabi
```
需要注意的是，GNU_INSTALL_ROOT选项的目录指的是gcc-arm-none-eabi的安装位置(bin文件夹所在的目录)，我直接用apt install安装的，所以直接用了"/usr/"。


## 4. 编译一个示例

在SDK的目录下打开一个示例文件夹

```
cd nRF5_SDK_12.3.0/examples/peripheral/led_softblink/pca10040/blank/armgcc
```

上面命令是打开一个led闪烁的例子，pca10040是我板子的版本.

在该目录下运行"make"，就会编译文件并生成二进制文件。

```
$ make

mkdir _build
Compiling file: nrf_log_backend_serial.c
Compiling file: nrf_log_frontend.c
Compiling file: app_error.c
Compiling file: app_error_weak.c
Compiling file: app_timer.c
Compiling file: app_util_platform.c
Compiling file: led_softblink.c
Compiling file: low_power_pwm.c
Compiling file: nrf_assert.c
Compiling file: sdk_errors.c
Compiling file: boards.c
Compiling file: nrf_drv_clock.c
Compiling file: nrf_drv_common.c
Compiling file: nrf_drv_uart.c
Compiling file: nrf_nvic.c
Compiling file: nrf_soc.c
Compiling file: main.c
Compiling file: RTT_Syscalls_GCC.c
Compiling file: SEGGER_RTT.c
Compiling file: SEGGER_RTT_printf.c
Assembling file: gcc_startup_nrf52.S
Compiling file: system_nrf52.c
Linking target: _build/nrf52832_xxaa.out

   text    data     bss     dec     hex filename
      7944      116     480    8540    215c _build/nrf52832_xxaa.out

      Preparing: _build/nrf52832_xxaa.hex
      Preparing: _build/nrf52832_xxaa.bin
```

如果执行make后输出跟上面一样，说明交叉编译gcc已经正确配置。下面就可以将文件烧录到板子中啦。

## 5. Jlink驱动工具

下载程序需要Jlink驱动工具，因此要先行安装。可以去[这里](https://www.segger.com/downloads/jlink)下载J-link软件，并安装。对于ubuntu系统，可以直接下载deb安装包进行安装。

## 6. nrfjprog工具下载

这是Nordic提供的命令行固件烧录工具，既有windows版本也有linux版本。在[这里](https://www.nordicsemi.com/eng/nordic/Products/nRF51-DK/nRF5x-Tools-Linux/51392)下载。

下载之后解压的到自己习惯的目录，然后将nrfjprog可执行文件所在路径添加到PATH路径中。之后输入"nrfjprog -v"查看是否配置正确。

```
$ nrfjprog -v
nrfjprog version: 9.7.2
JLinkARM.dll version: 6.22d
```

## 7. 下载程序到板子

进入到这一步的话，需要的工具都已安装好，可以烧写程序进板子观察现象啦。

```
$ nrfjprog --family nRF52 -e
Erasing code and UICR flash areas.
Applying system reset.

$ nrfjprog --family nRF52 --program _build/nrf52832_xxaa.hex 
Parsing hex file.
Reading flash area to program to guarantee it is erased.
Checking that the area to write is not protected.
Programing device.

$ nrfjprog --family nRF52 -r
Applying system reset.
Run.
```

以上三个命令执行完毕，会看到板子上的四个led灯闪烁，说明烧写成功。现在整个编译烧写的流程也就走通啦。


