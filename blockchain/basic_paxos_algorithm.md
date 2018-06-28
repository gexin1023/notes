## Paxos共识算法

---
paxos是一族用来解决分布式系统共识的基础算法，共识过程就是在一组节点上达成一个一致的结果。由于节点可能会错误，通讯消息也可能会丢失，所以建立共识是一个比较复杂的过程。

### paxos算法的假定

#### Processors（可理解为节点）

+ Processor以任意速度运行
+ Processor可能会出错
+ Processor失败失败后会重新配置恢复到网络中
+ Processor不会撒谎或者是违反协议，**即不会发生拜占庭错误**

#### Network

+ 一个节点可以给任何其他节点发送消息
+ 消息可以被异步发送，传输过程可以花费任意时间
+ 消息可能会丢失、重复或者是顺序错乱
+ 消息在传输过程中不会被篡改、破坏，**即不发生拜占庭错误**

#### 节点数量

一般情况下，共识算法用$2f+1$各节点，以保证$f$个节点发生错误时，系统依然可以正常运作。

### Roles(角色)

Paxos算法中根据不同节点的行为将其分为不同的角色：client、proposer、acceptor、learner、leader。在算法实现中，一个节点可以承担一个或多个角色。

+ Client

    Client向分布式系统发起一个请求，并等待回应。比如，在分布式文件系统中，发起一个写文件的请求。

+ Acceptor

    Acceptor被分成组（Quorum），每个组中包含大多数的Acceptor。任何发送给某个acceptor的消息，都必须给组内的任意节点都发该消息。如果一个acceptor收到一个消息，但是该消息的副本没有发送给组内所有的acceptor，那么该消息将会被忽略。

+ Proposer

    提出一个倡议的值，并试图让acceptor在该值达成一致，当出现冲突时，会承担一个协调者的角色

+ Learned

    当达成一致时，动作的执行者。当一个client的请求，被acceptor一致接收，那么learner会执行该请求，并返回。

+ Leader

    一个卓越的Proposer用来推进达成共识的过程。很多节点都可能认为他们自己是leader。


####  Quorums

理解为一个组，这个组里包括了大多数（超过半数）的acceptor，这样任何两个Quorum都会至少有一个共同的节点。比方说，节点{A,B,C,D}中，然和三个节点都可以组成一个Quorum。可以给每个节点一个权重，一个Quorum中所有节点的权重之和大于50%。

### Basic Paxos

共识的过程一般是这样的，client向分布式系统发起一个请求，然后proposer将该请求（proposal）发送给acceptor，当取得一致时候，learner来执行请求。

Proposal的形式是这样的，用一个整数N表示其ID，每个节点发出的proposal的ID是不断增大的，然后proposal的提议值用value表示。

共识过程的建立有以下四个步骤完成

#### 1. Prepare

Proposer先将proposal的ID（用N表示）发送给一个Quorum的acceptor（即发送给大多数节点）。

#### 2. Promise

Acceptor收到Prepare消息，如果消息中的N是目前为止收到的最大的值，那么就会返回一个Promise消息。如果之前有收到更大的N值，那么本次收到的Prepare消息便被忽略。如果本Acceptor节点之前有接受的proposal，那么会在返回的promise消息上，加上之前已接收的proposal的N与value值。

#### 3. Accept Request

如果Proposer收到足够多的promise，那么就需要给该proposal设置value值。如果在收到的promise消息中，有已经被Acceptor接受的proposal，那么会从中选出N值最大的proposal，并用其中的value值设置本次proposal。

然后proposer给一个Quorum的Acceptor发送“Accept Request”消息，消息中包含了本次proposal的ID及value。

#### 4. Accepted

当一个Acceptor收到一个“Accept Request”消息时，只要改accept还没有promise更大的ID的proposal，那么就必须接受该proposal。注册该proposal的Value值，发送一个Accepted消息给Proposer和每一个Learner。

需要注意的是，Acceptor可以接受proposal，这些proposal可能有着不一样的value值。但是，Paxos协议可以保证最终只会在一个值上达成一致。

Paxos过程的消息流程如下图所示：

```
Client   Proposer      Acceptor     Learner
   |         |          |  |  |       |  |
   X-------->|          |  |  |       |  |  Request
   |         X--------->|->|->|       |  |  Prepare(1)
   |         |<---------X--X--X       |  |  Promise(1,{Va,Vb,Vc})
   |         X--------->|->|->|       |  |  Accept!(1,Vn)
   |         |<---------X--X--X------>|->|  Accepted(1,Vn)
   |<---------------------------------X--X  Response
   |         |          |  |  |       |  |
```




