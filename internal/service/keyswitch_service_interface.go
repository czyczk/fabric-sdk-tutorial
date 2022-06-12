package service

import (
	"crypto"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/bcao"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/utils/cipherutils"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/keyswitch"
	"github.com/XiaoYao-austin/ppks"
	"github.com/tjfoc/gmsm/sm2"
)

// KeySwitchServiceInterface 定义了有关于密钥置换的服务的接口
type KeySwitchServiceInterface interface {
	// 创建密文访问申请/密钥置换触发器。
	//
	// 参数：
	//   资源 ID
	//   授权会话 ID
	//
	// 返回：
	//   交易 ID
	CreateKeySwitchTrigger(resourceID string, authSessionID string) (*bcao.TransactionCreationInfoWithManualID, error)

	// 创建密钥置换结果。
	//
	// 参数：
	//   密钥置换会话 ID
	//   个人份额
	//   关于份额的零知识证明
	//
	// 返回：
	//   交易 ID
	CreateKeySwitchResult(keySwitchSessionID string, share *ppks.CipherText, proof *cipherutils.ZKProof) (*bcao.TransactionCreationInfo, error)

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
	VerifyShare(share *ppks.CipherText, proof *cipherutils.ZKProof, shareCreatorPublicKey *sm2.PublicKey, targetPublicKey *sm2.PublicKey, encryptedKey *ppks.CipherText) (bool, error)

	// 获取解密后的对称密钥材料。调用前需要使用 `VerifyShare()` 对份额进行验证。
	//
	// 参数：
	//   所获的份额
	//   加密后的对称密钥材料
	//   目标用户用于密钥置换的私钥
	//
	// 返回：
	//   解密后的对称密钥材料
	GetDecryptedKey(shares []*ppks.CipherText, encryptedKey *ppks.CipherText, targetPrivateKey *sm2.PrivateKey) (*ppks.CurvePoint, error)

	// 等待并收集密钥置换结果。
	//
	// 参数：
	//   密钥置换会话 ID
	//   预期的份额个数
	//   超时时限（可选）
	//
	// 返回：
	//   预期个数的份额列表
	AwaitKeySwitchResults(keySwitchSessionID string, numExpected int, timeout ...int) ([]*keyswitch.KeySwitchResultStored, error)

	// 获取集合权威公钥。
	//
	// 返回：
	//   集合权威公钥（SM2）
	GetCollectiveAuthorityPublicKey() (crypto.PublicKey, error)
}
