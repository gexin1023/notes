# 下半部笔记
------

## 1. 软中断
### 软中断实现
软中断是在编译期间静态分配，其结构如下所示，结构中包含一个接受该结构体指针作为参数的action函数。
```c
struct softirq_action{
    void (*action)(struct softirq_action *);
}
```
在kernel/softirq.c中定义了一个包含32个结构体的数组，每个数组成员都是一个被注册的软中断，数组如下所示：
```c
static struct softirq_action softirq_vec[NR_SOFTIRQS];
```
### 软中断处理程序
软中断处理函数action原型如下：
```c
void siftirq_handler(struct softirq_action *);
```
内核通过如下的方式调用软中断处理函数：
```c
my_softirq->action(my_softirq);
```
软中断不会抢占其他软中断，唯一可以抢占软中断的是中断处理程序。
### 软中断的执行
一个注册的软中断必须在被标记后才会被执行，软中断被唤起后，要在 do_softirq() 中执行，在do_softirq()函数中，遍历执行每一个被标记的软中断,如下所示：
```c
u32 pending;
//pending表示32位的标志，用来标记32个软中断，若位设置为1说明该位对应的软中断唤起。
pending = local_softirq_pending();

if(pending){
    struct softirq_action *h;
    
    set_softirq_pending(0);//重设置待处理的标志
    
    h = softirq_vec;
    do{
        if(pending&1)
            h->action(h);
        h++;
        pending>>1;
    }while(pending);
}
```
### 软中断的使用
#### 1）分配索引
在编译期间，通过在<linux/interrupt.h>中定义枚举类型来声明软中断，如下所示，其中软中断按照优先高低自上而下，新插入新的软中断时需要根据想要的优先级插入相应位置。
```c
enum
{
	HI_SOFTIRQ=0,       //优先级高的tasklet
	TIMER_SOFTIRQ,
	NET_TX_SOFTIRQ,
	NET_RX_SOFTIRQ,
	BLOCK_SOFTIRQ,
	BLOCK_IOPOLL_SOFTIRQ,
	TASKLET_SOFTIRQ,    //正常优先级的tasklet
	SCHED_SOFTIRQ,
	HRTIMER_SOFTIRQ,
	RCU_SOFTIRQ,	/* Preferable RCU should always be the last softirq */

	NR_SOFTIRQS
};
```
#### 2）注册处理程序

可以通过open_softirq(）函数注册软中断处理程序，两个参数：软中断索引号、处理函数。
```c
open_softirq(NET_RX_SOFTIRQ, net_tx_action);
```
软中断处理程序执行时候，允许相应中断，但不能自己休眠。
在一个处理器运行时候，当前处理器上软中断被禁止。

#### 3）触发中断
raise_softirq()函数可以实现软中断设置为挂起待执行，该函数在运行之前需要先禁止中断，触发后再恢复原来的状态。如果中断本来就已经被禁止，可以采用raise_softirq_irqoff()函数去唤醒中断。
```c
raise_softirq(NET_RX_SOFTIRQ)；//需要在使用前关中断，然后再恢复。

raise_softirq_irqoff(NET_RX_SOFTIRQ);//适用于中断本来就已经被禁止的情况

```

## 2.tasklet
        tasklet是在软中断基础上实现的，相当于对软中断中的HI_SOFTIRQ、TASKLET_SOFTIRQ的更改，将tasklet链表加入到以上两个软中断的处理函数中执行。
        通常情况下，我们使用tasklet而不是软中断，使用软中断的情况屈指可数。
### tasklet实现
#### tasklet结构
```cpp
struct tasklet_struct{
    struct tasklet_struct *next;    //链表中下一个tasklet
    unsigned long state;            //tasklet状态
    atomic_t count;                 //原子操作的计数器
    void (*func)(unsigned long);    //tasklet处理函数
    unsigned long data;             //给处理函数的参数
}
```
结构体重state成员，可以取0、TASKLET_STATE_RUN、TASKLET_STATE_SCHED。
    TASKLET_STATE_RUN->正在运行
    TASKLET_STATE_SCHED->已被调度

#### 调度
调度相当于将未调度的tasklet结构添加到两个链表结构：tasklet_vec(普通优先级)、tasklet_hi_vec(高优先级)。

```cpp
TASKLET_STATE_SCHED(); //tasklet调度函数

/*
 * 检查tasklet状态是否为TASKLET_STATE_SCHED，是的话已被调度，直接返回
 * 调用 __TASKLET_STATE_SCHED()函数
 * 保存中断状态，然后禁止中断状态
 * 将被调用的tasklet添加到tasklet链表
 * 唤醒软中断HI_SOFTIRQ或者TASKLET_SOFTIRQ
 * 恢复中断状态并返回
```

### 使用tasklet
#### 1) 声明自己的tasklet

```CPP
DECLEAR_TASKLET(name, func, data)   //声明后tasklet处于激活状态
DECLEAR_TASKLET_DISABLE(name, func, data)//声明后tasklet处于禁止状态
```

####2） 编写tasklet处理程序

```
void tasklet_handler(unsigned long data)
```
因为tasklet是靠软中断实现的，因此不能睡眠，也就是说在tasklet处理函数中不能使用信号量或者其他阻塞式函数。

#### 3)调度自己的tasklet
我们可以通过tasklet_schedule()函数并传递给他相应的tasklet指针来调度，如下所示：
```
tasklet_schedule(&mytasklet); //将tasklet指针传过去，来调度
```
**要注意：tasklet总在调度他的处理器上执行。**

## 工作队列
工作队列可以把工作推后，交由一个内个线程去执行，这个下半部分总是会在进程上下文中执行。

### 实现
工作队列最基本的是表现形式是把需要推后执行的任务交给特定的通用线程（工作队列也可以通过驱动程序创建工作者线程来处理推后工作，但是多数情况直接采用系统缺省的工作者线程来做推后工作）
####数据结构
1） 表示线程的数据结构
```
struct workqueue_struct{
    struct ypu_workqueue_struct cpu_wq[NR_CPUS];//数组每一项对应一个处理器
    struct list_head list;
    const char *name;
    int singlethread;
    int freezeable;
    int rt;
}
struct cpu_workqueue_struct{
    spinlock_t lock;// 锁来保护这种结构
    struct list_head worklist;//工作列表
    wait_queue_head_t more_work;
    struct workqueue_struct * wq;   //  关联工作队列结构
    task_t *thread;
    
}
```

注意：每一个工作者类型都关联一个自己的workqueue_struct.在该结构体里给每一个处理器（内的工作者线程）分配一个cpu_workqueue_struct。

2) 表示工作的数据结构

所有的工作者线程都是通过普通的内核线程实现的，他们都执行worker_thread()函数。在他们初始化完成以后，每个函数执行一个死循环并进入休眠，当有操作被传入队列里的时候，线程就会被唤醒，以执行这些操作。
    
  工作用work_struct结构体表示：
```
  struct work_struct{
        atomic_long_t data; //64位原子操作整数
        struct list_head enty;
        work_func_t func;
  }
```

这些结构体被连成链表，在每个处理器上的每种类型的队列都对应一个这样的链表。

###使用工作队列
####1）创建推后的工作
首先需要做的是创建一些需要推后完成的实际工作，通过宏DECLEAR_WORK在编译时静态创建结构体，如下所示：
```
DECLEAR_WORK(name; void(*func)(void*),void * data);
//这样会静态的创建一个名为name,处理函数为func，参数为data的结构体。
```
也可以在运行时通过指针创建一个工作，如下所示：
```
INIT_WORK(struct work_struct *work, void(*func), void *data);
//动态初始化一个由work指向的工作
```

####2）工作队列处理函数
```
void work_handler(void *data);//工作队列处理函数原型
```
这个函数会有工作者线程执行，因此函数运行于进程上下文中，默认情况下，允许相应中断，不能持有任和锁。
**需要注意的是，尽管操作函数运行于进程上下文中，但是他不能访问用户空间。**
####3）对工作进行调度
可以通过调用函数schedule_work() 把处理函数交给缺省的events工作线程
```
schedule_work(&work);
```
work马上就会被调度，一旦其所在的处理器上的工作者线程被唤醒，他就会被执行。
若不想work马上就工作，二十希望他进行一段延迟再执行，可以通过：
```
`schedule_delay_work(&work, delay);
//此时，直到delay的节拍时钟用完之后才会执行work
```
####4）刷新操作
进入队列的工作会在工作者线程的下一次被唤醒时候执行，在继续下一步工作之前，需要保证一些操作已经执行完毕。对于模块来说这一点很重要，在卸载之前，他可能需要调用以下函数。而在内核部分，为了防止竞争条件的出现，也可能需要确保不再持有处理工作。

出于以上目的，内核准备了一个用于刷新指定工作队列的函数
```
void flush_schedule_work(void);
```
函数会一直等待，直到队列中所有对象都被执行以后才会返回。
####5）创建新的工作队列
如果缺省的队列不能满足你的工作要求，需要创建新的工作队列与相应的工作者进程。由于这么做会在每个处理器上都创建一个工作者线程，所以只有在你明确了必须要自己创建一套线程来提高性能的情况下再创建自己的额工作队列。

这部分用的情况较少，需要的话再细看。
