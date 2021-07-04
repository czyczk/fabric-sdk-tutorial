package cipherutils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"testing"

	"github.com/XiaoYao-austin/ppks"
	"github.com/stretchr/testify/assert"
)

func TestAESEncryptionDecryption(t *testing.T) {
	key := ppks.GenPoint()
	documentBytes := []byte("Document for test")

	// 用 key 加密 documentBytes
	// 使用由 key 导出的 256 位信息来创建 AES256 block
	cipherBlock, err := aes.NewCipher(DeriveSymmetricKeyBytesFromCurvePoint(key))
	if isNoError := assert.NoError(t, err); !isNoError {
		t.FailNow()
	}

	aesGCM, err := cipher.NewGCM(cipherBlock)
	if isNoError := assert.NoError(t, err); !isNoError {
		t.FailNow()
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		t.FailNow()
	}

	encryptedDocumentBytes := aesGCM.Seal(nonce, nonce, documentBytes, nil)

	// 用对称密钥解密 encryptedDocumentBytes
	decCipherBlock, err := aes.NewCipher(DeriveSymmetricKeyBytesFromCurvePoint(key))
	if isNoError := assert.NoError(t, err); !isNoError {
		t.FailNow()
	}

	decAesGCM, err := cipher.NewGCM(decCipherBlock)
	if isNoError := assert.NoError(t, err); !isNoError {
		t.FailNow()
	}

	nonceSize := decAesGCM.NonceSize()
	if len(encryptedDocumentBytes) < nonceSize {
		t.FailNow()
	}

	decNonce, encryptedDocumentBytes := encryptedDocumentBytes[:nonceSize], encryptedDocumentBytes[nonceSize:]
	decDocumentBytes, err := aesGCM.Open(nil, decNonce, encryptedDocumentBytes, nil)
	if isNoError := assert.NoError(t, err); !isNoError {
		t.FailNow()
	}
	if isEqual := assert.Equal(t, documentBytes, decDocumentBytes); !isEqual {
		t.FailNow()
	}
}
