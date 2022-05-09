package common

// UserIdentity 表示当前用户的身份信息，包含部门信息
type UserIdentity struct {
	UserID        string `json:"userId"`        // 用户 ID
	OrgName       string `json:"orgName"`       // 组织名称
	DeptType      string `json:"deptType"`      // 部门类型
	DeptLevel     int    `json:"deptLevel"`     // 部门级别
	DeptName      string `json:"deptName"`      // 部门名称
	SuperDeptName string `json:"superDeptName"` // 上级部门名称
}
