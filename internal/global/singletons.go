package global

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

var SDKInstance *fabsdk.FabricSDK
var ResMgmtClientInstances map[string]map[string]*resmgmt.Client
var MSPClientInstances map[string]map[string]*msp.Client
