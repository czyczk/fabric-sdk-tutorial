package auth

import "time"

// AuthRequestStored 表示从链码得到的访问权申请请求
type AuthRequestStored struct {
	AuthSessionID string    `json:"authSessionID"` // 授权会话 ID
	ResourceID    string    `json:"resourceID"`    // 资源 ID
	Extensions    string    `json:"extensions"`    // 扩展字段。以 JSON 形式表示。
	Creator       string    `json:"creator"`       // 访问权申请者公钥（Base64 编码）
	Timestamp     time.Time `json:"timestamp"`     // 时间戳
}

// AuthResponseStored 表示从链码得到的访问申请批复
type AuthResponseStored struct {
	AuthSessionID string    `json:"authSessionID"` // 访问权申请会话 ID
	Result        bool      `json:"result"`        // 访问权批复结果
	Creator       string    `json:"creator"`       // 访问权批复者公钥（Base64 编码）
	Timestamp     time.Time `json:"timestamp"`     // 时间戳
}
