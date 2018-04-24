## 基于freeRTOS定时器实现闹钟(定时)任务

在智能硬件产品中硬件中，闹钟定时任务是基本的需求。一般通过APP设置定时任务，从云端或者是APP直连硬件将闹钟任务保存在硬件flash中，硬件运行时会去处理闹钟任务。

最简单的实现方式是在循环或者定时器处理函数中不断的去判断当前时间是否等于闹钟设定时间，若相等则产生相应的动作。

这样做虽然可行，但是做了太多无用的计算。我们可以根据当前时间距离下一次闹钟激发时间，设定一个对应的定时器，定时器激发时就是闹钟时间，然后继续根据下次激发时间设定新的定时器，这样可以减少不必要的时间比较。


### 闹钟任务的表示

闹钟任务的表示包含一下及部分的内容：闹钟时间、重复类型、响应操作。

闹钟任务的本地表示可以根据cron格式来定义，当然也可以DIY一个，只要包含以上三个方面内容的就可以。

cron格式的时间表示如下：

| Character | Descriptor | Acceptable Values |
| - | - | - |
| 1 | Minute | 0~59, * (no specific value) |
| 2 | Hour | 0~23, * (no specific value) |
| 3 | Day of month | 1~31, * (no specific value) |
| 4 | Month | 1~12, * (no specific value) |
| 5 | Day of week | 0~7, * (no specific value) |

比如对于一个周一到周五早上7:00的闹钟，可以表示如为：`0  7  *  *  1,2,3,4,5`。


我在代码中定义了如下的闹钟任务:

```
/* cron 格式时间表示 */
typedef struct
{
    int     min;    // minute,          0xFFFFFFFF表示*
    int     hour;   // hour，           0xFFFFFFFF表示*
    int     wday;   // day of week，    0xFFFFFFFF表示*，bit[0~6]  表示周一到周日,如周一到周五每天响铃，则0x1F
    int     mday;   // day of month，   0xFFFFFFFF表示*, bit[0~31] 表示1~31日,每月1,3,5号响铃，则0x15
    int     mon;    // month，          0xFFFFFFFF表示*，bit[0~11] 表示1~12月
} cron_tm_t

/* 闹钟任务 */
typedef struct
{
	int             id;         // 任务ID
	struct tm       tb;         // linux中标准的时间结构体（#include <time.h>），用以表示下一次闹钟激发时间
    cron_tm_t       cron_tm;    // cron格式时间
	unsigned int    flags;      // 重复类型
	int             action;     // 响应操作，对于灯控产品来说，action可以表示开关、颜色、场景等等
    TimerHandle_t   xTimer;     // 定时器句柄
    list_node_t     node;       // 节点，用于将一系列定时任务组织成list
} alarm_task_t;
```

### 闹钟下一次激发时间

基于定时器实现闹钟的原理是在每次设置闹钟时计算出距离下一次激发的间隔时间，然后设置一个相应时间的定时器。

首先需要根据cron格式时间计算出下一次激发的时间

```
/* 距离闹钟下一次激发的时间（min） */
int get_expiry_time(alarm_task_t  *alarm_task)
{
   /* 首先获取当前时间 */
   
   time_t    t;
   struct tm now;

   time(&t);
   localtime_r(&t, &now);

   int ret = 0;

   if(alarm_task->cron_tm.min != 0xFFFFFFFF)
   {
        ret += cron_tm.min - now.tm_min;
   }

   if(alarm_task->cron_tm.hour != 0xFFFFFFFF)
   {
        int flag = 0;

        for(int i=0; i<24; i++)
        {
            if(i == alarm_task->cron_tm.hour){
                flag=1;
                break;
            }
        }

        if(flag){
            ret += 60*(i-now.tm_hour);
        }
   }

   /* 找到wday或者mday中距离今天最近的一天 */
   int t = ret<0?1:0;
   
   int d_wday=0;
   int d_mday=0;

   for(int i=0; i<7; i++)
   {
        int wday = now.tm_wday + t +i;
        if(1<<(wday%7) & alarm_task->cron_tm.wday){
            d_wday=i+t;
            break;
        }
   }

   for(int i=0; i<30; i++)
   {
        int mday = now.tm_mday + t +i;

        if(1<<(mday%30) & alarm_task->cron_tm_.mday){
            d_mday = i+t;
            break;
        }
   }

   int d = d_mday<d_wday?d_mday:d_wday;

   ret += d*24*60;
  
   return ret;
}

```

### 设置定时器

每次硬件上电时，或者闹钟激发时，都要根据下一次激发的时间，设置一次定时器。

```
int set_timer(alarm_task_t  *alarm_task)
{
    alarm_task->xTimer = xTimerCreate( "timer",
                                        get_expiry_time(tm_task) / portTICK_RATE_MS,
                                        1,
                                        (void*)alarm_task,
                                        timer_task_callback );

    if(alarm_task->xTimer == NULL)
    {
        printf("!!! timer created failed\n");
        return -1;
    }
    else
    {
        xTimerStart(alarm_task->xTimer, 0);
    }


    return 0;
}

```
