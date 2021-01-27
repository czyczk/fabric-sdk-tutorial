# fabric-sdk-tutorial

#### 介绍

此仓库是一个示例工程，用于展示一种利用 Fabric SDK Go 来管理数字文档和资产记录的方式，其中使用了如密钥置换等加密特性以保证访问加密数据时留痕。

#### 使用说明

`make` 和 `make clean` 是通常使用的两个命令。

`make` 将清理之前启动的 Fabric 网络、重启网络，并运行程序。程序在初始化网络后将启动为一个服务器。

`make clean` 关闭并清理已启动的 Fabric 网络。

此外，以下是所有的可用命令，可以按需使用它们。

|命令|描述|
|-|-|
|`make`|清理已启动的 Fabric 网络，重启网络并运行程序。程序将初始化网络，并启动为一个服务器。|
|`make clean`|关闭已启动的 Fabric 网络并清理它。|
|`make env-up`|启动 Fabric 网络。不要在 `make clean` 之前使用。|
|`make env-down`|只关闭已启动的 Fabric 网络，不清理。|
|`make build`|编译程序。|
|`make run`|运行程序。初始化网络并启动为服务器。|
|`make run-init`|运行程序。只初始化网络。|
|`make run-serve`|运行程序。只启动为服务器。|
