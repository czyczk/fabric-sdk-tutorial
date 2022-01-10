package chaincodectx

import (
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
)

type FabricChaincodeCtx struct {
	ChannelID     string
	OrgName       string
	Username      string
	ChaincodeID   string
	ResMgmtClient *resmgmt.Client
	MSPClient     *msp.Client
	ChannelClient *channel.Client
	EventClient   *event.Client
	LedgerClient  *ledger.Client
}

func (ctx *FabricChaincodeCtx) GetType() blockchain.BCType {
	return blockchain.Fabric
}
