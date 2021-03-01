package global

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

var SDKInstance *fabsdk.FabricSDK
var ResMgmtClientInstances map[string]map[string]*resmgmt.Client            // A lookup takes `orgName` followed by `username`.
var MSPClientInstances map[string]map[string]*msp.Client                    // A lookup takes `orgName` followed by `username`.
var ChannelClientInstances map[string]map[string]map[string]*channel.Client // A lookup takes `channelID` followed by `orgName` and `username`.
var LedgerClientInstances map[string]map[string]map[string]*ledger.Client   // A lookup takes `channelID` followed by `orgName` and `username`.
