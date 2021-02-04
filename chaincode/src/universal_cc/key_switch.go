package main

import (
	"fmt"

	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/keyswitch"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

func (uc *UniversalCC) createKeySwitchTrigger(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	return shim.Error(codeNotImplemented)
}

func (uc *UniversalCC) createKeySwitchResult(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	return shim.Error(codeNotImplemented)
}

func (uc *UniversalCC) getKeySwitchTriggerHelper(stub shim.ChaincodeStubInterface, keySwitchSessionID string) (*keyswitch.KeySwitchTriggerStored, error) {
	return nil, fmt.Errorf(codeNotImplemented)
}

func (uc *UniversalCC) getKeySwitchResult(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	return shim.Error(codeNotImplemented)
}

func (uc *UniversalCC) listKeySwitchResultsByID(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	return shim.Error(codeNotImplemented)
}
