package polkadotbcao

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/bcao"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/chaincodectx"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/models/common"
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
	} else {
		funcArgs = append(funcArgs, nil)
	}

	result, err := sendTx(o.ctx, o.client, funcName, funcArgs, true)
	if err != nil {
		return "", err
	}

	return result.TxHash, nil
}

func (o *DataBCAOPolkadotImpl) CreateEncryptedData(encryptedData *data.EncryptedData, eventID ...string) (string, error) {
	funcName := "createEncryptedData"

	funcArgs := []interface{}{encryptedData}
	if len(eventID) != 0 {
		funcArgs = append(funcArgs, eventID[0])
	} else {
		funcArgs = append(funcArgs, nil)
	}

	result, err := sendTxWithTimer(o.ctx, o.client, funcName, funcArgs, true, "链上存储文档")
	if err != nil {
		return "", err
	}

	return result.TxHash, nil
}

func (o *DataBCAOPolkadotImpl) CreateOffchainData(offchainData *data.OffchainData, eventID ...string) (string, error) {
	funcName := "createOffchainData"

	funcArgs := []interface{}{offchainData}
	if len(eventID) != 0 {
		funcArgs = append(funcArgs, eventID[0])
	} else {
		funcArgs = append(funcArgs, nil)
	}

	result, err := sendTxWithTimer(o.ctx, o.client, funcName, funcArgs, true, "链上存储文档元数据与属性")
	if err != nil {
		return "", err
	}

	return result.TxHash, nil
}

func (o *DataBCAOPolkadotImpl) GetMetadata(resourceID string) (*data.ResMetadataStored, error) {
	funcName := "getMetadata"
	funcArgs := []interface{}{resourceID}
	result, err := sendQuery(o.ctx, o.client, funcName, funcArgs, false)
	if err != nil {
		return nil, bcao.GetClassifiedError(funcName, err)
	}

	metadataBytes, err := unwrapOk(result.Output)
	if err != nil {
		return nil, err
	}

	var resMetadataStored data.ResMetadataStored
	if err = json.Unmarshal(metadataBytes, &resMetadataStored); err != nil {
		return nil, errors.Wrap(err, "获取的元数据不合法")
	}
	return &resMetadataStored, nil
}

func (o *DataBCAOPolkadotImpl) GetData(resourceID string) ([]byte, error) {
	funcName := "getData"
	funcArgs := []interface{}{resourceID}
	result, err := sendQuery(o.ctx, o.client, funcName, funcArgs, false)
	if err != nil {
		return nil, bcao.GetClassifiedError(funcName, err)
	}

	contentsAsBase64Bytes, err := unwrapOk(result.Output)
	if err != nil {
		return nil, err
	}
	// The content of `Ok` is actually a JSON string. So there's one more step we should take:
	// Getting rid of the leading and trailing quotes is fine enough.
	contentsAsBase64Bytes = contentsAsBase64Bytes[1:]
	contentsAsBase64Bytes = contentsAsBase64Bytes[:len(contentsAsBase64Bytes)-1]

	contents, err := base64.StdEncoding.DecodeString(string(contentsAsBase64Bytes))
	if err != nil {
		return nil, errors.Wrap(err, "无法解析资源本体")
	}

	return contents, nil
}

func (o *DataBCAOPolkadotImpl) GetKey(resourceID string) ([]byte, error) {
	funcName := "getKey"
	funcArgs := []interface{}{resourceID}
	result, err := sendQuery(o.ctx, o.client, funcName, funcArgs, false)
	if err != nil {
		return nil, bcao.GetClassifiedError(funcName, err)
	}

	encryptedKeyAsBase64Bytes, err := unwrapOk(result.Output)
	if err != nil {
		return nil, err
	}

	encryptedKeyAsBase64Bytes = encryptedKeyAsBase64Bytes[1:]
	encryptedKeyAsBase64Bytes = encryptedKeyAsBase64Bytes[:len(encryptedKeyAsBase64Bytes)-1]

	encryptedKeyBytes, err := base64.StdEncoding.DecodeString(string(encryptedKeyAsBase64Bytes))
	if err != nil {
		return nil, errors.Wrap(err, "无法解析资源加密后的密钥")
	}

	return encryptedKeyBytes, nil
}

func (o *DataBCAOPolkadotImpl) GetPolicy(resourceID string) ([]byte, error) {
	funcName := "getPolicy"
	funcArgs := []interface{}{resourceID}
	result, err := sendQuery(o.ctx, o.client, funcName, funcArgs, false)
	if err != nil {
		return nil, bcao.GetClassifiedError(funcName, err)
	}

	policyAsBase64Bytes, err := unwrapOk(result.Output)
	if err != nil {
		return nil, err
	}

	policyAsBase64Bytes = policyAsBase64Bytes[1:]
	policyAsBase64Bytes = policyAsBase64Bytes[:len(policyAsBase64Bytes)-1]

	policyBytes, err := base64.StdEncoding.DecodeString(string(policyAsBase64Bytes))
	if err != nil {
		return nil, errors.Wrap(err, "无法解析策略")
	}

	return policyBytes, nil
}

func (o *DataBCAOPolkadotImpl) ListResourceIDsByCreator(dataType string, isDesc bool, pageSize int, bookmark string) (*query.IDsWithPagination, error) {
	funcName := "listResourceIdsByCreator"
	funcArgs := []interface{}{dataType, isDesc, pageSize, bookmark}
	result, err := sendQuery(o.ctx, o.client, funcName, funcArgs, false)
	if err != nil {
		return nil, bcao.GetClassifiedError(funcName, err)
	}

	resourceIDsBytes, err := unwrapOk(result.Output)
	if err != nil {
		return nil, err
	}

	var resourceIDs query.IDsWithPagination
	err = json.Unmarshal(resourceIDsBytes, &resourceIDs)
	if err != nil {
		return nil, errors.Wrap(err, "无法解析结果列表")
	}

	return &resourceIDs, nil
}

func (o *DataBCAOPolkadotImpl) ListResourceIDsByConditions(conditions common.QueryConditions, pageSize int) (*query.IDsWithPagination, error) {
	conditionsScaleReady := conditions.ToScaleReadyStructure()

	funcName := "listResourceIdsByConditions"
	funcArgs := []interface{}{conditionsScaleReady, pageSize}
	result, err := sendQuery(o.ctx, o.client, funcName, funcArgs, false)
	if err != nil {
		return nil, bcao.GetClassifiedError(funcName, err)
	}

	chaincodeResourceIDsBytes, err := unwrapOk(result.Output)
	if err != nil {
		return nil, err
	}

	var chaincodeResourceIDs query.IDsWithPagination
	err = json.Unmarshal(chaincodeResourceIDsBytes, &chaincodeResourceIDs)
	if err != nil {
		return nil, errors.Wrap(err, "无法解析结果列表")
	}

	return &chaincodeResourceIDs, nil
}
