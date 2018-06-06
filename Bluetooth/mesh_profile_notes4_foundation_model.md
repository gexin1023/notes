##  Foundation Model

Foundation Model 定义了配置管理网络所需的access层状态、消息、model等内容。


### 基本约定

#### 大小端

该部分使用小端

#### Log转换

为了将两字节数据压缩到一个字节，使用了如下的转换。转换规则是将[$2^{(n-1)}$, $2^n-1$]范围的数，映射到 $n$。

![ch4_log_trans](pic\ch4_log_trans.png)

###  状态定义

状态定义的规范在《Mesh Model specification》中有详细描述。

#### 组成数据 Composition Data

组成数据包含了节点的的信息，比如节点所包含的element、及其支持的model。组成数据是由一些信息页组成。Page0 是必须的，其他页是可选择的。


