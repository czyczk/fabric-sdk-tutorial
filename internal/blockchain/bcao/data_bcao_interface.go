package bcao

import "gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/data"

type IDataBCAO interface {
	CreatePlainData(plainData *data.PlainData, eventID ...string) (string, error)
	CreateEncryptedData(encryptedData *data.EncryptedData, eventID ...string) (string, error)
	CreateOffchainData(offchainData *data.OffchainData, eventID ...string) (string, error)
	GetMetadata(resourceID string) ([]byte, error)
	GetData(resourceID string) ([]byte, error)
	GetKey(resourceID string) ([]byte, error)
	GetPolicy(resourceID string) ([]byte, error)
	ListResourceIDsByCreator(dataType string, isDesc bool, pageSize int, bookmark string) ([]byte, error)
	ListResourceIDsByConditions(queryConditions map[string]interface{}, pageSize int, bookmark string) ([]byte, error)
}
