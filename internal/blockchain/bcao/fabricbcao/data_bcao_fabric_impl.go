package fabricbcao

import (
	"encoding/json"
	"fmt"
	"strconv"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/bcao"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/chaincodectx"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/models/common"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/errorcode"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/data"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/query"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type DataBCAOFabricImpl struct {
	ctx *chaincodectx.FabricChaincodeCtx
}

func NewDataBCAOFabricImpl(ctx *chaincodectx.FabricChaincodeCtx) *DataBCAOFabricImpl {
	return &DataBCAOFabricImpl{
		ctx: ctx,
	}
}

func (o *DataBCAOFabricImpl) CreatePlainData(plainData *data.PlainData, eventID ...string) (string, error) {
	plainDataBytes, err := json.Marshal(plainData)
	if err != nil {
		return "", errors.Wrap(err, "无法序列化链码参数")
	}

	chaincodeFcn := "createPlainData"
	chaincodeArgs := [][]byte{plainDataBytes}
	if len(eventID) != 0 {
		chaincodeArgs = append(chaincodeArgs, []byte(eventID[0]))
	}
	channelReq := channel.Request{
		ChaincodeID: o.ctx.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        chaincodeArgs,
	}

	resp, err := o.ctx.ChannelClient.Execute(channelReq)
	if err != nil {
		return "", bcao.GetClassifiedError(chaincodeFcn, err)
	} else {
		return string(resp.TransactionID), nil
	}
}

func (o *DataBCAOFabricImpl) CreateEncryptedData(encryptedData *data.EncryptedData, eventID ...string) (string, error) {
	encryptedDataBytes, err := json.Marshal(encryptedData)
	if err != nil {
		return "", errors.Wrapf(err, "无法序列化链码参数")
	}

	chaincodeFcn := "createEncryptedData"
	chaincodeArgs := [][]byte{encryptedDataBytes}
	if len(eventID) != 0 {
		chaincodeArgs = append(chaincodeArgs, []byte(eventID[0]))
	}
	channelReq := channel.Request{
		ChaincodeID: o.ctx.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        chaincodeArgs,
	}

	resp, err := executeChannelRequestWithTimer(o.ctx.ChannelClient, &channelReq, "链上存储文档")
	if err != nil {
		return "", bcao.GetClassifiedError(chaincodeFcn, err)
	} else {
		return string(resp.TransactionID), nil
	}
}

func (o *DataBCAOFabricImpl) CreateOffchainData(offchainData *data.OffchainData, eventID ...string) (string, error) {
	offchainDataBytes, err := json.Marshal(offchainData)
	if err != nil {
		return "", errors.Wrapf(err, "无法序列化链码参数")
	}

	chaincodeFcn := "createOffchainData"
	chaincodeArgs := [][]byte{offchainDataBytes}
	if len(eventID) != 0 {
		chaincodeArgs = append(chaincodeArgs, []byte(eventID[0]))
	}
	channelReq := channel.Request{
		ChaincodeID: o.ctx.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        chaincodeArgs,
	}

	resp, err := executeChannelRequestWithTimer(o.ctx.ChannelClient, &channelReq, "链上存储文档元数据与属性")
	if err != nil {
		return "", bcao.GetClassifiedError(chaincodeFcn, err)
	}

	return string(resp.TransactionID), nil
}

func (o *DataBCAOFabricImpl) GetMetadata(resourceID string) (*data.ResMetadataStored, error) {
	chaincodeFcn := "getMetadata"
	channelReq := channel.Request{
		ChaincodeID: o.ctx.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{[]byte(resourceID)},
	}

	resp, err := o.ctx.ChannelClient.Query(channelReq)
	if err != nil {
		return nil, bcao.GetClassifiedError(chaincodeFcn, err)
	}

	metadataBytes := resp.Payload
	var resMetadataStored data.ResMetadataStored
	if err = json.Unmarshal(metadataBytes, &resMetadataStored); err != nil {
		return nil, errors.Wrap(err, "获取的元数据不合法")
	}
	return &resMetadataStored, nil
}

func (o *DataBCAOFabricImpl) GetData(resourceID string) ([]byte, error) {
	chaincodeFcn := "getData"
	channelReq := channel.Request{
		ChaincodeID: o.ctx.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{[]byte(resourceID)},
	}

	resp, err := o.ctx.ChannelClient.Query(channelReq)
	if err != nil {
		return nil, bcao.GetClassifiedError(chaincodeFcn, err)
	}

	dataBytes := resp.Payload
	return dataBytes, nil
}

func (o *DataBCAOFabricImpl) GetKey(resourceID string) ([]byte, error) {
	chaincodeFcn := "getKey"
	channelReq := channel.Request{
		ChaincodeID: o.ctx.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{[]byte(resourceID)},
	}

	resp, err := o.ctx.ChannelClient.Query(channelReq)
	if err != nil {
		return nil, bcao.GetClassifiedError(chaincodeFcn, err)
	}

	encryptedKeyBytes := resp.Payload
	return encryptedKeyBytes, nil
}

func (o *DataBCAOFabricImpl) GetPolicy(resourceID string) ([]byte, error) {
	// TODO
	return nil, errorcode.ErrorNotImplemented
}

func (o *DataBCAOFabricImpl) ListResourceIDsByCreator(dataType string, isDesc bool, pageSize int, bookmark string) (*query.IDsWithPagination, error) {
	chaincodeFcn := "listResourceIDsByCreator"
	channelReq := channel.Request{
		ChaincodeID: o.ctx.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{[]byte(dataType), []byte(fmt.Sprintf("%v", isDesc)), []byte(strconv.Itoa(pageSize)), []byte(bookmark)},
	}

	resp, err := o.ctx.ChannelClient.Query(channelReq)
	err = bcao.GetClassifiedError(chaincodeFcn, err)
	if err != nil {
		return nil, err
	}

	var resourceIDs query.IDsWithPagination
	err = json.Unmarshal(resp.Payload, &resourceIDs)
	if err != nil {
		return nil, errors.Wrap(err, "无法解析结果列表")
	}

	return &resourceIDs, nil
}

func (o *DataBCAOFabricImpl) ListResourceIDsByConditions(conditions common.QueryConditions, pageSize int) (*query.IDsWithPagination, error) {
	// 为链码层所用的 CouchDB 生成查询条件
	couchDBConditions, err := conditions.ToCouchDBConditions()
	if err != nil {
		return nil, err
	}

	conditionsBytes, err := json.Marshal(couchDBConditions)
	if err != nil {
		return nil, errors.Wrap(err, "无法序列化查询条件")
	}

	// TODO: Debug 用
	log.Debug(string(conditionsBytes))

	// 为满足 3 个参数，最后的 bookmark 参数为空列表
	chaincodeFcn := "listResourceIDsByConditions"
	channelReq := channel.Request{
		ChaincodeID: o.ctx.ChaincodeID,
		Fcn:         chaincodeFcn,
		Args:        [][]byte{conditionsBytes, []byte(strconv.Itoa(pageSize)), {}},
	}

	resp, err := o.ctx.ChannelClient.Query(channelReq)
	err = bcao.GetClassifiedError(chaincodeFcn, err)
	if err != nil {
		return nil, err
	}

	// 这里虽然包含查询后的新书签信息，但该书签信息无用
	var chaincodeResourceIDs query.IDsWithPagination
	err = json.Unmarshal(resp.Payload, &chaincodeResourceIDs)
	if err != nil {
		return nil, errors.Wrap(err, "无法解析结果列表")
	}

	return &chaincodeResourceIDs, nil
}
