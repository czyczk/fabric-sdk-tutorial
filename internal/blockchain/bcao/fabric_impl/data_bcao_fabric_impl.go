package fabric

import (
	"encoding/json"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/bcao"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/chaincodectx"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/data"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/pkg/errors"
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

func (o *DataBCAOFabricImpl) GetMetadata(resourceID string) ([]byte, error) {
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
	return metadataBytes, nil
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
