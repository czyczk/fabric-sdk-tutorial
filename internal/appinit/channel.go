package appinit

import (
	"fmt"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/global"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	providersmsp "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	errors "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// ApplyChannelConfigTx applies a channel config trasaction file to create a channel or configure a channel.
//
// Parameters:
//   channel ID
//   channel config info
func ApplyChannelConfigTx(channelID string, info *ChannelConfigInfo) error {
	// Make sure the MSP client is instantiated
	mspClient := global.MSPClientInstances[info.OrgName][info.UserID]
	if mspClient == nil {
		return fmt.Errorf("%v@%v 的 MSP 客户端未实例化", info.UserID, info.OrgName)
	}

	// Make sure the resource management client is instantiated
	resMgmtClient := global.ResMgmtClientInstances[info.OrgName][info.UserID]
	if resMgmtClient == nil {
		return fmt.Errorf("%v@%v 的资源管理客户端未实例化", info.UserID, info.OrgName)
	}

	// Create signing identity from the MSP client
	signingIdentity, err := mspClient.GetSigningIdentity(info.UserID)
	if err != nil {
		return errors.Wrapf(err, "无法获取 %v@%v 的签名身份", info.UserID, info.OrgName)
	}

	// Create a "save channel" request
	channelReq := resmgmt.SaveChannelRequest{
		ChannelID:         channelID,
		ChannelConfigPath: info.Path,
		SigningIdentities: []providersmsp.SigningIdentity{signingIdentity},
	}

	// Get the channel creation response with a transaction ID
	_, err = resMgmtClient.SaveChannel(channelReq,
		resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		return errors.Wrapf(err, "为通道 '%v' 应用通道配置交易文件 '%v' 失败", channelID, info.Path)
	}

	log.Printf("已为通道 '%v' 应用通道配置交易文件 '%v'。\n", channelID, info.Path)

	return nil
}

// JoinChannel joins the peers of the specified org to the specified channel with the specified operating identity.
//
// Parameters:
//   channel ID
//   organization name
//   operating identity
func JoinChannel(channelID, orgName string, operatingIdentity *OperatingIdentity) error {
	// Check if the res mgmt client of the operating identity is instantiated
	resMgmtClient := global.ResMgmtClientInstances[operatingIdentity.OrgName][operatingIdentity.UserID]
	if resMgmtClient == nil {
		return fmt.Errorf("%v@%v 的资源管理客户端未实例化", operatingIdentity.UserID, operatingIdentity.OrgName)
	}

	// Peers are not specified in options, so it will join all peers that belong to the client's MSP.
	err := resMgmtClient.JoinChannel(channelID, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		return errors.Wrapf(err, "无法将 '%v' 的节点加入通道 '%v'", orgName, channelID)
	}

	log.Printf("已将 '%v' 的节点加入通道 '%v'。\n", operatingIdentity.OrgName, channelID)

	return nil
}
