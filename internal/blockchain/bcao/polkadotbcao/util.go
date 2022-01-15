package polkadotbcao

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/chaincodectx"
	"github.com/pkg/errors"
)

func fillPostFormBase(ctx *chaincodectx.PolkadotChaincodeCtx, form *url.Values) {
	form.Set("abi", ctx.ContractABI)
	form.Set("contractAddress", ctx.ContractAddress)
	form.Set("signerAddress", ctx.CallerAddress)
}

func fillGetQueryStringBase(ctx *chaincodectx.PolkadotChaincodeCtx, queryString *url.Values) {
	queryString.Set("abi", ctx.ContractABI)
	queryString.Set("contractAddress", ctx.ContractAddress)
	queryString.Set("callerAddress", ctx.CallerAddress)
}

func sendTx(ctx *chaincodectx.PolkadotChaincodeCtx, client *http.Client, funcName string, funcArgs []interface{}, queryIfNoEvent bool) (*ContractTxSuccessResult, error) {
	// Prepare a POST form
	endpoint := ctx.APIPrefix + "/contract/tx"
	form := url.Values{}
	fillPostFormBase(ctx, &form)
	form.Set("funcName", funcName)

	argsBytes, err := json.Marshal(funcArgs)
	if err != nil {
		return nil, errors.Wrap(err, "无法序列化合约参数")
	}
	form.Set("funcArgs", string(argsBytes))

	formEncoded := form.Encode()
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(formEncoded))
	if err != nil {
		return nil, errors.Wrap(err, "无法调用合约")
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(formEncoded)))

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
		querySuccessResult, err := sendQuery(ctx, client, funcName, funcArgs, true)
		if err != nil {
			return nil, errors.Wrapf(err, "无法获取合约错误消息")
		}

		// The error message is inside the output. The output is usually in the form of
		// {
		//   "err": "error message here"
		// }
		errMsg, err := unwrapErr(querySuccessResult.Output)
		if err != nil {
			return nil, errors.Wrapf(err, "无法获取合约错误消息")
		}

		return txSuccessResult, fmt.Errorf(errMsg)
	} else if resp.StatusCode == 500 {
		txErrorResult, err := parseTx500(respBodyBytes)
		if err != nil {
			return nil, errors.Wrapf(err, "无法解析合约执行结果")
		}

		if txErrorResult.ExplainedModuleError == nil {
			return nil, fmt.Errorf("合约执行错误: %v", string(respBodyBytes))
		}

		return nil, fmt.Errorf("合约执行错误: %v", txErrorResult.ExplainedModuleError)
	} else {
		return nil, fmt.Errorf("合约执行错误: %v", string(respBodyBytes))
	}
}

func sendQuery(ctx *chaincodectx.PolkadotChaincodeCtx, client *http.Client, funcName string, funcArgs []interface{}, interpretError200AsOk bool) (*ContractQuerySuccessResult, error) {
	//// Prepare a GET query string
	//endpoint := ctx.APIPrefix + "/contract/query"
	//req, err := http.NewRequest("GET", endpoint, nil)
	//if err != nil {
	//	return nil, errors.Wrap(err, "无法调用合约")
	//}

	//queryString := req.URL.Query()
	//fillGetQueryStringBase(ctx, &queryString)
	//queryString.Set("funcName", funcName)

	//argsBytes, err := json.Marshal(funcArgs)
	//if err != nil {
	//	return nil, errors.Wrap(err, "无法序列化合约参数")
	//}
	//queryString.Set("funcArgs", string(argsBytes))

	//req.URL.RawQuery = queryString.Encode()

	//// Perform a GET request

	// <-- Above is preserved as an example of Golang's GET request -->

	// Prepare a POST form
	endpoint := ctx.APIPrefix + "/contract/query"
	form := url.Values{}
	fillGetQueryStringBase(ctx, &form)
	form.Set("funcName", funcName)

	argsBytes, err := json.Marshal(funcArgs)
	if err != nil {
		return nil, errors.Wrap(err, "无法序列化合约参数")
	}
	form.Set("funcArgs", string(argsBytes))

	formEncoded := form.Encode()
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(formEncoded))
	if err != nil {
		return nil, errors.Wrap(err, "无法调用合约")
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(formEncoded)))

	// Perform a POST request
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

		if interpretError200AsOk {
			return querySuccessResult, nil
		} else {
			// Status code of 200 may still imply an error. We should check into it.
			errMsg, err := unwrapErr(querySuccessResult.Output)
			if err != nil {
				// Not in the form of Err(_) -> Either a normal result or an Ok(_) result
				// Thus it should be treated as a success query
				return querySuccessResult, nil
			}

			return nil, fmt.Errorf(errMsg)
		}
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

// Returns JSON bytes
func unwrapOk(returnedOutput interface{}) ([]byte, error) {
	if reflect.TypeOf(returnedOutput).Kind() != reflect.Map {
		return nil, fmt.Errorf("合约返回值不是 Ok(_) 形式")
	}

	// Expect the result is in the form of
	// {
	//   "ok": JSONObj
	// }
	ret, ok := returnedOutput.(map[string]interface{})["ok"]
	if !ok {
		return nil, fmt.Errorf("合约返回值不是 Ok(_) 形式")
	}

	retJsonBytes, err := json.Marshal(ret)
	if err != nil {
		// Should not happen
		return nil, errors.Wrap(err, "无法序列化合约返回值")
	}

	return retJsonBytes, nil
}

// Returns error message as string
func unwrapErr(returnedOutput interface{}) (string, error) {
	if reflect.TypeOf(returnedOutput).Kind() != reflect.Map {
		return "", fmt.Errorf("合约返回值不是 Err(_) 形式")
	}

	// Expect the result is in the form of
	// {
	//   "err": error message (only strings are allowed)
	// }
	ret, ok := returnedOutput.(map[string]interface{})["err"]
	if !ok {
		return "", fmt.Errorf("合约返回值不是 Err(_) 形式")
	}

	if reflect.TypeOf(ret).Kind() != reflect.String {
		return "", fmt.Errorf("合约返回值 Err(_) 内容不是字符串")
	}

	return ret.(string), nil
}
