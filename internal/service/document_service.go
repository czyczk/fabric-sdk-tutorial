package service

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"runtime"
	"strconv"
	"strings"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/db"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/global"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/models/common"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/utils/cipherutils"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/errorcode"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/data"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/keyswitch"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/query"
	"github.com/XiaoYao-austin/ppks"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/tjfoc/gmsm/sm2"
)

// DocumentService 用于管理数字文档。
type DocumentService struct {
	ServiceInfo      *Info
	KeySwitchService KeySwitchServiceInterface
}

// 用于放置在元数据的 extensions.dataType 中的值
const documentDataType = "document"

// CreateDocument 创建数字文档。
//
// 参数：
//   数字文档
//
// 返回：
//   交易 ID
func (s *DocumentService) CreateDocument(document *common.Document) (string, error) {
	if document == nil {
		return "", fmt.Errorf("文档对象不能为 nil")
	}

	// 检查 ID 是否为空。若上层忽略此项检查此项为空，将可能对链码层造成混乱。
	if strings.TrimSpace(document.ID) == "" {
		return "", fmt.Errorf("文档 ID 不能为空")
	}

	// 将 struct 转成 map 后再序列化，以使得键按字典序排序。
	// 因为在 CouchDB 中存储 JSON 序列化后的对象时，不会保存键的顺序，取出时键将以字典序排序。
	// 若此时直接按 struct 序列化，保持原始键顺序的话，此时计算的哈希将与取出后的哈希不同。
	documentAsMap := make(map[string]interface{})
	err := mapstructure.Decode(document, &documentAsMap)
	if err != nil {
		return "", errors.Wrap(err, "无法序列化文档")
	}

	documentBytes, err := json.Marshal(documentAsMap)
	if err != nil {
		return "", errors.Wrap(err, "无法序列化文档")
	}

	// 计算哈希，获取大小并准备可公开的扩展字段
	hash := sha256.Sum256(documentBytes)
	hashBase64 := base64.StdEncoding.EncodeToString(hash[:])
	size := len(documentBytes)
	extensions := deriveExtensionsMapFromDocumentProperties(&document.DocumentProperties, nil)

	metadata := data.ResMetadata{
		ResourceType: data.Plain,
		ResourceID:   document.ID,
		Hash:         hashBase64,
		Size:         uint64(size),
		Extensions:   extensions,
	}

	// 组装要传入链码的参数，其中数据本体转换为 Base64 编码
	plainData := data.PlainData{
		Metadata: metadata,
		Data:     base64.StdEncoding.EncodeToString(documentBytes),
	}
	plainDataBytes, err := json.Marshal(plainData)
	if err != nil {
		return "", errors.Wrap(err, "无法序列化链码参数")
	}

	chaincodeFcn := "createPlainData"
	channelReq := channel.Request{
		ChaincodeID: s.ServiceInfo.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{plainDataBytes},
	}

	resp, err := s.ServiceInfo.ChannelClient.Execute(channelReq)
	if err != nil {
		return "", GetClassifiedError(chaincodeFcn, err)
	} else {
		return string(resp.TransactionID), nil
	}
}

// CreateEncryptedDocument 创建加密数字文档。
//
// 参数：
//   数字文档
//   对称密钥（SM2 曲线上的点）
//   访问策略
//
// 返回：
//   交易 ID
func (s *DocumentService) CreateEncryptedDocument(document *common.Document, key *ppks.CurvePoint, policy string) (string, error) {
	if document == nil {
		return "", fmt.Errorf("文档对象不能为 nil")
	}
	// 检查 ID 是否为空。若上层忽略此项检查此项为空，将可能对链码层造成混乱。
	if strings.TrimSpace(document.ID) == "" {
		return "", fmt.Errorf("文档 ID 不能为空")
	}

	documentBytes, err := json.Marshal(document)
	if err != nil {
		return "", errors.Wrap(err, "无法序列化文档")
	}
	documentPropertiesBytes, err := json.Marshal(document.DocumentProperties)
	if err != nil {
		return "", errors.Wrap(err, "无法序列化文档属性")
	}

	document = nil

	// 用 key 加密 documentBytes 和 documentPropertiesBytes
	// 使用由 key 导出的 256 位信息来创建 AES256 block
	encryptedDocumentBytes, err := encryptDataWithTimer(documentBytes, key, "无法加密文档", "加密文档")
	if err != nil {
		return "", err
	}

	encryptedDocumentPropertiesBytes, err := encryptDataWithTimer(documentPropertiesBytes, key, "无法加密文档属性", "加密文档属性")
	if err != nil {
		return "", err
	}

	// 获取集合公钥（当前实现为 SM2 公钥）
	collPubKey, err := s.KeySwitchService.GetCollectiveAuthorityPublicKey()
	if err != nil {
		return "", err
	}
	collPubKeyInSM2 := collPubKey.(*sm2.PublicKey)

	// 用集合公钥加密 key
	encryptedKey, err := ppks.PointEncrypt(collPubKeyInSM2, key)
	if err != nil {
		return "", errors.Wrap(err, "无法加密对称密钥")
	}
	// 序列化加密后的 key
	encryptedKeyBytes := cipherutils.SerializeCipherText(encryptedKey)

	// 计算原始内容的哈希，获取大小并准备扩展字段
	hash := sha256.Sum256(documentBytes)
	hashBase64 := base64.StdEncoding.EncodeToString(hash[:])
	size := len(documentBytes)
	extensions := deriveExtensionsMapFromDocumentProperties(&document.DocumentProperties, encryptedDocumentPropertiesBytes)

	documentBytes = nil

	metadata := data.ResMetadata{
		ResourceType: data.Encrypted,
		ResourceID:   document.ID,
		Hash:         hashBase64,
		Size:         uint64(size),
		Extensions:   extensions,
	}

	// 组装要传入链码的参数，其中密文本体和对称密钥的密文转换为 Base64 编码
	encryptedData := data.EncryptedData{
		Metadata: metadata,
		Data:     base64.StdEncoding.EncodeToString(encryptedDocumentBytes),
		Key:      base64.StdEncoding.EncodeToString(encryptedKeyBytes),
		Policy:   policy,
	}
	encryptedDataBytes, err := json.Marshal(encryptedData)
	if err != nil {
		return "", errors.Wrapf(err, "无法序列化链码参数")
	}

	encryptedData.Data = ""

	chaincodeFcn := "createEncryptedData"
	channelReq := channel.Request{
		ChaincodeID: s.ServiceInfo.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{encryptedDataBytes, []byte(encryptedResourceCreationEventName)},
	}

	resp, err := executeChannelRequestWithTimer(s.ServiceInfo.ChannelClient, &channelReq, "链上存储文档")
	if err != nil {
		return "", GetClassifiedError(chaincodeFcn, err)
	} else {
		return string(resp.TransactionID), nil
	}
}

// CreateOffchainDocument 创建链下加密数字文档。
//
// 参数：
//   数字文档
//   对称密钥（SM2 曲线上的点）
//   访问策略
//
// 返回：
//   交易 ID
func (s *DocumentService) CreateOffchainDocument(document *common.Document, key *ppks.CurvePoint, policy string) (string, error) {
	if document == nil {
		return "", fmt.Errorf("文档对象不能为 nil")
	}

	// 检查 ID 是否为空。若上层忽略此项检查此项为空，将可能对链码层造成混乱。
	if strings.TrimSpace(document.ID) == "" {
		return "", fmt.Errorf("文档 ID 不能为空")
	}

	documentPropertiesBytes, err := json.Marshal(document.DocumentProperties)
	if err != nil {
		return "", errors.Wrap(err, "无法序列化文档属性")
	}
	documentBytes, err := json.Marshal(document)
	if err != nil {
		return "", errors.Wrap(err, "无法序列化文档")
	}

	// 用 key 加密 documentBytes 和 documentPropertiesBytes
	// 使用由 key 导出的 256 位信息来创建 AES256 block
	encryptedDocumentPropertiesBytes, err := encryptDataWithTimer(documentPropertiesBytes, key, "无法加密文档属性", "加密文档属性")
	if err != nil {
		return "", err
	}

	// 提前准备扩展字段，以便回收 `document`
	extensions := deriveExtensionsMapFromDocumentProperties(&document.DocumentProperties, encryptedDocumentPropertiesBytes)
	documentID := document.ID
	document = nil
	runtime.GC()

	// 提前计算原始内容的哈希，获取大小，以便回收 `documentBytes`
	hash := sha256.Sum256(documentBytes)
	hashBase64 := base64.StdEncoding.EncodeToString(hash[:])
	size := len(documentBytes)

	encryptedDocumentBytes, err := encryptDataWithTimer(documentBytes, key, "无法加密文档", "加密文档")
	if err != nil {
		return "", err
	}

	documentBytes = nil
	runtime.GC()

	// 获取集合公钥（当前实现为 SM2 公钥）
	collPubKey, err := s.KeySwitchService.GetCollectiveAuthorityPublicKey()
	if err != nil {
		return "", err
	}
	collPubKeyInSM2 := collPubKey.(*sm2.PublicKey)

	// 用集合公钥加密 key
	encryptedKey, err := ppks.PointEncrypt(collPubKeyInSM2, key)
	if err != nil {
		return "", errors.Wrap(err, "无法加密对称密钥")
	}
	// 序列化加密后的 key
	encryptedKeyBytes := cipherutils.SerializeCipherText(encryptedKey)

	metadata := data.ResMetadata{
		ResourceType: data.Offchain,
		ResourceID:   documentID,
		Hash:         hashBase64,
		Size:         uint64(size),
		Extensions:   extensions,
	}

	// 将加密后的文档上传至 IPFS 网络
	cid, err := uploadBytesToIPFSWithTimer(s.ServiceInfo.IPFSSh, encryptedDocumentBytes, "无法将加密后的文档上传至 IPFS 网络", "上传至 IPFS 网络")
	if err != nil {
		return "", err
	}

	encryptedDocumentBytes = nil
	runtime.GC()

	// 在传给链码的参数中传入 IPFS CID
	offchainData := data.OffchainData{
		Metadata: metadata,
		CID:      cid,
		Key:      base64.StdEncoding.EncodeToString(encryptedKeyBytes),
		Policy:   policy,
	}
	offchainDataBytes, err := json.Marshal(offchainData)
	if err != nil {
		return "", errors.Wrapf(err, "无法序列化链码参数")
	}

	chaincodeFcn := "createOffchainData"
	channelReq := channel.Request{
		ChaincodeID: s.ServiceInfo.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{offchainDataBytes, []byte(encryptedResourceCreationEventName)},
	}

	resp, err := executeChannelRequestWithTimer(s.ServiceInfo.ChannelClient, &channelReq, "链上存储文档元数据与属性")
	if err != nil {
		return "", GetClassifiedError(chaincodeFcn, err)
	}

	return string(resp.TransactionID), nil
}

// GetDocumentMetadata 获取数字文档的元数据。
//
// 参数：
//   文档 ID
//
// 返回：
//   元数据
func (s *DocumentService) GetDocumentMetadata(id string) (*data.ResMetadataStored, error) {
	return getResourceMetadata(id, s.ServiceInfo)
}

// GetDocument 获取明文数字文档，调用前应先获取元数据。
//
// 参数：
//   文档 ID
//   文档元数据
//
// 返回：
//   文档本体
func (s *DocumentService) GetDocument(id string, metadata *data.ResMetadataStored) (*common.Document, error) {
	// 检查元数据中该资源类型是否为明文资源
	if metadata.ResourceType != data.Plain {
		return nil, &ErrorBadRequest{
			errMsg: "该资源不是明文资源。",
		}
	}

	// 调用链码 getData 获取该资源的本体
	chaincodeFcn := "getData"
	channelReq := channel.Request{
		ChaincodeID: s.ServiceInfo.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{[]byte(id)},
	}

	resp, err := s.ServiceInfo.ChannelClient.Query(channelReq)
	if err != nil {
		return nil, GetClassifiedError(chaincodeFcn, err)
	}

	documentBytes := resp.Payload

	// 检查所获数据的大小与哈希是否匹配
	err = checkSizeAndHashForDecryptedData(documentBytes, metadata).toError(metadata.ResourceType)
	if err != nil {
		return nil, err
	}

	// 解析所获数据，得到 common.Document 作为结果
	var document common.Document
	if err = json.Unmarshal(documentBytes, &document); err != nil {
		return nil, fmt.Errorf("获取的数据不是合法的数字文档")
	}
	return &document, nil
}

// GetEncryptedDocument 获取加密数字文档。提供密钥置换会话，函数将使用密钥置换结果尝试进行解密后，返回明文。调用前应先获取元数据。
//
// 参数：
//   文档 ID
//   密钥置换会话 ID
//   预期的份额数量
//   文档元数据
//
// 返回：
//   解密后的文档
func (s *DocumentService) GetEncryptedDocument(id string, keySwitchSessionID string, numSharesExpected int, metadata *data.ResMetadataStored) (*common.Document, error) {
	// 检查元数据中该资源类型是否为密文资源
	if metadata.ResourceType != data.Encrypted {
		return nil, &ErrorBadRequest{
			errMsg: "该资源不是加密资源。",
		}
	}

	// 调用链码 getData 获取该资源的密文本体
	chaincodeFcn := "getData"
	channelReq := channel.Request{
		ChaincodeID: s.ServiceInfo.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{[]byte(id)},
	}

	resp, err := s.ServiceInfo.ChannelClient.Query(channelReq)
	err = GetClassifiedError(chaincodeFcn, err)
	if err != nil {
		return nil, err
	}

	encryptedDocumentBytes := resp.Payload

	// 调用链码 getKey 获取该资源的加密后的密钥
	chaincodeFcn = "getKey"
	channelReq = channel.Request{
		ChaincodeID: s.ServiceInfo.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{[]byte(id)},
	}

	resp, err = s.ServiceInfo.ChannelClient.Query(channelReq)
	err = GetClassifiedError(chaincodeFcn, err)
	if err != nil {
		return nil, err
	}

	// 解析加密后的密钥材料
	encryptedKeyBytes := resp.Payload
	encryptedKeyAsCipherText, err := cipherutils.DeserializeCipherText(encryptedKeyBytes)
	if err != nil {
		return nil, err
	}

	// 检查加密的内容的大小和哈希是否匹配
	err = checkSizeAndHashForEncryptedData(encryptedDocumentBytes, metadata).toError(metadata.ResourceType)
	if err != nil {
		return nil, err
	}

	// 调用链码 listKeySwitchResultsByID 看是否有 numSharesExpected 份。若不足则报错。
	chaincodeFcn = "listKeySwitchResultsByID"
	channelReq = channel.Request{
		ChaincodeID: s.ServiceInfo.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{[]byte(keySwitchSessionID)},
	}

	resp, err = s.ServiceInfo.ChannelClient.Query(channelReq)
	err = GetClassifiedError(chaincodeFcn, err)
	if err != nil {
		return nil, err
	}

	var ksResults []*keyswitch.KeySwitchResultStored
	err = json.Unmarshal(resp.Payload, &ksResults)
	if err != nil {
		return nil, errors.Wrap(err, "无法解析密钥置换结果列表")
	}

	if len(ksResults) != numSharesExpected {
		return nil, fmt.Errorf("密钥置换结果只有 %v 份，不足 %v 份", len(ksResults), numSharesExpected)
	}

	// 解析并验证份额
	shares, err := parseAndVerifySharesFromKeySwitchResults(ksResults, global.KeySwitchKeys.PublicKey, encryptedKeyAsCipherText, s.KeySwitchService)
	if err != nil {
		return nil, err
	}

	// 调用 KeySwitchService 中的 GetDecryptedKey 得到解密的对称密钥材料
	decryptedKey, err := s.KeySwitchService.GetDecryptedKey(shares, encryptedKeyAsCipherText, global.KeySwitchKeys.PrivateKey)
	if err != nil {
		return nil, errors.Wrap(err, "无法解密对称密钥")
	}

	// 用对称密钥解密 encryptedDocumentBytes
	documentBytes, err := cipherutils.DecryptBytesUsingAESKey(encryptedDocumentBytes, cipherutils.DeriveSymmetricKeyBytesFromCurvePoint(decryptedKey))
	if err != nil {
		return nil, errors.Wrap(err, "无法解密文档")
	}

	// 检查解密出的内容的大小和哈希是否匹配
	err = checkSizeAndHashForDecryptedData(documentBytes, metadata).toError(metadata.ResourceType)
	if err != nil {
		return nil, err
	}

	// 解析 documentBytes 为 common.Document
	var document common.Document
	err = json.Unmarshal(documentBytes, &document)
	if err != nil {
		return nil, fmt.Errorf("获取的数据不是合法的数字文档")
	}

	// 将解密的文档属性与内容存入数据库（若已存在则覆盖）
	err = db.SaveDecryptedDocumentAndDocumentPropertiesToLocalDB(&document, metadata.Timestamp, s.ServiceInfo.DB)
	if err != nil {
		return nil, err
	}

	return &document, nil
}

// 获取链下加密数字文档。提供密钥置换会话，函数将从 IPFS 网络获得密文，使用密钥置换结果尝试进行解密后，返回明文。调用前应先获取元数据。
//
// 参数：
//   文档 ID
//   密钥置换会话 ID
//   预期的份额数量
//   文档元数据
//
// 返回：
//   解密后的文档
func (s *DocumentService) GetOffchainDocument(id string, keySwitchSessionID string, numSharesExpected int, metadata *data.ResMetadataStored) (*common.Document, error) {
	// 检查元数据中该资源类型是否为链下加密资源
	if metadata.ResourceType != data.Offchain {
		return nil, &ErrorBadRequest{
			errMsg: "该资源不是链下加密资源。",
		}
	}

	// 调用链码 getData 获取该资源在 IPFS 网络上的 CID
	chaincodeFcn := "getData"
	channelReq := channel.Request{
		ChaincodeID: s.ServiceInfo.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{[]byte(id)},
	}

	resp, err := s.ServiceInfo.ChannelClient.Query(channelReq)
	err = GetClassifiedError(chaincodeFcn, err)
	if err != nil {
		return nil, err
	}

	cidBytes := resp.Payload
	cid := string(cidBytes)

	// 检查链上记录的内容（CID）的大小和哈希是否匹配
	err = checkSizeAndHashForEncryptedData(cidBytes, metadata).toError(metadata.ResourceType)
	if err != nil {
		return nil, err
	}

	// 从 IPFS 网络中获取文档的密文
	reader, err := s.ServiceInfo.IPFSSh.Cat(cid)
	if err != nil {
		return nil, errors.Wrap(err, "无法从 IPFS 网络获取数字文档")
	}
	encryptedDocumentBytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errors.Wrap(err, "无法从 IPFS 网络获取数字文档")
	}

	// 调用链码 getKey 获取该资源的加密后的密钥
	chaincodeFcn = "getKey"
	channelReq = channel.Request{
		ChaincodeID: s.ServiceInfo.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{[]byte(id)},
	}

	resp, err = s.ServiceInfo.ChannelClient.Query(channelReq)
	err = GetClassifiedError(chaincodeFcn, err)
	if err != nil {
		return nil, err
	}

	// 解析加密后的密钥材料
	encryptedKeyBytes := resp.Payload
	encryptedKeyAsCipherText, err := cipherutils.DeserializeCipherText(encryptedKeyBytes)
	if err != nil {
		return nil, err
	}

	// 调用链码 listKeySwitchResultsByID 看是否有 numSharesExpected 份。若不足则报错。
	chaincodeFcn = "listKeySwitchResultsByID"
	channelReq = channel.Request{
		ChaincodeID: s.ServiceInfo.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{[]byte(keySwitchSessionID)},
	}

	resp, err = s.ServiceInfo.ChannelClient.Query(channelReq)
	err = GetClassifiedError(chaincodeFcn, err)
	if err != nil {
		return nil, err
	}

	var ksResults []*keyswitch.KeySwitchResultStored
	err = json.Unmarshal(resp.Payload, &ksResults)
	if err != nil {
		return nil, errors.Wrap(err, "无法解析密钥置换结果列表")
	}

	if len(ksResults) != numSharesExpected {
		return nil, fmt.Errorf("密钥置换结果只有 %v 份，不足 %v 份", len(ksResults), numSharesExpected)
	}

	// 解析并验证份额
	shares, err := parseAndVerifySharesFromKeySwitchResults(ksResults, global.KeySwitchKeys.PublicKey, encryptedKeyAsCipherText, s.KeySwitchService)
	if err != nil {
		return nil, err
	}

	// 调用 KeySwitchService 中的 GetDecryptedKey 得到解密的对称密钥材料
	decryptedKey, err := s.KeySwitchService.GetDecryptedKey(shares, encryptedKeyAsCipherText, global.KeySwitchKeys.PrivateKey)
	if err != nil {
		return nil, errors.Wrap(err, "无法解密对称密钥")
	}

	// 用对称密钥解密 encryptedDocumentBytes
	documentBytes, err := cipherutils.DecryptBytesUsingAESKey(encryptedDocumentBytes, cipherutils.DeriveSymmetricKeyBytesFromCurvePoint(decryptedKey))
	if err != nil {
		return nil, errors.Wrap(err, "无法解密文档")
	}

	// 检查解密出的内容的大小和哈希是否匹配
	err = checkSizeAndHashForDecryptedData(documentBytes, metadata).toError(metadata.ResourceType)
	if err != nil {
		return nil, err
	}

	// 解析 documentBytes 为 common.Document
	var document common.Document
	err = json.Unmarshal(documentBytes, &document)
	if err != nil {
		return nil, fmt.Errorf("获取的数据不是合法的数字文档")
	}

	// 将解密的文档属性与内容存入数据库（若已存在则覆盖）
	err = db.SaveDecryptedDocumentAndDocumentPropertiesToLocalDB(&document, metadata.Timestamp, s.ServiceInfo.DB)
	if err != nil {
		return nil, err
	}

	return &document, nil
}

// GetEncryptedDocumentProperties 获取加密与链下加密数字文档的加密属性部分，并使用密钥置换结果尝试进行解密。调用前应先获取元数据。
//
// 参数：
//   文档 ID
//   密钥置换会话 ID
//   预期的份额数量
//   文档元数据
//
// 返回：
//   解密后的文档属性
func (s *DocumentService) GetEncryptedDocumentProperties(id string, keySwitchSessionID string, numSharesExpected int, metadata *data.ResMetadataStored) (*common.DocumentProperties, error) {
	// 检查该文档是否为 Encrypted 或 Offchain 资源
	if metadata.ResourceType != data.Encrypted && metadata.ResourceType != data.Offchain {
		return nil, &ErrorBadRequest{
			errMsg: "该资源是明文资源。",
		}
	}

	// 调用链码 getKey 获取该资源的加密后的密钥
	chaincodeFcn := "getKey"
	channelReq := channel.Request{
		ChaincodeID: s.ServiceInfo.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{[]byte(id)},
	}

	resp, err := s.ServiceInfo.ChannelClient.Query(channelReq)
	err = GetClassifiedError(chaincodeFcn, err)
	if err != nil {
		return nil, err
	}

	// 解析加密后的密钥材料
	encryptedKeyBytes := resp.Payload
	encryptedKeyAsCipherText, err := cipherutils.DeserializeCipherText(encryptedKeyBytes)
	if err != nil {
		return nil, err
	}

	// 调用链码 listKeySwitchResultsByID 看是否有 numSharesExpected 份。若不足则报错。
	chaincodeFcn = "listKeySwitchResultsByID"
	channelReq = channel.Request{
		ChaincodeID: s.ServiceInfo.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{[]byte(keySwitchSessionID)},
	}

	resp, err = s.ServiceInfo.ChannelClient.Query(channelReq)
	err = GetClassifiedError(chaincodeFcn, err)
	if err != nil {
		return nil, err
	}

	var ksResults []*keyswitch.KeySwitchResultStored
	err = json.Unmarshal(resp.Payload, &ksResults)
	if err != nil {
		return nil, errors.Wrap(err, "无法解析密钥置换结果列表")
	}

	if len(ksResults) != numSharesExpected {
		return nil, fmt.Errorf("密钥置换结果只有 %v 份，不足 %v 份", len(ksResults), numSharesExpected)
	}

	// 解析并验证份额
	shares, err := parseAndVerifySharesFromKeySwitchResults(ksResults, global.KeySwitchKeys.PublicKey, encryptedKeyAsCipherText, s.KeySwitchService)
	if err != nil {
		return nil, err
	}

	// 调用 KeySwitchService 中的 GetDecryptedKey 得到解密的对称密钥材料
	decryptedKey, err := s.KeySwitchService.GetDecryptedKey(shares, encryptedKeyAsCipherText, global.KeySwitchKeys.PrivateKey)
	if err != nil {
		return nil, errors.Wrap(err, "无法解密对称密钥")
	}

	// 获取文档的加密属性部分
	encryptedDocumentPropertiesBytes, err := base64.StdEncoding.DecodeString(metadata.Extensions["encrypted"].(string))
	if err != nil {
		return nil, errors.Wrap(err, "无法解析加密属性")
	}

	// 用对称密钥解密加密的文档属性
	documentPropertiesBytes, err := cipherutils.DecryptBytesUsingAESKey(encryptedDocumentPropertiesBytes, cipherutils.DeriveSymmetricKeyBytesFromCurvePoint(decryptedKey))
	if err != nil {
		return nil, errors.Wrap(err, "无法解密文档属性")
	}

	// 解析解密的文档属性
	var documentProperties common.DocumentProperties
	err = json.Unmarshal(documentPropertiesBytes, &documentProperties)
	if err != nil {
		return nil, errors.Wrap(err, "无法解析解密后的文档属性")
	}

	// 将解密的文档属性存入数据库（若已存在则覆盖）
	err = db.SaveDecryptedDocumentPropertiesToLocalDB(&documentProperties, metadata.Timestamp, s.ServiceInfo.DB)
	if err != nil {
		return nil, err
	}

	return &documentProperties, nil
}

// GetDecryptedDocumentFromDB 从数据库中获取经解密的数字文档。返回解密后的明文。调用前应先获取元数据。
//
// 参数：
//   文档 ID
//   文档元数据
//
// 返回：
//   解密后的文档
func (s *DocumentService) GetDecryptedDocumentFromDB(id string, metadata *data.ResMetadataStored) (*common.Document, error) {
	// 检查元数据中该资源类型是否为密文资源
	if metadata.ResourceType != data.Encrypted && metadata.ResourceType != data.Offchain {
		return nil, &ErrorBadRequest{
			errMsg: "该资源为明文资源。",
		}
	}

	// 从数据库中读取解密后的文档内容部分
	documentDB, err := db.GetDecryptedDocumentContentsFromLocalDB(id, s.ServiceInfo.DB)
	if err != nil {
		return nil, err
	}

	// 如果没有 contents 部分，则此条只可用于获取属性，还不可用于获取整个文档，当 ErrorNotFound 处理。
	if len(documentDB.Contents) == 0 {
		return nil, errorcode.ErrorNotFound
	}

	// 从数据库中读取解密后的文档属性部分
	documentPropertiesDB, err := db.GetDecryptedDocumentPropertiesFromLocalDB(id, s.ServiceInfo.DB)
	if err != nil {
		return nil, err
	}

	document, err := documentDB.ToModel(documentPropertiesDB)
	if err != nil {
		return nil, err
	}

	// 检查解密出的内容的大小和哈希是否匹配
	documentBytes, err := json.Marshal(document)
	if err != nil {
		return nil, errors.Wrap(err, "无法检查文档完整性")
	}

	decryptedSize := uint64(len(documentBytes))
	if decryptedSize != metadata.Size {
		return nil, &ErrorCorruptedDatabaseResult{
			errMsg: "从数据库获取的文档大小不正确",
		}
	}

	decryptedHash := sha256.Sum256(documentBytes)
	decryptedHashBase64 := base64.StdEncoding.EncodeToString(decryptedHash[:])
	if decryptedHashBase64 != metadata.Hash {
		return nil, &ErrorCorruptedDatabaseResult{
			errMsg: "从数据库获取文档哈希不匹配",
		}
	}

	return document, nil
}

// GetDecryptedDocumentPropertiesFromDB 从数据库中获取经解密的数字文档的属性部分。返回解密后的属性明文。调用前应先获取元数据。
//
// 参数：
//   文档 ID
//   文档元数据
//
// 返回：
//   解密后的文档属性
func (s *DocumentService) GetDecryptedDocumentPropertiesFromDB(id string, metadata *data.ResMetadataStored) (*common.DocumentProperties, error) {
	// 检查元数据中该资源类型是否为密文资源
	if metadata.ResourceType != data.Encrypted && metadata.ResourceType != data.Offchain {
		return nil, &ErrorBadRequest{
			errMsg: "该资源为明文资源。",
		}
	}

	// 从数据库中读取解密后的文档属性部分
	documentPropertiesDB, err := db.GetDecryptedDocumentPropertiesFromLocalDB(id, s.ServiceInfo.DB)
	if err != nil {
		return nil, err
	}

	documentProperties := documentPropertiesDB.ToModel()

	return documentProperties, nil
}

// ListDocumentIDsByCreator 获取所有调用者创建的数字文档的资源 ID。
//
// 参数：
//   倒序排列
//   分页大小
//   分页书签
//
// 返回：
//   带分页的资源 ID 列表
func (s *DocumentService) ListDocumentIDsByCreator(isDesc bool, pageSize int, bookmark string) (*query.IDsWithPagination, error) {
	// 调用 listResourceIDsByCreator 拿到一个 ID 列表
	chaincodeFcn := "listResourceIDsByCreator"
	channelReq := channel.Request{
		ChaincodeID: s.ServiceInfo.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{[]byte(documentDataType), []byte(fmt.Sprintf("%v", isDesc)), []byte(strconv.Itoa(pageSize)), []byte(bookmark)},
	}

	resp, err := s.ServiceInfo.ChannelClient.Query(channelReq)
	err = GetClassifiedError(chaincodeFcn, err)
	if err != nil {
		return nil, err
	}

	var resourceIDs query.IDsWithPagination
	err = json.Unmarshal(resp.Payload, &resourceIDs)
	if err != nil {
		return nil, errors.Wrap(err, "无法解析结果列表")
	}

	return &resourceIDs, nil
}

// ListDocumentIDsByConditions 获取满足所提供的搜索条件的数字文档的资源 ID。
//
// 参数：
//   搜索条件
//   分页大小
//
// 返回：
//   带分页的资源 ID 列表
func (s *DocumentService) ListDocumentIDsByConditions(conditions DocumentQueryConditions, pageSize int) (*query.IDsWithPagination, error) {
	// 从两处获取资源 ID。
	// 第一是调用链码从链上获取，这部分的结果包括 明文资源以及所查寻属性为公开的那部分资源 中符合条件的条目；
	// 第二是从本地数据库中获取，这部分内容为 用户已解密过的资源 中符合条件的条目。
	//
	// 最终目标是按资源 ID 正序或倒序（由查询参数决定）排列，凑够 `pageSize` 条结果，带书签信息返回。
	// 关键点在于：
	//   1. 必须从两个来源都获取资源，为确保凑够数量，它们各需查询 `pageSize` 条，在排序过程中剔除多余的部分，剩下的作为本次查询的结果；
	//   2. 两个数据来源得到的内容并非互补的，即它们可能有交集，存在重复。以下一小段解释了这为什么会发生。
	//      若要满足在以下两个用例中都能取到完整的条目信息，我们就将面临这样的副作用。
	//      可能用例 1：资源 x 与资源 y 的 `name` 相同。x 的 `name` 是公开的。y 的 `name` 是非公开的。需要按照 `name` 精确匹配查询。
	//      可能用例 2：资源 x 的 `name` 是公开的，但 `precedingDocumentID` 是非公开的。需要按照 `name` 和 `precedingDocumentID` 精确匹配查询。
	//      为了满足用例 1 下资源 x 和 y 都能被获取到，只需要
	//          - 在链上按 `name` 查找
	//          - 在本地数据库上按 `name` 查找并限定 `is_name_public` 为 `false`。即隐藏了 `name` 的条目的范围内寻找匹配项。
	//      这样在单条件情况下，两个数据源的查询结果互补。然而这并不适用于用例 2 中多条件查询。
	//      在用例 2 中，为了得到资源 x，我们还是从两个数据源中查询。我们不能从链上查得，因为其 `precedingDocumentID` 是非公开的。
	//      但按上例中的方法，我们同样无法从本地数据库中获得，因为 `name` 是公开的。值得提醒的是，我们在搜索时并不知道一个资源的某个属性是否公开。
	//      从而，这决定了我们从本地数据库中获取条目时，**不能将范围限定在所有属性或任意某些属性是非公开属性的条目内**。
	//      换而言之，我们在从本地数据库上搜索时，也要包括属性是公开的那一部分，这将导致用例 1 中，属性公开的资源 x 会从两个数据源中同时获得。
	//
	// 为了解决这些现象，我们应在从两个数据源各获取 `pageSize` 条之后，进行以下操作。简单起见，假设 `pageSize` 为 10。
	//   1: 若两个数据源得到的结果均为空，则直接返回。
	//   2: 为两个数据源分别维护一个 `consumed` 变量，从一个数据源中取得一条就将相应变量 +1。选择哪个数据源作为下一个条目取决于哪个符合我们的排序规则。
	//      因为结果可能重复，我们在加入时，若遇重复项，则只为相应的 `consumed` 变量 +1，而将结果本身舍弃，不将其采纳进结果列表。
	//      我们在两个列表均被遍历完或者结果列表达到 10 条时，停止这一过程。这之后可能出现 4 种情况：
	//      2.1: 结果列表不满 10 条：两个数据源均用完，则采用该列表以及最后的 ID 信息返回。
	//      2.2: 结果列表满 10 条：直接返回以及最后的 ID 信息。

	// 为链码层所用的 CouchDB 生成查询条件。遇到错误视为参数错误。
	couchDBConditions, err := conditions.ToCouchDBConditions()
	if err != nil {
		return nil, &ErrorBadRequest{
			errMsg: err.Error(),
		}
	}
	couchDBConditionsBytes, err := json.Marshal(couchDBConditions)
	if err != nil {
		return nil, errors.Wrap(err, "无法序列化查询条件")
	}
	fmt.Println(string(couchDBConditionsBytes))

	// 生成 GORM 可用的带查询条件的 DB 对象
	gormConditionedDB, err := conditions.ToGormConditionedDB(s.ServiceInfo.DB)
	if err != nil {
		return nil, &ErrorBadRequest{
			errMsg: err.Error(),
		}
	}

	// 从链码获取资源 ID
	chaincodeFcn := "listResourceIDsByConditions"
	channelReq := channel.Request{
		ChaincodeID: s.ServiceInfo.ChaincodeID,
		Fcn:         chaincodeFcn,
		// 单独从链码获取是支持书签的，但这里不用（已经在查询条件中限定了）。为满足 3 个参数，最后的 bookmark 参数为空列表。
		Args: [][]byte{couchDBConditionsBytes, []byte(strconv.Itoa(pageSize)), {}},
	}

	resp, err := s.ServiceInfo.ChannelClient.Query(channelReq)
	err = GetClassifiedError(chaincodeFcn, err)
	if err != nil {
		return nil, err
	}

	// 这里虽然包含查询后的新书签信息，但该书签信息无用
	var chaincodeResourceIDs query.IDsWithPagination
	err = json.Unmarshal(resp.Payload, &chaincodeResourceIDs)
	if err != nil {
		return nil, errors.Wrap(err, "无法解析结果列表")
	}

	// 从本地数据库获取资源 ID
	// TODO: Debug 期使用
	log.Debug(gormConditionedDB.Statement.SQL.String())
	var localDBResourceIDs []string
	gormConditionedDB.Table("documents").Select("id").Limit(pageSize).Find(&localDBResourceIDs)

	// 若两个数据源得到的结果均为空，则直接返回
	if len(chaincodeResourceIDs.IDs) == 0 && len(localDBResourceIDs) == 0 {
		ret := &query.IDsWithPagination{
			IDs: []string{},
		}
		return ret, nil
	}

	// 从两个来源收集资源 ID
	resourceIDs := collectResourceIDsFromSources(chaincodeResourceIDs.IDs, localDBResourceIDs, pageSize, conditions.IsDesc)

	// 空结果早已提前返回，到此结果列表必有值。返回的书签是结果列表最后一项，即最后出现的资源 ID。
	retBookmark := resourceIDs[len(resourceIDs)-1]

	ret := &query.IDsWithPagination{
		IDs:      resourceIDs,
		Bookmark: retBookmark,
	}

	return ret, nil
}

func deriveExtensionsMapFromDocumentProperties(publicProperties *common.DocumentProperties, encryptedProperties []byte) map[string]interface{} {
	extensions := make(map[string]interface{})
	extensions["dataType"] = documentDataType
	if publicProperties.IsNamePublic {
		extensions["name"] = publicProperties.Name
	}
	if publicProperties.IsTypePublic {
		extensions["documentType"] = publicProperties
	}
	if publicProperties.IsPrecedingDocumentIDPublic {
		extensions["precedingDocumentID"] = publicProperties.PrecedingDocumentID
	}
	if publicProperties.IsHeadDocumentIDPublic {
		extensions["headDocumentID"] = publicProperties.HeadDocumentID
	}
	if publicProperties.IsEntityAssetIDPublic {
		extensions["entityAssetID"] = publicProperties.EntityAssetID
	}

	if encryptedProperties != nil {
		encryptedPropertiesBase64 := base64.StdEncoding.EncodeToString(encryptedProperties)
		extensions["encrypted"] = encryptedPropertiesBase64
	}

	return extensions
}
