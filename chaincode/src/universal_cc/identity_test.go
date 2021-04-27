package main

import (
	"encoding/json"
	"testing"

	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/identity"
	"github.com/google/uuid"
)

func TestGetDepartmentIdentityWithAttributedCert(t *testing.T) {
	// 用带属性的证书初始化
	stub := createMockStubWithCert(t, "TestGetDepartmentIdentityWithAttributedCert", exampleCertUser3)
	_ = initChaincode(stub, [][]byte{})

	// 获取部门身份信息
	targetFunction := "getDepartmentIdentity"
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction)})
	expectResponseStatusOK(t, &resp)
	var deptIdentity identity.DepartmentIdentityStored
	err := json.Unmarshal(resp.Payload, &deptIdentity)
	expectNil(t, err)

	// 验证身份信息是否正确
	expectEqual(t, "812", deptIdentity.DeptName)
	expectEqual(t, 2, deptIdentity.DeptLevel)
	expectEqual(t, "computer", deptIdentity.DeptType)
	expectEqual(t, "804", deptIdentity.SuperDeptName)
}

func TestGetDepartmentIdentityWithExccesiveParameters(t *testing.T) {
	// 用带属性的证书初始化
	stub := createMockStubWithCert(t, "TestGetDepartmentIdentityWithAttributedCert", exampleCertUser3)
	_ = initChaincode(stub, [][]byte{})

	// 获取部门身份信息
	targetFunction := "getDepartmentIdentity"
	resp := stub.MockInvoke(uuid.NewString(), [][]byte{[]byte(targetFunction), []byte("EXCCESIVE_PARAMETER")})
	expectResponseStatusERROR(t, &resp)
}
