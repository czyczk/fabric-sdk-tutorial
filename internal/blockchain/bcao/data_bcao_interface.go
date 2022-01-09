package bcao

import "gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/data"

type DataBCAOInterface interface {
	CreatePlainData(plainData data.PlainData, eventID *string) (string, error)
	GetData(resourceID string) ([]byte, error)
	// TODO: GetMetadata
}
