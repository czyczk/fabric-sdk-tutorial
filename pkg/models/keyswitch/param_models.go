package keyswitch

// KeySwitchTrigger 表示要传给链码的密文资源访问请求
type KeySwitchTrigger struct {
	ResourceID    string   `json:"resourceID"`    // 资源 ID
	AuthSessionID string   `json:"authSessionID"` // 授权会话 ID。为零值时可忽略。
	KeySwitchPK   [64]byte `json:"keySwitchPK"`   // 访问申请者用于密钥置换的公钥
}

// KeySwitchResult 表示要传给链码的密钥置换结果
type KeySwitchResult struct {
	KeySwitchSessionID string   `json:"keySwitchSessionID"` // 密钥置换会话 ID
	Share              [64]byte `json:"share"`              // 个人份额
}

// KeySwitchResultQuery 表示密钥置换的查询请求
type KeySwitchResultQuery struct {
	KeySwitchSessionID string `json:"keySwitchSessionID"` // 密钥置换会话 ID
	ResultCreator      string `json:"resultCreator"`      // 密钥置换结果的创建者公钥（Base 64 编码）
}
