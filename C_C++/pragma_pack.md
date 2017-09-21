## #pragma pack(n)对齐格式

#pragma pack(n) 是预处理器用来指定对齐格式的指令，表示n对齐。当元素字节小于n时，要扩展到n；若元素字节大于n则占用其实际大小。

```
struct tmp{
	int a;
	char b;
	int c;
};

int main()
{	
	struct tmp temp;

	printf("a:0x%x\n",&(temp.a));
	printf("b:0x%x\n",&(temp.b));
	printf("c:0x%x\n",&(temp.c));
	printf("size of temp is %d\n",sizeof(struct tmp));

	return 0;
}
```

对于tmp结构体，32位系统中默认情况下按照4字节对齐，该结构体占用12个字节，上面函数的输出如下：

```
a:0x28ff34
b:0x28ff38
c:0x28ff3c
size of temp is 12
```

当在结构体定义前加上#pragma pack(1)后，表示1字节对齐，也就是char类型数据不会扩展，该结构体占用9个字节，上述函数输出如下：

```
#pragma pack(1)
struct tmp{
	int a;
	char b;
	int c;
};

///////////输出如下////////////

a:0x28ff37
b:0x28ff3b
c:0x28ff3c
size of temp is 9
```

使用指令#pragma pack ()编译器将取消自定义的字节对齐方式，恢复到默认对齐方式，下面例子中，tmp将占用9个字节，而tmp1占用12个字节。
```
#pragma pack(1)
struct tmp{
	int a;
	char b;
	int c;
};

#pragma pack()

struct tmp1{
	int a;
	char b;
	int c;
};
```

另外，注意别的#pragma pack 的其他用法：
\#pragma pack(push) //保存当前对其方式到packing stack
\#pragma pack(push,n) 等效于
\#pragma pack(push)
\#pragma pack(n) //n=1,2,4,8,16 保存当前对齐方式，设置按n 字节对齐
\#pragma pack(pop) //packing stack 出栈，并将对其方式设置为出栈的对齐方
