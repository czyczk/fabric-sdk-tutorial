package polkadotbcao

import (
	"net/http"
	"time"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/chaincodectx"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/errorcode"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/identity"
)

type IdentityBCAOPolkadotImpl struct {
	ctx    *chaincodectx.PolkadotChaincodeCtx
	client *http.Client
}

func NewIdentityBCAOPolkadotImpl(ctx *chaincodectx.PolkadotChaincodeCtx) *IdentityBCAOPolkadotImpl {
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	return &IdentityBCAOPolkadotImpl{
		ctx:    ctx,
		client: client,
	}
}

func (o *IdentityBCAOPolkadotImpl) GetDepartmentIdentity() (*identity.DepartmentIdentityStored, error) {
	// TODO
	return nil, errorcode.ErrorNotImplemented
}
