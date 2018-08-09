# Hyperledger Fabric(v1.2.0)代码分析1——channel创建

## 0. e2e_cli

Hyperledger Fabric提供了一个e2e的例子，该例中创建了一个基础的区块链网络，并进行了交易并查询。麻雀虽小，五脏俱全，我们可以根据e2e的例子来对Fabric区块链网络有一个基本的认识，包括网络创建流程、智能合约（chaincode）实现等。作为底层技术开发者，可以根据e2e的例子来追踪整个代码流程，从而对fabric源码结构有一个清晰的认识。

ele示例的区块链网络由以下几部分顺序执行：

### 0.1 Generate加密所需要的材料

本过程是根据定义了网络拓扑结构的yaml文件来产生对应关系的加密材料，其主要过程是调用Fabric编译生成的cryptogen工具，cryptogen输入为crypto-config.yaml该文件定义了网络的拓扑结构，具体内容可以查看该文件。

```shell
 $CRYPTOGEN generate --config=./crypto-config.yaml
```
### 0.2  Generate ChannelArtifacts

这一步主要是生成Orderer的创世区块和一些配置transaction。

```shell 
# 生成Orderer的创世块
$CONFIGTXGEN -profile TwoOrgsOrdererGenesis -channelID e2e-orderer-syschan -outputBlock ./channel-artifacts/genesis.block

# 生成Channel-create.tx
# 这是一个配置transaction，在channel创建时会用到这一步生成的transaction
$CONFIGTXGEN -profile TwoOrgsChannel -outputCreateChannelTx ./channel-artifacts/channel.tx -channelID $CHANNEL_NAME

# 生成anchor-peer update-transaction for peer0.org1
$CONFIGTXGEN -profile TwoOrgsChannel -outputAnchorPeersUpdate ./channel-artifacts/Org1MSPanchors.tx -channelID $CHANNEL_NAME -asOrg Org1MSP

# 生成anchor-peer update-transaction
$CONFIGTXGEN -profile TwoOrgsChannel -outputAnchorPeersUpdate ./channel-artifacts/Org2MSPanchors.tx -channelID $CHANNEL_NAME -asOrg Org2MSP
```

### 0.3 启动docker容器

这一步会根据docker-compose-cli.yaml文件来启动docker容器，其中包括了org1下的两个peer、org2下的两个peer、一个orderer、一个cli。cli容器是我们进行网络控制入口，在cli容器上，我们可以进行网络配置、发起交易、查询交易等操作。

```shell
 CHANNEL_NAME=$CH_NAME TIMEOUT=$CLI_TIMEOUT docker-compose -f $COMPOSE_FILE -f $COMPOSE_FILE_COUCH up -d 2>&1
```

### 0.4 script脚本执行

在cli容器启动后，会执行scripts/script.sh脚本，该脚本的内容主要包括以下操作;

+ createChannel
+ joinChannel
+ installChaincode
+ instantiateChaincode
+ invokeChaincode
+ queryChainclde

我们对网络的的操作都是从这个脚本发出的，所以我们可以根据script.sh脚本中的步骤来分析fabric代码


## 1. Create Channel

Channel的创建过程是这样的：
1. 首先的读取本文0.2章节生成的channel.tx，根据其中内容创建一个channel-create的transaction；
2. 经过一些基本的验证签名之后，会将channel-create-transaction通过gRPC调用发送给orderer节点；
3. orderer节点收到创建channel的transaction后，创建一个对应于该channel的创世块，然后将该创世块返回给cli，在后续join-channel的操作中会用到该创世块。

下面来看channel创建过程的代码是如何实现的。

channel创建是通过`peer channel create [flags & args]`命令来实现的，我们需要首先定位到`peer->channel->create`中到底干了啥。

生成的peer工具的源码位置是在`${Fabric}/peer/`目录下，打开该目录的下main.go源码可以看到如下内容：

```go
	mainCmd.AddCommand(version.Cmd())
	mainCmd.AddCommand(node.Cmd())
	mainCmd.AddCommand(chaincode.Cmd(nil))
	mainCmd.AddCommand(clilogging.Cmd(nil))
	mainCmd.AddCommand(channel.Cmd(nil))
```
在fabric中的命令行处理采用了比较流行的cobra库，以上几行代码是为peer添加几个子命令，我们这里关注的channel相关的命令，因此我们跳转进入channel相关的源码。

在`${Fabric}/peer/channel`目录下，是channel相关的实现，其中包含了create、join等子命令。

在文件`${Fabric}/peer/channel/create.go`中，找到create命令的执行函数

```go
func createCmd(cf *ChannelCmdFactory) *cobra.Command {
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a channel",
		Long:  "Create a channel and write the genesis block to a file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return create(cmd, args, cf)
		},
	}
	flagList := []string{
		"channelID",
		"file",
		"outputBlock",
		"timeout",
	}
	attachFlags(createCmd, flagList)

	return createCmd
}
```

当执行channel create命令时，会进入create(cmd, args, cf)函数，该函数位于源码文件`${Fabric}/peer/channel/create.go`最下面，内容如下：

```go
func create(cmd *cobra.Command, args []string, cf *ChannelCmdFactory) error {
	// the global chainID filled by the "-c" command
	if channelID == common.UndefinedParamValue {
		return errors.New("must supply channel ID")
	}

	// Parsing of the command line is done so silence cmd usage
	cmd.SilenceUsage = true

	var err error
	if cf == nil {
		cf, err = InitCmdFactory(EndorserNotRequired, PeerDeliverNotRequired, OrdererRequired)
		if err != nil {
			return err
		}
	}
	return executeCreate(cf)
}
```
create(cmd, args, cf)函数首先检查了是否提供了channelID，之后调用executeCreate(cf)。executeCreate(cf)函数进行了以下几部分工作：

+ sendCreateChainTransaction(cf)，发送transaction给order
+ getGenesisBlock(cf)，获取创世块
+ 将上一步获取的block尽心字节化编码
+ 写文件，将获取的创世块写到文件里保存

```go
func executeCreate(cf *ChannelCmdFactory) error {
	// 发送transaction到orderer
    err := sendCreateChainTransaction(cf)
	if err != nil {
		return err
	}
	
    // 从orderer获取生成的channel对应的创世块
	block, err := getGenesisBlock(cf)
	if err != nil {
		return err
	}

    // 将block字节化编码
	b, err := proto.Marshal(block)
	if err != nil {
		return err
	}

    // 字节序列化后的block写入文件中，后面join-channel时会用到该文件
	file := channelID + ".block"
	if outputBlock != common.UndefinedParamValue {
		file = outputBlock
	}
	err = ioutil.WriteFile(file, b, 0644)
	if err != nil {
		return err
	}
	return nil
}
```
看到这个地方，我们已经知道了创建channel时主要干了啥，后面我们一次分析每一项是具体怎么实现的。

### 1.1 sendCreateChainTransaction(cf)

首先是发送transaction，我们跳转进入该函数。

```go
func sendCreateChainTransaction(cf *ChannelCmdFactory) error {
	var err error
	var chCrtEnv *cb.Envelope

	if channelTxFile != "" {
		if chCrtEnv, err = createChannelFromConfigTx(channelTxFile); err != nil {
			return err
		}
	} else {
		if chCrtEnv, err = createChannelFromDefaults(cf); err != nil {
			return err
		}
	}

	if chCrtEnv, err = sanityCheckAndSignConfigTx(chCrtEnv); err != nil {
		return err
	}

	var broadcastClient common.BroadcastClient
	broadcastClient, err = cf.BroadcastFactory()
	if err != nil {
		return errors.WithMessage(err, "error getting broadcast client")
	}

	defer broadcastClient.Close()
	err = broadcastClient.Send(chCrtEnv)

	return err
}
```

sendCreateChainTransaction()函数以此做了如下工作：

1. 根据传入的channel.tx文件创建transaction格式封装chCrtEnv，如果没有提供该文件的话，就根据默认情况创建chCrtEnv；
2. 对于生成的chCrtEnv进行一系列检查，并生成签名；
3. broadcastClient.Send(chCrtEnv)，将chCrtEnv发送给Orderer。

这里需要讲一下，对于channel相关的配置管理通过ConfigUpdate数据结构来实现的，创建channel的过程就是一个特殊的ConfigUpdate，当发现ConfigUpdate对应的channel不存在时，就会创建channel。关于channel配置相关的数据类型可以参考博文(https://www.cnblogs.com/gexin/p/9332719.html)。

后面的具体创建工作的是在orderer节点上进行的，接下来将目光转移到orderer实现代码中。

### 1.2 Orderer创建channel

Orderer节点会启动gRPC server，用于帧听client过来的broadcast 消息。

消息的处理函数位于`orderer\common\broadcast\broadcast.go`中，如下所示：

```go

// Handle starts a service thread for a given gRPC connection and services the broadcast connection
func (bh *handlerImpl) Handle(srv ab.AtomicBroadcast_BroadcastServer) error {
	
		// 一系列检查
	    // 检查代码略过
    	// 根据msg解析出其中的channelHeader
		chdr, isConfig, processor, err := bh.sm.BroadcastChannelSupport(msg)
		if !isConfig {
			// 过来的消息不是config消息
		} else { // isConfig
			// 该代码部分是配置transaction的处理部分
			// 应用msg中的配置到当前配置
            config, configSeq, err := processor.ProcessConfigUpdateMsg(msg)
			if err != nil {
				// 错误处理
				}
             // 然后将config应用 
			err = processor.Configure(config, configSeq)
			if err != nil {
				// 错误处理
				}
		}

		err = srv.Send(&ab.BroadcastResponse{Status: cb.Status_SUCCESS})
		if err != nil {
			// 错误处理
		}
}
```
由以上代码，我们可以看到在gRPC的处理函数中的做了以下几个工作：

1. BroadcastChannelSupport(msg)，该函数定义在`orderer/common/mutichannel/registar.go`中，这一步骤是根据chainID去找到map已有的channelSupport类型变量（如果没有就直接新建了一个）并判断是否是config类型的transaction。
2. ProcessConfigUpdateMsg(msg)，该函数定义在`orderer/common/msgprocessor/systemchannel.go`中，对于一个创建channel的消息而言，该过程会返回一个ORDERER_TRANSACTION。
3. Configure(config, configSeq)，该函数做的就是将上一步产生的ORDERER_TRANSACTION转化为kafka类型的消息，之后通过enqueue()发送到kafka的响应topic中用以排序。

接下来，kafka会对收到的消息进行排序，排序完成后，进行实际的创建genesis块的工作，该部分是在processRegular()函数中实现的，该部分定义在`orderer/consensus/kafka/chain.go`中。其源码如下所示：


```
func (chain *chainImpl) processRegular(regularMessage *ab.KafkaMessageRegular, receivedOffset int64) error {



case ab.KafkaMessageRegular_CONFIG:
		// Any messages coming in here may or may not have been re-validated
		// and re-ordered, BUT they are definitely valid here
		// advance lastOriginalOffsetProcessed iff message is re-validated and re-ordered
		
		offset := regularMessage.OriginalOffset
		if offset == 0 {
			offset = chain.lastOriginalOffsetProcessed
		}

		commitConfigMsg(env, offset)
		
}
```

commitConfigMsg()函数就是实际写入block的函数，其代码如下所示：

```
// When committing a config message, we also update `lastOriginalOffsetProcessed` with `newOffset`.
// It is caller's responsibility to deduce correct value of `newOffset` based on following rules:
// - if Resubmission is switched off, it should always be zero
// - if the message is committed on first pass, meaning it's not re-validated and re-ordered, this value
//   should be the same as current `lastOriginalOffsetProcessed`
// - if the message is re-validated and re-ordered, this value should be the `OriginalOffset` of that
//   Kafka message, so that `lastOriginalOffsetProcessed` is advanced
commitConfigMsg := func(message *cb.Envelope, newOffset int64) {
logger.Debugf("[channel: %s] Received config message", chain.ChainID())
batch := chain.BlockCutter().Cut()

chain.lastOriginalOffsetProcessed = newOffset
block := chain.CreateNextBlock([]*cb.Envelope{message})
metadata := utils.MarshalOrPanic(&ab.KafkaMetadata{
LastOffsetPersisted:         receivedOffset,
LastOriginalOffsetProcessed: chain.lastOriginalOffsetProcessed,
LastResubmittedConfigOffset: chain.lastResubmittedConfigOffset,
})
chain.WriteConfigBlock(block, metadata)
chain.lastCutBlockNumber++
chain.timer = nil
}
```



至此，一个channel的genesis区块也就创建完成了。




