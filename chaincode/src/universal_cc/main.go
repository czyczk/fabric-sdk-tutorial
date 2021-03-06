package main

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

// UniversalCC 实现 Chaincode 接口。它将负责数据上链、访问权申请批准与密钥置换等功能的相关数据在区块链上的存取。
type UniversalCC struct{}

// Init 用于初始化链码。
func (uc *UniversalCC) Init(stub shim.ChaincodeStubInterface) peer.Response {
	args := stub.GetArgs()
	if len(args) != 0 {
		return shim.Error("初始化不接收参数")
	}

	return shim.Success(nil)
}

// Invoke 用于分流链码调用。
func (uc *UniversalCC) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	// 解出具体函数名与参数
	funcName, args := stub.GetFunctionAndParameters()

	switch funcName {
	// data.go
	case "createPlainData":
		return uc.createPlainData(stub, args)
	case "createEncryptedData":
		return uc.createEncryptedData(stub, args)
	case "createOffchainData":
		return uc.createOffchainData(stub, args)
	case "getMetadata":
		return uc.getMetadata(stub, args)
	case "getData":
		return uc.getData(stub, args)
	case "getKey":
		return uc.getKey(stub, args)
	case "getPolicy":
		return uc.getPolicy(stub, args)
	case "listResourceIDsByCreator":
		return uc.listResourceIDsByCreator(stub, args)
	case "listResourceIDsByConditions":
		return uc.listResourceIDsByConditions(stub, args)
	// auth.go
	case "createAuthRequest":
		return uc.createAuthRequest(stub, args)
	case "createAuthResponse":
		return uc.createAuthResponse(stub, args)
	case "getAuthRequest":
		return uc.getAuthRequest(stub, args)
	case "getAuthResponse":
		return uc.getAuthResponse(stub, args)
	case "listPendingAuthSessionIDsByResourceCreator":
		return uc.listPendingAuthSessionIDsByResourceCreator(stub, args)
	case "listAuthSessionIDsByRequestor":
		return uc.listAuthSessionIDsByRequestor(stub, args)
	// key_switch.go
	case "createKeySwitchTrigger":
		return uc.createKeySwitchTrigger(stub, args)
	case "createKeySwitchResult":
		return uc.createKeySwitchResult(stub, args)
	case "getKeySwitchResult":
		return uc.getKeySwitchResult(stub, args)
	case "listKeySwitchResultsByID":
		return uc.listKeySwitchResultsByID(stub, args)
	// identity.go
	case "getDepartmentIdentity":
		return uc.getDepartmentIdentity(stub, args)
	}

	return shim.Error("未知的链码函数调用")
}

func main() {
	err := shim.Start(new(UniversalCC))
	if err != nil {
		fmt.Printf("无法启动 UniversalCC: %s", err)
	}
}
