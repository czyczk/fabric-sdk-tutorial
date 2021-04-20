package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/global"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/models/common"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/errorcode"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/data"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/keyswitch"
	"github.com/XiaoYao-austin/ppks"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/pkg/errors"
	"github.com/tjfoc/gmsm/sm2"
)

// DocumentService 用于管理数字文档。
type DocumentService struct {
	ServiceInfo      *Info
	KeySwitchService KeySwitchServiceInterface
}

// CreateDocument 创建数字文档。
//
// 参数：
//   文档 ID
//   文档名称
//   文档内容
//   文档属性（JSON）
//
// 返回：
//   交易 ID
func (s *DocumentService) CreateDocument(id string, name string, contents []byte, property string) (string, error) {
	// 检查 ID 是否为空。若上层忽略此项检查此项为空，将可能对链码层造成混乱。
	if strings.TrimSpace(id) == "" {
		return "", fmt.Errorf("文档 ID 不能为空")
	}

	document := common.Document{
		ID:       id,
		Name:     name,
		Contents: contents,
		Property: property,
	}

	documentBytes, err := json.Marshal(document)
	if err != nil {
		return "", errors.Wrap(err, "无法序列化文档")
	}

	// 计算哈希，获取大小并准备扩展字段
	hash := sha256.Sum256(documentBytes)
	hashBase64 := base64.StdEncoding.EncodeToString(hash[:])
	size := len(documentBytes)
	extensions := make(map[string]string)
	extensions["name"] = name
	extensionsBytes, err := json.Marshal(extensions)
	if err != nil {
		return "", errors.Wrap(err, "无法序列化扩展字段")
	}

	metadata := data.ResMetadata{
		ResourceType: data.Plain,
		ResourceID:   id,
		Hash:         hashBase64,
		Size:         uint64(size),
		Extensions:   string(extensionsBytes),
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
//   文档 ID
//   文档名称
//   文档内容
//   文档属性（JSON）
//   对称密钥（SM2 曲线上的点）
//   访问策略
//
// 返回：
//   交易 ID
func (s *DocumentService) CreateEncryptedDocument(id string, name string, contents []byte, property string, key *ppks.CurvePoint, policy string) (string, error) {
	// 检查 ID 是否为空。若上层忽略此项检查此项为空，将可能对链码层造成混乱。
	if strings.TrimSpace(id) == "" {
		return "", fmt.Errorf("文档 ID 不能为空")
	}

	document := common.Document{
		ID:       id,
		Name:     name,
		Contents: contents,
		Property: property,
	}

	documentBytes, err := json.Marshal(document)
	if err != nil {
		return "", errors.Wrap(err, "无法序列化文档")
	}

	// 用 key 加密 documentBytes
	// 使用由 key 导出的 256 位信息来创建 AES256 block
	cipherBlock, err := aes.NewCipher(deriveSymmetricKeyBytesFromCurvePoint(key))
	if err != nil {
		return "", errors.Wrap(err, "无法加密文档")
	}

	aesGCM, err := cipher.NewGCM(cipherBlock)
	if err != nil {
		return "", errors.Wrap(err, "无法加密文档")
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", errors.Wrap(err, "无法加密文档")
	}

	encryptedDocumentBytes := aesGCM.Seal(nonce, nonce, documentBytes, nil)

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
	encryptedKeyBytes := SerializeCipherText(encryptedKey)

	// 计算原始内容的哈希，获取大小并准备扩展字段
	hash := sha256.Sum256(documentBytes)
	hashBase64 := base64.StdEncoding.EncodeToString(hash[:])
	size := len(documentBytes)
	extensions := make(map[string]string)
	extensions["name"] = name
	extensionsBytes, err := json.Marshal(extensions)
	if err != nil {
		return "", errors.Wrap(err, "无法序列化扩展字段")
	}

	metadata := data.ResMetadata{
		ResourceType: data.Encrypted,
		ResourceID:   id,
		Hash:         hashBase64,
		Size:         uint64(size),
		Extensions:   string(extensionsBytes),
	}

	// TODO: policy 要强制加上 regulator

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

	chaincodeFcn := "createEncryptedData"
	channelReq := channel.Request{
		ChaincodeID: s.ServiceInfo.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{encryptedDataBytes},
	}

	resp, err := s.ServiceInfo.ChannelClient.Execute(channelReq)
	if err != nil {
		return "", GetClassifiedError(chaincodeFcn, err)
	} else {
		return string(resp.TransactionID), nil
	}
}

// CreateRegulatorEncryptedDocument 创建监管者加密数字文档。
//
// 参数：
//   文档 ID
//   文档名称
//   文档内容
//   文档属性（JSON）
//   对称密钥（SM2 曲线上的点）
//
// 返回：
//   交易 ID
func (s *DocumentService) CreateRegulatorEncryptedDocument(id string, name string, contents []byte, property string, key *ppks.CurvePoint) (string, error) {
	return "", errorcode.ErrorNotImplemented
}

// CreateOffchainDocument 创建链下加密数字文档。
//
// 参数：
//   文档 ID
//   文档名称
//   文档属性（JSON）
//   对称密钥（SM2 曲线上的点）
//   访问策略
//
// 返回：
//   交易 ID
func (s *DocumentService) CreateOffchainDocument(id string, name string, property string, key *ppks.CurvePoint, policy string) (string, error) {
	// 检查 ID 是否为空。若上层忽略此项检查此项为空，将可能对链码层造成混乱。
	if strings.TrimSpace(id) == "" {
		return "", fmt.Errorf("文档 ID 不能为空")
	}

	document := common.Document{
		ID:       id,
		Name:     name,
		Property: property,
	}

	documentBytes, err := json.Marshal(document)
	if err != nil {
		return "", errors.Wrap(err, "无法序列化文档")
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
	encryptedKeyBytes := SerializeCipherText(encryptedKey)

	// 计算原始内容的哈希，获取大小并准备扩展字段
	hash := sha256.Sum256(documentBytes)
	hashBase64 := base64.StdEncoding.EncodeToString(hash[:])
	size := len(documentBytes)
	extensions := make(map[string]string)
	extensions["name"] = name
	extensionsBytes, err := json.Marshal(extensions)
	if err != nil {
		return "", errors.Wrap(err, "无法序列化扩展字段")
	}

	metadata := data.ResMetadata{
		ResourceType: data.Plain,
		ResourceID:   id,
		Hash:         hashBase64,
		Size:         uint64(size),
		Extensions:   string(extensionsBytes),
	}

	// TODO: policy 要强制加上 regulator

	// 组装要传入链码的参数，其中密文本体和对称密钥的密文转换为 Base64 编码
	encryptedData := data.OffchainData{
		Metadata: metadata,
		Key:      base64.StdEncoding.EncodeToString(encryptedKeyBytes),
		Policy:   policy,
	}
	encryptedDataBytes, err := json.Marshal(encryptedData)
	if err != nil {
		return "", errors.Wrapf(err, "无法序列化链码参数")
	}

	chaincodeFcn := "createOffchainData"
	channelReq := channel.Request{
		ChaincodeID: s.ServiceInfo.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{encryptedDataBytes},
	}

	resp, err := s.ServiceInfo.ChannelClient.Execute(channelReq)
	if err != nil {
		return "", GetClassifiedError(chaincodeFcn, err)
	} else {
		return string(resp.TransactionID), nil
	}
}

// GetDocumentMetadata 获取数字文档的元数据。
//
// 参数：
//   文档 ID
//
// 返回：
//   元数据
func (s *DocumentService) GetDocumentMetadata(id string) (*data.ResMetadataStored, error) {
	chaincodeFcn := "getMetadata"
	channelReq := channel.Request{
		ChaincodeID: s.ServiceInfo.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{[]byte(id)},
	}

	resp, err := s.ServiceInfo.ChannelClient.Query(channelReq)
	if err != nil {
		return nil, GetClassifiedError(chaincodeFcn, err)
	} else {
		var resMetadataStored data.ResMetadataStored
		if err = json.Unmarshal(resp.Payload, &resMetadataStored); err != nil {
			return nil, errors.Wrap(err, "获取的元数据不合法")
		}
		return &resMetadataStored, nil
	}
}

// GetDocument 获取明文数字文档。
//
// 参数：
//   文档 ID
//
// 返回：
//   文档本体
func (s *DocumentService) GetDocument(id string) (*common.Document, error) {
	// 检查元数据中该资源类型是否为明文资源
	resMetadataStored, err := s.GetDocumentMetadata(id)
	if err != nil {
		return nil, err
	}
	if resMetadataStored.ResourceType != data.Plain {
		return nil, fmt.Errorf("该资源不是明文资源")
	}

	chaincodeFcn := "getData"
	channelReq := channel.Request{
		ChaincodeID: s.ServiceInfo.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{[]byte(id)},
	}

	resp, err := s.ServiceInfo.ChannelClient.Query(channelReq)
	if err != nil {
		return nil, GetClassifiedError(chaincodeFcn, err)
	} else {
		var document common.Document
		if err = json.Unmarshal(resp.Payload, &document); err != nil {
			return nil, fmt.Errorf("获取的数据不是合法的数字文档")
		}
		return &document, nil
	}
}

// GetEncryptedDocument 获取加密数字文档。提供密钥置换会话，函数将使用密钥置换结果尝试进行解密后，返回明文。
//
// 参数：
//   文档 ID
//   密钥置换会话 ID
//   预期的份额数量
//
// 返回：
//   解密后的文档
func (s *DocumentService) GetEncryptedDocument(id string, keySwitchSessionID string, numSharesExpected int) (*common.Document, error) {
	// 检查元数据中该资源类型是否为密文资源
	resMetadataStored, err := s.GetDocumentMetadata(id)
	if err != nil {
		return nil, err
	}
	if resMetadataStored.ResourceType != data.Encrypted {
		return nil, fmt.Errorf("该资源不是加密资源")
	}

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

	encryptedKey := resp.Payload

	// 调用链码 getData 获取该资源的密文本体
	chaincodeFcn = "getData"
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

	encryptedDocumentBytes := resp.Payload

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

	var ksResults []keyswitch.KeySwitchResultStored
	err = json.Unmarshal(resp.Payload, &ksResults)
	if err != nil {
		return nil, errors.Wrap(err, "无法解析密钥置换结果列表")
	}

	if len(ksResults) != numSharesExpected {
		return nil, fmt.Errorf("密钥置换结果只有 %v 份，不足 %v 份", len(ksResults), numSharesExpected)
	}

	// 调用 KeySwitchService 中的 GetDecryptedKey 得到解密的对称密钥材料
	var shares [][]byte
	for _, ksResult := range ksResults {
		share, err := base64.StdEncoding.DecodeString(ksResult.Share)
		if err != nil {
			return nil, errors.Wrap(err, "无法解析份额")
		}
		shares = append(shares, share)
	}

	decryptedKey, err := s.KeySwitchService.GetDecryptedKey(shares, encryptedKey, global.KeySwitchKeys.PrivateKey)
	if err != nil {
		return nil, errors.Wrap(err, "无法解密对称密钥")
	}

	// 用对称密钥解密 encryptedDocumentBytes
	cipherBlock, err := aes.NewCipher(deriveSymmetricKeyBytesFromCurvePoint(decryptedKey))
	if err != nil {
		return nil, errors.Wrap(err, "无法解密文档")
	}

	aesGCM, err := cipher.NewGCM(cipherBlock)
	if err != nil {
		return nil, errors.Wrap(err, "无法解密文档")
	}

	nonceSize := aesGCM.NonceSize()
	if len(encryptedDocumentBytes) < nonceSize {
		return nil, fmt.Errorf("无法解密文档: 密文长度太短")
	}

	nonce, encryptedDocumentBytes := encryptedDocumentBytes[:nonceSize], encryptedDocumentBytes[nonceSize:]
	documentBytes, err := aesGCM.Open(nil, nonce, encryptedDocumentBytes, nil)
	if err != nil {
		return nil, errors.Wrap(err, "无法解密文档")
	}

	// 解析 documentBytes 为 common.Document
	var document common.Document
	err = json.Unmarshal(documentBytes, &document)
	if err != nil {
		return nil, fmt.Errorf("获取的数据不是合法的数字文档")
	}

	return &document, nil
}

// GetRegulatorEncryptedDocument 获取由监管者公钥加密的文档。函数将获取数据本体并尝试使用调用者的公钥解密后，返回明文。
//
// 参数：
//   文档 ID
//
//  返回：
//    解密后的文档
func (s *DocumentService) GetRegulatorEncryptedDocument(id string) (*common.Document, error) {
	return nil, errorcode.ErrorNotImplemented
}

// ListDocumentIDsByCreator 获取所有调用者创建的数字文档的资源 ID。
//
// 返回：
//   资源 ID 列表
func (s *DocumentService) ListDocumentIDsByCreator() ([]string, error) {
	// 调用 listDocumentIDsByCreator 拿到一个 ID 列表
	chaincodeFcn := "listDocumentIDsByCreator"
	channelReq := channel.Request{
		ChaincodeID: s.ServiceInfo.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{},
	}

	resp, err := s.ServiceInfo.ChannelClient.Query(channelReq)
	err = GetClassifiedError(chaincodeFcn, err)
	if err != nil {
		return nil, err
	}

	var resourceIDs []string
	err = json.Unmarshal(resp.Payload, &resourceIDs)
	if err != nil {
		return nil, errors.Wrap(err, "无法解析结果列表")
	}

	return resourceIDs, nil
}

// ListDocumentIDsByPartialName 获取名称包含所提供的部分名称的数字文档的资源 ID。
//
// 返回：
//   资源 ID 列表
func (s *DocumentService) ListDocumentIDsByPartialName(partialName string) ([]string, error) {
	// 调用 listDocumentIDsByPartialName 拿到一个 ID 列表
	chaincodeFcn := "listDocumentIDsByPartialName"
	channelReq := channel.Request{
		ChaincodeID: s.ServiceInfo.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{},
	}

	resp, err := s.ServiceInfo.ChannelClient.Query(channelReq)
	err = GetClassifiedError(chaincodeFcn, err)
	if err != nil {
		return nil, err
	}

	var resourceIDs []string
	err = json.Unmarshal(resp.Payload, &resourceIDs)
	if err != nil {
		return nil, errors.Wrap(err, "无法解析结果列表")
	}

	return resourceIDs, nil
}

// 对称密钥的生成是由 curvePoint 导出的 256 位信息，可用于创建 AES256 block
func deriveSymmetricKeyBytesFromCurvePoint(curvePoint *ppks.CurvePoint) []byte {
	return curvePoint.X.Bytes()
}
