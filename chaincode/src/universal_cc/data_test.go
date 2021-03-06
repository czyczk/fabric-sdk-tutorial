package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/bits"
	"testing"

	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/data"
	"github.com/google/uuid"
)

const (
	data1 = "data1"
	data2 = "data2"
)

func TestCreatePlainDataWithNormalData(t *testing.T) {
	targetFunction := "createPlainData"

	stub := createMockStub(t, "TestCreatePlainDataWithNormalData")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the arg
	samplePlainData1 := getSamplePlainData1()
	resourceID := samplePlainData1.Metadata.ResourceID
	dataBytes, _ := json.Marshal(samplePlainData1)

	// Invoke with samplePlainData1 and expect the response status to be OK
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes})
	expectResponseStatusOK(t, &resp)

	// Check if the data in State is as expected
	expectStateEqual(t, stub, getKeyForResData(resourceID), []byte(data1))

	// Check if the stored metadata can be parsed
	metadataStoredBytes := stub.State[getKeyForResMetadata(resourceID)]
	metadataStored := data.ResMetadataStored{}
	if err := json.Unmarshal(metadataStoredBytes, &metadataStored); err != nil {
		testLogger.Infof("Cannot read stored metadata: %v\n", err)
		t.FailNow()
	}

	// Check if the stored metadata is correct
	expectEqual(t, resourceID, metadataStored.ResourceID)
	expectEqual(t, samplePlainData1.Metadata.ResourceType, metadataStored.ResourceType)
	expectEqual(t, samplePlainData1.Metadata.Hash, metadataStored.Hash)
	expectEqual(t, samplePlainData1.Metadata.Hash, metadataStored.HashStored)
	expectEqual(t, samplePlainData1.Metadata.Size, metadataStored.Size)
	expectEqual(t, samplePlainData1.Metadata.Size, metadataStored.SizeStored)
}

func TestCreatePlainDataWithDuplicateResourceIDs(t *testing.T) {
	targetFunction := "createPlainData"

	stub := createMockStub(t, "TestCreatePlainDataWithDuplicateResourceIDs")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the args
	samplePlainData1 := getSamplePlainData1()
	samplePlainData2 := getSamplePlainData2()
	// Deliberately change the resource ID of data2 to be the same as data1
	samplePlainData2.Metadata.ResourceID = samplePlainData1.Metadata.ResourceID

	data1Bytes, _ := json.Marshal(samplePlainData1)

	// Invoke with data1 and expect the response status to be OK
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), data1Bytes})
	expectResponseStatusOK(t, &resp)

	// Invoke with data2 and expect the response status to be ERROR
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), data1Bytes})
	expectResponseStatusERROR(t, &resp)
}

func TestCreatePlainDataWithExcessiveParameters(t *testing.T) {
	targetFunction := "createPlainData"

	stub := createMockStub(t, "TestCreatePlainDataWithExcessiveParameters")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the arg
	samplePlainData1 := getSamplePlainData1()
	dataBytes, _ := json.Marshal(samplePlainData1)

	// Invoke with samplePlainData1 with excessive parameters and expect the response status to be ERROR
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes, []byte("someEventID"), []byte("EXCESSIVE PARAMETER")})
	expectResponseStatusERROR(t, &resp)
}

func TestCreatePlainDataWithCorruptHashAndSize(t *testing.T) {
	targetFunction := "createPlainData"

	stub := createMockStub(t, "TestCreatePlainDataWithCorruptHashAndSize")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the args
	samplePlainDataCorruptHash := getSamplePlainData1()
	samplePlainDataCorruptHash.Metadata.Hash[13] = bits.Reverse8(samplePlainDataCorruptHash.Metadata.Hash[13])
	samplePlainDataCorruptHashBytes, _ := json.Marshal(samplePlainDataCorruptHash)

	samplePlainDataCorruptSize := getSamplePlainData2()
	samplePlainDataCorruptSize.Metadata.Size = samplePlainDataCorruptSize.Metadata.Size + 233
	samplePlainDataCorruptSizeBytes, _ := json.Marshal(samplePlainDataCorruptSize)

	// Invoke with dataCorruptHash and expect the response status to be ERROR
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), samplePlainDataCorruptHashBytes})
	expectResponseStatusOK(t, &resp)

	// Invoke with dataCorruptSize and expect the response status to be ERROR
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), samplePlainDataCorruptSizeBytes})
	expectResponseStatusERROR(t, &resp)
}

func TestCreatePlainDataIndexStatus(t *testing.T) {
	targetFunction := "createPlainData"

	stub := createMockStub(t, "TestCreatePlainDataIndexStatus")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the args
	samplePlainData1 := getSamplePlainData1()
	samplePlainData2 := getSamplePlainData2()

	samplePlainData1Bytes, _ := json.Marshal(samplePlainData1)
	samplePlainData2Bytes, _ := json.Marshal(samplePlainData2)

	// Invoke with data1 and expect the response status to be OK
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), samplePlainData1Bytes})
	expectResponseStatusOK(t, &resp)

	// Invoke with data2 and expect the response status to be OK
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), samplePlainData2Bytes})
	expectResponseStatusOK(t, &resp)

	creator, err := getPKDERFromCertString(exampleCertUser1)
	if err != nil {
		testLogger.Infof("Error parsing certificate: %v\n", err)
		t.FailNow()
	}
	creatorAsBase64 := base64.StdEncoding.EncodeToString(creator)

	it, err := stub.GetStateByPartialCompositeKey("creator~resourceid", []string{creatorAsBase64})
	if err != nil {
		testLogger.Infof("Cannot query index 'creator~resourceid': %v\n", err)
		t.FailNow()
	}

	// Iterate the composite key value entries and expect 2 distinct resource IDs
	resourceIDSet := make(map[string]bool)
	for it.HasNext() {
		entry, _ := it.Next()
		_, ckParts, _ := stub.SplitCompositeKey(entry.Key)
		resourceIDSet[ckParts[1]] = true
	}

	expectEqual(t, 2, len(resourceIDSet))
}

func TestCreateEncryptedDataWithNormalData(t *testing.T) {
	targetFunction := "createEncryptedData"

	stub := createMockStub(t, "TestCreateEncryptedDataWithNormalData")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the arg
	sampleEncryptedData1 := getSampleEncryptedData1()
	resourceID := sampleEncryptedData1.Metadata.ResourceID
	dataBytes, _ := json.Marshal(sampleEncryptedData1)

	// Invoke with samplePlainData1 and expect the response status to be OK
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes})
	expectResponseStatusOK(t, &resp)

	// Check if the data in State is as expected
	expectStateEqual(t, stub, getKeyForResData(resourceID), []byte(data1))

	// Check if the stored metadata can be parsed
	metadataStoredBytes := stub.State[getKeyForResMetadata(resourceID)]
	metadataStored := data.ResMetadataStored{}
	if err := json.Unmarshal(metadataStoredBytes, &metadataStored); err != nil {
		testLogger.Infof("Cannot read stored metadata: %v\n", err)
		t.FailNow()
	}

	// Check if the stored metadata is correct
	expectEqual(t, resourceID, metadataStored.ResourceID)
	expectEqual(t, sampleEncryptedData1.Metadata.ResourceType, metadataStored.ResourceType)
	expectEqual(t, sampleEncryptedData1.Metadata.Hash, metadataStored.Hash)
	expectEqual(t, sampleEncryptedData1.Metadata.Hash, metadataStored.HashStored)
	expectEqual(t, sampleEncryptedData1.Metadata.Size, metadataStored.Size)
	expectEqual(t, sampleEncryptedData1.Metadata.Size, metadataStored.SizeStored)
}

func TestCreateEncryptedDataWithExcessiveParameters(t *testing.T) {
	targetFunction := "createEncryptedData"

	stub := createMockStub(t, "TestCreateEncryptedDataWithExcessiveParameters")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the arg
	sampleEncryptedData1 := getSampleEncryptedData1()
	dataBytes, _ := json.Marshal(sampleEncryptedData1)

	// Invoke with samplePlainData1 with excessive parameters and expect the response status to be ERROR
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes, []byte("someEventID"), []byte("EXCESSIVE PARAMETER")})
	expectResponseStatusERROR(t, &resp)
}

func TestCreateEncryptedDataWithDuplicateResourceIDs(t *testing.T) {
	targetFunction := "createEncryptedData"

	stub := createMockStub(t, "TestCreatePlainDataWithDuplicateResourceIDs")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the args
	sampleEncryptedData1 := getSampleEncryptedData1()
	sampleEncryptedData2 := getSampleEncryptedData2()
	// Deliberately change the resource ID of data2 to be the same as data1
	sampleEncryptedData2.Metadata.ResourceID = sampleEncryptedData1.Metadata.ResourceID

	data1Bytes, _ := json.Marshal(sampleEncryptedData1)
	data2Bytes, _ := json.Marshal(sampleEncryptedData1)
	// Invoke with data1 and expect the response status to be OK
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), data1Bytes})
	expectResponseStatusOK(t, &resp)

	// Invoke with data2 and expect the response status to be ERROR
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), data2Bytes})
	expectResponseStatusERROR(t, &resp)
}

func TestCreateEncryptedDataIndexStatus(t *testing.T) {
	targetFunction := "createEncryptedData"

	stub := createMockStub(t, "TestCreateEncryptedDataIndexStatus")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the args
	getSampleEncryptedData1 := getSampleEncryptedData1()
	getSampleEncryptedData2 := getSampleEncryptedData2()

	sampleEncryptedData1Bytes, _ := json.Marshal(getSampleEncryptedData1)
	sampleEncryptedData2Bytes, _ := json.Marshal(getSampleEncryptedData2)

	// Invoke with data1 and expect the response status to be OK
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), sampleEncryptedData1Bytes})
	expectResponseStatusOK(t, &resp)

	// Invoke with data2 and expect the response status to be OK
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), sampleEncryptedData2Bytes})
	expectResponseStatusOK(t, &resp)

	creator, err := getPKDERFromCertString(exampleCertUser1)
	if err != nil {
		testLogger.Infof("Error parsing certificate: %v\n", err)
		t.FailNow()
	}
	creatorAsBase64 := base64.StdEncoding.EncodeToString(creator)

	it, err := stub.GetStateByPartialCompositeKey("creator~resourceid", []string{creatorAsBase64})
	if err != nil {
		testLogger.Infof("Cannot query index 'creator~resourceid': %v\n", err)
		t.FailNow()
	}

	// Iterate the composite key value entries and expect 2 distinct resource IDs
	resourceIDSet := make(map[string]bool)
	for it.HasNext() {
		entry, _ := it.Next()
		_, ckParts, _ := stub.SplitCompositeKey(entry.Key)
		resourceIDSet[ckParts[1]] = true
	}

	expectEqual(t, 2, len(resourceIDSet))
}

func TestCreateOffchainDataWithNormalData(t *testing.T) {
	targetFunction := "createOffchainData"

	stub := createMockStub(t, "TestCreatePlainDataWithNormalData")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the arg
	sampleOffchainData1 := getSampleOffchainData1()
	resourceID := sampleOffchainData1.Metadata.ResourceID
	dataBytes, _ := json.Marshal(sampleOffchainData1)

	// Invoke with samplePlainData1 and expect the response status to be OK
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes})
	expectResponseStatusOK(t, &resp)

	// Check if the stored metadata can be parsed
	metadataStoredBytes := stub.State[getKeyForResMetadata(resourceID)]
	metadataStored := data.ResMetadataStored{}
	if err := json.Unmarshal(metadataStoredBytes, &metadataStored); err != nil {
		testLogger.Infof("Cannot read stored metadata: %v\n", err)
		t.FailNow()
	}

	// Check if the stored metadata is correct
	expectEqual(t, resourceID, metadataStored.ResourceID)
	expectEqual(t, sampleOffchainData1.Metadata.ResourceType, metadataStored.ResourceType)
	expectEqual(t, sampleOffchainData1.Metadata.Hash, metadataStored.Hash)
	expectEqual(t, sampleOffchainData1.Metadata.Hash, metadataStored.HashStored)
	expectEqual(t, sampleOffchainData1.Metadata.Size, metadataStored.Size)
	expectEqual(t, sampleOffchainData1.Metadata.Size, metadataStored.SizeStored)
}

func TestCreateOffchainDataWithExcessiveParameters(t *testing.T) {
	targetFunction := "createOffchainData"

	stub := createMockStub(t, "TestCreatePlainDataWithExcessiveParameters")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the arg
	sampleOffchainData1 := getSampleOffchainData1()
	dataBytes, _ := json.Marshal(sampleOffchainData1)

	// Invoke with samplePlainData1 with excessive parameters and expect the response status to be ERROR
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes, []byte("someEventID"), []byte("EXCESSIVE PARAMETER")})
	expectResponseStatusERROR(t, &resp)
}

func TestCreateOffchainDataWithDuplicateResourceIDs(t *testing.T) {
	targetFunction := "createOffchainData"

	stub := createMockStub(t, "TestCreatePlainDataWithDuplicateResourceIDs")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the args
	sampleOffchainData1 := getSampleOffchainData1()
	sampleOffchainData2 := getSampleOffchainData2()
	// Deliberately change the resource ID of data2 to be the same as data1
	sampleOffchainData2.Metadata.ResourceID = sampleOffchainData1.Metadata.ResourceID

	data1Bytes, _ := json.Marshal(sampleOffchainData1)

	// Invoke with data1 and expect the response status to be OK
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), data1Bytes})
	expectResponseStatusOK(t, &resp)

	// Invoke with data2 and expect the response status to be ERROR
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), data1Bytes})
	expectResponseStatusERROR(t, &resp)
}

func TestCreateOffchainDataIndexStatus(t *testing.T) {
	targetFunction := "createOffchainData"

	stub := createMockStub(t, "TestCreateEncryptedDataIndexStatus")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the args
	getSampleOffchainData1 := getSampleOffchainData1()
	getSampleOffchainData2 := getSampleOffchainData2()

	sampleOffchainData1Bytes, _ := json.Marshal(getSampleOffchainData1)
	sampleOffchainData2Bytes, _ := json.Marshal(getSampleOffchainData2)

	// Invoke with data1 and expect the response status to be OK
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), sampleOffchainData1Bytes})
	expectResponseStatusOK(t, &resp)

	// Invoke with data2 and expect the response status to be OK
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), sampleOffchainData2Bytes})
	expectResponseStatusOK(t, &resp)

	creator, err := getPKDERFromCertString(exampleCertUser1)
	if err != nil {
		testLogger.Infof("Error parsing certificate: %v\n", err)
		t.FailNow()
	}
	creatorAsBase64 := base64.StdEncoding.EncodeToString(creator)

	it, err := stub.GetStateByPartialCompositeKey("creator~resourceid", []string{creatorAsBase64})
	if err != nil {
		testLogger.Infof("Cannot query index 'creator~resourceid': %v\n", err)
		t.FailNow()
	}

	// Iterate the composite key value entries and expect 2 distinct resource IDs
	resourceIDSet := make(map[string]bool)
	for it.HasNext() {
		entry, _ := it.Next()
		_, ckParts, _ := stub.SplitCompositeKey(entry.Key)
		resourceIDSet[ckParts[1]] = true
	}

	expectEqual(t, 2, len(resourceIDSet))
}

func TestGetMetadata(t *testing.T) {
	targetFunction := "createPlainData"

	stub := createMockStub(t, "TestGetMetadata")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the arg
	samplePlainData1 := getSamplePlainData1()
	resourceID := samplePlainData1.Metadata.ResourceID
	dataBytes, _ := json.Marshal(samplePlainData1)

	// Invoke with samplePlainData1 and expect the response status to be OK
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes})

	targetFunction = "getMetadata"
	// Prepare the arg

	resourceID = samplePlainData1.Metadata.ResourceID
	dataBytes, _ = json.Marshal(resourceID)

	// Invoke with samplePlainData1 and expect the response status to be OK
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes})
	resMetadata := data.ResMetadata{}
	json.Unmarshal(resp.Payload, &resMetadata)
	//check the return result
	expectEqual(t, samplePlainData1.Metadata.ResourceID, resMetadata.ResourceID)
	expectEqual(t, samplePlainData1.Metadata.Hash, resMetadata.Hash)
	expectEqual(t, samplePlainData1.Metadata.Extensions, resMetadata.Extensions)
	expectEqual(t, samplePlainData1.Metadata.ResourceType, resMetadata.ResourceType)
	expectEqual(t, samplePlainData1.Metadata.Size, resMetadata.Size)
}

func TestGetMetadataWithExcessiveParameters(t *testing.T) {
	targetFunction := "createPlainData"

	stub := createMockStub(t, "TestGetMetadataWithExcessiveParameters")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the arg
	samplePlainData1 := getSamplePlainData1()

	dataBytes, _ := json.Marshal(samplePlainData1)
	resourceID := samplePlainData1.Metadata.ResourceID
	dataBytes, _ = json.Marshal(samplePlainData1)
	// Invoke with samplePlainData1 and expect the response status to be OK
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes})

	targetFunction = "getMetadata"
	// Prepare the arg
	dataBytes, _ = json.Marshal(resourceID)
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes, []byte("EXCESSIVE PARAMETER")})
	expectResponseStatusERROR(t, &resp)
}

func TestGetMetadataWithNonExistentID(t *testing.T) {
	targetFunction := "createPlainData"

	stub := createMockStub(t, "TestGetMetadataWithNonExistentID")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the arg
	samplePlainData1 := getSamplePlainData1()

	dataBytes, _ := json.Marshal(samplePlainData1)
	resourceID := samplePlainData1.Metadata.ResourceID
	dataBytes, _ = json.Marshal(samplePlainData1)
	// Invoke with samplePlainData1 and expect the response status to be OK
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes})

	targetFunction = "getMetadata"
	// Prepare the arg
	samplePlainData2 := getSamplePlainData2()
	resourceID = samplePlainData2.Metadata.ResourceID
	dataBytes, _ = json.Marshal(resourceID)

	// Invoke with samplePlainData1 and expect the response status to be OK
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes})
	expectResponseStatusERROR(t, &resp)
}

func TestGetData(t *testing.T) {
	targetFunction := "createPlainData"

	stub := createMockStub(t, "TestGetData")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the arg
	samplePlainData1 := getSamplePlainData1()

	dataBytes, _ := json.Marshal(samplePlainData1)
	resourceID := samplePlainData1.Metadata.ResourceID
	dataBytes, _ = json.Marshal(samplePlainData1)
	// Invoke with samplePlainData1 and expect the response status to be OK
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes})

	targetFunction = "getData"
	resourceID = samplePlainData1.Metadata.ResourceID
	dataBytes, _ = json.Marshal(resourceID)

	// Invoke with samplePlainData1 and expect the response status to be OK
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes})

	var i string = string(resp.Payload[:])
	//check the return result
	i = base64.StdEncoding.EncodeToString([]byte(i))
	expectEqual(t, samplePlainData1.Data, i)
}

func TestGetDataWithNonExistentID(t *testing.T) {
	targetFunction := "createPlainData"

	stub := createMockStub(t, "TestGetData")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the arg
	samplePlainData1 := getSamplePlainData1()

	dataBytes, _ := json.Marshal(samplePlainData1)
	resourceID := samplePlainData1.Metadata.ResourceID
	dataBytes, _ = json.Marshal(samplePlainData1)
	// Invoke with samplePlainData1 and expect the response status to be OK
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes})

	targetFunction = "getData"
	samplePlainData2 := getSamplePlainData2()
	resourceID = samplePlainData2.Metadata.ResourceID
	dataBytes, _ = json.Marshal(resourceID)

	// Invoke with samplePlainData1 and expect the response status to be OK
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes})
	expectResponseStatusERROR(t, &resp)

}

func TestGetKey(t *testing.T) {
	targetFunction := "createEncryptedData"

	stub := createMockStub(t, "TestGetData")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the arg
	sampleEncryptedData1 := getSampleEncryptedData1()

	dataBytes, _ := json.Marshal(sampleEncryptedData1)
	resourceID := sampleEncryptedData1.Metadata.ResourceID
	dataBytes, _ = json.Marshal(sampleEncryptedData1)
	// Invoke with samplePlainData1 and expect the response status to be OK
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes})

	targetFunction = "getKey"
	resourceID = sampleEncryptedData1.Metadata.ResourceID

	dataBytes, _ = json.Marshal(resourceID)

	// Invoke with samplePlainData1 and expect the response status to be OK
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes})

	sampleEncryptedData2 := getSampleEncryptedData1()
	json.Unmarshal(resp.Payload, &sampleEncryptedData2.Key)

	expectEqual(t, sampleEncryptedData1.Key, sampleEncryptedData2.Key)
}

func TestGetKeyWithNonExistentID(t *testing.T) {
	targetFunction := "createEncryptedData"

	stub := createMockStub(t, "TestGetKeyWithNonExistentID")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the arg
	sampleEncryptedData1 := getSampleEncryptedData1()

	dataBytes, _ := json.Marshal(sampleEncryptedData1)
	resourceID := sampleEncryptedData1.Metadata.ResourceID
	dataBytes, _ = json.Marshal(sampleEncryptedData1)
	// Invoke with samplePlainData1 and expect the response status to be OK
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes})

	sampleEncryptedData1 = getSampleEncryptedData2()
	targetFunction = "getKey"
	resourceID = sampleEncryptedData1.Metadata.ResourceID

	dataBytes, _ = json.Marshal(resourceID)

	// Invoke with samplePlainData1 and expect the response status to be OK
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes})
	expectResponseStatusERROR(t, &resp)

}

func TestGetPolicy(t *testing.T) {
	targetFunction := "createOffchainData"

	stub := createMockStub(t, "TestGetPolicy")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the arg
	sampleEncryptedData1 := getSampleEncryptedData1()

	dataBytes, _ := json.Marshal(sampleEncryptedData1)
	resourceID := sampleEncryptedData1.Metadata.ResourceID
	dataBytes, _ = json.Marshal(sampleEncryptedData1)
	// Invoke with samplePlainData1 and expect the response status to be OK
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes})

	targetFunction = "getPolicy"
	resourceID = sampleEncryptedData1.Metadata.ResourceID

	dataBytes, _ = json.Marshal(resourceID)

	// Invoke with samplePlainData1 and expect the response status to be OK
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes})

	var i string
	//check the return result
	json.Unmarshal(resp.Payload, &i)
	expectEqual(t, sampleEncryptedData1.Policy, i)
}

func TestGetPolicyWithNonExistentID(t *testing.T) {
	targetFunction := "createEncryptedData"

	stub := createMockStub(t, "TestGetPolicyWithNonExistentID")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the arg
	sampleEncryptedData1 := getSampleEncryptedData1()

	dataBytes, _ := json.Marshal(sampleEncryptedData1)
	resourceID := sampleEncryptedData1.Metadata.ResourceID
	dataBytes, _ = json.Marshal(sampleEncryptedData1)
	// Invoke with samplePlainData1 and expect the response status to be OK
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes})

	sampleEncryptedData1 = getSampleEncryptedData2()
	targetFunction = "getKey"
	resourceID = sampleEncryptedData1.Metadata.ResourceID

	dataBytes, _ = json.Marshal(resourceID)

	// Invoke with samplePlainData1 and expect the response status to be OK
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes})
	expectResponseStatusERROR(t, &resp)
}

func getSamplePlainData1() data.PlainData {
	return data.PlainData{
		Metadata: data.ResMetadata{
			ResourceType: data.Plain,
			ResourceID:   "001",
			Hash:         sha256.Sum256([]byte("data1")),
			Size:         uint64(len([]byte("data1"))),
			Extensions:   "{\"name\":\"examplePlainData1\"}",
		},
		Data: base64.StdEncoding.EncodeToString([]byte(data1)),
	}
}

func getSamplePlainData2() data.PlainData {
	return data.PlainData{
		Metadata: data.ResMetadata{
			ResourceType: data.Plain,
			ResourceID:   "002",
			Hash:         sha256.Sum256([]byte("data2")),
			Size:         uint64(len([]byte("data2"))),
			Extensions:   "{\"name\":\"examplePlainData2\"}",
		},
		Data: base64.StdEncoding.EncodeToString([]byte(data2)),
	}
}

func getSampleEncryptedData1() data.EncryptedData {
	return data.EncryptedData{
		Metadata: data.ResMetadata{
			ResourceType: data.Encrypted,
			ResourceID:   "001",
			Hash:         sha256.Sum256([]byte("data1")),
			Size:         uint64(len([]byte("data1"))),
			Extensions:   "{\"name\":\"exampleEncryptedData1\"}",
		},
		Data:   base64.StdEncoding.EncodeToString([]byte(data1)),
		Key:    []byte("123456"),
		Policy: "Encryption strategy",
	}
}

func getSampleEncryptedData2() data.EncryptedData {
	return data.EncryptedData{
		Metadata: data.ResMetadata{
			ResourceType: data.Encrypted,
			ResourceID:   "002",
			Hash:         sha256.Sum256([]byte("data2")),
			Size:         uint64(len([]byte("data2"))),
			Extensions:   "{\"name\":\"exampleEncryptedData2\"}",
		},
		Data:   base64.StdEncoding.EncodeToString([]byte(data2)),
		Key:    []byte("123456"),
		Policy: "Encryption strategy",
	}
}

func getSampleOffchainData1() data.OffchainData {
	return data.OffchainData{
		Metadata: data.ResMetadata{
			ResourceType: data.Offchain,
			ResourceID:   "001",
			Hash:         sha256.Sum256([]byte("data1")),
			Size:         uint64(len([]byte("data1"))),
			Extensions:   "{\"name\":\"exampleOffchainData1\"}",
		},
		Key:    []byte("123456"),
		Policy: "Encryption strategy",
	}
}

func getSampleOffchainData2() data.OffchainData {
	return data.OffchainData{
		Metadata: data.ResMetadata{
			ResourceType: data.Offchain,
			ResourceID:   "002",
			Hash:         sha256.Sum256([]byte("data1")),
			Size:         uint64(len([]byte("data1"))),
			Extensions:   "{\"name\":\"exampleOffchainData2\"}",
		},
		Key:    []byte("123456"),
		Policy: "Encryption strategy",
	}
}
