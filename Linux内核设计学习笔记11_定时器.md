## 定时器与时间管理笔记

### 内核中的时间

  - 时钟中断：内核中的系统定时器以某种频率触发中断，该频率可以通过编程预定。
  - 节拍率HZ：时钟中断的频率称为节拍率。
  - 节拍：相邻两次中断的时间间隔称为节拍，1/节拍率。

### 节拍率HZ

系统定时器的节拍率是通过静态预处理定义的，也就是HZ值，在系统启动时按照HZ值对硬件进行设置。比如，x86系统的HZ默认值为100，即每10ms触发一次时钟中断。

### jiffies

该全局变量用来记录自系统启动以来产生的节拍数总数。在启动时，内核将该值初始化为0，每次时钟中断该值都会增加1。
	```
		time = jiffies /HZ;		//自启动以来经历的时间

	```
	
jiffies在系统中的定义如下，它是无符号长整数，因此在32位系统是32位，在64位系统上是64位。	

```
extern unsigned long jiffies;
extern  u64 jiffies_64;

```

系统还定义了一个变量jiffies_64，该变量在32位系统也是64位长，这样可以保证该变量不会溢出。在32位系统上，通过赋值将jiffies_64的低32位给jiffies，因为我们使用jiffies主要是来计算经过的相对时间，因此低32位足够使用了。在64位机上两个变量值是相同的。


当jiffies值超过他的最大存放范围之后，就会发生溢出。对于32位系统而言，当jiffies超过 $2^32-1$ 后，会再次回到0，这个过程称之为回绕。

我们在代码时需要注意jiffies是否会发生回绕，比如下面这个例子，当我们判断是否超时时，就需要考虑到回绕的情况。在下面的例子中，加入发生回绕，即使超时了，$timeout > jiffies$ 也是成立的。
```
unsigned long timeout = jiffies + HZ/2;//0.5秒之后的时刻

if(timeout > jiffies)
	printf("未超时")；
else printf("已超时")；
```

为了解决这种情况，内核提供了四个宏来比较节拍数，这是通过将 unsigned long 类型强制转换为 long 类型实现的。通过强制类型装换，即使发生回绕，也是从负数回绕到了正数，其大小关系不变。

```
#define time_after(unkown,kwon)		((long)(kown)-(long)(unkown))<0
#define time_before(unkown,kown)	((long)(unkown) -(long)(kown))<0
#deifne time_after_eq(unkown,kwon)		((long)(unkown) -(long)(kown))>=0
#define time_before_eq(unkown,kown)	 ((long)(kown)-(long)(unkown))>=0
```

### 硬时钟和定时器

系统中存在两种设备进行计时，分别是系统定时器和实时时钟。

#### 实时时钟

实时时钟（RTC）是用来持久存放系统时间的设备，即使系统关闭后，它依然可以靠着主板上的微型电池供电保持系统计时。

当系统启动时，内核通过读取RTC来初始化墙上时间，改时间存放在xtime变量中。内核通常不会在系统启动后再读取xtime变量，有些体系结构会周期性的将当前时间存回RTC中。实时时钟的主要作用就是在启动时初始化xtime。

#### 系统定时器

系统定时器提供一种周期性的触发中断机制，

### 时钟中断处理程序

时钟中断处理程序分为两个部分：体系结构相关部分和体系结构无关部分。

与体系相关的例程作为系统定时器的中断被注册到内核中，以便在产生时钟中断时能够相应运行。该部分一般包括以下内容：
	- 获得xtime_lock锁，一遍对jiffies_64和墙上时间xtime进行保护
	- 需要时应答或重新设置系统时钟
	- 周期性使用墙上时间更新实时时钟
	- 调用体系结构无关例程tick_periodic()

中断服务主要通过体系结构无关部分执行更多工作，tick_periodic()：
	- 给jiffies_64增加1
	- 更新资源的统计值
	- 执行一定到期的动态定时器
	- 更新墙上时间
	- 计算平均负载
	
### 定时器

定时器也称为动态定时器或内核定时器，是管理内核时间流逝的基础。

定时器结构如下：
```
struct timer_list{
	struct list_head entry;	//定时器链表入口
	unsigned ling expires;	// 以jiffies为单位的定时值
	void (*function) (unsigned ling);//定时器处理函数
	unsigned long data;			// 传给处理函数的长整型
	struct tvec_t_base_s *base;	//定时器内部值，用户不需要
}
```

创建定时器时，首先要定义
```
struct timer_list  my_timer;
```

接着初始化定时器结构
```
init_timer(&my_timer);
```

然后给定时器结构中的成员赋值
```
my_timer.expires = jiffies + delay;
my_timer.data = 0;
my_timer.function = my_function;
```
定时器处理函数的定义如下
```
void my_function(unsigned long data);
```

最后，还需要激活定时器
```
add_timer(&my_timer);
```

当节拍数大于等于指定的额超时时，内核就开始执行定时器处理函数。虽然内核可以保证在超时时间到达之前不会执行处理函数，但是会有延误。一般来说，定时器在超时后会马上执行，但也有可能推迟到下次节拍再运行。

有时候需要更改已经激活的定时器时间，可以通过内核函数mod_timer来实现
```
mod_timer(&my_timer, jiffies + new_delay);
```

mod_timer()函数也可以操作那些已经初始化，但还没有激活的定时器。若还没有激活，mod_timer()就会激活该函数。一旦从mod_timer()函数返回，定时器将被激活，并设置新的定时时间。

若需要在定时器超时之前停止计时器可通过del_timer()来实现
```
del_timer(&my_timer);
```

被激活的或未被激活的定时器都可以使用该函数，若未被激活则返回0；否则返回1.

当删除定时器时，必须注意一个潜在的竞争条件，当del_timer()返回后，可以保证的只是，定时器将来不会再被激活，但是在多处理器机器上定时器中断可能已经在其他处理器上运行了。所以删除定时器时需要等待可能在其他处理器上运行的定时器处理程序都结束，这时就要用del_timer_sync()
```
del_timer_sync(&my_timer);
```

### 延迟执行

内核代码（尤其是驱动程序代码）除了使用定时器或下半部意外还需要其他方法来退出任务。这种推迟常发生在等待硬件完成某项工作时，而且等待时间很短。

#### 忙等待

最简单的延迟方法是忙等待，实现方法如下：
```
unsigned long timeout = jiffies + 10;

while(time_before(jiffies,before);	// 忙等待10个节拍的时间

```

这种方法十分低效，尽量别用，丢人。。。

更好的方法是在等待的时候，运行内核重新调度执行其他任务：
```
unsigned long timeout = jiffies + 2*HZ;

while(time_before(jiffies,before)
	// 忙等待2秒的时间
	cond_resched();

```

cond_resched()函数将调度一个新的程序投入运行，但它只有在设置完need_resched标志后才能生效。也就是说，该方法有效的条件是系统中存在更重要的任务需要运行。注意，该方法需要调用调度程序，所以他不能再中断上下文中使用。事实上，所有的延迟方法在进程上下文中使用很好，而中断处理程序需要更快的执行（忙循环与这种目标相反）。延迟执行不管在那种情况下都不应该在持有锁时或者禁止中断时发生。

#### 短延迟

内核提供了三个可以处理短延迟的函数
```
void udelay(unsigned long usecs);	// 微秒
void ndelay(unsigned long nsecs);	// ns
void mdelay(unsigned long msecs);	// ms
```

#### schedule_timeout()

更理想的执行延迟的方法是使用schedule_timeout()函数，该方法让延迟执行的任务睡眠到指定的延迟时间耗尽后再重新运行。但该方法也不能保证睡眠时间正好等于指定的延迟时间，只能尽量使睡眠时间接近指定的延迟时间。当指定的时间到期后，内核唤醒被延迟的任务并将其重新投入运行队列。

```
set_current_state(TASK_INTERRUPTIBLE);

schedule_timeout(seconds*HZ)
```


