package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/auth"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/data"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

func (uc *UniversalCC) createAuthRequest(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	// 检查参数数量
	lenArgs := len(args)
	if lenArgs < 1 || lenArgs > 2 {
		return shim.Error("参数数量不正确。应为 1 或 2 个")
	}

	// 解析第 0 个参数为 auth.AuthRequest
	request := []byte(args[0])
	var authRequest auth.AuthRequest
	err := json.Unmarshal(request, &authRequest)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法解析参数中的 JSON 对象: %v", err))
	}

	// 若第 1 个参数有指定，则解析为 eventID
	var eventID string
	if lenArgs == 2 {
		eventID = args[1]
	}

	// 检查资源是否存在
	resourceID := authRequest.ResourceID
	key := getKeyForResMetadata(resourceID)
	metadata, err := stub.GetState(key)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法确定资源 ID 可用性: %v", err))
	}
	if metadata == nil {
		return shim.Error("资源 ID 不存在")
	}

	// 获取创建者与时间戳
	creator, err := getPKDERFromStub(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法获取创建者: %v", err))
	}

	timestamp, err := getTimeFromStub(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法获得时间戳: %v", err))
	}

	// 获取 authSessionID
	authSessionID := stub.GetTxID()

	// 构建authRequestStored，并存储上链
	authRequestStored := auth.AuthRequestStored{authSessionID, authRequest.ResourceID, authRequest.Extensions, creator, timestamp}
	authRequestStoredByte, err := json.Marshal(authRequestStored)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法序列化 authRequestStored: %v", err))
	}
	key = getKeyForAuthRequest(authSessionID)
	err = stub.PutState(key, authRequestStoredByte)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法存储 authRequestStored: %v", err))
	}

	// 构建并获取 resMetadataStored，以此得到资源的creator
	var Metadata data.ResMetadataStored
	err = json.Unmarshal(metadata, &Metadata)
	if err != nil {
		return shim.Error(fmt.Sprintf("元数据无法解析成 JSON 对象: %v", err))
	}

	// 建立索引
	// resourcecreator~authsessionid 绑定资源创建者和auth会话ID
	creatorAsBase64 := base64.StdEncoding.EncodeToString(creator)
	indexName := "resourcecreator~authsessionid"
	indexKey, err := stub.CreateCompositeKey(indexName, []string{creatorAsBase64, authSessionID})
	if err != nil {
		return shim.Error(fmt.Sprintf("无法创建索引 '%v': %v", indexName, err))
	}
	value := []byte{0x00}
	err = stub.PutState(indexKey, value)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法创建索引 '%v': %v", indexName, err))
	}

	// 发事件
	if err = stub.SetEvent(eventID, []byte(authSessionID)); err != nil {
		return shim.Error(fmt.Sprintf("无法生成事件 '%v': %v", eventID, err))
	}

	return shim.Success([]byte(authSessionID))
}

func (uc *UniversalCC) createAuthResponse(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// 检查参数数量
	lenArgs := len(args)
	if lenArgs < 1 || lenArgs > 2 {
		return shim.Error("参数数量不正确。应为 1 或 2 个")
	}

	// 解析第0个参数为 auth.AuthResponse
	response := []byte(args[0])
	var authResponse auth.AuthResponse
	err := json.Unmarshal(response, &authResponse)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法解析参数中的 JSON 对象: %v", err))
	}

	// 若第 1 个参数有指定，则解析为 eventID
	var eventID string
	if lenArgs == 2 {
		eventID = args[1]
	}

	// 检查授权会话的请求是否存在
	key := getKeyForAuthRequest(authResponse.AuthSessionID)
	authReq, err := stub.GetState(key)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法确定授权会话是否存在: %v", err))
	}
	if authReq == nil {
		return shim.Error("该授权会话不存在")
	}

	// 检查授权请求是否已经被回复
	key = getKeyForAuthResponse(authResponse.AuthSessionID)
	authResp, err := stub.GetState(key)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法确定该授权会话的批复状态: %v", err))
	}
	if authResp != nil {
		return shim.Error("该授权请求已经被批复")
	}

	// 构建 AuthRequetStored 去得到 资源 ID
	var authRequestStored auth.AuthRequestStored
	err = json.Unmarshal(authReq, &authRequestStored)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法解析 authRequestStored 的 JSON 对象: %v", err))
	}

	// 根据资源id，得到资源的元数据的序列化结果，以此检查资源是否存在
	key = getKeyForResMetadata(authRequestStored.ResourceID)
	metadata, err := stub.GetState(key)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法确定资源 ID 可用性: %v", err))
	}
	if metadata == nil {
		return shim.Error("该资源不存在")
	}

	// 解析resMetadataStored，以此得到资源创建者
	var Metadata data.ResMetadataStored
	err = json.Unmarshal(metadata, &Metadata)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法解析 ResMetadataStored 的 JSON 对象: %v", err))
	}

	// 获取创建者与时间戳
	creator, err := getPKDERFromStub(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法获取创建者: %v", err))
	}

	timestamp, err := getTimeFromStub(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法获得时间戳: %v", err))
	}

	// 检查该交易的创建者是否为资源创建者
	if string(creator) != string(Metadata.Creator) {
		return shim.Error("你不是资源创建者，无权限批复该请求")
	}

	// 构建 AuthResponseStored 并存储上链
	authResponseStored := auth.AuthResponseStored{authResponse.AuthSessionID, authResponse.Result, creator, timestamp}
	data, err := json.Marshal(authResponseStored)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法序列化 AuthResponseStored: %v", err))
	}
	key = getKeyForAuthResponse(authResponse.AuthSessionID)
	err = stub.PutState(key, data)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法存储 AuthResponseStored: %v", err))
	}

	// 删除 resourcecreator~authsessionid 索引
	indexName := "resourcecreator~authsessionid"
	creatorAsBase64 := base64.StdEncoding.EncodeToString(Metadata.Creator)
	indexKey, err := stub.CreateCompositeKey(indexName, []string{creatorAsBase64, authResponse.AuthSessionID})
	err = stub.DelState(indexKey)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法删除索引 '%v': %v", indexName, err))
	}

	// 获取交易ID
	txID := stub.GetTxID()

	// 发事件
	if err = stub.SetEvent(eventID, []byte(txID)); err != nil {
		return shim.Error(fmt.Sprintf("无法生成事件 '%v': %v", eventID, err))
	}
	return shim.Success([]byte(txID))
}

func (uc *UniversalCC) getAuthRequest(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// 检查参数个数
	if len(args) != 1 {
		return shim.Error("参数数量不正确。应为 1 个")
	}

	// 第 0 个参数为授权会话 ID
	authSessionID := args[0]

	// 从链上读取 AuthRequestStored
	key := getKeyForAuthRequest(authSessionID)
	authReq, err := stub.GetState(key)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法确定授权会话的可用性: %v", err))
	}
	if authReq == nil {
		return shim.Error("授权会话不存在")
	}

	return shim.Success(authReq)
}

func (uc *UniversalCC) getAuthResponseHelper(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// 检查参数个数
	if len(args) != 1 {
		return shim.Error("参数数量不正确。应为 1 个")
	}

	// 第 0 个参数为授权会话 ID
	authSessionID := args[0]

	// 从链上读取 AuthResponseStored
	key := getKeyForAuthResponse(authSessionID)
	authResp, err := stub.GetState(key)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法确定授权会话状态: %v", err))
	}
	if authResp == nil {
		return shim.Error("该授权请求未被批复")
	}

	return shim.Success(authResp)
}

func (uc *UniversalCC) listPendingAuthSessionIDsByResourceCreator(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// 检查参数数量
	lenArgs := len(args)
	if lenArgs != 2 {
		return shim.Error("参数数量不正确。应为 2 个")
	}

	// args = [pageSize int32, bookmark string]
	pageSizeStr := args[0]
	bookmark := args[1]

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法解析参数 pageSize，值为 %v。应为正整数", pageSizeStr))
	}
	if pageSize <= 0 {
		return shim.Error(fmt.Sprintf("参数 pageSize 值为 %v。应为正整数", pageSizeStr))
	}

	// 获取调用者信息
	creator, err := getPKDERFromStub(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法获取调用者信息: %v", err))
	}

	// 提供 creator 项以获取迭代器
	indexName := "resourcecreator~authsessionid"
	it, _, err := stub.GetStateByPartialCompositeKeyWithPagination(indexName, []string{string(creator)}, int32(pageSize), bookmark)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法查询索引 '%v': %v", indexName, err))
	}

	// 遍历迭代器，解出 authsessionid 项，组为列表
	authSessionIDs := []string{}
	for it.HasNext() {
		entry, err := it.Next()
		if err != nil {
			return shim.Error(fmt.Sprintf("无法查询索引 '%v': %v", indexName, err))
		}

		_, ckParts, err := stub.SplitCompositeKey(entry.Key)
		if err != nil {
			return shim.Error(fmt.Sprintf("无法查询索引 '%v': %v", indexName, err))
		}

		authSessionIDs = append(authSessionIDs, ckParts[1])
	}

	// 序列化结果
	authSessionIDsAsBytes, err := json.Marshal(authSessionIDs)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法序列化结果列表: %v", err))
	}

	return shim.Success(authSessionIDsAsBytes)
}
