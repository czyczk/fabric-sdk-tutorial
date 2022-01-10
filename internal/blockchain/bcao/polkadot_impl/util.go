package polkadot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/chaincodectx"
	"github.com/pkg/errors"
)

func fillPostFormBase(ctx *chaincodectx.PolkadotChaincodeCtx, form *url.Values) {
	form.Add("abi", ctx.ContractABI)
	form.Add("contractAddress", ctx.ContractAddress)
	form.Add("signerAddress", ctx.CallerAddress)
}

func fillGetQueryStringBase(ctx *chaincodectx.PolkadotChaincodeCtx, queryString *url.Values) {
	queryString.Add("abi", ctx.ContractABI)
	queryString.Add("contractAddress", ctx.ContractAddress)
	queryString.Add("signerAddress", ctx.CallerAddress)
}

func sendTx(ctx *chaincodectx.PolkadotChaincodeCtx, client *http.Client, funcName string, funcArgs []interface{}, queryIfNoEvent bool) (*ContractTxSuccessResult, error) {
	// Prepare a POST form
	endpoint := ctx.APIPrefix + "/contract/tx"
	form := url.Values{}
	fillPostFormBase(ctx, &form)
	form.Add("funcName", funcName)

	argsBytes, err := json.Marshal(funcArgs)
	if err != nil {
		return nil, errors.Wrap(err, "无法序列化合约参数")
	}
	form.Add("funcArgs", string(argsBytes))

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, errors.Wrap(err, "无法调用合约")
	}

	// Perform a POST request
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "无法调用合约")
	}
	defer resp.Body.Close()

	// Process the response.
	// 200 -> parseTx200
	// 500 -> parseTx500
	// Other -> response body as error message
	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "无法获取合约执行结果")
	}

	if resp.StatusCode == 200 {
		txSuccessResult, err := parseTx200(respBodyBytes)
		if err != nil {
			return nil, errors.Wrapf(err, "无法解析合约执行结果")
		}

		if txSuccessResult.ParsedContractEvents != nil || !queryIfNoEvent {
			return txSuccessResult, nil
		}

		// No event. Perform a query to retrieve the error message.
		querySuccessResult, err := sendQuery(ctx, client, funcName, funcArgs)
		if err != nil {
			return nil, errors.Wrapf(err, "无法获取合约错误消息")
		}

		return txSuccessResult, fmt.Errorf(querySuccessResult.Output)
	} else if resp.StatusCode == 500 {
		txErrorResult, err := parseTx500(respBodyBytes)
		if err != nil {
			return nil, errors.Wrapf(err, "无法解析合约执行结果")
		}

		return nil, fmt.Errorf("合约执行错误: %v", txErrorResult.ExplainedModuleError)
	} else {
		return nil, fmt.Errorf("合约执行错误: %v", string(respBodyBytes))
	}
}

func sendQuery(ctx *chaincodectx.PolkadotChaincodeCtx, client *http.Client, funcName string, funcArgs []interface{}) (*ContractQuerySuccessResult, error) {
	// Prepare a GET query string
	endpoint := ctx.APIPrefix + "/contract/query"
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, errors.Wrap(err, "无法调用合约")
	}

	queryString := req.URL.Query()
	fillGetQueryStringBase(ctx, &queryString)
	queryString.Add("funcName", funcName)

	argsBytes, err := json.Marshal(funcArgs)
	if err != nil {
		return nil, errors.Wrap(err, "无法序列化合约参数")
	}
	queryString.Add("funcArgs", string(argsBytes))

	req.URL.RawQuery = queryString.Encode()

	// Perform a GET request
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "无法调用合约")
	}
	defer resp.Body.Close()

	// Process the response.
	// 200 -> parseQuery200
	// 500 -> parseQuery500
	// Other -> response body as error message
	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "无法获取合约执行结果")
	}

	if resp.StatusCode == 200 {
		querySuccessResult, err := parseQuery200(respBodyBytes)
		if err != nil {
			return nil, errors.Wrapf(err, "无法解析合约执行结果")
		}

		return querySuccessResult, nil
	} else if resp.StatusCode == 500 {
		queryErrorResult, err := parseQuery500(respBodyBytes)
		if err != nil {
			return nil, errors.Wrapf(err, "无法解析合约执行结果")
		}

		if queryErrorResult.DebugMessage != "" {
			return nil, fmt.Errorf("合约执行错误: %v (%v)", queryErrorResult.ExplainedModuleError, queryErrorResult.DebugMessage)
		} else {
			return nil, fmt.Errorf("合约执行错误: %v", queryErrorResult.ExplainedModuleError)
		}
	} else {
		return nil, fmt.Errorf("合约执行错误: %v", string(respBodyBytes))
	}
}

func parseTx200(respBody []byte) (*ContractTxSuccessResult, error) {
	var ret ContractTxSuccessResult
	err := json.Unmarshal(respBody, &ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func parseTx500(respBody []byte) (*ContractTxErrorResult, error) {
	var ret ContractTxErrorResult
	err := json.Unmarshal(respBody, &ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func parseQuery200(respBody []byte) (*ContractQuerySuccessResult, error) {
	var ret ContractQuerySuccessResult
	err := json.Unmarshal(respBody, &ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func parseQuery500(respBody []byte) (*ContractQueryErrorResult, error) {
	var ret ContractQueryErrorResult
	err := json.Unmarshal(respBody, &ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func unwrapOk(returnedOutput string) (string, error) {
	// The result is usually "Ok(_)"
	if strings.HasPrefix(returnedOutput, "Ok(") && strings.HasSuffix(returnedOutput, ")") {
		return returnedOutput[3 : len(returnedOutput)-1], nil
	}

	return "", fmt.Errorf("合约返回值不是 Ok(_) 形式")
}
