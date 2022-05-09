package networkinfo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type PolkadotNetworkConfig struct {
	Organizations map[string]PolkadotOrganization `yaml:"organizations"`
	APIPrefix     string                          `yaml:"apiPrefix"`
	Chaincodes    map[string]polkaodtChaincode    `yaml:"chaincodes"`
}

type PolkadotOrganization struct {
	Users map[string]PolkadotUser `yaml:"users"`
}

type PolkadotUser struct {
	Phrase  string `yaml:"phrase"`
	Address string `yaml:"address"`
}

type polkaodtChaincode struct {
	Address string `yaml:"address"`
	ABIPath string `yaml:"abiPath"`
}

func (c *PolkadotNetworkConfig) GetUserAddress(orgName string, userID string) (string, error) {
	org, ok := c.Organizations[orgName]
	if !ok {
		return "", fmt.Errorf("无法获取用户地址: 未找到组织 '%v'", orgName)
	}

	user, ok := org.Users[userID]
	if !ok {
		return "", fmt.Errorf("无法获取用户地址: 未找到用户 '%v'", userID)
	}

	return user.Address, nil
}

func (c *PolkadotNetworkConfig) GetChaincodeAddress(chaincodeID string) string {
	return c.Chaincodes[chaincodeID].Address
}

func (c *PolkadotNetworkConfig) GetChaincodeABI(chaincodeID string) (string, error) {
	chaincode, ok := c.Chaincodes[chaincodeID]
	if !ok {
		return "", fmt.Errorf("无法获取链码 ABI: 未找到链码 '%v'", chaincodeID)
	}

	abiBytes, err := ioutil.ReadFile(chaincode.ABIPath)
	if err != nil {
		return "", errors.Wrap(err, "无法读取链码 ABI")
	}

	// Must be converted to compact JSON or HTTP requests containing the ABI will always timeout
	compactAbiBytes := bytes.NewBuffer([]byte{})
	if err := json.Compact(compactAbiBytes, abiBytes); err != nil {
		return "", errors.Wrap(err, "无法读取链码 ABI")
	}

	return compactAbiBytes.String(), nil
}

// ParsePolkadotNetworkConfig creates a `PolkadotNetworkConfig` object from the specified config file.
//
// Parameters:
//   the path to the config file
//
// Returns:
//   an object containing the network config info
func ParsePolkadotNetworkConfig(configFilePath string) (*PolkadotNetworkConfig, error) {
	yamlStr, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "读取 Polkadot 网络配置文件失败")
	}

	var config *PolkadotNetworkConfig
	err = yaml.Unmarshal(yamlStr, &config)
	if err != nil {
		return nil, errors.Wrap(err, "解析 YAML 文件时出现错误")
	}

	return config, nil
}
