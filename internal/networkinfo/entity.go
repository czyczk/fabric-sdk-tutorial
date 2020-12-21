package networkinfo

import "github.com/hyperledger/fabric-sdk-go/pkg/core/config/endpoint"

// Config contains config info about the network which is needed by a client.
type Config struct {
	Orderers      map[string]Orderer
	Organizations map[string]Organization
	Peers         map[string]Peer
}

// Orderer contains info about an orderer which is needed by a client.
type Orderer struct {
	Name string
	URL  string
}

// Organization contains info about an organization which is needed by a client.
type Organization struct {
	Name       string
	MSPID      string
	CryptoPath string
	Peers      []string
	Users      []string
}

// OrganizationUser represents a user in an Organization instance.
type OrganizationUser struct {
	Name    string
	KeyPair endpoint.TLSKeyPair
}

// Peer contains info about a peer which is needed by a client.
type Peer struct {
	Name string
	URL  string
}
