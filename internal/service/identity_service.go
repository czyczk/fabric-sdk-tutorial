package service

import (
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/appinit"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/bcao"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/models/common"
)

// IdentityService 实现了 `IdentityServiceInterface` 接口，提供有关于使用者身份的服务
type IdentityService struct {
	ServiceInfo  *Info
	IdentityBCAO bcao.IIdentityBCAO
	ServerInfo   *appinit.ServerInfo
}

// 获取当前使用者的身份信息。
//
// 返回：
//   用户身份信息
func (s *IdentityService) GetIdentityInfo() (*common.UserIdentity, error) {
	deptIdentity, err := s.IdentityBCAO.GetDepartmentIdentity()
	if err != nil {
		return nil, err
	}

	userIdentity := common.UserIdentity{
		UserID:        s.ServerInfo.User.UserID,
		OrgName:       s.ServerInfo.User.OrgName,
		DeptType:      deptIdentity.DeptType,
		DeptLevel:     deptIdentity.DeptLevel,
		DeptName:      deptIdentity.DeptName,
		SuperDeptName: deptIdentity.SuperDeptName,
	}

	return &userIdentity, nil
}
