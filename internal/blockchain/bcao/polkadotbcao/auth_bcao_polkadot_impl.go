package polkadotbcao

import (
	"encoding/json"
	"net/http"
	"time"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/bcao"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/chaincodectx"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/utils/idutils"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/auth"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/query"
	"github.com/pkg/errors"
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
	funcName := "createAuthRequest"

	id, err := idutils.GenerateSnowflakeId()
	if err != nil {
		return "", errors.Wrap(err, "无法为访问授权申请生成 ID")
	}

	funcArgs := []interface{}{id, authRequest}
	if len(eventID) != 0 {
		funcArgs = append(funcArgs, eventID[0])
	} else {
		funcArgs = append(funcArgs, nil)
	}

	_, err = sendTx(o.ctx, o.client, funcName, funcArgs, true)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (o *AuthBCAOPolkadotImpl) CreateAuthResponse(authResponse *auth.AuthResponse, eventID ...string) (string, error) {
	funcName := "createAuthResponse"

	id := authResponse.AuthSessionID

	funcArgs := []interface{}{id, authResponse}
	if len(eventID) != 0 {
		funcArgs = append(funcArgs, eventID[0])
	} else {
		funcArgs = append(funcArgs, nil)
	}

	result, err := sendTx(o.ctx, o.client, funcName, funcArgs, true)
	if err != nil {
		return "", err
	}

	return result.TxHash, nil
}

func (o *AuthBCAOPolkadotImpl) GetAuthRequest(authSessionID string) (*auth.AuthRequestStored, error) {
	funcName := "getAuthRequest"
	funcArgs := []interface{}{authSessionID}
	result, err := sendQuery(o.ctx, o.client, funcName, funcArgs, false)
	if err != nil {
		return nil, bcao.GetClassifiedError(funcName, err)
	}

	authRequestBytes, err := unwrapOk(result.Output)
	if err != nil {
		return nil, err
	}

	var authRequestStored auth.AuthRequestStored
	if err = json.Unmarshal(authRequestBytes, &authRequestStored); err != nil {
		return nil, errors.Wrap(err, "获取的访问权申请不合法")
	}
	return &authRequestStored, nil
}

func (o *AuthBCAOPolkadotImpl) GetAuthResponse(authSessionID string) (*auth.AuthResponseStored, error) {
	funcName := "getAuthResponse"
	funcArgs := []interface{}{authSessionID}
	result, err := sendQuery(o.ctx, o.client, funcName, funcArgs, false)
	if err != nil {
		return nil, bcao.GetClassifiedError(funcName, err)
	}

	authResponseBytes, err := unwrapOk(result.Output)
	if err != nil {
		return nil, err
	}

	var authResponseStored auth.AuthResponseStored
	if err = json.Unmarshal(authResponseBytes, &authResponseStored); err != nil {
		return nil, errors.Wrap(err, "获取的访问权响应不合法")
	}
	return &authResponseStored, nil
}

func (o *AuthBCAOPolkadotImpl) ListPendingAuthSessionIDsByResourceCreator(pageSize int, bookmark string) (*query.IDsWithPagination, error) {
	funcName := "listPendingAuthSessionIDsByResourceCreator"
	funcArgs := []interface{}{pageSize, bookmark}
	result, err := sendQuery(o.ctx, o.client, funcName, funcArgs, false)
	if err != nil {
		return nil, bcao.GetClassifiedError(funcName, err)
	}

	pendingAuthSessionIDsBytes, err := unwrapOk(result.Output)
	if err != nil {
		return nil, err
	}

	var resourceIDs query.IDsWithPagination
	err = json.Unmarshal(pendingAuthSessionIDsBytes, &resourceIDs)
	if err != nil {
		return nil, errors.Wrap(err, "无法解析结果列表")
	}

	return &resourceIDs, nil
}

func (o *AuthBCAOPolkadotImpl) ListAuthSessionIDsByRequestor(pageSize int, bookmark string, isLatestFirst bool) (*query.IDsWithPagination, error) {
	funcName := "listAuthSessionIDsByRequestor"
	funcArgs := []interface{}{pageSize, bookmark, isLatestFirst}
	result, err := sendQuery(o.ctx, o.client, funcName, funcArgs, false)
	if err != nil {
		return nil, bcao.GetClassifiedError(funcName, err)
	}

	authSessionIDsBytes, err := unwrapOk(result.Output)
	if err != nil {
		return nil, err
	}

	var resourceIDs query.IDsWithPagination
	err = json.Unmarshal(authSessionIDsBytes, &resourceIDs)
	if err != nil {
		return nil, errors.Wrap(err, "无法解析结果列表")
	}

	return &resourceIDs, nil
}
