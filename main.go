package main

import (
	"fmt"
	"os"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/service"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/global"
	log "github.com/sirupsen/logrus"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/appinit"
)

func main() {
	//setupLogger()

	workingDirectory, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}

	// Specify init info
	channelID := "mychannel"
	channelInitInfo := &appinit.ChannelInitInfo{
		ChannelID:         channelID,
		ChannelConfigPath: workingDirectory + "/fixtures/channel-artifacts/channel.tx",
	}
	org1AnchorPeerInitInfo := &appinit.ChannelInitInfo{
		ChannelID:         channelID,
		ChannelConfigPath: workingDirectory + "/fixtures/channel-artifacts/Org1MSPanchors.tx",
	}
	org2AnchorPeerInitInfo := &appinit.ChannelInitInfo{
		ChannelID:         channelID,
		ChannelConfigPath: workingDirectory + "/fixtures/channel-artifacts/Org2MSPanchors.tx",
	}

	org1InitInfo := &appinit.OrgInitInfo{
		AdminID:         "Admin",
		UserID:          "User1",
		OrgName:         "Org1",
		OrdererEndpoint: "orderer.lab805.com",
	}

	org2InitInfo := &appinit.OrgInitInfo{
		AdminID:         "Admin",
		UserID:          "User1",
		OrgName:         "Org2",
		OrdererEndpoint: "orderer2.lab805.com",
	}

	chaincodeInitInfo := &appinit.ChaincodeInitInfo{
		ChaincodeID:      "screwCc",
		ChaincodeVersion: "0.1",
		ChaincodePath:    "screw_example",
		ChaincodeGoPath:  workingDirectory + "/chaincode",
		Policy:           "OR('Org1MSP.member', 'Org2MSP.member')",
		InitArgs: [][]byte{[]byte("init"),
			[]byte(org1InitInfo.OrgName), []byte("200"),
			[]byte(org2InitInfo.OrgName), []byte("100")},
	}

	// Init the app
	initApp([]*appinit.OrgInitInfo{org1InitInfo, org2InitInfo})
	defer global.SDKInstance.Close()

	// Configure a channel
	configureChannel(org1InitInfo, org2InitInfo, channelInitInfo, org1AnchorPeerInitInfo, org2AnchorPeerInitInfo)

	// Install and instantiate the chaincode
	installAndInstantiateChaincode(org1InitInfo, org2InitInfo, channelInitInfo, chaincodeInitInfo)

	// Instantiate a screw service.
	serviceInfo := &service.Info{
		ChaincodeID:   chaincodeInitInfo.ChaincodeID,
		ChannelClient: global.ChannelClientInstances[channelInitInfo.ChannelID][org1InitInfo.OrgName][org1InitInfo.AdminID],
	}

	screwSvc := &service.ScrewService{ServiceInfo: serviceInfo}

	// Make a "transfer" request to transfer 10 screws from "Org1" to "Org2" and show the transaction ID.
	respMsg, err := screwSvc.TransferAndShowEvent(org1InitInfo.OrgName, org2InitInfo.OrgName, 10)
	if err != nil {
		log.Fatalln(err)
	} else {
		fmt.Printf("Transaction ID: %v\n", respMsg)
	}

	// Make a "query" request for "Org1" and show the response payload.
	respMsg, err = screwSvc.Query(org1InitInfo.OrgName)
	if err != nil {
		log.Fatalln(err)
	} else {
		fmt.Printf("Screw amount in Org 1 after the transfer: %v\n", respMsg)
	}
}

//func setupLogger() {
//	log.SetFormatter(&log.JSONFormatter{})
//	log.WithField("package", "main")
//}

// Both the instantiations of the resource management clients and the MSP clients will be invoked here.
func initApp(orgInitInfoList []*appinit.OrgInitInfo) {
	// Setup the SDK
	err := appinit.SetupSDK("config-network.yaml")
	if err != nil {
		log.Fatalln(err)
	}

	// Instantiate resource management clients and MSP clients for the orgs in the list.
	for _, info := range orgInitInfoList {
		// Instantiate clients
		err = appinit.InstantiateResMgmtClients(info.AdminID, info.OrgName)
		if err != nil {
			log.Fatalln(err)
		}
		err = appinit.InstantiateResMgmtClients(info.UserID, info.OrgName)
		if err != nil {
			log.Fatalln(err)
		}
		err = appinit.InstantiateMSPClients(info.AdminID, info.OrgName)
		if err != nil {
			log.Fatalln(err)
		}
		err = appinit.InstantiateMSPClients(info.UserID, info.OrgName)
		if err != nil {
			log.Fatalln(err)
		}
	}

}

// This function creates and configures a channel according to the channel init info and joins the peers to the channel.
func configureChannel(org1InitInfo *appinit.OrgInitInfo, org2InitInfo *appinit.OrgInitInfo,
	channelInitInfo *appinit.ChannelInitInfo,
	org1AnchorPeerInitInfo *appinit.ChannelInitInfo, org2AnchorPeerInitInfo *appinit.ChannelInitInfo) {
	// Create a channel
	if err := appinit.ApplyChannelTx(channelInitInfo, org1InitInfo); err != nil {
		log.Fatalln(err)
	}

	// Update anchor peers
	if err := appinit.ApplyChannelTx(org1AnchorPeerInitInfo, org1InitInfo); err != nil {
		log.Fatalln(err)
	}
	if err := appinit.ApplyChannelTx(org2AnchorPeerInitInfo, org2InitInfo); err != nil {
		log.Fatalln(err)
	}

	// Join peers of org1 to the channel
	if err := appinit.JoinChannel(channelInitInfo, org1InitInfo); err != nil {
		log.Fatalln(err)
	}

	// Join peers of org2 to the channel
	if err := appinit.JoinChannel(channelInitInfo, org2InitInfo); err != nil {
		log.Fatalln(err)
	}
}

func installAndInstantiateChaincode(org1InitInfo, org2InitInfo *appinit.OrgInitInfo,
	channelInitInfo *appinit.ChannelInitInfo, chaincodeInitInfo *appinit.ChaincodeInitInfo) {
	// Install the chaincode for peers in org1 and org2
	if err := appinit.InstallCC(chaincodeInitInfo, org1InitInfo); err != nil {
		log.Fatalln(err)
	}

	if err := appinit.InstallCC(chaincodeInitInfo, org2InitInfo); err != nil {
		log.Fatalln(err)
	}

	// Instantiate the chaincode on the channel
	if err := appinit.InstantiateCC(global.SDKInstance, chaincodeInitInfo, channelInitInfo, org1InitInfo); err != nil {
		log.Fatalln(err)
	}

	// Instantiate channel clients
	for _, orgInfo := range []*appinit.OrgInitInfo{org1InitInfo, org2InitInfo} {
		if err := appinit.InstantiateChannelClient(global.SDKInstance, channelInitInfo.ChannelID, orgInfo.AdminID, orgInfo.OrgName); err != nil {
			log.Fatalln(err)
		}

		if err := appinit.InstantiateChannelClient(global.SDKInstance, channelInitInfo.ChannelID, orgInfo.UserID, orgInfo.OrgName); err != nil {
			log.Fatalln(err)
		}
	}
}
