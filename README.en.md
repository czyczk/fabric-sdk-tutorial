# fabric-sdk-tutorial

#### Description

This repo is a demo project to show a possible use case to make use of Fabric SDK Go to manage digital documents and asset records featuring encryption techniques like the key-switch process to enforce leaving traces when accessing encrypted data.

#### Instructions

`make` and `make clean` are the two commands for common use.

`make` will clean the previously started Fabric network, restart the Fabric network, run the app to initialize the network and then start as a server.

`make clean` shutdowns and cleans the Fabric network.

Besides, these are all the commands you can use to achieve different tasks.

|Command|Description|
|-|-|
|`make`|Cleans the started Fabric network, restarts the network and runs the app to initialize the network and then start as a server.|
|`make clean`|Brings down the started Fabric network and cleans it.|
|`make env-up`|Brings up the Fabric network. Don't run it until `make clean`.|
|`make env-down`|Brings down the started Fabric network without cleaning.|
|`make build`|Compiles the app.|
|`make run`|Run the app to initialize the network and start as a server.|
|`make run-init`|Run the app only to initialize the network.|
|`make run-serve`|Run the app only to start a server.|

