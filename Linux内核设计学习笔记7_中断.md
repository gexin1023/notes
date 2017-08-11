## 中断与中断处理

### 何为中断？
+ 一种由设备发向处理器的电信号
+ 中断不与处理器时钟同步，随时可以发生，内核随时可能因为中断到来而被打断。
+ 每一个中断都有唯一一个数字标志，称之为中断线（IRQ）
+ 异常是由软件产生，与处理器时钟同步。

### 中断处理程序
+ 由内核调用来响应中断
+ 运行于中断上下文
+ 中断的执行不可阻塞
+ 中断处理分为两个部分，中断处理程序是上半部（top half），还有下半部（bottom halves）

#### 中断处理程序注册
+ 中断处理程序是管理硬件驱动程序的组成部分，如果设备使用中断，其相应的驱动程序就会注册一个中断处理程序。
+ 通过request_irq（）函数来注册中断处理程序
```	
int request_irq( unsigned irq,
		irq_handler_t handler,
		unsigned long flags,
		count char* name,
		void *dev)
```
+ 第一个参数irq表示要分配的中断号
+ 第二个参数handler表示中断处理程序指针
+ 第三个表示标志，可以为0、IRQF_DISABLE、IRQF_SAMPLE_RANDOM、IRQF_TIMER、IRQF_SHARED
	+ **IRQF_DISABLE** 表示该中断处理期间，禁用所有其他中断
	+ **IRQF_SAMPLE_RANDOM** 这个设备产生的中断对内核熵池有贡献
	+ **IRQF_TIMER** 为系统定时器中断而准备的
	+ **IRQF_SHARED** 表示多个中断处理程序共享中断线。
+ 第四个参数name表示设备的文本表示
+ 第五个参数dev用于共享中断线，dev提供唯一的标志信息。

  需要注意的是，request_irq( )可能睡眠，因此不能再中断上下文或者其他不允许阻塞的代码中调用该函数。
### 中断处理程序释放
  卸载驱动程序时，需要用**free_irq（）**注销相应的中断处理程序，并释放中断线。

```
  void free_irq(unsigned int irq, void *dev);
```
如果指定的中断线不是共享的，那么该函数删除处理程序的同时将禁用这条中断线。如果是共享的，只删除dev对应的中断处理程序。
### 编写中断处理程序
```
static irqreturn_t intr_handler(int irq, void * dev);
```
  当一个给定的中断处理程序正在执行时，相应的中断线在所有的处理器上都会被屏蔽掉，以防止在同一条中断线上接受另一个新的中断。

## 中断上下文
+ 当执行一个中断时，内核处于中断上下文。
+ 中断上下文没有后备进程，不可以睡眠。
+ 中断上下文有着严格的时间限制，因为其打断了其他代码（有可能打断了其他中断处理程序）。中断上下文中的 代码应该迅速简洁，尽量不要使用循环去处理繁重的工作。

### 中断控制
Linux内核提供了一组接口用于控制机器上的中断状态

+ 禁止和激活中断
用于禁止、激活当前处理器的本地中断，
```
local_irq_disable();
local_irqenable();
```
+ 禁止指定中断线
```
void disable_irq(unsigned int irq);			//禁止控制器上某一条中断线，函数只有在当前执行的所有处理程序完成后，才能返回 
void disable_irq_nosync(unsigned int irq);	//禁止控制器上某一条中断线，不会等待当前中断处理程序执行完毕。
void enable_irq(unsigned int irq);			//激活控制器上某一条中断线， 
void synchronize_irq(unsigned int irq);		//等待下一个特定的中断处理程序退出
```
在一条中断线上，每次调用disable_irq_nosync()、disable_irq()，都需要调用一次enable_irq()，只有在enable_irq()完成了最后一次调用后，才完成了中断线的激活。

+ 这三个函数可以从中断或进程上下文中调用，而且不会睡眠。





