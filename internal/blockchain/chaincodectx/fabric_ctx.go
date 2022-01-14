package chaincodectx

import (
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
)

type FabricChaincodeCtx struct {
	ChannelID     string
	OrgName       string
	Username      string
	ChaincodeID   string
	ChannelClient *channel.Client
	EventClient   *event.Client
	LedgerClient  *ledger.Client
}

func (ctx *FabricChaincodeCtx) GetBCType() blockchain.BCType {
	return blockchain.Fabric
}
