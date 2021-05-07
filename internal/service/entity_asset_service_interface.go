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

	// 获取明文实体资产。
	//
	// 参数：
	//   资产 ID
	//
	// 返回：
	//   实体资产条目本体
	GetEntityAsset(id string) (*common.EntityAsset, error)

	// 获取加密实体资产。提供密钥置换会话，函数将使用密钥置换结果尝试进行解密后，返回明文。
	//
	// 参数：
	//   资产 ID
	//   密钥置换会话 ID
	//   预期的份额数量
	//
	// 返回：
	//   解密后的实体资产条目
	GetEncryptedEntityAsset(id string, keySwitchSessionID string, numSharesExpected int) (*common.EntityAsset, error)

	// 用于列出与该实体资产有关的文档。
	//
	// 参数：
	//   实体资产 ID
	//   分页大小
	//   分页书签
	//
	// 返回：
	//   带分页的资源 ID 列表
	ListDocumentIDsByEntityID(id string, pageSize int, bookmark string) (*query.ResourceIDsWithPagination, error)
}
