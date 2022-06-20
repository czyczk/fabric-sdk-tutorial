package service

import (
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/data"
)

// ResourceServiceInterface 定义了用于管理资源的服务的接口。
type ResourceServiceInterface interface {
	// 获取资源的元数据
	//
	// 参数：
	//   资源 ID
	//
	// 返回：
	//   元数据
	GetResourceMetadata(id string) (*data.ResMetadataStored, error)
}
