package main

import (
	"testing"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-chaincode-go/shimtest"
	"github.com/hyperledger/fabric-protos-go/peer"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var testLogger = log.StandardLogger()

func checkState(t *testing.T, stub *shimtest.MockStub, key string, value string) {
	bytes := stub.State[key]

	isNotNil := assert.NotNil(t, bytes)
	if !isNotNil {
		testLogger.Infof("Failed to get value for state of key '%v'", key)
		t.FailNow()
	}

	isEqual := assert.Equal(t, value, string(bytes))
	if !isEqual {
		testLogger.Infof("State value of key '%v' was '%v', not '%v' as expected", key, string(bytes), value)
		t.FailNow()
	} else {
		testLogger.Infof("State value '%v' is '%v' as expected", key, string(bytes))
	}
}

func checkNoState(t *testing.T, stub *shimtest.MockStub, key string) {
	bytes := stub.State[key]
	if bytes != nil {
		testLogger.Infof("State of key '%v' should be absent; found value", key)
		t.FailNow()
	} else {
		testLogger.Infof("State of key '%v' is absent as it should be", key)
	}
}

// Creates a MockStub bound to the chaincode struct ScrewInventory.
func createMockStub(stubName string) *shimtest.MockStub {
	si := new(ScrewInventory)
	mockStub := shimtest.NewMockStub(stubName, si)

	return mockStub
}

// Initializes the chaincode with the specified parameters using mockStub.MockInit.
func initChaincode(mockStub *shimtest.MockStub, arguments [][]byte) *peer.Response {
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
		testLogger.Infof("%s failed: value was '%v' instead of '%v' as expected", invokeFunction, payload, expectedValue)
		t.FailNow()
	}

	testLogger.Infof("%s invoked. Got '%v' as expected", invokeFunction, payload)
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
	invokeResp := mockStub.MockInvoke("1", [][]byte{invokeFunction, []byte("CorpA"), []byte("CorpB"), []byte("10"), []byte("eventTransfer")})

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
		testLogger.Infof("%s failed: value was '%v' instead of '%v' as expected", invokeFunction, payload, expectedValue)
		t.FailNow()
	}

	testLogger.Infof("%s invoked. Got '%v' as expected", invokeFunction, payload)
}
