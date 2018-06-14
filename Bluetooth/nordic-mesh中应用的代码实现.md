## nordic-mesh中应用的代码实现


Nordic-Mesh遵循SIG-Mesh-Profile中的mesh定义，实现了element、model等概念。

一个应用中包含一个或多个element，element是可以寻址的实体；每个element中包含多个model，model定义了应用的功能。

每个设备在provision阶段，其中的每个element都会获得一个unicast-address；在config阶段，设置每个model的APP-Key等内容，该过程通过configure_model实现。每个model的发布地址只有一个，订阅地址可以有多个。

### Provision阶段

provision过程就是先扫描un_provision帧，然后根据UUID选择进行provision的过程，provision就是给未配网节点设置unicast-address、Netkey、IV_Index的过程。

在nordic的示例中，将provisioner相关的接口封装到了provisioner_helper.c(h)中，由以下四个接口函数控制provision过程。

```c 
/* 接口初始化 */
void prov_helper_init(mesh_provisioner_init_params_t * p_prov_init_info);

/* 开始扫描beacon帧 */
void prov_helper_scan_start(void);

/* 根据UUID的过滤字段进行provision， UUID中包含过滤字段的设备会被provision */
void prov_helper_provision_next_device(uint8_t retry_cnt, uint16_t address,
                                       prov_helper_uuid_filter_t * p_uuid_filter);
/* 给provisioner节点本身配置NetKey、Unicast-address */									   
void prov_helper_provision_self(void);									   
```
我们从prov_helper_provision_self()函数的实现中，理解provision的过程，配置本节点与配置其他节点本质是一致的，只是一个直接修改本地状态，一个通过网络传输在接收端通过操作码处理函数修改状态。

```c 
void prov_helper_provision_self(void)
{
    /* Add addresses */
    /* Set and add local addresses and keys, if flash recovery fails. */
    dsm_local_unicast_address_t local_address = {PROVISIONER_ADDRESS, ACCESS_ELEMENT_COUNT};
    ERROR_CHECK(dsm_local_unicast_addresses_set(&local_address));

    /* Generate keys, 随机产生各种KEY*/
    rand_hw_rng_get(m_provisioner.p_nw_data->netkey, NRF_MESH_KEY_SIZE);
    rand_hw_rng_get(m_provisioner.p_nw_data->appkey, NRF_MESH_KEY_SIZE);
    rand_hw_rng_get(m_provisioner.p_nw_data->self_devkey, NRF_MESH_KEY_SIZE);

    /* Add default Netkey and App Key */
    ERROR_CHECK(dsm_subnet_add(0, m_provisioner.p_nw_data->netkey, &m_provisioner.p_dev_data->m_netkey_handle));
    __LOG(LOG_SRC_APP, LOG_LEVEL_INFO, "netkey_handle: %d\n", m_provisioner.p_dev_data->m_netkey_handle);
    ERROR_CHECK(dsm_appkey_add(0, m_provisioner.p_dev_data->m_netkey_handle, m_provisioner.p_nw_data->appkey, &m_provisioner.p_dev_data->m_appkey_handle));

    /* Add device key for the own config server */
    ERROR_CHECK(dsm_devkey_add(PROVISIONER_ADDRESS, m_provisioner.p_dev_data->m_netkey_handle, m_provisioner.p_nw_data->self_devkey, &m_provisioner.p_dev_data->m_self_devkey_handle));

}
```

### Configure阶段

provision之后节点就获得了unicast地址与Netkey，只需要再给节点配置Appkey以及发布订阅地址就可以正常实现功能了，这个过程在nordic_mesh的示例也封装在了node_setup.c(h)中了。

配置server端的的过程如下：

```c
static const config_steps_t server1_server2_config_steps[] =
{
	// 获取composition_data
	// 这里面包含了待配置节点的基本信息，比如有多少了model、多少个element等等
    NODE_SETUP_CONFIG_COMPOSITION_GET,

	// 添加appkey，并绑定到health-server。
	// appkey_add过程，相当于把key保存在本地数据库中，并返回handle
	// appkey_bind过程，相当于把key传输给health_server
    NODE_SETUP_CONFIG_APPKEY_ADD,
    NODE_SETUP_CONFIG_APPKEY_BIND_HEALTH,
	
	// appkey绑定到light_server
    NODE_SETUP_CONFIG_APPKEY_BIND_ONOFF_SERVER,
	
	// 配置health_server的发布地址
    NODE_SETUP_CONFIG_PUBLICATION_HEALTH,
	
	// 配置light_server的发布地址
    NODE_SETUP_CONFIG_PUBLICATION_ONOFF_SERVER1_2,
	
	// 配置light_server的订阅地址
	// 将节点分组，就是给节点的订阅地址加一个组地址
    NODE_SETUP_CONFIG_SUBSCRIPTION_ONOFF_SERVER,
	
    NODE_SETUP_DONE
};
```

AppKey的绑定通过函数config_client_model_app_bind()实现，该函数把需要设置的内容发送到对应的地址，在接收端根据Config_server的操作码处理函数中进行相关的操作。函数注释及定义如下：

```c 
/**
 * Sends a application bind request.
 *
 * @note Response: @ref CONFIG_OPCODE_MODEL_APP_STATUS
 *
 * @param[in] element_address Element address of the model.
 * @param[in] appkey_index    Application key index to bind/unbind.
 * @param[in] model_id        Model ID of the model.
 *
 * @retval NRF_SUCCESS             Successfully sent request.
 * @retval NRF_ERROR_BUSY          The client is in a transaction. Try again later.
 * @retval NRF_ERROR_NO_MEM        Not enough memory available for sending request.
 * @retval NRF_ERROR_INVALID_STATE Client not initialized.
 */
uint32_t config_client_model_app_bind(uint16_t element_address, uint16_t appkey_index, access_model_id_t model_id)
{
    return app_bind_unbind_send(element_address, appkey_index, model_id, CONFIG_OPCODE_MODEL_APP_BIND);
}
```
	有一点需要注意，在nordic_mesh的实现中，不管是key的设置还是address的设置，都是通过handle进行的，这个handle实际上就是数组的index。因此需要首先将地址或key添加到本地数据库，这个歌添加过程会获得一个handle，然后通过handle设置对应内容。

### Model管理

在nordic的实现中，element、model的管理是在access.c(h)中实现的，消息发布是在access_publish.c(h) access_reliable.c(h)中实现的，地址、netkey、appkey的管理是在device_state_manager.c(h)中实现的。

#### 操作码-处理函数

mesh中的功能都是通过model来定义的，SIG_Mesh_Profile文档中定义了四种基本的model，分别是config_server、config_client、health_server、health_client。其中config_server、health_server是默认存在的，且存在于主element中（element_pool[0]即为主element）。

model的功能是通过Opcode-Handler来定义了，一个model中的opcode与响应的处理函数决定了这个model的功能。nordic定义了一个基本的开关灯的model，其支持如下的操作码，并定义了每个操作码消息的内容（即操作码的参数）。我们可以在此基础上添加新的操作码来实现更加复杂的功能，从简单的开关到RGB灯多路控制，再到参数存储、定时任务等。

```c 
/** Simple OnOff opcodes. */
typedef enum
{
    SIMPLE_ON_OFF_OPCODE_SET = 0xC1,            /**< Simple OnOff Acknowledged Set. */
    SIMPLE_ON_OFF_OPCODE_GET = 0xC2,            /**< Simple OnOff Get. */
    SIMPLE_ON_OFF_OPCODE_SET_UNRELIABLE = 0xC3, /**< Simple OnOff Set Unreliable. */
    SIMPLE_ON_OFF_OPCODE_STATUS = 0xC4          /**< Simple OnOff Status. */
} simple_on_off_opcode_t;

/* 开关设置消息的参数 */
typedef struct __attribute((packed))
{
    uint8_t on_off; /**< State to set. */
    uint8_t tid;    /**< Transaction number. */
} simple_on_off_msg_set_t;

/** Message format for th Simple OnOff Set Unreliable message. */
typedef struct __attribute((packed))
{
    uint8_t on_off; /**< State to set. */
    uint8_t tid;    /**< Transaction number. */
} simple_on_off_msg_set_unreliable_t;

/** Message format for the Simple OnOff Status message. */
typedef struct __attribute((packed))
{
    uint8_t present_on_off; /**< Current state. */
} simple_on_off_msg_status_t;

```

在server、client分别定义对应每个Opcode的处理函数，就可以实现每个操作码实现什么操作。对于开关model，其操作码与处理函数对应如下：
```c 
/* server 端 Opcode-Handler */
static const access_opcode_handler_t m_opcode_handlers[] =
{
    {ACCESS_OPCODE_VENDOR(SIMPLE_ON_OFF_OPCODE_SET,            SIMPLE_ON_OFF_COMPANY_ID), handle_set_cb},
    {ACCESS_OPCODE_VENDOR(SIMPLE_ON_OFF_OPCODE_GET,            SIMPLE_ON_OFF_COMPANY_ID), handle_get_cb},
    {ACCESS_OPCODE_VENDOR(SIMPLE_ON_OFF_OPCODE_SET_UNRELIABLE, SIMPLE_ON_OFF_COMPANY_ID), handle_set_unreliable_cb}
};

/* Client 端  Opcode-Handler */
static const access_opcode_handler_t m_opcode_handlers[] =
{
    {{SIMPLE_ON_OFF_OPCODE_STATUS, SIMPLE_ON_OFF_COMPANY_ID}, handle_status_cb}
};
```

#### Model添加

model是通过一个数组结构`m_model_pool[]`来管理，在access.c中定义了一个m_model_pool的全局变量用来管理所有的model。

需要在应用中实现某个Model的话，首先需要将其加入到Model池，这是通过函数`access_model_add()`实现的，下面代码段是该函数的注释及定义。model初始化参数作为函数参数传入，model_handle通过地址方式返回新添加model在model_pool中的index。在函数实现中，首先在model_pool数组中找到未被占用的位置，然后以该位置作为model_handle。
```c 
/**
 * Allocates, initializes and adds a model to the element at the given element index.
 *
 * @param[in]  p_model_params            Pointer to model initialization parameter structure.
 * @param[out] p_model_handle            Pointer to store allocated model handle.
 *
 * @retval     NRF_SUCCESS               Successfully added model to the given element.
 * @retval     NRF_ERROR_NO_MEM          @ref ACCESS_MODEL_COUNT number of models already allocated.
 * @retval     NRF_ERROR_NULL            One or more of the function parameters was NULL.
 * @retval     NRF_ERROR_FORBIDDEN       Multiple model instances per element is not allowed.
 * @retval     NRF_ERROR_NOT_FOUND       Invalid access element index.
 * @retval     NRF_ERROR_INVALID_LENGTH  Number of opcodes was zero.
 * @retval     NRF_ERROR_INVALID_PARAM   One or more of the opcodes had an invalid format.
 * @see        access_opcode_t for documentation of the valid format.
 */
uint32_t access_model_add(const access_model_add_params_t * p_model_params,
                          access_model_handle_t * p_model_handle)
{
	/*
		参数有效性校验		
	*/
	{
        *p_model_handle = find_available_model();
        if (ACCESS_HANDLE_INVALID == *p_model_handle)
        {
            return NRF_ERROR_NO_MEM;
        }

        m_model_pool[*p_model_handle].model_info.publish_address_handle = DSM_HANDLE_INVALID;
        m_model_pool[*p_model_handle].model_info.publish_appkey_handle = DSM_HANDLE_INVALID;
        m_model_pool[*p_model_handle].model_info.element_index = p_model_params->element_index;
        m_model_pool[*p_model_handle].model_info.model_id.model_id = p_model_params->model_id.model_id;
        m_model_pool[*p_model_handle].model_info.model_id.company_id = p_model_params->model_id.company_id;
        m_model_pool[*p_model_handle].model_info.publish_ttl = m_default_ttl;
        increment_model_count(p_model_params->element_index, p_model_params->model_id.company_id);
        ACCESS_INTERNAL_STATE_OUTDATED_SET(m_model_pool[*p_model_handle].internal_state);
    }

    m_model_pool[*p_model_handle].p_args = p_model_params->p_args;
    m_model_pool[*p_model_handle].p_opcode_handlers = p_model_params->p_opcode_handlers;
    m_model_pool[*p_model_handle].opcode_count = p_model_params->opcode_count;

    m_model_pool[*p_model_handle].publication_state.publish_timeout_cb = p_model_params->publish_timeout_cb;
    m_model_pool[*p_model_handle].publication_state.model_handle = *p_model_handle;
    ACCESS_INTERNAL_STATE_ALLOCATED_SET(m_model_pool[*p_model_handle].internal_state);

    return NRF_SUCCESS;
}						  
```

#### Model配置

Model添加后，需要配置过Appkey及发布订阅地址才可以正常工作，配置的过程在config_client端实现，参照前面Config过程。Config_client_model所在的节点，可以直接配置。比如，我们在config-client_model所在的节点上，实现light_client_model的过程如下：

```c 
	/*
     * 初始化 light-client-model，就是一个access_model_add()的过程
     */	
	uint32_t simple_on_off_client_init(simple_on_off_client_t * p_client, uint16_t element_index)
	{
		if (p_client == NULL ||
			p_client->status_cb == NULL)
		{
			return NRF_ERROR_NULL;
		}

		access_model_add_params_t init_params;
		init_params.model_id.model_id = SIMPLE_ON_OFF_CLIENT_MODEL_ID;
		init_params.model_id.company_id = SIMPLE_ON_OFF_COMPANY_ID;
		init_params.element_index = element_index;
		init_params.p_opcode_handlers = &m_opcode_handlers[0];
		init_params.opcode_count = sizeof(m_opcode_handlers) / sizeof(m_opcode_handlers[0]);
		init_params.p_args = p_client;
		init_params.publish_timeout_cb = handle_publish_timeout;
		return access_model_add(&init_params, &p_client->model_handle);
	}

	/*
     * 初始化 四个light_client
     */
    __LOG(LOG_SRC_APP, LOG_LEVEL_INFO, "Initializing and adding light-client models\n");

    for (uint32_t i = 0; i < CLIENT_MODEL_INSTANCE_COUNT; ++i)
    {
        m_clients[i].status_cb = client_status_cb;
        m_clients[i].timeout_cb = client_publish_timeout_cb;
        uint32_t ret=simple_on_off_client_init(&m_clients[i], i + 1);
        
        ERROR_CHECK(access_model_subscription_list_alloc(m_clients[i].model_handle));
    }

    /*
     * 绑定appkey，及设置发布订阅地址
     */
    for(int i =0 ; i<4; i++){
        ERROR_CHECK(access_model_application_bind(m_clients[i].model_handle, m_dev_handles.m_appkey_handle));
        ERROR_CHECK(access_model_publish_application_set(m_clients[i].model_handle, m_dev_handles.m_appkey_handle));
    
        dsm_handle_t address_handle;
        uint32_t status = dsm_address_publish_add(0x100+i, &address_handle); 

        __LOG(LOG_SRC_APP, LOG_LEVEL_INFO, "dsm_address_publish_add status: %d \n", status);
        
        status = access_model_publish_address_set(m_clients[i].model_handle, address_handle);
        __LOG(LOG_SRC_APP, LOG_LEVEL_INFO, "model_pulish_set status: %d \n", status);
        
    }
	
	

```

