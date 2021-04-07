package main

import (
	"fmt"
	"os"
	"runtime"
	"sync"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/appinit"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/controller"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/global"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/networkinfo"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/service"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	//setupLogger()

	var configPath, sdkConfigPath string

	// Functions to be used by the cli helper
	initFunc := getInitFunc(&configPath, &sdkConfigPath)
	serveFunc := getServeFunc(&configPath, &sdkConfigPath)

	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "init",
				Aliases: []string{"i"},
				Usage:   "Initialize the network",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "conf",
						Aliases:     []string{"c"},
						Value:       "init.yaml",
						EnvVars:     []string{"FST_CONF"},
						Destination: &configPath,
					},
					&cli.StringFlag{
						Name:        "sdkconf",
						Aliases:     []string{"s"},
						Value:       "config-network.yaml",
						EnvVars:     []string{"FST_SDK_CONF"},
						Destination: &sdkConfigPath,
					},
				},
				Action: initFunc,
			},
			{
				Name:    "serve",
				Aliases: []string{"s"},
				Usage:   "Start as server",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "conf",
						Aliases:     []string{"c"},
						Value:       "server.yaml",
						EnvVars:     []string{"FST_CONF"},
						Destination: &configPath,
					},
					&cli.StringFlag{
						Name:        "sdkconf",
						Aliases:     []string{"s"},
						Value:       "config-network.yaml",
						EnvVars:     []string{"FST_SDK_CONF"},
						Destination: &sdkConfigPath,
					},
				},
				Action: serveFunc,
			},
		},
	}

	// Run the cli helper
	if err := app.Run(os.Args); err != nil {
		log.Fatalln(err)
	}
}

func getInitFunc(configPath *string, sdkConfigPath *string) func(c *cli.Context) error {
	// The func for subcommand "init"
	initFunc := func(c *cli.Context) error {
		// Create a Fabric SDK instance
		err := appinit.SetupSDK(*sdkConfigPath)
		if err != nil {
			return err
		}

		defer global.SDKInstance.Close()

		// Load init info from `init.yaml`
		initInfo, err := appinit.LoadInitInfo(*configPath)
		if err != nil {
			return err
		}

		// Init the app
		if err := appinit.InitApp(&initInfo); err != nil {
			return err
		}

		// Fetch the network config info
		sdkConfig, err := global.SDKInstance.Config()
		if err != nil {
			return err
		}
		networkConfig, err := networkinfo.ParseConfig(sdkConfig)
		if err != nil {
			return err
		}
		fmt.Println(networkConfig)

		return nil
	}

	return initFunc
}

func getServeFunc(configPath *string, sdkConfigPath *string) func(c *cli.Context) error {
	serveFunc := func(c *cli.Context) error {
		// Create a Fabric SDK instance
		err := appinit.SetupSDK(*sdkConfigPath)
		if err != nil {
			return err
		}

		defer global.SDKInstance.Close()

		// Load serve info from `serve.yaml`
		serverInfo, err := appinit.LoadServerInfo(*configPath)
		if err != nil {
			return err
		}

		// Extract some info from the config for later use
		orgName := serverInfo.User.OrgName
		userID := serverInfo.User.UserID
		isKeySwitchServer := serverInfo.IsKeySwitchServer

		// Create clients
		if err = appinit.InstantiateResMgmtClient(orgName, userID); err != nil {
			return err
		}

		if err = appinit.InstantiateMSPClient(orgName, userID); err != nil {
			return err
		}

		for _, channelID := range serverInfo.Channels {
			if err = appinit.InstantiateChannelClient(global.SDKInstance, channelID, orgName, userID); err != nil {
				return err
			}
		}

		// Prepare the channels for key switch go routines. They will be of use if the app is enabled as a key switch server.
		ksServerJobsChan := make(chan string)
		ksServerQuitChan := make(chan int)
		var ksServerWg sync.WaitGroup
		defer ksServerWg.Wait()

		// Prepare to load key switch keys
		if serverInfo.KeySwitchKeys == nil || serverInfo.KeySwitchKeys.CollectivePublicKey == "" {
			return fmt.Errorf("未指定密钥置换集合公钥")
		}

		if isKeySwitchServer {
			// Make sure the private key and the public key are specified if the app is enabled as a key switch server
			if serverInfo.KeySwitchKeys.PrivateKey == "" || serverInfo.KeySwitchKeys.PublicKey == "" {
				return fmt.Errorf("密钥置换所需的私钥和/或公钥未指定。请指定公私钥或将密钥置换服务器关闭")
			}

		}

		appinit.LoadKeySwitchServerKeys(serverInfo.KeySwitchKeys)

		if isKeySwitchServer {
			// Start #LogicalCPUs go routines to listen key switch triggers
			for i := 0; i < runtime.NumCPU(); i++ {
				ksServerWg.Add(1)
				// TODO: Start a key switch trigger listener
			}
		}

		// Instantiate a screw service
		serviceInfo := &service.Info{
			ChaincodeID:   "screwCc",
			ChannelClient: global.ChannelClientInstances["mychannel"][orgName][userID],
		}

		screwSvc := &service.ScrewService{ServiceInfo: serviceInfo}

		// Instantiate a key switch service
		universalCcServiceInfo := &service.Info{
			ChaincodeID:   "universalCc",
			ChannelClient: global.ChannelClientInstances["mychannel"][orgName][userID],
		}

		keySwitchSvc := &service.KeySwitchService{ServiceInfo: universalCcServiceInfo}

		// Instantiate a document service
		documentSvc := &service.DocumentService{ServiceInfo: universalCcServiceInfo, KeySwitchSvc: keySwitchSvc}

		// TODO: Instantiate a auth service

		// Make a "transfer" request to transfer 10 screws from "Org1" to "Org2" and show the transaction ID
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

		// Instantiate controllers
		// Instantiate a ping pong controller
		pingPongController := &controller.PingPongController{}

		// Instantiate a screw controller
		screwController := &controller.ScrewController{
			GroupName: "/screw",
			ScrewSvc:  screwSvc,
		}

		// Instantiate a document controller
		documentController := &controller.DocumentController{
			GroupName:   "/document",
			DocumentSvc: documentSvc,
		}

		// Register controller handlers
		router := gin.Default()
		router.Use(controller.CORSMiddleware())
		apiv1Group := router.Group("/api/v1")
		controller.RegisterHandlers(apiv1Group, pingPongController)
		controller.RegisterHandlers(apiv1Group, screwController)
		controller.RegisterHandlers(apiv1Group, documentController)
		router.Run(fmt.Sprintf(":%v", serverInfo.Port))

		// TODO: Listen to Ctrl+C signal and send #LogicalCPUs quit signals to `ksServerQuitChan`

		return nil
	}

	return serveFunc
}
