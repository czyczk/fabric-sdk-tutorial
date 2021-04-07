package main

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/errorcode"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/auth"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/data"
	"github.com/google/uuid"
	"github.com/hyperledger/fabric-chaincode-go/shimtest"
	"github.com/hyperledger/fabric-protos-go/peer"
)

func createEncryptedDataHelper(stub *shimtest.MockStub, encryptedData data.EncryptedData) {
	// 创建加密数据
	targetFunction := "createEncryptedData"

	// Prepare the encryptedData
	dataBytes, _ := json.Marshal(encryptedData)

	// Invoke with sampleEncryptedData
	stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes, []byte("1")})
}

func createAuthRequestHelper(stub *shimtest.MockStub, authRequest auth.AuthRequest) peer.Response {
	// 创建授权请求
	targetFunction := "createAuthRequest"

	// Prepare authrequest
	dataBytes, _ := json.Marshal(authRequest)

	// Invoke
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes, []byte("1")})
	return resp
}

func createAuthResponseHelper(stub *shimtest.MockStub, authResponse auth.AuthResponse) peer.Response {
	// 创建授权批复
	targetFunction := "createAuthResponse"

	// Prepare authresponse
	dataBytes, _ := json.Marshal(authResponse)

	// invoke
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes, []byte("1")})
	return resp
}

func TestCreateAuthRequestWithEncryptedData(t *testing.T) {
	// 初始化
	stub := createMockStub(t, "TestCreateAuthRequestWithEncryptedData")
	_ = initChaincode(stub, [][]byte{})

	// 创建加密文档
	encryptedData := getSampleEncryptedData1()
	createEncryptedDataHelper(stub, encryptedData)

	// 创建授权请求
	authRequest := getSampleAuthRequest1(encryptedData.Metadata.ResourceID)
	resp := createAuthRequestHelper(stub, authRequest)
	expectResponseStatusOK(t, &resp)

	// 验证
	authSessionID := string(resp.Payload)
	authRequestStoredBytes := stub.State[getKeyForAuthRequest(authSessionID)]
	authRequestStored := auth.AuthRequestStored{}
	if err := json.Unmarshal(authRequestStoredBytes, &authRequestStored); err != nil {
		testLogger.Infof("Cannot read stored authRequest: %v\n", err)
		t.FailNow()
	}

	expectEqual(t, authSessionID, authRequestStored.AuthSessionID)
	expectEqual(t, authRequest.ResourceID, authRequestStored.ResourceID)
	expectEqual(t, authRequest.Extensions, authRequestStored.Extensions)
}

func TestCreateAuthRequestWithOffchainData(t *testing.T) {
	// 初始化
	stub := createMockStub(t, "TestCreateAuthRequestWithOffchainData")
	_ = initChaincode(stub, [][]byte{})

	// 创建链下数据
	targetFunction := "createOffchainData"
	sampleOffchainData1 := getSampleOffchainData1()
	dataBytes, _ := json.Marshal(sampleOffchainData1)
	stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes})

	// 创建授权请求
	authRequest := getSampleAuthRequest1(sampleOffchainData1.Metadata.ResourceID)
	resp := createAuthRequestHelper(stub, authRequest)
	expectResponseStatusOK(t, &resp)

	// 验证
	authSessionID := string(resp.Payload)
	authRequestStoredBytes := stub.State[getKeyForAuthRequest(authSessionID)]
	authRequestStored := auth.AuthRequestStored{}
	if err := json.Unmarshal(authRequestStoredBytes, &authRequestStored); err != nil {
		testLogger.Infof("Cannot read stored authRequest: %v\n", err)
		t.FailNow()
	}

	expectEqual(t, authSessionID, authRequestStored.AuthSessionID)
	expectEqual(t, authRequest.ResourceID, authRequestStored.ResourceID)
	expectEqual(t, authRequest.Extensions, authRequestStored.Extensions)
}

func TestCreateAuthRequestWithPlainData(t *testing.T) {
	// 初始化
	stub := createMockStub(t, "TestCreateAuthRequestWithPlainData")
	_ = initChaincode(stub, [][]byte{})

	// 创建明文文档
	targetFunction := "createPlainData"
	samplePlainData1 := getSamplePlainData1()
	dataBytes, _ := json.Marshal(samplePlainData1)
	stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes})

	// 创建授权请求
	authRequest := getSampleAuthRequest1(samplePlainData1.Metadata.ResourceID)
	resp := createAuthRequestHelper(stub, authRequest)
	expectResponseStatusERROR(t, &resp)
}

func TestCreateAuthRequestWithRegulatorEncryptedData(t *testing.T) {

}
func TestCreateAuthRequestWithExcessiveParameters(t *testing.T) {
	// 初始化
	stub := createMockStub(t, "TestCreateAuthRequestWithExcessiveParameters")
	_ = initChaincode(stub, [][]byte{})

	// 创建加密文档
	encryptedData := getSampleEncryptedData1()
	createEncryptedDataHelper(stub, encryptedData)

	// 错误的参数创建授权请求
	targetFunction := "createAuthRequest"

	// Prepare authrequest
	sampleAuthRequest := getSampleAuthRequest1(encryptedData.Metadata.ResourceID)
	dataBytes, _ := json.Marshal(sampleAuthRequest)

	// invoke
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes, []byte("1"), []byte("0")})
	expectResponseStatusERROR(t, &resp)
}

func TestCreateAuthRequestWithNonExistentResourceID(t *testing.T) {
	// 初始化
	stub := createMockStub(t, "TestCreateAuthRequestWithNonExistentResourceID")
	_ = initChaincode(stub, [][]byte{})

	// 创建授权请求
	authRequest := getSampleAuthRequest1("NON_EXISTENT_RESOURCE_ID")
	resp := createAuthRequestHelper(stub, authRequest)
	expectResponseStatusERROR(t, &resp)
}

func TestCreateAuthRequestIndexStatus(t *testing.T) {
	// 初始化
	stub := createMockStub(t, "TestCreateAuthRequestIndexStatus")
	_ = initChaincode(stub, [][]byte{})

	// 创建两份加密数据
	sampleEncryptedData1 := getSampleEncryptedData1()
	sampleEncryptedData2 := getSampleEncryptedData2()

	createEncryptedDataHelper(stub, sampleEncryptedData1)
	createEncryptedDataHelper(stub, sampleEncryptedData2)

	// 创建授权请求,改用 user2 的证书
	setMockStubCreator(t, stub, "Org1MSP", []byte(exampleCertUser2))

	sampleAuthRequest1 := getSampleAuthRequest1(sampleEncryptedData1.Metadata.ResourceID)
	sampleAuthRequest2 := getSampleAuthRequest2(sampleEncryptedData2.Metadata.ResourceID)
	createAuthRequestHelper(stub, sampleAuthRequest1)
	createAuthRequestHelper(stub, sampleAuthRequest2)

	// 验证，resourcecreator 为 user1，有两条索引记录。
	creator, err := getPKDERFromCertString(exampleCertUser1)
	if err != nil {
		testLogger.Infof("Error parsing certificate: %v\n", err)
		t.FailNow()
	}
	creatorAsBase64 := base64.StdEncoding.EncodeToString(creator)

	it, err := stub.GetStateByPartialCompositeKey("resourcecreator~authsessionid", []string{creatorAsBase64})
	if err != nil {
		testLogger.Infof("Cannot query index 'resourcecreator~authsessionid': %v\n", err)
		t.FailNow()
	}

	// Iterate the composite key value entries and expect 2 distinct authSession IDs
	authSessionIDSet := make(map[string]bool)
	for it.HasNext() {
		entry, _ := it.Next()
		_, ckParts, _ := stub.SplitCompositeKey(entry.Key)
		authSessionIDSet[ckParts[1]] = true
	}
	expectEqual(t, 2, len(authSessionIDSet))
}

func TestCreateAuthResponseWithNormalProcess(t *testing.T) {
	// 初始化
	stub := createMockStub(t, "TestCreateAuthResponseWithNormalProcess")
	_ = initChaincode(stub, [][]byte{})

	// user1 创建加密数据
	sampleEncryptedData := getSampleEncryptedData1()
	createEncryptedDataHelper(stub, sampleEncryptedData)

	// user2 创建授权请求
	setMockStubCreator(t, stub, "Org1MSP", []byte(exampleCertUser2))
	sampleAuthRequest := getSampleAuthRequest1(sampleEncryptedData.Metadata.ResourceID)
	resp := createAuthRequestHelper(stub, sampleAuthRequest)

	// user1 创建授权批复
	setMockStubCreator(t, stub, "Org1MSP", []byte(exampleCertUser1))
	sampleAuthResponse := getSampleAuthResponse1()
	authSessionID := string(resp.Payload)
	sampleAuthResponse.AuthSessionID = authSessionID
	resp = createAuthResponseHelper(stub, sampleAuthResponse)
	expectResponseStatusOK(t, &resp)

	// 验证 state
	authResponseStoredBytes := stub.State[getKeyForAuthResponse(authSessionID)]
	authResponseStored := auth.AuthResponseStored{}
	if err := json.Unmarshal(authResponseStoredBytes, &authResponseStored); err != nil {
		testLogger.Infof("Cannot read stored authResponse: %v\n", err)
		t.FailNow()
	}

	expectEqual(t, authSessionID, authResponseStored.AuthSessionID)
	expectEqual(t, sampleAuthResponse.Result, authResponseStored.Result)
}

func TestCreateAuthResponseWithExcessiveParameters(t *testing.T) {
	// 初始化
	stub := createMockStub(t, "TestCreateAuthResponseWithExcessiveParameters")
	_ = initChaincode(stub, [][]byte{})

	// user1 创建加密数据
	sampleEncryptedData := getSampleEncryptedData1()
	createEncryptedDataHelper(stub, sampleEncryptedData)

	// user2 创建授权请求
	setMockStubCreator(t, stub, "Org1MSP", []byte(exampleCertUser2))
	sampleAuthRequest := getSampleAuthRequest1(sampleEncryptedData.Metadata.ResourceID)
	resp := createAuthRequestHelper(stub, sampleAuthRequest)

	// user1 创建授权批复
	setMockStubCreator(t, stub, "Org1MSP", []byte(exampleCertUser1))
	targetFunction := "createAuthResponse"
	sampleAuthResponse := getSampleAuthResponse1()
	authSessionID := string(resp.Payload)
	sampleAuthResponse.AuthSessionID = authSessionID
	dataBytes, _ := json.Marshal(sampleAuthResponse)
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), dataBytes, []byte("1"), []byte("2")})
	expectResponseStatusERROR(t, &resp)
}

func TestCreateAuthResponseTwice(t *testing.T) {
	// 初始化
	stub := createMockStub(t, "TestCreateAuthResponseTwice")
	_ = initChaincode(stub, [][]byte{})

	// user1 创建加密数据
	sampleEncryptedData := getSampleEncryptedData1()
	createEncryptedDataHelper(stub, sampleEncryptedData)

	// user2 创建授权请求
	setMockStubCreator(t, stub, "Org1MSP", []byte(exampleCertUser2))
	sampleAuthRequest := getSampleAuthRequest1(sampleEncryptedData.Metadata.ResourceID)
	resp := createAuthRequestHelper(stub, sampleAuthRequest)

	// user1 创建授权批复
	setMockStubCreator(t, stub, "Org1MSP", []byte(exampleCertUser1))
	sampleAuthResponse := getSampleAuthResponse1()
	authSessionID := string(resp.Payload)
	sampleAuthResponse.AuthSessionID = authSessionID
	createAuthResponseHelper(stub, sampleAuthResponse)

	// user1 再次创建授权批复
	resp = createAuthResponseHelper(stub, sampleAuthResponse)
	expectResponseStatusERROR(t, &resp)
}

func TestCreateAuthResponseWithNonExistentSessionID(t *testing.T) {

	// 初始化
	stub := createMockStub(t, "TestCreateAuthResponseWithNonExistentSessionID")
	_ = initChaincode(stub, [][]byte{})

	// 直接创建授权批复
	sampleAuthResponse := getSampleAuthResponse1()
	resp := createAuthResponseHelper(stub, sampleAuthResponse)
	expectResponseStatusERROR(t, &resp)
}

func TestCreateAuthResponseWithOthersResource(t *testing.T) {
	// 初始化
	stub := createMockStub(t, "TestCreateAuthResponseWithOthersResource")
	_ = initChaincode(stub, [][]byte{})

	// user1 创建加密数据
	sampleEncryptedData := getSampleEncryptedData1()
	createEncryptedDataHelper(stub, sampleEncryptedData)

	// user2 创建授权请求
	setMockStubCreator(t, stub, "Org1MSP", []byte(exampleCertUser2))
	sampleAuthRequest := getSampleAuthRequest1(sampleEncryptedData.Metadata.ResourceID)
	resp := createAuthRequestHelper(stub, sampleAuthRequest)

	// user2 创建授权批复
	sampleAuthResponse := getSampleAuthResponse1()
	authSessionID := string(resp.Payload)
	sampleAuthResponse.AuthSessionID = authSessionID
	resp = createAuthResponseHelper(stub, sampleAuthResponse)

	// 状态应为 ERROR 且错误内容为 `codeForbidden`
	expectResponseStatusERROR(t, &resp)
	expectStringEndsWith(t, errorcode.CodeForbidden, resp.Message)
}

func TestCreateAuthResponseIndexStatus(t *testing.T) {
	// 初始化
	stub := createMockStub(t, "TestCreateAuthResponseIndexStatus")
	_ = initChaincode(stub, [][]byte{})

	// user1 创建两份加密数据
	sampleEncryptedData1 := getSampleEncryptedData1()
	sampleEncryptedData2 := getSampleEncryptedData2()

	createEncryptedDataHelper(stub, sampleEncryptedData1)
	createEncryptedDataHelper(stub, sampleEncryptedData2)

	// user2 创建两份授权请求
	setMockStubCreator(t, stub, "Org1MSP", []byte(exampleCertUser2))

	sampleAuthRequest1 := getSampleAuthRequest1(sampleEncryptedData1.Metadata.ResourceID)
	sampleAuthRequest2 := getSampleAuthRequest2(sampleEncryptedData2.Metadata.ResourceID)

	createAuthRequestHelper(stub, sampleAuthRequest1)
	resp := createAuthRequestHelper(stub, sampleAuthRequest2)

	// user1 批复上述授权2
	setMockStubCreator(t, stub, "Org1MSP", []byte(exampleCertUser1))
	sampleAuthResponse := getSampleAuthResponse1()
	authSessionID := string(resp.Payload)
	sampleAuthResponse.AuthSessionID = authSessionID
	createAuthResponseHelper(stub, sampleAuthResponse)

	// 验证，resourcecreator 为 user1 ，有 1 条索引记录
	creator, err := getPKDERFromCertString(exampleCertUser1)
	if err != nil {
		testLogger.Infof("Error parsing certificate: %v\n", err)
		t.FailNow()
	}
	creatorAsBase64 := base64.StdEncoding.EncodeToString(creator)

	it, err := stub.GetStateByPartialCompositeKey("resourcecreator~authsessionid", []string{creatorAsBase64})
	if err != nil {
		testLogger.Infof("Cannot query index 'resourcecreator~authsessionid': %v\n", err)
		t.FailNow()
	}

	// Iterate the composite key value entries and expect 1 distinct authSession IDs
	authSessionIDSet := make(map[string]bool)
	for it.HasNext() {
		entry, _ := it.Next()
		_, ckParts, _ := stub.SplitCompositeKey(entry.Key)
		authSessionIDSet[ckParts[1]] = true
	}

	expectEqual(t, 1, len(authSessionIDSet))
}

func TestGetAuthRequestWithNormalParameters(t *testing.T) {
	// 初始化
	stub := createMockStub(t, "TestGetAuthRequestWithNormalParameters")
	_ = initChaincode(stub, [][]byte{})

	// user1 创建加密数据
	sampleEncryptedData := getSampleEncryptedData1()
	createEncryptedDataHelper(stub, sampleEncryptedData)

	// user2 创建授权请求
	setMockStubCreator(t, stub, "Org1MSP", []byte(exampleCertUser2))
	sampleAuthRequest := getSampleAuthRequest1(sampleEncryptedData.Metadata.ResourceID)
	resp := createAuthRequestHelper(stub, sampleAuthRequest)

	// GetAuthRequest
	targetFunction := "getAuthRequest"
	authSessionID := string(resp.Payload)
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), []byte(authSessionID)})
	expectResponseStatusOK(t, &resp)

	// get authRequestStored
	authRequestStored := auth.AuthRequestStored{}
	if err := json.Unmarshal(resp.Payload, &authRequestStored); err != nil {
		testLogger.Infof("Cannot read stored authRequest: %v\n", err)
		t.FailNow()
	}

	expectEqual(t, authSessionID, authRequestStored.AuthSessionID)
	expectEqual(t, sampleAuthRequest.ResourceID, authRequestStored.ResourceID)
	expectEqual(t, sampleAuthRequest.Extensions, authRequestStored.Extensions)
}

func TestGetAuthRequestWithExcessiveParameters(t *testing.T) {
	// 初始化
	stub := createMockStub(t, "TestGetAuthRequestWithExcessiveParameters")
	_ = initChaincode(stub, [][]byte{})

	// user1 创建加密数据
	sampleEncryptedData := getSampleEncryptedData1()
	createEncryptedDataHelper(stub, sampleEncryptedData)

	// user2 创建授权请求
	setMockStubCreator(t, stub, "Org1MSP", []byte(exampleCertUser2))
	sampleAuthRequest := getSampleAuthRequest1(sampleEncryptedData.Metadata.ResourceID)
	resp := createAuthRequestHelper(stub, sampleAuthRequest)

	// GetAuthRequest
	targetFunction := "getAuthRequest"
	authSessionID := string(resp.Payload)
	resp = stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), []byte(authSessionID), []byte("1")})
	expectResponseStatusERROR(t, &resp)
}

func TestGetAuthRequestWithNonExistentSessionID(t *testing.T) {
	// 初始化
	stub := createMockStub(t, "TestGetAuthRequestWithNonExistentSessionID")
	_ = initChaincode(stub, [][]byte{})

	// GetAuthRequest
	targetFunction := "getAuthRequest"
	authSessionID := "01"
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), []byte(authSessionID)})

	// Expect the response status to be ERROR and containing error message of `errorcode.CodeNotFound`.
	expectResponseStatusERROR(t, &resp)
	expectStringEndsWith(t, errorcode.CodeNotFound, resp.Message)
}

func getSampleAuthRequest1(resourceID string) auth.AuthRequest {
	return auth.AuthRequest{
		ResourceID: resourceID,
		Extensions: "{\"name\":\"exampleAuthRequest1\"}",
	}
}

func getSampleAuthRequest2(resourceID string) auth.AuthRequest {
	return auth.AuthRequest{
		ResourceID: resourceID,
		Extensions: "{\"name\":\"exampleAuthRequest2\"}",
	}
}

func getSampleAuthResponse1() auth.AuthResponse {
	return auth.AuthResponse{
		AuthSessionID: "1",
		Result:        true,
	}
}

func getSampleAuthResponse2() auth.AuthResponse {
	return auth.AuthResponse{
		AuthSessionID: "2",
		Result:        false,
	}
}
