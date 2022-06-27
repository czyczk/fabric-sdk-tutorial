package background

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/bcao"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/eventmgr"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/global"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/service"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/utils/cipherutils"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/keyswitch"
	"github.com/XiaoYao-austin/ppks"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type KeySwitchServer struct {
	EventManager           eventmgr.IEventManager
	DataBCAO               bcao.IDataBCAO
	KeySwitchService       service.KeySwitchServiceInterface
	wg                     sync.WaitGroup
	chanQuit               chan struct{}
	chanKeySwitchSessionID chan string
	reg                    eventmgr.IEventRegistration
	serviceStatus          *backgroundServerStatus
}

func NewKeySwitchServer(serviceInfo *service.Info, eventManager eventmgr.IEventManager, dataBCAO bcao.IDataBCAO, keySwitchService service.KeySwitchServiceInterface) *KeySwitchServer {
	return &KeySwitchServer{
		EventManager:           eventManager,
		DataBCAO:               dataBCAO,
		KeySwitchService:       keySwitchService,
		wg:                     sync.WaitGroup{},
		chanQuit:               make(chan struct{}),
		chanKeySwitchSessionID: make(chan string),
		reg:                    nil,
		serviceStatus:          newBackgroundServerStatus(),
	}
}

// Start starts the key switch server to listen key switch triggers. The keys to be used by the key switch process must be ensured to have existed before starting.
func (s *KeySwitchServer) Start() error {
	// Don't start the server again if it has been started.
	log.Infoln("正在启动密钥置换服务器...")

	if s.serviceStatus.getIsStarting() {
		return fmt.Errorf("密钥置换服务器正在启动")
	} else if s.serviceStatus.getIsStarted() {
		return fmt.Errorf("密钥置换服务器已启动")
	}

	s.serviceStatus.setIsStarting(true)

	// Register the event chaincode and pass the chan object to the workers to be created.
	eventID := "ks_trigger"
	log.Debugf("正在尝试监听事件 '%v'...", eventID)
	reg, notifier, err := s.EventManager.RegisterEvent(eventID)
	if err != nil {
		s.EventManager.UnregisterEvent(reg)
		return errors.Wrap(err, "无法监听密钥置换触发器")
	}

	s.reg = reg

	// Start a Go routine to run a worker
	log.Debug("正在创建密钥置换工作单元...")
	go s.createKeySwitchServerWorker(notifier, &s.wg)

	s.serviceStatus.setIsStarting(false)
	s.serviceStatus.setIsStarted(true)
	log.Infoln("密钥置换服务器已启动。")

	return nil
}

func (s *KeySwitchServer) createKeySwitchServerWorker(chanKeySwitchSessionIDNotifier <-chan eventmgr.IEvent, wg *sync.WaitGroup) {
	log.Debug("密钥置换工作单元已创建。")

workerLoop:
	for {
		select {
		case event := <-chanKeySwitchSessionIDNotifier:
			s.wg.Add(1)
			go processEvent(event, s)
		case <-s.chanQuit:
			// Break the for loop when receiving a quit signal
			log.Debug("密钥置换工作单元收到退出信号。")
			break workerLoop
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}
}

// Process an event
func processEvent(event eventmgr.IEvent, s *KeySwitchServer) {
	defer s.wg.Done()

	// On receiving a key switch session ID, calculate the share and invoke the service function to save the result onto the chain
	// First parse the event payload
	var keySwitchTriggerStored keyswitch.KeySwitchTriggerStored
	if err := json.Unmarshal(event.GetPayload(), &keySwitchTriggerStored); err != nil {
		log.Errorln(errors.Wrap(err, "密钥置换工作单元无法解析事件内容"))
		return
	}

	log.Debugf("密钥置换工作单元收到触发器，会话 ID: %v。\n", keySwitchTriggerStored.KeySwitchSessionID)

	// Check if the validation result is true. Ignore the trigger if it's false.
	if !keySwitchTriggerStored.ValidationResult {
		log.Debugf("密钥置换工作单元: 未通过验证，将忽略该会话。会话 ID: %v。", keySwitchTriggerStored.KeySwitchSessionID)
		return
	}

	// Parse the target public key
	targetPubKeyBytes, err := base64.StdEncoding.DecodeString(keySwitchTriggerStored.KeySwitchPK)
	if err != nil {
		log.Errorln(errors.Wrapf(err, "密钥置换工作单元无法解析目标密钥。会话 ID: %v", keySwitchTriggerStored.KeySwitchSessionID))
		return
	}

	targetPubKey, err := cipherutils.DeserializeSM2PublicKey(targetPubKeyBytes)
	if err != nil {
		log.Errorln(errors.Wrapf(err, "密钥置换工作单元无法解析目标密钥。会话 ID: %v", keySwitchTriggerStored.KeySwitchSessionID))
		return
	}

	// Invoke the chaincode function to retrieve the encrypted symmetric key
	encryptedKeyBytes, err := s.getResourceKeyFromCC(keySwitchTriggerStored.ResourceID)
	if err != nil {
		log.Errorln(errors.Wrapf(err, "密钥置换工作单元无法获取资源 '%v' 的加密密钥。会话 ID: %v", keySwitchTriggerStored.ResourceID, keySwitchTriggerStored.KeySwitchSessionID))
		return
	}

	curvePoints, err := cipherutils.DeserializeCipherText(encryptedKeyBytes)
	if err != nil {
		log.Errorln(errors.Wrapf(err, "密钥置换工作单元无法获取资源 '%v' 的加密密钥。会话 ID: %v", keySwitchTriggerStored.ResourceID, keySwitchTriggerStored.KeySwitchSessionID))
		return
	}

	// Do share calculation
	timeBeforeShareCalc := time.Now()
	share, zkpRi, err := ppks.ShareCal(targetPubKey, &curvePoints.K, global.KeySwitchKeys.PrivateKey) // `zkpRi` for calculating zkp
	if err != nil {
		log.Errorln(errors.Wrapf(err, "密钥置换工作单元无法获取用户的密钥置换密钥。会话 ID: %v", keySwitchTriggerStored.KeySwitchSessionID))
		return
	}
	timeAfterShareCalc := time.Now()
	timeDiffShareCalc := timeAfterShareCalc.Sub(timeBeforeShareCalc)
	log.Debugf("密钥置换工作单元完成份额计算，耗时 %v。会话 ID: %v。", timeDiffShareCalc, keySwitchTriggerStored.KeySwitchSessionID)

	// Generate a ZKP for the share
	timeBeforeProofGen := time.Now()
	proof := &cipherutils.ZKProof{}
	proof.C, proof.R1, proof.R2, err = ppks.ShareProofGenNoB(zkpRi, global.KeySwitchKeys.PrivateKey, share, targetPubKey, &curvePoints.K)
	timeAfterProofGen := time.Now()
	if err != nil {
		log.Errorln(errors.Wrapf(err, "密钥置换工作单元无法为份额生成零知识证明。会话 ID: %v", keySwitchTriggerStored.KeySwitchSessionID))
		return
	}
	timeDiffProofGen := timeAfterProofGen.Sub(timeBeforeProofGen)
	log.Debugf("密钥置换工作单元完成为份额生成零知识证明，耗时 %v。会话 ID: %v", timeDiffProofGen, keySwitchTriggerStored.KeySwitchSessionID)

	// Invoke the service function to save the result onto the chain
	timeBeforeUploading := time.Now()
	txID, err := s.KeySwitchService.CreateKeySwitchResult(keySwitchTriggerStored.KeySwitchSessionID, share, proof)
	if err != nil {
		log.Errorln(errors.Wrapf(err, "密钥置换工作单元无法将份额结果上链。会话 ID: %v", keySwitchTriggerStored.KeySwitchSessionID))
		return
	}
	timeAfterUploading := time.Now()
	timeDiffUploading := timeAfterUploading.Sub(timeBeforeUploading)
	log.Debugf("密钥置换工作单元完成份额结果上链，耗时 %v。会话 ID: %v。交易 ID: %v。", timeDiffUploading, keySwitchTriggerStored.KeySwitchSessionID, txID)
}

// Stop stops the key switch server from responding to key switch triggers.
//
// Returns
//   a wait group that can be used to block the caller Go routine
func (s *KeySwitchServer) Stop() (*sync.WaitGroup, error) {
	// Don't send stop signals again if the server has already been called to stop.
	if s.serviceStatus.getIsStopping() {
		return nil, fmt.Errorf("密钥置换服务器正在停止")
	} else if !s.serviceStatus.getIsStarted() {
		return nil, fmt.Errorf("密钥置换服务器已停止")
	}

	s.serviceStatus.setIsStopping(true)

	// Send a stop signal to the worker
	s.chanQuit <- struct{}{}

	// Unregister the chaincode event
	s.EventManager.UnregisterEvent(s.reg)

	s.serviceStatus.setIsStarted(false)

	return &s.wg, nil
}

func (s *KeySwitchServer) getResourceKeyFromCC(resourceID string) ([]byte, error) {
	return s.DataBCAO.GetKey(resourceID)
}
