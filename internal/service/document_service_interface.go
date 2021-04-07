package service

import (
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/models/common"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/data"
	"github.com/XiaoYao-austin/ppks"
)

// DocumentServiceInterface 定义了用于管理数字文档的服务的接口。
type DocumentServiceInterface interface {
	// 创建数字文档
	//
	// 参数：
	//   文档 ID
	//   文档名称
	//   文档内容
	//   文档属性（JSON）
	//
	// 返回：
	//   交易 ID
	CreateDocument(id string, name string, contents []byte, property string) (string, error)

	// 创建加密数字文档
	//
	// 参数：
	//   文档 ID
	//   文档名称
	//   文档内容
	//   文档属性（JSON）
	//   对称密钥（SM2 曲线上的点）
	//   访问策略
	//
	// 返回：
	//   交易 ID
	CreateEncryptedDocument(id string, name string, contents []byte, property string, key *ppks.CurvePoint, policy string) (string, error)

	// 创建监管者加密数字文档
	//
	// 参数：
	//   文档 ID
	//   文档名称
	//   文档内容
	//   文档属性（JSON）
	//   对称密钥（SM2 曲线上的点）
	//
	// 返回：
	//   交易 ID
	CreateRegulatorEncryptedDocument(id string, name string, contents []byte, property string, key *ppks.CurvePoint) (string, error)

	// 创建链下加密数字文档
	//
	// 参数：
	//   文档 ID
	//   文档名称
	//   文档属性（JSON）
	//   对称密钥（SM2 曲线上的点）
	//   访问策略
	//
	// 返回：
	//   交易 ID
	CreateOffchainDocument(id string, name string, property string, key *ppks.CurvePoint, policy string) (string, error)

	// 获取数字文档的元数据
	//
	// 参数：
	//   文档 ID
	//
	// 返回：
	//   元数据
	GetDocumentMetadata(id string) (*data.ResMetadataStored, error)

	// 获取明文数字文档
	//
	// 参数：
	//   文档 ID
	//
	// 返回：
	//   文档本体
	GetDocument(id string) (*common.Document, error)

	// 获取加密数字文档。提供密钥置换会话，函数将使用密钥置换结果尝试进行解密后，返回明文。
	//
	// 参数：
	//   文档 ID
	//   密钥置换会话 ID
	//   预期的份额数量
	//
	// 返回：
	//   解密后的文档
	GetEncryptedDocument(id string, keySwitchSessionID string, numSharesExpected int) (*common.Document, error)

	// 获取由监管者公钥加密的文档。函数将获取数据本体并尝试使用调用者的公钥解密后，返回明文。
	//
	// 参数：
	//   文档 ID
	//
	//  返回：
	//    解密后的文档
	GetRegulatorEncryptedDocument(id string) (*common.Document, error)

	// 获取所有调用者创建的数字文档的资源 ID。
	//
	// 返回：
	//   资源 ID 列表
	ListDocumentIDsByCreator() ([]string, error)

	// 获取名称包含所提供的部分名称的数字文档的资源 ID。
	//
	// 返回：
	//   资源 ID 列表
	ListDocumentIDsByPartialName(partialName string) ([]string, error)
}
