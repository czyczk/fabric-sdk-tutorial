package main

import (
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

func (uc *UniversalCC) getRegulatorKey(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	return shim.Error(codeNotImplemented)
}

func (uc *UniversalCC) getRegulatorKeyHistory(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	return shim.Error(codeNotImplemented)
}

func (uc *UniversalCC) updateRegulatorKey(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	return shim.Error(codeNotImplemented)
}
