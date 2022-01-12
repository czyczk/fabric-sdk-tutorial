package service

import (
	"fmt"
	"strings"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/bcao"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/models/common"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/errorcode"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/auth"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/query"
	"github.com/pkg/errors"
)

// AuthService 用于管理访问权请求。
type AuthService struct {
	AuthBCAO bcao.IAuthBCAO
}

// 创建访问权申请。
//
// 参数：
//   资源 ID
//   理由
//
// 返回：
//   交易 ID
func (s *AuthService) CreateAuthRequest(resourceID string, reason string) (string, error) {
	// 组装 Extensions 结构
	extensions := make(map[string]interface{})
	extensions["dataType"] = "authRequest"
	if reason != "" {
		extensions["reason"] = reason
	}

	// 组装 AuthRequest 结构并调用链码
	authRequest := auth.AuthRequest{
		ResourceID: resourceID,
		Extensions: extensions,
	}

	return s.AuthBCAO.CreateAuthRequest(&authRequest)
}

// 创建访问权批复。
//
// 参数：
//   授权会话 ID
//   批复结果
//
// 返回：
//   交易 ID
func (s *AuthService) CreateAuthResponse(authSessionID string, result bool) (string, error) {
	// Defensive check: authSessionID 不能为空
	if strings.TrimSpace(authSessionID) == "" {
		return "", fmt.Errorf("授权会话 ID 不能为空")
	}

	// 组装 Extensions 结构
	extensions := make(map[string]interface{})
	extensions["dataType"] = "authResponse"

	// 组装 AuthResponse 结构并调用链码
	authResponse := auth.AuthResponse{
		AuthSessionID: authSessionID,
		Result:        result,
		Extensions:    extensions,
	}

	return s.AuthBCAO.CreateAuthResponse(&authResponse)
}

// 获取访问权会话。
//
// 参数：
//   授权会话 ID
//
// 返回：
//   授权会话
func (s *AuthService) GetAuthSession(authSessionID string) (*common.AuthSession, error) {
	// 获取授权申请
	authRequestStored, err := s.AuthBCAO.GetAuthRequest(authSessionID)
	if err != nil {
		return nil, err
	}

	// 装填一部分结果
	result := &common.AuthSession{
		AuthSessionID:     authSessionID,
		ResourceID:        authRequestStored.ResourceID,
		Reason:            authRequestStored.Extensions["reason"].(string),
		Requestor:         authRequestStored.Creator,
		RequestTimestsamp: authRequestStored.Timestamp,
	}

	// 获取授权批复
	authResponseStored, err := s.AuthBCAO.GetAuthResponse(authSessionID)

	if err == nil {
		// 可以从链码中获得该会话的批复，则将批复结果附加在结果中
		if authResponseStored.Result {
			result.Status = common.Approved
		} else {
			result.Status = common.Rejected
		}

		result.Responder = &authResponseStored.Creator
		result.ResponseTimestamp = &authResponseStored.Timestamp
	} else if err == errorcode.ErrorNotFound {
		// 如果从链码中找不到该会话的批复，则记录该会话的状态为未批复
		result.Status = common.Pending
	} else {
		// 从链码中获取批复时出错
		return nil, errors.Wrap(err, "无法获取授权会话批复情况")
	}

	return result, nil
}

// 列出用户可批复但未批复的授权会话 ID。
//
// 参数：
//   每页大小
//   书签（上次访问位置）
//
// 返回：
//   带分页书签的授权会话 ID 列表
func (s *AuthService) ListPendingAuthSessionIDsByResourceCreator(pageSize int, bookmark string) (*query.IDsWithPagination, error) {
	// 调用链码函数 listPendingAuthSessionIDsByResourceCreator
	return s.AuthBCAO.ListPendingAuthSessionIDsByResourceCreator(pageSize, bookmark)
}

// 列出当前用户申请的授权会话 ID。
//
// 参数：
//   每页大小
//   书签（上次访问位置）
//   最新置于最前
//
// 返回：
//   带分页书签的授权会话 ID 列表
func (s *AuthService) ListAuthSessionIDsByRequestor(pageSize int, bookmark string, isLatestFirst bool) (*query.IDsWithPagination, error) {
	// 调用链码函数 listAuthSessionIDsByRequestor
	return s.AuthBCAO.ListAuthSessionIDsByRequestor(pageSize, bookmark, isLatestFirst)
}
