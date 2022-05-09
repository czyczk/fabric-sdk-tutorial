package data

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ResourceType 用于标志一个资源的加密类别
type ResourceType int

const (
	// Plain 表示资源加密类别为“明文”。
	Plain ResourceType = iota
	// Encrypted 表示资源加密类别为“加密”。需要经由密钥置换流程来解密。
	Encrypted
	// Offchain 表示资源加密类别为“链下”。数据本身需要经由密钥置换流程来解密。
	Offchain
)

// 序列化时大写首字母以兼容 SCALE codec
var resourceTypeToStringMap = map[ResourceType]string{
	Plain:     "Plain",
	Encrypted: "Encrypted",
	Offchain:  "Offchain",
}

// 反序列化时大小写均接受
var resourceTypeFromStringMap = map[string]ResourceType{
	"plain":     Plain,
	"encrypted": Encrypted,
	"offchain":  Offchain,
}

func (t ResourceType) String() string {
	str, ok := resourceTypeToStringMap[t]
	if ok {
		return str
	}

	return fmt.Sprintf("%d", int(t))
}

// NewResourceTypeFromString 从 enum 名称获得 ResourceType enum。
func NewResourceTypeFromString(enumString string) (ret ResourceType, err error) {
	// 不要区分大小写
	enumStringCaseInsensitive := strings.ToLower(enumString)

	ret, ok := resourceTypeFromStringMap[enumStringCaseInsensitive]
	if !ok {
		err = fmt.Errorf("不正确的 enum 字符串 '%v'", enumString)
		return
	}

	return
}

// MarshalJSON marshals the enum as a quoted JSON string
func (t ResourceType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// UnmarshalJSON unmarshals a quoted JSON string to the enum value
func (t *ResourceType) UnmarshalJSON(b []byte) error {
	var jsonStr string
	err := json.Unmarshal(b, &jsonStr)
	if err != nil {
		return err
	}

	enum, err := NewResourceTypeFromString(jsonStr)
	if err != nil {
		return err
	}

	*t = enum
	return nil
}

// ResMetadata 包含要传入链码的资源的元数据
type ResMetadata struct {
	ResourceType ResourceType           `json:"resourceType"` // 资源加密类别
	ResourceID   string                 `json:"resourceId"`   // 资源 ID
	Hash         string                 `json:"hash"`         // 资源的明文的哈希值（[32]byte 的 Base64 编码）
	Size         uint64                 `json:"size"`         // 资源的明文的内容部分的大小
	Extensions   map[string]interface{} `json:"extensions"`   // 扩展字段（包含可公开的属性）
}

// PlainData 用于表示要传入链码的明文资源
type PlainData struct {
	Metadata ResMetadata `json:"metadata"` // 资源的元数据
	Data     string      `json:"data"`     // 资源的数据本体（Base64 编码）
}

// EncryptedData 用于表示要传入链码的加密资源
type EncryptedData struct {
	Metadata ResMetadata `json:"metadata"` // 资源的元数据
	Data     string      `json:"data"`     // 资源的数据本体（密文）（Base64 编码）
	Key      string      `json:"key"`      // 对称密钥（密文）（Base64 编码）
	Policy   string      `json:"policy"`   // 策略
}

// OffchainData 用于表示要传入链码的链下资源
type OffchainData struct {
	Metadata ResMetadata `json:"metadata"` // 资源的元数据
	CID      string      `json:"cid"`      // 资源在 IPFS 网络上的内容 ID
	Key      string      `json:"key"`      // 对称密钥（密文）（Base64 编码）
	Policy   string      `json:"policy"`   // 策略
}
