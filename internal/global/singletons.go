package global

import (
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/chaincodectx"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/polkadotnetwork"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/tjfoc/gmsm/sm2"
)

type keySwitchKeys struct {
	EncryptionAlgorithm  string          // The encryption algorithm used by the keys
	CollectivePrivateKey *sm2.PrivateKey // The collective private key to be used in regulator functions
	CollectivePublicKey  *sm2.PublicKey  // The collective public key to be used in the key switch process
	PrivateKey           *sm2.PrivateKey // The private key to be used in the key switch process
	PublicKey            *sm2.PublicKey  // The public key to be used in the key switch process
}

var BlockchainType blockchain.BCType             // Indicates the current Blockchain type
var BlockchainContext chaincodectx.IChaincodeCtx // The chaincode context to be used when the app is started as a server. Not useful when the app is started as the network initializer.
var KeySwitchKeys keySwitchKeys                  // The keys to be used in the key switch process
var ShowTimingLogs bool                          // Whether timers in several modules should be enabled and time consumption logged

// The SDK instnace and client instances below are for Fabric only
var FabricSDKInstance *fabsdk.FabricSDK                                     // Will be instantiated if the "BlockchainType" is "Fabric"
var ResMgmtClientInstances map[string]map[string]*resmgmt.Client            // A lookup takes `orgName` followed by `username`.
var MSPClientInstances map[string]map[string]*msp.Client                    // A lookup takes `orgName` followed by `username`.
var ChannelClientInstances map[string]map[string]map[string]*channel.Client // A lookup takes `channelID` followed by `orgName` and `username`.
var EventClientInstances map[string]map[string]map[string]*event.Client     // A lookup takes `channelID` followed by `orgName` and `username`.
var LedgerClientInstances map[string]map[string]map[string]*ledger.Client   // A lookup takes `channelID` followed by `orgName` and `username`.

// The instances below are for Polkadot only
var PolkadotNetworkConfig *polkadotnetwork.PolkadotNetworkConfig
