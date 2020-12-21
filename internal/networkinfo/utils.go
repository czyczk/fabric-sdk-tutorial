package networkinfo

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab"
)

// ParseOrderers parses the "orderers" section of the config from the config backend instance provided and returns a map of Orderer instances.
func ParseOrderers(configBackend core.ConfigBackend) (result map[string]Orderer, err error) {
	// Lookup the "orderers" section
	configBackendOrderers, ok := configBackend.Lookup("orderers")
	if !ok {
		err = fmt.Errorf("error parsing orderers")
		return
	}

	// Parse the section into a map
	orderersMap := make(map[string]fab.OrdererConfig)
	orderersBytes, err := json.Marshal(configBackendOrderers)
	if err != nil {
		return
	}

	err = json.Unmarshal(orderersBytes, &orderersMap)
	if err != nil {
		return
	}

	// Extract useful info for clients
	result = make(map[string]Orderer)
	for k, v := range orderersMap {
		result[k] = Orderer{Name: k, URL: v.URL}
	}

	return
}

// ParseOrganizations parses the "organizations" section of the config from the config backend instance provided and returns a map of Organization instances.
func ParseOrganizations(configBackend core.ConfigBackend) (result map[string]Organization, err error) {
	// Lookup the "organizations" section
	configBackendOrganizations, ok := configBackend.Lookup("organizations")
	if !ok {
		err = fmt.Errorf("error parsing organizations")
	}

	// Parse the section into a map
	organizationsMap := make(map[string]fab.OrganizationConfig)
	organizationsBytes, err := json.Marshal(configBackendOrganizations)
	if err != nil {
		return
	}

	err = json.Unmarshal(organizationsBytes, &organizationsMap)
	if err != nil {
		return
	}

	// Extract useful info for clients
	result = make(map[string]Organization)
	for k, v := range organizationsMap {
		var users []string
		for userName, _ := range v.Users {
			users = append(users, userName)
		}

		result[k] = Organization{
			Name:       k,
			MSPID:      v.MSPID,
			CryptoPath: v.CryptoPath,
			Peers:      v.Peers,
			Users:      users,
		}
	}

	return
}

// ParsePeers parses the "peers" section of the config from the config backend instance provided and returns a map of Peer instances.
func ParsePeers(configBackend core.ConfigBackend) (result map[string]Peer, err error) {
	// Lookup the "peers" section
	configBackendOrganizations, ok := configBackend.Lookup("peers")
	if !ok {
		err = fmt.Errorf("error parsing peers")
	}

	// Parse the section into a map
	peersMap := make(map[string]fab.PeerConfig)
	peersBytes, err := json.Marshal(configBackendOrganizations)
	if err != nil {
		return
	}

	err = json.Unmarshal(peersBytes, &peersMap)
	if err != nil {
		return
	}

	// Extract useful info for clients
	result = make(map[string]Peer)
	for k, v := range peersMap {
		result[k] = Peer{
			Name: k,
			URL:  v.URL,
		}
	}

	return
}

// ParseConfig parses multiple sections of the SDK config from the config backend instance provided and returns Config containing maps of multiple config instances.
func ParseConfig(config core.ConfigBackend) (result Config, err error) {
	orderers, err := ParseOrderers(config)
	if err != nil {
		return
	}

	organizations, err := ParseOrganizations(config)
	if err != nil {
		return
	}

	peers, err := ParsePeers(config)
	if err != nil {
		return
	}

	result = Config{Orderers: orderers, Organizations: organizations, Peers: peers}
	return
}
