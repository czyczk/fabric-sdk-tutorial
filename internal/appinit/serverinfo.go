package appinit

import (
	"io/ioutil"

	errors "github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

// ServerInfo is the Go struct for contents in serve.yaml.
type ServerInfo struct {
	User              *OperatingIdentity     `yaml:"user"`
	Channels          []string               `yaml:"channels"`
	Port              int                    `yaml:"port"`
	LocalDBSourceName string                 `yaml:"localDBSourceName"`
	IPFSAPI           string                 `yaml:"ipfsAPI"`
	IsKeySwitchServer bool                   `yaml:"isKeySwitchServer"`
	IsRegulator       bool                   `yaml:"isRegulator"`
	KeySwitchKeys     *KeySwitchKeyLocations `yaml:"keySwitchKeys"`
	ShowTimingLogs    bool                   `yaml:"showTimingLogs"`
}

// LoadServerInfo loads the server config file (in YAML) which contains info needed to start a server.
//
// Parameters:
//   the path to the config file
//
// Returns:
//   the `ServerInfo` struct containing the info needed to start a server
func LoadServerInfo(configFilePath string) (ret ServerInfo, err error) {
	yamlStr, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		err = errors.Wrap(err, "读取服务器配置文件失败")
		return
	}

	err = yaml.Unmarshal(yamlStr, &ret)
	if err != nil {
		err = errors.Wrap(err, "解析 YAML 文件时出现错误")
		return
	}

	return
}
