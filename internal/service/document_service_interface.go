package service

import (
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/models/common"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/data"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/query"
	"github.com/XiaoYao-austin/ppks"
)

// DocumentServiceInterface 定义了用于管理数字文档的服务的接口。
type DocumentServiceInterface interface {
	// 创建数字文档
	//
	// 参数：
	//   数字文档
	//
	// 返回：
	//   交易 ID
	CreateDocument(document *common.Document) (string, error)

	// 创建加密数字文档
	//
	// 参数：
	//   数字文档
	//   对称密钥（SM2 曲线上的点）
	//   访问策略
	//
	// 返回：
	//   交易 ID
	CreateEncryptedDocument(document *common.Document, key *ppks.CurvePoint, policy string) (string, error)

	// 创建链下加密数字文档
	//
	// 参数：
	//   数字文档
	//   对称密钥（SM2 曲线上的点）
	//   访问策略
	//
	// 返回：
	//   交易 ID
	CreateOffchainDocument(document *common.Document, key *ppks.CurvePoint, policy string) (string, error)

	// 获取数字文档的元数据
	//
	// 参数：
	//   文档 ID
	//
	// 返回：
	//   元数据
	GetDocumentMetadata(id string) (*data.ResMetadataStored, error)

	// 获取明文数字文档，调用前应先获取元数据。
	//
	// 参数：
	//   文档 ID
	//   文档元数据
	//
	// 返回：
	//   文档本体
	GetDocument(id string, metadata *data.ResMetadataStored) (*common.Document, error)

	// 获取加密数字文档。提供密钥置换会话，函数将使用密钥置换结果尝试进行解密后，返回明文。调用前应先获取元数据。
	//
	// 参数：
	//   文档 ID
	//   密钥置换会话 ID
	//   预期的份额数量
	//   文档元数据
	//
	// 返回：
	//   解密后的文档
	GetEncryptedDocument(id string, keySwitchSessionID string, numSharesExpected int, metadata *data.ResMetadataStored) (*common.Document, error)

	// GetDecryptedDocumentFromDB 从数据库中获取经解密的数字文档。返回解密后的明文。调用前应先获取元数据。
	//
	// 参数：
	//   文档 ID
	//   文档元数据
	//
	// 返回：
	//   解密后的文档
	GetDecryptedDocumentFromDB(id string, metadata *data.ResMetadataStored) (*common.Document, error)

	// 获取所有调用者创建的数字文档的资源 ID。
	//
	// 参数：
	//   分页大小
	//   分页书签
	//
	// 返回：
	//   带分页的资源 ID 列表
	ListDocumentIDsByCreator(pageSize int, bookmark string) (*query.ResourceIDsWithPagination, error)

	// 获取名称包含所提供的部分名称的数字文档的资源 ID。
	//
	// 参数：
	//   部分名称
	//   分页大小
	//   分页书签
	//
	// 返回：
	//   带分页的资源 ID 列表
	ListDocumentIDsByPartialName(partialName string, pageSize int, bookmark string) (*query.ResourceIDsWithPagination, error)
}
