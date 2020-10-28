package appinit

// ChannelInitInfo needed to create channels.
type ChannelInitInfo struct {
	ChannelID         string
	ChannelConfigPath string
}

// ClientInitInfo needed to create resource management clients and MSP clients.
type ClientInitInfo struct {
	AdminID         string // The admin ID of the organization
	UserID          string // The user ID of the organization
	OrgName         string // The organization name
	OrdererEndpoint string // The endpoint of any orderer of the organization
}

