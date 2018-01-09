## 块I/O层

### 基本概念

系统中可以随机访问固定大小数据片的硬件设备称做块设备，这些固定大小的数据片称之为块。还有一种基本的设备称之为字符设备，其需要按照顺序访问，比如键盘。

- 扇区：块设备中最小的寻址单元称为扇区，扇区是块设备的物理属性。
- 块： 文件系统最小的逻辑可寻址单元。是文件系统的一种抽象。
- 缓冲区： 当一个快被调入内存时候，存在一个缓冲区中。每个缓冲区与一个块对应，相当于磁盘块在内存中的表示。
- 缓冲区头： 每个缓冲区都有一个描述符 buffer_head ，用来描述内核处理数据时的相关控制信息。

```
struct buffer_head {
	unsigned long b_state;		/* buffer state bitmap (see below) */
	struct buffer_head *b_this_page;/* circular list of page's buffers */
	struct page *b_page;		/* the page this bh is mapped to */

	sector_t b_blocknr;		/* start block number */
	size_t b_size;			/* size of mapping */
	char *b_data;			/* pointer to data within the page */

	struct block_device *b_bdev;
	bh_end_io_t *b_end_io;		/* I/O completion */
 	void *b_private;		/* reserved for b_end_io */
	struct list_head b_assoc_buffers; /* associated with another mapping */
	struct address_space *b_assoc_map;	/* mapping this buffer is
						   associated with */
	atomic_t b_count;		/* users using this buffer_head */
};

// b_state成员的标志如下所示：

enum bh_state_bits {
	BH_Uptodate,	/* Contains valid data */
	BH_Dirty,	/* Is dirty */
	BH_Lock,	/* Is locked */
	BH_Req,		/* Has been submitted for I/O */
	BH_Uptodate_Lock,/* Used by the first bh in a page, to serialise
			  * IO completion of other buffers in the page
			  */

	BH_Mapped,	/* Has a disk mapping */
	BH_New,		/* Disk mapping was newly created by get_block */
	BH_Async_Read,	/* Is under end_buffer_async_read I/O */
	BH_Async_Write,	/* Is under end_buffer_async_write I/O */
	BH_Delay,	/* Buffer is not yet allocated on disk */
	BH_Boundary,	/* Block is followed by a discontiguity */
	BH_Write_EIO,	/* I/O error on write */
	BH_Ordered,	/* ordered write */
	BH_Eopnotsupp,	/* operation not supported (barrier) */
	BH_Unwritten,	/* Buffer is allocated on disk but not written */
	BH_Quiet,	/* Buffer Error Prinks to be quiet */

	BH_PrivateStart,/* not a state bit, but the first bit available
			 * for private allocation by other entities
			 */
};
```2

缓冲区头结构的第一个成员是b_state表示缓冲区状态，其可以是一种或几种标志的组合。

b_count是缓冲区使用计数，可以使用get_bh()和put_bh()对该成员进行增减。在使用缓冲区之前应该首先使用get_bh()增加缓存区计数，使用完之后使用put_bh()减少其使用计数。

与缓冲区对应的物理磁盘块由b_blocknr_th成员索引，该值是b_bdev指明的块设备中的逻辑块号。

与缓冲区对应的而物理内存页是由b_page表示，另外b_data直接指向相应的块（位于b_page所指向的页的某个位置上），块的大小由b_size表示，起始位置在b_data处，结束位置在b_data+b_size处。

缓冲区头的目的在于描述磁盘快和物理缓冲区之间的映射关系。

### bio结构体

内核中块IO的基本操作由bio结构体表示，该结构体代表了正在活动的以片断链表形式组织的块IO操作，一个片段是一小块连续的内存缓冲区。这样的话就不需要保证单个缓冲区一定要连续起来。

```
struct bio {
	sector_t				bi_sector;	/* device address in 512 byte 磁盘上的扇区
										sectors */
	struct bio				*bi_next;	/* request queue link */
	struct block_device		*bi_bdev;
	unsigned long			bi_flags;	/* status, command, etc */
	unsigned long			bi_rw;		/* bottom bits READ/WRITE,
										 * top bits priority
										 */

	unsigned short			bi_vcnt;	/* how many bio_vec's */
	unsigned short			bi_idx;		/* current index into bvl_vec */

	/* Number of segments in this BIO after
	 * physical address coalescing is performed.
	 */
	unsigned int			bi_phys_segments;

	unsigned int			bi_size;	/* residual I/O count */

	/*
	 * To keep track of the max segment size, we account for the
	 * sizes of the first and last mergeable segments in this bio.
	 */
	unsigned int			bi_seg_front_size;
	unsigned int			bi_seg_back_size;

	unsigned int			bi_max_vecs;	/* max bvl_vecs we can hold */

	unsigned int			bi_comp_cpu;	/* completion CPU */

	atomic_t				bi_cnt;		/* pin count */

	struct bio_vec			*bi_io_vec;	/* the actual vec list */

	bio_end_io_t			*bi_end_io;

	void					*bi_private;
	
#if defined(CONFIG_BLK_DEV_INTEGRITY)
	struct bio_integrity_payload *bi_integrity;  /* data integrity */
#endif

	bio_destructor_t		*bi_destructor;	/* destructor */

	/*
	 * We can inline a number of vecs at the end of the bio, to avoid
	 * double allocations for a small number of bio_vecs. This member
	 * MUST obviously be kept at the very end of the bio.
	 */
	 
	struct bio_vec			bi_inline_vecs[0];
};
```

bi_io_vec指向一个bio_vec结构体数组，每个bio_vec结构包含<page, offset, len>三个元素，描述一个特定片断：片断所在的物理页、块在物理页中的偏移，从给定偏移量开始的块长度。

bi_vcnt表示bi_io_vec所指向的数组中bio_vec的数量。当块IO操作执行完后，bi_idx指向数组的当前索引。

每个IO请求都通过一个bio结构体表示，每个请求包含了一个或多个块，这些块存储在bio_vec中。bio_vec结构体描述了每个片断在物理页中的实际位置。bi_idx指向数组中当前的bio_vec片断，块I/O层可以通过它跟踪块IO完成的进度。

缓冲区头和bio结构体之间有着明显的差别，bio结构体代表的是IO操作，它可以包括内存中的一个或多个页；而另一方面，buffer_head结构体代表的是一个缓冲区，它描述的仅仅是磁盘中的一个块。因为缓冲区头是关联单独页中的单独块，所以它可能引起不必要的分割，将请求按块进行分割，只能靠以后重新组合。bio结构体是轻量级的，他表述的块不需要连续存储区，并且不需要分割I/O操作。

### 请求队列

块设备将它们挂起的块IO请求保存在请求队列中，该队列有request_queue结构体表示，定义在文件<linux/blkdev.h>中，包含一个双向请求链表以及相关的控制信息。通过内核中文件系统这样的高层代码将请求加入到队列中。请求队列只要不为空，队列对应的块设备驱动程序就从队列中获取请求，然后将其送到对应的块设备上去。

### I/O调度程序

内核不会简单的按照请求产生的次序将IO请求提供给相应的块设备，而是进行了一定的优化。

I/O调度程序将磁盘I/O资源分配给系统中挂起的块I/O请求。I/O调度程序觉得队列中请求的排序，以及在什么时候将请求发送到块设备。具体来说，这种资源分配是通过将请求队列中挂起的请求进行请求合并和排序来完成的。

	进程调度程序是将处理器资源分配给系统中运行的进程。

合并是指将两个多多个请求结合成一个新请求，比如有两个请求是从磁盘的相邻位置读取数据，那么就可以将两个请求合并为一个请求。合并请求可以减少对磁盘寻址的次数，从而减少系统开销。

排列是指将请求队列按照磁盘扇区增长的方向排列，这样可以使磁盘头始终沿着同一个方向访问磁盘，减小了访问磁盘的时间。

#### linus电梯——I/O调度程序

当有新的调度程序加入请求队列时，首先检查在已有队列中是否有请求可以与之合并；若没有就插入请求队列，插入时要保证整个请求队列依然时沿着扇区方向的，若没有合适的位置，就插入到队列尾部；除此之外，当发现队列中有队列驻留时间过长的请求，那么新的请求也会加入到尾队列尾部。这是为了避免相邻磁盘的请求太多而造成的访问其他位置的请求难以得到执行。

#### 最终期限I/O调度程序

为了解决I/O调度程序饥饿而提出的方法，出于减少磁盘访问时间的考虑，对某个磁盘区域的繁重操作显然会使其他区域的访问得不到执行。

I/O请求中，写操作通常是在内核有空的时候才将请求提交给磁盘，也就是是说写操作和提交给他的应用程序异步执行。而读操作恰恰相反，通常当应用程序提交一个读请求时，应用程序会发生阻塞，直到读请求被满足，即读操作与提交他的应用程序同步执行。可见，读请求对系统性能性能影响很大。

在最后期限IO调度程序中，每个请求都有一个超时时间（比如 读500ms， 写5s）。最后期限调度类似于Linus电梯，也是以物理磁盘扇区增长的顺序排序的。当新的请求到来也是类似于linus电梯方式进行合并和插入。但是，最后期限调度程序也会同时将请求插入到额外的队列中。读请求按照次序被插入到特定的FIFO队列中，写请求被插入到特定的FIFO队列中。虽然普通队列是以磁盘扇区顺序排列的，但是着这两个额外的队列是FIFO队列（先进先出）。普通情况下，是将排序队列头部的请求不断发送给磁盘。

如果在写FIFO队列头，或者在读FIFO队列头的请求超时，那么最后期限调度程序便会在FIFO队列中提取请求进行操作。该方法可以一定程度上减小饥饿情况的发生。

#### 预测I/O调度程序

预测I/O调度程序在最终期限调度程序的基础上，增加了一个空闲时间，也就说当请求提交后，不会立即返回处理其他请求，而是可以空闲几个ms，等待磁盘相邻位置的请求。对于应用程序来说，这是一个提交请求的机会，任何相邻磁盘位置操作请求都会得到立即处理。在等待时间结束后，预测IO处理程序返回原来位置基础执行以前剩下的请求。如果一个相邻磁盘位置的请求在空闲时刻到来，那么它就会立刻得到执行，从而减小了二次寻址。


####  完全公正的排队I/O调度程序

该方法是为了专有工作负荷设计的，在实际中也为多挣工作负荷提供了较好的性能。该方法是将I/O请求根据引起该请求的进程来组织的，比如进程foo发起的请求放进foo队列，bar进程发起的请求放入bar队列。然后，以时间轮片一次从每个队列中选取请求。

#### 空操作I/O调度程序

所谓空操作是指，该调度方法仅仅进行请求的合并，而不进行其他操作。这主要针对闪存这一类的块设备，这类设备是正在的随机存储，没有必要按序访问。

