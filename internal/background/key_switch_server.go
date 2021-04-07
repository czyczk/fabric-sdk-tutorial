package ksserver

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"sync"
	"time"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/global"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/service"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/keyswitch"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/sm2keyutils"
	"github.com/XiaoYao-austin/ppks"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type KeySwitchServer struct {
	ServiceInfo            service.Info
	KeySwitchService       *service.KeySwitchServiceInterface
	wg                     sync.WaitGroup
	chanQuit               chan int
	chanKeySwitchSessionID chan string
	NumWorkers             int // The number of Go routines that will be created to perform the task. Don't change the value after creation or the server might not be able to stop as expected.
	reg                    *fab.Registration
	isStarting             bool
	isStarted              bool
	isStopping             bool
}

func NewKeySwitchServer(serviceInfo service.Info, keySwitchService *service.KeySwitchServiceInterface, numWorkers int) *KeySwitchServer {
	return &KeySwitchServer{
		ServiceInfo:            serviceInfo,
		KeySwitchService:       keySwitchService,
		wg:                     sync.WaitGroup{},
		chanQuit:               make(chan int),
		chanKeySwitchSessionID: make(chan string),
		NumWorkers:             numWorkers,
		reg:                    nil,
		isStarting:             false,
		isStarted:              false,
		isStopping:             false,
	}
}

// Start starts the key switch server to listen key switch triggers. The keys to be used by the key switch process must be ensured to have existed before starting.
func (s *KeySwitchServer) Start() error {
	// Don't start the server again if it has been started.
	log.Infoln("正在启动密钥置换服务器...")

	if s.isStarting {
		return fmt.Errorf("密钥置换服务器正在启动")
	} else if s.isStarted {
		return fmt.Errorf("密钥置换服务器已启动")
	}

	s.isStarting = true

	// Register the event chaincode and pass the chan object to the workers to be created.
	eventID := "ks_trigger"
	log.Tracef("正在尝试监听事件 '%v'...\n", eventID)
	reg, notifier, err := service.RegisterEvent(s.ServiceInfo.ChannelClient, s.ServiceInfo.ChaincodeID, eventID)
	if err != nil {
		s.ServiceInfo.ChannelClient.UnregisterChaincodeEvent(reg)
		return errors.Wrap(err, "无法监听密钥置换触发器")
	}

	s.reg = &reg

	// Start #NumWorkers Go routines with each running a worker.
	log.Tracef("正在创建 %v 个密钥置换工作单元...\n", s.NumWorkers)
	for id := 0; id < s.NumWorkers; id++ {
		go s.createKeySwitchServerWorker(id, notifier)
	}

	s.isStarting = false
	s.isStarted = true
	log.Infoln("密钥置换服务器已启动。")

	return nil
}

func (s *KeySwitchServer) createKeySwitchServerWorker(id int, chanKeySwitchSessionIDNotifier <-chan *fab.CCEvent) {
	s.wg.Add(1)

	for {
		select {
		case event := <-chanKeySwitchSessionIDNotifier:
			// On receiving a key switch session ID, calculate the share and invoke the service function to save the result onto the chain
			// First parse the event payload
			var keySwitchTriggerStored keyswitch.KeySwitchTriggerStored
			if err := json.Unmarshal(event.Payload, &keySwitchTriggerStored); err != nil {
				log.Errorln(errors.Wrapf(err, "密钥置换工作单元 #%v 无法解析事件内容\n", id))
				continue
			}

			log.Debugf("密钥置换工作单元 #%v 收到触发器，会话 ID: %v。\n", id, keySwitchTriggerStored.KeySwitchSessionID)

			// Parse the target public key
			targetPubKeyBytes, err := base64.StdEncoding.DecodeString(keySwitchTriggerStored.KeySwitchPK)
			if err != nil || len(targetPubKeyBytes) != 64 {
				log.Errorln(errors.Wrapf(err, "密钥置换工作单元 #%v 无法解析目标密钥", id))
				continue
			}
			if len(targetPubKeyBytes) != 64 {
				log.Errorf("密钥置换工作单元 #%v 无法解析目标密钥: 密钥长度不正确。\n", id)
			}

			targetPubKeyX, targetPubKeyY := big.Int{}, big.Int{}
			_ = targetPubKeyX.SetBytes(targetPubKeyBytes[:32])
			_ = targetPubKeyY.SetBytes(targetPubKeyBytes[32:])

			targetPubKey, err := sm2keyutils.ConvertBigIntegersToPublicKey(&targetPubKeyX, &targetPubKeyY)
			if err != nil {
				log.Errorln(errors.Wrapf(err, "密钥置换工作单元 #%v 无法解析目标密钥", id))
				continue
			}

			// Invoke the chaincode function to retrieve the encrypted symmetric key
			encryptedKeyBytes, err := s.getResourceKeyFromCC(keySwitchTriggerStored.ResourceID)
			if err != nil {
				log.Errorln(errors.Wrapf(err, "密钥置换工作单元 #%v 无法获取资源 '%v' 的加密密钥", id, keySwitchTriggerStored.ResourceID))
				continue
			}
			var curvePoint *ppks.CurvePoint
			curvePointX, curvePointY := big.Int{}, big.Int{}
			_ = curvePointX.SetBytes(encryptedKeyBytes[:32])
			_ = curvePointY.SetBytes(encryptedKeyBytes[32:])

			encryptedKey, err := sm2keyutils.ConvertBigIntegersToPublicKey(&curvePointX, &curvePointY)
			*curvePoint = ppks.CurvePoint(*encryptedKey)

			// Do share calculation
			share, err := ppks.ShareCal(targetPubKey, curvePoint, global.KeySwitchKeys.PrivateKey)

			// TODO: Invoke the service function to save the result onto the chain

		case <-s.chanQuit:
			s.wg.Done()
			break
		default:
			time.Sleep(50 * time.Second)
		}
	}
}

// Stop stops the key switch server from responding to key switch triggers.
//
// Returns
//   a wait group that can be used to block the caller Go routine
func (s *KeySwitchServer) Stop() (*sync.WaitGroup, error) {
	// Don't send stop signals again if the server has already been called to stop.
	if s.isStopping {
		return nil, fmt.Errorf("密钥置换服务器正在停止")
	} else if !s.isStarted {
		return nil, fmt.Errorf("密钥置换服务器已停止")
	}

	s.isStopping = true

	// Start sending stop signals to all the workers
	for id := 0; id < s.NumWorkers; id++ {
		s.chanQuit <- 0
	}

	// Unregister the chaincode event
	s.ServiceInfo.ChannelClient.UnregisterChaincodeEvent(s.reg)

	s.isStarted = false

	return &s.wg, nil
}

func (s *KeySwitchServer) getResourceKeyFromCC(resourceID string) ([]byte, error) {
	chaincodeFcn := "getKey"
	channelReq := channel.Request{
		ChaincodeID: s.ServiceInfo.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{[]byte(resourceID)},
	}

	resp, err := s.ServiceInfo.ChannelClient.Query(channelReq)
	if err != nil {
		return nil, service.GetClassifiedError(chaincodeFcn, err)
	}

	return resp.Payload, nil
}