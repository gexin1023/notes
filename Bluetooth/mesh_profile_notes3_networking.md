## 3. Mesh Networking

本部分以mesh网络的分层结构的顺序自下而上地介绍mesh网络。mesh网络结构如下所示：

![ ](pic/layer_of_mseh.png  "mesh网络结构")


### 3.1 Bearers 承载层

本规范定义了两种承载层：

+ **Advertising bearer**
+ **GATT bearer**

#### 3.1.1 Advertising bearer

使用 advertising bearer 时，mesh数据包可以使用Advertising Data发送，BLE advertising PDU使用 Mesh Message AD Type标识。

| Length | AD Type | Contens |
| - | - | - | 
| 0xXX | Mesh Message | network PDU |

任何使用Mesh Message AD Type的广播消息应该是无需连接（ non-connectable）、无需扫描的（ non-scannable） 非直接广告事件。如果一个节点在一个连接的或者扫描的广告事件中收到了一个Mesh Message AD Type 消息， 那么该消息会被忽略。

一个只支持 advertising bearer的设备应该使用尽可能高的占空比（接近100%）来扫描消息，以避免遗失消息或者Provisioning PDUs。

所有设备都应该支持GAP Observer role 和 GAP Broadcaster role。

#### 3.1.2 GATT bearer

GATT bearer可以使那些不支持advertising bearer的设备可以加入到mesh网中，GATT bearer 使用Proxy protocol通过GATT连接在设备之间转发、接受Proxy PDUs。

The GATT bearer uses a characteristic to write to and receive notifications of mesh messages using the attribute protocol.（这句话没理解什么意思）

GATT bearer定义了两种角色，分别是Client 和 Server。GATT Bearer Server应该实例化一个且只能一个 Mesh Proxy Service，GATT Bearer Client 应该支持Mesh Proxy Service。

### 3.2  Network Layer 网络层

网路层定义了可以使Lower Transport PDUs被bearer层转发的 Network PDU格式。网络层将从input interface收到的incoming消息进行解密（decrypt）、授权(anthenticate)并转发向output interface 或是更高的层级；将outgoing消息进行加密、授权并转发到其他网络接口。


#### 3.2.1 字节序

该层使用大端字节序

#### 3.2.2 地址

地址是16bit长度的值(两个字节），如下所示：

| 地址二进制值 | 地址类型 |
| - | - |
| 0b0000000000000000 | 未分配地址 |
| 0b0xxxxxxxxxxxxxxx（0x0000除外） | unicast address |
| 0b10xxxxxxxxxxxxxx | virtual address |
| ob11xxxxxxxxxxxxxx | group address |


##### 3.2.2.1 未分配地址

未分配地址是当一个节点的element还没有配置或者还没有分配地址时的一个地址。当不需要发布消息时，可以将publish address设置为unassigned address。

##### 3.2.2.2 Unicast address

unicast address是分配给每一个element的唯一地址，取值范围是0x0001到0x7FFF。unicast address在一个节点的生命周期中保持不变。unicast地址被用作消息的源地址，也可能用于消息的目的地址。如果一个消息是发往一个unicast 地址，那么该消息最多被一个element处理。

##### 3.2.2.3 虚拟地址

虚拟地址代表一系列的目的地址，每一个虚拟地址逻辑上代表一个128-bit的标签UUID。一个或多个element可能发布或订阅一个标签UUID， 标签UUID不被传递，应该被当作消息完整性检查值的附加域。

虚拟地址的15bit被设置为1, 14bit设置为0, 13~0bit是一个hash值。这个哈希值派生于UUID。

当一个发往虚拟地址的Access消息收到时，每一个匹配该虚拟地址的UUID都会被upper transport层用来当作附加唉的数据作为认证消息的一部分。

控制消息(control message)不可以使用虚拟地址。

##### 3.2.2.4 group address

Group address 是一个被写入0个或多个element的地址。地址位的15bit 14bit均被设置为1.

Group address只能被用作目的地址，被发往group address的消息被发送到所有订阅该group address的modale interfaces。

group address分为两种，一种是动态分配的，另一种是固定的。

| 地址值 | Fixed Group Address |
| - | - | 
| 0xFF00-0XFFFB | 	RFU	|
| 0xFFFC | all-proxies |
| 0xFFFD | all-friends |
| 0xFFFE | all-relays |
| 0xFFFF | all-nodes |


### 3.2.3 Address validity

| 地址类型 | 源地址 | 目的地址
| - | - | - | - | 









