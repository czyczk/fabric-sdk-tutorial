package main

import (
	"crypto/x509"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
)

const (
	// Res 对应“数据资源”的 key 的前缀
	Res = "res"
	// Auth 对应“访问权授权”申请和批复的 key 的前缀
	Auth = "auth"
	// KS 对应“Key-Switch”触发器与结果的 key 的前缀
	KS = "ks"
	// RegKey 对应监管者公钥的 key
	RegKey = "regKey"
)

func getKeyForResData(resourceID string) string {
	return fmt.Sprintf("res_%s", resourceID)
}

func getKeyForResMetadata(resourceID string) string {
	return fmt.Sprintf("res_%s_metadata", resourceID)
}

func getKeyForResKey(resourceID string) string {
	return fmt.Sprintf("res_%s_key", resourceID)
}

func getKeyForResPlicy(resourceID string) string {
	return fmt.Sprintf("res_%s_policy", resourceID)
}

func getKeyForAuthRequest(authSessionID string) string {
	return fmt.Sprintf("auth_%s_req", authSessionID)
}

func getKeyForAuthResponse(authSessionID string) string {
	return fmt.Sprintf("auth_%s_resp", authSessionID)
}

func getKeyForKeySwitchTrigger(keySwitchSessionID string) string {
	return fmt.Sprintf("ks_%s_trigger", keySwitchSessionID)
}

func getKeyForKeySwitchResponse(keySwitchSessionID string, resultCreator []byte) string {
	return fmt.Sprintf("ks_%s_result_%s", keySwitchSessionID, string(resultCreator))
}

func getTimeFromStub(stub shim.ChaincodeStubInterface) (ret time.Time, err error) {
	// 从 stub 中得到交易提案创建时间
	timestamp, err := stub.GetTxTimestamp()
	if err != nil {
		return
	}

	// 转为 Go 中的 time.Time
	ret = time.Unix(timestamp.GetSeconds(), int64(timestamp.GetNanos()))
	return
}

func getPKDERFromStub(stub shim.ChaincodeStubInterface) ([]byte, error) {
	// 从 stub 中抽取 X509 证书
	cert, err := cid.GetX509Certificate(stub)
	if err != nil {
		return nil, err
	}

	// 返回公钥的 DER 表示
	ret, err := x509.MarshalPKIXPublicKey(cert.PublicKey)
	if err != nil {
		return nil, err
	}

	return ret, nil
}
