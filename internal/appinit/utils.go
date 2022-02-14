package appinit

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/polkadotnetwork"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/global"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	errors "github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

func ParseBlockchainType(blockchainTypeStr string) error {
	switch strings.ToLower(blockchainTypeStr) {
	case "fabric":
		global.BlockchainType = blockchain.Fabric
	case "polkadot":
		global.BlockchainType = blockchain.Polkadot
	default:
		return fmt.Errorf("未知的区块链类型")
	}

	return nil
}

// SetupSDK creates a Fabric SDK instance from the specified config file. The SDK instance will be available as `global.SDKInstance`.
//
// Parameters:
//   the path to the config file
func SetupSDK(configFilePath string) error {
	configProvider := config.FromFile(configFilePath)
	sdk, err := fabsdk.New(configProvider)
	if err != nil {
		return errors.Wrap(err, "初始化 Fabric SDK 失败")
	}
	global.FabricSDKInstance = sdk

	return nil
}

// LoadPolkadotNetworkConfig creates a `polkadotnetwork.PolkadotNetworkConfig` object from the specified config file. The config object will be available as `global.PolkadotNetworkConfig`.
//
// Parameters:
//   the path to the config file
func LoadPolkadotNetworkConfig(configFilePath string) error {
	yamlStr, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return errors.Wrap(err, "读取 Polkadot 网络配置文件失败")
	}

	var config *polkadotnetwork.PolkadotNetworkConfig
	err = yaml.Unmarshal(yamlStr, &config)
	if err != nil {
		return errors.Wrap(err, "解析 YAML 文件时出现错误")
	}

	global.PolkadotNetworkConfig = config
	return nil
}

// InitApp instantiates clients, configure channels and chaincodes according to the init info
func InitApp(initInfo *InitInfo) error {
	// Make sure the SDK instance is instantiated
	sdk := global.FabricSDKInstance
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

// RegisterPolkadotUsers registers the polkadot users in the Polkadot network config in the HTTP API.
func RegisterPolkadotUsers(orgMap map[string]polkadotnetwork.Organization, apiPrefix string) error {
	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	endpoint := apiPrefix + "/keyring/from-uri"

	for orgName, orgInfo := range orgMap {
		for userID, userInfo := range orgInfo.Users {
			// Prepare a POST form
			form := url.Values{}
			form.Set("phrase", userInfo.Phrase)

			formEncoded := form.Encode()
			req, err := http.NewRequest("POST", endpoint, strings.NewReader(formEncoded))
			if err != nil {
				return errors.Wrapf(err, "无法注册 Polkadot 用户 '%v@%v'", userID, orgName)
			}

			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Add("Content-Length", strconv.Itoa(len(formEncoded)))

			// Perform a POST request
			resp, err := client.Do(req)
			if err != nil {
				return errors.Wrap(err, "无法调用合约")
			}
			defer resp.Body.Close()

			// Process the response.
			// 200 -> The response body doesn't matter.
			// !200 -> Error registering the user. Resp body as the error message.
			if resp.StatusCode != 200 {
				respBodyBytes, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return errors.Wrapf(err, "无法注册 Polkadot 用户 '%v@%v': 无法获取注册结果", userID, orgName)
				}

				return fmt.Errorf("无法注册 Polkadot 用户 '%v@%v': %v", userID, orgName, string(respBodyBytes))
			}
		}
	}

	return nil
}
