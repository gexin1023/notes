## PKI（Public Key Infrastucture）介绍

**根据Wikipedia PKI词条整理。**

PKI（Public Key Infrastucture）是一系列的规则、策略以及过程，可以用来创建、管理、分发、使用、存储、移除数字证书以及管理公钥加密。PKI是用来实现信息的安全传输，在包括电子贸易、网上银行、保密电邮等网络活动中。在一些活动中，简单的密码验证已经不再适用，需要更加严格的验证来确认信息交互参与者的身份并验证正在传输的信息。

在密码学中，PKI是一种将实体的标识（identity）与公钥绑定在一起的组织结构。这种绑定是通过 Certificate Authority(CA)注册和颁发证书的过程建立的。根据绑定的保证级别，这可以通过自动化过程或在人工监督下进行。

PKI中实现有效正确的注册的角色称之为Registration Authority（RA），RA负责接受对数字证书的请求，并对发出请求的实体进行身份验证。

一个实体必须根据该实体的信息在每个CA域中惟一地标识。第三方验证机构(VA)可以代表CA提供此实体信息。

公钥加密技术可以在不可靠的公共网络中实现安全通讯，并通过数字签名技术来验证实体的身份。PKI是这样一个系统，用来创建、存储、分发数字签名，这些数字签名用来验证一个特定的公钥属于某个特定实体。PKI创建数字签名，这些数字签名将公钥map到实体上，并且在中心仓库中安全的存储这些签名，也会移除失效的签名。

PKI包含如下几个部分：

+  Certificate Authority (CA)，存储、发行、签署证书
+  Registration Authority（RA），通过查询他们存储在CA中的数字签名来验证实体的身份
+  Central Directory，存储索引键值得位置
+  Certificate Management System，管理证书的存储或者证书分发等事情
+  Certificate Policy，说明PKI对其程序的要求。它的目的是让外人分析PKI的可信度。


###  Methods of certification

有三方法实现这种证书加密，分别是：certificate authorities (CAs),  web of trust (WoT), and simple public key infrastructure (SPKI).

#### Certificate authorities

CA的主要角色是数字签名的签署分发，以及将公钥绑定到用户实体。这个过程是基于CA自己的私钥完成的，因此对用户密钥的信任依赖于对CA密钥有效性的信任。当CA是从用户系统分离的第三方应用时，它被称为 Registration Authority （RA），可以从CA中分离。"Key—User"的绑定关系的建立依赖于绑定保证的等级，根据保证等级可以选择软件自动化实现或者在人的监管下进行。

##### Issuer market share

在这种信任关系的模型中，CA是一个受信任的第三方，同时被证书的拥有者及依赖于证书的其他部分所信任。

##### Temporary certificates and single sign-on

这种方法引入了一个Single sign-on 服务器，作为一个离线的证书颁发角色。这个服务器会向客户端颁发数字证书，但是不会存储他们。用户可以用临时的证书来执行程序，这在基于X.509的证书中是很常见的。




