package ksserver

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/service"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/keyswitch"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/sm2keyutils"
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
		go createKeySwitchServerWorker(id, &s.wg, notifier, s.chanQuit)
	}

	s.isStarting = false
	s.isStarted = true
	log.Infoln("密钥置换服务器已启动。")

	return nil
}

func createKeySwitchServerWorker(id int, wg *sync.WaitGroup, chanKeySwitchSessionIDNotifier <-chan *fab.CCEvent, chanQuit chan int) {
	wg.Add(1)

	for {
		select {
		case event := <-chanKeySwitchSessionIDNotifier:
			// On receiving a key switch session ID, calculate the share and invoke the service function to save the result onto the chain
			keySwitchTriggerStoredBytes := event.Payload
			var keySwitchTriggerStored keyswitch.KeySwitchTriggerStored
			if err := json.Unmarshal(keySwitchTriggerStoredBytes, &keySwitchTriggerStored); err != nil {
				log.Errorf("密钥置换工作单元 #%v 无法解析事件内容。\n", id)
				continue
			}

			log.Debugf("密钥置换工作单元 #%v 收到触发器，会话 ID: %v。\n", id, keySwitchTriggerStored.KeySwitchSessionID)
			// Do share calculation
			targetPubKey, err := sm2keyutils.ConvertBigIntegersToPublicKey
			fmt.Println(keySwitchSessionID)
		case <-chanQuit:
			wg.Done()
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
