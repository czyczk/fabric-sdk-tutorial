package appinit

// ChannelInitInfo needed to create channels.
type ChannelInitInfo struct {
	ChannelID         string
	ChannelConfigPath string
}

// OrgInitInfo needed to create resource management clients and MSP clients.
type OrgInitInfo struct {
	AdminID string // The admin ID of the organization
	UserID  string // The user ID of the organization
	OrgName string // The organization name. Used for display and client map lookup within the app.
}

// ChaincodeInitInfo needed to install and instantiate a chaincode.
type ChaincodeInitInfo struct {
	ChaincodeID      string
	ChaincodeVersion string
	ChaincodePath    string
	ChaincodeGoPath  string
	Policy           string
	InitArgs         [][]byte
}
