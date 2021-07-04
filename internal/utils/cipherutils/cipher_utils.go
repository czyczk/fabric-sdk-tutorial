// This package contains helper functions that can be used within the entire app.
// On one hand, it includes functions as extensions to the `ppks` package,
// like the functions of serialization and deserialization of `*ppks.CipherText`.
// On the other hand, it includes other handy tools for symmetric encryption and decryption using AES keys, etc..
package cipherutils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"math/big"

	"github.com/XiaoYao-austin/ppks"

	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/sm2keyutils"
)

// SerializeCipherText serializes a `CipherText` object into a byte slice of length of 128.
func SerializeCipherText(cipherText *ppks.CipherText) []byte {
	// 将左侧点 K 装入 [0:64]，将右侧点 C 装入 [64:128]
	encryptedKeyBytes := make([]byte, 128)
	copy(encryptedKeyBytes[:32], cipherText.K.X.Bytes())
	copy(encryptedKeyBytes[32:64], cipherText.K.Y.Bytes())
	copy(encryptedKeyBytes[64:96], cipherText.C.X.Bytes())
	copy(encryptedKeyBytes[96:], cipherText.C.Y.Bytes())

	return encryptedKeyBytes
}

// DeserializeCipherText parses a byte slice of length of 128 into a `CipherText` object.
func DeserializeCipherText(encryptedKeyBytes []byte) (*ppks.CipherText, error) {
	// 解析加密后的密钥材料，将其转化为两个 CurvePoint 后，分别作为 CipherText 的 K 和 C
	if len(encryptedKeyBytes) != 128 {
		return nil, fmt.Errorf("密钥材料长度不正确，应为 128 字节")
	}
	var pointKX, pointKY big.Int
	_ = pointKX.SetBytes(encryptedKeyBytes[:32])
	_ = pointKY.SetBytes(encryptedKeyBytes[32:64])

	encryptedKeyAsPubKeyK, err := sm2keyutils.ConvertBigIntegersToPublicKey(&pointKX, &pointKY)
	if err != nil {
		return nil, err
	}

	var pointCX, pointCY big.Int
	_ = pointCX.SetBytes(encryptedKeyBytes[64:96])
	_ = pointCY.SetBytes(encryptedKeyBytes[96:])

	encryptedKeyAsPubKeyC, err := sm2keyutils.ConvertBigIntegersToPublicKey(&pointCX, &pointCY)
	if err != nil {
		return nil, err
	}

	encryptedKeyAsCipherText := ppks.CipherText{
		K: (ppks.CurvePoint)(*encryptedKeyAsPubKeyK),
		C: (ppks.CurvePoint)(*encryptedKeyAsPubKeyC),
	}

	return &encryptedKeyAsCipherText, nil
}

// DeriveSymmetricKeyBytesFromCurvePoint 从 curvePoint 中导出 256 位信息，在应用内作为对称密钥。具体使用上可用于创建 AES256 block。
func DeriveSymmetricKeyBytesFromCurvePoint(curvePoint *ppks.CurvePoint) []byte {
	return curvePoint.X.Bytes()
}

// EncryptBytesUsingAESKey 使用 AES 对称密钥加密数据
func EncryptBytesUsingAESKey(b []byte, key []byte) (encryptedBytes []byte, err error) {
	cipherBlock, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	aesGCM, err := cipher.NewGCM(cipherBlock)
	if err != nil {
		return
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return
	}

	encryptedBytes = aesGCM.Seal(nonce, nonce, b, nil)
	return
}

// DecryptBytesUsingAESKey 使用 AES 对称密钥解密数据
func DecryptBytesUsingAESKey(b []byte, key []byte) (decryptedBytes []byte, err error) {
	cipherBlock, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	aesGCM, err := cipher.NewGCM(cipherBlock)
	if err != nil {
		return
	}

	nonceSize := aesGCM.NonceSize()
	if len(b) < nonceSize {
		err = fmt.Errorf("密文长度太短")
		return
	}

	nonce, b := b[:nonceSize], b[nonceSize:]
	decryptedBytes, err = aesGCM.Open(nil, nonce, b, nil)
	if err != nil {
		return
	}

	return
}
