package service

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

	// 列出用户可批复但未批复的授权会话 ID。
	//
	// 参数：
	//   每页大小
	//   书签（上次访问位置）
	//
	// 返回：
	//   授权会话 ID
	ListPendingAuthSessionIDsByResourceCreator(pageSize int, bookmark string) ([]string, error)
}
