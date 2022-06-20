package service

import (
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/bcao"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/data"
)

// ResourceService 用于管理资源。
type ResourceService struct {
	ServiceInfo *Info
	DataBCAO    bcao.IDataBCAO
}

// GetResourceMetadata 获取资源的元数据。
//
// 参数：
//   资源 ID
//
// 返回：
//   元数据
func (s *ResourceService) GetResourceMetadata(id string) (*data.ResMetadataStored, error) {
	return getResourceMetadata(id, s.DataBCAO)
}
