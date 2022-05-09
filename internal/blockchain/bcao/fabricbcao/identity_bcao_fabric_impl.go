package fabricbcao

import (
	"encoding/json"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/bcao"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/chaincodectx"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/identity"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/pkg/errors"
)

type IdentityBCAOFabricImpl struct {
	ctx *chaincodectx.FabricChaincodeCtx
}

func NewIdentityBCAOFabricImpl(ctx *chaincodectx.FabricChaincodeCtx) *IdentityBCAOFabricImpl {
	return &IdentityBCAOFabricImpl{
		ctx: ctx,
	}
}

func (o *IdentityBCAOFabricImpl) GetDepartmentIdentity() (*identity.DepartmentIdentityStored, error) {
	chaincodeFcn := "getDepartmentIdentity"
	channelReq := channel.Request{
		ChaincodeID: o.ctx.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{},
	}

	resp, err := o.ctx.ChannelClient.Query(channelReq)
	if err != nil {
		return nil, bcao.GetClassifiedError(chaincodeFcn, err)
	}

	var deptIdentity identity.DepartmentIdentityStored
	err = json.Unmarshal(resp.Payload, &deptIdentity)
	if err != nil {
		return nil, errors.Wrap(err, "无法解析部门身份信息")
	}

	return &deptIdentity, nil
}
