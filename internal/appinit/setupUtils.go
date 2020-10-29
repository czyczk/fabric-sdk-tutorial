package appinit

import (
	"fmt"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/global"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	providersmsp "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/gopackager"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/policydsl"
	log "github.com/sirupsen/logrus"
)

// SetupSDK creates a Fabric SDK instance from the specified config file(s).
func SetupSDK(configFile string) error {
	configProvider := config.FromFile(configFile)
	sdk, err := fabsdk.New(configProvider)
	if err != nil {
		return fmt.Errorf("failed initializing Fabric SDK: %v", err)
	}
	global.SDKInstance = sdk

	return nil
}

// InstantiateResMgmtClients creates resource management clients for both the admin and the user of the org as singletons.
func InstantiateResMgmtClients(username, orgName string) error {
	if global.ResMgmtClientInstances == nil {
		global.ResMgmtClientInstances = make(map[string]map[string]*resmgmt.Client)
	}

	if global.ResMgmtClientInstances[orgName] == nil {
		global.ResMgmtClientInstances[orgName] = make(map[string]*resmgmt.Client)
	}

	if global.ResMgmtClientInstances[orgName][username] != nil {
		return fmt.Errorf("res mgmt client for %v.%v already instantiated", username, orgName)
	}

	// Create client contexts using the initialized SDK instance and the init info
	clientContext := global.SDKInstance.Context(fabsdk.WithUser(username), fabsdk.WithOrg(orgName))
	if clientContext == nil {
		return fmt.Errorf("failed to create a client context for %v.%v", username, orgName)
	}

	// Create an admin resource management client instance using the admin client context.
	resMgmtClient, err := resmgmt.New(clientContext)
	if err != nil {
		return fmt.Errorf("failed to create a resource management client: %v", err)
	}

	global.ResMgmtClientInstances[orgName][username] = resMgmtClient

	return nil
}

// InstantiateMSPClients creates MSP clients for both the org admin and user.
func InstantiateMSPClients(username, orgName string) error {
	if global.MSPClientInstances == nil {
		global.MSPClientInstances = make(map[string]map[string]*msp.Client)
	}

	if global.MSPClientInstances[orgName] == nil {
		global.MSPClientInstances[orgName] = make(map[string]*msp.Client)
	}

	if global.MSPClientInstances[orgName][username] != nil {
		return fmt.Errorf("MSP client for %v.%v already instantiated", username, orgName)
	}

	// Create clients contexts using the initialized SDK instance and the init info
	clientCtx := global.SDKInstance.Context(fabsdk.WithUser(username), fabsdk.WithOrg(orgName))
	if clientCtx == nil {
		return fmt.Errorf("failed to create a client context for %v.%v", username, orgName)
	}

	// Create an MSP client
	mspClient, err := msp.New(clientCtx, msp.WithOrg(orgName))
	if err != nil {
		return fmt.Errorf("failed to create an MSP client: %v", err)
	}

	global.MSPClientInstances[orgName][username] = mspClient

	return nil
}

// CreateChannel creates a channel in the network using the channel creation transaction files.
func CreateChannel(channelInfo *ChannelInitInfo, orgInfo *OrgInitInfo) error {
	// Make sure the admin MSP client is initialized
	adminMSPClient := global.MSPClientInstances[orgInfo.OrgName][orgInfo.AdminID]
	if adminMSPClient == nil {
		return fmt.Errorf("admin MSP client not initialized")
	}

	// Make sure the admin resource management client is initialized
	adminResMgmtClient := global.ResMgmtClientInstances[orgInfo.OrgName][orgInfo.AdminID]
	if adminResMgmtClient == nil {
		return fmt.Errorf("admin resource management client not initialized")
	}

	// Create signing identity from the MSP client
	adminIdentity, err := adminMSPClient.GetSigningIdentity(orgInfo.AdminID)
	if err != nil {
		return fmt.Errorf("failed to get the admin signing identity: %v", err)
	}

	// Create a "save channel" request
	channelReq := resmgmt.SaveChannelRequest{
		ChannelID:         channelInfo.ChannelID,
		ChannelConfigPath: channelInfo.ChannelConfigPath,
		SigningIdentities: []providersmsp.SigningIdentity{adminIdentity},
	}

	// Get the channel creation response with a transaction ID
	_, err = adminResMgmtClient.SaveChannel(channelReq,
		resmgmt.WithRetry(retry.DefaultResMgmtOpts),
		resmgmt.WithOrdererEndpoint(orgInfo.OrdererEndpoint))
	if err != nil {
		return fmt.Errorf("failed to create channel for the app: %v", err)
	}

	log.Println("Channel created successfully.")

	return nil
}

// JoinChannel joins the peers of the specified org to the specified channel.
func JoinChannel(channelInfo *ChannelInitInfo, orgInfo *OrgInitInfo) error {
	adminResMgmtClient := global.ResMgmtClientInstances[orgInfo.OrgName][orgInfo.AdminID]
	if adminResMgmtClient == nil {
		return fmt.Errorf("admin res mgmt client of %v not initialized", orgInfo.OrgName)
	}

	// Peers are not specified in options, so it will join all peers that belong to the client's MSP.
	err := adminResMgmtClient.JoinChannel(channelInfo.ChannelID, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint(orgInfo.OrdererEndpoint))
	if err != nil {
		return fmt.Errorf("failed to join peers to the client: %v", err)
	}

	log.Printf("Peers of %v joined to the channel successfully.\n", orgInfo.OrgName)
	return nil
}

// InstallAndInstantiateCC installs the specified chaincode on all the peers of the specified org and instantiate it on the peers.
func InstallAndInstantiateCC(sdk *fabsdk.FabricSDK, chaincodeInfo *ChaincodeInitInfo, channelInfo *ChannelInitInfo, orgInfo *OrgInitInfo) (*channel.Client, error) {
	adminResMgmtClient := global.ResMgmtClientInstances[orgInfo.OrgName][orgInfo.AdminID]
	if adminResMgmtClient == nil {
		return nil, fmt.Errorf("admin res mgmt client of %v not initialized", orgInfo.OrgName)
	}

	log.Printf("Starting to install chaincode of ID '%v' for peers in org %v...\n", chaincodeInfo.ChaincodeID, orgInfo.OrgName)

	// Create a new Go chaincode package
	ccPkg, err := gopackager.NewCCPackage(chaincodeInfo.ChaincodePath, chaincodeInfo.ChaincodeGoPath)
	if err != nil {
		return nil, fmt.Errorf("failed creating chaincode package for chaincode ID %v: %v", chaincodeInfo.ChaincodeID, err)
	}

	// Make a request containing parameters to install the chaincode
	installCCReq := resmgmt.InstallCCRequest{
		Name:    chaincodeInfo.ChaincodeID,
		Path:    chaincodeInfo.ChaincodePath,
		Version: chaincodeInfo.ChaincodeVersion,
		Package: ccPkg,
	}

	// Install the chaincode. No peers are specified here so it will be installed on all the peers in the org.
	_, err = adminResMgmtClient.InstallCC(installCCReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		return nil, fmt.Errorf("failed installing chaincode for peers in org %v: %v", orgInfo.OrgName, err)
	}

	log.Printf("Chaincode of ID '%v' installed for peer in org %v.\n", chaincodeInfo.ChaincodeID, orgInfo.OrgName)
	log.Printf("Starting to instantiate chaincode of ID '%v' for peers in org %v...\n", chaincodeInfo.ChaincodeID, orgInfo.OrgName)

	// Parse the endorsement policy
	ccPolicy, err := policydsl.FromString(chaincodeInfo.Policy)
	if err != nil {
		return nil, fmt.Errorf("failed instantiating chaincode of ID '%v': %v", chaincodeInfo.ChaincodeID, err)
	}

	// Make a request containing parameters to instantiate the chaincode
	instantiateCCReq := resmgmt.InstantiateCCRequest{
		Name:    chaincodeInfo.ChaincodeID,
		Path:    chaincodeInfo.ChaincodePath,
		Version: chaincodeInfo.ChaincodeVersion,
		Args:    [][]byte{[]byte("init")},
		Policy:  ccPolicy,
	}

	// Instantiate the chaincode for all peers in the org
	_, err = adminResMgmtClient.InstantiateCC(channelInfo.ChannelID, instantiateCCReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		return nil, fmt.Errorf("failed instantiating chaincode of ID %v for peers in org %v: %v", chaincodeInfo.ChaincodeID, orgInfo.OrgName, err)
	}

	log.Printf("Chaincode of ID '%v' instantiated for peers in org %v.\n", chaincodeInfo.ChaincodeID, orgInfo.OrgName)

	// Returns a channel client instance. Channel clients can query chaincode, execute chaincode and register chaincode events on specific channel.
	clientChannelCtx := sdk.ChannelContext(channelInfo.ChannelID, fabsdk.WithUser(orgInfo.AdminID), fabsdk.WithOrg(orgInfo.OrgName))
	channelClient, err := channel.New(clientChannelCtx)
	if err != nil {
		return nil, fmt.Errorf("failed creating channel client: %v", err)
	}

	log.Println("Channel client created.")

	return channelClient, nil
}
