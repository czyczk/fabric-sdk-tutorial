package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

type corporation struct {
	ObjectType string `json:"docType"`
	Name       string `json:"name"`
	ScrewAmnt  int    `json:"screwAmnt"`
}

// newCorporation creates an `organization` object with the info specified.
func newCorporation(name string, screwAmnt int) corporation {
	return corporation{
		ObjectType: "corporation",
		Name:       name,
		ScrewAmnt:  screwAmnt,
	}
}

type accessRecord struct {
	ObjectType  string `json:"docType"`
	ClientID    string `json:"clientID"`
	ResourceKey string `json:"resourceKey"`
}

// newAccessRecord creates an `accessRecord` object with the info specified.
func newAccessRecord(clientID string, resourceKey string) accessRecord {
	return accessRecord{
		ObjectType:  "accessRecord",
		ClientID:    clientID,
		ResourceKey: resourceKey,
	}
}

// ScrewInventory implements interface Chaincode.
type ScrewInventory struct{}

// Init specifies the initial amount of available screws in each company.
func (si *ScrewInventory) Init(stub shim.ChaincodeStubInterface) peer.Response {
	// Extract the parameters and check the number of them
	_, args := stub.GetFunctionAndParameters()

	if len(args) != 4 {
		return shim.Error(fmt.Sprintf("Incorrect number of arguments. Got %v. Expecting 4", len(args)))
	}

	// Check the validity of the parameters
	invalidValErrorStr := "Expecting an integer value >= 0 for asset holding"

	corp1Name := args[0]
	corp1Amnt, err := strconv.Atoi(args[1])
	if err != nil || corp1Amnt < 0 {
		return shim.Error(invalidValErrorStr)
	}

	corp2Name := args[2]
	corp2Amnt, err := strconv.Atoi(args[3])
	if err != nil || corp2Amnt < 0 {
		return shim.Error(invalidValErrorStr)
	}

	// Construct corporation objects
	corp1 := newCorporation(corp1Name, corp1Amnt)
	corp2 := newCorporation(corp2Name, corp2Amnt)

	corp1JsonBytes, err := json.Marshal(corp1)
	if err != nil {
		return shim.Error(err.Error())
	}

	corp2JsonBytes, err := json.Marshal(corp2)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Write the state to the ledger
	err = stub.PutState(corp1Name, corp1JsonBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(corp2Name, corp2JsonBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// Invoke handles query and transfer requests.
func (si *ScrewInventory) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	// Extract the function name and the parameters
	function, args := stub.GetFunctionAndParameters()

	// Act according to the function name
	if function == "transfer" {
		// Transfers the spceified amount of asset between the specified two corporations
		return si.transfer(stub, args)
	} else if function == "query" {
		// Queries the amount of asset of the specified corporation
		return si.query(stub, args)
	}

	return shim.Error("Invalid invoke function name. Expecting \"transfer\" or \"query\"")
}

// This function will receive parameters as `peer chaincode invoke ... -c '{"Args":["transfer","CorpA","CorpB","10","eventTransfer"]}'`
func (si *ScrewInventory) transfer(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// Check if the parameters are valid
	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	invalidAmntErrorStr := "Expecting an integer value >= 0 for asset holding"

	sourceCorpName := args[0]
	targetCorpName := args[1]
	transferAmnt, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error(invalidAmntErrorStr)
	}
	eventID := args[3]

	// Fetch the source and target corporations from the ledger
	failedFetchingAssetStatusErrorStr := "Failed fetching the asset status of \"%v\""

	sourceCorpBytes, err := stub.GetState(sourceCorpName)
	if err != nil {
		return shim.Error(fmt.Sprintf(failedFetchingAssetStatusErrorStr, sourceCorpName))
	}
	sourceCorp := corporation{}
	err = json.Unmarshal(sourceCorpBytes, &sourceCorp)
	if err != nil {
		return shim.Error(fmt.Sprintf(failedFetchingAssetStatusErrorStr, sourceCorpName))
	}

	targetCorpBytes, err := stub.GetState(targetCorpName)
	if err != nil {
		return shim.Error(fmt.Sprintf(failedFetchingAssetStatusErrorStr, targetCorpName))
	}
	targetCorp := corporation{}
	err = json.Unmarshal(targetCorpBytes, &targetCorp)
	if err != nil {
		return shim.Error(fmt.Sprintf(failedFetchingAssetStatusErrorStr, targetCorpName))
	}

	// Check if the asset source has enough amount to transfer
	sourceRemainingAmnt := sourceCorp.ScrewAmnt
	if sourceRemainingAmnt < transferAmnt {
		return shim.Error(fmt.Sprintf("Failed to transfer: \"%v\" does not have the enough amount of assets", sourceCorpName))
	}

	// Perform the transfer and update the ledger
	sourceCorp.ScrewAmnt -= transferAmnt
	targetCorp.ScrewAmnt += transferAmnt
	sourceCorpBytes, err = json.Marshal(&sourceCorp)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error JSON conversion: %v", err.Error()))
	}
	targetCorpBytes, err = json.Marshal(&targetCorp)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error JSON conversion: %v", err.Error()))
	}

	err = stub.PutState(sourceCorpName, sourceCorpBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(targetCorpName, targetCorpBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.SetEvent(eventID, []byte{})
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// This function will receive parameters as `peer chaincode query ... -c '{"Args":["query","CorpA"]}'`
func (si *ScrewInventory) query(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// Check if the parameters are valid
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	// Extract the query target
	targetCorpName := args[0]

	// Get the current state of the specified corporation
	corpBytes, err := stub.GetState(targetCorpName)
	if err != nil {
		return shim.Error(err.Error())
	}
	corp := corporation{}
	err = json.Unmarshal(corpBytes, &corp)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte(strconv.Itoa(corp.ScrewAmnt)))
}

func main() {
	err := shim.Start(new(ScrewInventory))
	if err != nil {
		fmt.Printf("failed to start ScrewInventory: %s", err)
	}
}
