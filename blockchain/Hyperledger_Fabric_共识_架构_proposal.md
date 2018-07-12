# Hyperledger Fabic——共识、结构、Proposal

Hyperledger Fabric中的节点分为*peer*和*orderer*两个角色。*peer*负责账本管理、chaincode执行、block验证等，orderers负责transaction的排序。除此之外还引入了*endorsing peer*，endorser是一种特殊节点，用来给要执行的proposal背书并模拟chaincode执行。

## 1. 系统结构

区块链是一个有很多节点组成的分布式系统，节点间可以相互通信。fabric可以运行chainnode、运行transaction、保存状态及账本数据，这里的chaincode就是智能合约。Transaction必须被endorsed才可以被commit，并写入到账本上。fabric中还存在一个或多个特殊的chaincode（系统chaincode)用来管理系统参数。

### 1.1 Transactions

Transaction可以分为两种类型：

+  **Deploy Transactions** 创建新的chaincode并将其作为程序的参数，当一个deploy transaction运行成功时，chaincode就被安装在了区块链上。
+  **Invoke Transaction**在某个chaincode的上下文中执行操作。Invoke Transaction指向一个chaincode并与该chaincode的一个处理程序关联。当运行成功时，chaincode执行特定的函数来修改状态并返回结果。

实际上，deploy transaction是一种特殊的Invoke transaction，deploy transaction通过系统chaincode来创建新的chaincode。

### 1.2 Blockchain datastructures

#### 1.2.1 State

区块链的状态可以用用key-value的方式来存储，及key-value store。这些存储单元的实体可以用chaincode通过`get-value`、`put-value`的操作来管理。

通常情况下，状态`s`用这样一个map模型`K -> (V*N)`来表示。

+ `K`是一组key的集合
+ `V`是一组value的集合
+ `N`是有限的、有序的version集合。比如， `next: N -> N'`，takes an element of 'N' and return the next version number。

+ `put(k,v)`（`k属于K`，`v属于V`），将区块链的状态从`s`变化到了`s'`，`s'(k)=(v, next(s(k).version))`
+ `get(k)`返回`s(k)`

区块链的状态（state）是由peer通过chaincode管理的，而不是client，也不是orderer。

#### 1.2.2 Ledger

Ledger（账本）提供了一个可以验证的状态变化的历史记录，这包括在系统运行时状态的变化以及没有成功的对状态改变的尝试。

账本是由ordering-service所打包生成的，是由transaction组成的block所形成的全局排序的hash链。哈希链固定了账本中区块的顺序，每个区块中都包含了一组经过排序的transaction。通过区块链的结构，所有的交易顺序都是固定的，不可修改的。

账本是在所有peer上保存的，在部分orderer中也有保存。保存在Orderer中的账本我们称之为"Orderer-Ledger"，与之对应的Peer中的账本称之为"Peer-Ledger"。

### 1.3 Node（节点）

区块链中的节点可以相互通信，多个节点可以在一个server上运行。在Fabric中有三种不同的节点：

+ **Client** 向endorse-peer提交实际的事务处理申请，并等待返回，然后将proposal组成transaction之后发送给order-service进行排序。
+ **Peer**对交易进行commit，Peer包含了账本及state。endorse-peer是一种特殊的peer用来对交易进行背书并模拟运行。
+ **Orderer**进行transaction全局排序，将区块进行原子广播。

#### 1.3.1 Client

Client表示一个终端用户，它必须与peer相连以保证与区块链的通信。

Client既可以与peer通信，也可以与orderer通信。


#### 1.3.2 Peer

一个peer节点接收经过ordering-service排序的区块，并管理本peer上的state和账本。

Peer可以实现一个特殊endorse角色，可以称之为endorser。由client发起的proposal，首先会到endorse-peer尽心背书，并模拟执行。每个chaincode都会定义一个endorsement策略，应用于一组endorser节点上。背书策略规定了endorsement完成的一些要求，比如必须超过多少个endorer确认才算成功，或者必须经过某几个的endorser的背书才算成功。

#### 1.3.3 Ordering Service nodes

Orderer实现了区块的可靠传送，ordering-Service可以实现为中心化的服务（solo），或者是分布式协议（kafka）。

Ordering-service向client和peer提供一个共享的通信channel，该channel提供一个广播服务，可以用来传递包含transaction的消息。Client连接到这个channel，可以在这个channel上广播消息，这会传送到所有的peer节点。channel 支持对所有类型的消息进行原子传输，也就是说，该channel向所有的peer节点传送一组消息，所有的peer都以同样的顺序收到这组消息。这种元素通信，保证了全局排序的广播，及原子广播。

Ordering-Service可以支持多个channel，就像Kafka中多个topic一样。可以将channel看做是不同的分区，处于不同channel的peer是不能相互通信的。Client连接到一个channel上之后，不会知道另一个channel的存在，而Client可以连接多个channel。


## 2. Basic workflow oftransaction endorsement

### 2.1 Client 创建transaction并向endorser发送

开始一个transaction，client会首先向endorser节点发送`Propose`消息，在特定channel上的endoring-peer节点会收到该消息。

#### 2.1.1 `Propose`消息格式

`Propose`消息格式为`<Propose, tx, [anchor]>`,其中`anchor`是可选的。

-  `tx = <clientID, chaincodeID, txPayload, timestamp, clientSig>`
	- `clientID`是发起该proposal的client的ID
	- `chaincodeID`该transaction需要执行的chaincode
	- `txPayload`是发出的transaction的payload
	- `timestamp`client维护的一个时间戳
	- `cilentSig`client的signature
	
	对于invoke transaction和deploy transaction，其txPayload是不一样的。
	
	For **invoke transaction**:
	
	- `txPlayload = <operation, metadata>`
		- `operation` denotes the chaincode operation (function) and arguments
		- `metadata` denotes attributes related to the invocation.
	
	For  **deploy transaction** :
	
    - `txPayload = <source, metadata, policies>` 
    	- `source` denotes the source code of the chaincode
    	- `metadata` denotes attributes related to the chaincode and application
    	- `policies` contains policies related to the chaincode that are accessible to all peers, such as the endorsement policy.
- `anchor` contains _read version dependencies_, or more specifically, key-version pairs (i.e., `anchor` is a subset of `KxN`), that binds or "anchors" the `PROPOSE` request to specified versions of keys in a KVS (see Section 1.2.). If the client specifies the `anchor` argument, an endorser endorses a transaction only upon _read_ version numbers of corresponding keys in its local KVS match `anchor` (see Section 2.2. for more details).

### 2.2 Endorsing peer 模拟交易并产生endorsement signature

endorsing-peer收到proposal消息之后，首先验证client的签名`clientSig`，然后模拟transaction。

模拟交易运行时会根据`chaincodeID`和当前的state来执行一个chaincode。





















