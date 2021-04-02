package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/global"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/errorcode"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/keyswitch"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/pkg/errors"
)

// KeySwitchService 实现了 `KeySwitchServiceInterface` 接口，提供有关于密钥置换的服务
type KeySwitchService struct {
	ServiceInfo *Info
}

// 创建密文访问申请/密钥置换触发器。
//
// 参数：
//   资源 ID
//   授权会话 ID
//
// 返回：
//   交易 ID（亦即密钥置换会话 ID）
func (s *KeySwitchService) CreateKeySwitchTrigger(resourceID string, authSessionID string) (string, error) {
	if strings.TrimSpace(resourceID) == "" {
		return "", fmt.Errorf("资源 ID 不能为空")
	}

	// 将公钥序列化为定长字节切片
	ksPubKey := [64]byte{}
	copy(ksPubKey[:32], global.SM2PublicKey.X.Bytes())
	copy(ksPubKey[32:], global.SM2PublicKey.Y.Bytes())

	// 组装一个 KeySwitchTrigger 对象，并调用链码
	ksTrigger := keyswitch.KeySwitchTrigger{
		ResourceID:    resourceID,
		AuthSessionID: authSessionID,
		KeySwitchPK:   ksPubKey,
	}

	ksTriggerBytes, err := json.Marshal(ksTrigger)
	if err != nil {
		return "", errors.Wrapf(err, "无法序列化链码参数")
	}

	chaincodeFcn := "createKeySwitchTrigger"
	channelReq := channel.Request{
		ChaincodeID: s.ServiceInfo.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{ksTriggerBytes},
	}

	resp, err := s.ServiceInfo.ChannelClient.Execute(channelReq)
	if err != nil {
		return "", getClassifiedError(chaincodeFcn, err)
	} else {
		return string(resp.TransactionID), nil
	}
}

// 等待并收集密钥置换结果。
//
// 参数：
//   密钥置换会话 ID
//   预期的份额个数
//   超时时限（可选）
//
// 返回：
//   预期个数的份额列表
func (s *KeySwitchService) AwaitKeySwitchResults(keySwitchSessionID string, numExpected int, timeout ...int) ([][]byte, error) {
	// keySwitchSessionID 不能为空
	if strings.TrimSpace(keySwitchSessionID) == "" {
		return nil, fmt.Errorf("密钥转换会话 ID 不能为空")
	}

	// 解析超时时限参数，只允许指定 1 个，为空时默认 20 秒。
	timeoutInSec := 20
	if len(timeout) > 1 {
		return nil, fmt.Errorf("只可指定 1 个超时时限")
	} else if len(timeout) == 1 {
		timeoutInSec = timeout[0]
	}

	// 尝试监听事件 "ks_${keySwitchSessionID}_result"。若失败则提前返回。
	eventID := "ks_" + keySwitchSessionID + "_result"
	reg, notifier, err := registerEvent(s.ServiceInfo.ChannelClient, s.ServiceInfo.ChaincodeID, eventID)
	if err != nil {
		return nil, errors.Wrapf(err, "无法监听密钥置换结果事件")
	}

	// 接收 channel 内容（一个密钥置换会话 ID），对每个 ID 新开一个 Go routine 调用链码获取一次置换份额，并将结果放入列表。若在时限内未能接收到预期数量的 ID，则等待已发送的 Go routine 结束后，放弃此次结果，并返回超时错误。
	receivedIDs := []string{}
	var wg sync.WaitGroup
	for {
		select {
			case 
		}
	}
}

// 创建密钥置换结果。
//
// 参数：
//   密钥置换会话 ID
//   个人份额
//
// 返回：
//   交易 ID
func (s *KeySwitchService) CreateKeySwitchResult(keySwitchSessionID string, share []byte) (string, error) {
	return "", errorcode.ErrorNotImplemented
}

// 获取集合权威公钥。
//
// 返回：
//   集合权威公钥
func (s *KeySwitchService) GetCollectiveAuthorityPublicKey() ([]byte, error) {
	return nil, errorcode.ErrorNotImplemented
}
