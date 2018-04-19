## STM32_UART 中断方式接收数据(基于HAL库)

### 1. 使用STM32Cube生成工程


### 2. 使能uart中断


### 3. 数据接收

- HAL_UART_Receive_IT() 中断方式接收数据

- 该函数设置了接收缓存地址（huart->pTxBuffPtr = aRxBufer）、缓存长度(huart->TxXferSize)、缓存计数（huart->TxXferCount）、接收中断使能（USART_CR1_RXNEIE）。

- 真正的数据接收是在接收到数据时进入中断接收的（每收到一个字符进入一次中断）。在中断处理函数中，每接收一个字符，huart->pTxBuffPtr 指针后移一个，同时huart->TxXferCount减小1。当TxXferCount减小为0时，进入回调函数HAL_UART_RxCpltCallback()，处理缓存的数据。处理完缓存后，pTxBuffPtr是指向缓存区最后的，为了防止内存泄漏需要再次调用HAL_UART_Receive_IT()，重置缓存地址、缓存计数等内容。

- 必须注意，接收回调函数HAL_UART_RxCpltCallback()是在接收到与接收缓存长度相同数目的字节时，才会触发，而不是收到一个字符串就触发。