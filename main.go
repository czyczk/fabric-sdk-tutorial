package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"time"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/appinit"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/background"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/controller"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/global"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/networkinfo"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
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
		documentSvc := &service.DocumentService{
			ServiceInfo:      universalCcServiceInfo,
			KeySwitchService: keySwitchSvc,
		}

		// Instantiate an entity asset service
		entityAssetSvc := &service.EntityAssetService{
			ServiceInfo:      universalCcServiceInfo,
			KeySwitchService: keySwitchSvc,
		}

		// TODO: Instantiate a auth service

		// Prepare a key switch server. It will be of use if the app is enabled as a key switch server.
		ksServer := background.NewKeySwitchServer(serviceInfo, keySwitchSvc, runtime.NumCPU())
		if isKeySwitchServer {
			// Start the server to listen key switch triggers
			err := ksServer.Start()
			if err != nil {
				return err
			}
		}

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
			GroupName:      "/",
			DocumentSvc:    documentSvc,
			EntityAssetSvc: entityAssetSvc,
		}

		// Instantiate a entity controller
		entityAssetController := &controller.EntityAssetController{
			GroupName:      "/asset",
			EntityAssetSvc: entityAssetSvc,
		}

		// Register controller handlers
		router := gin.Default()
		router.Use(controller.CORSMiddleware())
		apiv1Group := router.Group("/api/v1")
		controller.RegisterHandlers(apiv1Group, pingPongController)
		controller.RegisterHandlers(apiv1Group, screwController)
		controller.RegisterHandlers(apiv1Group, documentController)
		controller.RegisterHandlers(apiv1Group, entityAssetController)

		// Start the HTTP server
		httpServer := &http.Server{
			Addr:    fmt.Sprintf(":%v", serverInfo.Port),
			Handler: router,
		}

		chanError := make(chan error)
		go func() {
			if err := httpServer.ListenAndServe(); err != nil {
				chanError <- errors.Wrap(err, "无法启动 HTTP 服务器")
			}
		}()

		// Listen Ctrl+C signals. On receiving a signal stops the app elegantly
		chanQuit := make(chan os.Signal)
		signal.Notify(chanQuit, os.Interrupt)
		select {
		case err := <-chanError:
			return err
		case <-chanQuit:
			log.Infoln("收到 Ctrl+C 信号，正在退出程序...")

			// Stop the HTTP server elegantly
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			log.Infoln("正在停止 HTTP 服务器...")
			if err := httpServer.Shutdown(ctx); err != nil {
				return errors.Wrap(err, "无法正常停止 HTTP 服务器")
			}

			// Stop the key switch server if enabled
			if isKeySwitchServer {
				log.Infoln("正在停止密钥置换服务器...")
				wg, err := ksServer.Stop()
				defer wg.Wait()
				if err != nil {
					return err
				}
			}
		}

		return nil
	}

	return serveFunc
}
