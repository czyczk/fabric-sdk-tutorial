package common

import (
	"encoding/json"
	"fmt"
	"time"
)

// AuthSession 表示一个授权会话，包含会话 ID、理由、批复结果等信息。
type AuthSession struct {
	AuthSessionID     string            `json:"authSessionID"`               // 授权会话 ID
	ResourceID        string            `json:"resourceID"`                  // 申请的资源 ID
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

var authSessionStatusToStringMap = map[AuthSessionStatus]string{
	Pending:  "pending",
	Approved: "approved",
	Rejected: "rejected",
}

var authSessionStatusFromStringMap = map[string]AuthSessionStatus{
	"pending":  Pending,
	"approved": Approved,
	"rejected": Rejected,
}

func (t AuthSessionStatus) String() string {
	str, ok := authSessionStatusToStringMap[t]
	if ok {
		return str
	}

	return fmt.Sprintf("%d", int(t))
}

// NewAuthSessionStatusFromString 从 enum 名称获得 AuthSessionStatus enum。
func NewAuthSessionStatusFromString(enumString string) (ret AuthSessionStatus, err error) {
	ret, ok := authSessionStatusFromStringMap[enumString]
	if !ok {
		err = fmt.Errorf("不正确的 enum 字符串")
		return
	}

	return
}

// MarshalJSON marshals the enum as a quoted JSON string
func (s AuthSessionStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// UnmarshalJSON unmarshals a quoted JSON string to the enum value
func (s *AuthSessionStatus) UnmarshalJSON(b []byte) error {
	var jsonStr string
	err := json.Unmarshal(b, &jsonStr)
	if err != nil {
		return err
	}

	enum, err := NewAuthSessionStatusFromString(jsonStr)
	if err != nil {
		return err
	}

	*s = enum
	return nil
}
