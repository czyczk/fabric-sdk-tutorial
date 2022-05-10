package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"time"

	cryptorand "crypto/rand"
	mathrand "math/rand"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/appinit"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/background"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/bcao"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/bcao/fabricbcao"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/bcao/polkadotbcao"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/chaincodectx"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/eventmgr"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/eventmgr/fabriceventmgr"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/controller"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/global"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/models/sqlmodel"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/networkinfo"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/service"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/utils/timingutils"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	ipfs "github.com/ipfs/go-ipfs-api"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	//setupLogger()

	var blockchainTypeStr, configPath, blockchainConfigPath string

	// Functions to be used by the cli helper
	initFunc := getInitFunc(&blockchainTypeStr, &configPath, &blockchainConfigPath)
	serveFunc := getServeFunc(&blockchainTypeStr, &configPath, &blockchainConfigPath)

	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "init",
				Aliases: []string{"i"},
				Usage:   "Initialize the network",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "blockchaintype",
						Aliases:     []string{"t"},
						Value:       "fabric",
						EnvVars:     []string{"FST_BLOCKCHAIN_TYPE"},
						Destination: &blockchainTypeStr,
					},
					&cli.StringFlag{
						Name:        "conf",
						Aliases:     []string{"c"},
						Value:       "init.yaml",
						EnvVars:     []string{"FST_CONF"},
						Destination: &configPath,
					},
					&cli.StringFlag{
						Name:        "blockchainconf",
						Aliases:     []string{"b"},
						Value:       "fabric-config-network.yaml",
						EnvVars:     []string{"FST_BLOCKCHAIN_CONF"},
						Destination: &blockchainConfigPath,
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
						Name:        "blockchaintype",
						Aliases:     []string{"t"},
						Value:       "fabric",
						EnvVars:     []string{"FST_BLOCKCHAIN_TYPE"},
						Destination: &blockchainTypeStr,
					},
					&cli.StringFlag{
						Name:        "conf",
						Aliases:     []string{"c"},
						Value:       "server.yaml",
						EnvVars:     []string{"FST_CONF"},
						Destination: &configPath,
					},
					&cli.StringFlag{
						Name:        "blockchainconf",
						Aliases:     []string{"b"},
						Value:       "fabric-config-network.yaml",
						EnvVars:     []string{"FST_BLOCKCHAIN_CONF"},
						Destination: &blockchainConfigPath,
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

func getInitFunc(blockchainTypeStr *string, configPath *string, blockchainConfigPath *string) func(c *cli.Context) error {
	// The func for subcommand "init"
	initFunc := func(c *cli.Context) error {
		// Parse blockchain type
		if blockchainTypeStr == nil {
			return fmt.Errorf("未指定区块链类型")
		}

		if err := appinit.ParseBlockchainType(*blockchainTypeStr); err != nil {
			return err
		}

		if global.BlockchainType == blockchain.Fabric {
			// Create a Fabric SDK instance only if the blockchain type is Fabric
			err := appinit.SetupSDK(*blockchainConfigPath)
			if err != nil {
				return err
			}

			defer global.FabricSDKInstance.Close()

			// Fetch the network config info
			sdkConfig, err := global.FabricSDKInstance.Config()
			if err != nil {
				return err
			}
			networkConfig, err := networkinfo.ParseFabricNetworkConfig(sdkConfig)
			if err != nil {
				return err
			}

			global.FabricNetworkConfig = &networkConfig
		} else if global.BlockchainType == blockchain.Polkadot {
			// Load a Polkadot network config
			networkConfig, err := networkinfo.ParsePolkadotNetworkConfig(*blockchainConfigPath)
			if err != nil {
				return err
			}

			global.PolkadotNetworkConfig = networkConfig
		} else {
			return fmt.Errorf("未实现的区块链类型")
		}

		// Load init info from the init config file
		initInfo, err := appinit.LoadInitInfo(*configPath)
		if err != nil {
			return err
		}

		// Init the app
		if err := appinit.InitApp(initInfo); err != nil {
			return err
		}

		return nil
	}

	return initFunc
}

func getServeFunc(blockchainTypeStr *string, configPath *string, blockchainConfigPath *string) func(c *cli.Context) error {
	log.SetLevel(log.DebugLevel)
	serveFunc := func(c *cli.Context) error {
		// Parse blockchain type
		if blockchainTypeStr == nil {
			return fmt.Errorf("未指定区块链类型")
		}

		if err := appinit.ParseBlockchainType(*blockchainTypeStr); err != nil {
			return err
		}

		// Load blockchain network config
		if global.BlockchainType == blockchain.Fabric {
			// Create a Fabric SDK instance
			err := appinit.SetupSDK(*blockchainConfigPath)
			if err != nil {
				return err
			}

			defer global.FabricSDKInstance.Close()
		} else if global.BlockchainType == blockchain.Polkadot {
			// Load a Polkadot network config
			networkConfig, err := networkinfo.ParsePolkadotNetworkConfig(*blockchainConfigPath)
			if err != nil {
				return err
			}

			global.PolkadotNetworkConfig = networkConfig
		} else {
			return fmt.Errorf("未实现的区块链类型")
		}

		// Load server info from `server.yaml`
		serverInfo, err := appinit.LoadServerInfo(*configPath)
		if err != nil {
			return err
		}

		// Apply global settings
		global.ShowTimingLogs = serverInfo.ShowTimingLogs

		// Extract some info from the config for later use
		orgName := serverInfo.User.OrgName
		userID := serverInfo.User.UserID
		isKeySwitchServer := serverInfo.IsKeySwitchServer
		isRegulator := serverInfo.IsRegulator

		if global.BlockchainType == blockchain.Fabric {
			// Create clients only for Fabric
			if err = appinit.InstantiateResMgmtClient(orgName, userID); err != nil {
				return err
			}

			if err = appinit.InstantiateMSPClient(orgName, userID); err != nil {
				return err
			}

			for _, channelID := range serverInfo.Channels {
				if err = appinit.InstantiateChannelClient(global.FabricSDKInstance, channelID, orgName, userID); err != nil {
					return err
				}

				if err = appinit.InstantiateEventClient(global.FabricSDKInstance, channelID, orgName, userID); err != nil {
					return err
				}

				if err = appinit.InstantiateLedgerClient(global.FabricSDKInstance, channelID, orgName, userID); err != nil {
					return err
				}
			}
		} else if global.BlockchainType == blockchain.Polkadot {
			// Register the logged-in user in the HTTP API
			if err := appinit.RegisterPolkadotUsers(global.PolkadotNetworkConfig.Organizations, global.PolkadotNetworkConfig.APIPrefix); err != nil {
				return err
			}
		}

		// Check and create a MySQL database connection
		db, err := gorm.Open(mysql.Open(serverInfo.LocalDBSourceName), &gorm.Config{})
		if err != nil {
			return errors.Wrap(err, "无法连接数据库")
		}

		// Auto migrate schemas
		err = db.AutoMigrate(&sqlmodel.DocumentProperties{}, &sqlmodel.Document{}, &sqlmodel.EntityAsset{}, &sqlmodel.Component{})
		if err != nil {
			return errors.Wrap(err, "无法创建数据库表")
		}

		log.Info("已连接数据库并创建数据库表。")

		// Prepare to load key switch keys
		if serverInfo.KeySwitchKeys == nil || serverInfo.KeySwitchKeys.CollectivePublicKey == "" {
			return fmt.Errorf("未指定密钥置换集合公钥")
		}

		// Make sure the private key and public key are specified to participate in retrieving encrypted data
		if serverInfo.KeySwitchKeys.PrivateKey == "" || serverInfo.KeySwitchKeys.PublicKey == "" {
			return fmt.Errorf("未指定密钥置换所需的私钥和/或公钥。将无法获取加密资源")
		}

		// Load key switch keys
		err = appinit.LoadKeySwitchServerKeys(serverInfo.KeySwitchKeys)
		if err != nil {
			return err
		}

		// Connect to IPFS service
		ipfsSh := ipfs.NewShell(serverInfo.IPFSAPI)
		ipfsSh.SetTimeout(20 * time.Second)
		if !ipfsSh.IsUp() {
			return fmt.Errorf("无法连接到 IPFS 节点")
		}

		// If enabled, load regulator config
		if isRegulator {
			// Make sure the collective private key is specified
			if serverInfo.KeySwitchKeys.CollectivePrivateKey == "" {
				return fmt.Errorf("未指定监管者功能所需的集合私钥。请指定集合私钥或将监管者关闭")
			}
		}

		var screwCcCtx chaincodectx.IChaincodeCtx
		var universalCcCtx chaincodectx.IChaincodeCtx
		switch global.BlockchainType {
		case blockchain.Fabric:
			screwCcCtx = &chaincodectx.FabricChaincodeCtx{
				ChannelID:     "mychannel",
				OrgName:       orgName,
				Username:      userID,
				ChaincodeID:   "screwCc",
				ChannelClient: global.ChannelClientInstances["mychannel"][orgName][userID],
				EventClient:   global.EventClientInstances["mychannel"][orgName][userID],
				LedgerClient:  global.LedgerClientInstances["mychannel"][orgName][userID],
			}

			universalCcCtx = &chaincodectx.FabricChaincodeCtx{
				ChannelID:     "mychannel",
				OrgName:       orgName,
				Username:      userID,
				ChaincodeID:   "universalCc",
				ChannelClient: global.ChannelClientInstances["mychannel"][orgName][userID],
				EventClient:   global.EventClientInstances["mychannel"][orgName][userID],
				LedgerClient:  global.LedgerClientInstances["mychannel"][orgName][userID],
			}
		case blockchain.Polkadot:
			// TODO: Polkadot screw CC ctx cannot be created because the chaincode address and ABI are missing

			callerAddress, err := global.PolkadotNetworkConfig.GetUserAddress(orgName, userID)
			if err != nil {
				return err
			}

			universalCcABI, err := global.PolkadotNetworkConfig.GetChaincodeABI("universalCc")
			if err != nil {
				return err
			}

			universalCcCtx = &chaincodectx.PolkadotChaincodeCtx{
				CallerAddress:   callerAddress,
				APIPrefix:       global.PolkadotNetworkConfig.APIPrefix,
				ContractAddress: global.PolkadotNetworkConfig.GetChaincodeAddress("universalCc"),
				ContractABI:     universalCcABI,
			}
		default:
			return fmt.Errorf("未知的区块链类型")
		}

		// Instantiate BCAOs
		var screwBCAO bcao.IScrewBCAO
		var dataBCAO bcao.IDataBCAO
		var authBCAO bcao.IAuthBCAO
		var keySwitchBCAO bcao.IKeySwitchBCAO
		var identityBCAO bcao.IIdentityBCAO

		switch global.BlockchainType {
		case blockchain.Fabric:
			screwBCAO = fabricbcao.NewScrewBCAOFabricImpl(screwCcCtx.(*chaincodectx.FabricChaincodeCtx))
			dataBCAO = fabricbcao.NewDataBCAOFabricImpl(universalCcCtx.(*chaincodectx.FabricChaincodeCtx))
			authBCAO = fabricbcao.NewAuthBCAOFabricImpl(universalCcCtx.(*chaincodectx.FabricChaincodeCtx))
			keySwitchBCAO = fabricbcao.NewKeySwitchBCAOFabricImpl(universalCcCtx.(*chaincodectx.FabricChaincodeCtx))
			identityBCAO = fabricbcao.NewIdentityBCAOFabricImpl(universalCcCtx.(*chaincodectx.FabricChaincodeCtx))
		case blockchain.Polkadot:
			// TODO: Polkadot screw BCAO cannot be created because the CC ctx is missing
			dataBCAO = polkadotbcao.NewDataBCAOPolkadotImpl(universalCcCtx.(*chaincodectx.PolkadotChaincodeCtx))
			authBCAO = polkadotbcao.NewAuthBCAOPolkadotImpl(universalCcCtx.(*chaincodectx.PolkadotChaincodeCtx))
			keySwitchBCAO = polkadotbcao.NewKeySwitchBCAOPolkadotImpl(universalCcCtx.(*chaincodectx.PolkadotChaincodeCtx))
			identityBCAO = polkadotbcao.NewIdentityBCAOPolkadotImpl(universalCcCtx.(*chaincodectx.PolkadotChaincodeCtx))
		default:
			return fmt.Errorf("未知的区块链类型")
		}

		// Instantiate event managers
		// TODO: screw cc event manager
		var universalCcEventManager eventmgr.IEventManager

		switch global.BlockchainType {
		case blockchain.Fabric:
			universalCcEventManager = fabriceventmgr.NewFabricEventManager(universalCcCtx.(*chaincodectx.FabricChaincodeCtx))
		case blockchain.Polkadot:
			// TODO: Not implemented yet
		default:
			return fmt.Errorf("未知的区块链类型")
		}

		// Instantiate services
		screwSvc := &service.ScrewService{ScrewBCAO: screwBCAO} // TODO: EventManager

		// Instantiate a key switch service
		universalCcServiceInfo := &service.Info{
			DB:     db,
			IPFSSh: ipfsSh,
		}

		keySwitchSvc := &service.KeySwitchService{
			KeySwitchBCAO: keySwitchBCAO,
			EventManager:  universalCcEventManager,
		}

		// Instantiate a document service
		// Create file loggers for document service
		chanLoggerErr := make(chan error)
		go func() {
			for {
				err := <-chanLoggerErr
				if err != nil {
					log.Error(err)
				}
			}
		}()
		defer close(chanLoggerErr)

		// Generate an ID for logger
		var loggerID string
		{
			ran := rand.Int63()
			loggerID = fmt.Sprintf("%v", ran)
		}

		os.Mkdir("logs", 0755)
		fileLoggerDocumentServicePreProcess, err := timingutils.NewStartEndFileLogger(loggerID, "logs/tb-ds-preprocess.out", "logs/ta-ds-preprocess.out")
		if err != nil {
			return errors.Wrap(err, "无法为前处理任务创建文件日志器")
		}
		defer fileLoggerDocumentServicePreProcess.Close()

		fileLoggerDocumentServiceBcUpload, err := timingutils.NewStartEndFileLogger(loggerID, "logs/tb-ds-bcupload.out", "logs/ta-ds-bcupload.out")
		if err != nil {
			return errors.Wrap(err, "无法为属性上链创建文件日志器")
		}
		defer fileLoggerDocumentServiceBcUpload.Close()

		fileLoggerDocumentServiceOffchainIpfsUpload, err := timingutils.NewStartEndFileLogger(loggerID, "logs/tb-ds-ipfsupload.out", "logs/ta-ds-ipfsupload.out")
		if err != nil {
			return errors.Wrap(err, "无法为 IPFS 上传任务创建文件日志器")
		}
		defer fileLoggerDocumentServiceOffchainIpfsUpload.Close()

		fileLoggerDocumentServiceBcRetrieval, err := timingutils.NewStartEndFileLogger(loggerID, "logs/tb-ds-bcretrieval.out", "logs/ta-ds-bcretrieval.out")
		if err != nil {
			return errors.Wrap(err, "无法为属性获取任务创建文件日志器")
		}
		defer fileLoggerDocumentServiceBcRetrieval.Close()

		fileLoggerDocumentServiceOffchainIpfsRetrieval, err := timingutils.NewStartEndFileLogger(loggerID, "logs/tb-ds-ipfsretrieval.out", "logs/ta-ds-ipfsretrieval.out")
		if err != nil {
			return errors.Wrap(err, "无法为 IPFS 获取任务创建文件日志器")
		}
		defer fileLoggerDocumentServiceOffchainIpfsRetrieval.Close()

		documentSvc := &service.DocumentService{
			ServiceInfo:                     universalCcServiceInfo,
			DataBCAO:                        dataBCAO,
			KeySwitchBCAO:                   keySwitchBCAO,
			KeySwitchService:                keySwitchSvc,
			FileLoggerPreProcess:            fileLoggerDocumentServicePreProcess,
			FileLoggerBcUpload:              fileLoggerDocumentServiceBcUpload,
			FileLoggerOffchainIpfsUpload:    fileLoggerDocumentServiceOffchainIpfsUpload,
			FileLoggerBcRetrieval:           fileLoggerDocumentServiceBcRetrieval,
			FileLoggerOffchainIpfsRetrieval: fileLoggerDocumentServiceOffchainIpfsRetrieval,
			ChanLoggerErr:                   chanLoggerErr,
		}

		// Instantiate an entity asset service
		entityAssetSvc := &service.EntityAssetService{
			ServiceInfo:      universalCcServiceInfo,
			DataBCAO:         dataBCAO,
			KeySwitchBCAO:    keySwitchBCAO,
			KeySwitchService: keySwitchSvc,
		}

		// Instantiate an auth service
		authSvc := &service.AuthService{
			AuthBCAO: authBCAO,
		}

		// Instantiate an identity service
		identitySvc := &service.IdentityService{
			IdentityBCAO: identityBCAO,
			ServerInfo:   &serverInfo,
		}

		// Prepare a key switch server. It will be of use if the app is enabled as a key switch server.
		ksServer := background.NewKeySwitchServer(universalCcServiceInfo, universalCcEventManager, dataBCAO, keySwitchSvc)
		if isKeySwitchServer {
			// Start the server to listen key switch triggers
			err := ksServer.Start()
			if err != nil {
				return err
			}
		}

		// Prepare a regulator server. It will be of use if the app is enabled as a regulator server.
		regulatorServer := background.NewRegulatorServer(universalCcServiceInfo, universalCcEventManager, dataBCAO, entityAssetSvc)
		if isRegulator {
			// Start the server to listen encrypted resource creation events
			err := regulatorServer.Start()
			if err != nil {
				return err
			}
		}

		// // Make a "transfer" request to transfer 10 screws from "Org1" to "Org2" and show the transaction ID
		// respMsg, err := screwSvc.TransferAndShowEvent("Org1", "Org2", 10)
		// if err != nil {
		// 	log.Fatalln(err)
		// } else {
		// 	fmt.Printf("Transaction ID: %v\n", respMsg)
		// }

		// // Make a "query" request for "Org1" and show the response payload.
		// respMsg, err = screwSvc.Query("Org1")
		// if err != nil {
		// 	log.Fatalln(err)
		// } else {
		// 	fmt.Printf("Screw amount in Org 1 after the transfer: %v\n", respMsg)
		// }

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
			GroupName:   "/",
			DocumentSvc: documentSvc,
		}

		// Instantiate an entity controller
		entityAssetController := &controller.EntityAssetController{
			GroupName:      "/",
			EntityAssetSvc: entityAssetSvc,
		}

		// Instantiate an auth controller
		authController := &controller.AuthController{
			GroupName: "/auth",
			AuthSvc:   authSvc,
		}

		// Instantiate a key switch controller
		keySwitchController := &controller.KeySwitchController{
			GroupName:    "/ks",
			KeySwitchSvc: keySwitchSvc,
		}

		// Instantiate an identity controller
		identityController := &controller.IdentityController{
			GroupName:      "/identity",
			DocumentSvc:    documentSvc,
			EntityAssetSvc: entityAssetSvc,
			AuthSvc:        authSvc,
			IdentitySvc:    identitySvc,
		}

		// Register controller handlers
		router := gin.Default()
		router.Use(controller.CORSMiddleware())
		apiv1Group := router.Group("/api/v1")
		_ = controller.RegisterHandlers(apiv1Group, pingPongController)
		_ = controller.RegisterHandlers(apiv1Group, screwController)
		_ = controller.RegisterHandlers(apiv1Group, documentController)
		_ = controller.RegisterHandlers(apiv1Group, entityAssetController)
		_ = controller.RegisterHandlers(apiv1Group, authController)
		_ = controller.RegisterHandlers(apiv1Group, keySwitchController)
		_ = controller.RegisterHandlers(apiv1Group, identityController)

		// Start the HTTP server
		log.Infoln(fmt.Sprintf("正在端口 %v 上启动 HTTP 服务器...", serverInfo.Port))
		httpServer := &http.Server{
			Addr:    fmt.Sprintf(":%v", serverInfo.Port),
			Handler: router,
		}

		chanError := make(chan error)
		go func() {
			if err := httpServer.ListenAndServe(); err != nil {
				chanError <- errors.Wrap(err, "无法启动 HTTP 服务器")
				return
			}
		}()

		// Listen Ctrl+C signals. On receiving a signal stops the app elegantly
		chanQuit := make(chan os.Signal, 1)
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

// Seed the rand generators appropriately
// Solution from https://stackoverflow.com/a/54491783/7616443
func init() {
	var b [8]byte
	_, err := cryptorand.Read(b[:])
	if err != nil {
		panic("cannot seed math/rand package with cryptographically secure random number generator")
	}
	mathrand.Seed(int64(binary.LittleEndian.Uint64(b[:])))
}
