package bcao

import (
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/auth"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/query"
)

type IAuthBCAO interface {
	CreateAuthRequest(authRequest *auth.AuthRequest, eventID ...string) (string, error)
	CreateAuthResponse(authResponse *auth.AuthResponse, eventID ...string) (string, error)
	GetAuthRequest(authSessionID string) (*auth.AuthRequestStored, error)
	GetAuthResponse(authSessionID string) (*auth.AuthResponseStored, error)
	ListPendingAuthSessionIDsByResourceCreator(pageSize int, bookmark string) (*query.IDsWithPagination, error)
	ListAuthSessionIDsByRequestor(pageSize int, bookmark string, isLatestFirst bool) (*query.IDsWithPagination, error)
}
