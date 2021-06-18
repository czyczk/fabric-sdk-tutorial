package service

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	ipfs "github.com/ipfs/go-ipfs-api"
	"gorm.io/gorm"
)

// Info needed for a service to know which `chaincodeID` it's serving and which channel it's using.
type Info struct {
	ChaincodeID   string
	ChannelClient *channel.Client
	EventClient   *event.Client
	LedgerClient  *ledger.Client
	DB            *gorm.DB
	IPFSSh        *ipfs.Shell
}
