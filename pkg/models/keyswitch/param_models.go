package keyswitch

// KeySwitchTrigger 表示要传给链码的密文资源访问请求
type KeySwitchTrigger struct {
	ResourceID    string `json:"resourceId"`    // 资源 ID
	AuthSessionID string `json:"authSessionId"` // 授权会话 ID。为零值时可忽略。
	KeySwitchPK   string `json:"keySwitchPk"`   // 访问申请者用于密钥置换的公钥（[64]byte 的 Base64 编码）
}

// KeySwitchResult 表示要传给链码的密钥置换结果
type KeySwitchResult struct {
	KeySwitchSessionID string `json:"keySwitchSessionId"` // 密钥置换会话 ID
	Share              string `json:"share"`              // 个人份额（[64]byte 的 Base64 编码）
	ZKProof            string `json:"zkproof"`            // 零知识证明（[96]byte 的 Base64 编码），用于验证份额
	KeySwitchPK        string `json:"keySwitchPk"`        // 份额生成者的密钥置换公钥（[64]byte 的 Base64 编码），用于验证份额
}

// KeySwitchResultQuery 表示密钥置换的查询请求
type KeySwitchResultQuery struct {
	KeySwitchSessionID string `json:"keySwitchSessionId"` // 密钥置换会话 ID
	ResultCreator      string `json:"resultCreator"`      // 密钥置换结果的创建者公钥（Base 64 编码）
}
