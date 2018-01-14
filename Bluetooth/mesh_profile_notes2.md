## Mesh Architechture

蓝牙mesh的结构如下图所示：

![](/home/gexin/repo/notes/Bluetooth/pic/layer_of_mseh.png) 

蓝牙mesh网络是一种基于泛洪管理的mesh网络，消息是通过广播信道传递，收到消息的节点可以继续转发消息，这样就可以实现更远范围的消息传递。为了防止出现消息不受限制的转发传递，规定了以下两种方法来避免：

+ 节点不会转发之前收到的消息，当收到新的消息时，会首先在缓存中检查是否存在相同消息，若存在，则忽略新的消息。
+ 每个消息都会包含一个TTL的字段，这是用来限制消息中继的次数，每次转发消息后，TTL的值就会减1，当TTL的值到达1时，消息就不会再次被中继。

 
 共享以下几种网络资源的节点组成一个mesh网络:
 
 + 用来identify消息源地址及目的地址的网络地址
 + Netwoek Key
 + Application Key
 + IV index
 
 一个mesh网络中可以存在一个或多个子网，子网中的节点有着同样的Network Key，他们可以在网络层通讯。一个节点可以属于多个子网，即一个节点可以配置多个Network Key。在Provision阶段，一个节点会被配到一个子网中，可以通过Configuration Model将节点陪孩子到多个子网中。
 
 子网中有一类特殊的子网，被称为主子网，主子网是基于主NetKey的。主子网中的节点参与IV更新操作，并将IV值传递给其他子网。
 
 网络资源是通过实现了的Configureation Client Model的节点来配置的。通常情况下，Provisioner 负责在Provision阶段给节点分配unicast address以保证不会有重复的地址。Configuration Client负责分配NetKey, AppKey以保证网络中的节点可以在网络层和access层通讯。
 
 