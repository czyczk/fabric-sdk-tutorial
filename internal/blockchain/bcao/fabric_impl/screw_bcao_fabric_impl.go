package fabricbcao

import (
	"strconv"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/chaincodectx"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
)

type ScrewBCAOFabricImpl struct {
	ctx *chaincodectx.FabricChaincodeCtx
}

func NewScrewBCAOFabricImpl(ctx *chaincodectx.FabricChaincodeCtx) *ScrewBCAOFabricImpl {
	return &ScrewBCAOFabricImpl{
		ctx: ctx,
	}
}

func (o *ScrewBCAOFabricImpl) Transfer(sourceCorpName string, targetCorpName string, transferAmnt uint, eventID string) (string, error) {
	channelReq := channel.Request{
		ChaincodeID: o.ctx.ChaincodeID,
		Fcn:         "transfer",
		Args:        [][]byte{[]byte(sourceCorpName), []byte(targetCorpName), []byte(strconv.Itoa(int(transferAmnt))), []byte(eventID)},
	}

	resp, err := o.ctx.ChannelClient.Execute(channelReq)
	if err != nil {
		return "", err
	}

	return string(resp.TransactionID), nil
}

func (o *ScrewBCAOFabricImpl) Query(targetCorpName string) (string, error) {
	channelReq := channel.Request{
		ChaincodeID: o.ctx.ChaincodeID,
		Fcn:         "query",
		Args:        [][]byte{[]byte(targetCorpName)},
	}

	resp, err := o.ctx.ChannelClient.Query(channelReq)
	if err != nil {
		return "", err
	}

	return string(resp.Payload), nil
}
