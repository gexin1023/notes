## Linux内核同步笔记

### 几个基本概念

	- 临界区（critical region）：访问和操作共享数据的代码段；
	- 原子操作：操作在执行中不被打断，要么不执行，要么执行完；
	- 竞争条件： 两个线程处于同一个临界区内执行，对数据同时访问或操作，称之为竞争；
	- 同步（synchronization）：避免并发和防止竞争条件成为同步。
	
### 预防死锁

	- 按顺序加锁，使用嵌套锁时，必须注意按顺序加锁，可以防止拥抱类死锁。
	- 防止饥饿
	- 不要重复请求同一个锁
	- 设计力求简单。
	
### 原子操作

原子操作可以保证执行过程不被打断，内核提供了两种原子操作接口：一组针对整数进行操作，另一组针对单独的位进行操作。

#### 原子整数操作

针对整数的操作只能对atomic_t类型的数据操作，atomic_t定义如下所示。
```
typedef struct{
	volatile int counter;
} atomic_t;
```

	尽管linux支持32整数，但是atomic_t类型只能当做24位来用，这是因为在SPARC体系结构上，原子操作的实现不同于其他系统，在int中的八位引入了一个锁来避免数据并发访问。

```
atomic_t v;						//	定义一个原子整数 v
atomic_t u = ATOMIC_INIT(0); 	//定义并初始化u

atomic_set(&v, 4);				//设置，赋值
atomic_add(2, &v);				//加
atomic_inc(&v);					//自增

atomic_read(&v);				//返回int型数据
```

原子操作通常是内联函数，往往是通过内嵌汇编指令来实现的。在编写代码时，能使用原子操作时，就尽量不要使用复杂的加锁机制。

#### 64位整数的原子操作

64位整数操作时，只需要讲atomic_前缀相关的类型及操作函数修改为 atomic64_前缀就可以了。

```
typedef struct{
	volatile long counter;
} atomic64_t;

atomic64_t v;						//	定义一个原子整数 v
atomic64_t u = ATOMIC_INIT(0); 	//定义并初始化u

atomic64_set(&v, 4);				//设置，赋值
atomic64_add(2, &v);				//加
atomic64_inc(&v);					//自增

atomic64_read(&v);				//返回int型数据
```

#### 原子位操作

位操作函数是针对普通地址进行操作的，无需特殊的原子类型，它的参数是一个指针和一个位号。32位系统上位号是0~31，64位系统上位号是0~63。

```
usigned long word = 0;

set_bit(0, &word);		//第0位被设置
set_bit(1, &word);		//第1位被设置

printk("%ul", &word);	//此处将打印3

clear_bit(1,&word);		//第1位被清除
change_bit(0, &word);	//第0位被翻转
```

### 自旋锁

自旋锁最多被一个可执行线程持有，如果一个线程试图获得一个已经被持有的自旋锁（即所谓的争用），那么该线程就会一直进行忙循环——旋转——等待锁重新可用。一个被争用的自旋锁使得请求它的线程在等待锁重新可用时自旋（自旋特别浪费处理器时间）。
	
	需要注意的是，自旋锁不应被长时间持有，自旋锁适用于短时间内进行轻量级加锁。

#### 自旋锁方法

```
DEFINE_SPIN;OCK(mr_lock);
spin_lock(&mr_lock);

/* 临界区
 * 操作共享数据的代码放于此处
 * 以避免对共享数据的操作并发
 */
 
spin_unlock(&mr_lock);
```

	需要注意，在单处理器机器上，编译的时候并不会加入自旋锁，它仅仅被当做一个设置内核抢占机制是否被启用的开关。如果禁止内核抢占，那么在编译时自旋锁会被完全踢出内核。
	
自旋锁可以在中断中使用，在中断中使用自旋锁时，要在使用之前首先禁止本地中断。否则，中断会打断正在持有锁的内核代码，有可能回去争用这个已经被持有的自旋锁，从而造成死锁。内核提供了同时禁止中断和获取锁的函数：
```
DEFINE_SPIN(mr_lock);
unsigned long flags;

spin_lock_irqsave(&mr_lock, flags);

//临界区（critical

spin_unlock_irqrestore(&mr_lock, flags);
```

	要对数据加锁，而不是对代码加锁。加锁保护的是临界区内的数据，而非代码。
	
除了上述静态方法添加自旋锁，还可以动态的添加。可以用spin_lock_init()函数动态创建自旋锁。

```
spin_lock();			//	获取指定的自旋锁
spin_lock_irq();		// 禁止本地中断，并获取自旋锁
spin_lock_irqsave();	// 获取本地中断状态，禁止本地中断，并获取自旋锁

spin_unlock();			// 释放指定的锁
spin_unlock_irq();		// 释放指定的锁，并打开中断
spin_unlock_irqrestore();// 释放指定锁，并将中断恢复的原有状态

spin_lock_init();		//	动态初始化指定的spinlock_t
spin_trylock();			// 试图获取指定的锁，若未获取就返回非0
spin_is_locked();		// 如果指定的锁正咋被获取，则返回非0，否则返回0
```

### 读写自旋锁

读取共享数据时，不会对数据造成改变，因此可以多个线程同时对数据进行读取。但是，写数据与读数据是不能同时的。

```
DEFINE_RWLOCK(mr_rwlock);

read_lock(&mr_rwlock);
// 临界区(只读)
read_unlock(&mr_rwlock);

write_lock(&mr_lock);
// 临界区(读写)，只能被一个线程获取
write_unlock(&mr_lock);

```

通常情况下，读锁和写锁处于完全分割的代码分支中。

```
read_lock();		// 获得指定的读锁
read_lock_irq();	// 禁止本地中断，并获取读锁
read_lock_irqsave();// 保存本地中断，禁止本地中断，获取读锁

read_unlock();		// 释放指定读锁
read_unlock_irq();	// 激活本地中断，并释放读锁
read_lock_irqstore();// 恢复中断到原有状态，并释放读锁

write_lock();		// 获得指定的写锁
write_lock_irq();	// 禁止本地中断，并获取写锁
write_lock_irqsave();// 保存本地中断，禁止本地中断，获取写锁

write_unlock();		// 释放指定写锁
write_unlock_irq();	// 激活本地中断，并释放写锁
write_lock_irqstore();// 恢复中断到原有状态，并释放写锁

write_trylock();	// 试图获取写锁，写锁不成功则返回非0值
rwlock_init();		// 初始化指定的rwlock
```
	
	需要注意的是，读写自旋锁照顾读锁更多一点，当读锁被占用时，写操作处于等待状态。但是，读锁却可以继续占用锁，大量的读锁被挂起，会导致写锁处于饥饿状态。

### 信号量

信号量是一种睡眠锁，如果有一个任务试图获得一个不可用的信号量时，信号量会将其推进一个等待队列。

	- 由于争优信号量的进程在等待时会睡眠，所以信号量适用于锁会被长时间锁定的情况。
	- 锁被短时间持有的情况不适合使用信号量，因为睡眠、维护等待队列以及唤醒所花费的时间可能比锁占用的时间还要长
	- 由于执行线程在锁被争用时会睡眠，所以只能在进程上下文中才能获取信号量锁，因为在中断上下文中是不可睡眠的。
	- 可以在持有锁时去睡眠，其他进程试图获取该锁时，并不会死锁，而是去睡眠了。
	- 占用信号量时不可以同时占用自旋锁，因为在等待信号量时有可能睡眠，而持有自旋锁时是不允许睡眠的。
	
#### 计数信号量和二值信号量

信号量可以同时允许任意数量的锁持有者，而自旋锁在一个时刻最多允许一个任务持有它，通过一个计数count来表示。只允许一个持有者的信号量称为二值信号量（也成为互斥信号量），其count为1。

Linux通过down()操作来请求获得一个信号量，down()对信号量计数减1，若结果大于等于0则持有该信号量，若小于0则线程被放入等待队列。临界区内操作完成之后，通过up()来释放信号量，信号量计数加1。

#### 创建和初始化信号量

```
struct semaphore name;		// 定义信号量
sema_init(&name, count);	// 初始化， count表示信号量的使用数量

```

创建互斥信号量可以用以下更简洁的方式
```
static DECLEAR_MUTEX(name);
```

更常见的情况是，信号量作为一个大数据结构动态创建。此时，只有指向该动态创建的信号量的简介指针，可以使用如下函数来对他进行初始化：
```
sama_init(sem, count);
```

动态初始化可以通过如下函数：
```
init_MUTEX(sem);
```

#### 使用信号量

通过函数**down_interruptible()**获取指定信号量，如果信号量不可用，就将调用进程设置成TASK_INTERRUPTIBLE状态进入睡眠。

使用down_trylock()函数可以尝试以堵塞的方式来获取指定的信号量。在信号已经被占领时，它返回非0值；否则返回0，并让你成功持有信号量锁。

```
static DECLEAR_MUTEX(mr_sem);	//定义并声明一个信号量锁

if(down_interruptible(&mr_sem)){
	
	//信号量未获取
}

/*	临界区... */

up(&mr_sem);	//释放给定的信号量

```

```
//信号量主要方法

sema_init(struct semaphore *, int);		// 以指定的计数值初始化动态创建信号量
init_MUTEX(struct semaphore *);			// 以计数值1初始化动态创建信号量
down_interruptible(struct semaphore *)  // 试图获取指定信号量，若信号被占用，则进入中断休眠状态
down(struct semaphore*)					// 试图获取指定信号量，若信号被占用，则进入不可中断睡眠状态
down_trylock(struct semaphore*)			// 试图获取指定信号量，若信号被争用，则立刻返回非0值
up(struct semaphore*)					// 释放信号量，如果睡眠队列不空，则唤醒其中一个任务	
```

### 读写信号量

与读写自旋锁类似，将信号量更具体为读写信号量。所有的读写信号量都是互斥信号量，他们只针对写操作互斥，不针对读者。也就是说，只要没有写锁定，并发的读锁数量不限。只要有写锁，就不可以有其他读锁或者写锁。

```
// 静态创建
static DECLEAR_RWSEM(name)

// 动态创建
init_rwsem(struct rw_semaphore * sem)

// 使用例子
static DECLEAR_RWSEM(mr_rwsem)；

down_read(&mr_rwsem);	//获取读信号量锁

// 临界区（只读）

up_read(&mr_rwsem);		//释放读信号量锁

down_write(&mr_rwsem);	// 获取写信号量锁

// 临界区（写）

up_write(&mr_rwsem);	// 释放写信号量锁

```

### 互斥体（mutex）

一种互斥的睡眠锁，其操作与计数为1的信号量相似，其接口更简单。

```
DEFINE_MUTEX(name);		// 静态初始化

mutex_init(&mutex);		// 动态初始化

mutex_lock(&mutex);		// 获取锁
	/* 临界区  */
mutex_unlock(&mutex);	// 释放锁

```

	- 任何时刻只有一个任务可以持有mutex
	- 给mutex上锁者必须负责给其解锁
	- 在同一个上下文中上锁和解锁
	- 当持有一个mutex时，进程不可以退出。
	- mutex不能在中断或者下半部中使用，即使是mutex_trylock()也不行
	- mutex只能通过官方API管理

	信号量和互斥体二者很相似，在使用时要优先使用mutex，只有在很特殊的场合才会使用信号量（一般在底层）。

### 完成变量（completion variable）

如果在内核中一个任务需要发出信号通知另一个任务发生了某个特定事件，可以利用完成变量使两个任务得以同步。

```
DECLEAR_COMPLETION(mr_comp);	// 静态初始化
init_completion(&mr_comp);		//动态创建

wait_for_completion(&mr_comp);	//等待某完成变量接受信号
complete(&mr_comp);				//发信号唤醒任何等待的任务

```

### 顺序锁

顺序锁的实现是通过一个序列计数器实现的，当有疑义的数据被写入之后，会得到一个锁，并且计数值增加。在读取数据前后，序列号都会被读取。如果读取的序列号值相同，说明在读操作过程中没有被写操作打断过。此外，如果序列数是偶数说明没有写操作发生，因为写锁会使值变成基数，读操作后值恢复到偶数。

```
seqlock_t mr_seq_lock = DEFINE_SEQLOCK(mr_seq_lock);

write_seqlock(&mr_seq_lock);
//写锁被读取
write_sequnlock(&mr_seq_lock);

// 写锁与自旋锁类似，差异在于读的时候
unsigned long seq;

do{
	seq = read_seqbegin(&mr_seq_lock);
	//开始读数据。。
}while(read_seqretry(&mr_seq_lock, seq));

```

### 禁止抢占

由于内核是抢占性的，内核的进程在任何时候都可以停下来执行更改优先权的进程，这意味着一个任务与被强占的任务可能在同一个临界区内运行。为了避免这种情况，内核抢占代码使用自旋锁作为非抢占区域的标志。如果一个自旋锁被持有，则内核不能进行抢占。

可以通过preempt_disable()来禁止内核抢占。

```
preempt_disable();
//内核禁止被抢占
preempt_enable();

```

### 顺序和屏障


