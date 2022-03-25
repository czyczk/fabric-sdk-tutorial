# fabric-sdk-tutorial

## 介绍

此仓库是一个示例工程，用于展示一种利用 Fabric SDK Go 来管理数字文档和资产记录的方式，其中使用了如密钥置换等加密技术以保证在用户访问加密数据时强制留痕留证。

## 使用说明

以下是所有的可用命令，可以按需使用它们。

|命令|描述|
|-|-|
|`make`|清理已启动的 Fabric 网络，重启网络并运行程序。程序将初始化网络，并启动为一个服务器。|
|`make clean`|关闭已启动的 Fabric 网络并清理它。|
|`make env-up`|启动 Fabric 网络。不要在 `make clean` 之前使用。|
|`make env-down`|只关闭已启动的 Fabric 网络，不清理。|
|`make build`|编译程序。|
|`make run`|运行程序。初始化网络并启动为服务器。|
|`make run-init`|运行程序。只初始化网络。|
|`make run-serve`|运行程序。只启动为服务器。以 User1@org1.lab805.com 登录。|
|`make run-serve-ado1`|运行程序。只启动为服务器。以 Admin@org1.lab805.com 登录。|
|`make run-serve-u1o2`|运行程序。只启动为服务器。以 User1@org2.lab805.com 登录。|

一种典型的启动方案（按顺序执行）：

`make clean` 以清理干净上一次的运行。  
`make build` 或 `go build` 以编译应用。  
`make env-up` 以启动 Docker 容器。  
`make run-init` 以初始化网络并在节点上安装链码。  
`make run-serve` 以启动一个服务器在端口 8081 上上。以 User1@org1.lab805.com 登录。  
在新的 shell 中，`make run-serve-ado1` 以启动第二个服务器在端口 8082上。以 Admin@org1.lab805.com 登录。  
再开一个新 shell，`make run-serve-u1o2` 以启动第三个服务器在端口 8083 上。以 User1@org2.lab805.com 登录。

## 密钥置换服务说明
### 密钥转换服务的用途

密钥置换流程中共涉及到 4 种密钥，分别是私钥、公钥、集合私钥和集合公钥。

在网络初始化之前，规划者需要提前制定好密钥置换服务的服务器。集合私钥是由这些服务器的私钥生成的，与之类似集合公钥是由这些服务器的公钥生成的。

在平台中，用于加密数据的对称密钥是用集合公钥加密的。因此，解密这些数据就只需要解密对称密钥。解密对称密钥有 2 种方式。第一种是使用集合私钥，这就是监管者的解密方式（因为只有监管者持有集合私钥）；另一种是通过完整的密钥置换流程，这要求指定的密钥置换服务器上传它们所计算出的份额。

为了完成密钥置换流程，客户端要向密钥置换服务器公开其公钥作为目标，密钥置换服务器需要配以私钥以计算其份额，之后客户端便可以使用其私钥解密对称密钥。

反过来，客户端要上传加密数据需要知道集合公钥，这样才能加密对称密钥。

总结来说，密钥置换服务器需要配以其私钥；需要上传加密数据的客户端需要配以集合公钥，需要解密加密数据的客户端需要配以其公钥和私钥；监管者为了方便其快速解密与绕过密钥置换流程，需要配以集合私钥。

### 密钥生成器

上述的密钥可以通过专用的密钥置换密钥生成器或使用其他相兼容的使用 SM2 算法以及集合密钥生成算法的工具。

专用工具在 `cmd/sm2keygen` 文件夹中，可以使用 `go run ./cmd/sm2keygen` 来运行它。同文件夹下的配置文件 `users.yaml` 可以根据需要修改，以表示需要为哪些用户生成密钥。集合私钥和集合公钥也将由为这些用户生成的私钥和公钥所生成。生成出的密钥文件将被放在 `sm2keys` 文件夹中。

**说明：**

每次运行该工具前注意备份 `sm2keys` 文件夹。当集合密钥只需要由部分用户（子集）的密钥生成时，可以通过运行该工具两次来达成。第一次运行中，只为这个子集生成密钥，将生成的文件保存（备份）下来以备后用。在第二次运行中，为剩下的用户生成密钥，只采用其中的公、私钥。

## 角色说明
### 密钥置换服务器

密钥置换服务器是一个在配置文件中开启了密钥置换服务器选项的应用实例。为保证平台密钥置换服务的可用性，需要保持服务器一直处在运行状态。密钥置换服务只有在所有服务器都有应答时才可用。

密钥置换服务器负责在来自客户端的密钥置换请求出现时，尽快地计算其份额并上传上链。

这个功能需要在配置文件中指定密钥置换用的私钥。

### 监管者

监管者是一个在配置文件中开启了监管者身份选项的应用实例，在后台会运行有一个监管者服务器。

处于运行状态中的监管者服务器会在后台检测任何加密数据的创建，并使用监管者特权解密它们（对于一些资源只解密属性）。

为使监管者功能能正常使用，需要在配置文件中指定集合私钥。

## 应用配置文件说明

### 初始化配置文件

此类型的配置文件包含创建一个 Fabric 网络的所有所需信息，包括网络中会出现的所有用户身份信息、用于创建通道、安装和实例化链码的信息。

- **顶层字段：**

|字段名称|值类型|描述|
|--------|------|----|
|users|（见下）|参与网络的用户信息|
|channels|（见下）|网络中的通道信息|
|chaincodes|（见下）|网络中链码的信息|

- **users：**

这一段包含网络中组织与其中用户的信息，用于在初始化过程中创建一些必要的客户端。  
格式为一个关于组织 → 用户的列表。例如：

```
users:
  Org1:
    adminIDs:
	  - Admin
	userIDs:
	  - User1
  Org2:
    adminIDs:
	  - Admin
	userIDs:
	  - User1
```

- **channels:**

这一段包含网络中通道的信息，以及每个通道的通道配置文件。  
从一个通道列表开始，每个通道可以详配以参与者与通道配置文件。配置如以下形式。

```
channels:
  channel1:
    participants:
	  ...
	configs:
	  ...
  channel2:
    participants:
	  ...
	configs:
	  ...
  ...
```

`participants` 部分包含通道参与者的信息，配置如以下形式。
```
...
    participants:
	  Org1:
	    orgName: ...
		userID: ...
	  Org2:
	    orgName: ...
		userID: ...
...
```

`configs` 部分包含通道配置文件的信息，每条该信息有一个应用者（应用这个配置文件的用户身份，通常即通道的管理员）。这一部分如以下形式。

```
...
    configs:
	  - path: path/to/config1.tx
	    orgName: ...
		userID: Admin
	  - path: path/to/config2.tx
	    orgName: ...
		userID: Admin
...
```

该部分的一个完整示例如下。

```
channels:
  mychannel:
    participants:
      Org1:
        orgName: Org1
        userID: Admin
      Org2:
        orgName: Org2
        userID: Admin
    configs:
      - path: fixtures/channel-artifacts/channel.tx
        orgName: Org1
        userID: Admin
      - path: fixtures/channel-artifacts/Org1MSPanchors.tx
        orgName: Org1
        userID: Admin
      - path: fixtures/channel-artifacts/Org2MSPanchors.tx
        orgName: Org2
        userID: Admin

```

- **chaincodes:**

这一段定义了网络中会被用到的链码的属性，包括链码的版本、路径、安装和实例化等信息。  
从一个链码列表开始，每个链码配以其基本属性和安装和实例化信息。这一段的具体内容以如以下格式。

```
chaincodes:
  chaincode1:
    version: ...
	path: ...
	goPath: ...
	installations:
	  ...
	instantiations:
	  ...
  chaincode2:
    version: ...
	path: ...
	goPath: ...
	installations:
	  ...
	instantiations:
	  ...
  ...

```

值得注意的是，链码的实际路径按 `${goPath}/src/${path}` 的拼接来算。

`installations` 部分设计用于提供信息以指导链码应该安装在哪些组织中。指定一个组织与一个用户身份（通常即组织管理员），就可以为该组织的所有节点安装链码。这一部分如以下形式。

```
...
  chaincode1:
    ...
	installations:
	  Org1:
	    orgName: Org1
		userID: Admin
	  Org2:
	    orgName: Org2
		userID: Admin
	  ...
...
```

`instantiations` 部分设计用于提供信息以指导链码应该在哪些通道上实例化，以及该实例的策略和初始化参数。这一部分如以下形式。

```
...
  chaincode1:
    ...
	instantiations:
	  channel1:
	    policy: OR('Org1MSP.member', 'Org2MSP.member')
		initArgs:
		  - initFuncName
		  - arg1
		  - arg2
		  ...
      channel2:
	    policy: ...
		initArgs:
		  - initFuncName
		  - arg3
		  - arg4
		  ...
	  ...
...
```

这一段的一个完整示例如下：

```
chaincodes:
  screwCc:
    version: 0.1
    path: screw_example
    goPath: chaincode
    installations:
      Org1:
        orgName: Org1
        userID: Admin
      Org2:
        orgName: Org2
        userID: Admin
    instantiations:
      mychannel:
        policy: OR('Org1MSP.member', 'Org2MSP.member')
        initArgs:
          - init
          - Org1
          - 200
          - Org2
          - 100
        orgName: Org1
        userID: Admin
  universalCc:
    version: 0.1
    path: universal_cc
    goPath: chaincode
    installations:
      Org1:
        orgName: Org1
        userID: Admin
      Org2:
        orgName: Org2
        userID: Admin
    instantiations:
      mychannel:
        policy: OR('Org1MSP.member', 'Org2MSP.member')
        initArgs:
        orgName: Org1
        userID: Admin
```

### 服务器配置文件

- **顶层字段：**  

|字段名称|值类型|描述|
|--------|------|----|
|user|（见下）|用于登录的用户身份信息|
|channels|list|用户参与的通道|
|port|uint|服务器运行在这个端口上|
|localDBSourceName|string|节点的本地数据库连接字符串|
|ipfsAPI|string|节点的 IPFS 网络 API 地址|
|isKeySwitchServer|bool|是否启动为密钥置换服务器|
|isRegulator|bool|是否启动为监管者|
|keySwitchKeys|（见下）|启动为该角色所需要的密钥|
|showTimingLogs|bool|是否显示一些关键流程的时间消耗，有助于了解性能。|

- **user:**  

|字段名称|值类型|描述|
|--------|------|----|
|orgName|string|用户所属的组织名称|
|userID|string|用户身份|

示例：  
以下是 MSP 名称为 `Org1` 的组织中的一个用户身份 `User1` 在这一段配置中的的正确写法。

```
user:
  orgName: Org1
  userID: User1
```

- **keySwitchKeys:**  

|字段名称|值类型|描述|
|--------|------|----|
|encryptionAlgorithm|string|The algorithm used in key switch keys generation and the key switch process. Only "SM2" is supported for now.密钥置换流程中用于生成密钥置换密钥的算法。目前只支持 `"SM2"`。|
|collectivePrivateKey|string|集合私钥的相对或绝对路径。仅当用户是监管者时需要指定。|
|collectivePublicKey|string|集合公钥的相对或绝对路径。要参与上传加密数据的用户需要指定。|
|privateKey|string|用于密钥置换流程的用户私钥的相对或绝对路径。要解密数据的用户需要指定；密钥置换服务器要用它计算其份额，也需要指定。|
|publicKey|string|用于密钥置换流程的用户公钥的相对或绝对路径。要解密数据的用户需要指定。|

示例：

```
keySwitchKeys:
  encryptionAlgorithm: "SM2"
  collectivePrivateKey: "sm2keys/collPrivKey.pem"
  collectivePublicKey: "sm2keys/collPubKey.pem"
  privateKey: "sm2keys/Admin@org1.lab805.com/sk"
  publicKey: "sm2keys/Admin@org1.lab805.com/Admin@org1.lab805.com.pem"
```

- **内置的服务器配置文件：**

|配置文件名|为该用户身份准备|
|----------|----------------|
|server.yaml|User1@org1.lab805.com|
|server-ado1.yaml|Admin@org1.lab805.com|
|server-u1o2.yaml|User1@org2.lab805.com|

## 内置的 Fabric 网络用户

仓库中有 4 个默认的 Fabric 网络的用户的证书文件，其中 3 个拥有应用配置文件，可以在启动为服务器时作为登录用户。这 3 个每个用户都有它们的*角色*（在之前小节中描述过）。用户情况如下表。

|用户|密钥置换服务器|监管者|
|----|--------------|------|
|User1@org1.lab805.com|✔|✖|
|Admin@org1.lab805.com|✖|✔|
|User1@org2.lab805.com|✔|✖|

## FAQ

- 运行 `run-init` 时发生关于无法构建/编译链码的错误。

出现像这样的错误  
```
failed to generate platform-specific docker build
```

或其他更明显的信息意味着链码与 Docker 镜像中的 Go 编译器存在不兼容问题。这通常是因为 `vendor` 文件夹中陈旧的内容所致。  
如果发生这种错误，尝试删除所有链码文件夹中的 `vendor` 文件夹。运行 `make clean` 后再从头来一次。
