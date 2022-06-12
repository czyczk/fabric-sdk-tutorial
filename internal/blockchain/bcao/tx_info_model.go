package bcao

// TransactionCreationInfo 包含交易成功创建时应该返回的信息
type TransactionCreationInfo struct {
	TransactionID string `json:"transactionId"`     // 交易 ID
	BlockID       string `json:"blockId,omitempty"` // 区块 ID
}

type TransactionCreationInfoWithManualID struct {
	ManualID string
	*TransactionCreationInfo
}
