package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/bwesterb/go-ristretto"
	"github.com/hyperledger/fabric-chaincode-go/shim"
)

const BLOCK_GENERATION_TIME = 10

type TxInformation struct {
	Sender              string
	Amount              []byte
	ProposalBlockNumber int64
	isValid             bool //will be needed to avoid double spending, otherwise a recipient could approve several times within the time the contract exists.
} //Since we are using a different temp address per each transaction, this won't happen anyway. Implementing this allows us to have a single temp account

// TODO: use pointers, you MUST on stubs for example
func createTxInfo(stub shim.ChaincodeStubInterface, sender string, amount ristretto.Point) (TxInformation, error) {
	amountBytes, err := amount.MarshalBinary()
	if err != nil {
		return TxInformation{}, err
	}
	blockNumber, err := GetBlockNumber(stub)
	if err != nil {
		return TxInformation{}, err
	}

	txInfo := TxInformation{
		Sender:              sender,
		Amount:              amountBytes,
		ProposalBlockNumber: blockNumber,
		isValid:             true,
	}
	return txInfo, nil
}

func storeTxInfo(stub shim.ChaincodeStubInterface, sender string, amount ristretto.Point) error {
	txInfo, err := createTxInfo(stub, sender, amount)
	if err != nil {
		return err
	}

	transferDetailsBytes, err := json.Marshal(txInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal: %v", err)
	}
	txId := stub.GetTxID()
	err = stub.PutState(txId, transferDetailsBytes)
	if err != nil {
		return err
	}

	return nil
}

// TODO: better to return the whole struct and afterwards unmarshal the amount. In this way we can create a modifier to invalidate it
func getTxInfo(stub shim.ChaincodeStubInterface, TxId string) (TxInformation, error) {
	TxInfoBytes, err := stub.GetState(TxId)
	if err != nil {
		return TxInformation{}, fmt.Errorf("failed to read transaction information from world state: %v", err)
	} //TODO: may not be correct to return true if fails
	var txInfo TxInformation
	json.Unmarshal(TxInfoBytes, &txInfo)
	if txInfo.Amount == nil {
		return TxInformation{}, fmt.Errorf("temporary account has no balance")
	}
	var committedAmount ristretto.Point                  //variable to store the current committed balance of sender
	err = committedAmount.UnmarshalBinary(txInfo.Amount) //recipient should be clientId
	if err != nil {
		return TxInformation{}, fmt.Errorf("error unmarshalling")
	}
	return txInfo, nil
}

func GetBlockNumber(stub shim.ChaincodeStubInterface) (int64, error) {
	// Get the transaction timestamp
	txTimestamp, err := stub.GetTxTimestamp()
	if err != nil {
		return 0, err
	}
	// Calculate the approximate block number
	blockNumber := txTimestamp.GetSeconds() / BLOCK_GENERATION_TIME // Assuming 10 seconds per block

	return blockNumber, nil
}

func (txInfo *TxInformation) InvalidateTx() {
	txInfo.isValid = false
}
