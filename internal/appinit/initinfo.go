package appinit

import (
	"io/ioutil"

	errors "github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

// InitInfo is the Go struct for contents in init.yaml.
type InitInfo struct {
	Users      map[string]*OrgInfo       `yaml:"users"`
	Channels   map[string]*ChannelInfo   `yaml:"channels"`
	Chaincodes map[string]*ChaincodeInfo `yaml:"chaincodes"`
}

// OperatingIdentity represents the client / user that performs the operation.
type OperatingIdentity struct {
	OrgName string `yaml:"orgName"` // The name of the organization to which the user that performs the operation belongs
	UserID  string `yaml:"userID"`  // The ID of the user
}

// OrgInfo contains the users of an org. During the init process, a resource management client, MSP client and channel client will be created for each ID in it.
type OrgInfo struct {
	Name     string   `yaml:"name"`     // The name of the organization
	AdminIDs []string `yaml:"adminIDs"` // The IDs of the admin users of the organization
	UserIDs  []string `yaml:"userIDs"`  // The IDs of the normal users of the organization
}

// ChannelInfo needed to create a channel.
type ChannelInfo struct {
	Participants map[string]*OperatingIdentity `yaml:"participants"` // The participating organization names -> the operating user. Peers in the organizations will be joined to the channel.
	Configs      []*ChannelConfigInfo          `yaml:"configs"`      // The configurations that should be applied to the channel
}

// ChannelConfigInfo contains info about a channel config transaction and the user that is used to apply the config.
type ChannelConfigInfo struct {
	Path    string `yaml:"path"`    // The path to the config transaction file
	OrgName string `yaml:"orgName"` // The name of the organization to which the operating user belongs
	UserID  string `yaml:"userID"`  // The ID of the operating user
}

// ChaincodeInfo contains info about a chaincode as well as the installation and instantiation schemes.
type ChaincodeInfo struct {
	ID             string                                 `yaml:"id"`             // The ID of the chaincode
	Version        string                                 `yaml:"version"`        // The version of the chaincode
	Path           string                                 `yaml:"path"`           // The path to the chaincode. Chaincode files should be in ${GoPath}/src/${Path}.
	GoPath         string                                 `yaml:"goPath"`         // The GoPath of the chaincode. Chaincode files dshould be in ${GoPath}/src/${Path}.
	Installations  map[string]*OperatingIdentity          `yaml:"installations"`  // The organization that is to be installed with the chaincode -> the operating user
	Instantiations map[string]*ChaincodeInstantiationInfo `yaml:"instantiations"` // The instantiation info of the chaincode
}

// ChaincodeInstantiationInfo needed to instantiate a chaincode.
type ChaincodeInstantiationInfo struct {
	Policy   string   `yaml:"policy"`   // The instantiation policy
	InitArgs []string `yaml:"initArgs"` // The instantiation arguments
	OrgName  string   `yaml:"orgName"`  // The name of the organization of the operating user
	UserID   string   `yaml:"userID"`   // The ID of the operating user
}

// LoadInitInfo loads the init config file (in YAML) which contains info needed during the init process.
//
// Parameters:
//   the path to the config file
//
// Returns:
//   the `InitInfo` struct containing the info needed during the init process
func LoadInitInfo(configFilePath string) (ret InitInfo, err error) {
	yamlStr, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		err = errors.Wrap(err, "读取初始化配置文件失败")
		return
	}

	err = yaml.Unmarshal(yamlStr, &ret)
	if err != nil {
		err = errors.Wrap(err, "解析 YAML 文件时出现错误")
		return
	}

	return
}
