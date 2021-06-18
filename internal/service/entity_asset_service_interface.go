package service

import (
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/models/common"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/data"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/query"
	"github.com/XiaoYao-austin/ppks"
)

// EntityAssetServiceInterface 定义了用于管理实体资产的服务的接口。
type EntityAssetServiceInterface interface {
	// 创建实体资产。
	//
	// 参数：
	//   实体资产
	//
	// 返回：
	//   交易 ID
	CreateEntityAsset(asset *common.EntityAsset) (string, error)

	// 创建加密的实体资产。
	//
	// 参数：
	//   实体资产
	//   加密后的对称密钥
	//   访问策略
	//
	// 返回：
	//   交易 ID
	CreateEncryptedEntityAsset(asset *common.EntityAsset, key *ppks.CurvePoint, policy string) (string, error)

	// 获取实体资产的元数据。
	//
	// 参数：
	//   资产 ID
	//
	// 返回：
	//   资产资源元数据
	GetEntityAssetMetadata(id string) (*data.ResMetadataStored, error)

	// 获取明文实体资产。调用前应先获取元数据。
	//
	// 参数：
	//   资产 ID
	//   资产元数据
	//
	// 返回：
	//   实体资产条目本体
	GetEntityAsset(id string, metadata *data.ResMetadataStored) (*common.EntityAsset, error)

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
	GetEncryptedEntityAsset(id string, keySwitchSessionID string, numSharesExpected int, metadata *data.ResMetadataStored) (*common.EntityAsset, error)

	// GetDecryptedEntityAssetFromDB 从数据库中获取经解密的实体资产。返回解密后的明文。调用前应先获取元数据。
	//
	// 参数：
	//   资产 ID
	//   资产元数据
	//
	// 返回：
	//   解密后的资产
	GetDecryptedEntityAssetFromDB(id string, metadata *data.ResMetadataStored) (*common.EntityAsset, error)

	// ListEntityAssetIDsByConditions 获取满足所提供的搜索条件的实体资产的资源 ID。
	//
	// 参数：
	//   搜索条件
	//   分页大小
	//
	// 返回：
	//   带分页的资源 ID 列表
	ListEntityAssetIDsByConditions(conditions EntityAssetQueryConditions, pageSize int) (*query.IDsWithPagination, error)
}
