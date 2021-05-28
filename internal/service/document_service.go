package service

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"sort"
	"strconv"
	"strings"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/global"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/models/common"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/models/sqlmodel"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/errorcode"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/data"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/keyswitch"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/query"
	"github.com/XiaoYao-austin/ppks"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/tjfoc/gmsm/sm2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DocumentService 用于管理数字文档。
type DocumentService struct {
	ServiceInfo      *Info
	KeySwitchService KeySwitchServiceInterface
}

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

	documentBytes, err := json.Marshal(document)
	if err != nil {
		return "", errors.Wrap(err, "无法序列化文档")
	}

	// 计算哈希，获取大小并准备可公开的扩展字段
	hash := sha256.Sum256(documentBytes)
	hashBase64 := base64.StdEncoding.EncodeToString(hash[:])
	size := len(documentBytes)
	extensions := deriveExtensionsMapFromDocument(document)

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
	extensions := deriveExtensionsMapFromDocument(document)

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
	extensions := deriveExtensionsMapFromDocument(document)

	metadata := data.ResMetadata{
		ResourceType: data.Offchain,
		ResourceID:   document.ID,
		Hash:         hashBase64,
		Size:         uint64(size),
		Extensions:   extensions,
	}

	// 将加密后的文档上传至 IPFS 网络
	cid, err := s.ServiceInfo.IPFSSh.Add(bytes.NewReader(encryptedDocumentBytes))
	if err != nil {
		return "", errors.Wrap(err, "无法将加密后的文档上传至 IPFS 网络")
	}

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
		Args:        [][]byte{offchainDataBytes},
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

	encryptedKey := resp.Payload

	// 检查加密的内容的大小和哈希是否匹配
	encryptedSize := uint64(len(encryptedDocumentBytes))
	if encryptedSize != metadata.SizeStored {
		return nil, fmt.Errorf("获取的密文大小不正确")
	}

	encryptedHash := sha256.Sum256(encryptedDocumentBytes)
	encryptedHashBase64 := base64.StdEncoding.EncodeToString(encryptedHash[:])
	if encryptedHashBase64 != metadata.HashStored {
		return nil, fmt.Errorf("获取的密文哈希不匹配")
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

	// 检查解密出的内容的大小和哈希是否匹配
	decryptedSize := uint64(len(documentBytes))
	if decryptedSize != metadata.Size {
		return nil, fmt.Errorf("解密后的文档大小不正确")
	}

	decryptedHash := sha256.Sum256(documentBytes)
	decryptedHashBase64 := base64.StdEncoding.EncodeToString(decryptedHash[:])
	if decryptedHashBase64 != metadata.Hash {
		return nil, fmt.Errorf("解密后的文档哈希不匹配")
	}

	// 解析 documentBytes 为 common.Document
	var document common.Document
	err = json.Unmarshal(documentBytes, &document)
	if err != nil {
		return nil, fmt.Errorf("获取的数据不是合法的数字文档")
	}

	// 将解密的文档存入数据库（若已存在则覆盖）
	documentDB, err := sqlmodel.NewDocumentFromModel(&document, metadata.Timestamp)
	if err != nil {
		return nil, err
	}

	dbResult := s.ServiceInfo.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(documentDB)
	if dbResult.Error != nil {
		return nil, errors.Wrap(dbResult.Error, "无法将解密后的文档存入数据库")
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
	encryptedSize := uint64(len(cidBytes))
	if encryptedSize != metadata.SizeStored {
		return nil, fmt.Errorf("获取的密文大小不正确")
	}

	encryptedHash := sha256.Sum256(cidBytes)
	encryptedHashBase64 := base64.StdEncoding.EncodeToString(encryptedHash[:])
	if encryptedHashBase64 != metadata.HashStored {
		return nil, fmt.Errorf("获取的密文哈希不匹配")
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

	encryptedKey := resp.Payload

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

	// 检查解密出的内容的大小和哈希是否匹配
	decryptedSize := uint64(len(documentBytes))
	if decryptedSize != metadata.Size {
		return nil, fmt.Errorf("解密后的文档大小不正确")
	}

	decryptedHash := sha256.Sum256(documentBytes)
	decryptedHashBase64 := base64.StdEncoding.EncodeToString(decryptedHash[:])
	if decryptedHashBase64 != metadata.Hash {
		return nil, fmt.Errorf("解密后的文档哈希不匹配")
	}

	// 解析 documentBytes 为 common.Document
	var document common.Document
	err = json.Unmarshal(documentBytes, &document)
	if err != nil {
		return nil, fmt.Errorf("获取的数据不是合法的数字文档")
	}

	// 将解密的文档存入数据库（若已存在则覆盖）
	documentDB, err := sqlmodel.NewDocumentFromModel(&document, metadata.Timestamp)
	if err != nil {
		return nil, err
	}

	dbResult := s.ServiceInfo.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(documentDB)
	if dbResult.Error != nil {
		return nil, errors.Wrap(dbResult.Error, "无法将解密后的文档存入数据库")
	}

	return &document, nil
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

	// 从数据库中读取解密后的文档
	var documentDB sqlmodel.Document
	dbResult := s.ServiceInfo.DB.Where("id = ?", id).Take(&documentDB)
	if dbResult.Error != nil {
		if errors.Cause(dbResult.Error) == gorm.ErrRecordNotFound {
			return nil, errorcode.ErrorNotFound
		} else {
			return nil, errors.Wrap(dbResult.Error, "无法从数据库中获取文档")
		}
	}

	document := documentDB.ToModel()

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

// ListDocumentIDsByCreator 获取所有调用者创建的数字文档的资源 ID。
//
// 参数：
//   分页大小
//   分页书签
//
// 返回：
//   带分页的资源 ID 列表
func (s *DocumentService) ListDocumentIDsByCreator(pageSize int, bookmark string) (*query.ResourceIDsWithPagination, error) {
	// 调用 listDocumentIDsByCreator 拿到一个 ID 列表
	chaincodeFcn := "listDocumentIDsByCreator"
	channelReq := channel.Request{
		ChaincodeID: s.ServiceInfo.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{[]byte(strconv.Itoa(pageSize)), []byte(bookmark)},
	}

	resp, err := s.ServiceInfo.ChannelClient.Query(channelReq)
	err = GetClassifiedError(chaincodeFcn, err)
	if err != nil {
		return nil, err
	}

	var resourceIDs query.ResourceIDsWithPagination
	err = json.Unmarshal(resp.Payload, &resourceIDs)
	if err != nil {
		return nil, errors.Wrap(err, "无法解析结果列表")
	}

	return &resourceIDs, nil
}

// TODO: 删掉
// ListDocumentIDsByPartialName 获取名称包含所提供的部分名称的数字文档的资源 ID。
//
// 参数：
//   部分名称
//   分页大小
//   分页书签
//
// 返回：
//   带分页的资源 ID 列表
func (s *DocumentService) ListDocumentIDsByPartialName(partialName string, pageSize int, bookmark string) (*query.ResourceIDsWithPagination, error) {
	// 调用 listDocumentIDsByPartialName 拿到一个 ID 列表
	chaincodeFcn := "listDocumentIDsByPartialName"
	channelReq := channel.Request{
		ChaincodeID: s.ServiceInfo.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{[]byte(partialName), []byte(strconv.Itoa(pageSize)), []byte(bookmark)},
	}

	resp, err := s.ServiceInfo.ChannelClient.Query(channelReq)
	err = GetClassifiedError(chaincodeFcn, err)
	if err != nil {
		return nil, err
	}

	var resourceIDs query.ResourceIDsWithPagination
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
//   分页书签
//
// 返回：
//   带分页的资源 ID 列表
func (s *DocumentService) ListDocumentIDsByConditions(conditions DocumentQueryConditions, pageSize int, bookmarks QueryBookmarks) (*query.ResourceIDsWithPagination, error) {
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
	//   2: 将两个数据源得到的结果按用户需求各自正序或倒序排列。
	//   3: 为两个数据源分别维护一个 `consumed` 变量，从一个数据源中取得一条就将相应变量 +1。选择哪个数据源作为下一个条目取决于哪个符合我们的排序规则。
	//      因为结果可能重复，我们在加入时，若遇重复项，则只为相应的 `consumed` 变量 +1，而将结果本身舍弃，不将其采纳进结果列表。
	//      我们在两个列表均被遍历完或者结果列表达到 10 条时，停止这一过程。这之后可能出现 4 种情况：
	//      3.1: 结果列表不满 10 条：已经遍历完的数据源如果它这次获取的量达到 10 条（说明没到底）需要用适当的书签再获取一次，然后回到第 1 步（或者说重置该数据源的计数器使循环继续）。
	//           若两个数据源获取的量均 < 10 条，则采用该列表以及最新的书签信息返回。
	//      3.2: 两个来源的结果都被用完：直接返回。返回结果中，链码书签就是链码所返回的书签，消耗数量为 0；本地数据库书签对应 offset 值，即上次的值 + 相应 `consumed`。
	//      3.3: 链码来源的结果没用完：链码书签是上次的链码书签，消耗数量为这次从链码中消耗的数量；本地数据库书签为上次的值 + 相应 `consumed`。
	//      3.4: 本地数据库结果没用完：链码书签是最新书签；本地数据库书签即上次的值 + `pageSize`。

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

	// 生成 GORM 可用的带查询条件的 DB 对象
	gormConditionedDB, err := conditions.ToGormConditionedDB(s.ServiceInfo.DB)
	if err != nil {
		return nil, &ErrorBadRequest{
			errMsg: err.Error(),
		}
	}

	// 从链码获取资源 ID
	chaincodeFcn := "listDocumentIDsByConditions"
	channelReq := channel.Request{
		ChaincodeID: s.ServiceInfo.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{couchDBConditionsBytes, []byte(strconv.Itoa(pageSize)), []byte(bookmarks.ChaincodeBookmark)},
	}

	resp, err := s.ServiceInfo.ChannelClient.Query(channelReq)
	err = GetClassifiedError(chaincodeFcn, err)
	if err != nil {
		return nil, err
	}

	// 这里将包含查询后的新书签信息。稍后应根据此次查询的内容是否被用完来决定使用旧书签还是新书签。
	var chaincodeResourceIDs query.ResourceIDsWithPagination
	err = json.Unmarshal(resp.Payload, &chaincodeResourceIDs)
	if err != nil {
		return nil, errors.Wrap(err, "无法解析结果列表")
	}

	// 从本地数据库获取资源 ID
	// TODO: Debug 期使用
	log.Debug(gormConditionedDB.Statement.SQL.String())
	var localDBResourceIDs []string
	gormConditionedDB.Table("documents").Select("id").Limit(pageSize).Offset(bookmarks.LocalDBBookmark).Find(&localDBResourceIDs)

	// 若两个数据源得到的结果均为空，则直接返回
	if len(chaincodeResourceIDs.ResourceIDs) == 0 && len(localDBResourceIDs) == 0 {
		retBookmarks := &QueryBookmarks{
			ChaincodeBookmark:           chaincodeResourceIDs.Bookmark,
			ChaincodeEntriesNumConsumed: 0,
			LocalDBBookmark:             0,
		}
		ret := &query.ResourceIDsWithPagination{
			ResourceIDs: []string{},
			Bookmark:    retBookmarks.ToBase64(),
		}
		return ret, nil
	}

	// 将两个数据源得到的结果按用户需求各自正序或倒序排列
	if !conditions.IsReverse {
		sort.Strings(chaincodeResourceIDs.ResourceIDs)
		sort.Strings(localDBResourceIDs)
	} else {
		sort.Sort(sort.Reverse(sort.StringSlice(chaincodeResourceIDs.ResourceIDs)))
		sort.Sort(sort.Reverse(sort.StringSlice(localDBResourceIDs)))
	}

	// 为两个数据源分别维护一个 `consumed` 变量，从一个数据源中取得一条就将相应变量 +1。
	var resourceIDs []string
	chaincodeSrcConsumed := bookmarks.ChaincodeEntriesNumConsumed
	localSrcConsumed := 0

	// 从两个来源采纳条目进结果列表。当结果列表的条目数量足够，或者两个来源均被遍历完，则停止这一过程
	for len(resourceIDs) < pageSize && chaincodeSrcConsumed < len(chaincodeResourceIDs.ResourceIDs) && localSrcConsumed < len(localDBResourceIDs) {
		// 按用户需求选择下一个条目。因为结果可能重复，在加入时若遇重复项，则只增对应的 `consumed` 变量，而将结果舍弃。
		if !conditions.IsReverse {

		}
	}
}

func deriveExtensionsMapFromDocument(document *common.Document) map[string]string {
	extensions := make(map[string]string)
	extensions["dataType"] = "document"
	if document.IsNamePublic {
		extensions["name"] = document.Name
	}
	if document.IsPrecedingDocumentIDPublic {
		extensions["precedingDocumentID"] = document.PrecedingDocumentID
	}
	if document.IsHeadDocumentIDPublic {
		extensions["headDocumentID"] = document.HeadDocumentID
	}
	if document.IsEntityAssetIDPublic {
		extensions["entityAssetID"] = document.EntityAssetID
	}

	return extensions
}
