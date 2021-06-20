package common

import "fmt"

// DocumentProperties 表示数字文档的属性部分
type DocumentProperties struct {
	ID                          string       `json:"id"`                          // 数字文档 ID
	Name                        string       `json:"name"`                        // 数字文档名称
	Type                        DocumentType `json:"documentType"`                // 数字文档的文档类型
	PrecedingDocumentID         string       `json:"precedingDocumentID"`         // 数字文档的前置文档 ID
	HeadDocumentID              string       `json:"headDocumentID"`              // 数字文档的头文档 ID
	EntityAssetID               string       `json:"entityAssetID"`               // 数字文档所关联的实体资产的 ID
	IsNamePublic                bool         `json:"isNamePublic"`                // 是否公开标记。用于创建扩展字段。本地数据库中应保留该字段。
	IsTypePublic                bool         `json:"isTypePublic"`                // 是否公开标记。用于创建扩展字段。本地数据库中应保留该字段。
	IsPrecedingDocumentIDPublic bool         `json:"isPrecedingDocumentIDPublic"` // 是否公开标记。用于创建扩展字段。本地数据库中应保留该字段。
	IsHeadDocumentIDPublic      bool         `json:"isHeadDocumentIDPublic"`      // 是否公开标记。用于创建扩展字段。本地数据库中应保留该字段。
	IsEntityAssetIDPublic       bool         `json:"isEntityAssetIDPublic"`       // 是否公开标记。用于创建扩展字段。本地数据库中应保留该字段。
}

// Document 表示数字文档
type Document struct {
	DocumentProperties
	Contents []byte `json:"contents"` // 数字文档内容
}

// DocumentType 表示数字文档的文档类型
type DocumentType int

const (
	// DesignDocument 表示设计文档
	DesignDocument DocumentType = iota
	// ProductionDocument 表示生产文档
	ProductionDocument
	// TransferDocument 表示转移文档
	TransferDocument
	// UsageDocument 表示使用文档
	UsageDocument
	// RepairDocument 表示维修文档
	RepairDocument
)

func (t DocumentType) String() string {
	switch t {
	case DesignDocument:
		return "designDocument"
	case ProductionDocument:
		return "productionDocument"
	case TransferDocument:
		return "transferDocument"
	case UsageDocument:
		return "usageDocument"
	case RepairDocument:
		return "repairDocument"
	default:
		return fmt.Sprintf("%d", int(t))
	}
}

// NewDocumentTypeFromString 从 enum 名称获得 DocumentType enum。
func NewDocumentTypeFromString(enumString string) (ret DocumentType, err error) {
	switch enumString {
	case "designDocument":
		ret = DesignDocument
		return
	case "productionDocument":
		ret = ProductionDocument
		return
	case "transferDocument":
		ret = TransferDocument
		return
	case "usageDocument":
		ret = UsageDocument
		return
	case "repairDocument":
		ret = RepairDocument
		return
	default:
		err = fmt.Errorf("不正确的 enum 字符串")
		return
	}
}
