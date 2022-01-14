package polkadotbcao

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/chaincodectx"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/errorcode"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/data"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/query"
	"github.com/pkg/errors"
)

type DataBCAOPolkadotImpl struct {
	ctx    *chaincodectx.PolkadotChaincodeCtx
	client *http.Client
}

func NewDataBCAOPolkadotImpl(ctx *chaincodectx.PolkadotChaincodeCtx) *DataBCAOPolkadotImpl {
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	return &DataBCAOPolkadotImpl{
		ctx:    ctx,
		client: client,
	}
}

func (o *DataBCAOPolkadotImpl) CreatePlainData(plainData *data.PlainData, eventID ...string) (string, error) {
	funcName := "createPlainData"
	funcArgs := []interface{}{plainData}
	if len(eventID) != 0 {
		funcArgs = append(funcArgs, eventID[0])
	}

	result, err := sendTx(o.ctx, o.client, funcName, funcArgs, true)
	if err != nil {
		return "", err
	}

	return result.TxHash, nil
}

func (o *DataBCAOPolkadotImpl) CreateEncryptedData(encryptedData *data.EncryptedData, eventID ...string) (string, error) {
	// TODO
	return "", errorcode.ErrorNotImplemented
}

func (o *DataBCAOPolkadotImpl) CreateOffchainData(offchainData *data.OffchainData, eventID ...string) (string, error) {
	// TODO
	return "", errorcode.ErrorNotImplemented
}

func (o *DataBCAOPolkadotImpl) GetMetadata(resourceID string) (*data.ResMetadataStored, error) {
	funcName := "getMetadata"
	funcArgs := []interface{}{resourceID}
	result, err := sendQuery(o.ctx, o.client, funcName, funcArgs)
	if err != nil {
		return nil, err
	}

	metadataStr, err := unwrapOk(result.Output)
	if err != nil {
		return nil, err
	}

	metadataBytes := []byte(metadataStr)

	var resMetadataStored data.ResMetadataStored
	if err = json.Unmarshal(metadataBytes, &resMetadataStored); err != nil {
		return nil, errors.Wrap(err, "获取的元数据不合法")
	}
	return &resMetadataStored, nil
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

func (o *DataBCAOPolkadotImpl) GetKey(resourceID string) ([]byte, error) {
	// TODO
	return nil, errorcode.ErrorNotImplemented
}

func (o *DataBCAOPolkadotImpl) GetPolicy(resourceID string) ([]byte, error) {
	// TODO
	return nil, errorcode.ErrorNotImplemented
}

func (o *DataBCAOPolkadotImpl) ListResourceIDsByCreator(dataType string, isDesc bool, pageSize int, bookmark string) (*query.IDsWithPagination, error) {
	// TODO
	return nil, errorcode.ErrorNotImplemented
}

func (o *DataBCAOPolkadotImpl) ListResourceIDsByConditions(queryConditions map[string]interface{}, pageSize int, bookmark string) (*query.IDsWithPagination, error) {
	// TODO
	return nil, errorcode.ErrorNotImplemented
}
