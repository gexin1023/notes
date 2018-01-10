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


### 3.2.3 Address validity 地址有效性

| 地址类型 | 源地址 | control message 目的地址 | access message 目的地址 |
| - | - | - | - | 
| 未分配地址 | NO | NO | NO |
| unicast address | yes | yes |yes |
| virtual address | no | no | no |
| group address | no | yes | yes |

下表表示地址是否可以使用device key 或者是 application key： 

| 地址类型 | Device Key 有效 | Application Key 有效 |
| - | - | - |
| unassigned address | no | no |
| unicast address | yes | yes |
| virtual address | no | yes |
| group address | no | yes |


### 3.2.4 Network PDU

Network PDU的结构定义如下表所示：

| filed | bits | notes |
| - | - | - |
| IVI | 1 | Least significant bit of IV Index |
| NID | 7 | Value derived from the NetKey used to identify the Encryption Key and Privacy Key used to secure this PDU |
| CTL | 1 | Network control |
| TTL | 7 | Time to live |
| SEQ | 24 | Sequence Number |
| SRC | 16 | 源地址 |
| DST | 16 | 目的地址  |
| TransportPDU | 8 to 128 | 传输单元 |
| NetMIC | 32 or 64 | 网络层信息完整性检查 |

#### 3.2.4.1 IVI

IVI是 用来认证加密Network PDU的IV Index的最低位。

#### 3.2.4.2 NID

 NID域包含一个7bit的 network identifier，用来提供一种简单的方式查找加密认证Network PDU所使用的Encryption Key和Privacy Key。

NID是派生于Network Key，与Encryption Key和Privacy Key关联。

#### 3.2.4.3 CTL

CTL位判断消息是否是控制消息，当该位是1时，表示消息是control message。该位为0,则表示消息位access message。

当CTL为0时，NetMIC是32-bit的值，下传输层包含access message。

当CTL为1时，NetMIC是64-bit的值，下传输层包含control message。

#### 3.2.4.4 TTL

TTL是一个7-bit的值，表示消息跳转的次数。

	0=没有被中继且不会被中继。
	1=已经被中继过，不会再次中继
	2~126 = 可能被中继过，仍将继续中继
	127= 没有被中继，可以被中继
	
#### 3.2.4.5 SEQ

该成员是一个24-bit的值，由IV index组成，对于每一个network PDU来说，这是一个由节点产生的唯一的值。



#### 3.2.4.6 SRC

源地址，必须是unicast address。可以根据源地址识别产生该消息的element。

源地址由产生该消息的element设置，并且在传输过程中不会被中继节点接触（可以理解为对中继节点不可见？）。

#### 3.2.4.7 DST

目的地址，16-bit值，可以时unicast address、virtual address、group address。

在传输过程中，不会被中继节点的网络层接触（理解为对中继节点网络层不可见）。

#### 3.2.4.8 Transport PDU

传输的字节数据，当CTL设置为1时，该域最大96 bit。当CTL为0时，最大长度为128 Bits。

Transport PDU 是被产生该消息的下传输层设置，不能被网络层改变。

#### 3.2.4.9 NetMIC

该域长度取决于CTL，当CTL为0时，该域为64-bit；当CTL为1时，该域为32-bit。

NetMIC用于确认DST和Transport PDU没有被破坏。

NetMIC会被每一个传输、中继该消息的节点的网络层设置。

#### 3.2.5 Network Interfaces

网络层支持通过多个承载层来收发消息，每个bearer可以通过network interfaces与网络层连接。同一个节点内部的不同element间的消息传送是通过local interfaces实现的。

举个例子，比如一个节点可能存在三个interfaces，一个用来通过advertising bearers收发消息；另外两个通过GATT bearers收发消息。

Interface 可以提供filter（过滤器）用来控制消息的进出规则。

#### 3.2.5.1 Interface input filter

用来确定进来的消息是丢弃还是传送到网络层。

#### 3.2.5.2 Interface output filter 

用来确定出去的消息是被丢弃还是被传送到承载层。

当TTL的值为1时，会丢弃所有要传送到承载层的消息。

#### 3.2.5.3 Local network interface

这用来在一个节点内部的不同element之间传送消息。当该interface受到消息时，会将消息传送到节点内的所有element.

#### 3.2.5.4 Advertising bearer network interfaces

该interface允许通过advertising bearer传送消息。

当收到一个消息且消息没有被标记中继消息时，Advertising bearer network interface会使用Network Transmit state的值在advertisng bearer上传送该消息。

当收到一个消息且被标记为中继消息，Advertising bearer network interface会使用Relay Transmit state的值在advertisng bearer上传送该消息。

### 3.2.6 Network layer behavior

#### 3.2.6.1 Relay feature

中继特性指的是将从advertising bearer收到的Network PDU进行转发。该特性是选配的，可以启用或者不启用。如果支持代理特性，那么节点应该同时支持GATT和advertising两种bearer(承载层)。

#### 3.2.6.2 Proxy feature

用来转发在GATT和advertising bearer之间传送的消息。

#### 3.2.6.3 接收network PDU

消息是通过network interface从bearer layer 传送到network layer。网络层可以对消息标记一些附加的标签，以供后续使用。

当网络层收到消息时，会首先检查NID的值是否匹配已知的NID值，若不匹配则忽略该消息。若可以匹配已知的NID，则节点会根据相匹配的Network Key来认证该消息。如果认证成功，而且SRC、DST是有效的，且网络层消息缓存中没有该消息，那么消息会被传送到下传输层被继续处理。

消息被转发时，其中的IV index应该与收到时保持一致。

如果从advertising bearer传送过来的消息被传送到下传输层处理，且节点支持中继特性（并使能），TTL不小于2，目的地址不是本节点的unicast address，那么TTL的值减1，Network PDU被标记为relay，Network PDU被重发向所有连接到advertising bearer的网络接口（network interfaces）。建议在收到Network PDU后，加一个任意时间的小延时，再进行转发。这样可以避免多次中继同时发生。


如果从GATT承载层传过来的消息被传送到下传输层处理，节点支持并使能了代理特性，TTL不小于2，目的地址不是本节点的unicast，那么TTL减1，并且Network PDU重发到所有的network interfaces。

如果从advertsing bearer传过来的消息被下传输层处理，并且代理特性被支持使能，TTL不小于2，目的地址不是本节点的unicast address，那么TTL减1，并且Network PDU将会被重发到所有连接GATT bearer的网络接口。










