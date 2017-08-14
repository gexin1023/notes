## 虚拟文件系统

内核在它的底层文件系统系统接口上建立一个抽象层，该抽象层使Linux可以支持各种文件系统，即便他们在功能和行为上存在很大差异。

VFS抽象层定义了各个文件系统都支持的基本的、概念上的接口和结构数据。

### VFS对象及其数据结构

VFS中有四个主要的对象类型：

- 超级块：表示一个具体的已安装文件系统，包含文件的控制信息等内容。
- 索引点对象：代表一个具体的文件，包含文件的相关信息，比如文件大小、拥有者、创建时间等。
- 目录项对象：代表一个目录项，是路径的组成部分
- 文件对象：代表进程已经打开的文件，显然一个文件可以被多个进程打开，也就是说一个文件可能对应多个文件对象


对于对象都存在对应的操作方法：

- super_operationss对象，包含内核针对特定文件系统所能调用的方法
- inode_operations对象，包含内核针对特定文件所能调用的方法
- dentry_operations对象，包含内核针对特定目录所能进行的操作
- file_operations对象，其中进程针对已打开的文件所进行的操作

#### 超级快对象 super_block

各种文件系统都必须实现超级块对象，该对象用于存储特定文件系统的信息，通常对应于存放在磁盘特定扇区。

```
struct super_block {
	struct list_head	s_list;		/* Keep this first */
	dev_t			s_dev;		/* search index; _not_ kdev_t */
	unsigned char		s_dirt;
	unsigned char		s_blocksize_bits;
	unsigned long		s_blocksize;
	loff_t			s_maxbytes;	/* Max file size */
	struct file_system_type	*s_type;
	const struct super_operations	*s_op;		// 超级块操作方法
	const struct dquot_operations	*dq_op;
	const struct quotactl_ops	*s_qcop;
	const struct export_operations *s_export_op;
	unsigned long		s_flags;
	unsigned long		s_magic;
	struct dentry		*s_root;
	struct rw_semaphore	s_umount;
	struct mutex		s_lock;
	int			s_count;
	int			s_need_sync;
	atomic_t		s_active;
#ifdef CONFIG_SECURITY
	void                    *s_security;
#endif
	struct xattr_handler	**s_xattr;

	struct list_head	s_inodes;	/* all inodes */
	struct hlist_head	s_anon;		/* anonymous dentries for (nfs) exporting */
	struct list_head	s_files;
	/* s_dentry_lru and s_nr_dentry_unused are protected by dcache_lock */
	struct list_head	s_dentry_lru;	/* unused dentry lru */
	int			s_nr_dentry_unused;	/* # of dentry on lru */

	struct block_device	*s_bdev;
	struct backing_dev_info *s_bdi;
	struct mtd_info		*s_mtd;
	struct list_head	s_instances;
	struct quota_info	s_dquot;	/* Diskquota specific options */

	int			s_frozen;
	wait_queue_head_t	s_wait_unfrozen;

	char s_id[32];				/* Informational name */

	void 			*s_fs_info;	/* Filesystem private info */
	fmode_t			s_mode;

	/* Granularity of c/m/atime in ns.
	   Cannot be worse than a second */
	u32		   s_time_gran;

	/*
	 * The next field is for VFS *only*. No filesystems have any business
	 * even looking at it. You had been warned.
	 */
	struct mutex s_vfs_rename_mutex;	/* Kludge */

	/*
	 * Filesystem subtype.  If non-empty the filesystem type field
	 * in /proc/mounts will be "type.subtype"
	 */
	char *s_subtype;

	/*
	 * Saved mount options for lazy filesystems using
	 * generic_show_options()
	 */
	char *s_options;
};
```

超级块对象中断s_op成员指向超级块的操作函数表，其形式如下，该结构体的每一项成员都是一个函数指针，

```
struct super_operations {
   	struct inode *(*alloc_inode)(struct super_block *sb);	// 在给定的超级块下创建并初始化一个新的节点对象
	void (*destroy_inode)(struct inode *);		//释放给定的索引点

   	void (*dirty_inode) (struct inode *);		// 索引节点脏（被修改）时调用此函数
	int (*write_inode) (struct inode *, struct writeback_control *wbc);//将给定索引节点写到磁盘
	void (*drop_inode) (struct inode *);	// 在最后一个指向索引节点的引用被释放后，VFS会调用该函数。VFS只需简单删除这个索引节点。
	void (*delete_inode) (struct inode *);
	void (*put_super) (struct super_block *);
	void (*write_super) (struct super_block *);
	int (*sync_fs)(struct super_block *sb, int wait);
	int (*freeze_fs) (struct super_block *);
	int (*unfreeze_fs) (struct super_block *);
	int (*statfs) (struct dentry *, struct kstatfs *);
	int (*remount_fs) (struct super_block *, int *, char *);
	void (*clear_inode) (struct inode *);
	void (*umount_begin) (struct super_block *);

	int (*show_options)(struct seq_file *, struct vfsmount *);
	int (*show_stats)(struct seq_file *, struct vfsmount *);
#ifdef CONFIG_QUOTA
	ssize_t (*quota_read)(struct super_block *, int, char *, size_t, loff_t);
	ssize_t (*quota_write)(struct super_block *, int, const char *, size_t, loff_t);
#endif
	int (*bdev_try_to_free_page)(struct super_block*, struct page*, gfp_t);
};
```

#### 索引节点对象

索引节点对象包含了内核在操作文件或目录时需要的全部信息，其定义如下

```
struct inode {
	struct hlist_node	i_hash;
	struct list_head	i_list;		/* backing dev IO list */
	struct list_head	i_sb_list;
	struct list_head	i_dentry;
	unsigned long		i_ino;
	atomic_t		i_count;
	unsigned int		i_nlink;
	uid_t			i_uid;
	gid_t			i_gid;
	dev_t			i_rdev;
	unsigned int		i_blkbits;
	u64			i_version;
	loff_t			i_size;
#ifdef __NEED_I_SIZE_ORDERED
	seqcount_t		i_size_seqcount;
#endif
	struct timespec		i_atime;
	struct timespec		i_mtime;
	struct timespec		i_ctime;
	blkcnt_t		i_blocks;
	unsigned short          i_bytes;
	umode_t			i_mode;
	spinlock_t		i_lock;	/* i_blocks, i_bytes, maybe i_size */
	struct mutex		i_mutex;
	struct rw_semaphore	i_alloc_sem;
	const struct inode_operations	*i_op;
	const struct file_operations	*i_fop;	/* former ->i_op->default_file_ops */
	struct super_block	*i_sb;
	struct file_lock	*i_flock;
	struct address_space	*i_mapping;
	struct address_space	i_data;
#ifdef CONFIG_QUOTA
	struct dquot		*i_dquot[MAXQUOTAS];
#endif
	struct list_head	i_devices;
	union {
		struct pipe_inode_info	*i_pipe;
		struct block_device	*i_bdev;
		struct cdev		*i_cdev;
	};

	__u32			i_generation;

#ifdef CONFIG_FSNOTIFY
	__u32			i_fsnotify_mask; /* all events this inode cares about */
	struct hlist_head	i_fsnotify_mark_entries; /* fsnotify mark entries */
#endif

#ifdef CONFIG_INOTIFY
	struct list_head	inotify_watches; /* watches on this inode */
	struct mutex		inotify_mutex;	/* protects the watches list */
#endif

	unsigned long		i_state;
	unsigned long		dirtied_when;	/* jiffies of first dirtying */

	unsigned int		i_flags;

	atomic_t		i_writecount;
#ifdef CONFIG_SECURITY
	void			*i_security;
#endif
#ifdef CONFIG_FS_POSIX_ACL
	struct posix_acl	*i_acl;
	struct posix_acl	*i_default_acl;
#endif
	void			*i_private; /* fs or device private pointer */
};
```

一个索引点代表文件系统中的一个文件，它也可以是设备或管道这样的特殊文件。要注意到的是，索引点仅当文件被访问时，才会在内存中创建。 

索引节点操作由结构inode_operations定义，如下所示

```
struct inode_operations {
	int (*create) (struct inode *,struct dentry *,int, struct nameidata *);
	// VFS通过调用create()和open()来调用函数，从而为entry对象创建一个新的索引节点
	
	struct dentry * (*lookup) (struct inode *,struct dentry *, struct nameidata *);
	// 该函数在特定目录寻找索引节点
	
	int (*link) (struct dentry *,struct inode *,struct dentry *);
	// 该函数被系统调用link()调用，穿件硬链接。
	
	int (*unlink) (struct inode * dir,struct dentry *);
	// 从目录dir中删除dentry指定的索引节点对象
	
	int (*symlink) (struct inode *,struct dentry *,const char *);
	int (*mkdir) (struct inode *,struct dentry *,int);
	int (*rmdir) (struct inode *,struct dentry *);
	int (*mknod) (struct inode *,struct dentry *,int,dev_t);
	int (*rename) (struct inode *, struct dentry *,
			struct inode *, struct dentry *);
	int (*readlink) (struct dentry *, char __user *,int);
	void * (*follow_link) (struct dentry *, struct nameidata *);
	void (*put_link) (struct dentry *, struct nameidata *, void *);
	void (*truncate) (struct inode *);
	int (*permission) (struct inode *, int);
	int (*check_acl)(struct inode *, int);
	int (*setattr) (struct dentry *, struct iattr *);
	int (*getattr) (struct vfsmount *mnt, struct dentry *, struct kstat *);
	int (*setxattr) (struct dentry *, const char *,const void *,size_t,int);
	ssize_t (*getxattr) (struct dentry *, const char *, void *, size_t);
	ssize_t (*listxattr) (struct dentry *, char *, size_t);
	int (*removexattr) (struct dentry *, const char *);
	void (*truncate_range)(struct inode *, loff_t, loff_t);
	long (*fallocate)(struct inode *inode, int mode, loff_t offset,
			  loff_t len);
	int (*fiemap)(struct inode *, struct fiemap_extent_info *, u64 start,
		      u64 len);
};

```

#### 目录项

VFS把目录当文件对待，所以路径 /bin/vi中，bin和vi都属于文件，路径中每个组成部分都有一个索引节点对象表示。

为了方便查找操作，VFS引入了目录项的概念，每个dentry代表路径中的一个特定部分，对于上面路径来说，/、bin、vi都是目录项对象。也就是说，路径的每一个部分都是目录项对象。

目录项由dentry结构表示：

```
struct dentry {
	atomic_t d_count;			/* 使用计数 */
	unsigned int d_flags;		/* 目录项标识 protected by d_lock */
	spinlock_t d_lock;			/* 单目录项锁 per dentry lock */
	int d_mounted;				/* 是登陆点的目录项吗 */
	struct inode *d_inode;		/* 相关联的索引节点 Where the name belongs to - NULL is
					 * negative */
	/*
	 * The next three fields are touched by __d_lookup.  Place them here
	 * so they all fit in a cache line.
	 */
	struct hlist_node d_hash;	/* 散列表lookup hash list */
	struct dentry *d_parent;	/* 父目录的目录项对象 parent directory */
	struct qstr d_name;		    /* 目录项的名称 */

	struct list_head d_lru;		/* 未使用链表 LRU list */
	/*
	 * d_child and d_rcu can share memory
	 */
	union {
		struct list_head d_child;	/* 目录项内部形成的链表 child of parent list */
	 	struct rcu_head d_rcu;		/* RCU加锁 */
	} d_u;
	struct list_head d_subdirs;	/* 子目录链表 our children */
	struct list_head d_alias;	/* 索引点别名链表 inode alias list */
	unsigned long d_time;		/* 重置时间 used by d_revalidate */
	const struct dentry_operations *d_op; /* 目录项操作指针 */
	struct super_block *d_sb;	/* 文件的超级块 The root of the dentry tree */
	void *d_fsdata;			/* 文件系统特有数据 fs-specific data */

	unsigned char d_iname[DNAME_INLINE_LEN_MIN];	/* 短文件名 small names */
};
```

目录项状态有三种，被使用、未被使用、负状态。

一个被使用的目录项对应一个有效索引节点，表面该对象有一个或多个使用者。一个目录项处于使用状态意味它被VFS使用并且指向有效数据。

一个未被使用的目录项对应一个有效的索引节点，但是VFS当前并未使用它。该目录项依然指向一个有效的对象，而且被保留在缓存中一遍需要时使用它。

一个负目录项表示没有对应有效的索引节点，因为节点已经被删除了，或路径不在正确了，但是项目依然保留，以便快速解析以后的目录查找。

#### 目录项缓存

如果VFS层遍历路径名中所有的元素并将它们逐个解析成目录项对象，还要达到最深层次的目录，将是一件费力的事，所以内核将目录项对象保存在目录项缓存中（D-Cache）。

目录项缓存主要包括三个主要部分：

- “被使用的”目录项链表，由相关的索引节点中i_dentry成员所引出的目录项构成的链表。该链表通过索引节点对象中的i_dentry项连接相关的目录项对象，因为一个给定的索引节点可能有多个链接，所以就有可能有多个目录项对象，因此用一个链表来表示他们。
Lists of “used” dentries linked off their associated inode via the i_dentry field of
the inode object. Because a given inode can have multiple links, there might be
multiple dentry objects; consequently, a list is used.

- “最近被使用的”目录项双向链表，该链表含有未被使用和负状态的目录项对象
A doubly linked “least recently used” list of unused and negative dentry objects.The
list is inserted at the head, such that entries toward the head of the list are newer
than entries toward the tail.When the kernel must remove entries to reclaim memory, the entries are removed from the tail; those are the oldest and presumably have
the least chance of being used in the near future.

- 散列表和相应的散列函数用来快速地将给定路径解析为相关的目录项对象。
A hash table and hashing function used to quickly resolve a given path into the
associated dentry object.

#### 目录项操作 

```
struct dentry_operations {
	int (*d_revalidate)(struct dentry *, struct nameidata *);
	/* 该函数判断目录项是否有效 */
	
	int (*d_hash) (struct dentry *, struct qstr *);
	/* 为目录项生成散列值，当目录项需要加入散列表时，调用该函数 */
	
	int (*d_compare) (struct dentry *, struct qstr *, struct qstr *);
	/* 用来比较两个文件名字 */
	
	int (*d_delete)(struct dentry *);
	/* 当目录项d_count计数为0时，调用该函数 */
	
	void (*d_release)(struct dentry *);
	/*  当目录项将要被释放时，调用该函数 */
	
	void (*d_iput)(struct dentry *, struct inode *);
	/*当一个目录项丢失了一个索引点时调用该函数 */
	
	char *(*d_dname)(struct dentry *, char *, int);
};

```

### 文件对象

文件对象表示进程已经打开的文件，是已经打开的文件在内存中的表示。该对象有open()系统调用创建，有close()系统调用撤销。这些文件相关的调用在文件操作表中定义。多个进程可以打开同一个文件，所以一个文件可能存在多个文件对象。文件对象仅仅在进程观念上代表已经打开的文件，它所对应的目录项才代表已打开的文件。一个文件对应的文件对象不是唯一的，但是对应的索引节点和目录项是唯一的。

	The file object is the in-memory representation of an open file.The object (but not the physical file) is created in response to the open() system call and destroyed in response to the close() system call.All these file-related calls are actually methods defined in the file operations table. Because multiple processes can open and manipulate a file at the same time, there can be multiple file objects in existence for the same file.The file object merely represents a process’s view of an open file.The object points back to the dentry (which in turn points back to the inode) that actually represents the open file.The inode and dentry objects, of course, are unique.

```
struct file {
	/*
	 * fu_list becomes invalid after file_free is called and queued via
	 * fu_rcuhead for RCU freeing
	 */
	union {
		struct list_head	fu_list;	/* 文件对象链表 */
		struct rcu_head 	fu_rcuhead; /* 释放之后的RCU链表 */
	} f_u;
	struct path		f_path;				/* 包含目录项 */
#define f_dentry	f_path.dentry		
#define f_vfsmnt	f_path.mnt
	const struct file_operations	*f_op;/* 文件操作方法 */
	spinlock_t		f_lock;  /* f_ep_links, f_flags, no IRQ */
	atomic_long_t		f_count;
	unsigned int 		f_flags;
	fmode_t			f_mode;
	loff_t			f_pos;
	struct fown_struct	f_owner;
	const struct cred	*f_cred;
	struct file_ra_state	f_ra;

	u64			f_version;
#ifdef CONFIG_SECURITY
	void			*f_security;
#endif
	/* needed for tty driver, and maybe others */
	void			*private_data;

#ifdef CONFIG_EPOLL
	/* Used by fs/eventpoll.c to link all the hooks to this file */
	struct list_head	f_ep_links;
#endif /* #ifdef CONFIG_EPOLL */
	struct address_space	*f_mapping;
#ifdef CONFIG_DEBUG_WRITECOUNT
	unsigned long f_mnt_write_state;
#endif
};
```

#### 文件操作

```
struct file_operations {
	struct module *owner;
	loff_t (*llseek) (struct file *, loff_t, int);
	ssize_t (*read) (struct file *, char __user *, size_t, loff_t *);
	ssize_t (*write) (struct file *, const char __user *, size_t, loff_t *);
	ssize_t (*aio_read) (struct kiocb *, const struct iovec *, unsigned long, loff_t);
	ssize_t (*aio_write) (struct kiocb *, const struct iovec *, unsigned long, loff_t);
	int (*readdir) (struct file *, void *, filldir_t);
	unsigned int (*poll) (struct file *, struct poll_table_struct *);
	int (*ioctl) (struct inode *, struct file *, unsigned int, unsigned long);
	long (*unlocked_ioctl) (struct file *, unsigned int, unsigned long);
	long (*compat_ioctl) (struct file *, unsigned int, unsigned long);
	int (*mmap) (struct file *, struct vm_area_struct *);
	int (*open) (struct inode *, struct file *);
	int (*flush) (struct file *, fl_owner_t id);
	int (*release) (struct inode *, struct file *);
	int (*fsync) (struct file *, struct dentry *, int datasync);
	int (*aio_fsync) (struct kiocb *, int datasync);
	int (*fasync) (int, struct file *, int);
	int (*lock) (struct file *, int, struct file_lock *);
	ssize_t (*sendpage) (struct file *, struct page *, int, size_t, loff_t *, int);
	unsigned long (*get_unmapped_area)(struct file *, unsigned long, unsigned long, unsigned long, unsigned long);
	int (*check_flags)(int);
	int (*flock) (struct file *, int, struct file_lock *);
	ssize_t (*splice_write)(struct pipe_inode_info *, struct file *, loff_t *, size_t, unsigned int);
	ssize_t (*splice_read)(struct file *, loff_t *, struct pipe_inode_info *, size_t, unsigned int);
	int (*setlease)(struct file *, long, struct file_lock **);
};
```

### 和文件系统相关的数据结构

除了几个VFS对象，内核还使用了另一些标准数据结构来管理文件系统的其他相关数据。第一个对象是file_system_type，用来描述特定文件系统类型，必须ext3,ext4等。第二个结构体是vfsmount，用来描述一个安装文件的实例。

因为Linux支持众多的文件系统，因此需要一个特殊结构来描述各种文件系统的功能和行为，也就是file_system_type

```
struct file_system_type {
	const char *name;
	int fs_flags;
	int (*get_sb) (struct file_system_type *, int,
		       const char *, void *, struct vfsmount *);
	void (*kill_sb) (struct super_block *);
	struct module *owner;
	struct file_system_type * next;
	struct list_head fs_supers;

	struct lock_class_key s_lock_key;
	struct lock_class_key s_umount_key;

	struct lock_class_key i_lock_key;
	struct lock_class_key i_mutex_key;
	struct lock_class_key i_mutex_dir_key;
	struct lock_class_key i_alloc_sem_key;
};
```

每个文件系统不管有多少个实例安装在系统中，还是根本没有安装到系统中，都只有一个file_system_type结构。

当文件系统被安装时，将有一个 vfsmount结构体在安装点被创建。

```
struct vfsmount {
	struct list_head mnt_hash;
	struct vfsmount *mnt_parent;	/* fs we are mounted on */
	struct dentry *mnt_mountpoint;	/* dentry of mountpoint */
	struct dentry *mnt_root;	/* root of the mounted tree */
	struct super_block *mnt_sb;	/* pointer to superblock */
	struct list_head mnt_mounts;	/* list of children, anchored here */
	struct list_head mnt_child;	/* and going through their mnt_child */
	int mnt_flags;
	/* 4 bytes hole on 64bits arches */
	const char *mnt_devname;	/* Name of device e.g. /dev/dsk/hda1 */
	struct list_head mnt_list;
	struct list_head mnt_expire;	/* link in fs-specific expiry list */
	struct list_head mnt_share;	/* circular list of shared mounts */
	struct list_head mnt_slave_list;/* list of slave mounts */
	struct list_head mnt_slave;	/* slave list entry */
	struct vfsmount *mnt_master;	/* slave is on master->mnt_slave_list */
	struct mnt_namespace *mnt_ns;	/* containing namespace */
	int mnt_id;			/* mount identifier */
	int mnt_group_id;		/* peer group identifier */
	/*
	 * We put mnt_count & mnt_expiry_mark at the end of struct vfsmount
	 * to let these frequently modified fields in a separate cache line
	 * (so that reads of mnt_flags wont ping-pong on SMP machines)
	 */
	atomic_t mnt_count;
	int mnt_expiry_mark;		/* true if marked for expiry */
	int mnt_pinned;
	int mnt_ghosts;
#ifdef CONFIG_SMP
	int __percpu *mnt_writers;
#else
	int mnt_writers;
#endif
};
```
