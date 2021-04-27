package common

// EntityAsset 表示实体资产
type EntityAsset struct {
	ID           string   `json:"id"`           // 实体资产 ID
	Name         string   `json:"name"`         // 实体资产名称
	ComponentIDs []string `json:"componentIDs"` // 组件的序列号数组
	Property     string   `json:"property"`     // 所有额外扩展字段。以 JSON 表示。
}

// Document 表示数字文档
type Document struct {
	ID       string `json:"id"`       // 数字文档 ID
	Name     string `json:"name"`     // 数字文档名称
	Contents []byte `json:"contents"` // 数字文档内容
	Property string `json:"property"` // 所有额外扩展字段。以 JSON 表示。
}

// UserIdentity 表示当前用户的身份信息，包含部门信息
type UserIdentity struct {
	UserID        string `json:"userID"`        // 用户 ID
	OrgName       string `json:"orgName"`       // 组织名称
	DeptType      string `json:"deptType"`      // 部门类型
	DeptLevel     int    `json:"deptLevel"`     // 部门级别
	DeptName      string `json:"deptName"`      // 部门名称
	SuperDeptName string `json:"superDeptName"` // 上级部门名称
}
