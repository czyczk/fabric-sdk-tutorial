package main

import (
	"testing"

	"github.com/hyperledger/fabric-chaincode-go/shim"
)

// Init with no parameter. Should pass.
func TestInitWithNoParameter(t *testing.T) {
	mockStub := createMockStub(t, "TestInitWithNoParameter")

	// Expect the chaincode to get initialized normally
	arguments := [][]byte{}
	resp := initChaincode(mockStub, arguments)

	if resp.Status != shim.OK {
		testLogger.Infof("Failed to initialize chaincode: %v\n", resp.Message)
		t.FailNow()
	}
}

// Init with one parameter. Should return error.
func TestInitWithOneParameter(t *testing.T) {
	mockStub := createMockStub(t, "TestInitWithOneParameter")

	// Expect the chaincode to return an error
	arguments := [][]byte{[]byte("Whatever")}
	resp := initChaincode(mockStub, arguments)

	if resp.Status != shim.ERROR {
		testLogger.Infof("Should return error\n")
		t.FailNow()
	}
}
