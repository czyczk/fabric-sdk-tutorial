package controller

import "gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/bcao"

// ResourceCreationInfo 包含资源成功创建时应该返回给客户端的信息
type ResourceCreationInfo struct {
	ResourceID           string `json:"resourceId"`                     // 资源 ID
	SymmetricKeyMaterial string `json:"symmetricKeyMaterial,omitempty"` // 用于生成加密该资源的对称密钥的原材料（SM2 公钥 PEM）（可选）
	*bcao.TransactionCreationInfo
}

// AuthSessionCreationInfo 包含授权会话成功创建时应该返回给客户端的信息
type AuthSessionCreationInfo struct {
	AuthSessionID string `json:"authSessionId"`
	*bcao.TransactionCreationInfo
}

// KeySwitchSessionCreationInfo 包含密钥置换会话成功创建时应该返回给客户端的信息
type KeySwitchSessionCreationInfo struct {
	KeySwitchSessionID string `json:"keySwitchSessionId"`
	*bcao.TransactionCreationInfo
}

func NewAuthSessionCreationInfoFromTransactionCreationInfo(txCreationInfo *bcao.TransactionCreationInfoWithManualID) *AuthSessionCreationInfo {
	return &AuthSessionCreationInfo{
		AuthSessionID:           txCreationInfo.ManualID,
		TransactionCreationInfo: txCreationInfo.TransactionCreationInfo,
	}
}

func NewKeySwitchSessionCreationInfoFromTransactionCreationInfo(txCreationInfo *bcao.TransactionCreationInfoWithManualID) *KeySwitchSessionCreationInfo {
	return &KeySwitchSessionCreationInfo{
		KeySwitchSessionID:      txCreationInfo.ManualID,
		TransactionCreationInfo: txCreationInfo.TransactionCreationInfo,
	}
}
