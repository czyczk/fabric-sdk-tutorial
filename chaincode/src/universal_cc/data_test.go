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
	expectStateEqual(t, stub, getKeyForResData(resourceID), samplePlainData1.Data)

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

func getSamplePlainData1() data.PlainData {
	return data.PlainData{
		Metadata: data.ResMetadata{
			ResourceType: data.Plain,
			ResourceID:   "001",
			Hash:         sha256.Sum256([]byte("data1")),
			Size:         uint64(len([]byte("data1"))),
			Extensions:   "{\"name\":\"examplePlainData1\"}",
		},
		Data: []byte("data1"),
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
		Data: []byte("data2"),
	}
}
