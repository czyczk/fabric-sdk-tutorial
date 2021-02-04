package service

// KeySwitchServiceInterface 定义了有关于密钥置换的服务的接口
type KeySwitchServiceInterface interface {
	// 创建密文访问申请/密钥置换触发器。
	//
	// 参数：
	//   资源 ID
	//   授权会话 ID
	//
	// 返回：
	//   交易 ID
	CreateKeySwitchTrigger(resourceID string, authSessionID string) (string, error)

	// 等待并收集密钥置换结果。
	//
	// 参数：
	//   密钥置换会话 ID
	//   预期的份额个数
	//   超时时限（可选）
	//
	// 返回：
	//   预期个数的份额列表
	AwaitKeySwitchResults(keySwitchSessionID string, numExpected int, timeout ...int) ([][]byte, error)

	// 创建密钥置换结果。
	//
	// 参数：
	//   密钥置换会话 ID
	//   个人份额
	//
	// 返回：
	//   交易 ID
	CreateKeySwitchResult(keySwitchSessionID string, share []byte) (string, error)

	// 获取集合权威公钥。
	//
	// 返回：
	//   集合权威公钥
	GetCollectiveAuthorityPublicKey() ([]byte, error)
}
