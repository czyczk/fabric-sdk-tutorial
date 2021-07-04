package service

import (
	"testing"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/utils/cipherutils"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/sm2keyutils"

	"github.com/XiaoYao-austin/ppks"
	"github.com/stretchr/testify/assert"
)

func TestKeySwitchProcess(t *testing.T) {
	collPubKeyPEM := "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0DQgAE2rffx6xbyAJYplmi2tfY7CI87Ls+\nObY6vcYIKD/qllWp9bOcC/KwojfrxRBDv54dVpJkn22v0PfxX8qZ5GF1vA==\n-----END PUBLIC KEY-----"
	collPubKeyInSM2, err := sm2keyutils.ConvertPEMToPublicKey([]byte(collPubKeyPEM))
	if isNoError := assert.NoError(t, err); !isNoError {
		t.FailNow()
	}

	participant1KeyPEM := "-----BEGIN PRIVATE KEY-----\nMIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQgdl2HAVMULxxRwLVY\nOVMcex1V/XM4Xc8tWNmqR7WVFaGgCgYIKoEcz1UBgi2hRANCAAQXrSPKlA4WBE/i\nGipzuOA4yndecO/eI8dWsPWvIb/yJlxdxkXUYzrxV0pPfFNO1efuoK3cwBUkkFFz\nYsVgZNWs\n-----END PRIVATE KEY-----"
	participant1KeyInSM2, err := sm2keyutils.ConvertPEMToPrivateKey([]byte(participant1KeyPEM))
	if isNoError := assert.NoError(t, err); !isNoError {
		t.FailNow()
	}

	participant2KeyPEM := "-----BEGIN PRIVATE KEY-----\nMIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQgvMIvGAX44pja2IHV\ng1w7Ki7T3MBkydwuLirtpm3a01+gCgYIKoEcz1UBgi2hRANCAAQDlT1WR9KYBPqJ\nayuAkBkZXS2paMAfdSUQg1Z2eP9IhHU8hIFxrQcQEUrOemfx7buzGGNY+dQ+JRQr\n7NGcC3qG\n-----END PRIVATE KEY-----"
	participant2KeyInSM2, err := sm2keyutils.ConvertPEMToPrivateKey([]byte(participant2KeyPEM))
	if isNoError := assert.NoError(t, err); !isNoError {
		t.FailNow()
	}

	targetPubKeyPEM := "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0DQgAEF60jypQOFgRP4hoqc7jgOMp3XnDv\n3iPHVrD1ryG/8iZcXcZF1GM68VdKT3xTTtXn7qCt3MAVJJBRc2LFYGTVrA==\n-----END PUBLIC KEY-----"
	targetPubKeyInSM2, err := sm2keyutils.ConvertPEMToPublicKey([]byte(targetPubKeyPEM))
	if isNoError := assert.NoError(t, err); !isNoError {
		t.FailNow()
	}

	targetPrivKeyPEM := participant1KeyPEM
	targetPrivKeyInSM2, err := sm2keyutils.ConvertPEMToPrivateKey([]byte(targetPrivKeyPEM))
	if isNoError := assert.NoError(t, err); !isNoError {
		t.FailNow()
	}

	key := ppks.GenPoint()

	// 用集合公钥加密 key
	encryptedKey, err := ppks.PointEncrypt(collPubKeyInSM2, key)
	if isNoError := assert.NoError(t, err); !isNoError {
		t.FailNow()
	}

	// 序列化加密后的 key
	encryptedKeyBytes := cipherutils.SerializeCipherText(encryptedKey)
	// 反序列化加密后的 key
	unsEncryptedKeyBytes, err := cipherutils.DeserializeCipherText(encryptedKeyBytes)
	if isNoError := assert.NoError(t, err); !isNoError {
		t.FailNow()
	}
	if isEqual := assert.Equal(t, unsEncryptedKeyBytes, encryptedKey); !isEqual {
		t.FailNow()
	}

	// 参与者 1 计算份额
	share1, err := ppks.ShareCal(targetPubKeyInSM2, &encryptedKey.K, participant1KeyInSM2)
	if isNoError := assert.NoError(t, err); !isNoError {
		t.FailNow()
	}

	// 参与者 2 计算份额
	share2, err := ppks.ShareCal(targetPubKeyInSM2, &encryptedKey.K, participant2KeyInSM2)
	if isNoError := assert.NoError(t, err); !isNoError {
		t.FailNow()
	}

	var shares [][]byte
	shares = append(shares, cipherutils.SerializeCipherText(share1))
	shares = append(shares, cipherutils.SerializeCipherText(share2))

	// 解密
	// 组建一个 CipherVector。将每份 share 转化为两个 CurvePoint 后，分别作为 CipherText 的 K 和 C，将 CipherText 放入 CipherVector。
	var cipherVector ppks.CipherVector
	for _, share := range shares {
		if isEqual := assert.Equal(t, 128, len(share)); !isEqual {
			t.FailNow()
		}

		cipherText, err := cipherutils.DeserializeCipherText(share)
		if isNoError := assert.NoError(t, err); !isNoError {
			t.FailNow()
		}
		cipherVector = append(cipherVector, *cipherText)
	}

	// 解析加密后的密钥材料
	encryptedKeyAsCipherText, err := cipherutils.DeserializeCipherText(encryptedKeyBytes)
	if isNoError := assert.NoError(t, err); !isNoError {
		t.FailNow()
	}

	// 密钥置换
	shareReplacedCipherText, err := ppks.ShareReplace(&cipherVector, encryptedKeyAsCipherText)
	if isNoError := assert.NoError(t, err); !isNoError {
		t.FailNow()
	}

	// 用用户的私钥解密 CipherText
	decryptedKey, err := ppks.PointDecrypt(shareReplacedCipherText, targetPrivKeyInSM2)
	if isNoError := assert.NoError(t, err); !isNoError {
		t.FailNow()
	}
	if isEqual := assert.Equal(t, key, decryptedKey); !isEqual {
		t.FailNow()
	}
}
