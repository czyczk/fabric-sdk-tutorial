package service

import "gitee.com/czyczk/fabric-sdk-tutorial/pkg/errorcode"

// AuthService 用于管理访问权请求。
type AuthService struct {
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
	return "", errorcode.ErrorNotImplemented
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
	return "", errorcode.ErrorNotImplemented
}

// 列出用户可批复但未批复的授权会话 ID。
//
// 参数：
//   每页大小
//   书签（上次访问位置）
//
// 返回：
//   授权会话 ID
func (s *AuthService) ListPendingAuthSessionIDsByResourceCreator(pageSize int, bookmark string) ([]string, error) {
	return nil, errorcode.ErrorNotImplemented
}
