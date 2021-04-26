package data

import "time"

// ResMetadataStored 包含从链码中读出的资源的元数据
type ResMetadataStored struct {
	ResourceType ResourceType      `json:"resourceType"` // 资源加密类别
	ResourceID   string            `json:"resourceID"`   // 资源 ID
	Hash         string            `json:"hash"`         // 资源的明文该有的哈希值（[32]byte 的 Base64 编码）
	Size         uint64            `json:"size"`         // 资源的明文该有的大小
	Extensions   map[string]string `json:"extensions"`   // 扩展字段
	Creator      string            `json:"creator"`      // 资源创建者公钥（Base64 编码）
	Timestamp    time.Time         `json:"timestamp"`    // 时间戳
	HashStored   string            `json:"hashStored"`   // 上传的密文哈希值。明文时应与 `Hash` 有相同值。（[32]byte 的 Base64 编码）
	SizeStored   uint64            `json:"sizeStored"`   // 上传的密文的大小。明文时应与 `Size` 有相同值。
}
