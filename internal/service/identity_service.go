package service

import (
	"encoding/json"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/appinit"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/models/common"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/identity"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/pkg/errors"
)

// IdentityService 实现了 `IdentityServiceInterface` 接口，提供有关于使用者身份的服务
type IdentityService struct {
	ServiceInfo *Info
	ServerInfo  *appinit.ServerInfo
}

// 获取当前使用者的身份信息。
//
// 返回：
//   用户身份信息
func (s *IdentityService) GetIdentityInfo() (*common.UserIdentity, error) {
	chaincodeFcn := "getDepartmentIdentity"
	channelReq := channel.Request{
		ChaincodeID: s.ServiceInfo.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{},
	}

	resp, err := s.ServiceInfo.ChannelClient.Query(channelReq)
	if err != nil {
		return nil, GetClassifiedError(chaincodeFcn, err)
	}

	var deptIdentity identity.DepartmentIdentityStored
	err = json.Unmarshal(resp.Payload, &deptIdentity)
	if err != nil {
		return nil, errors.Wrap(err, "无法解析部门身份信息")
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
