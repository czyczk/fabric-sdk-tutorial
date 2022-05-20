package polkadoteventmgr

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/chaincodectx"
	"github.com/pkg/errors"
)

func registerPolkadotEvent(ctx *chaincodectx.PolkadotChaincodeCtx, contractAddress string, eventID string) (err error) {
	// Prepare a POST form
	var client *http.Client
	var resp *http.Response
	endpoint := ctx.APIPrefix + "/event/subscription"
	form := url.Values{}
	form.Set("contractAddress", ctx.ContractAddress)
	form.Set("eventId", eventID)

	formEncoded := form.Encode()
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(formEncoded))
	if err != nil {
		return errors.Wrap(err, "无法订阅 Polkadot 事件")
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(formEncoded)))

	// Perform a POST request
	resp, err = client.Do(req)
	if err != nil {
		return errors.Wrap(err, "无法订阅 Polkadot 事件")
	}
	defer resp.Body.Close()

	// Process the response
	// 201 -> return nil
	// 400, 500, Other -> response body as error message
	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrapf(err, "无法获取事件订阅结果")
	}

	if resp.StatusCode == 201 {
		return nil
	} else if resp.StatusCode == 400 {
		return fmt.Errorf("无法获取事件订阅结果 (400 Bad Request): %v", string(respBodyBytes))
	} else if resp.StatusCode == 500 {
		return fmt.Errorf("无法获取事件订阅结果 (500 Internal Server Error): %v", string(respBodyBytes))
	} else {
		return fmt.Errorf("无法获取事件订阅结果: %v", string(respBodyBytes))
	}
}

func registerPolkadotEvents(ctx *chaincodectx.PolkadotChaincodeCtx, reg *PolkadotEventRegistration) (polkadotEvents []PolkadotEvent, err error) {
	// Prepare a GET query string
	var client *http.Client
	var resp *http.Response
	endpoint := ctx.APIPrefix + "/event/subscription/" + reg.contractAddress + "/" + reg.eventID
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, errors.Wrap(err, "无法订阅 Polkadot 事件")
	}

	// Perform a GET request
	resp, err = client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "无法订阅 Polkadot 事件")
	}
	defer resp.Body.Close()

	// Process the response
	// 200 -> parseReg200
	// Other -> response body as error message
	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "无法获取事件订阅结果")
	}

	if resp.StatusCode == 200 {
		polkadotEvents, err := parseReg200(respBodyBytes)
		if err != nil {
			return nil, errors.Wrapf(err, "无法获取事件订阅结果 (200 OK)")
		}

		return polkadotEvents, nil
	} else {
		return nil, fmt.Errorf("无法获取事件订阅结果: %v", string(respBodyBytes))
	}
}

func parseReg200(respBody []byte) ([]PolkadotEvent, error) {
	var ret []PolkadotEvent
	err := json.Unmarshal(respBody, &ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}
