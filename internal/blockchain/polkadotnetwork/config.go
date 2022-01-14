package polkadotnetwork

import (
	"io/ioutil"

	"github.com/pkg/errors"
)

type PolkadotNetworkConfig struct {
	Organizations map[string]organization `yaml:"organizations"`
	APIPrefix     string                  `yaml:"apiPrefix"`
	Chaincodes    map[string]chaincode    `yaml:"chaincodes"`
}

type organization map[string]*user

type user struct {
	Address string `yaml:"address"`
}

type chaincode struct {
	Address string `yaml:"address"`
	ABIPath string `yaml:"abiPath"`
}

func (c *PolkadotNetworkConfig) GetUserAddress(orgName string, userID string) string {
	return c.Organizations[orgName][userID].Address
}

func (c *PolkadotNetworkConfig) GetChaincodeAddress(chaincodeID string) string {
	return c.Chaincodes[chaincodeID].Address
}

func (c *PolkadotNetworkConfig) GetChaincodeABI(chaincodeID string) (string, error) {
	bytes, err := ioutil.ReadFile(c.Chaincodes[chaincodeID].ABIPath)
	if err != nil {
		return "", errors.Wrap(err, "无法读取链码 ABI")
	}

	return string(bytes), nil
}
