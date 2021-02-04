package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

func (uc *UniversalCC) createAuthRequest(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	return shim.Error(codeNotImplemented)
}

func (uc *UniversalCC) createAuthResponse(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	return shim.Error(codeNotImplemented)
}

func (uc *UniversalCC) getAuthRequest(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	return shim.Error(codeNotImplemented)
}

func (uc *UniversalCC) listPendingAuthSessionIDsByResourceCreator(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// 检查参数数量
	lenArgs := len(args)
	if lenArgs != 2 {
		return shim.Error("参数数量不正确。应为 2 个")
	}

	// args = [pageSize int32, bookmark string]
	pageSizeStr := args[0]
	bookmark := args[1]

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法解析参数 pageSize，值为 %v。应为正整数", pageSizeStr))
	}
	if pageSize <= 0 {
		return shim.Error(fmt.Sprintf("参数 pageSize 值为 %v。应为正整数", pageSizeStr))
	}

	// 获取调用者信息
	creator, err := getPKDERFromStub(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法获取调用者信息: %v", err))
	}

	// 提供 creator 项以获取迭代器
	it, _, err := stub.GetStateByPartialCompositeKeyWithPagination("resourcecreator~authsessionid", []string{string(creator)}, int32(pageSize), bookmark)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法查询索引: %v", err))
	}

	// 遍历迭代器，解出 authsessionid 项，组为列表
	authSessionIDs := []string{}
	for it.HasNext() {
		entry, err := it.Next()
		if err != nil {
			return shim.Error(fmt.Sprintf("无法查询索引: %v", err))
		}

		_, ckParts, err := stub.SplitCompositeKey(entry.Key)
		if err != nil {
			return shim.Error(fmt.Sprintf("无法查询索引: %v", err))
		}

		authSessionIDs = append(authSessionIDs, ckParts[1])
	}

	// 序列化结果
	authSessionIDsAsBytes, err := json.Marshal(authSessionIDs)
	if err != nil {
		return shim.Error(fmt.Sprintf("无法序列化结果列表: %v", err))
	}

	return shim.Success(authSessionIDsAsBytes)
}
