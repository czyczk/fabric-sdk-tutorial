package identity

// DepartmentIdentityStored 表示由链码返回的部门身份信息
type DepartmentIdentityStored struct {
	DeptType      string `json:"deptType"`      // 部门类型
	DeptLevel     int    `json:"deptLevel"`     // 部门级别
	DeptName      string `json:"deptName"`      // 部门名称
	SuperDeptName string `json:"superDeptName"` // 上级部门名称
}
