## APUE学习笔记3_文件IO

Unix中的文件IO函数主要包括以下几个：open()、read()、write()、lseek()、close()等。这类I/O函数也被称为不带缓冲的I/O，标准I/O是带缓冲的I/O（当然，标准I/O也可以设置为不带缓冲）。

### 文件描述符

对于内核而言，所有打开的文件都通过文件描述符引用，比如read、write等操作都是通过文件描述符来实现的。文件描述符其实就是一个非负整数。当打开或者创建一个文件时，内核会向进程传递一个非负整数作为文件描述符，该文件描述符就可以作为参数传递给read、write等函数，进行文件操作。

UNIX系统中，通常是把0作为标准输入的描述符、1作为标准输出的描述符、2作为标准错误的描述符。

```
// <unistd.h> 

#define	STDIN_FILENO  0
#define STDOUT_FILENO 1
#define STDERR_FILENO 2

```

### 函数open()和openat()

```
#include <fcntl.h>

int open(const char *path, int oflag, ...);
int openat(int fd, const cahr *path, int oflag, ...);
```

使用open() 或者 openat() 函数打开或者创建文件。在函数open()中，path参数是文件的绝对路径或相对于当前文件的相对路径，oflag表示打开的方式。而在函数openat()中，path表示绝对路径（此时fd被忽略）或者是相对于fd的相对路径（fd 表示某已打开目录的文件描述符），openat()函数可以实现以相对路径来打开某些不便直接用相对路径表示的文件。

open和openat函数返回的文件描述符一定是最小的未使用描述符，比如一个应用程序可以先关闭标准输出（文件描述符是1）,然后打开另一文件，这样该文件会返回1作为其文件描述符。

```
int fd1 = open("test/test.c", O_RDONLY);	//只读方式打开文件test.c，此处为相对路径。

int fd2 = open("test", O_RDONLY | O_DIRECTORY);	//打开当前目录下的文件夹test
int fd3 = openat(fd2, "test.c", O_RDONLY); 	//打开相对于test文件夹下的test.c

```

#### 打开标志

文件打开标志有五个必选标志（必须指定一个且只能指定一个），与其他可选标志。

```
// 以下五个必须选一个且只能选一个

O_RDONLY // 只读打开
O_WRONLY // 只写打开
O_RDWR	 // 读写打开
O_EXEC   // 只执行打开
O_SEARCH // 只搜索打开（应用于目录）

// 以下为可选标志，

O_APPEND	// 每次写都追加到文件的结尾，文件以该标志打开时，如果使用lseek对文件重定位，若是读操作，重定位可以生效，若是写操作，重定位不生效，依然写在文件结尾。

O_CREAT 	// 若是文件爱你不存在则创建它，使用该标志时需要同时说明open第三个参数mode，即创建文件的权限。

O_TRUNC		// 如果此文件存在且以只读或读写打开，则将文件长度截为0.

O_DIRECTORY // 打开目录，若path不是目录则出错。

```
### 函数creat()

该函数只能以只写方式打开，我们可以直接使用open来实现创建新的文件，所以该函数就不用了。

### 函数close

```
#include <fcntl.h>

int close(int fd);
```

关闭一个文件时，还会释放加在该文件上的所有记录锁。

当一个进程结束时，内核会自动关闭它所有的打开文件，很多程序都用了这一点而不显示的调用close关闭打开文件。


### 函数lseek()

每个打开的文件都有一个与其相关的文件偏移量，表示从文件开始处到当前的字节数，通常是一个非负值。

```
#include <unistd.h>

off_t lseek(int fd, off_t offset, int whence);

// whence 可能取值如下
// SEEK_SET, 将文件的偏移量设置为距文件开始处offset个字节
// SEEK_CUR, 将文件的偏移量设置为距当前位置offset个字节（offset可正可负）
// SEEK_END, 将文件的偏移量设置为距结尾处offset个字节（offset可正可负）

```

通常，读写操作都是从当前文件偏移量处开始，并使偏移量增加所读写的字节数。根据系统默认情况，除非设置了O_APPEND，否则打开文件时，其偏移量被设置为0。


lseek若成功，则返回新的文件偏移量。文件的偏移量可以大于文件的长度，在这种情况下，对文件的下一次写操作将会增长改文件，并在文件中构成一个空洞。

文件空洞并不在磁盘占用存储区，读出的数据为0.



