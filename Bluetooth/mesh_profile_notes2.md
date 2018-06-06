##SIG 蓝牙 mesh 组成

[TOC]

### mesh网络概述

蓝牙mesh网络是一种基于泛洪管理的mesh网络，消息是通过广播信道传递，收到消息的节点可以继续转发消息，这样就可以实现更远范围的消息传递。为了防止出现消息不受限制的转发传递，规定了以下两种方法来避免：

+ 节点不会转发之前收到的消息，当收到新的消息时，会首先在缓存中检查是否存在相同消息，若存在，则忽略新的消息。
+ 每个消息都会包含一个TTL（Time to Live）的字段，这是用来限制消息中继的次数，每次转发消息后，TTL的值就会减1，当TTL的值到达1时，消息就不会再次被中继。


###  网络和子网
 共享以下几种网络资源的节点组成一个mesh网络:

 + 用来identify消息源地址及目的地址的网络地址
 + Netwoek Key, 用来在网络层加密
 + Application Key， 用来在access layer加密
 + IV index

  > IV Index 是一个32位的值，是一种共享网络资源，比如一个mesh网中的所有节点都共享相同的IV Index值。
  >
  > IV Index从`0x00000000`开始，在IV-Update-procedure过程中递增，并由特定的进程维护，以保证整个mesh网内共享相同的IV。
  >
  > IV Index在网络中通过Secure Network beacons 共享， 一个子网收到 IV-Update时，会处理并传播该update。传播是通过在子网中传输Secure Network beacons实现的。如果一个在主子网中的节点收到IV-update时，会将其传播到其他所有的子网。 
  >
  > 如果一个节点从网络中消失一段时间，该节点会扫描Secure Network beacons ，或者是使用IV Index Recovery procedure来重新设置其IV-Index的值。


一个mesh网络中可以存在一个或多个子网（比如一个宾馆的所有节点构成一个mesh网络，每个房间的节点都构成一个子网），子网中的节点有着同样的Network Key，他们可以在网络层相互通讯。一个节点可以属于多个子网，即一个节点可以配置多个Network Key。在Provision阶段，一个节点会被配到一个子网中，可以通过Configuration Model将节点陪孩子到多个子网中。

子网中有一类特殊的子网，被称为主子网，主子网是基于主NetKey的。主子网中的节点参与IV更新操作，并将IV值传递给其他子网。与之相对的是，其他组网中 的设备只是向子网中传播IV-Index。

包含**Configureation Client Model**的节点来配置网络资源， 这类节点称之为**Configuration Client**。通常情况下，Provisioner 负责在Provision阶段给节点分配unicast address以保证不会有重复的地址。Configuration Client负责分配NetKey, AppKey以保证网络中的节点可以在网络层和access层通讯。

### 设备和节点 devices & nodes

加入mesh网络的设备称之为节点，未配网设备称之为device。** Provisioner** 就是用来管理节点与未配网设备之间的消息传输。

未配网的设备不能收发mesh消息，但是它可以通过广播消息告诉provisioner该设备的存在。provisioner可以将一个未配网设备接入网络使之成为网络节点。

网络节点可以收发mesh消息，其是由**Configuration Client**来管理的（**Configuration Client**可以与**Provisioner**是同一个）。**Configuration Client**是配置节点之间如何传递消息。**Configuration Client**可以把节点从mesh网络中移除，使之成为未配网设备。

### 入网

当设备被**provisioner**添加到网络时，就成为网络节点。设备的配网过程与传统的蓝牙点对点配对过程不同。设备配网过程是通过**advertising bearer **或者**point-to-point GATT-based bearer**实现的。通过**advertising bearer **配网的过程是被所有节点所支持的，而通过**point-to-point GATT-based bearer**配网允许智能手机这类设备成为配网者（provisioner）。

多个节点配网时，**Provisioner**可以在正在配网节点上设置一个attention 定时器，当设置该定时器设置为非零值时，设备可以向外界表现出其正在配网，比如闪灯、响铃、震动等。当定时器设置时间到达时，设备结束闪灯恢复正常。

### mesh中的几个概念

mesh网络结构使用以下几个概念：states、messages、bindings、element、addressing、models、publish-subscribe、mesh keys、association。

#### States（状态）

state是一个用来表示element状态的一个值。

可以表现state的element称之为server。比如，灯控中，灯节点一般为server，其开关状态则是一个state， server根据收到的消息改变其状态。

可以访问state的element称之为client。比如，灯控中的开关，可以通过发送开关消息给server，实现开关控制。

包含多个值得状态称之为复合状态，比如灯的颜色会包含RGB等分量。

#### Bound states（关联状态）

当一个状态关联另一个状态，其中一个改变，会导致另一个状态改变，这种状态就称之为关联状态。关联状态可以存在于一个或多个element的不同models之间。在灯控中，较为常见的关联状态是电平状态跟开关状态的关联，当电平变化到0时，会导致开关状态也变为关。

#### Messages（消息）

消息会作用于状态， 对于每个状态（state）都会定义一系列消息，这些消息被server支持并且可以被client用来获取并改变server的状态。server也可以主动发送自身状态的消息，比如状态改变时。

消息包含操作码、相关的参数以及行为。操作码可以是1~3个字节，比如泰凌微的私有mesh就是用的三个字节（op[3]）表示，其中op[0]为操作码，op[1~2]表示厂商ID。一字节的操作码用于某些特殊情况，比如需要将允许的参数个数设置为最大时；2个操作码时标准的消息。

一条包含操作码的消息的总长度是在下传输层决定的，这里可能会用的分段重组机制。为了最优化性能，我们应避免分段重组，最好将消息限制在一个分段长度以内。传输层允许最多11个字节的不分段消息，当使用一个字节操作码时，参数个数最多允许10个。当操作码为3个时，最多允许8个参数。

传输层的分段重组机制最多允许32个分段，因此消息的最大长度是384个字节。除去MIC检验的四个字节，一个操作码的消息最多允许379个参数。

消息可以是需要回应的，也可以是无需回应的。

#### Element

element是节点中一个可寻址的实体，每个节点都至少有一个element，一个主element，多个附加element。element的数量和结构是固定的，在节点的生命周期中保持不变。

主element使用节点在配网时的唯一地址（unicast address）寻址，每个附加element都使用子序列地址寻址。这些element地址允许节点识别节点内的收发消息。

如果一个节点内，element的数量和结构发生了变化，比如在固件升级时，那么需要再次配网。

models中的消息是根据消息的操作码和element地址来分发的。

一个element中不允许包含多个使用相同消息的model，当一个element中的多个model使用相同消息时，会造成过载。为了避免过载，一个节点内允许多个element存在，这样具备相同消息的model放在不同element中就可以了。

比如，对于一个灯设备来说，可能存在两个灯泡，每个都实现了亮度控制。这时就需要节点包含两个element来分别表示两个灯泡，当收到亮度命令时，节点会根据element的地址来决定哪个灯泡亮度应被修改。

#### Addresses 地址

有三种类型地址，唯一地址（unicast address）、虚拟地址、组地址。

唯一地址是在配网阶段被分配给节点的主element，每个mesh网络可以有32767个唯一地址。

虚拟地址可以代表多个element，每个虚拟地址表示一个标签化的UUID。每个被发往UUID的消息都在完整检验值中包含可整个标签UUID，以验证消息。为了避免检查每个已知的UUID，我们使用UUID的哈希值。共有16384个哈希值，每个都代表了一族虚拟地址。在虚拟地址中，只使用了16384个哈希值，因此每个哈希值都可以表示上百万个UUID，因此虚拟地址的数量可以被认为很大。

组地址可以表示多个element，每个mesh网络中共有16384个组地址。有一些固定的组地址用来访问所有主element。共有256个固定组地址，16128个动态分配的组地址。


#### Models

Model定义了一个节点的基础功能，一个节点可能包含多个Model。一个Model定义了其所需要的状态、作用于状态的消息、以及相关行为。

mesh应用使用了client-server结构，通讯使用“发布-订阅”机制。鉴于mesh网络的特点，以及配网的行为，应用不是简单的“端到端”模式。应用是定义在client model、server model和control model之中。

+ **Server model:** 由一个或者多个element中的多个state组成。Server Model 定义了一族主消息，收发主消息时element的行为，还有收发消息之后的附加行为。

+ **Client model：**定义了一族消息，client可以用来请求、改变相关server的状态。client没有state.

+ **Control model:** 包含client功能用来与其他server通讯，也包含server功能用来与其他client通讯。一个control model可以包含控制逻辑，即与其他model相互作用时，该model的一些规则及表现。

一个简单的设备可以包含server、client、control model。下图表示了一个control model例子。

![ch2_Control_model_comm](pic\ch2_Control_model_comm.png)


照明控制器是一个control model实现，照明控制器需要作为client来控制灯设备（或传感器），也需要作为一个server被智能手机等设备控制。这样一个照明控制器可以存在于传感器中、灯设备中或者是一个独立的设备中。

Models可以定义网络节点的功能，比如秘钥管理、地址分配、消息中继灯。Models也定义了设备的表现，比如电源控制、照明控制、传感器数据收集。有的节点只实现了中继中能，比如代理节点。

一个消息可以应用于多个不同的models，消息的表现在每个model中都是相同的。

Model的特性是不可改变的，不能向一个model移除或者添加行为。如果一个新的行为必须要添加，则需要产生一个新的model。model支持扩展，新的model扩展原model时，会继承原model的相关行为。

一个element所包含的的model决定了该element的表现。

Model可以被SIG定义（SIG Model）也可以被厂商（Vendor Model）定义，model被唯一的标识符所识别，可以是16bit（SIG Model）或者是32bi（Vendor Model）t的一个值。

下面的例子中，一个设备包含了两个element，主element中包含了一个model，该model是扩展自附属model的，在附属model的基础上添加了一个新的状态。

由于State_X1与State_X2可以接收相同的消息，因此必须要放在两个element中。

![ch2_Element-model-structure](pic\ch2_Element-model-structure.png)

###  智能插座例子

![ch2_dual-socket_smart_device](pic\ch2_dual-socket_smart_device.png)

如上图所示，该设备由两个插座组成，包含两个element，分别代表两个插座。每个element都分配一个唯一地址。

每个element的功能由**Generic Power Level Server model**定义，该model定义了一族状态以及作用于状态的消息。

Generic Power Level Set 消息被传送到设备上用来控制电平，然后根据element的地址，来控制对应的插座。

插座也可以被实现了**Generic Level Client model**的设备控制，该model简单的设置电平，插座的实际控制是通过状态关联实现的。在每个插座中，**Generic Power Actual state ** 是与**Generic Level state**关联的。**Generic Level Client **向**Generic Level Server **发送电平设置消息，然后**Generic Level**改变，与此同时**Generic Power Actual state **也会改变，Power Actual state实际控制插座电源。

element可以报告状态，在插座的例子中，每个插座都可以汇报电源状态以统计耗电量。能耗是通过**Sensor Server model**定义的消息实现的，每个消息都有个element地址，这个地址唯一识别插座。

![ch2_dual-socket_element_struct](pic\ch2_dual-socket_element_struct.png)

### 发布订阅与消息交换

节点产生消息并向特定的地址发布该消息，订阅相应地址的节点会处理消息。

消息可以发往唯一地址、预分配的组地址或者是虚拟地址，消息作为响应消息被发出，也可以作为独立的消息。当作为响应消息被发出时，model将会使用收到消息的源地址作为该响应消息的目的地址。当model传输独立消息时，使用该model的发布地址作为目的地址。每个model都有一个发布地址。

在接收端，每个model都可以订阅一个或多个组地址或是虚拟地址，如果收到消息的目的地址在本model的订阅列表中，就会处理该消息。当消息的目的地址是element的唯一地址时，也会被处理。

model的发布地址和订阅列表时在Model-publication-and-Subscription-List-state中定义的，这个state被Configuration Server Model所管理。

每个消息都是从一个确定的唯一地址发出，而且有一个唯一的序列号用来避免中继攻击。

### 加密安全

所有的消息都使用两种类型的秘钥，一种是用来网络层加密的，称之为"NetKey"，一种是应用秘钥，称之为"AppKey"。网络密钥和应用秘钥可以实现敏感消息与非敏感消息的分离。比如，中继节点可以解密网络层消息，而不会影响到应用数据，在智能家居的mesh环境里，一个灯控设备可以中继门锁设备的消息，但是由于没有应用秘钥，所以不会改变门锁消息，故而不会对门锁产生影响。

