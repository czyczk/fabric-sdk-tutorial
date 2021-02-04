package regkey

import "time"

// RegulatorKeyStored 表示从链码中读出的监管者公钥信息
type RegulatorKeyStored struct {
	Key       []byte    `json:"key"`       // 监管者的公钥
	Timestamp time.Time `json:"timestamp"` // 监管者更新为该公钥的时间
}
