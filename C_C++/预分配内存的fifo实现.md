## 预分配内存fifo实现可变长度字节序列存储


github链接[https://github.com/gexin1023/utils/tree/master/fifo](https://github.com/gexin1023/utils/tree/master/fifo) 

fifo即先进先出队列，可以用链表来实现，在链表头部插入数据，尾部读数据，每次插入新的数据都动态分配一段内存用于数据存储,适用于变长数据的队列实现。也可以用数组实现，用一个数组`buf[LEN]`作为缓存，用两个整数分别记录写数据和读数据的位置，适用于每次读取相同长度数据的场景。

有的场景中，要避免频繁的malloc/free动态分配释放，与此同时数据长度不定。因此，需要预分配一段空间存储数据，也需要记录每一个数据的长度，方便存取。

### fifo数据结构
```
typedef struct
{
    unsigned int pos;   // position index in buffer
    unsigned int len;   // the length of data
    list_node_t  node;
}pos_t;

typedef struct
{
    unsigned char   *buffer;
    unsigned int    size;
    unsigned int    in;
    unsigned int    out;
    list_node_t     pos_head;
} fifo_t;
```

设计以上的数据结构，`buffer`即为fifo的存储空间，开始时根据需要预分配，`size`表示`buffer`的长度。`in`和`out`分别记录读写数据的位置，`pos_t`结构组成的链表用于记录每次写入数据的位置及长度。

### fifo接口

```
fifo_t * fifo_init(unsigned char *buf, unsigned int size);

fifo_t *fifo_alloc(unsigned int size);

void fifo_free(fifo_t *fifo);


/* fifo_put, 向fifo加入数据
 * @fifo,   目标fifo
 * @buf,    数据
 * @len,    数据长度
 * 如果空间不够，就删除最旧的数据，新数据覆盖旧数据 
 */
unsigned int fifo_put(fifo_t *fifo, unsigned char *buf, unsigned int len);

/* fifo_put_tail
 * 有时会存在优先级比较高的数据需要放在最先出队的位置 
 * /
unsigned int fifo_put_tail(fifo_t *fifo, unsigned char *buf, unsigned int len);

/* fifo_get
 * 取数据
 */
int fifo_get(fifo_t *fifo, unsigned char *buf, unsigned int *p_len);

/* fifo_get_len
 * 获取数据长度
 */
int fifo_get_len(fifo_t *fifo);
```

fifo接口的实现如下：

```

/* fifo_init:   create a fifo using a preallocated memory
 *
 * buf:     preallocated memory
 * size:    the length of the preallocated memory, 取以2为底的整数
 */
fifo_t * fifo_init(unsigned char *buf, unsigned int size)
{
    fifo_t *fifo = (fifo_t *)malloc(sizeof(fifo_t));
    
    fifo->buffer = buf;
    fifo->size   = size;
    fifo->in = fifo->out = 0;
    fifo->pos_head.next = &(fifo->pos_head);
    fifo->pos_head.prev = &(fifo->pos_head);
    return fifo;
}

/* fifo_alloc:   create a fifo 
 *
 * size: the length of the allocated memory
 */
fifo_t *fifo_alloc(unsigned int size)
{
    unsigned char * buf = (unsigned char *)malloc(size);
    return fifo_init(buf, size);
}

/* fifo_free: 
 *
 */

void fifo_free(fifo_t *fifo)
{
    free(fifo->buffer);
    free(fifo);
}

/* fifo_put, 向fifo加入数据
 * @fifo,   目标fifo
 * @buf,    数据
 * @len,    数据长度
 * 如果空间不够，就删除最旧的数据，新数据覆盖旧数据 */
static unsigned int __fifo_put(fifo_t *fifo, unsigned char *buf, unsigned int len)
{
    unsigned int l;
    
    /* fifo 空间不足时，删除旧内容，直到可以容纳新的数据 */
    while(len>(fifo->size - fifo->in + fifo->out))
    {
        pos_t *pos = list_entry(fifo->pos_head.prev, pos_t, node);
        fifo->out += pos->len;
        list_del(fifo->pos_head.prev);
        free(pos);
    }
    

    /* 首先复制数据从（ in % buf_size）位置到buffer结尾 */
    l = min(len , fifo->size - (fifo->in & (fifo->size-1)));
    memcpy(fifo->buffer + (fifo->in & (fifo->size-1)), buf ,l);

    /* 然后复制剩下的数据从buffer开头开始 */
    memcpy(fifo->buffer, buf+l, len-l);

    /* 加入新的位置节点 */
    pos_t *pos = (pos_t *)malloc(sizeof(pos_t));
    pos->len=len;
    pos->pos=fifo->in;
    list_add(&(fifo->pos_head), &(pos->node));

    /* 更改写入点索引 */
    fifo->in += len;
    
    return len;
}

unsigned int fifo_put(fifo_t *fifo, unsigned char *buf, unsigned int len)
{
    return __fifo_put(fifo, buf, len);
}

unsigned int fifo_put_tail(fifo_t *fifo, unsigned char *buf, unsigned int len)
{
    unsigned int l;
    
    /* fifo 空间不足时，删除旧内容，直到可以容纳新的数据 */
    while(len>(fifo->size - fifo->in + fifo->out))
    {
        pos_t *pos = list_entry(fifo->pos_head.prev, pos_t, node);
        fifo->out += pos->len;
        list_del(fifo->pos_head.prev);
        free(pos);
    }

    
    fifo->out -= len;

    /* 首先复制数据从（ out % buf_size）位置到buffer结尾 */
    l = min(len , fifo->size - (fifo->out & (fifo->size-1)));
    memcpy(fifo->buffer + (fifo->out & (fifo->size-1)), buf ,l);

    /* 然后复制剩下的数据从buffer开头开始 */
    memcpy(fifo->buffer, buf+l, len-l);

    /* 加入新的位置节点 */
    pos_t *pos = (pos_t *)malloc(sizeof(pos_t));
    pos->len=len;
    pos->pos=fifo->out;
    list_add_tail(&(fifo->pos_head), &(pos->node));
    
    return len;
}

int fifo_get(fifo_t * fifo, unsigned char * buf, unsigned int * p_len)
{
    if(fifo->pos_head.next == &(fifo->pos_head))
    {
        // fifo is emperty
        return -1;
    }

    pos_t *pos = list_entry(fifo->pos_head.prev, pos_t, node);
    *p_len = pos->len;

    list_del(&(pos->node));

    free(pos);


    int l = min(*p_len, fifo->size - (fifo->out &(fifo->size-1)));
    memcpy(buf, fifo->buffer+(fifo->out & (fifo->size-1)), l);
    memcpy(buf+l, fifo->buffer, *p_len-l);
   

    fifo->out += *p_len;
    return *p_len;
}

int fifo_get_len(fifo_t * fifo)
{
    if(fifo->pos_head.next == &(fifo->pos_head))
    {
        // fifo is emperty
        return -1;
    }

    pos_t *pos = list_entry(fifo->pos_head.prev, pos_t, node);
   
    return (int)pos->len;
}
```




