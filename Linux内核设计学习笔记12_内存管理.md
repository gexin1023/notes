## 内存管理学习笔记

### 页

页是内核管理内存的基本单位，内存管理单元（MMU，管理内存并把虚拟地址转化为物理地址的硬件）通常以页为单位进行处理，从虚拟内存的角度看，页就是最小单位。

```
struct page{
	unsigned long flags;
	atomic_t _count;
	atomic_t _mapcoount;
	unsigned long private;
	struct address_space *mapping;
	pgoff_t index;
	struct list_head lru;
	void *virtual
}
```

   - flag成员：页的状态，比如页是不是脏的、是不是被锁在内存中。flag的每一位单独表示一种状态，也就是说最少可以表示32中状态（这取决于flag有多少位）
   - _count成员：存放页的引用计数，内核通过page_count()函数对该成员进行检查，若函数返回0则表示页空闲，若返回一个正整数则表示页正咋被使用。
   - virtual成员：是页的虚拟地址，通常情况下他就是页在虚拟内存中的地址，有些内存（比如 High Memory）并不会永久的映射到内核地址空间上，这事virtual成员的值为NULL
   - 系统中每个页都被分配一个这样的结构体
   
### 区

由于硬件的限制，内核并不能对所有的页一视同仁，比如有的硬件只能在特定的地址上执行DMA操作，有的结构其物理地址寻址范围大于虚拟地址寻址范围，导致部分内存不能永久性映射到内核空间上。

Linux主要使用了四个区将页分为不同的类型

	- ZONE_DMA：这个区包含用来执行DMA操作的页
	- ZONE_DMA32：这个区与ZONE_DMA类似负责执行DMA操作，但是该区内的页只能被32位设备访问。
	- ZONE_NORMAL：可以正常映射的页
	- ZONE_HIGHMEM：这个区包含哪些不能被永久映射到内核地址空间的页。

分区是与体系结构相关的，有的体系结构中在所有的内存上执行DMA操作都没问题，那么ZONE_DMA就为空。同样，有的体系结构中不存在High Memory，即所有的内存都可以映射到内核空间上。

	- 在x86结构上，ISA设备不能在整个32位地址空间中执行DMA操作，ISA只能使用地址空间的前24位，即16M地址空间，所以ZONE_DMA在x86上包含的页都在0-16M的范围内。
	- 在32位X86系统上，虚拟地址分为两部分，低地址开始的3G（0x0000 0000-0xC000 000）属于用户空间，高地址的1G（0xC000 0000 - 0xFFFF FFFF）属于内核空间。
	- 对于内核空间来说，较低的896M（0xC000 0000 - 0xF7FF FFFF）可以直接映射到内核空间的物理地址上，剩下的128M（0xF800 0000 - 0xFFFF FFFF）根据需求映射成高地址。
	
某些分配必须从特定区中获取页，比如用于DMA的内存必须从ZONE_DMA区获取页，但是一般用途的内存可以从任何区内获取。需要注意的是，内存获取不能跨区进行，也就是说不能从两个区内获取页。

### 获取页

#### 获取页
```
struct page * alloc_pages(gfp_t gfp_mask, unsigned int order);	
// 返回page指针，该函数分配$2^order$个连续的物理页

void* page_address(struct page *page);
// 该函数返回当前所在的逻辑地址。

unsigned long __get_free_pages(gfp_t gfp_mask, unsigned int order);
// 该返回获取 $2^order$个页，并返回第一个页的逻辑地址

struct page * alloc_page(gfp_t gfp_mask);
// 分配一个页,返回page指针
unsigned long __get_free_page(gfp_t gfp_mask);
// 分配一个页，返回页的逻辑地址

unsigned long get_zeroed_page(unsigned int gfp_mask);
// 获取填充为0的页
```

#### 释放页
```
void __free_pages(struct page* page, unsigned int order);
void free_pages(unsigned long addr, unsigned int order);
void free_page(unsigned long addr);
```

```
unsigned long page;

page = __get_free_pages(GFP_KERNEL, 3);
if(!page){
	return -ENOMEM;
}

// ....
// ....

free_pages(page,3);
```

#### gfp_mask标志

分配器标志可以分为三类：行为修饰符、区修饰符、类型。用于指定获得内存时的方式，包括怎么获取，从哪里获取等行为。

最常用的标志是GFP_KERNEL，这种分配可能会阻塞，该标志只能用在可以重新安排调度的进程上下文中（未被锁持有）。

GFP_ATOMIC标志表示不能睡眠的内存分配，如果当前的代码不能睡眠（如中断、软中断、tasklet等）那么可以用该标志获取内存。

GFP_NOIO表示分配内存时不会启动磁盘IO来帮助满足请求。

GFP_NOFS 表示它可能会启动磁盘IO，但是不会启动文件系统IO。

	在绝大多数代码中用到的标志要么是GFP_KERNEL，要么是GFP_ATOMIC。
	- 进程上下文，可以睡眠 => GFP_KERNEL
	- 进程上下文，不可以睡眠 => GFP_ATOMIC
	- 中断、软中断、tasklet =>  GFP_ATOMIC
	- 用于DMA，可睡眠   =>  GFP_DMA|GFP_KERNEL
	- 用于DMA, 不可睡眠 => GFP_DMA|GFP_ATOMIC

### kmalloc() and vmalloc()

kmalloc()函数用来获得指定字节数的连续内存，其函数声明如下：

```
void * kmalloc(size_t size, gfp_t flags);
// 该函数返回一个指向内存块的指针，其内存块至少为size字节大小
// 新分配的内存区域在物理上是连续的。

void kfree(const void* ptr);
// 释放不需要的内存
```

vmolloc()函数类似于kmalloc()，不过该函数分配的内存在虚拟地址上是连续，而在物理地址上不连续。

```
void * vmalloc(unsigned long size);
// 分配size字节大小的内存，在逻辑上连续，物理地址不一定连续

void vfree(const void * addr);
// 释放不需要的虚拟内存
```

由molloc()分配的内存在进程的虚拟地址上也是连续的，但是不能保证其在物理地址上连续。

大多数情况下，只由硬件需要得到物理地址连续的内存，因为硬件设备一般存在于内存管理单元之外，它根本不知道啥是虚拟内存。对于内核而言，所有的内存看起来都是逻辑上连续的。

尽管某些情况下才需要物理上连续的内存，但是内核一般是用kmalloc()函数的，因为vmalloc()为了把不连续的物理地址映射为连续的虚拟地址，还需要做额外的工作。只有在获得大块内存时才会用vmalloc()函数来分配。


### slab层

为了便于数据频繁的分配和回收，编程人员常会用到空闲链表作为缓存，空闲链表中包含已经分配好的数据结构块，可以直接获得，节省了分配内存的步骤。

Linux内核提供了slab层，slab分配器扮演了通用数据结构缓存层的角色。

每种类型的对象都对应一个高速缓存，每个高速缓存被分为不同的slab，slab由一个或多个物理上连续的页组成。一般情况下，每个slab由一个页组成。每个slab都包含一些对象成员（被缓存的数据结构）。每个slab处于三种状态的一种：满，空，部分。一个满的slab没有空闲对象，都被占用。空的slab则所有对象都是空闲的。

高速缓存使用kmem_cache结构表示，其包含三个链表：slabs_full, slabs_partial, slabs_empty，这三个链表均存放于kmem_list3结构内。

```
struct slab{
	struct list_head 	list;		// 满、空或部分满的链表
	unsigned long 		colouroff;	// slab着色偏移量
	void   				*s_mem;		// slab中的第一个对象
	unsigned int 		inuse;		// slab中已经分配的对象数
	kmem_bufctl_t 		free;		// 第一个空闲对象
}
```

slab描述符要么在slab之外另行分配，要么放下slab自身开始的地方。如果slab很小，或者slab内部有足够空间容纳slab描述符，那么描述符就放在slab里面。

#### slab分配器接口

一个新的高速缓存通过kmem_cache_create()函数类创建
```
struct kmem_cache * kmem_cache_create( const char *name,
										size_t size,
										size_t align,
										unsigned long flags,
										void (*ctor) (void*));
```

第一个参数是字符串，存放缓存的名字，第二个变量使缓存中每个元素的大小，第三个参数是slab内第一个对象的偏移，确保页内的特定对齐，flags表示特定的标志。

	- SLAB_HWCACHE_ALIGN: 把一个slab内的所有对象按高速缓存行对齐
	- SLAB_POISON: 使slab用已知的值（a5a5a5a5）填充slab
	- SLAB_RED_ZONE: 这个标志导致slab层在已分配内存的周围插入红色警戒区。
	- SLAB_PANIC: 这个标志当分配失败时提醒slab层。
	- SLAB_CACHE_DMA: 这个标志命令slab层使用可以执行DMA的内存给slab分配空间。

最后一个参数ctor是构造函数，只有在新的页被追加到高速缓存时，这个函数才会被调用。实际上Linux内核的二高速缓存不使用构造函数。

要撤销一个高速缓存使用 kmem_cache_destroy()，调用该函数时要保证高速缓存中的所有页都是空，而且在调用该函数的过程中不能在访问该高速缓存。
```
int kmem_cache_destroy(struct kmem_cache *cachep)
```

##### 缓存中分配与释放

```
void * kmem_cache_alloc(struct kmem_cache *cachep, gfp_t flags);

void kmem_cache_free(struct kmem_cache * cachep, void *objp);

// slab分配器的使用

// 首先定义一个全局变量，存放指向高速缓存的指针
struct kmem_cache * task_struct_cachep;

// 在内核初始化期间，创建高速缓存
task_struct_cachep = kmem_cache_create("task_struct",
										sizeof(struct task_struct),
										ARCH_MIN_TASKSTRUCT,
										SLAB_PANIC | SLAB_NOTRACK,
										NULL);

										
//创建新的进程描述符
struct task_struct *tsk;
tsk = kmem_cache_alloc(task_struct_cachep, 	GFP_KERNEL);
if(!tsk)
	//分配失败
	return NULL;

//进程执行完后，释放缓存
kmem_cache_free(task_struct_cachep, tsk);


// 最后如果有需要的话，可以撤销该高速缓存
int err;
err = kmem_cache_destroy(task_struct_cachep);

if(err)
	//...
	;
```




