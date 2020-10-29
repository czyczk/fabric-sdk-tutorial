package main

import (
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/global"
	log "github.com/sirupsen/logrus"
	"os"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/appinit"
)

func main() {
	//setupLogger()

	workingDirectory, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}

	// Specify init info
	channelInitInfo := &appinit.ChannelInitInfo{
		ChannelID:         "mychannel",
		ChannelConfigPath: workingDirectory + "/fixtures/channel-artifacts/channel.tx",
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

	// Init the app
	initApp([]*appinit.OrgInitInfo{org1InitInfo, org2InitInfo})
	defer global.SDKInstance.Close()

	// Create a channel
	if err = appinit.CreateChannel(channelInitInfo, org1InitInfo); err != nil {
		log.Fatalln(err)
	}

	// Join peers of org1 to the channel
	if err = appinit.JoinChannel(channelInitInfo, org1InitInfo); err != nil {
		log.Fatalln(err)
	}

	// Join peers of org2 to the channel
	if err = appinit.JoinChannel(channelInitInfo, org2InitInfo); err != nil {
		log.Fatalln(err)
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
