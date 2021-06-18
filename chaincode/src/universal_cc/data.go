package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/errorcode"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/data"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/query"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

func (uc *UniversalCC) createPlainData(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// 检查参数数量
	lenArgs := len(args)
	if lenArgs < 1 || lenArgs > 2 {
		return shim.Error("参数数量不正确。应为 1 或 2 个")
	}

	// 解析第 0 个参数为 data.PlainData
	plainData := data.PlainData{}
	if err := json.Unmarshal([]byte(args[0]), &plainData); err != nil {
		return shim.Error(fmt.Sprintf("无法解析参数中的 JSON 对象: %v", err))
	}

	// 若第 1 个参数有指定，则解析为 eventID
	var eventID string
	if lenArgs == 2 {
		eventID = args[1]
	}

	// 检查资源 ID 是否被占用
	resourceID := plainData.Metadata.ResourceID
	dbMetadataKey := getKeyForResMetadata(resourceID)
	dbMetadataVal, err := stub.GetState(dbMetadataKey)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法确定资源 ID 可用性: %v", err))
	}

	if len(dbMetadataVal) != 0 {
		return shim.Error(fmt.Sprintf("资源 ID '%v' 已被占用", resourceID))
	}

	// 将数据本体从 Base64 解码
	dataBytes, err := base64.StdEncoding.DecodeString(plainData.Data)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法解析数据本体: %v", err))
	}

	// 计算哈希和大小并检查是否与用户提供的值相同
	sizeStored := uint64(len(dataBytes))
	if sizeStored != plainData.Metadata.Size {
		return shim.Error(fmt.Sprintf("大小不匹配，应有大小为 %v，实际大小为 %v", plainData.Metadata.Size, sizeStored))
	}

	hashStored := sha256.Sum256(dataBytes)
	hashStoredBase64 := base64.StdEncoding.EncodeToString(hashStored[:])
	if hashStoredBase64 != plainData.Metadata.Hash {
		return shim.Error("哈希不匹配")
	}

	// 获取创建者与时间戳
	creator, err := getPKDERFromStub(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法获取创建者: %v", err))
	}
	creatorAsBase64 := base64.StdEncoding.EncodeToString(creator)

	timestamp, err := getTimeFromStub(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法获得时间戳: %v", err))
	}

	// 准备存储元数据
	metadataStored := data.ResMetadataStored{
		ResourceType: plainData.Metadata.ResourceType,
		ResourceID:   plainData.Metadata.ResourceID,
		Hash:         plainData.Metadata.Hash,
		Size:         plainData.Metadata.Size,
		Extensions:   plainData.Metadata.Extensions,
		Creator:      creatorAsBase64,
		Timestamp:    timestamp,
		HashStored:   hashStoredBase64,
		SizeStored:   sizeStored,
	}

	// 写入数据库
	dbDataKey := getKeyForResData(resourceID)
	if err = stub.PutState(dbDataKey, dataBytes); err != nil {
		return shim.Error(fmt.Sprintf("无法存储资源数据: %v", err))
	}

	metadataStoredBytes, err := json.Marshal(metadataStored)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法序列化元数据: %v", err))
	}
	if err = stub.PutState(dbMetadataKey, metadataStoredBytes); err != nil {
		return shim.Error(fmt.Sprintf("无法存储元数据: %v", err))
	}

	// 建立索引
	// creator~resourceid 绑定创建者与资源 ID
	ckObjectType := "creator~resourceid"
	ckCreatorResourceID, err := stub.CreateCompositeKey(ckObjectType, []string{creatorAsBase64, resourceID})
	if err != nil {
		return shim.Error(fmt.Sprintf("无法创建索引 '%v': %v", ckObjectType, err))
	}
	if err = stub.PutState(ckCreatorResourceID, []byte{0x00}); err != nil {
		return shim.Error(fmt.Sprintf("无法创建索引 '%v': %v", ckObjectType, err))
	}

	// name~resourceid 绑定元数据中 name 字段与资源 ID
	extensionsMap := plainData.Metadata.Extensions
	name, ok := extensionsMap["name"].(string)
	if ok {
		ckObjectType = "name~resourceid"
		ckNameResourceID, err := stub.CreateCompositeKey(ckObjectType, []string{name, resourceID})
		if err != nil {
			return shim.Error(fmt.Sprintf("无法创建索引 '%v': %v", ckObjectType, err))
		}
		if err = stub.PutState(ckNameResourceID, []byte{0x00}); err != nil {
			return shim.Error(fmt.Sprintf("无法创建索引 '%v': %v", ckObjectType, err))
		}
	}

	txID := stub.GetTxID()

	// 发事件
	if eventID != "" {
		if err = stub.SetEvent(eventID, []byte(txID)); err != nil {
			return shim.Error(fmt.Sprintf("无法生成事件 '%v': %v", eventID, err))
		}
	}

	return shim.Success([]byte(txID))
}

func (uc *UniversalCC) createEncryptedData(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// 检查参数数量
	lenArgs := len(args)
	if lenArgs < 1 || lenArgs > 2 {
		return shim.Error("参数数量不正确。应为 1 或 2 个")
	}

	// 解析第 0 个参数为 data.EncryptedData
	encryptedData := data.EncryptedData{}
	if err := json.Unmarshal([]byte(args[0]), &encryptedData); err != nil {
		return shim.Error(fmt.Sprintf("无法解析参数中的 JSON 对象: %v", err))
	}

	// 若第 1 个参数有指定，则解析为 eventID
	var eventID string
	if lenArgs == 2 {
		eventID = args[1]
	}

	// 检查资源 ID 是否被占用
	resourceID := encryptedData.Metadata.ResourceID
	dbMetadataKey := getKeyForResMetadata(resourceID)
	dbMetadataVal, err := stub.GetState(dbMetadataKey)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法确定资源 ID 可用性: %v", err))
	}

	if len(dbMetadataVal) != 0 {
		return shim.Error(fmt.Sprintf("资源 ID '%v' 已被占用", resourceID))
	}

	// 将数据本体从 Base64 解码
	dataBytes, err := base64.StdEncoding.DecodeString(encryptedData.Data)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法解析数据本体: %v", err))
	}

	hashStored := sha256.Sum256(dataBytes)
	hashStoredBase64 := base64.StdEncoding.EncodeToString(hashStored[:])
	sizeStored := len(dataBytes)

	// 获取创建者与时间戳
	creator, err := getPKDERFromStub(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法获取创建者: %v", err))
	}
	creatorAsBase64 := base64.StdEncoding.EncodeToString(creator)

	timestamp, err := getTimeFromStub(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法获得时间戳: %v", err))
	}

	// 准备存储元数据
	metadataStored := data.ResMetadataStored{
		ResourceType: encryptedData.Metadata.ResourceType,
		ResourceID:   encryptedData.Metadata.ResourceID,
		Hash:         encryptedData.Metadata.Hash,
		Size:         encryptedData.Metadata.Size,
		Extensions:   encryptedData.Metadata.Extensions,
		Creator:      creatorAsBase64,
		Timestamp:    timestamp,
		HashStored:   hashStoredBase64,
		SizeStored:   uint64(sizeStored),
	}

	// 写入数据库
	dbDataKey := getKeyForResData(resourceID)
	if err = stub.PutState(dbDataKey, dataBytes); err != nil {
		return shim.Error(fmt.Sprintf("无法存储资源数据: %v", err))
	}

	dbKeyKey := getKeyForResKey(resourceID)
	keyDecoded, err := base64.StdEncoding.DecodeString(encryptedData.Key)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法解析密钥: %v", err))
	}

	if err = stub.PutState(dbKeyKey, keyDecoded); err != nil {
		return shim.Error(fmt.Sprintf("无法存储密钥: %v", err))
	}

	dbPolicyKey := getKeyForResPolicy(resourceID)
	if err = stub.PutState(dbPolicyKey, []byte(encryptedData.Policy)); err != nil {
		return shim.Error(fmt.Sprintf("无法存储策略: %v", err))
	}

	metadataStoredBytes, err := json.Marshal(metadataStored)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法序列化元数据: %v", err))
	}
	if err = stub.PutState(dbMetadataKey, metadataStoredBytes); err != nil {
		return shim.Error(fmt.Sprintf("无法存储元数据: %v", err))
	}

	// 建立索引
	// creator~resourceid 绑定创建者与资源 ID
	ckObjectType := "creator~resourceid"
	ckCreatorResourceID, err := stub.CreateCompositeKey(ckObjectType, []string{creatorAsBase64, resourceID})
	if err != nil {
		return shim.Error(fmt.Sprintf("无法创建索引 '%v': %v", ckObjectType, err))
	}
	if err = stub.PutState(ckCreatorResourceID, []byte{0x00}); err != nil {
		return shim.Error(fmt.Sprintf("无法创建索引 '%v': %v", ckObjectType, err))
	}

	// name~resourceid 绑定元数据中 name 字段与资源 ID（仅当 name 字段可用时绑定）
	extensionsMap := encryptedData.Metadata.Extensions
	name, ok := extensionsMap["name"].(string)
	if ok {
		ckObjectType = "name~resourceid"
		ckNameResourceID, err := stub.CreateCompositeKey(ckObjectType, []string{name, resourceID})
		if err != nil {
			return shim.Error(fmt.Sprintf("无法创建索引 '%v': %v", ckObjectType, err))
		}
		if err = stub.PutState(ckNameResourceID, []byte{0x00}); err != nil {
			return shim.Error(fmt.Sprintf("无法创建索引 '%v': %v", ckObjectType, err))
		}
	}

	txID := stub.GetTxID()

	// 发事件
	if eventID != "" {
		if err = stub.SetEvent(eventID, []byte(txID)); err != nil {
			return shim.Error(fmt.Sprintf("无法生成事件 '%v': %v", eventID, err))
		}
	}

	return shim.Success([]byte(txID))
}

func (uc *UniversalCC) createOffchainData(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// 检查参数数量
	lenArgs := len(args)
	if lenArgs < 1 || lenArgs > 2 {
		return shim.Error("参数数量不正确。应为 1 或 2 个")
	}

	// 解析第 0 个参数为 data.PlainData
	offchainData := data.OffchainData{}
	if err := json.Unmarshal([]byte(args[0]), &offchainData); err != nil {
		return shim.Error(fmt.Sprintf("无法解析参数中的 JSON 对象: %v", err))
	}

	// 若第 1 个参数有指定，则解析为 eventID
	var eventID string
	if lenArgs == 2 {
		eventID = args[1]
	}

	// 检查资源 ID 是否被占用
	resourceID := offchainData.Metadata.ResourceID
	dbMetadataKey := getKeyForResMetadata(resourceID)
	dbMetadataVal, err := stub.GetState(dbMetadataKey)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法确定资源 ID 可用性: %v", err))
	}

	if len(dbMetadataVal) != 0 {
		return shim.Error(fmt.Sprintf("资源 ID '%v' 已被占用", resourceID))
	}

	// CID 不可为空
	if offchainData.CID == "" {
		return shim.Error("资源 CID 不可为空")
	}

	// 计算存储的哈希与大小
	cidBytes := []byte(offchainData.CID)
	hashStored := sha256.Sum256(cidBytes)
	hashStoredBase64 := base64.StdEncoding.EncodeToString(hashStored[:])
	sizeStored := len(cidBytes)

	// 获取创建者与时间戳
	creator, err := getPKDERFromStub(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法获取创建者: %v", err))
	}
	creatorAsBase64 := base64.StdEncoding.EncodeToString(creator)

	timestamp, err := getTimeFromStub(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法获得时间戳: %v", err))
	}

	// 准备存储元数据
	metadataStored := data.ResMetadataStored{
		ResourceType: offchainData.Metadata.ResourceType,
		ResourceID:   offchainData.Metadata.ResourceID,
		Hash:         offchainData.Metadata.Hash,
		Size:         offchainData.Metadata.Size,
		Extensions:   offchainData.Metadata.Extensions,
		Creator:      creatorAsBase64,
		Timestamp:    timestamp,
		HashStored:   hashStoredBase64,
		SizeStored:   uint64(sizeStored),
	}

	// 写入数据库
	dbDataKey := getKeyForResData(resourceID)
	if err = stub.PutState(dbDataKey, cidBytes); err != nil {
		return shim.Error(fmt.Sprintf("无法存储资源数据: %v", err))
	}

	dbKeykey := getKeyForResKey(resourceID)
	keyDecoded, err := base64.StdEncoding.DecodeString(offchainData.Key)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法存储密钥: %v", err))
	}

	if err = stub.PutState(dbKeykey, keyDecoded); err != nil {
		return shim.Error(fmt.Sprintf("无法存储密钥: %v", err))
	}

	dbPolicyKey := getKeyForResPolicy(resourceID)
	if err = stub.PutState(dbPolicyKey, []byte(offchainData.Policy)); err != nil {
		return shim.Error(fmt.Sprintf("无法存储策略: %v", err))
	}

	metadataStoredBytes, err := json.Marshal(metadataStored)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法序列化元数据: %v", err))
	}
	if err = stub.PutState(dbMetadataKey, metadataStoredBytes); err != nil {
		return shim.Error(fmt.Sprintf("无法存储元数据: %v", err))
	}

	// 建立索引
	// creator~resourceid 绑定创建者与资源 ID
	ckObjectType := "creator~resourceid"
	ckCreatorResourceID, err := stub.CreateCompositeKey(ckObjectType, []string{creatorAsBase64, resourceID})
	if err != nil {
		return shim.Error(fmt.Sprintf("无法创建索引 '%v': %v", ckObjectType, err))
	}
	if err = stub.PutState(ckCreatorResourceID, []byte{0x00}); err != nil {
		return shim.Error(fmt.Sprintf("无法创建索引 '%v': %v", ckObjectType, err))
	}

	// name~resourceid 绑定元数据中 name 字段与资源 ID（仅当 name 字段可用时绑定）
	extensionsMap := offchainData.Metadata.Extensions
	name, ok := extensionsMap["name"].(string)
	if ok {
		ckObjectType = "name~resourceid"
		ckNameResourceID, err := stub.CreateCompositeKey(ckObjectType, []string{name, resourceID})
		if err != nil {
			return shim.Error(fmt.Sprintf("无法创建索引 '%v': %v", ckObjectType, err))
		}
		if err = stub.PutState(ckNameResourceID, []byte{0x00}); err != nil {
			return shim.Error(fmt.Sprintf("无法创建索引 '%v': %v", ckObjectType, err))
		}
	}

	txID := stub.GetTxID()

	// 发事件
	if eventID != "" {
		if err = stub.SetEvent(eventID, []byte(txID)); err != nil {
			return shim.Error(fmt.Sprintf("无法生成事件 '%v': %v", eventID, err))
		}
	}

	return shim.Success([]byte(txID))
}

func (uc *UniversalCC) getMetadata(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// 检查参数数量
	lenArgs := len(args)
	if lenArgs != 1 {
		return shim.Error("参数数量不正确。应为 1 个")
	}

	// 解析第一个参数为 resourceID
	resourceID := args[0]

	// 读 metadata 并返回，若未找到则返回 codeNotFound
	dbKey := getKeyForResMetadata(resourceID)
	metadataBytes, err := stub.GetState(dbKey)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法读取元数据: %v", err))
	}

	if len(metadataBytes) == 0 {
		return shim.Error(errorcode.CodeNotFound)
	}

	return shim.Success(metadataBytes)
}

func (uc *UniversalCC) getData(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	lenArgs := len(args)
	if lenArgs != 1 {
		return shim.Error("参数数量不正确。应为 1 个")
	}

	// 解析第一个参数为 resourceID
	resourceID := args[0]

	// 读 data 并返回，若未找到则返回 codeNotFound
	dbKey := getKeyForResData(resourceID)
	dataBytes, err := stub.GetState(dbKey)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法读取数据: %v", err))
	}

	if len(dataBytes) == 0 {
		return shim.Error(errorcode.CodeNotFound)
	}

	return shim.Success(dataBytes)
}

func (uc *UniversalCC) getKey(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	lenArgs := len(args)
	if lenArgs != 1 {
		return shim.Error("参数数量不正确。应为 1 个")
	}

	// 解析第一个参数为 resourceID
	resourceID := args[0]

	// 读 key 并返回，若未找到则返回 codeNotFound
	dbKey := getKeyForResKey(resourceID)
	dataBytes, err := stub.GetState(dbKey)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法读取密钥: %v", err))
	}

	if len(dataBytes) == 0 {
		return shim.Error(errorcode.CodeNotFound)
	}

	return shim.Success(dataBytes)
}

func (uc *UniversalCC) getPolicy(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	lenArgs := len(args)
	if lenArgs != 1 {
		return shim.Error("参数数量不正确。应为 1 个")
	}

	// 解析第一个参数为 resourceID
	resourceID := args[0]

	// 读 policy 并返回，若未找到则返回 codeNotFound
	dbKey := getKeyForResPolicy(resourceID)
	dataBytes, err := stub.GetState(dbKey)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法读取策略: %v", err))
	}

	if len(dataBytes) == 0 {
		return shim.Error(errorcode.CodeNotFound)
	}

	return shim.Success(dataBytes)
}

// 此段保留作为 composite key 索引方案参考
//func (uc *UniversalCC) linkEntityIDWithDocumentID(stub shim.ChaincodeStubInterface, args []string) peer.Response {
//	// 检查参数数量
//	lenArgs := len(args)
//	if lenArgs != 2 {
//		return shim.Error("参数数量不正确。应为 2 个")
//	}
//
//	// args = [entityID, documentID]
//	entityID := args[0]
//	documentID := args[1]
//
//	// entityid~documentid 绑定实体资产 ID 与数字文档 ID
//	ckObjectType := "entityid~documentid"
//	ckEntityIDDocumentID, err := stub.CreateCompositeKey(ckObjectType, []string{entityID, documentID})
//	if err != nil {
//		return shim.Error(fmt.Sprintf("无法创建索引 '%v': %v", ckObjectType, err))
//	}
//	if err = stub.PutState(ckEntityIDDocumentID, []byte{0x00}); err != nil {
//		return shim.Error(fmt.Sprintf("无法创建索引 '%v': %v", ckObjectType, err))
//	}
//
//	txID := stub.GetTxID()
//	return shim.Success([]byte(txID))
//}

// 此段保留作为 composite key 条件查询示例
//func (uc *UniversalCC) listDocumentIDsByEntityID(stub shim.ChaincodeStubInterface, args []string) peer.Response {
//	// 检查参数数量
//	lenArgs := len(args)
//	if lenArgs != 3 {
//		return shim.Error("参数数量不正确。应为 3 个")
//	}
//
//	// args = [entityID string, pageSize int, bookmark string]
//	entityID := args[0]
//
//	pageSizeStr := args[1]
//	pageSize, err := strconv.Atoi(pageSizeStr)
//	if err != nil {
//		return shim.Error(fmt.Sprintf("无法解析参数 pageSize，值为 %v。应为正整数", pageSizeStr))
//	}
//	if pageSize <= 0 {
//		return shim.Error(fmt.Sprintf("参数 pageSize 值为 %v。应为正整数", pageSizeStr))
//	}
//
//	bookmarkAsBase64 := args[2]
//	bookmarkBytes, err := base64.StdEncoding.DecodeString(bookmarkAsBase64)
//	if err != nil {
//		return shim.Error(fmt.Sprintf("无法解析书签: %v", err))
//	}
//	bookmark := string(bookmarkBytes)
//
//	// 提供 entityid 项以获取迭代器
//	ckObjectType := "entityid~documentid"
//	it, respMetadata, err := stub.GetStateByPartialCompositeKeyWithPagination(ckObjectType, []string{entityID}, int32(pageSize), bookmark)
//	if err != nil {
//		return shim.Error(fmt.Sprintf("无法查询索引 '%v': %v", ckObjectType, err))
//	}
//
//	defer it.Close()
//
//	// 遍历迭代器，从中解出 documentid 项，组为列表
//	documentIDs := []string{}
//	for it.HasNext() {
//		entry, err := it.Next()
//		if err != nil {
//			return shim.Error(fmt.Sprintf("无法查询索引 '%v': %v", ckObjectType, err))
//		}
//
//		_, ckParts, err := stub.SplitCompositeKey(entry.Key)
//		if err != nil {
//			return shim.Error(fmt.Sprintf("无法查询索引 '%v': %v", ckObjectType, err))
//		}
//
//		documentIDs = append(documentIDs, ckParts[1])
//	}
//
//	// 序列化结果列表并返回
//	paginationResult := query.ResourceIDsWithPagination{
//		ResourceIDs: documentIDs,
//		Bookmark:    base64.StdEncoding.EncodeToString([]byte(respMetadata.Bookmark)),
//	}
//	paginationResultAsBytes, err := json.Marshal(paginationResult)
//	if err != nil {
//		return shim.Error(fmt.Sprintf("无法序列化结果列表: %v", err))
//	}
//
//	return shim.Success(paginationResultAsBytes)
//}

func (uc *UniversalCC) listDocumentIDsByCreator(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// 检查参数数量
	lenArgs := len(args)
	if lenArgs != 2 {
		return shim.Error("参数数量不正确。应为 2 个")
	}

	// args = [pageSize int, bookmark string]
	pageSizeStr := args[0]
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法解析参数 pageSize，值为 %v。应为正整数", pageSizeStr))
	}
	if pageSize <= 0 {
		return shim.Error(fmt.Sprintf("参数 pageSize 值为 %v。应为正整数", pageSizeStr))
	}

	bookmarkAsBase64 := args[1]
	bookmarkBytes, err := base64.StdEncoding.DecodeString(bookmarkAsBase64)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法解析书签: %v", err))
	}
	bookmark := string(bookmarkBytes)

	// 获取调用者信息
	creator, err := getPKDERFromStub(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法获取调用者信息: %v", err))
	}

	// 提供 creator 项以获取迭代器
	ckObjectType := "creator~resourceid"
	creatorAsBase64 := base64.StdEncoding.EncodeToString(creator)
	it, respMetadata, err := stub.GetStateByPartialCompositeKeyWithPagination(ckObjectType, []string{creatorAsBase64}, int32(pageSize), bookmark)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法查询索引 '%v': %v", ckObjectType, err))
	}

	defer it.Close()

	// 遍历迭代器，解出 resourceid 项，组成列表
	resourceIDs := []string{}
	for it.HasNext() {
		entry, err := it.Next()
		if err != nil {
			return shim.Error(fmt.Sprintf("无法查询索引 '%v': %v", ckObjectType, err))
		}

		_, ckParts, err := stub.SplitCompositeKey(entry.Key)
		if err != nil {
			return shim.Error(fmt.Sprintf("无法查询索引 '%v': %v", ckObjectType, err))
		}

		resourceIDs = append(resourceIDs, ckParts[1])
	}

	// 序列化结果列表并返回
	paginationResult := query.IDsWithPagination{
		IDs:      resourceIDs,
		Bookmark: base64.StdEncoding.EncodeToString([]byte(respMetadata.Bookmark)),
	}
	paginationResultAsBytes, err := json.Marshal(paginationResult)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法序列化结果列表: %v", err))
	}

	return shim.Success(paginationResultAsBytes)
}

// 此段保留作为 CouchDB 模糊查询参考
//func (uc *UniversalCC) listDocumentIDsByPartialName(stub shim.ChaincodeStubInterface, args []string) peer.Response {
//	// 检查参数数量
//	lenArgs := len(args)
//	if lenArgs != 3 {
//		return shim.Error("参数数量不正确。应为 3 个")
//	}
//
//	// args = [partialName string, pageSize int, bookmark string]
//	partialName := args[0]
//	pageSizeStr := args[1]
//	bookmarkAsBase64 := args[2]
//	bookmarkBytes, err := base64.StdEncoding.DecodeString(bookmarkAsBase64)
//	if err != nil {
//		return shim.Error(fmt.Sprintf("无法解析书签: %v", err))
//	}
//	bookmark := string(bookmarkBytes)
//
//	pageSize, err := strconv.Atoi(pageSizeStr)
//	if err != nil {
//		return shim.Error(fmt.Sprintf("无法解析参数 pageSize，值为 %v。应为正整数", pageSizeStr))
//	}
//	if pageSize <= 0 {
//		return shim.Error(fmt.Sprintf("参数 pageSize 值为 %v。应为正整数", pageSizeStr))
//	}
//
//	// 获取关于 extensions.name 模糊匹配的迭代器
//	queryConditions := map[string]interface{}{
//		"selector": map[string]interface{}{
//			"extensions.name": map[string]interface{}{
//				"$regex": partialName,
//			},
//			"extensions.dataType": "document",
//		},
//	}
//	queryConditionsBytes, err := json.Marshal(queryConditions)
//	if err != nil {
//		return shim.Error(fmt.Sprintf("无法序列化查询条件: %v", err))
//	}
//	fmt.Printf("%v\n", string(queryConditionsBytes))
//
//	it, respMetadata, err := stub.GetQueryResultWithPagination(string(queryConditionsBytes), int32(pageSize), bookmark)
//	if err != nil {
//		return shim.Error(fmt.Sprintf("无法执行关于关键词 '%v' 的富查询: %v", partialName, err))
//	}
//
//	defer it.Close()
//
//	// 遍历迭代器，获取所有的 key 并抽取其中的 resourceID，组成列表
//	resourceIDs := []string{}
//	for it.HasNext() {
//		entry, err := it.Next()
//		if err != nil {
//			return shim.Error(fmt.Sprintf("无法执行关于关键词 '%v' 的富查询: %v", partialName, err))
//		}
//
//		resourceID, err := extractResourceIDFromKeyForResMetadata(entry.Key)
//		if err != nil {
//			return shim.Error(err.Error())
//		}
//
//		resourceIDs = append(resourceIDs, resourceID)
//	}
//
//	// 记录书签位置
//	returnedBookmark := respMetadata.Bookmark
//
//	// 序列化结果并返回
//	result := query.ResourceIDsWithPagination{
//		ResourceIDs: resourceIDs,
//		Bookmark:    base64.StdEncoding.EncodeToString([]byte(returnedBookmark)),
//	}
//	resultAsBytes, err := json.Marshal(result)
//	if err != nil {
//		return shim.Error(fmt.Sprintf("无法序列化结果列表: %v", err))
//	}
//
//	return shim.Success(resultAsBytes)
//}

func (uc *UniversalCC) listResourceIDsByConditions(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// 检查参数数量
	lenArgs := len(args)
	if lenArgs != 3 {
		return shim.Error("参数数量不正确。应为 3 个")
	}

	// args = [queryConditions map[string]interface{}, pageSize int, bookmark string]
	queryConditionsBytes := []byte(args[0])
	queryConditions := make(map[string]interface{})
	err := json.Unmarshal(queryConditionsBytes, &queryConditions)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法解析查询条件: %v", err))
	}
	pageSizeStr := args[1]
	bookmarkAsBase64 := args[2]
	bookmarkBytes, err := base64.StdEncoding.DecodeString(bookmarkAsBase64)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法解析书签: %v", err))
	}
	bookmark := string(bookmarkBytes)

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法解析参数 pageSize，值为 %v。应为正整数", pageSizeStr))
	}
	if pageSize <= 0 {
		return shim.Error(fmt.Sprintf("参数 pageSize 值为 %v。应为正整数", pageSizeStr))
	}

	// 获取匹配查询条件的迭代器
	it, respMetadata, err := stub.GetQueryResultWithPagination(string(queryConditionsBytes), int32(pageSize), bookmark)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法执行条件查询: %v", err))
	}

	defer it.Close()

	// 遍历迭代器，获取所有的 key 并抽取其中的 resourceID，组成列表
	resourceIDs := []string{}
	for it.HasNext() {
		entry, err := it.Next()
		if err != nil {
			return shim.Error(fmt.Sprintf("无法执行条件查询: %v", err))
		}

		resourceID, err := extractResourceIDFromKeyForResMetadata(entry.Key)
		if err != nil {
			return shim.Error(err.Error())
		}

		resourceIDs = append(resourceIDs, resourceID)
	}

	// 记录书签位置
	returnedBookmark := respMetadata.Bookmark

	// 序列化结果并返回
	result := query.IDsWithPagination{
		IDs:      resourceIDs,
		Bookmark: base64.StdEncoding.EncodeToString([]byte(returnedBookmark)),
	}
	resultAsBytes, err := json.Marshal(result)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法序列化结果列表: %v", err))
	}

	return shim.Success(resultAsBytes)
}
