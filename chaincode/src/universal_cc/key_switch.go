package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/errorcode"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/auth"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/keyswitch"
	"github.com/casbin/casbin"
	"github.com/casbin/casbin/model"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

func (uc *UniversalCC) createKeySwitchTrigger(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// 检查参数个数
	if len(args) != 2 {
		return shim.Error("参数数量不正确。应为 2 个")
	}

	// 第0个参数解析为 keyswitch.KeySwitchTrigger
	var ksTrigger keyswitch.KeySwitchTrigger
	err := json.Unmarshal([]byte(args[0]), &ksTrigger)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法解析参数中的 JSON 对象: %v", err))
	}

	// 第1个参数解析为 eventID
	eventID := args[1]
	if eventID == "" {
		return shim.Error("事件 ID 不能为空")
	}

	// 获取ksSessionID
	ksSessionID := stub.GetTxID()

	// 获取 authSessionID
	validationResult := false
	authSessionID := ksTrigger.AuthSessionID

	// 获取 resourceID，验证资源是否存在
	resourceID := ksTrigger.ResourceID
	key := getKeyForResMetadata(resourceID)
	metadata, err := stub.GetState(key)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法确定元数据的可用性: %v", err))
	}
	if metadata == nil {
		return shim.Error("资源 ID 不存在")
	}

	// 获取创建者与时间戳
	creator, err := getPKDERFromStub(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法获取创建者: %v", err))
	}
	creatorAsBase64 := base64.StdEncoding.EncodeToString(creator)

	timestamp, err := getTimeFromStub(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法获得时间戳: %v", err))
	}

	// 检查用于密钥置换的公钥是否为空
	if ksTrigger.KeySwitchPK == "" {
		return shim.Error("用于密钥置换的公钥未指定")
	}

	if authSessionID != "" {
		// 获取 AuthRequestStored，验证其中资源 ID 是否相同。若请求不存在，则另外报错。
		authReq := uc.getAuthRequest(stub, []string{authSessionID})
		if authReq.Payload == nil {
			return shim.Error("该授权会话不存在")
		}
		var authRequestStored auth.AuthRequestStored
		err = json.Unmarshal(authReq.Payload, &authRequestStored)
		if err != nil {
			return shim.Error(fmt.Sprintf("AuthRequestStored 无法解析成 JSON 对象: %v", err))
		}
		if authRequestStored.ResourceID != resourceID {
			return shim.Error("资源 ID 与授权会话 ID 不匹配")
		}

		// 验证 AuthRequestStored.Creator 是否等于链码调用者 Creator
		if authRequestStored.Creator != creatorAsBase64 {
			return shim.Error("不是申请授权者本人")
		}

		// 如果 authSessionID 不为空值，则获取 AuthResponseStored，并解析成 JSON 对象。若批复不存在，则另外报错。
		authResp, err := uc.getAuthResponseHelper(stub, authSessionID)
		if err != nil {
			if err == errorcode.ErrorNotFound {
				return shim.Error("该授权会话申请未得到批复")
			} else {
				return shim.Error(err.Error())
			}
		}

		var authResponseStored auth.AuthResponseStored
		err = json.Unmarshal(authResp, &authResponseStored)
		if err != nil {
			return shim.Error(fmt.Sprintf("AuthResponseStored 无法解析成 JSON 对象: %v", err))
		}

		// 根据 AuthResponseStored 中的结果得到最终判断结果
		if authResponseStored.Result == true {
			validationResult = true
		} else {
			return shim.Error(errorcode.CodeForbidden)
		}
	} else {
		// 如果 authSessionID 为空值，执行 abac
		// 从当前客户端的证书上获取部门信息
		deptIdentity, err := uc.getDepartmentIdentityHelper(stub)
		if err != nil {
			return shim.Error(err.Error())
		}

		// 根据资源 ID，得到资源的访问策略
		resourceID := ksTrigger.ResourceID
		key := getKeyForResPolicy(resourceID)
		Policy, err := stub.GetState(key)
		if err != nil {
			return shim.Error(fmt.Sprintf("无法确定 policy 的可用性: %v", err))
		}
		if Policy == nil {
			return shim.Error("该资源的访问策略不存在")
		}

		// TODO:完善访问策略
		s := string(Policy)
		parts := strings.Split(s, "|| ")
		s = strings.Join(parts, "|| r.sub.")
		parts = strings.Split(s, "(")
		s = strings.Join(parts, "(r.sub.")
		parts = strings.Split(s, "&& ")
		s = strings.Join(parts, "&& r.sub.")
		policy := "m = " + s

		// 执行abac，并得到最终判断结果
		modeltext := `
		[request_definition]
		r = sub, obj
	
		[policy_definition]
		p = act
	
		[policy_effect]
		e = some(where (p.eft == allow))
	
		[matchers]
		`
		modeltext = modeltext + policy
		m := model.Model{}
		m.LoadModelFromText(modeltext)
		e := casbin.NewEnforcer(m)
		validationResult = e.Enforce(deptIdentity, "", "")
		if validationResult == false {
			return shim.Error(errorcode.CodeForbidden)
		}
	}

	// 构建 KeySwitchTriggerStored 并存储上链
	ksTriggerToBeStored := keyswitch.KeySwitchTriggerStored{
		KeySwitchSessionID: ksSessionID,
		ResourceID:         ksTrigger.ResourceID,
		AuthSessionID:      authSessionID,
		Creator:            creatorAsBase64,
		KeySwitchPK:        ksTrigger.KeySwitchPK,
		Timestamp:          timestamp,
		ValidationResult:   validationResult,
	}
	data, err := json.Marshal(ksTriggerToBeStored)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法序列化 KeySwitchTriggerStored: %v", err))
	}
	key = getKeyForKeySwitchTrigger(ksSessionID)
	err = stub.PutState(key, data)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法存储 KeySwitchTriggerStored: %v", err))
	}

	// 发事件
	err = stub.SetEvent(eventID, data)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法生成事件 '%v': %v", eventID, err))
	}

	return shim.Success([]byte(ksSessionID))
}

func (uc *UniversalCC) createKeySwitchResult(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// 检查参数个数
	if len(args) != 1 {
		return shim.Error("参数数量不正确。应为 1 个")
	}

	// 解析第0个参数为 keyswitch.KeySwitchResult
	var ksResult keyswitch.KeySwitchResult
	err := json.Unmarshal([]byte(args[0]), &ksResult)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法解析参数中的 JSON 对象: %v", err))
	}

	// 获取 ksSessionID
	ksSessionID := ksResult.KeySwitchSessionID

	// 检查 KeySwitchTriggerStored 是否存在
	ksTriggerStored, err := uc.getKeySwitchTriggerHelper(stub, ksSessionID)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法确定 KeySwitchTriggerStored 的可用性: %v", err))
	}
	if ksTriggerStored == nil {
		return shim.Error("该 KeySwitchTriggerStored 不存在")
	}

	// 获取创建者与时间戳
	creator, err := getPKDERFromStub(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法获取创建者: %v", err))
	}
	creatorAsBase64 := base64.StdEncoding.EncodeToString(creator)

	timestamp, err := getTimeFromStub(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法获得时间戳: %v", err))
	}

	// 构建 KeySwitchResultStored 并存储上链
	ksResultStored := keyswitch.KeySwitchResultStored{
		KeySwitchSessionID: ksSessionID,
		Share:              ksResult.Share,
		ZKProof:            ksResult.ZKProof,
		KeySwitchPK:        ksResult.KeySwitchPK,
		Creator:            creatorAsBase64,
		Timestamp:          timestamp,
	}
	data, err := json.Marshal(ksResultStored)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法序列化 KeySwitchResultStored: %v", err))
	}
	key := getKeyForKeySwitchResponse(ksSessionID, creatorAsBase64)
	err = stub.PutState(key, data)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法存储 KeySwitchResultStored: %v", err))
	}

	// 发事件
	eventID := getKeyPrefixForKeySwitchResponse(ksSessionID)
	value := getKeyForKeySwitchResponse(ksSessionID, creatorAsBase64)
	err = stub.SetEvent(eventID, []byte(value))
	if err != nil {
		return shim.Error(fmt.Sprintf("无法生成事件 '%v': %v", eventID, err))
	}

	// 获取交易ID
	txID := stub.GetTxID()

	return shim.Success([]byte(txID))
}

func (uc *UniversalCC) getKeySwitchTriggerHelper(stub shim.ChaincodeStubInterface, keySwitchSessionID string) (*keyswitch.KeySwitchTriggerStored, error) {
	// 获取 ksTriggerStored
	key := getKeyForKeySwitchTrigger(keySwitchSessionID)
	ksTriggerStored, err := stub.GetState(key)
	if err != nil {
		return nil, err
	}
	if ksTriggerStored == nil {
		return nil, err
	}

	var keySwitchTriggerStored keyswitch.KeySwitchTriggerStored
	err = json.Unmarshal(ksTriggerStored, &keySwitchTriggerStored)
	if err != nil {
		return nil, err
	}

	return &keySwitchTriggerStored, nil
}

func (uc *UniversalCC) getKeySwitchResult(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// 检查参数个数
	if len(args) != 1 {
		return shim.Error("参数数量不正确。应为 1 个")
	}

	// 解析第0个参数为 keyswitch.KeySwitchResultQuery
	var query keyswitch.KeySwitchResultQuery
	err := json.Unmarshal([]byte(args[0]), &query)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法解析参数中的 JSON 对象: %v", err))
	}

	// 获取 ksSessionID and resultCreator
	ksSessionID := query.KeySwitchSessionID
	resultCreator := query.ResultCreator

	// 获取 KeySwitchResultStore
	key := getKeyForKeySwitchResponse(ksSessionID, resultCreator)
	ksResultStored, err := stub.GetState(key)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法确定 KeySwitchResultStore 的可用性: %v", err))
	}
	if ksResultStored == nil {
		return shim.Error(errorcode.CodeNotFound)
	}

	return shim.Success(ksResultStored)
}

func (uc *UniversalCC) listKeySwitchResultsByID(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// 检查参数个数
	if len(args) != 1 {
		return shim.Error("参数数量不正确。应为 1 个")
	}

	// 获取 ksSessionID，确定搜索前缀
	ksSessionID := args[0]
	startKey := getKeyPrefixForKeySwitchResponse(ksSessionID) + "_"

	// 得到搜索截至key
	endKey := string(BytesPrefix([]byte(startKey)))

	// 开始搜索
	resultsIterator, err := stub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法查询密钥置换结果: %v", err))
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString(string(queryResponse.Value))
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")
	return shim.Success(buffer.Bytes())
}

// BytesPrefix get endKey
func BytesPrefix(prefix []byte) []byte {
	var limit []byte
	for i := len(prefix) - 1; i >= 0; i-- {
		c := prefix[i]
		if c < 0xff {
			limit = make([]byte, i+1)
			copy(limit, prefix)
			limit[i] = c + 1
			break
		}
	}
	return limit
}
