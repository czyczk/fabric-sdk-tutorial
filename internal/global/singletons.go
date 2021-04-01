package global

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/tjfoc/gmsm/sm2"
)

var SDKInstance *fabsdk.FabricSDK
var ResMgmtClientInstances map[string]map[string]*resmgmt.Client            // A lookup takes `orgName` followed by `username`.
var MSPClientInstances map[string]map[string]*msp.Client                    // A lookup takes `orgName` followed by `username`.
var ChannelClientInstances map[string]map[string]map[string]*channel.Client // A lookup takes `channelID` followed by `orgName` and `username`.
var LedgerClientInstances map[string]map[string]map[string]*ledger.Client   // A lookup takes `channelID` followed by `orgName` and `username`.
var SM2PrivateKey *sm2.PrivateKey                                           // SM2 private key of the current user
var SM2PublicKey *sm2.PublicKey                                             // SM2 public key of the current user
