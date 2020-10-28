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

	clientInitInfo := &appinit.ClientInitInfo{
		AdminID: "Admin",
		UserID:  "User1",
		OrgName: "Org1",
		OrdererEndpoint: "orderer.lab805.com",
	}

	// Init the app
	initApp(clientInitInfo)

	// Create a channel
	if err = appinit.CreateChannel(channelInitInfo, clientInitInfo,
		global.MSPClientInstances.AdminMSPClient, global.ResMgmtClientInstances.AdminResMgmtClient); err != nil {
		log.Fatalln(err)
	}

	// Join a channel
}

//func setupLogger() {
//	log.SetFormatter(&log.JSONFormatter{})
//	log.WithField("package", "main")
//}

// Both the instantiations of the resource management clients and the MSP clients will be invoked here.
func initApp(clientInitInfo *appinit.ClientInitInfo) {
	// Setup the SDK
	err := appinit.SetupSDK("config-network.yaml")
	if err != nil {
		log.Fatalln(err)
	}

	defer global.SDKInstance.Close()

	// Instantiate clients
	err = appinit.InstantiateResMgmtClients(clientInitInfo)
	if err != nil {
		log.Fatalln(err)
	}

	err = appinit.InstantiateMSPClients(clientInitInfo)
	if err != nil {
		log.Fatalln(err)
	}
}
