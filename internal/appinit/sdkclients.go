package appinit

import (
	"fmt"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/global"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	errors "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// InstantiateResMgmtClient creates a resource management client for the user of the org as specified. The client will be available as singletons in `global.ResMgmtClientInstances`.
//
// Parameters:
//   organization name
//   user ID
func InstantiateResMgmtClient(orgName, userID string) error {
	if global.ResMgmtClientInstances == nil {
		global.ResMgmtClientInstances = make(map[string]map[string]*resmgmt.Client)
	}

	if global.ResMgmtClientInstances[orgName] == nil {
		global.ResMgmtClientInstances[orgName] = make(map[string]*resmgmt.Client)
	}

	if global.ResMgmtClientInstances[orgName][userID] != nil {
		return fmt.Errorf("%v@%v 的资源管理客户端已实例化", userID, orgName)
	}

	// Create a client context using the SDK instance and the init info
	clientContext := global.SDKInstance.Context(fabsdk.WithUser(userID), fabsdk.WithOrg(orgName))
	if clientContext == nil {
		return fmt.Errorf("无法为 %v@%v 创建客户端环境", userID, orgName)
	}

	// Create a resource management client instance using the client context
	resMgmtClient, err := resmgmt.New(clientContext)
	if err != nil {
		return errors.Wrap(err, "无法创建资源管理端")
	}

	global.ResMgmtClientInstances[orgName][userID] = resMgmtClient

	return nil
}

// InstantiateMSPClient creates an MSP client for the user of the org as specified. The MSP client will be available as singletons in `global.MSPClientInstances`.
// Parameters:
//   organization name
//   user ID
func InstantiateMSPClient(orgName, userID string) error {
	if global.MSPClientInstances == nil {
		global.MSPClientInstances = make(map[string]map[string]*msp.Client)
	}

	if global.MSPClientInstances[orgName] == nil {
		global.MSPClientInstances[orgName] = make(map[string]*msp.Client)
	}

	if global.MSPClientInstances[orgName][userID] != nil {
		return fmt.Errorf("%v@%v 的 MSP 客户端已实例化", userID, orgName)
	}

	// Create a client context using the SDK instance and the init info
	clientCtx := global.SDKInstance.Context(fabsdk.WithUser(userID), fabsdk.WithOrg(orgName))
	if clientCtx == nil {
		return fmt.Errorf("无法为 %v@%v 创建客户端环境", userID, orgName)
	}

	// Create an MSP client
	mspClient, err := msp.New(clientCtx, msp.WithOrg(orgName))
	if err != nil {
		return errors.Wrap(err, "无法创建 MSP 客户端")
	}

	global.MSPClientInstances[orgName][userID] = mspClient

	return nil
}

// InstantiateChannelClient creates a channel client on the specified channel for the specified user in the specified org. The channel client will be available as singletons in `global.ChannelClientInstances`.
// Parameters:
//   initialized Fabric SDK instance
//   channel ID
//   organization name
//   user ID
func InstantiateChannelClient(sdk *fabsdk.FabricSDK, channelID, orgName, userID string) error {
	if global.ChannelClientInstances == nil {
		global.ChannelClientInstances = make(map[string]map[string]map[string]*channel.Client)
	}

	if global.ChannelClientInstances[channelID] == nil {
		global.ChannelClientInstances[channelID] = make(map[string]map[string]*channel.Client)
	}

	if global.ChannelClientInstances[channelID][orgName] == nil {
		global.ChannelClientInstances[channelID][orgName] = make(map[string]*channel.Client)
	}

	if global.ChannelClientInstances[channelID][orgName][userID] != nil {
		return fmt.Errorf("%v@%v 在通道 '%v' 上的通道客户端已实例化", userID, orgName, channelID)
	}

	// Creates a channel client instance. Channel clients can query chaincode, execute chaincode and register chaincode events on specific channel.
	clientCtx := sdk.ChannelContext(channelID, fabsdk.WithUser(userID), fabsdk.WithOrg(orgName))
	channelClient, err := channel.New(clientCtx)
	if err != nil {
		return errors.Wrapf(err, "无法在通道 '%v' 上为 %v@%v 创建通道客户端", channelID, userID, orgName)
	}
	global.ChannelClientInstances[channelID][orgName][userID] = channelClient

	log.Printf("已在通道 '%v' 上为 %v@%v 创建通道客户端。", channelID, userID, orgName)

	return nil
}

// InstantiateLedgerClient creates a ledger client for the specified channel, org and user ID. The ledger client will be available as singletons in `global.LedgerClientInstances`.
// Parameters:
//   initialized Fabric SDK instance
//   channel ID
//   organization name
//   user ID
func InstantiateLedgerClient(sdk *fabsdk.FabricSDK, channelID, orgName, userID string) error {
	if global.LedgerClientInstances == nil {
		global.LedgerClientInstances = make(map[string]map[string]map[string]*ledger.Client)
	}

	if global.LedgerClientInstances[channelID] == nil {
		global.LedgerClientInstances[channelID] = make(map[string]map[string]*ledger.Client)
	}

	if global.LedgerClientInstances[channelID][orgName] == nil {
		global.LedgerClientInstances[channelID][orgName] = make(map[string]*ledger.Client)
	}

	if global.LedgerClientInstances[channelID][orgName][userID] != nil {
		return fmt.Errorf("%v@%v 在通道 '%v' 上的账本客户端已实例化", userID, orgName, channelID)
	}

	// Creates a ledger client instance. Ledger clients can query blocks and transactions on the channel.
	clientCtx := sdk.ChannelContext(channelID, fabsdk.WithUser(userID), fabsdk.WithOrg(orgName))
	ledgerClient, err := ledger.New(clientCtx)
	if err != nil {
		return errors.Wrapf(err, "无法在通道 '%v' 上为 %v@%v 创建账本客户端", channelID, userID, orgName)
	}
	global.LedgerClientInstances[channelID][orgName][userID] = ledgerClient

	log.Printf("已在通道 '%v' 上为 %v@%v 创建账本客户端。", channelID, userID, orgName)

	return nil
}
