package query

// IDsWithPagination 结构体用于封装 ID 列表和书签，其中 ID 列表可用于资源 ID、授权会话 ID 等。
type IDsWithPagination struct {
	IDs      []string `json:"ids"`      // ID 列表
	Bookmark string   `json:"bookmark"` // 标识分页终点的书签
}
