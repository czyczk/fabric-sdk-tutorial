package fabric

import (
	"encoding/json"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/bcao"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/chaincodectx"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/errorcode"
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

func (o *KeySwitchBCAOFabricImpl) CreateKeySwitchTrigger(ksTrigger *keyswitch.KeySwitchTrigger, eventID ...string) (string, error) {
	// TODO
	return "", errorcode.ErrorNotImplemented
}

func (o *KeySwitchBCAOFabricImpl) CreateKeySwitchResult(ksResult *keyswitch.KeySwitchResult) (string, error) {
	// TODO
	return "", errorcode.ErrorNotImplemented
}

func (o *KeySwitchBCAOFabricImpl) GetKeySwitchResult(query *keyswitch.KeySwitchResultQuery) (*keyswitch.KeySwitchResultStored, error) {
	// TODO
	return nil, errorcode.ErrorNotImplemented
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
