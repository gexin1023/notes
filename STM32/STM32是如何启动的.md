## STM32单片机是如何启动的？

### STM32中的内存

STM32中的内存包含两块主要区域：flash memory（只读）、static ram memory（SRAM，读写）。其中，flash memory 起始于0x08000000，SRAM起始于0x20000000。flash memory的第一部分存放异常向量表，表中包含了指向各种异常处理程序的指针。比如说，RESET Handler便位于0x08000004的位置，在处理器上电或重启时执行。在0x08000000处存放的是内部栈指针。

STM32的存储器映射如下图所示：
![存储器映射](./pic/mmap.bmp)

程序执行时，机器代码位于flash区域，变量和运行时栈等易变的内容位于SRAM中。

### startup.s

下面看一下start.s的代码，了解下如何定义不同类型的代码。

```
Stack_Size		EQU     0x400			;定义一个变量Stack_Size，相当于 Stack_Size = 0x400

                AREA    STACK, NOINIT, READWRITE, ALIGN=3	;定义一个segment 命名为 STACK
Stack_Mem       SPACE   Stack_Size		;连续0x400个字节清零
__initial_sp


; <h> Heap Configuration
;   <o>  Heap Size (in Bytes) <0x0-0xFFFFFFFF:8>
; </h>

Heap_Size      EQU     0x200

                AREA    HEAP, NOINIT, READWRITE, ALIGN=3
__heap_base
Heap_Mem        SPACE   Heap_Size
__heap_limit

```
这段代码中主要是定义了两个段（segment），这两个段都是可读写的，涉及到两个汇编指令AREA和SPACE。

- AREA
	语法： 
	AREA sectionname{,attr}{,attr}...

	where: sectionname is the name to give to the section. Sections are independent, named, indivisible chunks of code or data that are manipulated by the linker.
	
	当遇到下一个AREA时，表示该段结束。或者是，碰到END也表示该段结束。

	AREA属性：

	NOINIT		该数据段无须初始化
	READWRITE 	可读写
	DATA 		数据而非指令，默认是可读写的
	ALIGN 		对齐


- SPACE
	语法：
	{label} SPACE expr
	The SPACE directive reserves a zeroed block of memory. 
	保留了一段零初始化的内存

紧接着，定义了一个RESET段，该段只读的数据段，该段主要包含异常向量表。异常向量表的每一个元素都是一个函数地址，CDC表示一个字长的整形数据。向量的第一个元素是栈顶地址，第二个元素是Reset_Handler。
	
```

                PRESERVE8
                THUMB


; Vector Table Mapped to Address 0 at Reset
                AREA    RESET, DATA, READONLY
                EXPORT  __Vectors
                EXPORT  __Vectors_End
                EXPORT  __Vectors_Size

__Vectors       DCD     __initial_sp               ; Top of Stack
                DCD     Reset_Handler              ; Reset Handler
                DCD     NMI_Handler                ; NMI Handler

				/******后面代码省略********/

```

定义完向量表之后，又定义了.text段，也就是存放程序代码段，该段也是只读的。

该段定义了向量表中的各个处理程序，每个程序以PROC开始，以ENDP结束。第一个是Reset_Handler处理函数，单片机器动时便是从这里开始执行的。我们可以看到，除了ResetHandler其他的函数都只有一个 "B ."这是一个空的跳转，相当于进了死循环，所以需要在外部定义相应的处理函数。

Reset_Handler函数首先执行函数SystemInit，完成硬件初始化工作，然后执行__main建立C运行环境并从中调到用户定义的main()函数执行。

```
                AREA    |.text|, CODE, READONLY

; Reset handler
Reset_Handler    PROC
                 EXPORT  Reset_Handler             [WEAK]
        IMPORT  SystemInit
        IMPORT  __main

                 LDR     R0, =SystemInit
                 BLX     R0
                 LDR     R0, =__main
                 BX      R0
                 ENDP

; Dummy Exception Handlers (infinite loops which can be modified)

NMI_Handler     PROC
                EXPORT  NMI_Handler                [WEAK]
                B       .
                ENDP

```

### 下载程序到单片机

startup.s汇编程序经过汇编器编译后，产生目标代码，然后通过链接器将各个子程序链接为可执行代码。在链接之前，目标代码中的地址都是相对地址，只有链接之后才能转变为可执行的目标代码。在链接过程中，会确定每一部分代码的地址。这个过程都被IDE封装起来了，所以用户看不到。

链接过程中，不同的段的地址是不一样的，比如可读写的段必须放在SRAM对应的地址中（0x20000000开始），只读的段放到flash中（0x08000000开始）。我们可以在keil开发环境的Linker选项面板中看到读写和只读存放的地址，如下图所示。
![](./pic/linker.bmp)

下载到单片机时，也是分别制定了RAM和flash的地址，下图所示。

![](./pic/download.bmp)
