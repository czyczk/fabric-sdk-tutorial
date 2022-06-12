package appinit

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/global"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/networkinfo"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	errors "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
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

// InitApp instantiates clients, configure channels and chaincodes according to the init info
func InitApp(initInfo InitInfo) error {
	if global.BlockchainType == blockchain.Fabric {
		var initInfo = initInfo.(*FabricInitInfo)

		// Make sure the SDK instance is instantiated
		sdk := global.FabricSDKInstance
		if sdk == nil {
			return fmt.Errorf("无法初始化应用: Fabric SDK 未实例化")
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
	} else if global.BlockchainType == blockchain.Polkadot {
		// Register the logged-in user in the HTTP API
		if err := RegisterPolkadotUsers(global.PolkadotNetworkConfig.Organizations, global.PolkadotNetworkConfig.APIPrefix); err != nil {
			return err
		}
	}

	// Configure chaincodes if the chaincodes section is specified
	if global.BlockchainType == blockchain.Fabric {
		var initInfo = initInfo.(*FabricInitInfo)

		if initInfo.Chaincodes != nil {
			// The target function receives an argument of type `map[string]ChaincodeInfo`
			// Thus a manual type conversion is needed
			chaincodes := make(map[string]ChaincodeInfo)
			for k, v := range initInfo.Chaincodes {
				chaincodes[k] = v
			}

			if err := configureChaincodes(chaincodes); err != nil {
				return err
			}
		}
	} else if global.BlockchainType == blockchain.Polkadot {
		var initInfo = initInfo.(*PolkadotInitInfo)

		if initInfo.Chaincodes != nil {
			// The target function receives an argument of type `map[string]ChaincodeInfo`
			// Thus a manual type conversion is needed
			chaincodes := make(map[string]ChaincodeInfo)
			for k, v := range initInfo.Chaincodes {
				chaincodes[k] = v
			}

			if err := configureChaincodes(chaincodes); err != nil {
				return err
			}
		}
	} else {
		return fmt.Errorf("无法初始化应用: 未实现的区块链类型")
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
func configureChaincodes(chaincodes map[string]ChaincodeInfo) error {
	// Install and instantiate each chaincode in the list
	for ccID, chaincodeInfo := range chaincodes {
		if global.BlockchainType == blockchain.Fabric {
			var chaincodeInfo = chaincodeInfo.(*FabricChaincodeInfo)

			// Perform installations for the chaincode
			for orgName, operatingIdentity := range chaincodeInfo.Installations {
				// For each organization ($orgName), install the chaincode using the operating identity ($operatingIdentity)
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
		} else if global.BlockchainType == blockchain.Polkadot {
			var chaincodeInfo = chaincodeInfo.(*PolkadotChaincodeInfo)
			log.Printf("开始实例化 Polkadot 合约 '%v'...\n", chaincodeInfo.ID)

			httpClient := http.DefaultClient

			abi, err := global.PolkadotNetworkConfig.GetChaincodeABI(ccID)
			if err != nil {
				return err
			}

			// Read the ONLY wasm file in the chaincode path.
			// If there are multiple ones or none, return an error
			var wasmFiles []string
			err = filepath.WalkDir(chaincodeInfo.Path, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					return err
				}

				if d.IsDir() {
					return nil
				}

				if matched, err := filepath.Match("*.wasm", filepath.Base(path)); err != nil {
					return err
				} else if matched {
					wasmFiles = append(wasmFiles, path)
				}

				return nil
			})
			if err != nil {
				return errors.Wrapf(err, "无法遍历合约文件夹 '%v'", chaincodeInfo.Path)
			}
			if len(wasmFiles) != 1 {
				return fmt.Errorf("合约文件夹 '%v' 中包含 0 个或多于 1 个 wasm 文件", chaincodeInfo.Path)
			}

			// Read the only wasm file
			wasmBytes, err := ioutil.ReadFile(wasmFiles[0])
			if err != nil {
				return errors.Wrapf(err, "无法读取合约文件 '%v'", wasmFiles[0])
			}

			ctorFuncName := "default"
			ctorArgsBytes := []byte{'['}
			for i, arg := range chaincodeInfo.Instantiation.InitArgs {
				ctorArgsBytes = append(ctorArgsBytes, []byte(arg)...)
				if i < len(chaincodeInfo.Instantiation.InitArgs)-1 {
					ctorArgsBytes = append(ctorArgsBytes, ',')
				}
			}
			ctorArgsBytes = append(ctorArgsBytes, ']')

			// Perform installations for the chaincode
			for orgName, operatingIdentity := range chaincodeInfo.Installations {
				// For each organization ($orgName), install the chaincode using the operating identity ($operatingIdentity)
				signerAddress, err := global.PolkadotNetworkConfig.GetUserAddress(orgName, operatingIdentity.UserID)
				if err != nil {
					return err
				}

				result, err := sendTxToInstantiateChaincode(global.PolkadotNetworkConfig.APIPrefix, httpClient, abi, wasmBytes, signerAddress, ctorFuncName, string(ctorArgsBytes))
				if err != nil {
					return errors.Wrapf(err, "无法实例化 Polkadot 合约 '%v'", chaincodeInfo.ID)
				}

				log.Printf("已实例化 Polkadot 合约 '%v'。\n", chaincodeInfo.ID)
				log.Printf("合约地址: %v\n", result.Address)
				log.Printf("交易 ID: %v\n", result.TxExecutionResult.TxHash)
				log.Printf("区块 ID: %v\n", result.InBlockStatus.InBlock)
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
func RegisterPolkadotUsers(orgMap map[string]networkinfo.PolkadotOrganization, apiPrefix string) error {
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
