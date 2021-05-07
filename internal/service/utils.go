package service

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/XiaoYao-austin/ppks"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"gorm.io/gorm"

	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/errorcode"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/sm2keyutils"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Info needed for a service to know which `chaincodeID` it's serving and which channel it's using.
type Info struct {
	ChaincodeID   string
	ChannelClient *channel.Client
	EventClient   *event.Client
	LedgerClient  *ledger.Client
	DB            *gorm.DB
}

const eventTimeout time.Duration = 20

// RegisterEvent registers an event using a channel client.
//
// Returns:
//   the registration (used to unregister the event)
//   the event channel
func RegisterEvent(client *event.Client, chaincodeID, eventID string) (fab.Registration, <-chan *fab.CCEvent, error) {
	reg, notifier, err := client.RegisterChaincodeEvent(chaincodeID, eventID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to register chaincode event: %v", err)
	}
	return reg, notifier, nil
}

// Receives an event from the event channel and prints the event contents.
func showEventResult(notifier <-chan *fab.CCEvent, eventID string) error {
	select {
	case ccEvent := <-notifier:
		log.Printf("Chaincode event received: %v\n", ccEvent)
	case <-time.After(eventTimeout * time.Second):
		return fmt.Errorf("can't receive chaincode event with event ID '%v'", eventID)
	}

	return nil
}

// GetClassifiedError is a general error handler that converts some errors returned from the chaincode to the predefined errors.
func GetClassifiedError(chaincodeFcn string, err error) error {
	if err == nil {
		return nil
	} else if strings.HasSuffix(err.Error(), errorcode.CodeForbidden) {
		return errorcode.ErrorForbidden
	} else if strings.HasSuffix(err.Error(), errorcode.CodeNotFound) {
		return errorcode.ErrorNotFound
	} else if strings.HasSuffix(err.Error(), errorcode.CodeNotImplemented) {
		return errorcode.ErrorNotImplemented
	} else {
		return errors.Wrapf(err, "无法调用链码函数 '%v'", chaincodeFcn)
	}
}

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

// 对称密钥的生成是由 curvePoint 导出的 256 位信息，可用于创建 AES256 block
func deriveSymmetricKeyBytesFromCurvePoint(curvePoint *ppks.CurvePoint) []byte {
	return curvePoint.X.Bytes()
}
