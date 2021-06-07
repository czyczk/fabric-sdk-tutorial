package service

import (
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/models/common"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/query"
)

// AuthServiceInterface 定义了用于管理访问权请求的服务的接口。
type AuthServiceInterface interface {
	// 创建访问权申请。
	//
	// 参数：
	//   资源 ID
	//   理由
	//
	// 返回：
	//   交易 ID
	CreateAuthRequest(resourceID string, reason string) (string, error)

	// 创建访问权批复。
	//
	// 参数：
	//   授权会话 ID
	//   批复结果
	//
	// 返回：
	//   交易 ID
	CreateAuthResponse(authSessionID string, result bool) (string, error)

	// 获取访问权会话。
	//
	// 参数：
	//   授权会话 ID
	//
	// 返回：
	//   授权会话
	GetAuthSession(authSessionID string) (*common.AuthSession, error)

	// 列出用户可批复但未批复的授权会话 ID。
	//
	// 参数：
	//   每页大小
	//   书签（上次访问位置）
	//
	// 返回：
	//   带分页书签的授权会话 ID 列表
	ListPendingAuthSessionIDsByResourceCreator(pageSize int, bookmark string) (*query.IDsWithPagination, error)

	// 列出当前用户申请的授权会话 ID。
	//
	// 参数：
	//   每页大小
	//   书签（上次访问位置）
	//   最新置于最前
	//
	// 返回：
	//   带分页书签的授权会话 ID 列表
	ListAuthSessionIDsByRequestor(pageSize int, bookmark string, isLatestFirst bool) (*query.IDsWithPagination, error)
}
