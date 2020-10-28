package appinit

import (
	"fmt"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/global"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	providersmsp "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"log"
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
func InstantiateResMgmtClients(info *ClientInitInfo) error {
	if global.ResMgmtClientInstances.AdminResMgmtClient != nil {
		log.Println("Admin resource management client already instantiated.")
	} else {
		// Create client contexts using the initialized SDK instance and the init info
		adminClientCtx := global.SDKInstance.Context(fabsdk.WithUser(info.AdminID), fabsdk.WithOrg(info.OrgName))
		if adminClientCtx == nil {
			return fmt.Errorf("failed to create an admin client context")
		}

		// Create an admin resource management client instance using the admin client context.
		adminResMgmtClient, err := resmgmt.New(adminClientCtx)
		if err != nil {
			return fmt.Errorf("failed to create an admin resource management client: %v", err)
		}

		global.ResMgmtClientInstances.AdminResMgmtClient = adminResMgmtClient
	}

	// TODO: Create a user client context and a client

	return nil
}

// InstantiateMSPClients creates MSP clients for both the org admin and user.
func InstantiateMSPClients(info *ClientInitInfo) error {
	// Create clients contexts using the initialized SDK instance and the init info
	adminClientCtx := global.SDKInstance.Context(fabsdk.WithUser(info.AdminID), fabsdk.WithOrg(info.OrgName))
	if adminClientCtx == nil {
		return fmt.Errorf("failed to create an admin client context")
	}
	// TODO: Create a user client context

	// Create an MSP client
	adminMspClient, err := msp.New(adminClientCtx, msp.WithOrg(info.OrgName))
	if err != nil {
		return fmt.Errorf("failed to create an admin MSP client:%v", err)
	}

	// TODO: Create a user MSP client.

	global.MSPClientInstances.AdminMSPClient = adminMspClient

	return nil
}

func CreateChannel(channelInfo *ChannelInitInfo, clientInfo *ClientInitInfo,
	adminMSPClient *msp.Client, adminResMgmtClient *resmgmt.Client) error {
	// Make sure the admin MSP client is initialized
	if adminMSPClient == nil {
		return fmt.Errorf("admin MSP client not initialized")
	}

	// Make sure the admin resource management client is initialized
	if adminResMgmtClient == nil {
		return fmt.Errorf("admin resource management client not initialized")
	}

	// Create signing identity from the MSP client
	adminIdentity, err := adminMSPClient.GetSigningIdentity(clientInfo.AdminID)
	if err != nil {
		return fmt.Errorf("failed to get the admin signing identity: %v", err)
	}

	// Create a "save channel" request
	channelReq := resmgmt.SaveChannelRequest{
		ChannelID: channelInfo.ChannelID,
		ChannelConfigPath: channelInfo.ChannelConfigPath,
		SigningIdentities: []providersmsp.SigningIdentity{adminIdentity},
	}

	// Get the channel creation response with a transaction ID
	_, err = adminResMgmtClient.SaveChannel(channelReq,
		resmgmt.WithRetry(retry.DefaultResMgmtOpts),
		resmgmt.WithOrdererEndpoint(clientInfo.OrdererEndpoint))
	if err != nil {
		return fmt.Errorf("failed to create channel for the app: %v", err)
	}

	log.Println("Channel created successfully.")

	return nil
}
