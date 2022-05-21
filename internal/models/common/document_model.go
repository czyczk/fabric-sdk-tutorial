package common

import (
	"encoding/json"
	"fmt"
	"strings"
)

// 用于放置在元数据的 extensions.dataType 中的值
const DocumentDataType = "Document"

// DocumentProperties 表示数字文档的属性部分
type DocumentProperties struct {
	ID                          string       `json:"id"`                          // 数字文档 ID
	Name                        string       `json:"name"`                        // 数字文档名称
	Type                        DocumentType `json:"documentType"`                // 数字文档的文档类型
	PrecedingDocumentID         string       `json:"precedingDocumentId"`         // 数字文档的前置文档 ID
	HeadDocumentID              string       `json:"headDocumentId"`              // 数字文档的头文档 ID
	EntityAssetID               string       `json:"entityAssetId"`               // 数字文档所关联的实体资产的 ID
	IsNamePublic                bool         `json:"isNamePublic"`                // 是否公开标记。用于创建扩展字段。本地数据库中应保留该字段。
	IsTypePublic                bool         `json:"isTypePublic"`                // 是否公开标记。用于创建扩展字段。本地数据库中应保留该字段。
	IsPrecedingDocumentIDPublic bool         `json:"isPrecedingDocumentIdPublic"` // 是否公开标记。用于创建扩展字段。本地数据库中应保留该字段。
	IsHeadDocumentIDPublic      bool         `json:"isHeadDocumentIdPublic"`      // 是否公开标记。用于创建扩展字段。本地数据库中应保留该字段。
	IsEntityAssetIDPublic       bool         `json:"isEntityAssetIdPublic"`       // 是否公开标记。用于创建扩展字段。本地数据库中应保留该字段。
}

// Document 表示数字文档
type Document struct {
	DocumentProperties `mapstructure:",squash"`
	Contents           []byte `json:"contents"` // 数字文档内容
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

// 序列化时大写首字母以兼容 SCALE codec
var documentTypeToStringMap = map[DocumentType]string{
	DesignDocument:     "DesignDocument",
	ProductionDocument: "ProductionDocument",
	TransferDocument:   "TransferDocument",
	UsageDocument:      "UsageDocument",
	RepairDocument:     "RepairDocument",
}

// 反序列化时大小写均接受
var documentTypeFromStringMap = map[string]DocumentType{
	"designdocument":     DesignDocument,
	"productiondocument": ProductionDocument,
	"transferdocument":   TransferDocument,
	"usagedocument":      UsageDocument,
	"repairdocument":     RepairDocument,
}

func (t DocumentType) String() string {
	str, ok := documentTypeToStringMap[t]
	if ok {
		return str
	}

	return fmt.Sprintf("%d", int(t))
}

// NewDocumentTypeFromString 从 enum 名称获得 DocumentType enum。
func NewDocumentTypeFromString(enumString string) (ret DocumentType, err error) {
	// 不要区分大小写
	enumStringCaseInsensitive := strings.ToLower(enumString)

	ret, ok := documentTypeFromStringMap[enumStringCaseInsensitive]
	if !ok {
		err = fmt.Errorf("不正确的 enum 字符串 '%v'", enumString)
		return
	}

	return
}

// MarshalJSON marshals the enum as a quoted JSON string
func (t DocumentType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// UnmarshalJSON unmarshals a quoted JSON string to the enum value
func (t *DocumentType) UnmarshalJSON(b []byte) error {
	var jsonStr string
	err := json.Unmarshal(b, &jsonStr)
	if err != nil {
		return err
	}

	enum, err := NewDocumentTypeFromString(jsonStr)
	if err != nil {
		return err
	}

	*t = enum
	return nil
}
