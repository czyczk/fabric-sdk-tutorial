package service

import (
	ipfs "github.com/ipfs/go-ipfs-api"
	"gorm.io/gorm"
)

// Info needed for a service to know which `chaincodeID` it's serving and which channel it's using.
type Info struct {
	DB     *gorm.DB
	IPFSSh *ipfs.Shell
}
