package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/bits"
	"testing"

	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/errorcode"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/data"
	"github.com/google/uuid"
)

const (
	data1 = "data1"
	data2 = "data2"
	data3 = "data3"
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
	stub := createMockStub(t, "TestCreateOffchainDataWithNormalData")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the arg
	sampleOffchainData1 := getSampleOffchainData1()
	resourceID := sampleOffchainData1.Metadata.ResourceID
	dataBytes, _ := json.Marshal(sampleOffchainData1)

	// Invoke with createOffchainData and expect the response status to be OK
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
	expectEqual(t, sampleOffchainData1.Metadata.Size, metadataStored.Size)
}

func TestCreateOffchainDataWithExcessiveParameters(t *testing.T) {
	targetFunction := "createOffchainData"
	stub := createMockStub(t, "TestCreateOffchainDataWithExcessiveParameters")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the arg
	sampleOffchainData1 := getSampleOffchainData1()
	dataBytes, _ := json.Marshal(sampleOffchainData1)

	// Invoke getSampleOffchainData1 with excessive parameters and expect the response status to be ERROR
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes, []byte("someEventID"), []byte("EXCESSIVE PARAMETER")})
	expectResponseStatusERROR(t, &resp)
}

func TestCreateOffchainDataWithDuplicateResourceIDs(t *testing.T) {
	targetFunction := "createOffchainData"
	stub := createMockStub(t, "TestCreateOffchainDataWithDuplicateResourceIDs")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the args
	sampleOffchainData1 := getSampleOffchainData1()
	sampleOffchainData2 := getSampleOffchainData2()

	// Deliberately change the resource ID of data2 to be the same as data1
	sampleOffchainData2.Metadata.ResourceID = sampleOffchainData1.Metadata.ResourceID
	data1Bytes, _ := json.Marshal(sampleOffchainData1)
	data2Bytes, _ := json.Marshal(sampleOffchainData1)

	// Invoke with data1 and expect the response status to be OK
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), data1Bytes})
	expectResponseStatusOK(t, &resp)

	// Invoke with data2 and expect the response status to be ERROR
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), data2Bytes})
	expectResponseStatusERROR(t, &resp)
}

func TestCreateOffchainDataIndexStatus(t *testing.T) {
	targetFunction := "createOffchainData"
	stub := createMockStub(t, "TestCreateOffchainDataIndexStatus")
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

	// Invoke with createPlainData and expect the response status to be OK
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes})
	expectResponseStatusOK(t, &resp)

	// Prepare the arg
	targetFunction = "getMetadata"
	resourceID = samplePlainData1.Metadata.ResourceID

	// Invoke with getMetadata and expect the response status to be OK
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), []byte(resourceID)})
	expectResponseStatusOK(t, &resp)
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

	// Invoke with samplePlainData1 and expect the response status to be OK
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes})
	expectResponseStatusOK(t, &resp)

	// Prepare the arg
	targetFunction = "getMetadata"

	// Invoke with getMetadata and expect the response status to be ERROR
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), []byte(resourceID), []byte("EXCESSIVE PARAMETER")})
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
	expectResponseStatusOK(t, &resp)

	// Prepare the arg
	targetFunction = "getMetadata"
	samplePlainData2 := getSamplePlainData2()
	resourceID = samplePlainData2.Metadata.ResourceID

	// Invoke with getMetadata and expect the response status to be ERROR
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), []byte(resourceID)})
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

	// Invoke with samplePlainData1 and expect the response status to be OK
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes})
	expectResponseStatusOK(t, &resp)

	// Invoke with getData and expect the response status to be OK
	targetFunction = "getData"
	resourceID = samplePlainData1.Metadata.ResourceID
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), []byte(resourceID)})
	expectResponseStatusOK(t, &resp)

	//check the return result
	var i string = string(resp.Payload[:])
	i = base64.StdEncoding.EncodeToString([]byte(i))
	expectEqual(t, samplePlainData1.Data, i)
}

func TestGetDataWithExcessiveParameters(t *testing.T) {
	targetFunction := "getData"
	stub := createMockStub(t, "TestGetDataWithExcessiveParameters")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the args
	samplePlainData1 := getSamplePlainData1()

	// Invoke with excessive parameters and expect the status to be ERROR
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), []byte(samplePlainData1.Metadata.ResourceID), []byte("EXCESSIVE PARAMETER")})
	expectResponseStatusERROR(t, &resp)
}

func TestGetDataWithNonExistentID(t *testing.T) {
	targetFunction := "createPlainData"
	stub := createMockStub(t, "TestGetDataWithNonExistentID")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the arg
	samplePlainData1 := getSamplePlainData1()
	resourceID := samplePlainData1.Metadata.ResourceID
	dataBytes, _ := json.Marshal(samplePlainData1)

	// Invoke with samplePlainData1 and expect the response status to be OK
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes})
	expectResponseStatusOK(t, &resp)

	// Prepare the args to test getData
	targetFunction = "getData"
	nonExistentResourceID := resourceID + "_NON_EXISTENT"

	// Invoke with a non existent resource ID and expect the response status to be ERROR
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), []byte(nonExistentResourceID)})
	expectResponseStatusERROR(t, &resp)
	expectStringEndsWith(t, errorcode.CodeNotFound, resp.Message)
}

func TestGetKey(t *testing.T) {
	targetFunction := "createEncryptedData"
	stub := createMockStub(t, "TestGetKey")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the EncryptedData
	sampleEncryptedData1 := getSampleEncryptedData1()
	dataBytes, _ := json.Marshal(sampleEncryptedData1)
	resourceID := sampleEncryptedData1.Metadata.ResourceID
	dataBytes, _ = json.Marshal(sampleEncryptedData1)

	// Invoke with createEncryptedData and expect the response status to be OK
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes})
	expectResponseStatusOK(t, &resp)

	targetFunction = "getKey"
	resourceID = sampleEncryptedData1.Metadata.ResourceID

	// Invoke with getKey and expect the response status to be ok
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), []byte(resourceID)})
	expectResponseStatusOK(t, &resp)
	sampleEncryptedData2 := getSampleEncryptedData1()
	json.Unmarshal(resp.Payload, &sampleEncryptedData2.Key)

	//check the return result
	expectEqual(t, sampleEncryptedData1.Key, sampleEncryptedData2.Key)
}

func TestGetKeyWithExcessiveParameters(t *testing.T) {
	stub := createMockStub(t, "TestGetKeyWithExcessiveParameters")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the arg
	targetFunction := "getKey"
	sampleEncryptedData1 := getSampleEncryptedData1()

	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), []byte(sampleEncryptedData1.Metadata.ResourceID), []byte("EXCESSIVE PARAMETER")})
	expectResponseStatusERROR(t, &resp)
}

func TestGetKeyWithNonExistentID(t *testing.T) {
	targetFunction := "createEncryptedData"
	stub := createMockStub(t, "TestGetKeyWithNonExistentID")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the arg
	sampleEncryptedData1 := getSampleEncryptedData1()

	dataBytes, _ := json.Marshal(sampleEncryptedData1)
	resourceID := sampleEncryptedData1.Metadata.ResourceID

	// Invoke with createEncryptedData and expect the response status to be OK
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes})
	expectResponseStatusOK(t, &resp)

	// Prepare a non-existent resource ID to test getKey
	nonExistentResourceID := resourceID + "_NON_EXISTENT"
	targetFunction = "getKey"

	// Invoke with getKey and expect the response status to be ERROR
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), []byte(nonExistentResourceID)})
	expectResponseStatusERROR(t, &resp)
	expectStringEndsWith(t, errorcode.CodeNotFound, resp.Message)

}

func TestGetPolicy(t *testing.T) {
	targetFunction := "createEncryptedData"
	stub := createMockStub(t, "TestGetPolicy")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the arg with encyprted data
	sampleEncryptedData1 := getSampleEncryptedData1()
	dataBytes, _ := json.Marshal(sampleEncryptedData1)
	resourceID := sampleEncryptedData1.Metadata.ResourceID

	// Invoke with createEncryptedData and expect the response status to be OK
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes})
	expectResponseStatusOK(t, &resp)

	// Test getPolicy
	targetFunction = "getPolicy"

	// Invoke with getPolicy and expect the response status to be OK
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), []byte(resourceID)})
	expectResponseStatusOK(t, &resp)

	// Expect the result to be the same
	expectEqual(t, resp.Payload, []byte(sampleEncryptedData1.Policy))

	// Prepare the arg with offchain data and do it again
	sampleOffchainData2 := getSampleOffchainData2()
	dataBytes, _ = json.Marshal(sampleOffchainData2)
	resourceID = sampleOffchainData2.Metadata.ResourceID
	targetFunction = "createOffchainData"
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes})
	expectResponseStatusOK(t, &resp)
	targetFunction = "getPolicy"
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), []byte(resourceID)})
	expectResponseStatusOK(t, &resp)
	expectEqual(t, resp.Payload, []byte(sampleOffchainData2.Policy))
}

func TestGetPolicyWithExcessiveParameters(t *testing.T) {
	stub := createMockStub(t, "TestGetPolicyWithExcessiveParameters")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the args
	targetFunction := "getPolicy"
	sampleEncryptedData1 := getSampleEncryptedData1()

	resourceID := sampleEncryptedData1.Metadata.ResourceID

	// Invoke with excessive parameters and expect the response status to be ERROR
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), []byte(resourceID), []byte("EXCESSIVE PARAMETER")})
	expectResponseStatusERROR(t, &resp)
}

func TestGetPolicyWithNonExistentID(t *testing.T) {
	// createEncryptedData
	targetFunction := "createEncryptedData"
	stub := createMockStub(t, "TestGetPolicyWithNonExistentID")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the arg
	sampleEncryptedData1 := getSampleEncryptedData1()

	dataBytes, _ := json.Marshal(sampleEncryptedData1)
	resourceID := sampleEncryptedData1.Metadata.ResourceID
	// Invoke with sampleEncryptedData1 and expect the response status to be OK
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes})
	expectResponseStatusOK(t, &resp)

	// Invoke getPolicy with a non existent resource ID
	nonExistentResourceID := resourceID + "_NON_EXISTENT"
	targetFunction = "getPolicy"

	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), []byte(nonExistentResourceID)})
	expectResponseStatusERROR(t, &resp)
	expectStringEndsWith(t, errorcode.CodeNotFound, resp.Message)
}

func TestLinkEntityIDWithDocumentID(t *testing.T) {
	stub := createMockStub(t, "TestLinkEntityIDWithDocumentID")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the args
	asset := getSamplePlainData1()
	doc2 := getSamplePlainData2()
	doc3 := getSamplePlainData3()

	assetBytes, _ := json.Marshal(asset)
	doc2Bytes, _ := json.Marshal(doc2)
	doc3Bytes, _ := json.Marshal(doc3)

	// Invoke to upload the data and expect the response status to be OK
	targetFunction := "createPlainData"
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), assetBytes})
	expectResponseStatusOK(t, &resp)
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), doc2Bytes})
	expectResponseStatusOK(t, &resp)
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), doc3Bytes})
	expectResponseStatusOK(t, &resp)

	targetFunction = "linkEntityIDWithDocumentID"
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), []byte(asset.Metadata.ResourceID), []byte(doc2.Metadata.ResourceID)})
	expectResponseStatusOK(t, &resp)
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), []byte(asset.Metadata.ResourceID), []byte(doc3.Metadata.ResourceID)})
	expectResponseStatusOK(t, &resp)

	// Query the composite key and expect to get 2 entries with values of the resource IDs we've just linked to
	expectedResourceIDs := make(map[string]bool)
	expectedResourceIDs[doc2.Metadata.ResourceID] = true
	expectedResourceIDs[doc3.Metadata.ResourceID] = true

	ckObjectType := "entityid~documentid"
	it, err := stub.GetStateByPartialCompositeKey(ckObjectType, []string{asset.Metadata.ResourceID})
	expectNil(t, err)

	defer it.Close()

	resourceIDs := []string{}
	for it.HasNext() {
		entry, err := it.Next()
		expectNil(t, err)

		_, ckParts, err := stub.SplitCompositeKey(entry.Key)
		expectNil(t, err)

		resourceIDs = append(resourceIDs, ckParts[1])
	}

	expectEqual(t, len(expectedResourceIDs), len(resourceIDs))
	for _, resourceID := range resourceIDs {
		expectEqual(t, true, expectedResourceIDs[resourceID])
	}
}

func TestLinkEntityIDWithDocumentIDWithExcessiveParameters(t *testing.T) {
	stub := createMockStub(t, "TestLinkEntityIDWithDocumentIDWithExcessiveParameters")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the args
	asset := getSamplePlainData1()
	doc2 := getSamplePlainData2()

	// Invoke to upload the data and expect the response status to be OK
	targetFunction := "linkEntityIDWithDocumentID"
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), []byte(asset.Metadata.ResourceID), []byte(doc2.Metadata.ResourceID), []byte("EXCESSIVE PARAMETER")})
	expectResponseStatusERROR(t, &resp)
}

func TestLinkEntityIDWithDocumentIDWithNonExistentEntityID(t *testing.T) {
	stub := createMockStub(t, "TestLinkEntityIDWithDocumentIDWithNonExistentEntityID")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the args
	asset := getSamplePlainData1()
	doc2 := getSamplePlainData2()

	assetBytes, _ := json.Marshal(asset)
	doc2Bytes, _ := json.Marshal(doc2)

	// Invoke to upload the data and expect the response status to be OK
	targetFunction := "createPlainData"
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), assetBytes})
	expectResponseStatusOK(t, &resp)
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), doc2Bytes})
	expectResponseStatusOK(t, &resp)

	targetFunction = "linkEntityIDWithDocumentID"
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), []byte(asset.Metadata.ResourceID), []byte(doc2.Metadata.ResourceID)})
	expectResponseStatusOK(t, &resp)

	// Check with a non-existent entity ID and expect the results to be empty
	doc3 := getSamplePlainData3()

	ckObjectType := "entityid~documentid"
	it, err := stub.GetStateByPartialCompositeKey(ckObjectType, []string{doc3.Metadata.ResourceID})
	expectNil(t, err)

	defer it.Close()

	expectEqual(t, false, it.HasNext())
}

func TestListDocumentIDsByEntityID(t *testing.T) {
	stub := createMockStub(t, "TestLinkEntityIDWithDocumentID")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the args
	asset := getSamplePlainData1()
	doc2 := getSamplePlainData2()
	doc3 := getSamplePlainData3()

	assetBytes, _ := json.Marshal(asset)
	doc2Bytes, _ := json.Marshal(doc2)
	doc3Bytes, _ := json.Marshal(doc3)

	// Invoke to upload the data and expect the response status to be OK
	targetFunction := "createPlainData"
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), assetBytes})
	expectResponseStatusOK(t, &resp)
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), doc2Bytes})
	expectResponseStatusOK(t, &resp)
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), doc3Bytes})
	expectResponseStatusOK(t, &resp)

	targetFunction = "linkEntityIDWithDocumentID"
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), []byte(asset.Metadata.ResourceID), []byte(doc2.Metadata.ResourceID)})
	expectResponseStatusOK(t, &resp)
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), []byte(asset.Metadata.ResourceID), []byte(doc3.Metadata.ResourceID)})
	expectResponseStatusOK(t, &resp)

	// Test listDocumentIDsByEntityID
	targetFunction = "listDocumentIDsByEntityID"
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), []byte(asset.Metadata.ResourceID)})
	results := []string{}
	err := json.Unmarshal(resp.Payload, &results)
	expectNil(t, err)

	// Expect the received results to be the same as expected
	expectedResults := make(map[string]bool)
	expectedResults[doc2.Metadata.ResourceID] = true
	expectedResults[doc3.Metadata.ResourceID] = true

	expectEqual(t, len(expectedResults), len(results))

	for _, resourceID := range results {
		expectEqual(t, true, expectedResults[resourceID])
	}
}

func TestListDocumentIDsByEntityIDWithExcessiveParameters(t *testing.T) {
	stub := createMockStub(t, "TestListDocumentIDsByEntityIDWithExcessiveParameters")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the args
	doc1 := getSamplePlainData1()

	// Invoke to upload the data and expect the response status to be OK
	targetFunction := "listDocumentIDsByEntityID"
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), []byte(doc1.Metadata.ResourceID), []byte("EXCESSIVE PARAMETER")})
	expectResponseStatusERROR(t, &resp)
}

func TestListDocumentIDsByNonExistentEntityID(t *testing.T) {
	stub := createMockStub(t, "TestLinkEntityIDWithDocumentID")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the args
	asset := getSamplePlainData1()
	doc2 := getSamplePlainData2()

	assetBytes, _ := json.Marshal(asset)
	doc2Bytes, _ := json.Marshal(doc2)

	// Invoke to upload the data and expect the response status to be OK
	targetFunction := "createPlainData"
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), assetBytes})
	expectResponseStatusOK(t, &resp)
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), doc2Bytes})
	expectResponseStatusOK(t, &resp)

	targetFunction = "linkEntityIDWithDocumentID"
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), []byte(asset.Metadata.ResourceID), []byte(doc2.Metadata.ResourceID)})
	expectResponseStatusOK(t, &resp)

	// Test listDocumentIDsByEntityID with a non-existent entity ID
	doc3 := getSamplePlainData3()
	targetFunction = "listDocumentIDsByEntityID"
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), []byte(doc3.Metadata.ResourceID)})
	expectResponseStatusOK(t, &resp)
	results := []string{}
	err := json.Unmarshal(resp.Payload, &results)
	expectNil(t, err)

	// Expect the received list to be empty
	expectEqual(t, 0, len(results))
}

func TestListDocumentIDsByCreatorWithNormalProcess(t *testing.T) {
	stub := createMockStub(t, "TestListDocumentIDsByCreatorWithNormalProcess")
	_ = initChaincode(stub, [][]byte{})

	// Prepare the environment: upload two different types of docs as user 1
	doc1 := getSamplePlainData1()
	doc2 := getSampleEncryptedData2()

	doc1Bytes, _ := json.Marshal(doc1)
	doc2Bytes, _ := json.Marshal(doc2)

	targetFunction := "createPlainData"
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), doc1Bytes})
	expectResponseStatusOK(t, &resp)

	targetFunction = "createEncryptedData"
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), doc2Bytes})
	expectResponseStatusOK(t, &resp)

	// Test listDocumentIDsByCreator and expect to receive 2 resource IDs
	targetFunction = "listDocumentIDsByCreator"
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction)})
	expectResponseStatusOK(t, &resp)

	expectedResourceIDs := make(map[string]bool)
	expectedResourceIDs[doc1.Metadata.ResourceID] = true
	expectedResourceIDs[doc2.Metadata.ResourceID] = true

	results := []string{}
	err := json.Unmarshal(resp.Payload, &results)
	expectNil(t, err)

	expectEqual(t, len(expectedResourceIDs), len(results))
	for _, resourceID := range results {
		expectEqual(t, true, expectedResourceIDs[resourceID])
	}

	// Change the user to user 2 and query again
	setMockStubCreator(t, stub, "Org1MSP", []byte(exampleCertUser2))
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction)})
	expectResponseStatusOK(t, &resp)
	results = []string{}
	err = json.Unmarshal(resp.Payload, &results)
	expectNil(t, err)

	expectEqual(t, 0, len(results))
}

func TestListDocumentIDsByCreatorWithExcessiveParameters(t *testing.T) {
	stub := createMockStub(t, "TestListDocumentIDsByCreatorWithExcessiveParameters")
	_ = initChaincode(stub, [][]byte{})

	targetFunction := "listDocumentIDsWithCreator"
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), []byte("EXCESSIVE PARAMETER")})
	expectResponseStatusERROR(t, &resp)
}

func TestListDocumentIDsByPartialNameWithExcessiveParameters(t *testing.T) {
	stub := createMockStub(t, "TestListDocumentIDsByPartialNameWithExcessiveParameters")
	_ = initChaincode(stub, [][]byte{})

	// Test listDocumentIDsByCreator
	targetFunction := "listDocumentIDsByPartialName"
	partialName := "1"
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), []byte(partialName), []byte("5"), []byte(""), []byte("EXCESSIVE_PARAMETER")})
	expectResponseStatusERROR(t, &resp)
}

// 明文数据
// 资源 ID: "001"
// 名称: "Sample Plain Data 1"
// 内容: base64(data1)
func getSamplePlainData1() data.PlainData {
	return data.PlainData{
		Metadata: data.ResMetadata{
			ResourceType: data.Plain,
			ResourceID:   "001",
			Hash:         sha256.Sum256([]byte(data1)),
			Size:         uint64(len([]byte(data1))),
			Extensions:   "{\"name\":\"Sample Plain Data 1\"}",
		},
		Data: base64.StdEncoding.EncodeToString([]byte(data1)),
	}
}

// 明文数据
// 资源 ID: "002"
// 名称: "示例明文数据2"
// 内容: base64(data2)
func getSamplePlainData2() data.PlainData {
	return data.PlainData{
		Metadata: data.ResMetadata{
			ResourceType: data.Plain,
			ResourceID:   "002",
			Hash:         sha256.Sum256([]byte(data2)),
			Size:         uint64(len([]byte(data2))),
			Extensions:   "{\"name\":\"示例明文数据2\"}",
		},
		Data: base64.StdEncoding.EncodeToString([]byte(data2)),
	}
}

// 明文数据
// 资源 ID: "003"
// 名称: "示例明文数据3"
// 内容: base64(data3)
func getSamplePlainData3() data.PlainData {
	return data.PlainData{
		Metadata: data.ResMetadata{
			ResourceType: data.Plain,
			ResourceID:   "003",
			Hash:         sha256.Sum256([]byte(data3)),
			Size:         uint64(len([]byte(data3))),
			Extensions:   "{\"name\":\"示例明文数据3\"}",
		},
		Data: base64.StdEncoding.EncodeToString([]byte(data3)),
	}
}

// 加密数据
// 资源 ID: "001"
// 名称: "Sample Encrypted Data 1"
// 内容：base64(encrypt(data1))
func getSampleEncryptedData1() data.EncryptedData {
	return data.EncryptedData{
		Metadata: data.ResMetadata{
			ResourceType: data.Encrypted,
			ResourceID:   "101",
			Hash:         sha256.Sum256([]byte(data1)),
			Size:         uint64(len([]byte(data1))),
			Extensions:   "{\"name\":\"Sample Encrypted Data 1\"}",
		},
		// TODO: 需要用对称密钥加密
		Data:   base64.StdEncoding.EncodeToString([]byte(data1)),
		Key:    []byte("123456"),
		Policy: `(DeptType == "computer" && DeptLevel == 2)`,
	}
}

// 加密数据
// 资源 ID: "002"
// 名称: "示例加密数据2"
// 内容：base64(encrypt(data2))
func getSampleEncryptedData2() data.EncryptedData {
	return data.EncryptedData{
		Metadata: data.ResMetadata{
			ResourceType: data.Encrypted,
			ResourceID:   "102",
			Hash:         sha256.Sum256([]byte(data2)),
			Size:         uint64(len([]byte(data2))),
			Extensions:   "{\"name\":\"示例加密数据2\"}",
		},
		// TODO: 需要用对称密钥加密
		Data:   base64.StdEncoding.EncodeToString([]byte(data2)),
		Key:    []byte("123456"),
		Policy: `(DeptType == "computer" && DeptLevel == 1)`,
	}
}

// 链下数据
// 资源 ID: "001"
// 名称: "Sample Offchain Data 1"
func getSampleOffchainData1() data.OffchainData {
	return data.OffchainData{
		Metadata: data.ResMetadata{
			ResourceType: data.Offchain,
			ResourceID:   "201",
			Hash:         sha256.Sum256([]byte(data1)),
			Size:         uint64(len([]byte(data1))),
			Extensions:   "{\"name\":\"Sample Offchain Data 1\"}",
		},
		Key: []byte("123456"),
		// TODO: 测试时填充
		Policy: "Encryption strategy",
	}
}

// 链下数据
// 资源 ID: "002"
// 名称: "示例链下数据2"
func getSampleOffchainData2() data.OffchainData {
	return data.OffchainData{
		Metadata: data.ResMetadata{
			ResourceType: data.Offchain,
			ResourceID:   "202",
			Hash:         sha256.Sum256([]byte(data2)),
			Size:         uint64(len([]byte(data2))),
			Extensions:   "{\"name\":\"示例链下数据2\"}",
		},
		Key: []byte("123456"),
		// TODO: 测试时填充
		Policy: "Encryption strategy",
	}
}
