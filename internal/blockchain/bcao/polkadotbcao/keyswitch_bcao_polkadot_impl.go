package polkadotbcao

import (
	"encoding/json"
	"net/http"
	"time"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/bcao"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/chaincodectx"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/utils/idutils"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/keyswitch"
	"github.com/pkg/errors"
)

type KeySwitchBCAOPolkadotImpl struct {
	ctx    *chaincodectx.PolkadotChaincodeCtx
	client *http.Client
}

func NewKeySwitchBCAOPolkadotImpl(ctx *chaincodectx.PolkadotChaincodeCtx) *KeySwitchBCAOPolkadotImpl {
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	return &KeySwitchBCAOPolkadotImpl{
		ctx:    ctx,
		client: client,
	}
}

func (o *KeySwitchBCAOPolkadotImpl) CreateKeySwitchTrigger(ksTrigger *keyswitch.KeySwitchTrigger, eventID ...string) (*bcao.TransactionCreationInfoWithManualID, error) {
	funcName := "createKeySwitchTrigger"

	id, err := idutils.GenerateSnowflakeId()
	if err != nil {
		return nil, errors.Wrap(err, "无法为密钥置换触发生成 ID")
	}

	// TODO: Retrieve and pass the department identity as the second param when the contract is ready for it.
	funcArgs := []interface{}{id, nil, ksTrigger}
	if len(eventID) != 0 {
		funcArgs = append(funcArgs, eventID[0])
	} else {
		funcArgs = append(funcArgs, nil)
	}

	result, err := sendTx(o.ctx, o.client, funcName, funcArgs, true)
	if err != nil {
		return nil, err
	}

	creationInfo := &bcao.TransactionCreationInfoWithManualID{
		ManualID: id,
		TransactionCreationInfo: &bcao.TransactionCreationInfo{
			TransactionID: result.TxHash,
			BlockID:       result.InBlockStatus.InBlock,
		},
	}

	return creationInfo, nil
}

func (o *KeySwitchBCAOPolkadotImpl) CreateKeySwitchResult(ksResult *keyswitch.KeySwitchResult) (*bcao.TransactionCreationInfo, error) {
	funcName := "createKeySwitchResult"

	funcArgs := []interface{}{ksResult}

	result, err := sendTx(o.ctx, o.client, funcName, funcArgs, true)
	if err != nil {
		return nil, err
	}

	creationInfo := &bcao.TransactionCreationInfo{
		TransactionID: result.TxHash,
		BlockID:       result.InBlockStatus.InBlock,
	}

	return creationInfo, nil
}

func (o *KeySwitchBCAOPolkadotImpl) GetKeySwitchResult(query *keyswitch.KeySwitchResultQuery) (*keyswitch.KeySwitchResultStored, error) {
	funcName := "getKeySwitchResult"
	funcArgs := []interface{}{query}
	result, err := sendQuery(o.ctx, o.client, funcName, funcArgs, false)
	if err != nil {
		return nil, bcao.GetClassifiedError(funcName, err)
	}

	keySwitchResultBytes, err := unwrapOk(result.Output)
	if err != nil {
		return nil, err
	}

	var keySwitchResultStored keyswitch.KeySwitchResultStored
	if err = json.Unmarshal(keySwitchResultBytes, &keySwitchResultStored); err != nil {
		return nil, errors.Wrap(err, "无法解析密钥置换结果")
	}
	return &keySwitchResultStored, nil
}

func (o *KeySwitchBCAOPolkadotImpl) ListKeySwitchResultsByID(ksSessionID string) ([]*keyswitch.KeySwitchResultStored, error) {
	funcName := "listKeySwitchResultsById"
	funcArgs := []interface{}{ksSessionID}
	result, err := sendQuery(o.ctx, o.client, funcName, funcArgs, false)
	if err != nil {
		return nil, bcao.GetClassifiedError(funcName, err)
	}

	keySwitchResultsBytes, err := unwrapOk(result.Output)
	if err != nil {
		return nil, err
	}

	var keySwitchResults []*keyswitch.KeySwitchResultStored
	err = json.Unmarshal(keySwitchResultsBytes, &keySwitchResults)
	if err != nil {
		return nil, errors.Wrap(err, "无法解析密钥置换结果列表")
	}

	return keySwitchResults, nil
}
