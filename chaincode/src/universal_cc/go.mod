module universalcc

go 1.15

replace gitee.com/czyczk/fabric-sdk-tutorial => ../../../

replace github.com/XiaoYao-austin/ppks => ../../../ppks/

require (
	gitee.com/czyczk/fabric-sdk-tutorial v0.0.0
	github.com/casbin/casbin v1.9.1
	github.com/gogo/protobuf v1.3.2
	github.com/google/uuid v1.2.0
	github.com/hyperledger/fabric-ca v1.4.9
	github.com/hyperledger/fabric-chaincode-go v0.0.0-20201119163726-f8ef75b17719
	github.com/hyperledger/fabric-protos-go v0.0.0-20210127161553-4f432a78f286
	github.com/mitchellh/mapstructure v1.4.1
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
)
