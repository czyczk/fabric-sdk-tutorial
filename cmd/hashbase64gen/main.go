package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// internal/models/common/document_model.go
type DocumentProperties struct {
	ID                          string `yaml:"id"`                          // 数字文档 ID
	Name                        string `yaml:"name"`                        // 数字文档名称
	Type                        string `yaml:"documentType"`                // 数字文档的文档类型
	PrecedingDocumentID         string `yaml:"precedingDocumentId"`         // 数字文档的前置文档 ID
	HeadDocumentID              string `yaml:"headDocumentId"`              // 数字文档的头文档 ID
	EntityAssetID               string `yaml:"entityAssetId"`               // 数字文档所关联的实体资产的 ID
	IsNamePublic                bool   `yaml:"isNamePublic"`                // 是否公开标记。用于创建扩展字段。本地数据库中应保留该字段。
	IsTypePublic                bool   `yaml:"isTypePublic"`                // 是否公开标记。用于创建扩展字段。本地数据库中应保留该字段。
	IsPrecedingDocumentIDPublic bool   `yaml:"isPrecedingDocumentIdPublic"` // 是否公开标记。用于创建扩展字段。本地数据库中应保留该字段。
	IsHeadDocumentIDPublic      bool   `yaml:"isHeadDocumentIdPublic"`      // 是否公开标记。用于创建扩展字段。本地数据库中应保留该字段。
	IsEntityAssetIDPublic       bool   `yaml:"isEntityAssetIdPublic"`       // 是否公开标记。用于创建扩展字段。本地数据库中应保留该字段。
}

type Document struct {
	DocumentProperties `mapstructure:",squash"`
	Contents           []byte `json:"contents"`
}

// internal/service/document_service.go
func calculateHashBase64(document *Document) (string, error) {
	if document == nil {
		return "", fmt.Errorf("文档对象不能为 nil")
	}

	if strings.TrimSpace(document.ID) == "" {
		return "", fmt.Errorf("文档 ID 不能为空")
	}

	documentAsMap := make(map[string]interface{})
	err := mapstructure.Decode(document, &documentAsMap)
	if err != nil {
		return "", errors.Wrap(err, "无法序列化文档")
	}

	documentBytes, err := json.Marshal(documentAsMap)
	if err != nil {
		return "", errors.Wrap(err, "无法序列化文档")
	}

	hash := sha256.Sum256(documentBytes)
	hashBase64 := base64.StdEncoding.EncodeToString(hash[:])

	return hashBase64, nil
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run main.go <document_path> <document_properties_path>")
		return
	}

	documentPath := os.Args[1]
	documentPropertiesPath := os.Args[2]

	contents, err := os.ReadFile(documentPath)
	if err != nil {
		fmt.Printf("无法读取文档：%v\n", err)
		return
	}

	documentPropertiesData, err := os.ReadFile(documentPropertiesPath)
	if err != nil {
		fmt.Printf("无法读取文档属性：%v\n", err)
		return
	}

	var documentProperties DocumentProperties
	err = yaml.Unmarshal(documentPropertiesData, &documentProperties)
	if err != nil {
		fmt.Printf("无法解析文档属性：%v\n", err)
		return
	}

	document := &Document{
		DocumentProperties: documentProperties,
		Contents:           contents,
	}

	hashBase64, err := calculateHashBase64(document)
	if err != nil {
		fmt.Printf("无法计算文档哈希：%v\n", err)
		return
	}

	// $ go run cmd/hashbase64gen/main.go cmd/hashbase64gen/document.txt cmd/hashbase64gen/document-properties.yaml
	// JV8Dx0kVg4bhzjsOm02BEvUmFgZT+whS0kRb1UdmKik=
	fmt.Println(hashBase64)
}
