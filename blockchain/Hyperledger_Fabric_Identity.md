## Hyperledger Fabric中的Identity

### 什么是Identity

区块链网络中存在如下的角色：peers， orderers, client application, administrators等等。每一个这样的角色都有一个身份标识（Identity），该身份标识是通过X.509 数字证书来表示的。这些身份标识决定了该角色的对区块链网络上资源的权限，比如是否有权限访问区块链上的某种信息。

数字身份有很多附加属性，供fabric来判断权限。数字身份给出了一个身份的组合结构，与之相关的属性称之为principal。Principals 就像用户ID或者是群组ID，但是更加复杂，因为principal中包含了该角色的一系列属性信息。当我们谈论principal的时候，就是在说觉得角色权限的各种属性信息。

为了保证身份（identity）是可以被验证的，Identity必须来自于一个受信任的颁发机构。在Fabric中，这是通过membership service provider (MSP) 来实现的。MSP是Fabric中的一个组件，它定义了管理有效identity的规则。Fabric中默认的MSP实现是使用X.509证书作为identity，采用传统的PKI（Public Key Infrastructure ）结构模型。

### 说明Identity使用的简单场景

假设你正在逛超市买东西，在买单时发现收银台只支持银联和visa的银行卡。这时如果你想使用一张不属于银联和visa的卡来支付，无论你的卡中是否有足够的余额，这都是不被收银员接受的。

有一张有效的信用卡是不够的，还需要被超市所支持。PKI和MSP以相同的方式运作，PKI提供了一系列identity，而MSP来指明哪一个identity才是区块链网络的参与者。

PKI证书颁发和MSP提供了相似的功能，PKI就像一个银行卡发行机构，它分发很多不同类型的可以验证的Identity。MSP就像被超市所接受的银行卡，决定哪些Identity是可以信任的区块链网络成员。MSP讲可验证的Identity转化为区块链网络的成员。

### 什么是PKI

A public key infrastructure (PKI)是一组提供网络通讯安全的信息技术。https中的"s"字母就是由PKI来实现的，在传统的http上加入PKI技术来实现网络通信安全。

PKI由Certificate Authorities（CA） 组成，它向各方(例如，服务的用户、服务提供者)颁发数字证书，然后使用它们在与环境交换的消息中进行身份验证。一个Certificate Revocation List (CRL) 中包含了那些已经失效的证书的引用。证书可以被撤销，比如当一个与证书关联的私有加密材料泄露时，该证书应该被撤销。

尽管区块链不仅仅是一个通讯网络，但是它也是依赖于PKI来实现在网络参与者之间的安全通讯的。理解基本的PKI与MSP，对于理解区块链中消息的传输是十分重要的。

PKI有一下四个关键元素：

+ Digital Certificate 数字证书
+ Public and Private Keys 公钥、私钥
+ Certificate Authorities 证书颁发机构
+ Certificate Revocation List 证书撤销列表

关于PKI的介绍可以看看Wikipedia的PKI词条。

####  Digital Certificates 数字证书

数字证书是有个文档，其中包含了一系列证书中包含的属性。最常见的数字证书类型是符合X.509标准的证书，该标准允许在其结构中编码identity细节。

比如说，位于底特律的某个汽车制造厂家可能在其数字证书中会包含诸如地点、行业、UID等一系列信息。数字证书就像身份证一样，记录了所有者的关键信息。在X.509证书中，有许多属性信息，我们仅仅看一下如下图所示的几个。


![identity.diagram](pic\identity.diagram.jpg)

上图这个数字证书描述了一个名为Mary Morris的组织，Mary是证书的所有者（主人），黑色加粗的一行文字描述了Mary的关键信息。该证书也描述了很多其他信息。更重要的是，Mary的公钥是随着证书分发的，而私钥却不是。

Mary的所有属性都可以用密码学加密记录，以防止被篡改。加密学允许Mary提供他的数字证书给其他人验证他的身份，只要其他人相信证书的颁发机构（CA）。只要CA保证这种加密信息的安全，任何读取该证书的人都可以确信Mary的信息没有被篡改过。可以将Mary的X.509证书看做是不会被篡改的数字身份。

#### Authentication, Public keys, and Private Keys

在安全通讯中，身份认证和消息完整性是两个很重要的概念。身份认证是说消息交换的双方知道是哪个发来的消息。而消息完整性是说，消息在传输过程中没有被破坏。比如说，你会去确认跟你交易的是否就是Mary而不是张三。Mary发给你一个信息，你想确认信息在传输中没有被张三更改。

传统的认证机制是基于数字签名的，允许用户对消息进行数字签名。数字签名也可以保证签名消息的完整性。

技术角度来看，数字签名需要各方都持有两个加密秘钥： 一个是公钥，可以广泛使用并朝哪个档身份验证锚；另一个是私钥，用于在消息上生成数字签名。数字签名的接受方可以通过检查其附属签名在于其发送方的公钥下是否有效来验证消息的来源和完整性。

私钥与公钥关系的唯一性使得加密的消息安全传输成为可能。唯一性的数学关系是这样的，私钥可以用来在消息上产生只符合特定公钥的签名，而且只在该消息上符合。


#### Certificate Authorities

一个节点加入区块链网络是通过节点从一个受信任的机构那里获取一个数字身份来达到的。在多数情况下，数字身份是以加密的数字证书（X.509）表示，该证书由Certificate Authorities（CA）所颁发。

CA是网络安全协议的一部分，你可能听过一些，比如：Symantec (originally Verisign), GeoTrust, DigiCert, GoDaddy, and Comodo等等。

CA向不同的角色分发证书，证书是被CA数字签名的，并与角色通过公钥绑定在一起的。因此，一个人如果信任CA（知道CA的public key），那么他就可以信任与该公钥绑定的角色。

证书可以被广泛的传播，因为证书中不包含角色或者是CA的私钥。

CA也有证书，它们可以广泛地提供证书。这允许给定CA颁发的标识的使用者通过检查证书是否只能由相应的私钥(CA)的持有者生成来验证它们。

在区块链网络中，每个角色如果想与网络中互动，都需要一个数字身份。您可能会说，可以使用一个或多个CA从数字的角度定义组织的成员。它是为组织的参与者提供可验证的数字身份的基础的CA。

####  Root CAs, Intermediate CAs and Chains of Trust

CA可以分为两种：Root CA和Intermediate CA。因为Root CAs (Symantec, Geotrust等)必须安全地向网络用户分发数以亿计的证书，所以将这个过程扩展到所谓的 Intermediate CA是有意义的。这些Intermediate CA由 Root CA或其他中介机构颁发证书，允许为链中的任何CA颁发的任何证书建立“信任链”。这种可以追踪回Root CA的能力，不仅让CA在提供安全的功能的同时进行扩展（这种方式允许组织受信任地使用Intermediate CA），而且CA链也避免了Root CA的泄露，ROOT CA泄露会危及整个链的信任。另一方面，如果一个中间CA被破坏了，暴露量就会小得多。

![CA_chain](pic/CA_chain.jpg)

如上图所示，信任链是通过一个ROOT CA和一系列intermediate CA来建立的。每个CA都可以签署新的CA来构成信任链的一部分。


当涉及到跨多个组织颁发证书时，中间的CAs提供了大量的灵活性，这对于许可的区块链系统(如Fabric)非常有用。例如，您将看到不同的组织可能使用不同的根CAs，或者相同的根CA使用不同的中间CAs——它确实取决于网络的需要。

#### Fabric CA

CA太重要了，所以在Fabric中提供了一个内建的CA组件，以在区块链中创建CA。这个组件就是Fabric CA，这是一个私有的Root CA提供者，可以用来管理网络成员的数字身份（通过X.509证书）。因为Fabric CA是针对Fabric的根CA需求的定制CA，所以它本质上不能为浏览器中的通用/自动使用提供SSL证书。但是，由于某些CA必须用于管理标识(甚至在测试环境中)，所以Fabric CA可以用于提供和管理证书。

#### Certificate Revocation Lists

Certificate Revocation Lists（CRL）很容易理解，他就是一系列已经被撤销的证书的引用。

如果一个第三方想去验证另一方的身份时，他首先会检查CA中的CRL是否有该证书，已确认该证书没有被撤销。
