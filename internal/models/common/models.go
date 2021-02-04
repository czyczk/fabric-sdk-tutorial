package common

// EntityAsset 表示实体资产
type EntityAsset struct {
	ID           string   // 实体资产 ID
	Name         string   // 实体资产名称
	ComponentIDs []string // 组件的序列号数组
	Property     string   // 所有额外扩展字段。以 JSON 表示。
}

// Document 表示数字文档
type Document struct {
	ID       string // 数字文档 ID
	Name     string // 数字文档名称
	Contents []byte // 数字文档内容
	Property string // 所有额外扩展字段。以 JSON 表示。
}
