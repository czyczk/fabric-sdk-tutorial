package service

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/models/common"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/errorcode"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/data"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/pkg/errors"
)

// DocumentService 用于管理数字文档。
type DocumentService struct {
	ServiceInfo *Info
}

// CreateDocument 创建数字文档。
//
// 参数：
//   文档 ID
//   文档名称
//   文档内容
//   文档属性（JSON）
//
// 返回：
//   交易 ID
func (s *DocumentService) CreateDocument(id string, name string, contents []byte, property string) (string, error) {
	// 检查 ID 是否为空。若上层忽略此项检查此项为空，将可能对链码层造成混乱。
	if strings.TrimSpace(id) == "" {
		return "", fmt.Errorf("文档 ID 不能为空。")
	}

	document := common.Document{
		ID:       id,
		Name:     name,
		Contents: contents,
		Property: property,
	}

	documentBytes, err := json.Marshal(document)
	if err != nil {
		return "", errors.Wrapf(err, "无法序列化文档")
	}

	// 计算哈希，获取大小并准备扩展字段
	hash := sha256.Sum256(documentBytes)
	size := len(documentBytes)
	extensions := make(map[string]string)
	extensions["name"] = name
	extensionsBytes, err := json.Marshal(extensions)
	if err != nil {
		return "", errors.Wrapf(err, "无法序列化扩展字段")
	}

	metadata := data.ResMetadata{
		ResourceType: data.Plain,
		ResourceID:   id,
		Hash:         hash,
		Size:         uint64(size),
		Extensions:   string(extensionsBytes),
	}

	// 组装要传入链码的参数，其中数据本体转换为 Base64 编码
	plainData := data.PlainData{
		Metadata: metadata,
		Data:     base64.StdEncoding.EncodeToString(documentBytes),
	}
	plainDataBytes, err := json.Marshal(plainData)
	if err != nil {
		return "", errors.Wrapf(err, "无法序列化链码参数")
	}

	chaincodeFcn := "createPlainData"
	channelReq := channel.Request{
		ChaincodeID: s.ServiceInfo.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{[]byte(plainDataBytes)},
	}

	resp, err := s.ServiceInfo.ChannelClient.Execute(channelReq)
	if err != nil {
		if strings.HasSuffix(err.Error(), errorcode.CodeNotImplemented) {
			return "", errorcode.ErrorNotImplemented
		} else {
			return "", errors.Wrapf(err, "无法调用链码函数 '%v'", chaincodeFcn)
		}
	} else {
		return string(resp.TransactionID), nil
	}
}

// CreateEncryptedDocument 创建加密数字文档。
//
// 参数：
//   文档 ID
//   文档名称
//   文档内容
//   文档属性（JSON）
//   加密后的对称公钥
//   访问策略
//
// 返回：
//   交易 ID
func (s *DocumentService) CreateEncryptedDocument(id string, name string, contents []byte, property string, key []byte, policy string) (string, error) {
	return "", errorcode.ErrorNotImplemented
}

// CreateRegulatorEncryptedDocument 创建监管者加密数字文档。
//
// 参数：
//   文档 ID
//   文档名称
//   文档内容
//   文档属性（JSON）
//   加密后的对称公钥
//
// 返回：
//   交易 ID
func (s *DocumentService) CreateRegulatorEncryptedDocument(id string, name string, contents []byte, property string, key []byte) (string, error) {
	return "", errorcode.ErrorNotImplemented
}

// CreateOffchainDocument 创建链下加密数字文档。
//
// 参数：
//   文档 ID
//   文档名称
//   文档属性（JSON）
//   加密的对称公钥
//   访问策略
//
// 返回：
//   交易 ID
func (s *DocumentService) CreateOffchainDocument(id string, name string, property string, key []byte, policy string) (string, error) {
	return "", errorcode.ErrorNotImplemented
}

// GetDocumentMetadata 获取数字文档的元数据。
//
// 参数：
//   文档 ID
//
// 返回：
//   元数据
func (s *DocumentService) GetDocumentMetadata(id string) (*data.ResMetadataStored, error) {
	chaincodeFcn := "getMetadata"
	channelReq := channel.Request{
		ChaincodeID: s.ServiceInfo.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{[]byte(id)},
	}

	resp, err := s.ServiceInfo.ChannelClient.Execute(channelReq)
	if err != nil {
		if strings.HasSuffix(err.Error(), errorcode.CodeNotFound) {
			return nil, errorcode.ErrorNotFound
		} else if strings.HasSuffix(err.Error(), errorcode.CodeNotImplemented) {
			return nil, errorcode.ErrorNotImplemented
		} else {
			return nil, errors.Wrapf(err, "无法调用链码函数 '%v'", chaincodeFcn)
		}
	} else {
		var resMetadataStored data.ResMetadataStored
		if err = json.Unmarshal(resp.Payload, &resMetadataStored); err != nil {
			return nil, errors.Wrapf(err, "获取的元数据不合法")
		}
		return &resMetadataStored, nil
	}
}

// GetDocument 获取明文数字文档。
//
// 参数：
//   文档 ID
//
// 返回：
//   文档本体
func (s *DocumentService) GetDocument(id string) (*common.Document, error) {
	// 检查元数据中该资源类型是否为明文资源
	resMetadataStored, err := s.GetDocumentMetadata(id)
	if err != nil {
		return nil, err
	}
	if resMetadataStored.ResourceType != data.Plain {
		return nil, fmt.Errorf("该资源不是明文资源")
	}

	chaincodeFcn := "getData"
	channelReq := channel.Request{
		ChaincodeID: s.ServiceInfo.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{[]byte(id)},
	}

	resp, err := s.ServiceInfo.ChannelClient.Execute(channelReq)
	if err != nil {
		if strings.HasSuffix(err.Error(), errorcode.CodeNotFound) {
			return nil, errorcode.ErrorNotFound
		} else if strings.HasSuffix(err.Error(), errorcode.CodeNotImplemented) {
			return nil, errorcode.ErrorNotImplemented
		} else {
			return nil, errors.Wrapf(err, "无法调用链码函数 '%v'", chaincodeFcn)
		}
	} else {
		var document common.Document
		if err = json.Unmarshal(resp.Payload, &document); err != nil {
			return nil, fmt.Errorf("获取的数据不是合法的数字文档")
		}
		return &document, nil
	}
}

// GetEncryptedDocument 获取加密数字文档。提供密钥置换会话，函数将使用密钥置换结果尝试进行解密后，返回明文。
//
// 参数：
//   文档 ID
//   密钥置换会话 ID
//   预期的份额数量
//
// 返回：
//   解密后的文档
func (s *DocumentService) GetEncryptedDocument(id string, keySwitchSessionID string, numSharesExpected int) (*common.Document, error) {
	return nil, fmt.Errorf(errorcode.CodeNotImplemented)
}

// GetRegulatorEncryptedDocument 获取由监管者公钥加密的文档。函数将获取数据本体并尝试使用调用者的公钥解密后，返回明文。
//
// 参数：
//   文档 ID
//
//  返回：
//    解密后的文档
func (s *DocumentService) GetRegulatorEncryptedDocument(id string) (*common.Document, error) {
	return nil, fmt.Errorf(errorcode.CodeNotImplemented)
}

// ListDocumentIDsByCreator 获取所有调用者创建的数字文档的资源 ID。
//
// 返回：
//   资源 ID 列表
func (s *DocumentService) ListDocumentIDsByCreator() ([]string, error) {
	return nil, fmt.Errorf(errorcode.CodeNotImplemented)
}

// ListDocumentIDsByPartialName 获取名称包含所提供的部分名称的数字文档的资源 ID。
//
// 返回：
//   资源 ID 列表
func (s *DocumentService) ListDocumentIDsByPartialName(partialName string) ([]string, error) {
	return nil, fmt.Errorf(errorcode.CodeNotImplemented)
}
