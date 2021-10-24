# fabric-sdk-tutorial

## Description

This repo is a demo project to show a possible use case to make use of Fabric SDK Go to manage digital documents and asset records featuring encryption techniques like the key-switch process to enforce leaving traceable evidence when encrypted data is accessed by users.

## Instructions

These are all the commands you can use to achieve different tasks.

|Command|Description|
|-|-|
|`make`|Cleans the started Fabric network, restarts the network and runs the app to initialize the network and then start as a server.|
|`make clean`|Brings down the started Fabric network and cleans it.|
|`make env-up`|Brings up the Fabric network. Don't run it until `make clean`.|
|`make env-down`|Brings down the started Fabric network without cleaning.|
|`make build`|Compiles the app.|
|`make run`|Run the app to initialize the network and start as a server.|
|`make run-init`|Run the app only to initialize the network.|
|`make run-serve`|Run the app only to start a server. Logged in as User1@org1.lab805.com.|
|`make run-serve-ado1`|Run the app only to start a server. Logged in as Admin@org1.lab805.com.|
|`make run-serve-u1o2`|Run the app only to start a server. Logged in as User1@org2.lab805.com.|

A typical scenario (following the order):

`make clean` to clean up the last run.  
`make build` or `go build` to compile the app.  
`make env-up` to start the docker containers.  
`make run-init` to initialize the network and install the chaincodes on the peers.  
`make run-serve` to start a server on port 8081. Logged in as User1@org1.lab805.com.  
In a new shell, `make run-serve-ado1` to start a second server on port 8082. Logged in as Admin@org1.lab805.com.  
In another new shell, `make run-serve-u1o2` to start a third server on port 8083. Logged in as User1@org2.lab805.com.

## About key-switch service
### What is key-switch service used for

There are 4 kinds of keys that are involed in the key-switch process, namely the private key, the public key, the collective private key and the collective public key.

Before the network is initialized, the servers of the key-switch service must be planned beforehand. The collective private key is derived from the private keys of the servers and similarly, the collective public key is derived from the public keys of the servers.

In the platform, the symmectric key used to encrypt the data is encrypted with the collective public key. Hence, to decrypt the data, one only needs to decrypt the symmectric key. There are 2 ways to do this. One is to simply use the collective private key, which is the regulator way of doing this (since only the regulator holds the collective private key). The other is to go through the whole key-switch process, which requires each key-switch server to upload their calculated share.

In order to complete the key-switch process, the user must expose its public key to inform the key-switch server of the target. A key-switch server must be provided with its private key to calculate its share. As a result, the user can decrypt the symmetric key using its private key.

In a reverse fashion, to upload encrypted data, the client must be informed of the collective public key so that it can encrypt the symmectric key.

So to sum it up, a key-switch server must be configured with its private key. A client that needs to upload encrypted data must be configured with the collective public key. A client that needs to fetch decrypted data must be configured with its public and private key. The regulator, for his/her convenience, should be configured with the collective private key.

### Key generator

The keys mentioned above can be generated using the dedicated key-switch key generator or using any tools that are compatible with the SM2 algorithm and the collective key generation algorithm.

The dedicated tool is located in `cmd/sm2keygen`. You can use `go run ./cmd/sm2keygen` to run it. The config file `users.yaml` located in the same directory can be modified to indicate which users are to be generated keys for. The collective private key and the collective public key are also derived from the keys of these users. The generated keys will be placed in directory `sm2keys`.

**Notes:**

Back up directory `sm2keys` every time before the next run of the tool. To generate collective keys that are derived from only the keys of part of the users (a subset), you can run the tool twice. In the first run, generate the keys for the subset. Store them (back them up) for later use. In the second run, generate the keys for the rest of the users and adopt only the private and public keys.

## About the roles
### Key-switch server

A key-switch server is an app instance that is configured with the key-switch server option on. In order to ensure the availablility of the key-switch service of the platform, always keep the server running. The key-switch service is available only when all the servers are responsive.

A key-switch server is responsible for calculating its share and uploading the share onto the chain as soon as possible when a key-switch request from a client emerges.

For that to work, a key-switch server must be configured with its private key.

### Regulator

A regulator is an app instance that is configured with the regulator identity option on, which will run a regulator server in the background.

A running regulator server can detect any creation of encrypted data and decrypt them (for some, only the properties) in the background with the regulator privileges.

For a regulator to function correctly, it must be configured with the collective private key.

## About the app config file

### Initialization config file

This type of config file contains all the needed information to create a Fabric network. It contains info of all the user identities that will appear in the network, along with the info needed to create channels, to install and instantiate chaincodes.

- **Top-level fields:**

|Field name|Value type|Description|
|----------|----------|-----------|
|users|(Seen below)|The info of the participating users in the network|
|channels|(Seen below)|The info of the channels in the network|
|chaincodes|(Seen below)|The info of the chaincodes in the network|

- **users:**

This section includes info about the organizations of the network and the users in them. The info is provided to create necessary clients needed during the initialization process.  
Organized as a list of organizations -> users. E.g.:

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

This section includes info about the channels in the network, as well as the channel configuration files for each channel.  
Starting from a list of channels, each channel is configured with its participants and config files. Organized as

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

The `participants` part includes info about the channel participants, in the form of
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

The `configs` part includes info about the channel configuration files, each with an applicant (the identity of the user to apply the configuration file, usually an admin of the channel). This part is in the form of

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

A complete example of this section is as follows.

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

This section defines the properties of the chaincodes to be used in the network, including info of the versions, paths of the chaincodes, and the installation and instantiation info.  
Starting from a list of chaincodes, each is configured with its basic properties and the installation and instantiation info. Specifically, the section is organized as follows.

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

Note that the actual chaincode path will be recognized by concatenating the two path fields— `${goPath}/src/${path}`.

The `installations` part is designed to provide info about for which organizations the chaincode is to be installed. By specifying an organization with the user identity as the installation performer (usually the organization admin), the chaincode will be installed on all peers of the organization. This part is in the form of

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

The `instantiations` part is designed to provide info about which channels the chaincode should be instantiate on, the policy of the instantiation and the init args. This part is in the form of

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

A complete example of this section:

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

### Server config file

- **Top-level fields:**  

|Field name|Value type|Description|
|----------|----------|-----------|
|user|(Seen below)|The info of the user to log in|
|channels|list|The channels that the user participate in|
|port|uint|The port on which the server will run|
|localDBSourceName|string|The node's connection string for the local database|
|ipfsAPI|string|The node's API address of the IPFS network|
|isKeySwitchServer|bool|Whether the server is a key-switch server|
|isRegulator|bool|Whether the server is a regulator|
|keySwitchKeys|(Seen below)|The required keys to be used by the role|

- **user:**  

|Field name|Value type|Description|
|----------|----------|-----------|
|orgName|string|The name of the organization to which the user belong|
|userID|string|User ID|

E.g.:  
An identity of `User1` of an organization with an MSP name of `Org1` leads to the following contents as the correct configuration.

```
user:
  orgName: Org1
  userID: User1
```

- **keySwitchKeys:**  

|Field name|Value type|Description|
|----------|----------|-----------|
|encryptionAlgorithm|string|The algorithm used in key-switch keys generation and the key-switch process. Only `"SM2"` is supported for now.|
|collectivePrivateKey|string|The relative or absolute path to the collective private key. Only required if the user is to be configured as a regulator.|
|collectivePublicKey|string|The relative or absolute path to the collective public key. Required by any user that needs to participate in uploading encrypted data.|
|privateKey|string|The relative or absolute path to the user's private key for the key-switch process. Required by any user in need of decrypting data. Also required by a key-switch server to calculate its share.|
|publicKey|string|The relative or absolute path to the user's public key for key-switch process. Required by any user in need of decrypting data.|

E.g.:

```
keySwitchKeys:
  encryptionAlgorithm: "SM2"
  collectivePrivateKey: "sm2keys/collPrivKey.pem"
  collectivePublicKey: "sm2keys/collPubKey.pem"
  privateKey: "sm2keys/Admin@org1.lab805.com/sk"
  publicKey: "sm2keys/Admin@org1.lab805.com/Admin@org1.lab805.com.pem"
```

- **Built-in server config files:**

|Config file name|For user identity|
|----------------|-----------------|
|server.yaml|User1@org1.lab805.com|
|server-ado1.yaml|Admin@org1.lab805.com|
|server-u1o2.yaml|User1@org2.lab805.com|

## Built-in users of the Fabric network

There are certificate files for the 4 default users of the Fabric network. 3 of them has an app config file and can be logged in while starting the server. Each of the 3 users has their *role*. The users are as follows.

|User|Key-switch server|Regulator|
|----|---------|---------|
|User1@org1.lab805.com|✔|✖|
|Admin@org1.lab805.com|✖|✔|
|User1@org2.lab805.com|✔|✖|

