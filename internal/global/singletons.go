package global

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

type ResMgmtClients struct {
	AdminResMgmtClient *resmgmt.Client
	User1ResMgmtClient *resmgmt.Client
}

type MSPClients struct {
	AdminMSPClient *msp.Client
	User1MSPClient *msp.Client
}

var SDKInstance *fabsdk.FabricSDK
var ResMgmtClientInstances *ResMgmtClients
var MSPClientInstances *MSPClients
