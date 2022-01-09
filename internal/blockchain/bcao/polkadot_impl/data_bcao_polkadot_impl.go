package polkadot

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/contractctx"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/data"
)

type DataBCAOPolkadotImpl struct {
	ctx    *contractctx.PolkadotContractCtx
	client *http.Client
}

func NewDataBCAOSubstrateImpl() *DataBCAOPolkadotImpl {
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	return &DataBCAOPolkadotImpl{
		client: client,
	}
}

func (o *DataBCAOPolkadotImpl) CreatePlainData(plainData data.PlainData, eventID *string) (string, error) {
	funcName := "createPlainData"
	funcArgs := []interface{}{plainData, eventID}
	result, err := sendTx(o.ctx, o.client, funcName, funcArgs, true)
	if err != nil {
		return "", err
	}

	return result.TxHash, nil
}

func (o *DataBCAOPolkadotImpl) GetData(resourceID string) ([]byte, error) {
	funcName := "getData"
	funcArgs := []interface{}{resourceID}
	result, err := sendQuery(o.ctx, o.client, funcName, funcArgs)
	if err != nil {
		return nil, err
	}

	contentsAsBase64, err := unwrapOk(result.Output)
	if err != nil {
		return nil, err
	}

	contents, err := base64.StdEncoding.DecodeString(contentsAsBase64)
	if err != nil {
		return nil, fmt.Errorf("无法解析资源本体")
	}

	return contents, nil
}

// TODO: GetMetadata
