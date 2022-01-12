package service

import (
	"crypto"
	"encoding/base64"
	"fmt"
	"strings"
	"sync"
	"time"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/bcao"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/eventmgr"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/global"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/utils/cipherutils"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/errorcode"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/keyswitch"
	"github.com/XiaoYao-austin/ppks"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/tjfoc/gmsm/sm2"
)

// KeySwitchService 实现了 `KeySwitchServiceInterface` 接口，提供有关于密钥置换的服务
type KeySwitchService struct {
	ServiceInfo   *Info
	KeySwitchBCAO bcao.IKeySwitchBCAO
	EventManager  eventmgr.IEventManager
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
	ksPubKey := cipherutils.SerializeSM2PublicKey(global.KeySwitchKeys.PublicKey)

	// 组装一个 KeySwitchTrigger 对象，并调用链码
	ksTrigger := keyswitch.KeySwitchTrigger{
		ResourceID:    resourceID,
		AuthSessionID: authSessionID,
		KeySwitchPK:   base64.StdEncoding.EncodeToString(ksPubKey),
	}

	// 固定的事件 ID
	eventID := "ks_trigger"

	return s.KeySwitchBCAO.CreateKeySwitchTrigger(&ksTrigger, eventID)
}

// 创建密钥置换结果。
//
// 参数：
//   密钥置换会话 ID
//   个人份额
//   关于份额的零知识证明
//
// 返回：
//   交易 ID
func (s *KeySwitchService) CreateKeySwitchResult(keySwitchSessionID string, share *ppks.CipherText, proof *cipherutils.ZKProof) (string, error) {
	if strings.TrimSpace(keySwitchSessionID) == "" {
		return "", fmt.Errorf("密钥置换会话 ID 不能为空")
	}

	// Serialize the materials
	// share.K and share.C each takes up 64 bytes
	shareBytes := cipherutils.SerializeCipherText(share)
	// Each big.Int in the proof takes up 32 bytes
	proofBytes := cipherutils.SerializeZKProof(proof)

	ksPubKeyBytes := cipherutils.SerializeSM2PublicKey(global.KeySwitchKeys.PublicKey)

	shareAsBase64 := base64.StdEncoding.EncodeToString(shareBytes)
	proofAsBase64 := base64.StdEncoding.EncodeToString(proofBytes)
	ksPubKeyAsBase64 := base64.StdEncoding.EncodeToString(ksPubKeyBytes)

	keySwitchResult := keyswitch.KeySwitchResult{
		KeySwitchSessionID: keySwitchSessionID,
		Share:              shareAsBase64,
		ZKProof:            proofAsBase64,
		KeySwitchPK:        ksPubKeyAsBase64,
	}

	return s.KeySwitchBCAO.CreateKeySwitchResult(&keySwitchResult)
}

// 验证所获得的份额。
//
// 参数：
//   所获的份额
//   所获的零知识证明
//   份额生成者的密钥置换公钥
//   目标用户的密钥置换公钥
//   加密后的对称密钥材料
//
// 返回：
//   该份额是否通过验证
func (s *KeySwitchService) VerifyShare(share *ppks.CipherText, proof *cipherutils.ZKProof, shareCreatorPublicKey *sm2.PublicKey, targetPublicKey *sm2.PublicKey, encryptedKey *ppks.CipherText) (bool, error) {
	isShareVerified, err := ppks.ShareProofVryNoB(proof.C, proof.R1, proof.R2, share, shareCreatorPublicKey, targetPublicKey, &encryptedKey.K)
	if err != nil {
		return isShareVerified, err
	}

	return isShareVerified, nil
}

// 获取解密后的对称密钥材料。调用前需要使用 `VerifyShare()` 对份额进行验证。
//
// 参数：
//   经零知识证明验证的份额
//   加密后的对称密钥材料
//   目标用户用于密钥置换的私钥
//
// 返回：
//   解密后的对称密钥材料
func (s *KeySwitchService) GetDecryptedKey(shares []*ppks.CipherText, encryptedKey *ppks.CipherText, targetPrivateKey *sm2.PrivateKey) (*ppks.CurvePoint, error) {
	// 组建一个 CipherVector，将每个 CipherText 放入 CipherVector。
	var cipherVector ppks.CipherVector
	for _, share := range shares {
		cipherVector = append(cipherVector, *share)
	}

	// 密钥置换
	shareReplacedCipherText, err := ppks.ShareReplace(&cipherVector, encryptedKey)
	if err != nil {
		return nil, errors.Wrap(err, "无法进行密钥置换")
	}

	// 用用户的私钥解密 CipherText
	decryptedKey, err := ppks.PointDecrypt(shareReplacedCipherText, targetPrivateKey)
	if err != nil {
		return nil, errors.Wrap(err, "无法解密对称密钥材料")
	}

	return decryptedKey, nil
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
func (s *KeySwitchService) AwaitKeySwitchResults(keySwitchSessionID string, numExpected int, timeout ...int) ([]*keyswitch.KeySwitchResultStored, error) {
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

	// 尝试监听事件 "ks_${keySwitchSessionID}_result"。事件内容为 "ks_${keySwitchSessionID}_result_${creator}"。若失败则提前返回。
	eventID := "ks_" + keySwitchSessionID + "_result"
	reg, notifier, err := s.EventManager.RegisterEvent(eventID)
	defer s.EventManager.UnregisterEvent(reg)

	if err != nil {
		return nil, errors.Wrap(err, "无法监听密钥置换结果事件")
	}

	// 接收 channel 内容（一个密钥置换会话 ID），对每个 ID 新开一个 Go routine 调用链码获取一次置换份额，并将结果放入列表。若在时限内未能接收到预期数量的 ID，则等待已发送的 Go routine 结束后，放弃此次结果，并返回超时错误。
	receivedIDs := []string{}
	var ret []*keyswitch.KeySwitchResultStored
	var wg sync.WaitGroup
	chanError := make(chan error)

eventHandler:
	for {
		select {
		case eventVal := <-notifier:
			log.Debugf("收到事件 {'%v': '%s'}", eventID, eventVal.GetPayload())
			dbKeyParts := strings.Split(string(eventVal.GetPayload()), "_")
			if len(dbKeyParts) != 4 {
				wg.Wait()
				return nil, fmt.Errorf("不合法的事件内容: %v", eventVal)
			}
			receivedIDs = append(receivedIDs, dbKeyParts[1])

			wg.Add(1)
			go func(chanErr chan error) {
				defer wg.Done()

				query := keyswitch.KeySwitchResultQuery{
					KeySwitchSessionID: dbKeyParts[1],
					ResultCreator:      dbKeyParts[3],
				}

				ksResult, err := s.KeySwitchBCAO.GetKeySwitchResult(&query)
				if err != nil {
					chanError <- err
					return
				}

				ret = append(ret, ksResult)
			}(chanError)

			err := <-chanError
			if err != nil {
				wg.Wait()
				return nil, err
			}

			// 收集到足够的结果便可停止接收事件
			if len(ret) == numExpected {
				break eventHandler
			}
		case <-time.After(time.Duration(timeoutInSec) * time.Second):
			// 只有收到的事件数量小于预期值才报超时错误，否则只是还有查询未完成，过一会自然会完成。
			if len(receivedIDs) < numExpected {
				wg.Wait()
				return nil, errorcode.ErrorGatewayTimeout
			}
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}

	wg.Wait()
	return ret, nil
}

// 获取集合权威公钥。
//
// 返回：
//   集合权威公钥（SM2）
func (s *KeySwitchService) GetCollectiveAuthorityPublicKey() (crypto.PublicKey, error) {
	// 当前设计为从单例 `global.KSCollPubKey` 中获取一个预指定的集合公钥。
	if global.KeySwitchKeys.CollectivePublicKey == nil {
		return nil, fmt.Errorf("集合公钥未指定")
	}

	ret := crypto.PublicKey(global.KeySwitchKeys.CollectivePublicKey)

	return ret, nil
}
