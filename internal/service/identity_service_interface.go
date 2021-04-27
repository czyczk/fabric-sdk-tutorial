package service

import (
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/models/common"
)

// IdentityServiceInterface 定义了有关于使用者身份的服务的接口。
type IdentityServiceInterface interface {
	// 获取当前使用者的身份信息。
	//
	// 返回：
	//   用户身份信息
	GetIdentityInfo() (*common.UserIdentity, error)
}
