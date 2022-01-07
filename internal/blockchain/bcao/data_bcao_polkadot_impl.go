package bcao

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/contractctx"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/data"
	"github.com/pkg/errors"
)

type DataBCAOPolkadotImpl struct {
	ctx    *contractctx.PolkadotContractCtx
	client *http.Client
}

func NewDataBCAOSubstrateImpl() *DataBCAOPolkadotImpl {
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	return &DataBCAOPolkadotImpl{
		client: client,
	}
}

func (o *DataBCAOPolkadotImpl) CreatePlainData(plainData data.PlainData, eventID *string) (string, error) {
	endpoint := o.ctx.APIPrefix + "/contract/tx"
	form := url.Values{}
	fillPostFormBase(o.ctx, &form)
	form.Add("funcName", "createPlainData")

	argsMap := make(map[string]interface{})
	argsMap["plainData"] = plainData
	argsMap["eventID"] = eventID
	argsBytes, err := json.Marshal(argsMap)
	if err != nil {
		return "", errors.Wrap(err, "无法序列化合约参数")
	}
	form.Add("funcArgs", string(argsBytes))

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", errors.Wrap(err, "无法调用合约")
	}

	resp, err := o.client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "无法调用合约")
	}
	defer resp.Body.Close()

	// TODO: Process the returned body
	body, _ := ioutil.ReadAll(resp.Body)
	return string(body), nil
}
