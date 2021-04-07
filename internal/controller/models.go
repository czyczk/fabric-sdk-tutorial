package controller

// ResourceCreationInfo 包含资源成功创建时该返回给客户端的信息
type ResourceCreationInfo struct {
	ResourceID           string `json:"resourceID"`                     // 资源 ID
	TransactionID        string `json:"transactionID"`                  // 交易 ID
	SymmetricKeyMaterial string `json:"symmetricKeyMaterial,omitempty"` // 用于生成加密该资源的对称密钥的原材料（SM2 公钥 PEM）（可选）
}
