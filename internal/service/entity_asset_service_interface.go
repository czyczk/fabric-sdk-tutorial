package service

import (
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/models/common"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/data"
	"github.com/XiaoYao-austin/ppks"
)

// EntityAssetServiceInterface 定义了用于管理实体资产的服务的接口。
type EntityAssetServiceInterface interface {
	// 创建实体资产。
	//
	// 参数：
	//   资产 ID
	//   资产名称
	//   组件 ID 列表
	//   扩展字段（JSON）
	//
	// 返回：
	//   交易 ID
	CreateEntityAsset(id string, name string, componenetsIDs []string, property string) (string, error)

	// 创建加密的实体资产。
	//
	// 参数：
	//   资产 ID
	//   资产名称
	//   组件 ID 列表
	//   扩展字段（JSON）
	//   加密后的对称密钥
	//   访问策略
	//
	// 返回：
	//   交易 ID
	CreateEncryptedEntityAsset(id string, name string, componentsIDs []string, property string, key *ppks.CurvePoint, policy string) (string, error)

	// 创建一条资产移交记录。
	//
	// 参数：
	//   移交记录 ID
	//   资产 ID
	//   新拥有者（身份的 key）
	//
	// 返回：
	//   交易 ID
	TransferEntityAsset(transferRecordID string, entityID string, newOwner string) (string, error)

	// 获取实体资产的元数据。
	//
	// 参数：
	//   资产 ID
	//
	// 返回：
	//   资产资源元数据
	GetEntityAssetMetadata(id string) (*data.ResMetadataStored, error)

	GetEntityAsset(id string) (*common.EntityAsset, error)

	GetEncryptedEntityAsset(id string, keySwitchSessionID string, numSharesExpected int) (*common.EntityAsset, error)

	// 用于列出与该实体资源有关的文档。
	//
	// 参数：
	//   实体资产 ID
	//
	// 返回：
	//   资源 ID 列表
	ListDocumentIDsByEntityID(id string) ([]string, error)
}
