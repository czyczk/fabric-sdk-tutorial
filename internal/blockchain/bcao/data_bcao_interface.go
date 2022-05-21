package bcao

import (
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/models/common"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/data"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/query"
)

type IDataBCAO interface {
	CreatePlainData(plainData *data.PlainData, eventID ...string) (string, error)
	CreateEncryptedData(encryptedData *data.EncryptedData, eventID ...string) (string, error)
	CreateOffchainData(offchainData *data.OffchainData, eventID ...string) (string, error)
	GetMetadata(resourceID string) (*data.ResMetadataStored, error)
	GetData(resourceID string) ([]byte, error)
	GetKey(resourceID string) ([]byte, error)
	GetPolicy(resourceID string) ([]byte, error)
	ListResourceIDsByCreator(dataType string, isDesc bool, pageSize int, bookmark string) (*query.IDsWithPagination, error)
	ListResourceIDsByConditions(conditions common.QueryConditions, pageSize int) (*query.IDsWithPagination, error)
}
