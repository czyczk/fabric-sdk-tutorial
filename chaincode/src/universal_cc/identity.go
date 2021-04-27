package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/identity"
	"github.com/hyperledger/fabric-ca/lib/attrmgr"
	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/mitchellh/mapstructure"
)

func (uc *UniversalCC) getDepartmentIdentity(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// 检查参数数量
	if len(args) != 0 {
		return shim.Error("参数数量不正确。应为 0 个")
	}

	// 从证书属性中获取部门身份信息
	identity, err := uc.getDepartmentIdentityHelper(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	identityBytes, err := json.Marshal(identity)
	if err != nil {
		return shim.Error("无法序列化身份信息")
	}

	return shim.Success(identityBytes)
}

func (uc *UniversalCC) getDepartmentIdentityHelper(stub shim.ChaincodeStubInterface) (*identity.DepartmentIdentityStored, error) {
	cert, err := cid.GetX509Certificate(stub)
	if err != nil {
		return nil, fmt.Errorf("无法获取证书: %v", err)
	}

	// 获取当前客户端证书上的属性
	attrs, err := attrmgr.New().GetAttributesFromCert(cert)
	if err != nil {
		return nil, fmt.Errorf("无法获取属性: %v", err)
	}

	// 将从证书中得到的属性 map 转为 struct
	var identity identity.DepartmentIdentityStored
	_ = mapstructure.Decode(attrs.Attrs, &identity)

	deptLevelStr := attrs.Attrs["DeptLevel"]
	if deptLevelStr != "" {
		deptLevel, err := strconv.Atoi(deptLevelStr)
		if err != nil {
			return nil, fmt.Errorf("不合法的证书，DeptLevel 应为整数: %v", err)
		}

		identity.DeptLevel = deptLevel
	}

	return &identity, nil
}
