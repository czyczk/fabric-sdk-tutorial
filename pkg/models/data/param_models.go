package data

import "fmt"

// ResourceType 用于标志一个资源的加密类别
type ResourceType int

const (
	// Plain 表示资源加密类别为“明文”。
	Plain ResourceType = iota
	// Encrypted 表示资源加密类别为“加密”。需要经由密钥置换流程来解密。
	Encrypted
	// Offchain 表示资源加密类别为“链下”。数据本身需要经由密钥置换流程来解密。
	Offchain
	// RegulatorEncrypted 表示资源加密类别为“由监管者公钥加密”。只由监管者使用。
	RegulatorEncrypted
)

func (t ResourceType) String() string {
	switch t {
	case Plain:
		return "Plain"
	case Encrypted:
		return "Encrypted"
	case Offchain:
		return "Offchain"
	case RegulatorEncrypted:
		return "RegulatorEncrypted"
	default:
		return fmt.Sprintf("%d", int(t))
	}
}

// NewResourceTypeFromString 从 enum 名称获得 ResourceType enum。
func NewResourceTypeFromString(enumString string) (ret ResourceType, err error) {
	switch enumString {
	case "Plain":
		ret = Plain
		return
	case "Encrypted":
		ret = Encrypted
		return
	case "Offchain":
		ret = Offchain
		return
	case "RegulatorEncrypted":
		ret = RegulatorEncrypted
		return
	default:
		err = fmt.Errorf("不正确的 enum 字符串")
		return
	}
}

// ResMetadata 包含要传入链码的资源的元数据
type ResMetadata struct {
	ResourceType ResourceType `json:"resourceType"` // 资源加密类别
	ResourceID   string       `json:"resourceID"`   // 资源 ID
	Hash         string       `json:"hash"`         // 资源的明文该有的哈希值（SHA256）（[32]byte 的 Base64 编码）
	Size         uint64       `json:"size"`         // 资源的明文该有的大小
	Extensions   string       `json:"extensions"`   // 扩展字段。以 JSON 形式表示。
}

// PlainData 用于表示要传入链码的明文资源
type PlainData struct {
	Metadata ResMetadata `json:"metadata"` // 资源的元数据
	Data     string      `json:"data"`     // 资源数据本体
}

// EncryptedData 用于表示要传入链码的加密资源
type EncryptedData struct {
	Metadata ResMetadata `json:"metadata"` // 资源的元数据
	Data     string      `json:"data"`     // 资源数据本体
	Key      string      `json:"key"`      // 加密的对称密钥（Base64 编码）
	Policy   string      `json:"policy"`   // 策略
}

// OffchainData 用于表示要传入链码的链下资源
type OffchainData struct {
	Metadata ResMetadata `json:"metadata"` // 资源的元数据
	Key      string      `json:"key"`      // 加密的对称密钥（Base64 编码）
	Policy   string      `json:"policy"`   // 策略
}
