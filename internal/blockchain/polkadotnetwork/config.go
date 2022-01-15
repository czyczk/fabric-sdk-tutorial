package polkadotnetwork

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
)

type PolkadotNetworkConfig struct {
	Organizations map[string]organization `yaml:"organizations"`
	APIPrefix     string                  `yaml:"apiPrefix"`
	Chaincodes    map[string]chaincode    `yaml:"chaincodes"`
}

type organization struct {
	Users map[string]user `yaml:"users"`
}

type user struct {
	Address string `yaml:"address"`
}

type chaincode struct {
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
