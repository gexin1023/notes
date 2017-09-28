## ARM架构中的Procedure Call Standard

### 几个名词
- ABI ：
1. 可执行文件必须遵守的规范，以在特定执行环境中运行；
2. 单独产生的可重定址的文件必须遵守的规范，以用来链接和执行。

- EABI: 
适用于嵌入式环境的ABI

- PCS：
程序调用规范（Procedure Call Standard）

- AAPCS：
PCS for ARM Architecture
AAPCS定义了单独编译、单独汇编的程序是如何一起工作的。

- Routine、subroutine 
控制可以进入的一段程序，调用之后，可以将控制返回给它的调用者。这里可分别理解为程序调用者、被调用者

- Procedure: 
A routine returns no result value.

- Function；
A routine returns a result value.

- Active stack、call-frame stack:
调用者栈帧


### 数据类型

#### 基础数据类型

- 整型
unsigned byte(8), signed byte(8), unsigned half-word(16), signed half-word(16), unsigned word(32), signed word(32), unsigned double-word(64), signed double-word(64)

- 浮点型
half precision(2), single precision(4), double precision(8)

- 容器向量
64-bit vector(8)，128-bit vector(16)

- 指针
数据指针(4)，指令指针(4)

#### 字节序

从软件视角看，内存是字节的阵列，每一个字节都是可寻址的。

- 小端字节序
数据在内存中，数据的最低字节放在内存中最低地址上。

- 大端字节序
数据在内存中，数据的最低字节放在内存中的最高地址上。

#### 复合类型

- an aggregate, 类似于C中的结构体, where the members are laid out sequentially in memmory

- a union, 枚举类型内的元素有相同的地址

- 数组， 相同类型数据的集合，连续地址存储

#### 数据对齐

- 数据自身对齐
比如，byte对齐为**1**个字节，word对齐为**4**个字节。如果数据的对齐值为N，则该数据的存放地址位于N的整数倍的位置，即“数据地址 % N == 0”。

- 结构体对齐值
结构体成员中最大的对齐值即为结构体对齐值。同样，该结构体存放的地址为对齐值N的整数倍。

- 在C中可以使用 #pragma pack(N) 来指定对齐值

下面以C语言为例，看一下结构体的对齐方式：

```
#include <stdio.h>
#include <stdlib.h>

typedef struct s{	
	int a;
	char b;
	long long c;
	short d
} s;

s t;

int main(void)
{
	char *c;
	t.a = 1;
	t.b = 1;
	t.c = 1;
	t.d = 1;
	
	printf("size of t is %d\n", sizeof(t));
	
	for(c = (char*)&t + sizeof(t) -1; c>=(char*)&t; c--)
	{
		printf("0x%x: |--%x--|\n",c,*c);
	}
	
	return 0;
}

```

![](./pic/align.jpeg)

上面程序输出如下图所示，可以看出这是小端字节序，即数据的最低字节存放在最低地址。b的存放位置位于0x40ea2c地址处，虽然只占一个字节。但是由于成员c是8字节对齐的，c的起始地址8位对齐，所以c应该位于0x40ea30，0x40ea2d到0x40ea2f这三个字节是为了对齐而扩充的。而该结构体的对齐值为成员的最大对齐值也为8，所以最后一个成员d，虽然只占2字节，但是需要扩充的8字节对齐。











