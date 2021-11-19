package background

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/global"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/service"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/utils/cipherutils"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/utils/timingutils"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/keyswitch"
	"github.com/XiaoYao-austin/ppks"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type KeySwitchServer struct {
	ServiceInfo            *service.Info
	KeySwitchService       service.KeySwitchServiceInterface
	wg                     sync.WaitGroup
	chanQuit               chan int
	chanKeySwitchSessionID chan string
	NumWorkers             int // The number of Go routines that will be created to perform the task. Don't change the value after creation or the server might not be able to stop as expected.
	reg                    *fab.Registration
	serviceStatus          *backgroundServerStatus
}

func NewKeySwitchServer(serviceInfo *service.Info, keySwitchService service.KeySwitchServiceInterface, numWorkers int) *KeySwitchServer {
	return &KeySwitchServer{
		ServiceInfo:            serviceInfo,
		KeySwitchService:       keySwitchService,
		wg:                     sync.WaitGroup{},
		chanQuit:               make(chan int),
		chanKeySwitchSessionID: make(chan string),
		NumWorkers:             numWorkers,
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
	reg, notifier, err := service.RegisterEvent(s.ServiceInfo.EventClient, s.ServiceInfo.ChaincodeID, eventID)
	if err != nil {
		s.ServiceInfo.ChannelClient.UnregisterChaincodeEvent(reg)
		return errors.Wrap(err, "无法监听密钥置换触发器")
	}

	s.reg = &reg

	// Start #NumWorkers Go routines with each running a worker.
	log.Debugf("正在创建 %v 个密钥置换工作单元...", s.NumWorkers)
	for id := 0; id < s.NumWorkers; id++ {
		s.wg.Add(1)
		go s.createKeySwitchServerWorker(id, notifier)
	}

	s.serviceStatus.setIsStarting(false)
	s.serviceStatus.setIsStarted(true)
	log.Infoln("密钥置换服务器已启动。")

	return nil
}

func (s *KeySwitchServer) createKeySwitchServerWorker(id int, chanKeySwitchSessionIDNotifier <-chan *fab.CCEvent) {
	log.Debugf("密钥置换工作单元 #%v 已创建。", id)

	// Get file descriptors to append timestamps to
	var openedFileDescriptors []*os.File

	filename := "time-before-share.out"
	fShareBefore, err := timingutils.GetFileDescriptorAppendMode(filename)
	if err != nil {
		log.Errorln(errors.Wrapf(err, "无法打开时间戳日志文件 %v", filename))
		return
	}
	openedFileDescriptors = append(openedFileDescriptors, fShareBefore)

	filename = "time-after-share.out"
	fShareAfter, err := timingutils.GetFileDescriptorAppendMode(filename)
	if err != nil {
		log.Errorln(errors.Wrapf(err, "无法打开时间戳日志文件 %v", filename))
		return
	}
	openedFileDescriptors = append(openedFileDescriptors, fShareAfter)

	filename = "time-before-proof.out"
	fProofBefore, err := timingutils.GetFileDescriptorAppendMode(filename)
	if err != nil {
		log.Errorln(errors.Wrapf(err, "无法打开时间戳日志文件 %v", filename))
		return
	}
	openedFileDescriptors = append(openedFileDescriptors, fProofBefore)

	filename = "time-after-proof.out"
	fProofAfter, err := timingutils.GetFileDescriptorAppendMode(filename)
	if err != nil {
		log.Errorln(errors.Wrapf(err, "无法打开时间戳日志文件 %v", filename))
		return
	}
	openedFileDescriptors = append(openedFileDescriptors, fProofAfter)

	filename = "time-before-upload.out"
	fUploadBefore, err := timingutils.GetFileDescriptorAppendMode(filename)
	if err != nil {
		log.Errorln(errors.Wrapf(err, "无法打开时间戳日志文件 %v", filename))
		return
	}
	openedFileDescriptors = append(openedFileDescriptors, fUploadBefore)

	filename = "time-after-upload.out"
	fUploadAfter, err := timingutils.GetFileDescriptorAppendMode(filename)
	if err != nil {
		log.Errorln(errors.Wrapf(err, "无法打开时间戳日志文件 %v", filename))
		return
	}
	openedFileDescriptors = append(openedFileDescriptors, fUploadAfter)

	for _, f := range openedFileDescriptors {
		defer f.Close()
	}

workerLoop:
	for {
		select {
		case event := <-chanKeySwitchSessionIDNotifier:
			// On receiving a key switch session ID, calculate the share and invoke the service function to save the result onto the chain
			// First parse the event payload
			var keySwitchTriggerStored keyswitch.KeySwitchTriggerStored
			if err := json.Unmarshal(event.Payload, &keySwitchTriggerStored); err != nil {
				log.Errorln(errors.Wrapf(err, "密钥置换工作单元 #%v 无法解析事件内容", id))
				continue
			}

			log.Debugf("密钥置换工作单元 #%v 收到触发器，会话 ID: %v。\n", id, keySwitchTriggerStored.KeySwitchSessionID)

			// Check if the validation result is true. Ignore the trigger if it's false.
			if !keySwitchTriggerStored.ValidationResult {
				log.Debugf("密钥置换工作单元 #%v: 未通过验证，将忽略该会话。会话 ID: %v。", id, keySwitchTriggerStored.KeySwitchSessionID)
				continue
			}

			// Parse the target public key
			targetPubKeyBytes, err := base64.StdEncoding.DecodeString(keySwitchTriggerStored.KeySwitchPK)
			if err != nil {
				log.Errorln(errors.Wrapf(err, "密钥置换工作单元 #%v 无法解析目标密钥。会话 ID: %v", id, keySwitchTriggerStored.KeySwitchSessionID))
				continue
			}

			targetPubKey, err := cipherutils.DeserializeSM2PublicKey(targetPubKeyBytes)
			if err != nil {
				log.Errorln(errors.Wrapf(err, "密钥置换工作单元 #%v 无法解析目标密钥。会话 ID: %v", id, keySwitchTriggerStored.KeySwitchSessionID))
				continue
			}

			// Invoke the chaincode function to retrieve the encrypted symmetric key
			encryptedKeyBytes, err := s.getResourceKeyFromCC(keySwitchTriggerStored.ResourceID)
			if err != nil {
				log.Errorln(errors.Wrapf(err, "密钥置换工作单元 #%v 无法获取资源 '%v' 的加密密钥。会话 ID: %v", id, keySwitchTriggerStored.ResourceID, keySwitchTriggerStored.KeySwitchSessionID))
				continue
			}

			curvePoints, err := cipherutils.DeserializeCipherText(encryptedKeyBytes)
			if err != nil {
				log.Errorln(errors.Wrapf(err, "密钥置换工作单元 #%v 无法获取资源 '%v' 的加密密钥。会话 ID: %v", id, keySwitchTriggerStored.ResourceID, keySwitchTriggerStored.KeySwitchSessionID))
				continue
			}

			// Do share calculation
			timeBeforeShareCalc := time.Now()
			share, zkpRi, err := ppks.ShareCal(targetPubKey, &curvePoints.K, global.KeySwitchKeys.PrivateKey) // `zkpRi` for calculating zkp
			if err != nil {
				log.Errorln(errors.Wrapf(err, "密钥置换工作单元 #%v 无法获取用户的密钥置换密钥。会话 ID: %v", id, keySwitchTriggerStored.KeySwitchSessionID))
				continue
			}
			timeAfterShareCalc := time.Now()
			timeDiffShareCalc := timeAfterShareCalc.Sub(timeBeforeShareCalc)
			log.Debugf("密钥置换工作单元 #%v 完成份额计算，耗时 %v。会话 ID: %v。", id, timeDiffShareCalc, keySwitchTriggerStored.KeySwitchSessionID)
			if timestampStr, err := timingutils.SerializeTimestamp(timeBeforeShareCalc); err != nil {
				log.Errorln(err)
			} else {
				timingutils.WriteStringToFile(fmt.Sprintf("%v~%v", id, timestampStr), fShareBefore)
			}
			if timestampStr, err := timingutils.SerializeTimestamp(timeAfterShareCalc); err != nil {
				log.Errorln(err)
			} else {
				timingutils.WriteStringToFile(fmt.Sprintf("%v~%v", id, timestampStr), fShareAfter)
			}

			// Generate a ZKP for the share
			timeBeforeProofGen := time.Now()
			proof := &cipherutils.ZKProof{}
			proof.C, proof.R1, proof.R2, err = ppks.ShareProofGenNoB(zkpRi, global.KeySwitchKeys.PrivateKey, share, targetPubKey, &curvePoints.K)
			if err != nil {
				log.Errorln(errors.Wrapf(err, "密钥置换工作单元 #%v 已为份额生成零知识证明。会话 ID: %v", id, keySwitchTriggerStored.KeySwitchSessionID))
				continue
			}
			timeAfterProofGen := time.Now()
			if timestampStr, err := timingutils.SerializeTimestamp(timeBeforeProofGen); err != nil {
				log.Errorln(err)
			} else {
				timingutils.WriteStringToFile(fmt.Sprintf("%v~%v", id, timestampStr), fProofBefore)
			}
			if timestampStr, err := timingutils.SerializeTimestamp(timeAfterProofGen); err != nil {
				log.Errorln(err)
			} else {
				timingutils.WriteStringToFile(fmt.Sprintf("%v~%v", id, timestampStr), fProofAfter)
			}

			// Invoke the service function to save the result onto the chain
			timeBeforeUploading := time.Now()
			txID, err := s.KeySwitchService.CreateKeySwitchResult(keySwitchTriggerStored.KeySwitchSessionID, share, proof)
			if err != nil {
				log.Errorln(errors.Wrapf(err, "密钥置换工作单元 #%v 无法将份额结果上链。会话 ID: %v", id, keySwitchTriggerStored.KeySwitchSessionID))
				continue
			}
			timeAfterUploading := time.Now()
			timeDiffUploading := timeAfterUploading.Sub(timeBeforeUploading)
			log.Debugf("密钥置换工作单元 #%v 完成份额结果上链，耗时 %v。会话 ID: %v。交易 ID: %v。", id, timeDiffUploading, keySwitchTriggerStored.KeySwitchSessionID, txID)
			if timestampStr, err := timingutils.SerializeTimestamp(timeBeforeUploading); err != nil {
				log.Errorln(err)
			} else {
				timingutils.WriteStringToFile(fmt.Sprintf("%v~%v", id, timestampStr), fUploadBefore)
			}
			if timestampStr, err := timingutils.SerializeTimestamp(timeAfterUploading); err != nil {
				log.Errorln(err)
			} else {
				timingutils.WriteStringToFile(fmt.Sprintf("%v~%v", id, timestampStr), fUploadAfter)
			}
		case <-s.chanQuit:
			// Break the for loop when receiving a quit signal
			log.Debugf("密钥置换工作单元 #%v 收到退出信号。", id)
			s.wg.Done()
			break workerLoop
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}
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

	// Start sending stop signals to all the workers
	for id := 0; id < s.NumWorkers; id++ {
		s.chanQuit <- 0
	}

	// Unregister the chaincode event
	s.ServiceInfo.ChannelClient.UnregisterChaincodeEvent(*s.reg)

	s.serviceStatus.setIsStarted(false)

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
