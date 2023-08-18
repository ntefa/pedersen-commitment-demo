package chaincode

import (
	"encoding/json"
	"fmt"
	"log"
	"pedersen-commitment-transfer/src/pedersen"
	"strconv"

	"github.com/bwesterb/go-ristretto"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Define key names for options
const nameKey = "name"
const symbolKey = "symbol"
const decimalsKey = "decimals"
const totalSupplyKey = "totalSupply"
const temporaryAccountAddress = "TempAccount"

// Define objectType names for prefix
const allowancePrefix = "allowance"

// Define key names for options

// SmartContract provides functions for transferring tokens between accounts
type SmartContract struct {
	contractapi.Contract
}

// event provides an organized struct for emitting events
type event struct {
	From  string          `json:"from"`
	To    string          `json:"to"`
	Value ristretto.Point `json:"value"`
}

// Mint creates new tokens and adds them to minter's account balance
// This function triggers a Transfer event
func (s *SmartContract) Mint(ctx contractapi.TransactionContextInterface, amount int64, committedAmount ristretto.Point) error {

	// Check if contract has been intilized first
	initialized, err := checkInitialized(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if contract is already initialized: %v", err)
	}
	if !initialized {
		return fmt.Errorf("Contract options need to be set before calling any function, call Initialize() to initialize contract")
	}

	//Check if the encryption is valid
	err = IsValidEncryption(ctx, amount, &committedAmount)
	if err != nil {
		return fmt.Errorf("Minting failed: %v", err)
	}

	_, _, zeroCommitted, err := GetPedersenParams(ctx)
	if err != nil {
		return fmt.Errorf("Minting failed: %v", err)
	}
	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get MSPID: %v", err)
	}
	if clientMSPID != "Org1MSP" {
		return fmt.Errorf("client is not authorized to mint new tokens")
	}

	// Get ID of submitting client identity
	minter, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("failed to get client id: %v", err)
	}

	if amount <= 0 {
		return fmt.Errorf("mint amount must be a positive integer")
	}

	currentBalanceBytes, err := ctx.GetStub().GetState(minter)
	if err != nil {
		return fmt.Errorf("failed to read minter account %s from world state: %v", minter, err)
	}

	var currentBalance ristretto.Point
	var updatedBalance ristretto.Point

	// If minter current balance doesn't yet exist, we'll create it with a current balance of 0
	if currentBalanceBytes == nil {
		currentBalance = *zeroCommitted
	} else {
		err = currentBalance.UnmarshalBinary(currentBalanceBytes) // Error handling not needed since Itoa() was used when setting the account balance, guaranteeing it was an integer.
	}

	if err != nil {
		return err
	}

	updatedBalance = pedersen.Add(&currentBalance, &committedAmount)

	updatedBalanceBytes, err := updatedBalance.MarshalBinary()
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(minter, updatedBalanceBytes)
	if err != nil {
		return err
	}

	// Update the totalSupply
	totalSupplyBytes, err := ctx.GetStub().GetState(totalSupplyKey)
	if err != nil {
		return fmt.Errorf("failed to retrieve total token supply: %v", err)
	}

	var totalSupply int

	// If no tokens have been minted, initialize the totalSupply
	if totalSupplyBytes == nil {
		totalSupply = 0
	} else {
		totalSupply, _ = strconv.Atoi(string(totalSupplyBytes)) // Error handling not needed since Itoa() was used when setting the totalSupply, guaranteeing it was an integer.
	}

	// Add the mint amount to the total supply and update the state
	totalSupply, err = add(totalSupply, int(amount)) //TODO: convert all to int64
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(totalSupplyKey, []byte(strconv.Itoa(totalSupply)))
	if err != nil {
		return err
	}

	// Emit the Transfer event
	transferEvent := event{"0x0", minter, committedAmount}
	transferEventJSON, err := json.Marshal(transferEvent)
	if err != nil {
		return fmt.Errorf("failed to obtain JSON encoding: %v", err)
	}
	err = ctx.GetStub().SetEvent("Transfer", transferEventJSON)
	if err != nil {
		return fmt.Errorf("failed to set event: %v", err)
	}

	log.Printf("minter account %s balance updated from %d to %d", minter, currentBalance, updatedBalance)

	return nil
}

// // Burn redeems tokens the minter's account balance
// // This function triggers a Transfer event
// func (s *SmartContract) Burn(ctx contractapi.TransactionContextInterface, amount int) error {

// 	// Check if contract has been intilized first
// 	initialized, err := checkInitialized(ctx)
// 	if err != nil {
// 		return fmt.Errorf("failed to check if contract is already initialized: %v", err)
// 	}
// 	if !initialized {
// 		return fmt.Errorf("Contract options need to be set before calling any function, call Initialize() to initialize contract")
// 	}
// 	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to burn new tokens
// 	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
// 	if err != nil {
// 		return fmt.Errorf("failed to get MSPID: %v", err)
// 	}
// 	if clientMSPID != "Org1MSP" {
// 		return fmt.Errorf("client is not authorized to mint new tokens")
// 	}

// 	// Get ID of submitting client identity
// 	minter, err := ctx.GetClientIdentity().GetID()
// 	if err != nil {
// 		return fmt.Errorf("failed to get client id: %v", err)
// 	}

// 	if amount <= 0 {
// 		return errors.New("burn amount must be a positive integer")
// 	}

// 	currentBalanceBytes, err := ctx.GetStub().GetState(minter)
// 	if err != nil {
// 		return fmt.Errorf("failed to read minter account %s from world state: %v", minter, err)
// 	}

// 	var currentBalance int

// 	// Check if minter current balance exists
// 	if currentBalanceBytes == nil {
// 		return errors.New("The balance does not exist")
// 	}

// 	currentBalance, _ = strconv.Atoi(string(currentBalanceBytes)) // Error handling not needed since Itoa() was used when setting the account balance, guaranteeing it was an integer.

// 	updatedBalance, err := sub(currentBalance, amount)
// 	if err != nil {
// 		return err
// 	}

// 	err = ctx.GetStub().PutState(minter, []byte(strconv.Itoa(updatedBalance)))
// 	if err != nil {
// 		return err
// 	}

// 	// Update the totalSupply
// 	totalSupplyBytes, err := ctx.GetStub().GetState(totalSupplyKey)
// 	if err != nil {
// 		return fmt.Errorf("failed to retrieve total token supply: %v", err)
// 	}

// 	// If no tokens have been minted, throw error
// 	if totalSupplyBytes == nil {
// 		return errors.New("totalSupply does not exist")
// 	}

// 	totalSupply, _ := strconv.Atoi(string(totalSupplyBytes)) // Error handling not needed since Itoa() was used when setting the totalSupply, guaranteeing it was an integer.

// 	// Subtract the burn amount to the total supply and update the state
// 	totalSupply, err = sub(totalSupply, amount)
// 	if err != nil {
// 		return err
// 	}

// 	err = ctx.GetStub().PutState(totalSupplyKey, []byte(strconv.Itoa(totalSupply)))
// 	if err != nil {
// 		return err
// 	}

// 	// Emit the Transfer event
// 	transferEvent := event{minter, "0x0", amount}
// 	transferEventJSON, err := json.Marshal(transferEvent)
// 	if err != nil {
// 		return fmt.Errorf("failed to obtain JSON encoding: %v", err)
// 	}
// 	err = ctx.GetStub().SetEvent("Transfer", transferEventJSON)
// 	if err != nil {
// 		return fmt.Errorf("failed to set event: %v", err)
// 	}

// 	log.Printf("minter account %s balance updated from %d to %d", minter, currentBalance, updatedBalance)

// 	return nil
// }

// Transfer transfers tokens from client account to recipient account
// recipient account must be a valid clientID as returned by the ClientID() function
// This function triggers a Transfer event
func (s *SmartContract) Transfer2Recipient(ctx contractapi.TransactionContextInterface, amount int64, committedAmount ristretto.Point, currentBalance int64, recipient string) error {

	if amount > currentBalance {
		return fmt.Errorf("You cannot send less than 0")
	}
	if amount > currentBalance {
		return fmt.Errorf("You cannot send more money that what you have available")
	}

	// Check if contract has been intilized first
	initialized, err := checkInitialized(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if contract is already initialized: %v", err)
	}
	if !initialized {
		return fmt.Errorf("Contract options need to be set before calling any function, call Initialize() to initialize contract")
	}

	//Check if the encryption is valid
	err = IsValidEncryption(ctx, amount, &committedAmount)
	if err != nil {
		return fmt.Errorf("Minting failed: %v", err)
	}

	// Get ID of submitting client identity
	clientID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("failed to get client id: %v", err)
	}

	err = transferHelper(ctx, clientID, recipient, committedAmount)
	if err != nil {
		return fmt.Errorf("failed to transfer: %v", err)
	}

	// Emit the Transfer event
	transferEvent := event{clientID, recipient, committedAmount}
	transferEventJSON, err := json.Marshal(transferEvent)
	if err != nil {
		return fmt.Errorf("failed to obtain JSON encoding: %v", err)
	}
	err = ctx.GetStub().SetEvent("Transfer", transferEventJSON)
	if err != nil {
		return fmt.Errorf("failed to set event: %v", err)
	}

	return nil
}

// Transfer transfers tokens from client account to recipient account
// recipient account must be a valid clientID as returned by the ClientID() function
// This function triggers a Transfer event
func (s *SmartContract) Transfer2TempAccount(ctx contractapi.TransactionContextInterface, amount int64, committedAmount ristretto.Point, currentBalance int64) error {

	if amount > currentBalance {
		return fmt.Errorf("You cannot send less than 0")
	}
	if amount > currentBalance {
		return fmt.Errorf("You cannot send more money that what you have available")
	}

	// Check if contract has been intilized first
	initialized, err := checkInitialized(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if contract is already initialized: %v", err)
	}
	if !initialized {
		return fmt.Errorf("Contract options need to be set before calling any function, call Initialize() to initialize contract")
	}

	//Check if the encryption is valid
	err = IsValidEncryption(ctx, amount, &committedAmount)
	if err != nil {
		return fmt.Errorf("Minting failed: %v", err)
	}

	// Get ID of submitting client identity
	clientID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("failed to get client id: %v", err)
	}

	err = transferHelper(ctx, clientID, temporaryAccountAddress, committedAmount)
	if err != nil {
		return fmt.Errorf("failed to transfer: %v", err)
	}

	// Emit the Transfer event
	transferEvent := event{clientID, temporaryAccountAddress, committedAmount}
	transferEventJSON, err := json.Marshal(transferEvent)
	if err != nil {
		return fmt.Errorf("failed to obtain JSON encoding: %v", err)
	}
	err = ctx.GetStub().SetEvent("Transfer", transferEventJSON)
	if err != nil {
		return fmt.Errorf("failed to set event: %v", err)
	}
	return nil
}

// // BalanceOf returns the balance of the given account
// func (s *SmartContract) BalanceOf(ctx contractapi.TransactionContextInterface, account string) (int, error) {

// 	// Check if contract has been intilized first
// 	initialized, err := checkInitialized(ctx)
// 	if err != nil {
// 		return 0, fmt.Errorf("failed to check if contract is already initialized: %v", err)
// 	}
// 	if !initialized {
// 		return 0, fmt.Errorf("Contract options need to be set before calling any function, call Initialize() to initialize contract")
// 	}

// 	balanceBytes, err := ctx.GetStub().GetState(account)
// 	if err != nil {
// 		return 0, fmt.Errorf("failed to read from world state: %v", err)
// 	}
// 	if balanceBytes == nil {
// 		return 0, fmt.Errorf("the account %s does not exist", account)
// 	}

// 	balance, _ := strconv.Atoi(string(balanceBytes)) // Error handling not needed since Itoa() was used when setting the account balance, guaranteeing it was an integer.

// 	return balance, nil
// }

// // ClientAccountBalance returns the balance of the requesting client's account
// func (s *SmartContract) ClientAccountBalance(ctx contractapi.TransactionContextInterface) (int, error) {

// 	// Check if contract has been intilized first
// 	initialized, err := checkInitialized(ctx)
// 	if err != nil {
// 		return 0, fmt.Errorf("failed to check if contract is already initialized: %v", err)
// 	}
// 	if !initialized {
// 		return 0, fmt.Errorf("Contract options need to be set before calling any function, call Initialize() to initialize contract")
// 	}

// 	// Get ID of submitting client identity
// 	clientID, err := ctx.GetClientIdentity().GetID()
// 	if err != nil {
// 		return 0, fmt.Errorf("failed to get client id: %v", err)
// 	}

// 	balanceBytes, err := ctx.GetStub().GetState(clientID)
// 	if err != nil {
// 		return 0, fmt.Errorf("failed to read from world state: %v", err)
// 	}
// 	if balanceBytes == nil {
// 		return 0, fmt.Errorf("the account %s does not exist", clientID)
// 	}

// 	balance, _ := strconv.Atoi(string(balanceBytes)) // Error handling not needed since Itoa() was used when setting the account balance, guaranteeing it was an integer.

// 	return balance, nil
// }

// // ClientAccountID returns the id of the requesting client's account
// // In this implementation, the client account ID is the clientId itself
// // Users can use this function to get their own account id, which they can then give to others as the payment address
// func (s *SmartContract) ClientAccountID(ctx contractapi.TransactionContextInterface) (string, error) {

// 	// Check if contract has been intilized first
// 	initialized, err := checkInitialized(ctx)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to check if contract is already initialized: %v", err)
// 	}
// 	if !initialized {
// 		return "", fmt.Errorf("Contract options need to be set before calling any function, call Initialize() to initialize contract")
// 	}

// 	// Get ID of submitting client identity
// 	clientAccountID, err := ctx.GetClientIdentity().GetID()
// 	if err != nil {
// 		return "", fmt.Errorf("failed to get client id: %v", err)
// 	}

// 	return clientAccountID, nil
// }

// // TotalSupply returns the total token supply
// func (s *SmartContract) TotalSupply(ctx contractapi.TransactionContextInterface) (int, error) {

// 	// Check if contract has been intilized first
// 	initialized, err := checkInitialized(ctx)
// 	if err != nil {
// 		return 0, fmt.Errorf("failed to check if contract is already initialized: %v", err)
// 	}
// 	if !initialized {
// 		return 0, fmt.Errorf("Contract options need to be set before calling any function, call Initialize() to initialize contract")
// 	}

// 	// Retrieve total supply of tokens from state of smart contract
// 	totalSupplyBytes, err := ctx.GetStub().GetState(totalSupplyKey)
// 	if err != nil {
// 		return 0, fmt.Errorf("failed to retrieve total token supply: %v", err)
// 	}

// 	var totalSupply int

// 	// If no tokens have been minted, return 0
// 	if totalSupplyBytes == nil {
// 		totalSupply = 0
// 	} else {
// 		totalSupply, _ = strconv.Atoi(string(totalSupplyBytes)) // Error handling not needed since Itoa() was used when setting the totalSupply, guaranteeing it was an integer.
// 	}

// 	log.Printf("TotalSupply: %d tokens", totalSupply)

// 	return totalSupply, nil
// }

// // Approve allows the spender to withdraw from the calling client's token account
// // The spender can withdraw multiple times if necessary, up to the value amount
// // This function triggers an Approval event
// func (s *SmartContract) Approve(ctx contractapi.TransactionContextInterface, spender string, value int) error {

// 	// Check if contract has been intilized first
// 	initialized, err := checkInitialized(ctx)
// 	if err != nil {
// 		return fmt.Errorf("failed to check if contract is already initialized: %v", err)
// 	}
// 	if !initialized {
// 		return fmt.Errorf("Contract options need to be set before calling any function, call Initialize() to initialize contract")
// 	}

// 	// Get ID of submitting client identity
// 	owner, err := ctx.GetClientIdentity().GetID()
// 	if err != nil {
// 		return fmt.Errorf("failed to get client id: %v", err)
// 	}

// 	// Create allowanceKey
// 	allowanceKey, err := ctx.GetStub().CreateCompositeKey(allowancePrefix, []string{owner, spender})
// 	if err != nil {
// 		return fmt.Errorf("failed to create the composite key for prefix %s: %v", allowancePrefix, err)
// 	}

// 	// Update the state of the smart contract by adding the allowanceKey and value
// 	err = ctx.GetStub().PutState(allowanceKey, []byte(strconv.Itoa(value)))
// 	if err != nil {
// 		return fmt.Errorf("failed to update state of smart contract for key %s: %v", allowanceKey, err)
// 	}

// 	// Emit the Approval event
// 	approvalEvent := event{owner, spender, value}
// 	approvalEventJSON, err := json.Marshal(approvalEvent)
// 	if err != nil {
// 		return fmt.Errorf("failed to obtain JSON encoding: %v", err)
// 	}
// 	err = ctx.GetStub().SetEvent("Approval", approvalEventJSON)
// 	if err != nil {
// 		return fmt.Errorf("failed to set event: %v", err)
// 	}

// 	log.Printf("client %s approved a withdrawal allowance of %d for spender %s", owner, value, spender)

// 	return nil
// }

// // Allowance returns the amount still available for the spender to withdraw from the owner
// func (s *SmartContract) Allowance(ctx contractapi.TransactionContextInterface, owner string, spender string) (int, error) {

// 	// Check if contract has been intilized first
// 	initialized, err := checkInitialized(ctx)
// 	if err != nil {
// 		return 0, fmt.Errorf("failed to check if contract is already initialized: %v", err)
// 	}
// 	if !initialized {
// 		return 0, fmt.Errorf("Contract options need to be set before calling any function, call Initialize() to initialize contract")
// 	}

// 	// Create allowanceKey
// 	allowanceKey, err := ctx.GetStub().CreateCompositeKey(allowancePrefix, []string{owner, spender})
// 	if err != nil {
// 		return 0, fmt.Errorf("failed to create the composite key for prefix %s: %v", allowancePrefix, err)
// 	}

// 	// Read the allowance amount from the world state
// 	allowanceBytes, err := ctx.GetStub().GetState(allowanceKey)
// 	if err != nil {
// 		return 0, fmt.Errorf("failed to read allowance for %s from world state: %v", allowanceKey, err)
// 	}

// 	var allowance int

// 	// If no current allowance, set allowance to 0
// 	if allowanceBytes == nil {
// 		allowance = 0
// 	} else {
// 		allowance, err = strconv.Atoi(string(allowanceBytes)) // Error handling not needed since Itoa() was used when setting the totalSupply, guaranteeing it was an integer.
// 	}

// 	log.Printf("The allowance left for spender %s to withdraw from owner %s: %d", spender, owner, allowance)

// 	return allowance, nil
// }

// // TransferFrom transfers the value amount from the "from" address to the "to" address
// // This function triggers a Transfer event
// func (s *SmartContract) TransferFrom(ctx contractapi.TransactionContextInterface, from string, to string, value int) error {

// 	// Check if contract has been intilized first
// 	initialized, err := checkInitialized(ctx)
// 	if err != nil {
// 		return fmt.Errorf("failed to check if contract is already initialized: %v", err)
// 	}
// 	if !initialized {
// 		return fmt.Errorf("Contract options need to be set before calling any function, call Initialize() to initialize contract")
// 	}

// 	// Get ID of submitting client identity
// 	spender, err := ctx.GetClientIdentity().GetID()
// 	if err != nil {
// 		return fmt.Errorf("failed to get client id: %v", err)
// 	}

// 	// Create allowanceKey
// 	allowanceKey, err := ctx.GetStub().CreateCompositeKey(allowancePrefix, []string{from, spender})
// 	if err != nil {
// 		return fmt.Errorf("failed to create the composite key for prefix %s: %v", allowancePrefix, err)
// 	}

// 	// Retrieve the allowance of the spender
// 	currentAllowanceBytes, err := ctx.GetStub().GetState(allowanceKey)
// 	if err != nil {
// 		return fmt.Errorf("failed to retrieve the allowance for %s from world state: %v", allowanceKey, err)
// 	}

// 	var currentAllowance int
// 	currentAllowance, _ = strconv.Atoi(string(currentAllowanceBytes)) // Error handling not needed since Itoa() was used when setting the totalSupply, guaranteeing it was an integer.

// 	// Check if transferred value is less than allowance
// 	if currentAllowance < value {
// 		return fmt.Errorf("spender does not have enough allowance for transfer")
// 	}

// 	// Initiate the transfer
// 	err = transferHelper(ctx, from, to, value)
// 	if err != nil {
// 		return fmt.Errorf("failed to transfer: %v", err)
// 	}

// 	// Decrease the allowance
// 	updatedAllowance, err := sub(currentAllowance, value)
// 	if err != nil {
// 		return err
// 	}

// 	err = ctx.GetStub().PutState(allowanceKey, []byte(strconv.Itoa(updatedAllowance)))
// 	if err != nil {
// 		return err
// 	}

// 	// Emit the Transfer event
// 	transferEvent := event{from, to, value}
// 	transferEventJSON, err := json.Marshal(transferEvent)
// 	if err != nil {
// 		return fmt.Errorf("failed to obtain JSON encoding: %v", err)
// 	}
// 	err = ctx.GetStub().SetEvent("Transfer", transferEventJSON)
// 	if err != nil {
// 		return fmt.Errorf("failed to set event: %v", err)
// 	}

// 	log.Printf("spender %s allowance updated from %d to %d", spender, currentAllowance, updatedAllowance)

// 	return nil
// }

// // Name returns a descriptive name for fungible tokens in this contract
// // returns {String} Returns the name of the token

// func (s *SmartContract) Name(ctx contractapi.TransactionContextInterface) (string, error) {

// 	// Check if contract has been intilized first
// 	initialized, err := checkInitialized(ctx)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to check if contract is already initialized: %v", err)
// 	}
// 	if !initialized {
// 		return "", fmt.Errorf("Contract options need to be set before calling any function, call Initialize() to initialize contract")
// 	}

// 	bytes, err := ctx.GetStub().GetState(nameKey)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to get Name bytes: %s", err)
// 	}

// 	return string(bytes), nil
// }

// // Symbol returns an abbreviated name for fungible tokens in this contract.
// // returns {String} Returns the symbol of the token

// func (s *SmartContract) Symbol(ctx contractapi.TransactionContextInterface) (string, error) {

// 	// Check if contract has been intilized first
// 	initialized, err := checkInitialized(ctx)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to check if contract is already initialized: %v", err)
// 	}
// 	if !initialized {
// 		return "", fmt.Errorf("Contract options need to be set before calling any function, call Initialize() to initialize contract")
// 	}

// 	bytes, err := ctx.GetStub().GetState(symbolKey)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to get Symbol: %v", err)
// 	}

// 	return string(bytes), nil
// }

// Set information for a token and intialize contract.
// param {String} name The name of the token
// param {String} symbol The symbol of the token
// param {String} decimals The decimals used for the token operations
func (s *SmartContract) Initialize(ctx contractapi.TransactionContextInterface, name string, symbol string, decimals string, H ristretto.Point, bindingFactor ristretto.Scalar) (bool, error) {

	err := InitPedersen(ctx, H, bindingFactor)
	if err != nil {
		return false, fmt.Errorf("failed to init Pedersen Params: %v", err)
	}
	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to intitialize contract
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return false, fmt.Errorf("failed to get MSPID: %v", err)
	}
	if clientMSPID != "Org1MSP" {
		return false, fmt.Errorf("client is not authorized to initialize contract")
	}

	// Check contract options are not already set, client is not authorized to change them once intitialized
	bytes, err := ctx.GetStub().GetState(nameKey)
	if err != nil {
		return false, fmt.Errorf("failed to get Name: %v", err)
	}
	if bytes != nil {
		return false, fmt.Errorf("contract options are already set, client is not authorized to change them")
	}

	err = ctx.GetStub().PutState(nameKey, []byte(name))
	if err != nil {
		return false, fmt.Errorf("failed to set token name: %v", err)
	}

	err = ctx.GetStub().PutState(symbolKey, []byte(symbol))
	if err != nil {
		return false, fmt.Errorf("failed to set symbol: %v", err)
	}

	err = ctx.GetStub().PutState(decimalsKey, []byte(decimals))
	if err != nil {
		return false, fmt.Errorf("failed to set token name: %v", err)
	}

	return true, nil
}

// Helper Functions

// transferHelper is a helper function that transfers tokens from the "from" address to the "to" address
// Dependant functions include Transfer and TransferFrom
func transferHelper(ctx contractapi.TransactionContextInterface, from string, to string, committedAmount ristretto.Point) error {

	if from == to {
		return fmt.Errorf("cannot transfer to and from same client account")
	}

	fromCurrentBalanceBytes, err := ctx.GetStub().GetState(from)
	if err != nil {
		return fmt.Errorf("failed to read client account %s from world state: %v", from, err)
	}
	if fromCurrentBalanceBytes == nil {
		return fmt.Errorf("client account %s has no balance", from)
	}

	var fromCurrentBalance ristretto.Point //variable to store the current committed balance of sender
	err = fromCurrentBalance.UnmarshalBinary(fromCurrentBalanceBytes)

	//Remove funds from committed amount of sender
	updatedFromBalance := pedersen.Sub(&fromCurrentBalance, &committedAmount)
	//

	toCurrentBalanceBytes, err := ctx.GetStub().GetState(from)
	if err != nil {
		return fmt.Errorf("failed to read target account %s from world state: %v", from, err)
	}

	var toCurrentBalance ristretto.Point //variable to store the current committed balance of sender
	err = toCurrentBalance.UnmarshalBinary(toCurrentBalanceBytes)

	//add funds to recipient
	updatedToBalance := pedersen.Add(&toCurrentBalance, &committedAmount)
	//

	updatedFromBalanceBytes, err := fromCurrentBalance.MarshalBinary()
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(from, updatedFromBalanceBytes)
	if err != nil {
		return err
	}

	updatedToBalanceBytes, err := toCurrentBalance.MarshalBinary()
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(from, updatedToBalanceBytes)
	if err != nil {
		return err
	}

	log.Printf("client %s balance updated from %d to %d", from, fromCurrentBalance, updatedFromBalance)
	log.Printf("recipient %s balance updated from %d to %d", to, toCurrentBalance, updatedToBalance)

	return nil
}

// add two number checking for overflow
func add(b int, q int) (int, error) {

	// Check overflow
	var sum int
	sum = q + b

	if (sum < q || sum < b) == (b >= 0 && q >= 0) {
		return 0, fmt.Errorf("Math: addition overflow occurred %d + %d", b, q)
	}

	return sum, nil
}

// Checks that contract options have been already initialized
func checkInitialized(ctx contractapi.TransactionContextInterface) (bool, error) {
	tokenName, err := ctx.GetStub().GetState(nameKey)
	if err != nil {
		return false, fmt.Errorf("failed to get token name: %v", err)
	}

	if tokenName == nil {
		return false, nil
	}

	return true, nil
}

// sub two number checking for overflow
func sub(b int, q int) (int, error) {

	// Check overflow
	var diff int
	diff = b - q

	if (diff > b) == (b >= 0 && q >= 0) {
		return 0, fmt.Errorf("Math: Subtraction overflow occurred  %d - %d", b, q)
	}

	return diff, nil
}
