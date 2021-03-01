package controller

// ResourceCreationInfo 包含资源成功创建时该返回给客户端的信息
type ResourceCreationInfo struct {
	ResourceID    string `json:"resourceID"`
	TransactionID string `json:"transactionID"`
}
