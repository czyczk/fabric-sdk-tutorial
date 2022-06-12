package fabricbcao

import (
	"encoding/json"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/bcao"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/chaincodectx"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/keyswitch"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/pkg/errors"
)

type KeySwitchBCAOFabricImpl struct {
	ctx *chaincodectx.FabricChaincodeCtx
}

func NewKeySwitchBCAOFabricImpl(ctx *chaincodectx.FabricChaincodeCtx) *KeySwitchBCAOFabricImpl {
	return &KeySwitchBCAOFabricImpl{
		ctx: ctx,
	}
}

func (o *KeySwitchBCAOFabricImpl) CreateKeySwitchTrigger(ksTrigger *keyswitch.KeySwitchTrigger, eventID ...string) (*bcao.TransactionCreationInfoWithManualID, error) {
	ksTriggerBytes, err := json.Marshal(ksTrigger)
	if err != nil {
		return nil, errors.Wrap(err, "无法序列化链码参数")
	}

	chaincodeFcn := "createKeySwitchTrigger"
	chaincodeArgs := [][]byte{ksTriggerBytes}
	if len(eventID) != 0 {
		chaincodeArgs = append(chaincodeArgs, []byte(eventID[0]))
	}
	channelReq := channel.Request{
		ChaincodeID: o.ctx.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        chaincodeArgs,
	}

	resp, err := o.ctx.ChannelClient.Execute(channelReq)
	if err != nil {
		return nil, bcao.GetClassifiedError(chaincodeFcn, err)
	}

	// Get the block ID from the ledger client
	txID := resp.TransactionID
	blockHashAsHex, err := getBlockHashFromTxID(o.ctx.LedgerClient, txID)
	if err != nil {
		return nil, bcao.GetClassifiedError(chaincodeFcn, err)
	}

	creationInfo := &bcao.TransactionCreationInfoWithManualID{
		ManualID: string(txID),
		TransactionCreationInfo: &bcao.TransactionCreationInfo{
			TransactionID: string(txID),
			BlockID:       blockHashAsHex,
		},
	}

	return creationInfo, nil
}

func (o *KeySwitchBCAOFabricImpl) CreateKeySwitchResult(ksResult *keyswitch.KeySwitchResult) (*bcao.TransactionCreationInfo, error) {
	keySwitchResultBytes, err := json.Marshal(ksResult)
	if err != nil {
		return nil, errors.Wrap(err, "无法序列化链码参数")
	}

	chaincodeFcn := "createKeySwitchResult"
	channelReq := channel.Request{
		ChaincodeID: o.ctx.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{keySwitchResultBytes},
	}

	resp, err := o.ctx.ChannelClient.Execute(channelReq)
	if err != nil {
		return nil, bcao.GetClassifiedError(chaincodeFcn, err)
	}

	// Get the block ID from the ledger client
	txID := resp.TransactionID
	blockHashAsHex, err := getBlockHashFromTxID(o.ctx.LedgerClient, txID)
	if err != nil {
		return nil, bcao.GetClassifiedError(chaincodeFcn, err)
	}

	creationInfo := &bcao.TransactionCreationInfo{
		TransactionID: string(txID),
		BlockID:       blockHashAsHex,
	}

	return creationInfo, nil
}

func (o *KeySwitchBCAOFabricImpl) GetKeySwitchResult(query *keyswitch.KeySwitchResultQuery) (*keyswitch.KeySwitchResultStored, error) {
	queryBytes, err := json.Marshal(query)
	if err != nil {
		return nil, errors.Wrap(err, "无法序列化链码参数")
	}

	chaincodeFcn := "getKeySwitchResult"
	channelReq := channel.Request{
		ChaincodeID: o.ctx.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{queryBytes},
	}

	resp, err := o.ctx.ChannelClient.Query(channelReq)
	if err != nil {
		return nil, bcao.GetClassifiedError(chaincodeFcn, err)
	}

	var ksResult *keyswitch.KeySwitchResultStored
	err = json.Unmarshal(resp.Payload, &ksResult)
	if err != nil {
		return nil, errors.Wrap(err, "无法解析密钥置换结果")
	}

	return ksResult, nil
}

func (o *KeySwitchBCAOFabricImpl) ListKeySwitchResultsByID(ksSessionID string) ([]*keyswitch.KeySwitchResultStored, error) {
	chaincodeFcn := "listKeySwitchResultsByID"
	channelReq := channel.Request{
		ChaincodeID: o.ctx.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{[]byte(ksSessionID)},
	}

	resp, err := o.ctx.ChannelClient.Query(channelReq)
	err = bcao.GetClassifiedError(chaincodeFcn, err)
	if err != nil {
		return nil, err
	}

	var ksResults []*keyswitch.KeySwitchResultStored
	err = json.Unmarshal(resp.Payload, &ksResults)
	if err != nil {
		return nil, errors.Wrap(err, "无法解析密钥置换结果列表")
	}

	return ksResults, nil
}
