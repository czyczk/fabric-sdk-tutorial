package polkadotbcao

import (
	"encoding/json"
	"net/http"
	"time"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/bcao"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/chaincodectx"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/identity"
	"github.com/pkg/errors"
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
	funcName := "getDepartmentIdentity"
	funcArgs := []interface{}{}
	result, err := sendQuery(o.ctx, o.client, funcName, funcArgs, false)
	if err != nil {
		return nil, bcao.GetClassifiedError(funcName, err)
	}

	deptIdentityBytes, err := unwrapOk(result.Output)
	if err != nil {
		return nil, err
	}

	var deptIdentity identity.DepartmentIdentityStored
	err = json.Unmarshal(deptIdentityBytes, &deptIdentity)
	if err != nil {
		return nil, errors.Wrap(err, "无法解析部门身份信息")
	}

	return &deptIdentity, nil
}
