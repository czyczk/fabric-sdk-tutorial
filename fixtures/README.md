# 一些小工具

在 `sciprts` 文件夹中。

|脚本名称|用途|
|--------|----|
|crypto-sk-normalize|使用 `cryptogen` 工具生成的密钥材料中，私钥都是 `***_sk` 的形式，其中前面的部分是随机的，不便于编写配置文件。此脚本将它们全部改成 `sk`。|
|ipfs-repo-gen|以 `ipfs-template/ipfs0` 为集群的主节点，生成一定数量的 IPFS 仓库。|

# 生成 fixtures 所需材料的方式

## 准备工作

设置 `$FABRIC_HOME` 变量指向 Fabric 的二进制文件夹 `bin`，以便后续操作。

## 密钥文件
```
$FABRIC_HOME/cryptogen generate --config=./crypto-config.yaml
```

要注意的是每次使用该工具都是创建新的 CA，因此  
1. 新密钥文件不能和之前生成的混用。
2. 创世块中含有密钥信息，故生成新材料后，旧的 channel artifacts 都作废。

另一个小细节是，`crypto-config` 文件夹中的要删干净再运行上面的命令，否则已有的身份不会被覆盖更新，这样会造成新旧混用的情况。

## 通道文件（etcdraft 版）
```
$FABRIC_HOME/configtxgen -profile SampleMultiNodeEtcdRaft -channelID byfn-sys-channel -outputBlock ./channel-artifacts/genesis.block
export CHANNEL_NAME=mychannel
$FABRIC_HOME/configtxgen -profile TwoOrgsChannel -outputCreateChannelTx ./channel-artifacts/channel.tx -channelID $CHANNEL_NAME
$FABRIC_HOME/configtxgen -profile TwoOrgsChannel -outputAnchorPeersUpdate ./channel-artifacts/Org1MSPanchors.tx -channelID $CHANNEL_NAME -asOrg Org1MSP
$FABRIC_HOME/configtxgen -profile TwoOrgsChannel -outputAnchorPeersUpdate ./channel-artifacts/Org2MSPanchors.tx -channelID $CHANNEL_NAME -asOrg Org2MSP
```
