package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/auth"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/query"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/pkg/errors"
)

// AuthService 用于管理访问权请求。
type AuthService struct {
	ServiceInfo *Info
}

// 创建访问权申请。
//
// 参数：
//   资源 ID
//   理由
//
// 返回：
//   交易 ID
func (s *AuthService) CreateAuthRequest(resourceID string, reason string) (string, error) {
	// 组装 Extensions 结构
	extensions := make(map[string]string)
	extensions["reason"] = reason
	extensionsBytes, err := json.Marshal(extensions)
	if err != nil {
		return "", errors.Wrap(err, "无法序列化扩展字段")
	}

	// 组装 AuthRequest 结构并调用链码
	authRequest := auth.AuthRequest{
		ResourceID: resourceID,
		Extensions: string(extensionsBytes),
	}
	authRequestBytes, err := json.Marshal(authRequest)
	if err != nil {
		return "", errors.Wrap(err, "无法序列化链码参数")
	}

	chaincodeFcn := "createAuthRequest"
	channelReq := channel.Request{
		ChaincodeID: s.ServiceInfo.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{authRequestBytes},
	}

	resp, err := s.ServiceInfo.ChannelClient.Execute(channelReq)
	if err != nil {
		return "", GetClassifiedError(chaincodeFcn, err)
	} else {
		return string(resp.TransactionID), nil
	}
}

// 创建访问权批复。
//
// 参数：
//   授权会话 ID
//   批复结果
//
// 返回：
//   交易 ID
func (s *AuthService) CreateAuthResponse(authSessionID string, result bool) (string, error) {
	// Defensive check: authSessionID 不能为空
	if strings.TrimSpace(authSessionID) == "" {
		return "", fmt.Errorf("授权会话 ID 不能为空")
	}

	// 组装 AuthResponse 结构并调用链码
	authResponse := auth.AuthResponse{
		AuthSessionID: authSessionID,
		Result:        result,
	}
	authResponseBytes, err := json.Marshal(authResponse)
	if err != nil {
		return "", errors.Wrap(err, "无法序列化链码参数")
	}

	chaincodeFcn := "createAuthResponse"
	channelReq := channel.Request{
		ChaincodeID: s.ServiceInfo.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{authResponseBytes},
	}

	resp, err := s.ServiceInfo.ChannelClient.Execute(channelReq)
	if err != nil {
		return "", GetClassifiedError(chaincodeFcn, err)
	} else {
		return string(resp.TransactionID), nil
	}
}

// 列出用户可批复但未批复的授权会话 ID。
//
// 参数：
//   每页大小
//   书签（上次访问位置）
//
// 返回：
//   带分页书签的授权会话 ID 列表
func (s *AuthService) ListPendingAuthSessionIDsByResourceCreator(pageSize int, bookmark string) (*query.ResourceIDsWithPagination, error) {
	// 调用链码函数 listPendingAuthSessionIDsByResourceCreator
	chaincodeFcn := "listPendingAuthSessionIDsByResourceCreator"
	channelReq := channel.Request{
		ChaincodeID: s.ServiceInfo.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{[]byte(strconv.Itoa(pageSize)), []byte(bookmark)},
	}

	resp, err := s.ServiceInfo.ChannelClient.Query(channelReq)
	if err != nil {
		return nil, GetClassifiedError(chaincodeFcn, err)
	}

	// 解析结果列表
	var result query.ResourceIDsWithPagination
	err = json.Unmarshal(resp.Payload, &result)
	if err != nil {
		return nil, errors.Wrap(err, "无法解析结果列表")
	}

	return &result, nil
}
