package main

import (
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

// ScrewInventory implements interface Chaincode.
type ScrewInventory struct {
}

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

	// Write the state to the ledger
	err = stub.PutState(corp1Name, []byte(strconv.Itoa(corp1Amnt)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(corp2Name, []byte(strconv.Itoa(corp2Amnt)))
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

// This function will receive parameters as `peer chaincode invoke ... -c '{"Args":["transfer","CorpA","CorpB","10"]}'`
func (si *ScrewInventory) transfer(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// Check if the parameters are valid
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	invalidAmntErrorStr := "Expecting an integer value >= 0 for asset holding"

	sourceCorpName := args[0]
	destCorpName := args[1]
	transferAmnt, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error(invalidAmntErrorStr)
	}

	// Check if the asset source has enough amount to transfer
	sourceRemainingAmntBytes, err := stub.GetState(sourceCorpName)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed fetching the asset status of \"%v\"", sourceCorpName))
	}

	sourceRemainingAmnt, _ := strconv.Atoi(string(sourceRemainingAmntBytes))
	if sourceRemainingAmnt < transferAmnt {
		return shim.Error(fmt.Sprintf("Failed to transfer: \"%v\" does not have the enough amount of assets", sourceCorpName))
	}

	// Perform the transfer and update the ledger
	destRemainingAmntBytes, err := stub.GetState(destCorpName)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed fetching the asset status of \"%v\"", destCorpName))
	}

	sourceRemainingAmnt -= transferAmnt
	destRemainingAmnt, _ := strconv.Atoi(string(destRemainingAmntBytes))
	destRemainingAmnt += transferAmnt

	err = stub.PutState(sourceCorpName, []byte(strconv.Itoa(sourceRemainingAmnt)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(destCorpName, []byte(strconv.Itoa(destRemainingAmnt)))
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
	amntBytes, err := stub.GetState(targetCorpName)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(amntBytes)
}

func main() {
	err := shim.Start(new(ScrewInventory))
	if err != nil {
		fmt.Printf("failed to start ScrewInventory: %s", err)
	}
}
