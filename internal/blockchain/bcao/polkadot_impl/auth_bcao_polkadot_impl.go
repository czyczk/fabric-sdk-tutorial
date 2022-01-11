package polkadot

import (
	"net/http"
	"time"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/chaincodectx"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/errorcode"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/auth"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/query"
)

type AuthBCAOPolkadotImpl struct {
	ctx    *chaincodectx.PolkadotChaincodeCtx
	client *http.Client
}

func NewAuthBCAOPolkadotImpl(ctx *chaincodectx.PolkadotChaincodeCtx) *AuthBCAOPolkadotImpl {
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	return &AuthBCAOPolkadotImpl{
		ctx:    ctx,
		client: client,
	}
}

func (o *AuthBCAOPolkadotImpl) CreateAuthRequest(authRequest *auth.AuthRequest, eventID ...string) (string, error) {
	// TODO
	return "", errorcode.ErrorNotImplemented
}

func (o *AuthBCAOPolkadotImpl) CreateAuthResponse(authResponse *auth.AuthResponse, eventID ...string) (string, error) {
	// TODO
	return "", errorcode.ErrorNotImplemented
}

func (o *AuthBCAOPolkadotImpl) GetAuthRequest(authSessionID string) (*auth.AuthRequestStored, error) {
	// TODO
	return nil, errorcode.ErrorNotImplemented
}

func (o *AuthBCAOPolkadotImpl) GetAuthResponse(authSessionID string) (*auth.AuthResponseStored, error) {
	// TODO
	return nil, errorcode.ErrorNotImplemented
}

func (o *AuthBCAOPolkadotImpl) ListPendingAuthSessionIDsByResourceCreator(pageSize int, bookmark string) (*query.IDsWithPagination, error) {
	// TODO
	return nil, errorcode.ErrorNotImplemented
}

func (o *AuthBCAOPolkadotImpl) ListAuthSessionIDsByRequestor(pageSize int, bookmark string, isLatestFirst bool) (*query.IDsWithPagination, error) {
	// TODO
	return nil, errorcode.ErrorNotImplemented
}
