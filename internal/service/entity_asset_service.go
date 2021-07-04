package service

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/db"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/global"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/models/common"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/utils/cipherutils"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/data"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/keyswitch"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/query"
	"github.com/XiaoYao-austin/ppks"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/tjfoc/gmsm/sm2"
)

// EntityAssetService 用于管理实体资产。
type EntityAssetService struct {
	ServiceInfo      *Info
	KeySwitchService KeySwitchServiceInterface
}

// 用于放置在元数据的 extensions.dataType 中的值
const entityAssetDataType = "entityAsset"

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
	encryptedAssetBytes, err := cipherutils.EncryptBytesUsingAESKey(assetBytes, cipherutils.DeriveSymmetricKeyBytesFromCurvePoint(key))
	if err != nil {
		return "", errors.Wrap(err, "无法加密资产")
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
		Args:        [][]byte{encryptedDataBytes, []byte(encryptedResourceCreationEventName)},
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
	return getResourceMetadata(id, s.ServiceInfo)
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

	// 用对称密钥解密 encryptedAssetBytes
	assetBytes, err := cipherutils.DecryptBytesUsingAESKey(encryptedAssetBytes, cipherutils.DeriveSymmetricKeyBytesFromCurvePoint(decryptedKey))
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
	if err := db.SaveDecryptedEntityAssetToLocalDB(&asset, metadata.Timestamp, s.ServiceInfo.DB); err != nil {
		return nil, err
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
func (s *EntityAssetService) GetDecryptedEntityAssetFromDB(id string, metadata *data.ResMetadataStored) (*common.EntityAsset, error) {
	// 检查元数据中该资源类型是否为密文资源
	if metadata.ResourceType != data.Encrypted {
		return nil, &ErrorBadRequest{
			errMsg: "该资源不是加密资源。",
		}
	}

	// 从数据库中读取解密后的实体资产
	assetDB, err := db.GetDecryptedEntityAssetFromLocalDB(id, s.ServiceInfo.DB)
	if err != nil {
		return nil, err
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

// ListEntityAssetIDsByCreator 获取所有调用者创建的实体资产的资源 ID。
//
// 参数：
//   倒序排列
//   分页大小
//   分页书签
//
// 返回：
//   带分页的资源 ID 列表
func (s *EntityAssetService) ListEntityAssetIDsByCreator(isDesc bool, pageSize int, bookmark string) (*query.IDsWithPagination, error) {
	// 调用 listResourceIDsByCreator 拿到一个 ID 列表
	chaincodeFcn := "listResourceIDsByCreator"
	channelReq := channel.Request{
		ChaincodeID: s.ServiceInfo.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{[]byte(entityAssetDataType), []byte(fmt.Sprintf("%v", isDesc)), []byte(strconv.Itoa(pageSize)), []byte(bookmark)},
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

// ListEntityAssetIDsByConditions 获取满足所提供的搜索条件的实体资产的资源 ID。
//
// 参数：
//   搜索条件
//   分页大小
//
// 返回：
//   带分页的资源 ID 列表
func (s *EntityAssetService) ListEntityAssetIDsByConditions(conditions EntityAssetQueryConditions, pageSize int) (*query.IDsWithPagination, error) {
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
	gormConditionedDB.Table("entity_assets").Select("id").Limit(pageSize).Find(&localDBResourceIDs)

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

func deriveExtensionsMapFromAsset(asset *common.EntityAsset) map[string]interface{} {
	extensions := make(map[string]interface{})
	extensions["dataType"] = entityAssetDataType
	if asset.IsNamePublic {
		extensions["name"] = asset.Name
	}
	if asset.IsDesignDocumentIDPublic {
		extensions["designDocumentID"] = asset.DesignDocumentID
	}

	return extensions
}
