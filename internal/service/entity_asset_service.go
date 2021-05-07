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
	"github.com/tjfoc/gmsm/sm2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// EntityAssetService 用于管理实体资产。
type EntityAssetService struct {
	ServiceInfo      *Info
	KeySwitchService KeySwitchServiceInterface
}

// 创建实体资产。
//
// 参数：
//   实体资产
//
// 返回：
//   交易 ID
func (s *EntityAssetService) CreateEntityAsset(asset *common.EntityAsset) (string, error) {
	if asset == nil {
		return "", fmt.Errorf("资产对象不能为 nil")
	}

	// 检查 ID 是否为空。若上层忽略此项检查此项为空，将可能对链码层造成混乱。
	if strings.TrimSpace(asset.ID) == "" {
		return "", fmt.Errorf("资产 ID 不能为空")
	}

	assetBytes, err := json.Marshal(asset)
	if err != nil {
		return "", errors.Wrap(err, "无法序列化资产")
	}

	// 计算哈希，获取大小并准备扩展字段
	hash := sha256.Sum256(assetBytes)
	hashBase64 := base64.StdEncoding.EncodeToString(hash[:])
	size := len(assetBytes)
	extensions := deriveExtensionsMapFromAsset(asset)

	metadata := data.ResMetadata{
		ResourceType: data.Plain,
		ResourceID:   asset.ID,
		Hash:         hashBase64,
		Size:         uint64(size),
		Extensions:   extensions,
	}

	// 组装要传入链码的参数，其中数据本体转换为 Base64 编码
	plainData := data.PlainData{
		Metadata: metadata,
		Data:     base64.StdEncoding.EncodeToString(assetBytes),
	}
	plainDataBytes, err := json.Marshal(plainData)
	if err != nil {
		return "", errors.Wrap(err, "无法序列化链码参数")
	}

	// 调用链码将数据上链
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

// 创建加密的实体资产。
//
// 参数：
//   实体资产
//   加密后的对称密钥
//   访问策略
//
// 返回：
//   交易 ID
func (s *EntityAssetService) CreateEncryptedEntityAsset(asset *common.EntityAsset, key *ppks.CurvePoint, policy string) (string, error) {
	if asset == nil {
		return "", fmt.Errorf("资产不能为 nil")
	}

	// 检查 ID 是否为空。若上层忽略此项检查此项为空，将可能对链码层造成混乱。
	if strings.TrimSpace(asset.ID) == "" {
		return "", fmt.Errorf("资产 ID 不能为空")
	}

	assetBytes, err := json.Marshal(asset)
	if err != nil {
		return "", errors.Wrap(err, "无法序列化资产")
	}

	// 用 key 加密 assetBytes
	// 使用由 key 导出的 256 位信息来创建 AES256 block
	cipherBlock, err := aes.NewCipher(deriveSymmetricKeyBytesFromCurvePoint(key))
	if err != nil {
		return "", errors.Wrap(err, "无法加密资产")
	}

	aesGCM, err := cipher.NewGCM(cipherBlock)
	if err != nil {
		return "", errors.Wrap(err, "无法加密资产")
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", errors.Wrap(err, "无法加密资产")
	}

	encryptedAssetBytes := aesGCM.Seal(nonce, nonce, assetBytes, nil)

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
	hash := sha256.Sum256(assetBytes)
	hashBase64 := base64.StdEncoding.EncodeToString(hash[:])
	size := len(assetBytes)
	extensions := deriveExtensionsMapFromAsset(asset)

	metadata := data.ResMetadata{
		ResourceType: data.Encrypted,
		ResourceID:   asset.ID,
		Hash:         hashBase64,
		Size:         uint64(size),
		Extensions:   extensions,
	}

	// 组装要传入链码的参数，其中密文本体和对称密钥的密文转换为 Base64 编码
	encryptedData := data.EncryptedData{
		Metadata: metadata,
		Data:     base64.StdEncoding.EncodeToString(encryptedAssetBytes),
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

// 获取实体资产的元数据。
//
// 参数：
//   资产 ID
//
// 返回：
//   资产资源元数据
func (s *EntityAssetService) GetEntityAssetMetadata(id string) (*data.ResMetadataStored, error) {
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

// 获取明文实体资产。调用前应先获取元数据。
//
// 参数：
//   资产 ID
//   资产元数据
//
// 返回：
//   实体资产条目本体
func (s *EntityAssetService) GetEntityAsset(id string, metadata *data.ResMetadataStored) (*common.EntityAsset, error) {
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
		var asset common.EntityAsset
		if err = json.Unmarshal(resp.Payload, &asset); err != nil {
			return nil, fmt.Errorf("获取的数据不是合法的实体资产")
		}
		return &asset, nil
	}
}

// 获取加密实体资产。提供密钥置换会话，函数将使用密钥置换结果尝试进行解密后，返回明文。调用前应先获取元数据。
//
// 参数：
//   资产 ID
//   密钥置换会话 ID
//   预期的份额数量
//   资产元数据
//
// 返回：
//   解密后的实体资产条目
func (s *EntityAssetService) GetEncryptedEntityAsset(id string, keySwitchSessionID string, numSharesExpected int, metadata *data.ResMetadataStored) (*common.EntityAsset, error) {
	// 检查元数据中该资源类型是否为密文资源
	if metadata.ResourceType != data.Encrypted {
		return nil, &ErrorBadRequest{
			errMsg: "该资源不是加密资源。",
		}
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

	encryptedAssetBytes := resp.Payload

	// 检查加密的内容的大小和哈希是否匹配
	encryptedSize := uint64(len(encryptedAssetBytes))
	if encryptedSize != metadata.SizeStored {
		return nil, fmt.Errorf("获取的密文大小不正确")
	}

	encryptedHash := sha256.Sum256(encryptedAssetBytes)
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
		return nil, errors.Wrap(err, "无法解密资产")
	}

	aesGCM, err := cipher.NewGCM(cipherBlock)
	if err != nil {
		return nil, errors.Wrap(err, "无法解密资产")
	}

	nonceSize := aesGCM.NonceSize()
	if len(encryptedAssetBytes) < nonceSize {
		return nil, fmt.Errorf("无法解密资产: 密文长度太短")
	}

	nonce, encryptedAssetBytes := encryptedAssetBytes[:nonceSize], encryptedAssetBytes[nonceSize:]
	assetBytes, err := aesGCM.Open(nil, nonce, encryptedAssetBytes, nil)
	if err != nil {
		return nil, errors.Wrap(err, "无法解密资产")
	}

	// 检查解密出的内容的大小和哈希是否匹配
	decryptedSize := uint64(len(assetBytes))
	if decryptedSize != metadata.Size {
		return nil, fmt.Errorf("解密后的资产大小不正确")
	}

	decryptedHash := sha256.Sum256(assetBytes)
	decryptedHashBase64 := base64.StdEncoding.EncodeToString(decryptedHash[:])
	if decryptedHashBase64 != metadata.Hash {
		return nil, fmt.Errorf("解密后的资产哈希不匹配")
	}

	// 解析 assetBytes 为 common.EntityAsset
	var asset common.EntityAsset
	err = json.Unmarshal(assetBytes, &asset)
	if err != nil {
		return nil, fmt.Errorf("获取的数据不是合法的实体资产")
	}

	// 将解密的资产存入数据库（若已存在则覆盖）
	assetDB, err := sqlmodel.NewEntityAssetFromModel(&asset)
	if err != nil {
		return nil, err
	}

	dbResult := s.ServiceInfo.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(assetDB)
	if dbResult.Error != nil {
		return nil, errors.Wrap(dbResult.Error, "无法将解密后的实体资产存入数据库")
	}

	return &asset, nil
}

// GetDecryptedEntityAssetFromDB 从数据库中获取经解密的实体资产。返回解密后的明文。调用前应先获取元数据。
//
// 参数：
//   资产 ID
//   资产元数据
//
// 返回：
//   解密后的资产
func (s *DocumentService) GetDecryptedEntityAssetFromDB(id string, metadata *data.ResMetadataStored) (*common.EntityAsset, error) {
	// 检查元数据中该资源类型是否为密文资源
	if metadata.ResourceType != data.Encrypted {
		return nil, &ErrorBadRequest{
			errMsg: "该资源不是加密资源。",
		}
	}

	// 从数据库中读取解密后的文档
	var assetDB sqlmodel.EntityAsset
	dbResult := s.ServiceInfo.DB.Where("id = ?", id).Take(&assetDB)
	if dbResult.Error != nil {
		if errors.Cause(dbResult.Error) == gorm.ErrRecordNotFound {
			return nil, errorcode.ErrorNotFound
		} else {
			return nil, errors.Wrap(dbResult.Error, "无法从数据库中获取资产")
		}
	}

	asset := assetDB.ToModel()

	// 检查解密出的内容的大小和哈希是否匹配
	assetBytes, err := json.Marshal(asset)
	if err != nil {
		return nil, errors.Wrap(err, "无法检查资产完整性")
	}

	decryptedSize := uint64(len(assetBytes))
	if decryptedSize != metadata.Size {
		return nil, &ErrorCorruptedDatabaseResult{
			errMsg: "从数据库获取的资产大小不正确",
		}
	}

	decryptedHash := sha256.Sum256(assetBytes)
	decryptedHashBase64 := base64.StdEncoding.EncodeToString(decryptedHash[:])
	if decryptedHashBase64 != metadata.Hash {
		return nil, &ErrorCorruptedDatabaseResult{
			errMsg: "从数据库获取资产哈希不匹配",
		}
	}

	return asset, nil
}

// 用于列出与该实体资产有关的文档。
//
// 参数：
//   实体资产 ID
//   分页大小
//   分页书签
//
// 返回：
//   带分页的资源 ID 列表
func (s *EntityAssetService) ListDocumentIDsByEntityID(id string, pageSize int, bookmark string) (*query.ResourceIDsWithPagination, error) {
	// 调用链码函数 listDocumentIDsByEntityID 获取有该资产
	chaincodeFcn := "listDocumentIDsByEntityID"
	channelReq := channel.Request{
		ChaincodeID: s.ServiceInfo.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{[]byte(id), []byte(strconv.Itoa(pageSize)), []byte(bookmark)},
	}

	resp, err := s.ServiceInfo.ChannelClient.Query(channelReq)
	err = GetClassifiedError(chaincodeFcn, err)
	if err != nil {
		return nil, errors.Wrap(err, "无法获取相关文档")
	}

	var resourceIDs query.ResourceIDsWithPagination
	err = json.Unmarshal(resp.Payload, &resourceIDs)
	if err != nil {
		return nil, errors.Wrap(err, "无法解析结果列表")
	}

	return &resourceIDs, nil
}

func deriveExtensionsMapFromAsset(asset *common.EntityAsset) map[string]string {
	extensions := make(map[string]string)
	extensions["dataType"] = "entityAsset"
	if asset.IsNamePublic {
		extensions["name"] = asset.Name
	}
	if asset.IsDesignDocumentIDPublic {
		extensions["designDocumentID"] = asset.DesignDocumentID
	}

	return extensions
}
