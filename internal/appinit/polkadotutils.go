package appinit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strconv"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/bcao/polkadotbcao"
	errors "github.com/pkg/errors"
)

func sendTxToInstantiateChaincode(apiPrefix string, client *http.Client, abi string, wasmBytes []byte, signerAddress string, ctorFuncName string, ctorArgs string) (*polkadotbcao.ContractInstantiationSuccessResult, error) {
	// Prepare a POST form
	endpoint := apiPrefix + "/contract/from-code"

	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	bodyWriter.WriteField("abi", abi)
	bodyWriter.WriteField("signerAddress", signerAddress)
	bodyWriter.WriteField("ctorFuncName", ctorFuncName)
	bodyWriter.WriteField("ctorArgs", ctorArgs)

	fileWriter, err := bodyWriter.CreateFormFile("wasm", "contract.wasm")
	if err != nil {
		return nil, errors.Wrap(err, "无法为 form-data 创建文件部分")
	}
	if _, err := fileWriter.Write(wasmBytes); err != nil {
		return nil, errors.Wrap(err, "无法在请求中写入合约 WASM 字节")
	}

	contentType := bodyWriter.FormDataContentType()

	if err := bodyWriter.Close(); err != nil {
		return nil, errors.Wrap(err, "无法准备请求参数")
	}

	req, err := http.NewRequest("POST", endpoint, bodyBuf)
	if err != nil {
		return nil, errors.Wrap(err, "无法实例化合约")
	}

	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Content-Length", strconv.Itoa(bodyBuf.Len()))

	// Perform a POST request
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "无法实例化合约")
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

		return txSuccessResult, nil
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

func parseTx200(respBody []byte) (*polkadotbcao.ContractInstantiationSuccessResult, error) {
	var ret polkadotbcao.ContractInstantiationSuccessResult
	err := json.Unmarshal(respBody, &ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func parseTx500(respBody []byte) (*polkadotbcao.ContractInstantiationErrorResult, error) {
	var ret polkadotbcao.ContractInstantiationErrorResult
	err := json.Unmarshal(respBody, &ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}
