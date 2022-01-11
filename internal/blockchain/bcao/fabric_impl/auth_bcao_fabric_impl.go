package fabric

import (
	"encoding/json"
	"fmt"
	"strconv"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/bcao"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/chaincodectx"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/auth"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/query"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/pkg/errors"
)

type AuthBCAOFabricImpl struct {
	ctx *chaincodectx.FabricChaincodeCtx
}

func NewAuthBCAOFabricImpl(ctx *chaincodectx.FabricChaincodeCtx) *AuthBCAOFabricImpl {
	return &AuthBCAOFabricImpl{
		ctx: ctx,
	}
}

func (o *AuthBCAOFabricImpl) CreateAuthRequest(authRequest *auth.AuthRequest, eventID ...string) (string, error) {
	authRequestBytes, err := json.Marshal(authRequest)
	if err != nil {
		return "", errors.Wrap(err, "无法序列化链码参数")
	}

	chaincodeFcn := "createAuthRequest"
	channelReq := channel.Request{
		ChaincodeID: o.ctx.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{authRequestBytes},
	}

	resp, err := o.ctx.ChannelClient.Execute(channelReq)
	if err != nil {
		return "", bcao.GetClassifiedError(chaincodeFcn, err)
	} else {
		return string(resp.TransactionID), nil
	}
}

func (o *AuthBCAOFabricImpl) CreateAuthResponse(authResponse *auth.AuthResponse, eventID ...string) (string, error) {
	authResponseBytes, err := json.Marshal(authResponse)
	if err != nil {
		return "", errors.Wrap(err, "无法序列化链码参数")
	}

	chaincodeFcn := "createAuthResponse"
	channelReq := channel.Request{
		ChaincodeID: o.ctx.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{authResponseBytes},
	}

	resp, err := o.ctx.ChannelClient.Execute(channelReq)
	if err != nil {
		return "", bcao.GetClassifiedError(chaincodeFcn, err)
	} else {
		return string(resp.TransactionID), nil
	}
}

func (o *AuthBCAOFabricImpl) GetAuthRequest(authSessionID string) (*auth.AuthRequestStored, error) {
	chaincodeFcn := "getAuthRequest"
	channelReq := channel.Request{
		ChaincodeID: o.ctx.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{[]byte(authSessionID)},
	}

	resp, err := o.ctx.ChannelClient.Query(channelReq)
	if err != nil {
		return nil, bcao.GetClassifiedError(chaincodeFcn, err)
	}

	var authRequestStored auth.AuthRequestStored
	err = json.Unmarshal(resp.Payload, &authRequestStored)
	if err != nil {
		return nil, errors.Wrap(err, "无法解析授权申请")
	}

	return &authRequestStored, nil
}

func (o *AuthBCAOFabricImpl) GetAuthResponse(authSessionID string) (*auth.AuthResponseStored, error) {
	chaincodeFcn := "getAuthResponse"
	channelReq := channel.Request{
		ChaincodeID: o.ctx.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{[]byte(authSessionID)},
	}

	resp, err := o.ctx.ChannelClient.Query(channelReq)
	if err != nil {
		return nil, bcao.GetClassifiedError(chaincodeFcn, err)
	}

	var authResponseStored auth.AuthResponseStored
	err = json.Unmarshal(resp.Payload, &authResponseStored)
	if err != nil {
		return nil, errors.Wrap(err, "无法解析授权批复")
	}

	return &authResponseStored, nil
}

func (o *AuthBCAOFabricImpl) ListPendingAuthSessionIDsByResourceCreator(pageSize int, bookmark string) (*query.IDsWithPagination, error) {
	chaincodeFcn := "listPendingAuthSessionIDsByResourceCreator"
	channelReq := channel.Request{
		ChaincodeID: o.ctx.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{[]byte(strconv.Itoa(pageSize)), []byte(bookmark)},
	}

	resp, err := o.ctx.ChannelClient.Query(channelReq)
	if err != nil {
		return nil, bcao.GetClassifiedError(chaincodeFcn, err)
	}

	// 解析结果列表
	var result query.IDsWithPagination
	err = json.Unmarshal(resp.Payload, &result)
	if err != nil {
		return nil, errors.Wrap(err, "无法解析结果列表")
	}

	return &result, nil
}

func (o *AuthBCAOFabricImpl) ListAuthSessionIDsByRequestor(pageSize int, bookmark string, isLatestFirst bool) (*query.IDsWithPagination, error) {
	chaincodeFcn := "listAuthSessionIDsByRequestor"
	channelReq := channel.Request{
		ChaincodeID: o.ctx.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{[]byte(strconv.Itoa(pageSize)), []byte(bookmark), []byte(fmt.Sprintf("%v", isLatestFirst))},
	}

	resp, err := o.ctx.ChannelClient.Query(channelReq)
	if err != nil {
		return nil, bcao.GetClassifiedError(chaincodeFcn, err)
	}

	// 解析结果列表
	var result query.IDsWithPagination
	err = json.Unmarshal(resp.Payload, &result)
	if err != nil {
		return nil, errors.Wrap(err, "无法解析结果列表")
	}

	return &result, nil
}
