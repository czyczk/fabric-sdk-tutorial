package appinit

import (
	"fmt"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/global"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/gopackager"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/policydsl"
	errors "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// InstallCC installs the specified chaincode on all the peers of the specified org.
//
// Parameters:
//   chaincode ID
//   chaincode version
//   chaincode path (chaincode files should be in ${goPath}/src/${path})
//   chaincode Go path (chaincode files should be in ${goPath}/src/${path})
//   organization name. The chaincode will be installed on the peers of the organization.
//   operating identity. The identity that performs the operation.
func InstallCC(chaincodeID, version, path, goPath, orgName string, operatingIdentity *OperatingIdentity) error {
	resMgmtClient := global.ResMgmtClientInstances[operatingIdentity.OrgName][operatingIdentity.UserID]
	if resMgmtClient == nil {
		return fmt.Errorf("'%v' 的资源管理客户端未初始化", operatingIdentity.OrgName)
	}

	log.Printf("开始为组织 '%v' 的节点安装链码 '%v'...\n", orgName, chaincodeID)

	// Create a new Go chaincode package
	ccPkg, err := gopackager.NewCCPackage(path, goPath)
	if err != nil {
		return errors.Wrapf(err, "为链码 '%v' 创建链码包失败", chaincodeID)
	}

	// Make a request containing parameters to install the chaincode
	installCCReq := resmgmt.InstallCCRequest{
		Name:    chaincodeID,
		Path:    path,
		Version: version,
		Package: ccPkg,
	}

	// Install the chaincode. No peers are specified here so it will be installed on all the peers in the org.
	_, err = resMgmtClient.InstallCC(installCCReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		return errors.Wrapf(err, "为组织 '%v' 的节点安装链码 '%v' 失败", orgName, chaincodeID)
	}

	log.Printf("已为组织 '%v' 的节点安装链码 '%v'。\n", orgName, chaincodeID)

	return nil
}

// InstantiateCC instantiates the specified chaincode on the specified channel.
//
// Parameters:
//   chaincode ID
//   chaincode path
//   chaincode version
//   channel ID
//   chaincode instantiation info
func InstantiateCC(chaincodeID, path, version, channelID string, info *ChaincodeInstantiationInfo) error {
	resMgmtClient := global.ResMgmtClientInstances[info.OrgName][info.UserID]
	if resMgmtClient == nil {
		return fmt.Errorf("'%v.%v' 的资源管理客户端未初始化", info.UserID, info.OrgName)
	}

	ledgerClient := global.LedgerClientInstances[channelID][info.OrgName][info.UserID]
	if ledgerClient == nil {
		return fmt.Errorf("'%v.%v' 在通道 '%v' 上的账本客户端未初始化", info.UserID, info.OrgName, channelID)
	}

	log.Printf("开始在通道 '%v' 上实例化链码 '%v'...\n", channelID, chaincodeID)

	// Parse the endorsement policy
	ccPolicy, err := policydsl.FromString(info.Policy)
	if err != nil {
		return errors.Wrapf(err, "在通道 '%v' 上实例化链码 '%v' 失败", channelID, chaincodeID)
	}

	var initArgsOfBytes [][]byte
	for _, arg := range info.InitArgs {
		initArgsOfBytes = append(initArgsOfBytes, []byte(arg))
	}

	// Make a request containing parameters to instantiate the chaincode
	instantiateCCReq := resmgmt.InstantiateCCRequest{
		Name:    chaincodeID,
		Path:    path,
		Version: version,
		Args:    initArgsOfBytes,
		Policy:  ccPolicy,
	}

	// Instantiate the chaincode for all peers in the org
	_, err = resMgmtClient.InstantiateCC(channelID, instantiateCCReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		return errors.Wrapf(err, "在通道 '%v' 上实例化链码 '%v' 失败", channelID, chaincodeID)
	}

	log.Printf("已在通道 '%v' 上实例化链码 '%v'。\n", channelID, chaincodeID)

	return nil
}
