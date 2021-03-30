package query

type ResourceIDsWithPagination struct {
	ResourceIDs []string `json:"resourceIDs"` // 资源 ID 列表
	Bookmark    string   `json:"bookmark"`    // 标识分页终点的书签
}
