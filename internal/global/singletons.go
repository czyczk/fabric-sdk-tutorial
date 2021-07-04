package global

import (
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

var SDKInstance *fabsdk.FabricSDK
var ResMgmtClientInstances map[string]map[string]*resmgmt.Client            // A lookup takes `orgName` followed by `username`.
var MSPClientInstances map[string]map[string]*msp.Client                    // A lookup takes `orgName` followed by `username`.
var ChannelClientInstances map[string]map[string]map[string]*channel.Client // A lookup takes `channelID` followed by `orgName` and `username`.
var EventClientInstances map[string]map[string]map[string]*event.Client     // A lookup takes `channelID` followed by `orgName` and `username`.
var LedgerClientInstances map[string]map[string]map[string]*ledger.Client   // A lookup takes `channelID` followed by `orgName` and `username`.
var KeySwitchKeys keySwitchKeys                                             // The keys to be used in the key switch process
