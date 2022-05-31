// This package contains helper functions that can be used within the entire app.
// On one hand, it includes functions as extensions to the `ppks` package,
// like the functions of serialization and deserialization for `*ppks.CipherText` and `*ZKProof`.
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
	"github.com/tjfoc/gmsm/sm2"

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
		return nil, fmt.Errorf("密钥材料或份额长度不正确，应为 128 字节")
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

// SerializeZKProof serializes a `ZKProof` object into a byte slice of length of 96.
func SerializeZKProof(proof *ZKProof) []byte {
	// 一份 proof 的内容是 3 个 big.Int，每个用 .Bytes() 提取信息得到长度为 32 的 []byte。
	// 3 个 big.Int 提取 []byte 后按顺序拼接成长度为 96 的 []byte，作为序列化结果。
	proofBytes := make([]byte, 96)
	copy(proofBytes[:32], proof.C.Bytes())
	copy(proofBytes[32:64], proof.R1.Bytes())
	copy(proofBytes[64:], proof.R2.Bytes())

	return proofBytes
}

// DeserializeZKProof parses a byte slice of length of 96 into a `ZKProof` object.
func DeserializeZKProof(proofBytes []byte) (*ZKProof, error) {
	// Byte slice 内容是 ZKProof 中的 3 个 big.Int 的 .Bytes() 信息，各占长度 32。
	// 将它们分 3 段装填入 3 个 big.Int 即可。
	if len(proofBytes) != 96 {
		return nil, fmt.Errorf("序列化的零知识证明长度不正确，应为 96 字节，得到 %v 字节", len(proofBytes))
	}

	var c, r1, r2 big.Int
	_ = c.SetBytes(proofBytes[:32])
	_ = r1.SetBytes(proofBytes[32:64])
	_ = r2.SetBytes(proofBytes[64:])

	proof := ZKProof{
		C:  &c,
		R1: &r1,
		R2: &r2,
	}

	return &proof, nil
}

// SerializeSM2PublicKey 将一个 SM2 公钥序列化成一个长度为 64 的字节切片。
func SerializeSM2PublicKey(publicKey *sm2.PublicKey) []byte {
	pubKeyBytes := [64]byte{}
	copy(pubKeyBytes[:32], publicKey.X.Bytes())
	copy(pubKeyBytes[32:], publicKey.Y.Bytes())
	return pubKeyBytes[:]
}

// DeserializeSM2PublicKey 解析一个长度为 64 的字节切片，得到 *sm2.PublicKey。
func DeserializeSM2PublicKey(publicKeyBytes []byte) (*sm2.PublicKey, error) {
	if len(publicKeyBytes) != 64 {
		return nil, fmt.Errorf("公钥字节切片长度不正确")
	}

	publicKeyX, publicKeyY := big.Int{}, big.Int{}
	_ = publicKeyX.SetBytes(publicKeyBytes[:32])
	_ = publicKeyY.SetBytes(publicKeyBytes[32:])

	publicKey, err := sm2keyutils.ConvertBigIntegersToPublicKey(&publicKeyX, &publicKeyY)
	if err != nil {
		return nil, err
	}

	return publicKey, nil
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

type ZKProof struct {
	C  *big.Int
	R1 *big.Int
	R2 *big.Int
}
