package main

import (
	"crypto/x509"
	"encoding/pem"
	"strings"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/google/uuid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-chaincode-go/shimtest"
	"github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric-protos-go/peer"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var testLogger = log.StandardLogger()

const exampleCertUser1 = `-----BEGIN CERTIFICATE-----
MIICKDCCAc6gAwIBAgIRAPstx377NKEjR+ohbQ2J0oUwCgYIKoZIzj0EAwIwcTEL
MAkGA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNhbiBG
cmFuY2lzY28xGDAWBgNVBAoTD29yZzEubGFiODA1LmNvbTEbMBkGA1UEAxMSY2Eu
b3JnMS5sYWI4MDUuY29tMB4XDTIwMTAyOTEyMDAwMFoXDTMwMTAyNzEyMDAwMFow
azELMAkGA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNh
biBGcmFuY2lzY28xDzANBgNVBAsTBmNsaWVudDEeMBwGA1UEAwwVVXNlcjFAb3Jn
MS5sYWI4MDUuY29tMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEHZYGK3Ck7LVg
u1YRK/vweROnZ6e1CSNzYviGXELedNZ/Rcv/1r/eEMP1hGhRjQdw1yz855N9I2FC
mSUdr1hgdKNNMEswDgYDVR0PAQH/BAQDAgeAMAwGA1UdEwEB/wQCMAAwKwYDVR0j
BCQwIoAgfPu1yXjxVzXCDv8yjNIBA6IhTAkvU5VROcg5ebdiB3cwCgYIKoZIzj0E
AwIDSAAwRQIhAO2h+8VLHnxBcbmPsc410N3dDCiSVx0b/2kSm53i801aAiAKZ/AG
mSmrm0zEPivFjOxDpd72v4tUS+O09sr28k+UPA==
-----END CERTIFICATE-----`

const exampleCertUser2 = `-----BEGIN CERTIFICATE-----
MIIC1TCCAnugAwIBAgIUYzfSJbtsxXJpUjVYLtuH3Aa+2UYwCgYIKoZIzj0EAwIw
czELMAkGA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNh
biBGcmFuY2lzY28xGTAXBgNVBAoTEG9yZzEuZXhhbXBsZS5jb20xHDAaBgNVBAMT
E2NhLm9yZzEuZXhhbXBsZS5jb20wHhcNMjEwMTE4MDgzMDAwWhcNMjIwMTE4MDgz
NTAwWjBdMQswCQYDVQQGEwJVUzEXMBUGA1UECBMOTm9ydGggQ2Fyb2xpbmExFDAS
BgNVBAoTC0h5cGVybGVkZ2VyMQ8wDQYDVQQLEwZjbGllbnQxDjAMBgNVBAMTBVVz
ZXIyMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEVze5OZbW+0dxmEirpAa8ZwOZ
0fHulXVr43e65Orlye7FMIwk6I5lFQv3NCsB55nGHN9aQcY8p2NDZkHimSAsnKOC
AQEwgf4wDgYDVR0PAQH/BAQDAgeAMAwGA1UdEwEB/wQCMAAwHQYDVR0OBBYEFCuq
p6CBmeiqw8fJOqmcNxOuqkX+MCsGA1UdIwQkMCKAIIX7/j1cvIs5NWgKG6PnenIm
LIm/atWWERrq16MMCQySMBQGA1UdEQQNMAuCCWxvY2FsaG9zdDB8BggqAwQFBgcI
AQRweyJhdHRycyI6eyJEZXBhcnRtZW50IjoiY29tcHV0ZXIiLCJHcmFkZSI6IjIi
LCJoZi5BZmZpbGlhdGlvbiI6IiIsImhmLkVucm9sbG1lbnRJRCI6IlVzZXIyIiwi
aGYuVHlwZSI6ImNsaWVudCJ9fTAKBggqhkjOPQQDAgNIADBFAiEAm3JbqyvB+aw+
0E5Mb3PsmiTFK/ZlXRvXDSXNW+YDkbQCIHff/qxbQvKQ5tfrMPuTMPzHh8mJkLD6
m8oUZXEh5bKi
-----END CERTIFICATE-----
`

const exampleCertUser3 = `-----BEGIN CERTIFICATE-----
MIIDHjCCAsWgAwIBAgIUGTUcF5Bj0nXRt/BtG0SslzZfVgMwCgYIKoZIzj0EAwIw
ajELMAkGA1UEBhMCVVMxFzAVBgNVBAgTDk5vcnRoIENhcm9saW5hMRQwEgYDVQQK
EwtIeXBlcmxlZGdlcjEPMA0GA1UECxMGRmFicmljMRswGQYDVQQDExJjYS5vcmcx
LmxhYjgwNS5jb20wHhcNMjEwMzA3MDk0NTAwWhcNMjIwMzA3MDk1NTAwWjBbMQsw
CQYDVQQGEwJVUzEXMBUGA1UECBMOTm9ydGggQ2Fyb2xpbmExFDASBgNVBAoTC0h5
cGVybGVkZ2VyMQ0wCwYDVQQLEwRwZWVyMQ4wDAYDVQQDEwVwZWVyMDBZMBMGByqG
SM49AgEGCCqGSM49AwEHA0IABJsfYuGbeTpHC0PUkbms0NWpEmhul89+nD+fjQ/i
HvGz4Qmicdz8Ydee0oyQbqim9nNrHeCa/Y3oBStZFrqqxuyjggFWMIIBUjAOBgNV
HQ8BAf8EBAMCB4AwDAYDVR0TAQH/BAIwADAdBgNVHQ4EFgQU2UAiJ0WoEKTT9bpe
0s2r/tCtyRQwHwYDVR0jBBgwFoAU6xcRmVzmh0HmJ111bISCbUJQ/xMwIAYDVR0R
BBkwF4IVcGVlcjAub3JnMS5sYWI4MDUuY29tMIHPBggqAwQFBgcIAQSBwnsiYXR0
cnMiOnsiRGVwdExldmVsIjoiMiIsIkRlcHROYW1lIjoiODEyIiwiRGVwdFR5cGUi
OiJjb21wdXRlciIsIlN1cGVyRGVwdE5hbWUiOiI4MDQiLCJhZG1pbiI6InRydWUi
LCJoZi5BZmZpbGlhdGlvbiI6IiIsImhmLkVucm9sbG1lbnRJRCI6InBlZXIwIiwi
aGYuUmVnaXN0cmFyLlJvbGVzIjoicGVlciIsImhmLlR5cGUiOiJwZWVyIn19MAoG
CCqGSM49BAMCA0cAMEQCIC4PUwJHhxi20JJT+yAdB+i4UWNcPmIFNNFHyHYwgvCm
AiBQxf8/6m576DKRpTB+x1BAOhnk2MoNdm9Qrv4OC5Oykw==
-----END CERTIFICATE-----`

// Check if the actual value is equal to the expected value.
func expectEqual(t *testing.T, expected interface{}, actual interface{}) {
	isEqual := assert.Equal(t, expected, actual)
	if !isEqual {
		testLogger.Infof("Value was '%v'. Expecting '%v'\n", actual, expected)
		t.FailNow()
	}
}

// Check if the actual value is NOT equal to the expected value.
func expectNotEqual(t *testing.T, expected interface{}, actual interface{}) {
	isEqual := assert.Equal(t, expected, actual)
	if isEqual {
		testLogger.Infof("Value was '%v'. Expecting to be different\n", actual)
		t.FailNow()
	}
}

// Check if the actual value is nil.
func expectNil(t *testing.T, actual interface{}) {
	isNil := assert.Nil(t, actual)
	if !isNil {
		testLogger.Infof("Value was '%v'. Expecting to be nil\n", actual)
		t.FailNow()
	}
}

// Check if the actual value is NOT nil.
func expectNotNil(t *testing.T, actual interface{}) {
	isNil := assert.Nil(t, actual)
	if isNil {
		testLogger.Infof("Value was nil. Expecting not to be nil\n")
		t.FailNow()
	}
}

// Check if the string ends with the specified phrases.
func expectStringEndsWith(t *testing.T, expectedSuffix string, actual string) {
	isCorrectEnding := strings.HasSuffix(actual, expectedSuffix)
	if !isCorrectEnding {
		testLogger.Infof("Value was '%v'. Expecting to end with '%v'\n", actual, expectedSuffix)
		t.FailNow()
	}
}

// Check if the state of the key is equal to the value specified.
func expectStateEqual(t *testing.T, stub *shimtest.MockStub, key string, expected []byte) {
	bytes := stub.State[key]

	isNotNil := assert.NotNil(t, bytes)
	if !isNotNil {
		testLogger.Infof("Failed to get value for key '%v'\n", key)
		t.FailNow()
	}

	isEqual := assert.Equal(t, expected, bytes)
	if !isEqual {
		testLogger.Infof("State value of key '%v' was '%v'. Expecting '%v'\n", key, bytes, expected)
		t.FailNow()
	}
}

// Check if the state of the key is NOT equal to the value specified.
func expectStateNotEqual(t *testing.T, stub *shimtest.MockStub, key string, expected []byte) {
	bytes := stub.State[key]

	isNotNil := assert.NotNil(t, bytes)
	if !isNotNil {
		testLogger.Infof("Failed to get value for key '%v'\n", key)
		t.FailNow()
	}

	isEqual := assert.Equal(t, expected, bytes)
	if isEqual {
		testLogger.Infof("State value of key '%v' was '%v'. Expecting to be different\n", key, bytes)
		t.FailNow()
	}
}

// Check if the state of the key is empty as specified.
func expectStateEmpty(t *testing.T, stub *shimtest.MockStub, key string) {
	bytes := stub.State[key]

	isNotNil := assert.NotNil(t, bytes)
	if !isNotNil {
		testLogger.Infof("Failed to get value for key '%v'\n", key)
		t.FailNow()
	}

	isEqual := assert.Equal(t, 0, len(bytes))
	if !isEqual {
		testLogger.Infof("State value of key '%v' was '%v'. Expecting empty value\n", key, bytes)
		t.FailNow()
	}
}

// Check if the response state is OK.
func expectResponseStatusOK(t *testing.T, resp *peer.Response) {
	if resp.Status != shim.OK {
		testLogger.Infof("Response status was ERROR with message '%v'. Expecting response status to be OK\n", resp.Message)
		t.FailNow()
	}
}

// Check if the response state is ERROR.
func expectResponseStatusERROR(t *testing.T, resp *peer.Response) {
	if resp.Status != shim.ERROR {
		testLogger.Infof("Expecting response status to be ERROR\n")
		t.FailNow()
	}
}

// Creates a MockStub bound to the chaincode struct UniversalCC.
// Certificate is default to `exampleCertUser1`
func createMockStub(t *testing.T, stubName string) *shimtest.MockStub {
	return createMockStubWithCert(t, stubName, exampleCertUser1)
}

// Creates a MockStub bound to the chaincode struct UniversalCC with the certificate specified.
func createMockStubWithCert(t *testing.T, stubName string, certPEM string) *shimtest.MockStub {
	uc := new(UniversalCC)
	mockStub := shimtest.NewMockStub(stubName, uc)
	setMockStubCreator(t, mockStub, "Org1MSP", []byte(certPEM))
	return mockStub
}

func setMockStubCreator(t *testing.T, stub *shimtest.MockStub, mspID string, idBytes []byte) {
	sid := &msp.SerializedIdentity{Mspid: mspID, IdBytes: idBytes}
	b, err := proto.Marshal(sid)
	if err != nil {
		testLogger.Infof("Cannot set stub creator: %v\n", err)
		t.FailNow()
	}

	stub.Creator = b
}

func getPKDERFromCertString(certPEM string) ([]byte, error) {
	blockBytes, _ := pem.Decode([]byte(certPEM))
	cert, err := x509.ParseCertificate(blockBytes.Bytes)
	if err != nil {
		return nil, err
	}

	ret, err := x509.MarshalPKIXPublicKey(cert.PublicKey)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

// Initializes the chaincode with the specified parameters using mockStub.MockInit.
func initChaincode(mockStub *shimtest.MockStub, arguments [][]byte) peer.Response {
	resp := mockStub.MockInit(uuid.NewString(), arguments)
	return resp
}
