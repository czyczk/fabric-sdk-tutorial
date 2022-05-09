package networkinfo

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/endpoint"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab"
)

// FabricNetworkConfig contains config info about the network which is needed by a client.
type FabricNetworkConfig struct {
	Orderers      map[string]FabricOrderer
	Organizations map[string]FabricOrganization
	Peers         map[string]FabricPeer
}

// FabricOrderer contains info about an orderer which is needed by a client.
type FabricOrderer struct {
	Name string
	URL  string
}

// FabricOrganization contains info about an organization which is needed by a client.
type FabricOrganization struct {
	Name       string
	MSPID      string
	CryptoPath string
	Peers      []string
	Users      []string
}

// FabricOrganizationUser represents a user in a FabricOrganization instance.
type FabricOrganizationUser struct {
	Name    string
	KeyPair endpoint.TLSKeyPair
}

// FabricPeer contains info about a peer which is needed by a client.
type FabricPeer struct {
	Name string
	URL  string
}

// ParseFabricOrderers parses the "orderers" section of the config from the config backend instance provided and returns a map of Orderer instances.
func ParseFabricOrderers(configBackend core.ConfigBackend) (result map[string]FabricOrderer, err error) {
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
	result = make(map[string]FabricOrderer)
	for k, v := range orderersMap {
		result[k] = FabricOrderer{Name: k, URL: v.URL}
	}

	return
}

// ParseFabricOrganizations parses the "organizations" section of the config from the config backend instance provided and returns a map of Organization instances.
func ParseFabricOrganizations(configBackend core.ConfigBackend) (result map[string]FabricOrganization, err error) {
	// Lookup the "organizations" section
	configBackendOrganizations, ok := configBackend.Lookup("organizations")
	if !ok {
		err = fmt.Errorf("error parsing organizations")
		return
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
	result = make(map[string]FabricOrganization)
	for k, v := range organizationsMap {
		var users []string
		for userName := range v.Users {
			users = append(users, userName)
		}

		result[k] = FabricOrganization{
			Name:       k,
			MSPID:      v.MSPID,
			CryptoPath: v.CryptoPath,
			Peers:      v.Peers,
			Users:      users,
		}
	}

	return
}

// ParseFabricPeers parses the "peers" section of the config from the config backend instance provided and returns a map of Peer instances.
func ParseFabricPeers(configBackend core.ConfigBackend) (result map[string]FabricPeer, err error) {
	// Lookup the "peers" section
	configBackendOrganizations, ok := configBackend.Lookup("peers")
	if !ok {
		err = fmt.Errorf("error parsing peers")
		return
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
	result = make(map[string]FabricPeer)
	for k, v := range peersMap {
		result[k] = FabricPeer{
			Name: k,
			URL:  v.URL,
		}
	}

	return
}

// ParseFabricNetworkConfig parses multiple sections of the SDK config from the config backend instance provided and returns Config containing maps of multiple config instances. Only to be used when the blockchain type is Fabric.
func ParseFabricNetworkConfig(config core.ConfigBackend) (result FabricNetworkConfig, err error) {
	orderers, err := ParseFabricOrderers(config)
	if err != nil {
		return
	}

	organizations, err := ParseFabricOrganizations(config)
	if err != nil {
		return
	}

	peers, err := ParseFabricPeers(config)
	if err != nil {
		return
	}

	result = FabricNetworkConfig{Orderers: orderers, Organizations: organizations, Peers: peers}
	return
}
