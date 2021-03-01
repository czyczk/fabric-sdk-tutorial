package appinit

import (
	"fmt"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/global"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	errors "github.com/pkg/errors"
)

// SetupSDK creates a Fabric SDK instance from the specified config file(s). The SDK instance will be available as `global.SDKInstance`.
//
// Parameters:
//   the path to the config file
func SetupSDK(configFilePath string) error {
	configProvider := config.FromFile(configFilePath)
	sdk, err := fabsdk.New(configProvider)
	if err != nil {
		return errors.Wrap(err, "初始化 Fabric SDK 失败")
	}
	global.SDKInstance = sdk

	return nil
}

// InitApp instantiates clients, configure channels and chaincodes according to the init info
func InitApp(initInfo *InitInfo) error {
	// Make sure the SDK instance is instantiated
	sdk := global.SDKInstance
	if sdk == nil {
		return fmt.Errorf("Fabric SDK 未实例化")
	}

	// Res mgmt clients and MSP clients
	if err := instantiateResMgmtClientsAndMSPClients(initInfo.Users); err != nil {
		return err
	}

	// Configure channels if the channels section is specified
	if initInfo.Channels != nil {
		if err := configureChannels(initInfo.Channels); err != nil {
			return err
		}
	}

	// Channel & ledger clients
	channelIDs := make([]string, 0, len(initInfo.Channels))
	for channelID := range initInfo.Channels {
		channelIDs = append(channelIDs, channelID)
	}
	if err := instantiateChannelClients(sdk, initInfo.Users, channelIDs); err != nil {
		return err
	}
	if err := instantiateLedgerClients(sdk, initInfo.Users, channelIDs); err != nil {
		return err
	}

	// Configure chaincodes if the chaincodes section is specified
	if initInfo.Chaincodes != nil {
		if err := configureChaincodes(initInfo.Chaincodes); err != nil {
			return err
		}
	}

	return nil
}

// Instantiates resource management clients and MSP clients for the orgs in the list.
func instantiateResMgmtClientsAndMSPClients(userInfo map[string]*OrgInfo) error {
	for orgName, orgInfo := range userInfo {
		// Clients for each admin ID
		for _, adminID := range orgInfo.AdminIDs {
			if err := InstantiateResMgmtClient(orgName, adminID); err != nil {
				return err
			}

			if err := InstantiateMSPClient(orgName, adminID); err != nil {
				return err
			}
		}

		// Clients for each user ID
		for _, userID := range orgInfo.UserIDs {
			if err := InstantiateResMgmtClient(orgName, userID); err != nil {
				return err
			}

			if err := InstantiateMSPClient(orgName, userID); err != nil {
				return err
			}
		}
	}

	return nil
}

// This function creates and configures channels according to the channel init info and joins the peers to the channels.
func configureChannels(channels map[string]*ChannelInfo) error {
	// Apply the specifications for each of the channels
	for channelID, channelInfo := range channels {
		// Apply channel configs in order
		for _, channelConfigInfo := range channelInfo.Configs {
			if err := ApplyChannelConfigTx(channelID, channelConfigInfo); err != nil {
				return err
			}
		}

		// Join peers to the channel
		for orgName, operatingIdentity := range channelInfo.Participants {
			if err := JoinChannel(channelID, orgName, operatingIdentity); err != nil {
				return err
			}
		}
	}

	return nil
}

// This function installs and instantiates chaincodes according to the init info.
func configureChaincodes(chaincodes map[string]*ChaincodeInfo) error {
	// Install and instantiate each chaincode in the list
	for ccID, chaincodeInfo := range chaincodes {
		// Perform installations for the chaincode
		for orgName, operatingIdentity := range chaincodeInfo.Installations {
			if err := InstallCC(ccID, chaincodeInfo.Version, chaincodeInfo.Path, chaincodeInfo.GoPath, orgName, operatingIdentity); err != nil {
				return err
			}
		}

		// Perform instantiations for the chaincode
		for channelID, instantiationInfo := range chaincodeInfo.Instantiations {
			if err := InstantiateCC(ccID, chaincodeInfo.Path, chaincodeInfo.Version, channelID, instantiationInfo); err != nil {
				return err
			}
		}
	}

	return nil
}

// Instantiates channel clients for users in the list.
func instantiateChannelClients(sdk *fabsdk.FabricSDK, userInfo map[string]*OrgInfo, channelIDs []string) error {
	for orgName, orgInfo := range userInfo {
		for _, channelID := range channelIDs {
			// Channel client for each admin ID
			for _, adminID := range orgInfo.AdminIDs {
				if err := InstantiateChannelClient(sdk, channelID, orgName, adminID); err != nil {
					return err
				}
			}

			// Channel client for each user ID
			for _, userID := range orgInfo.UserIDs {
				if err := InstantiateChannelClient(sdk, channelID, orgName, userID); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// Instantiates ledger clients for users in the list.
func instantiateLedgerClients(sdk *fabsdk.FabricSDK, userInfo map[string]*OrgInfo, channelIDs []string) error {
	for orgName, orgInfo := range userInfo {
		for _, channelID := range channelIDs {
			// Ledger client for each admin ID
			for _, adminID := range orgInfo.AdminIDs {
				if err := InstantiateLedgerClient(sdk, channelID, orgName, adminID); err != nil {
					return err
				}
			}

			// Ledger client for each user ID
			for _, userID := range orgInfo.UserIDs {
				if err := InstantiateLedgerClient(sdk, channelID, orgName, userID); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
