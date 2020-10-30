package main

import (
	"testing"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/stretchr/testify/assert"
)

var testLogger = shim.NewLogger("screw_example_test")

func checkState(t *testing.T, stub *shim.MockStub, key string, value string) {
	bytes := stub.State[key]

	isNotNil := assert.NotNil(t, bytes)
	if !isNotNil {
		testLogger.Info("State", key, "failed to get value")
		t.FailNow()
	}

	isEqual := assert.Equal(t, value, string(bytes))
	if !isEqual {
		testLogger.Infof("State value %v was %v, not %v as expected", string(bytes), value)
		t.FailNow()
	} else {
		testLogger.Infof("State value %v is %v as expected", string(bytes))
	}
}

func checkNoState(t *testing.T, stub *shim.MockStub, key string) {
	bytes := stub.State[key]
	if bytes != nil {
		testLogger.Infof("State %v should be absent; found value", key)
		t.FailNow()
	} else {
		testLogger.Infof("State %v is absent as it should be", key)
	}
}

// Creates a MockStub bound to the chaincode struct ScrewInventory.
func createMockStub(stubName string) *shim.MockStub {
	si := new(ScrewInventory)
	mockStub := shim.NewMockStub(stubName, si)

	return mockStub
}

// Initializes the chaincode with the specified parameters using mockStub.MockInit.
func initChaincode(mockStub *shim.MockStub, arguments [][]byte) *peer.Response {
	resp := mockStub.MockInit("1", arguments)

	return &resp
}

func TestInit(t *testing.T) {
	mockStub := createMockStub("Test: Init")

	// Expect the chaincode to be initialized normally
	arguments := [][]byte{[]byte("init"), []byte("CorpA"), []byte("200"), []byte("CorpB"), []byte("100")}
	resp := initChaincode(mockStub, arguments)

	if resp.Status != shim.OK {
		testLogger.Infof("Initialization failed: %v", string(resp.Message))
		t.FailNow()
	}
}

func TestInovkeQuery(t *testing.T) {
	mockStub := createMockStub("Test: Invoke Query")

	// Expect the chaincode to be initialized normally
	initArguments := [][]byte{[]byte("init"), []byte("CorpA"), []byte("200"), []byte("CorpB"), []byte("100")}
	initResp := initChaincode(mockStub, initArguments)

	if initResp.Status != shim.OK {
		testLogger.Infof("Initialization failed: %v", string(initResp.Message))
		t.FailNow()
	}

	// Prepare the information needed to invoke the function
	invokeFunction := []byte("query")
	invokeArgument := []byte("CorpA")
	expectedValue := "200"

	// Invoke query and expect the payload to be correct
	invokeResp := mockStub.MockInvoke("1", [][]byte{invokeFunction, invokeArgument})

	if invokeResp.Status != shim.OK {
		testLogger.Infof("%s failed: %v", invokeFunction, string(invokeResp.Message))
		t.FailNow()
	}
	if invokeResp.Payload == nil {
		testLogger.Infof("%s failed: cannot get the value", invokeFunction)
		t.FailNow()
	}

	payload := string(invokeResp.Payload)
	if payload != expectedValue {
		testLogger.Infof("%s failed: value was %v instead of %v as expected", invokeFunction, payload, expectedValue)
		t.FailNow()
	}

	testLogger.Infof("%s invoked. Got %v as expected", invokeFunction, payload)
}

func TestInvokeTransfer(t *testing.T) {
	mockStub := createMockStub("Test: Invoke Transfer")

	// Expect the chaincode to be initialized normally
	initArguments := [][]byte{[]byte("init"), []byte("CorpA"), []byte("200"), []byte("CorpB"), []byte("100")}
	initResp := initChaincode(mockStub, initArguments)

	if initResp.Status != shim.OK {
		testLogger.Infof("Initialization failed: %v", string(initResp.Message))
		t.FailNow()
	}

	// Invoke transfer and expect the payload to be correct
	invokeFunction := []byte("transfer")
	invokeResp := mockStub.MockInvoke("1", [][]byte{invokeFunction, []byte("CorpA"), []byte("CorpB"), []byte("10")})

	if invokeResp.Status != shim.OK {
		testLogger.Infof("%s failed: %v", invokeFunction, string(invokeResp.Message))
		t.FailNow()
	}

	testLogger.Infof("%s invoked", invokeFunction)

	// Invoke query and expect the result to be changed
	invokeFunction = []byte("query")
	expectedValue := "190"
	invokeResp = mockStub.MockInvoke("1", [][]byte{invokeFunction, []byte("CorpA")})
	payload := string(invokeResp.Payload)
	if invokeResp.Status != shim.OK {
		testLogger.Infof("%s failed: %v", invokeFunction, string(invokeResp.Message))
		t.FailNow()
	}
	if payload != expectedValue {
		testLogger.Infof("%s failed: value was %v instead of %v as expected", invokeFunction, payload, expectedValue)
		t.FailNow()
	}

	testLogger.Infof("%s invoked. Got %v as expected", invokeFunction, payload)
}
