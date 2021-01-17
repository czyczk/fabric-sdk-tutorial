package main

import (
	"fmt"
	"os"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/networkinfo"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/controller"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/service"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/global"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/appinit"
)

func main() {
	//setupLogger()

	workingDirectory, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}

	// Create a Fabric SDK instance
	err = appinit.SetupSDK("config-network.yaml")
	if err != nil {
		log.Fatalln(err)
	}

	defer global.SDKInstance.Close()

	// Load init info from `init.yaml`
	initInfoPath := workingDirectory + "/init.yaml"
	initInfo, err := appinit.LoadInitInfo(initInfoPath)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(initInfo)

	// Init the app
	appinit.InitApp(&initInfo)

	// Fetch the network config info
	sdkConfig, err := global.SDKInstance.Config()
	if err != nil {
		log.Fatalln(err)
	}
	networkConfig, err := networkinfo.ParseConfig(sdkConfig)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(networkConfig)

	// Instantiate a screw service.
	serviceInfo := &service.Info{
		ChaincodeID:   "screwCc",
		ChannelClient: global.ChannelClientInstances["mychannel"]["Org1"]["User1"],
	}

	screwSvc := &service.ScrewService{ServiceInfo: serviceInfo}

	// Make a "transfer" request to transfer 10 screws from "Org1" to "Org2" and show the transaction ID.
	respMsg, err := screwSvc.TransferAndShowEvent("Org1", "Org2", 10)
	if err != nil {
		log.Fatalln(err)
	} else {
		fmt.Printf("Transaction ID: %v\n", respMsg)
	}

	// Make a "query" request for "Org1" and show the response payload.
	respMsg, err = screwSvc.Query("Org1")
	if err != nil {
		log.Fatalln(err)
	} else {
		fmt.Printf("Screw amount in Org 1 after the transfer: %v\n", respMsg)
	}

	// Instantiate a ping pong controller
	pingPongController := &controller.PingPongController{}

	// Instantiate a screw controller
	screwController := &controller.ScrewController{
		GroupName: "/screw",
		ScrewSvc:  screwSvc,
	}

	// Register controller handlers
	router := gin.Default()
	router.Use(controller.CORSMiddleware())
	apiv1Group := router.Group("/api/v1")
	controller.RegisterHandlers(apiv1Group, pingPongController)
	controller.RegisterHandlers(apiv1Group, screwController)
	router.Run(":8081")
}
