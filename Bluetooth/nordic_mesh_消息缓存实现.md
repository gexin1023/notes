## nordic mesh中的消息缓存实现

代码文件`msg_cache.h`、`msg_cache.c`。

### 接口定义

头文件中定义了四个接口，供mesh协议栈调用，四个接口如下所示，接口的实现代码在`msg_cache.c`文件中。

```c
@file：msg_cache.h

// 消息缓存初始化
void msg_cache_init(void);

// 检查消息是否存在
bool msg_cache_entry_exists(uint16_t src_addr, uint32_t sequence_number);

// 添加消息到缓存
void msg_cache_entry_add(uint16_t src, uint32_t seq);

// 消息缓存清空
void msg_cache_clear(void);
```

### 实现代码

消息缓存用静态全局变量的一个数组`m_msg_cache[]`实现，该数组长度为32，数组每个元素表示消息。`m_msg_cache_head`表示新消息加入的位置，通过对`m_msg_cache_head`的控制实现一个环形的消息缓存。

其结构定义如下：

```c
typedef struct
{
    bool allocated;  /**< Whether the entry is in use. */
    uint16_t src;    /**< Source address from packet header. */
    uint32_t seq;    /**< Sequence number from the packet header. */
} msg_cache_entry_t;

/** Message cache buffer */
static msg_cache_entry_t m_msg_cache[MSG_CACHE_ENTRY_COUNT];

/** Message cache head index 
 *  新消息的加入位置
 */
static uint32_t m_msg_cache_head = 0; 
```

由缓存结构可知，消息缓存结构只是保存了消息的源地址及其序列号。收到消息时，先在消息缓存中检查是否已有该消息，若不存在则添加进去，否则忽略消息。

由于蓝牙mesh是基于泛洪管理网络的，所以某个节点会收到多条相同的消息（每个节点会中继转发消息），消息缓存主要用判断是否已收到该消息，用来避免消息拥塞。

#### msg_cache_init()

消息缓存初始化代码如下，就是将消息缓存数组的各元素设置为可用状态。

```c

void msg_cache_init(void)
{
    for (uint32_t i = 0; i < MSG_CACHE_ENTRY_COUNT; ++i)
    {
        m_msg_cache[i].src = NRF_MESH_ADDR_UNASSIGNED;
        m_msg_cache[i].seq = 0;
        m_msg_cache[i].allocated = 0;
    }

    m_msg_cache_head = 0;
}

```

#### msg_cache_entry_exists()

判断消息是否已经存在，从`m_msg_cache_head`的位置开始逆序遍历整个消息缓存数组，逐个对比源地址及序列号

```c

bool msg_cache_entry_exists(uint16_t src_addr, uint32_t sequence_number)
{
    /* Search backwards from head */
    uint32_t entry_index = m_msg_cache_head;
    for (uint32_t i = 0; i < MSG_CACHE_ENTRY_COUNT; ++i)
    {
        if (entry_index-- == 0) /* compare before subtraction */
        {
            entry_index = MSG_CACHE_ENTRY_COUNT - 1;
        }

        if (!m_msg_cache[entry_index].allocated)
        {
            return false; /* Gone past the last valid entry. */
        }

        if (m_msg_cache[entry_index].src == src_addr &&
            m_msg_cache[entry_index].seq == sequence_number)
        {
            return true;
        }
    }

    return false;
}

```

#### msg_cache_entry_add()

消息添加到缓存

```c 
void msg_cache_entry_add(uint16_t src, uint32_t seq)
{
    m_msg_cache[m_msg_cache_head].src = src;
    m_msg_cache[m_msg_cache_head].seq = seq;
    m_msg_cache[m_msg_cache_head].allocated = true;

    if ((++m_msg_cache_head) == MSG_CACHE_ENTRY_COUNT)
    {
        m_msg_cache_head = 0;
    }
}
```

#### msg_cache_clear()

清空消息缓存

```c 
void msg_cache_clear(void)
{
    for (uint32_t i = 0; i < MSG_CACHE_ENTRY_COUNT; ++i)
    {
        m_msg_cache[i].allocated = 0;
    }
}
```


















