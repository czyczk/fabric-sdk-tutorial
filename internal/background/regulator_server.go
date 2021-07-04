package background

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/db"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/global"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/models/common"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/service"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/utils/cipherutils"
	"github.com/XiaoYao-austin/ppks"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type RegulatorServer struct {
	ServiceInfo        *service.Info
	DocumentService    service.DocumentServiceInterface
	EntityAssetService service.EntityAssetServiceInterface
	wg                 sync.WaitGroup
	chanQuit           chan int
	chanResourceID     chan string
	reg                *fab.Registration
	serverStatus       *backgroundServerStatus
}

func NewRegulatorServer(serviceInfo *service.Info, documentService service.DocumentServiceInterface, entityAssetService service.EntityAssetServiceInterface) *RegulatorServer {
	return &RegulatorServer{
		ServiceInfo:        serviceInfo,
		DocumentService:    documentService,
		EntityAssetService: entityAssetService,
		wg:                 sync.WaitGroup{},
		chanQuit:           make(chan int),
		chanResourceID:     make(chan string),
		reg:                nil,
		serverStatus:       newBackgroundServerStatus(),
	}
}

// Start starts the regulator server to listen resource creation events. The resource ID will be used to fetch document properties and entity assets.
func (s *RegulatorServer) Start() error {
	// Don't start the server again if it has been started.
	log.Infoln("正在启动监管者服务器...")

	if s.serverStatus.getIsStarting() {
		return fmt.Errorf("监管者服务器正在启动")
	} else if s.serverStatus.getIsStarted() {
		return fmt.Errorf("监管者服务器已启动")
	}

	s.serverStatus.setIsStarting(true)

	// Register the event chaincode and pass the chan object to the workers to be created.
	eventID := "enc_res_creation"
	log.Debugf("正在尝试监听事件 '%v'...", eventID)
	reg, notifier, err := service.RegisterEvent(s.ServiceInfo.EventClient, s.ServiceInfo.ChaincodeID, eventID)
	if err != nil {
		s.ServiceInfo.ChannelClient.UnregisterChaincodeEvent(reg)
		return errors.Wrap(err, "无法监听加密资源创建事件")
	}

	s.reg = &reg

	// Start #NumWorkers Go routines with each running a worker.
	log.Debugf("正在创建监管者服务工作单元...")
	s.wg.Add(1)
	go s.createRegulatorServerWorker(notifier)

	s.serverStatus.setIsStarting(false)
	s.serverStatus.setIsStarted(true)
	log.Infoln("监管者服务器已启动。")

	return nil
}

func (s *RegulatorServer) createRegulatorServerWorker(chanResourceIDNotifier <-chan *fab.CCEvent) {
	log.Debugf("监管者服务工作单元已创建。")

workerLoop:
	for {
		select {
		case event := <-chanResourceIDNotifier:
			timeDiffBefore := time.Now()
			// On receiving a resource ID, fetch the resource metadata and act according to the data type
			// First parse the event payload
			resourceID := string(event.Payload)
			if resourceID == "" {
				log.Errorln("监管者服务工作单元无法解析事件内容")
				continue
			}

			log.Debugf("监管者服务工作单元收到加密资源创建事件，资源 ID: %v。", resourceID)

			// Fetch the resource metadata with the service function
			resourceMetadata, err := s.DocumentService.GetDocumentMetadata(resourceID)
			if err != nil {
				log.Errorln(errors.Wrapf(err, "监管者服务工作单元无法获取资源元数据，资源 ID: %v", resourceID))
				continue
			}

			// Fetch the encrypted key of the resource
			chaincodeFcn := "getKey"
			channelReq := channel.Request{
				ChaincodeID: s.ServiceInfo.ChaincodeID,
				Fcn:         chaincodeFcn,
				Args:        [][]byte{[]byte(resourceID)},
			}

			resp, err := s.ServiceInfo.ChannelClient.Query(channelReq)
			err = service.GetClassifiedError(chaincodeFcn, err)
			if err != nil {
				log.Errorln(errors.Wrapf(err, "监管者服务工作单元无法获取资源的加密密钥，资源 ID: %v", resourceID))
				continue
			}

			encryptedKey := resp.Payload

			// Decrypt the encrypted resource key
			symmetricKeyBytes, err := s.decryptResourceKey(encryptedKey)
			if err != nil {
				log.Errorln(err)
				continue
			}

			// Act according to the data type
			dataType, ok := resourceMetadata.Extensions["dataType"]
			if !ok {
				log.Errorf("监管者服务工作单元无法获取资源的数据类型，资源 ID: %v", resourceID)
			}

			switch dataType {
			case "document":
				// Decrypt only the document properties in the metadata
				encryptedDocumentPropertiesBytes, err := base64.StdEncoding.DecodeString(resourceMetadata.Extensions["encrypted"].(string))
				if err != nil {
					log.Errorln(errors.Wrapf(err, "监管者服务工作单元无法获取资源的加密属性，资源 ID: %v", resourceID))
					continue
				}

				documentProperties, err := s.decryptDocumentProperties(encryptedDocumentPropertiesBytes, symmetricKeyBytes)
				if err != nil {
					log.Errorln(err)
					continue
				}

				// Save the decrypted document properties
				err = db.SaveDecryptedDocumentPropertiesToLocalDB(documentProperties, resourceMetadata.Timestamp, s.ServiceInfo.DB)
				if err != nil {
					log.Errorln(err)
					continue
				}
			case "entityAsset":
				// Fetch the encrypted entity asset from the chaincode function
				chaincodeFcn := "getData"
				channelReq := channel.Request{
					ChaincodeID: s.ServiceInfo.ChaincodeID,
					Fcn:         chaincodeFcn,
					Args:        [][]byte{[]byte(resourceID)},
				}

				resp, err := s.ServiceInfo.ChannelClient.Query(channelReq)
				err = service.GetClassifiedError(chaincodeFcn, err)
				if err != nil {
					log.Errorln(errors.Wrapf(err, "监管者服务工作单元无法获取资源，资源 ID: %v", resourceID))
					continue
				}

				encryptedEntityAssetBytes := resp.Payload

				// Decrypt the entity asset
				entityAsset, err := s.decryptEntityAsset(encryptedEntityAssetBytes, symmetricKeyBytes)
				if err != nil {
					log.Errorln(err)
					continue
				}

				// Save the decrypted entity asset
				err = db.SaveDecryptedEntityAssetToLocalDB(entityAsset, resourceMetadata.Timestamp, s.ServiceInfo.DB)
				if err != nil {
					log.Errorln(err)
					continue
				}
			}

			timeDiffAfter := time.Now()
			timeDiff := timeDiffAfter.Sub(timeDiffBefore)
			log.Debugf("密钥置换工作单元完成事件处理，耗时 %v。资源 ID: %v。", timeDiff, resourceID)
		case <-s.chanQuit:
			// Break the for loop when receiving a quit signal
			log.Debug("密钥置换工作单元收到退出信号。")
			s.wg.Done()
			break workerLoop
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}
}

func (s *RegulatorServer) decryptResourceKey(encryptedKeyBytes []byte) (symmetricKeyBytes []byte, err error) {
	if global.KeySwitchKeys.CollectivePrivateKey == nil {
		err = fmt.Errorf("未指定集合私钥")
		return
	}

	// 反序列化加密后的 key
	encryptedKey, err := cipherutils.DeserializeCipherText(encryptedKeyBytes)
	if err != nil {
		return
	}

	// 用集合私钥解密 encryptedKey
	key, err := ppks.PointDecrypt(encryptedKey, global.KeySwitchKeys.CollectivePrivateKey)
	if err != nil {
		err = errors.Wrapf(err, "无法解密加密的密钥")
		return
	}

	symmetricKeyBytes = cipherutils.DeriveSymmetricKeyBytesFromCurvePoint(key)
	return
}

func (s *RegulatorServer) decryptDocumentProperties(encryptedDocumentPropertiesBytes []byte, symmetricKeyBytes []byte) (documentProperties *common.DocumentProperties, err error) {
	// 用对称密钥解密加密的文档属性
	documentPropertiesBytes, err := cipherutils.DecryptBytesUsingAESKey(encryptedDocumentPropertiesBytes, symmetricKeyBytes)
	if err != nil {
		err = errors.Wrap(err, "无法解密文档属性")
		return
	}

	// 解析解密的文档属性
	err = json.Unmarshal(documentPropertiesBytes, &documentProperties)
	if err != nil {
		err = errors.Wrap(err, "无法解析解密后的文档属性")
		return
	}

	return
}

func (s *RegulatorServer) decryptEntityAsset(encryptedEntityAssetBytes []byte, symmetricKeyBytes []byte) (entityAsset *common.EntityAsset, err error) {
	// 用对称密钥解密加密的实体资产
	entityAssetBytes, err := cipherutils.DecryptBytesUsingAESKey(encryptedEntityAssetBytes, symmetricKeyBytes)
	if err != nil {
		err = errors.Wrap(err, "无法解密实体资产")
		return
	}

	// 解析解密的实体资产
	err = json.Unmarshal(entityAssetBytes, &entityAsset)
	if err != nil {
		err = errors.Wrap(err, "无法解析解密后的实体资产")
		return
	}

	return
}
