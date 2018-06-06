## NORDIC MESH API

###  NRF_MESH Core Mesh API  @nrf_mesh.c(h)

####  nrf_mesh_init()
mesh协议栈的初始化
```c
/**
 * Initializes the Bluetooth Mesh stack.
 * 初始化mesh协议栈，在此之前需要保证，softdevice已经初始化。
 * 若要开始消息收发，还需要调用nrf_mesh_enable()函数。
 * 除此之外，还需要设置net-keys、app-keys
 */
uint32_t nrf_mesh_init(const nrf_mesh_init_params_t * p_init_params);
```

其参数结构如下：
```c
typedef struct
{
    nrf_clock_lf_cfg_t lfclksrc; 
	nrf_mesh_relay_check_cb_t relay_cb; 
	/**< 判断是否需要中继的回调函数，可以为NULL */
    uint8_t irq_priority; 	
    /**< 中断优先级 (NRF_MESH_IRQ_PRIORITY_LOWEST or NRF_MESH_IRQ_PRIORITY_THREAD). */    
    const uint8_t * p_uuid; 
    	/** 用于未配网设备的becaon帧中UUID，若为NULL，则会自动生成一个UUID */
} nrf_mesh_init_params_t;
```

设置了两个全局变量以判断mesh协议栈是否初始化及使能。

```c
static bool m_is_enabled;
static bool m_is_initialized;
```

`nrf_mesh_init()`中首先进行一系列检查，以排除异常状况。

```c
uint32_t nrf_mesh_init(const nrf_mesh_init_params_t * p_init_params)
{
	// 检查初始化状态
    if (m_is_initialized)
    {
        return NRF_ERROR_INVALID_STATE;
    }

	// 检查参数
    if (p_init_params == NULL)
    {
        return NRF_ERROR_NULL;
    }

	// 确保mesh中断优先级为 NRF_MESH_IRQ_PRIORITY_THREAD
	// 或者为NRF_MESH_IRQ_PRIORITY_LOWEST
    uint8_t irq_priority = p_init_params->irq_priority;
    if ((irq_priority != NRF_MESH_IRQ_PRIORITY_THREAD) &&
        (irq_priority != NRF_MESH_IRQ_PRIORITY_LOWEST))
    {
        return NRF_ERROR_INVALID_PARAM;
    }

	// 如果初始化参数没有设置UUID则自动生成
    if (p_init_params->p_uuid != NULL)
    {
        nrf_mesh_configure_device_uuid_set(p_init_params->p_uuid);
    }
    else
    {
        nrf_mesh_configure_device_uuid_reset();
    }
    
    // 确保softdevice已经初始化
#if !defined(HOST)
    uint8_t softdevice_enabled;
    uint32_t status = sd_softdevice_is_enabled(&softdevice_enabled);
    if (status != NRF_SUCCESS)
    {
        return status;
    }

    if (softdevice_enabled != 1)
    {
        return NRF_ERROR_SOFTDEVICE_NOT_ENABLED;
    }
#endif

    ...
    
}
```

经过一系列检查后，进行基本部件的初始化工作。

```c
{
	// 消息缓存初始化，definitions @msg_cache.c
    msg_cache_init();
    
    // 定时器调度初始化
    timer_sch_init();
    bearer_event_init(irq_priority);
    
    timeslot_init(lfclk_accuracy);
    
    bearer_handler_init();
    scanner_init(scanner_packet_process_cb);
    advertiser_init();

    mesh_flash_init();

#if PERSISTENT_STORAGE
    flash_manager_init();
    flash_manager_action_queue_empty_cb_set(flash_stable_cb);
#endif

#if EXPERIMENTAL_INSTABURST_ENABLED
    core_tx_instaburst_init();
#else
    core_tx_adv_init();
#endif
    network_init(p_init_params);
    transport_init(p_init_params);
    heartbeat_init();
    packet_mgr_init(p_init_params);

#if !defined(HOST)
    status = nrf_mesh_dfu_init();
    if ((status != NRF_SUCCESS) && (status != NRF_ERROR_NOT_SUPPORTED))
    {
        return status;
    }
#endif

    ticker_init();

    (void) ad_listener_subscribe(&m_nrf_mesh_listener);

    m_rx_cb = NULL;
    m_is_initialized = true;
}
```



#### nrf_mesh_enable(void)
使能mesh协议栈

```c
/**
 * Enables the Mesh.
 *
 * @note Calling this function alone will not generate any events unless:
 *         - Network and application keys have been added.
 *         - At least one RX address has been added.
 *
 * @see nrf_mesh_rx_addr_add()
 *
 * @retval NRF_SUCCESS             The Mesh was started successfully.
 * @retval NRF_ERROR_INVALID_STATE The mesh was not initialized,
 *                                 see @ref nrf_mesh_init().
 */
uint32_t nrf_mesh_enable(void);
```

#### nrf_mesh_disable()

停止mesh协议栈
```c

/**
 * Disables the Mesh.
 *
 * Calling this function will stop the Mesh, i.e, it will stop ordering
 * time slots from the SoftDevice and will not generate events.
 *
 * @retval NRF_SUCCESS The Mesh was stopped successfully.
 */
uint32_t nrf_mesh_disable(void);
```

#### nrf_mesh_packet_send()

将消息加入队列以发送

```c
/**
 * Queues a mesh packet for transmission.
 * 
 */
uint32_t nrf_mesh_packet_send(const nrf_mesh_tx_params_t * p_params,
                              uint32_t * const p_packet_reference);
```

