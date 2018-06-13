## SIG-mesh与Telink-mesh简单介绍

SIG-mesh与telink-mesh都采用泛洪管理的消息机制在广播信道收发消息。telink-mesh的实现方式较为简单，主要是根据用户名与密码来甄别同一mesh网内节点消息。SIG-mesh采用Network-Key与Application-Key进行两级加密，实现方式相对于telink-mesh更加复杂，但是适用性、安全性更好。

### telink-mesh实现方式

由于telink-mesh协议栈代码并未开源，所以telink-mesh实现方式是根据SDK中应用层API及应用代码推测的。。。

数据传输是在广播信道进行的，即一个节点发出的消息所有节点都可以收到。在一个节点收到的消息中，有非mesh蓝牙节点的广播数据、mesh网内节点的有效数据、mesh网外节点的无效数据。

在进来的消息中，如何识别本节点所属mesh网内的消息呢？telink-mesh是这样实现的，在包数据的特定位置插入用户名、密码，这样用户名密码不匹配的消息就被过滤掉了。使用相同用户名密码的节点，构成一个mesh网。**telink-mesh的组网过程，只是修改节点的用户名。**

出厂时，所有的节点都设置相同的默认用户名密码，用户使用时，会通过APP注册一个唯一的用户名（比如 g@qq.com）。app首先通过出厂默认用户名密码连接节点，然后将其修改为app用户名。通过这种方式，该app连接的节点共享一个用户名，构成一个mesh网。


### SIG-mesh实现方式

SIG-mesh是基于BLE5.0实现的，其网络层级结构自下而上一次是：

BLE协议栈 -> **bearer_layer** -> **network_layer** -> **transport_layer** -> **access_layer** -> **model_layer** -> **application**。

除了最下层的BLE协议栈，nordic都将其实现代码开源。

SIG-mesh的组网由**provision**和**configure**两个动作实现。

provision过程是给节点分配Network_key、IV_Index、Unicast_address，该过程将未配网设备加入到网络中，成为网络节点。

	SIG_mesh中，mesh网内节点共享相同的Network_key和IV_Index，其中IV_Index每隔30~60分钟会加1，网内节点同步更新IV_Index。
	一个节点可以配置多个Network_Key，即属于多个子网。

初始时，未经过provision的设备，会不断发出unprovision_beacon帧，beacon帧中的数据包含该设备的UUID。Provisioner如果扫描到该beacon帧时，会根据其中的UUID信息，决定是否provision该节点。之后，provisioner会选择一个UUID的设备，向该节点发送Network_Key、IV_Index等信息。设备获得正确的Network_Key、IV_Index时，就算是加入mesh网了，节点之间可以转发本mesh网内消息。

	UUID是一个16字节（128 bit）的数据，可以在UUID的特定位置做标记，比如灯、插座、开关、公司等。这样在provision阶段就可以获取扫描的设备信息。

经过provision阶段，节点已经组网了，相互之间可以转发消息。这时节点还没有地址、Application_Key等信息，还需要configure阶段才能构建完整的mesh网络。

Configure过程就是配置节点地址及Application_Key。

根据实际情况分配group_address及virtual_address

	element是SIG_mesh中的一个基本概念，举个例子，一个智能插排设备会有六个插座位置，这六个插座位置都可以接收开关消息，而我们需要独立控制每个插座位置，所以不能只给这个插排分配一个unicast_address。要给六个插座位都分配一个唯一的地址。每个插座位就是一个element。可以接收相同消息的单元（这里是插座位），需要分配在不同element种，以便独立控制。

分配Application_Key，是针对不同应用而言，比如门锁与灯。在实际应用中，一个家庭的所有mesh设备构成一个网络，其中包括门锁、插座、灯等设备，处于一个mesh网可以互发消息，但是为了防止灯的消息对门锁造成影响，因此需要对不同的应用分配不同的application_key。这样门锁根本就不会去解析灯的消息，但是由于共享相同的Network_Key，门锁可以转发灯节点的消息。






