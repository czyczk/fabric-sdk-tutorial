package auth

// AuthRequest 表示传给链码的访问权申请请求
type AuthRequest struct {
	ResourceID string                 `json:"resourceId"` // 资源 ID
	Extensions map[string]interface{} `json:"extensions"` // 扩展字段
}

// AuthResponse 表示传给链码的访问申请批复
type AuthResponse struct {
	AuthSessionID string                 `json:"authSessionId"` // 访问权申请会话 ID
	Result        bool                   `json:"result"`        // 访问权批复结果
	Extensions    map[string]interface{} `json:"extensions"`    // 扩展字段
}
