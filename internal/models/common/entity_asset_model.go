package common

// EntityAsset 表示实体资产
type EntityAsset struct {
	ID                       string   `json:"id"`                       // 实体资产 ID
	Name                     string   `json:"name"`                     // 实体资产名称
	DesignDocumentID         string   `json:"designDocumentId"`         // 实体资产的设计文档的 ID
	IsNamePublic             bool     `json:"isNamePublic"`             // 是否公开标记。用于创建扩展字段。本地数据库中应保留该字段。
	IsDesignDocumentIDPublic bool     `json:"isDesignDocumentIdPublic"` // 是否公开标记。用于创建扩展字段。本地数据库中应保留该字段。
	ComponentIDs             []string `json:"componentIds"`             // 组件的序列号数组
}
