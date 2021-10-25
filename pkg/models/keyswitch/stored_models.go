package keyswitch

import "time"

// KeySwitchTriggerStored 表示从链码得到的密文资源访问请求
type KeySwitchTriggerStored struct {
	KeySwitchSessionID string    `json:"keySwitchSessionID"` // 密钥置换会话 ID
	ResourceID         string    `json:"resourceID"`         // 资源 ID
	AuthSessionID      string    `json:"authSessionID"`      // 授权会话 ID。为零值时可忽略。
	Creator            string    `json:"creator"`            // 访问申请者公钥（Base64 编码）
	KeySwitchPK        string    `json:"keySwitchPK"`        // 访问申请者用于密钥置换的公钥（[64]byte 的 Base64 编码）
	Timestamp          time.Time `json:"timestamp"`          // 时间戳
	ValidationResult   bool      `json:"validationResult"`   // 访问申请是否通过验证
}

// KeySwitchResultStored 表示从链码得到的密钥置换结果
type KeySwitchResultStored struct {
	KeySwitchSessionID string    `json:"keySwitchSessionID"` // 密钥置换会话 ID
	Share              string    `json:"share"`              // 个人份额（[64]byte 的 Base64 编码）
	ZKProof            string    `json:"zkproof"`            // 零知识证明（[96]byte 的 Base64 编码），用于验证份额
	KeySwitchPK        string    `json:"keySwitchPK"`        // 份额生成者的密钥置换公钥（[64]byte 的 Base64 编码），用于验证份额
	Creator            string    `json:"creator"`            // 密钥置换响应者的公钥（Base64 编码）
	Timestamp          time.Time `json:"timestamp"`          // 时间戳
}
