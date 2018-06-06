## nordic mesh 任务调度实现

nordic mesh的任务调度室基于定时器实现的，有两个链表结构维护任务。

需要注意的是，任务调度的部分接口只能在“bearer event”的中段级别被调用，因此调用的形式是通过设置"bearer event"事件来实现的。

### 结构及接口 @timer_scheduler.h

任务调度的结构定义如下：

```c

typedef enum
{
    TIMER_EVENT_STATE_UNUSED,      /**< Not present in the scheduler */
    TIMER_EVENT_STATE_ADDED,       /**< Added for processing */
    TIMER_EVENT_STATE_QUEUED,      /**< Queued for firing */
    TIMER_EVENT_STATE_RESCHEDULED, /**< Rescheduled, but not resorted */
    TIMER_EVENT_STATE_ABORTED,     /**< Aborted, but still in the list */
    TIMER_EVENT_STATE_IGNORED,     /**< Aborted, but added for processing */
    TIMER_EVENT_STATE_IN_CALLBACK  /**< Currently being called */
} timer_event_state_t;

/**
 * Timer event structure for schedulable timers.
 */
typedef struct timer_event
{
    volatile timer_event_state_t state;     /**< Timer event state. */
    timestamp_t                  timestamp; /**< 定时器激发时间  */
	
    timer_sch_callback_t         cb;        /**< 定时器激发时的回调函数 */
    uint32_t                     interval;  /**< 周期事件的时间间隔，0时表示单次事件 */
    void *                       p_context; /**< 调用回调函数cb时传入的参数 */
    struct timer_event*          p_next;    /**< 指向下一个事件结构的指针，构成链表 */
} timer_event_t;
```

相关的几个接口如下：

```c

/**
 * Initializes the scheduler module.
 * 在nrf_mesh_init()中被调用
 */
void timer_sch_init(void);

/**
 * Schedules a timer event.
 *
 * @warning This function must be called from @ref BEARER_EVENT "bearer event" IRQ level or lower.
 * 			该函数必须在BEARER_EVENT的中断优先级上被调用
 * @warning The structure parameters must not be changed after the structure has been added to the
 *          scheduler, as this may cause race conditions. If a change in timing is needed, please use the
 *          @ref timer_sch_reschedule() function. If any other changes are needed, abort the event, change the
 *          parameter, and schedule it again.
 *
 * @param[in] p_timer_evt A pointer to a statically allocated timer event, which will be used as
 *                        context for the schedulable event.
 */
void timer_sch_schedule(timer_event_t* p_timer_evt);


/**
 * Aborts a previously scheduled event.
 *
 * @warning This function must be called from @ref BEARER_EVENT "bearer event" IRQ level or lower.
 *			该函数必须在BEARER_EVENT的中断优先级上被调用
 * @param[in] p_timer_evt A pointer to a previously scheduled event.
 */
void timer_sch_abort(timer_event_t* p_timer_evt);

/**
 * Reschedules a previously scheduled event.
 *
 * @warning This function must be called from @ref BEARER_EVENT "bearer event" IRQ level or lower.
 *			该函数必须在BEARER_EVENT的中断优先级上被调用
 * @param[in] p_timer_evt   A pointer to a previously scheduled event.
 * @param[in] new_timestamp When the event should time out, instead of the old time.
 */
void timer_sch_reschedule(timer_event_t* p_timer_evt, timestamp_t new_timestamp);
```

### 任务调度实现 @timer_scheduler.c

调度器中维护两个任务链表`p_head`、`p_add_head`，当有新的任务加入调度时，首先会加到`p_add_head`链表中，此时新任务的状态会设置为"TIMER_EVENT_STATE_ADDED"；之后在`process_add_list()`函数遍历`p_add_head`链表，将其中状态为"TIMER_EVENT_STATE_ADDED"的任务加入到第一个链表`p_head`中。定时器激发时会遍历第一个链表，将其中满足运行时间的事件运行。

调度器结构的定义如下：

```c

typedef struct
{
    timer_event_t* p_head;
    timer_event_t* p_add_head;
    uint32_t dirty_events; /**< Events in the fire list that needs processing */
    uint32_t event_count;
} scheduler_t;

/**
* Static globals
*/
static volatile scheduler_t m_scheduler;	/**<调度器实体 */
static bearer_event_flag_t m_event_flag;	/**<bearer event 事件 */

```
需要注意的是，部分定时器接口是在”bearer event“的中断优先级的被调用的。因此，定义了m_event_flag的全局变量，以产生bearer-event事件调用。

#### timer_sch_init()

调度器初始化函数的实现代码如下，在其中首先将调度器实体全部归零，然后注册了一个bearer-event事件。

```c
void timer_sch_init(void)
{
	// 调度器结构置零
    memset((scheduler_t*) &m_scheduler, 0, sizeof(m_scheduler));
	
	// 注册一个bearer-event事件
	// 当调用函数 bearer_event_flag_set(m_event_flag) 时
	// 会产生bearer-event中断，来调用flag_event_cb()回调函数
    m_event_flag = bearer_event_flag_add(flag_event_cb);
}
```

bearer-event事件回调函数

```c 

static bool flag_event_cb(void)
{
	/* 处理被污染的任务
	 * 被污染的任务要么直接移除
	 * 需要再次调度的，则加入到p_add_head链表中
	 */
    process_dirty_events();

    /* add all the popped events back in, at their right places. 
	 * 该函数遍历第二个链表，将其中的新加的任务添加到第一个链表中
	 * 第一个链表的任务是根据运行时间排序的，从头到尾运行时间越来越晚
	 * 即第一个链表的链表头是最先运行的节点
	 */
    process_add_list();

	/* 激发定时器
	 * 遍历第一个链表，一次运行到达运行时间的任务
	 * p_head依次后移
	 */
    fire_timers(timer_now());
	
	/* 
	 * 设置定时器函数下次激发时间
	 */
    setup_timeout(timer_now());

    return true;
}
```

timer_sch_schedule()函数实现如下：

```c

// 任务加入调度时，首先是加入第二个任务链表中
void timer_sch_schedule(timer_event_t* p_timer_evt)
{
    NRF_MESH_ASSERT(p_timer_evt != NULL);
    NRF_MESH_ASSERT(p_timer_evt->cb != NULL);

    uint32_t was_masked;
    _DISABLE_IRQS(was_masked);
    p_timer_evt->p_next = NULL;
    add_to_add_list(p_timer_evt);
    _ENABLE_IRQS(was_masked);

    bearer_event_flag_set(m_event_flag);
}
```

#### timer_sch_abort()

终止某个任务 

```c 

void timer_sch_abort(timer_event_t* p_timer_evt)
{
    NRF_MESH_ASSERT(p_timer_evt != NULL);
    uint32_t was_masked;
    _DISABLE_IRQS(was_masked);
    if (p_timer_evt->state == TIMER_EVENT_STATE_IN_CALLBACK)
    {
        p_timer_evt->state = TIMER_EVENT_STATE_UNUSED;
    }
    else if (p_timer_evt->state == TIMER_EVENT_STATE_ADDED)
    {
        p_timer_evt->state = TIMER_EVENT_STATE_IGNORED;
    }
    else if (p_timer_evt->state != TIMER_EVENT_STATE_UNUSED)
    {
        if (!is_dirty_state(p_timer_evt->state))
        {
            m_scheduler.dirty_events++;
        }
        p_timer_evt->state = TIMER_EVENT_STATE_ABORTED;
        bearer_event_flag_set(m_event_flag);
    }
    _ENABLE_IRQS(was_masked);
}
```

#### timer_sch_reschedule()

再次调度

```c 


void timer_sch_reschedule(timer_event_t* p_timer_evt, timestamp_t new_timeout)
{
    NRF_MESH_ASSERT(p_timer_evt != NULL);

    uint32_t was_masked;
    _DISABLE_IRQS(was_masked);
    /* The events in the added queue will reinsert themselves in the processing. */
    if (p_timer_evt->state == TIMER_EVENT_STATE_UNUSED ||
        p_timer_evt->state == TIMER_EVENT_STATE_IN_CALLBACK)
    {
        add_to_add_list(p_timer_evt);
    }
    else if (p_timer_evt->state == TIMER_EVENT_STATE_ADDED ||
             p_timer_evt->state == TIMER_EVENT_STATE_IGNORED)
    {
        p_timer_evt->state = TIMER_EVENT_STATE_ADDED;
    }
    else
    {
        /* Mark the rescheduled event as dirty, will be processed at the next opportunity. */
        if (!is_dirty_state(p_timer_evt->state))
        {
            m_scheduler.dirty_events++;
        }
        p_timer_evt->state = TIMER_EVENT_STATE_RESCHEDULED;
    }
    p_timer_evt->timestamp = new_timeout;
    bearer_event_flag_set(m_event_flag);
    _ENABLE_IRQS(was_masked);
}

```
















