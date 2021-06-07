package common

import (
	"fmt"
	"time"
)

// AuthSession 表示一个授权会话，包含会话 ID、理由、批复结果等信息。
type AuthSession struct {
	AuthSessionID     string            `json:"authSessionID"`               // 授权会话 ID
	Reason            string            `json:"reason"`                      // 申请理由
	Status            AuthSessionStatus `json:"status"`                      // 会话状态
	Requestor         string            `json:"requestor"`                   // 申请者的公钥（Base64 编码）
	Responder         *string           `json:"responder,omitempty"`         // 批复者的公钥（Base64 编码）
	RequestTimestsamp time.Time         `json:"requestTimestamp"`            // 申请时间
	ResponseTimestamp *time.Time        `json:"responseTimestamp,omitempty"` // 批复时间
}

type AuthSessionStatus int

const (
	// Pending 表示该授权会话尚未被处理
	Pending AuthSessionStatus = iota
	// Approved 表示该授权已被批准
	Approved
	// Rejected 表示该授权未被批准
	Rejected
)

func (t AuthSessionStatus) String() string {
	switch t {
	case Pending:
		return "pending"
	case Approved:
		return "approved"
	case Rejected:
		return "rejected"
	default:
		return fmt.Sprintf("%d", int(t))
	}
}

// NewAuthSessionStatusFromString 从 enum 名称获得 AuthSessionStatus enum。
func NewAuthSessionStatusFromString(enumString string) (ret AuthSessionStatus, err error) {
	switch enumString {
	case "pending":
		ret = Pending
		return
	case "approved":
		ret = Approved
		return
	case "rejected":
		ret = Rejected
		return
	default:
		err = fmt.Errorf("不正确的 enum 字符串")
		return
	}
}
